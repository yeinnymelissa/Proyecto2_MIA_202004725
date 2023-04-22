package main

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

//----------------------------STRUCTS-----------------------------
type task struct {
	Consola string `json:"Consola"`
}

type particion struct {
	part_status byte
	part_type   byte
	part_fit    byte
	part_start  int32
	part_s      int32
	part_name   [16]byte
}

type ebr struct {
	part_status byte
	part_fit    byte
	part_start  int32
	part_s      int32
	part_next   int32
	part_name   [16]byte
}

type mbr struct {
	mbr_tamano         int32
	mbr_fecha_creacion time.Time
	mbr_dsk_signature  int32
	dsk_fit            byte
	particiones        [4]particion
}

type mountFormat struct {
	id         int
	idMount    string
	name       string
	path       string
	discoName  string
	part_fit   byte
	part_start int
	part_s     int
	part_next  int
	part_type  byte
	s_mtime    *time.Time
}

type Analizador struct {
	mensaje  string
	mountMap map[string]mountFormat
}

func NewAnalizador(mensaje string, mountMap map[string]mountFormat) *Analizador {
	return &Analizador{
		mensaje:  mensaje,
		mountMap: mountMap,
	}
}

func (a *Analizador) ejecutarAnalizador() {
	var tokens []string
	ss := strings.NewReader(a.mensaje)
	scanner := bufio.NewScanner(ss)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		tokens = append(tokens, scanner.Text())
	}
	for _, token := range tokens {
		separarPorEspacios(token)
	}
}

func separarPorEspacios(instruccion string) {
	insChar := []byte(instruccion)
	var tokens []string
	if insChar[0] == '#' {
		fmt.Println("Se encontro un comentario.")
	} else {
		encontreComi := 0
		var token string
		for i := 0; i < len(instruccion)+1; i++ {
			if encontreComi == 1 {
				if insChar[i] == '"' {
					encontreComi = 0
					if i == len(instruccion) {
						tokens = append(tokens, token)
						token = ""
					}
				} else {
					token += string(insChar[i])
				}
			} else if i == len(instruccion) {
				tokens = append(tokens, token)
				token = ""
			} else {
				if insChar[i] == ' ' {
					tokens = append(tokens, token)
					token = ""
				} else if insChar[i] == '"' {
					encontreComi++
				} else {
					token += string(insChar[i])
				}
			}
		}
	}

	if tokens[0] == "execute" {
		exec(tokens)
	} else if tokens[0] == "mkdisk" {
		mkdisk(tokens)
	} else {
		fmt.Println("No se reconoce el comando a ejecutar.")
	}
}

func exec(tokens []string) {
	var execIns []string
	ss := strings.NewReader(tokens[1])
	scanner := bufio.NewScanner(ss)
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		execIns = append(execIns, scanner.Text())
	}
	nombreArchivo := execIns[1]
	archivo, err := os.Open(nombreArchivo)
	if err != nil {
		log.Fatal(err)
	}
	defer archivo.Close()
	var instruccionesExec []string
	scanner = bufio.NewScanner(archivo)
	for scanner.Scan() {
		linea := scanner.Text()
		if linea != "" {
			instruccionesExec = append(instruccionesExec, linea)
		}
	}
	for _, token := range instruccionesExec {
		tokenAux := strings.ToLower(token)
		separarPorEspacios(tokenAux)
	}
}

func mkdisk(tokens []string) {
	size := 0
	unit := ""
	path := ""
	fit := ""
	var fitD byte
	error := false
	for i := 1; i < len(tokens); i++ {
		parameter := strings.Split(tokens[i], "=")
		if parameter[0] == ">size" {
			size, _ = strconv.Atoi(parameter[1])
			if size <= 0 {
				fmt.Println("El size del disco debe ser mayor a 0.")
				error = true
			}
		} else if parameter[0] == ">unit" {
			unit = strings.ToLower(parameter[1])
		} else if parameter[0] == ">path" {
			path = parameter[1]
		} else if parameter[0] == ">fit" {
			fit = strings.ToLower(parameter[1])
		} else {
			fmt.Println("No se reconoce el comando.")
			error = true
		}
	}
	if size <= 0 || path == "" {
		fmt.Println("El tamaño y el path son obligatorios.")
		error = true
	}
	buffer := make([]byte, 1024)
	for i := range buffer {
		buffer[i] = '\x00'
	}

	var aux int = 0
	if unit == "k" {
		aux = size
	} else if unit == "m" {
		aux = size * 1024
	} else if unit == "" {
		aux = size * 1024
		unit = "m"
	} else {
		fmt.Println("La unidades no son validas.")
		error = true
	}
	if fit == "bf" {
		fitD = 'B'
	} else if fit == "ff" {
		fitD = 'F'
	} else if fit == "wf" {
		fitD = 'W'
	} else if fit == "" {
		fitD = 'F'
	} else {
		fmt.Println("El tipo de fit no es valido.")
		error = true
	}

	if error == false {
		var auxSize int
		if unit == "k" {
			auxSize = size * 1024
		} else if unit == "m" {
			auxSize = size * 1024 * 1024
		}
		rand.Seed(time.Now().Unix())
		var m mbr
		m.mbr_tamano = int32(auxSize)
		m.mbr_fecha_creacion = time.Now()
		m.dsk_fit = fitD
		m.mbr_dsk_signature = int32(rand.Intn(100000))
		var nada string = ""
		for i := 0; i < 4; i++ {
			m.particiones[i].part_start = -1
			m.particiones[i].part_type = 'X'
			m.particiones[i].part_fit = 'X'
			m.particiones[i].part_status = 'X'
			m.particiones[i].part_s = -1
			copy(m.particiones[i].part_name[:], []byte(nada))
		}
		crearCarpetas(path)
		f, err := os.Create(path)
		if err != nil {
			log.Fatal(err)
		}
		var i2 int = 0
		for i2 != aux {
			f.Write(buffer[:])
			i2++
		}
		f.Seek(0, 0)
		err = binary.Write(f, binary.LittleEndian, &m)
		if err != nil {
			fmt.Println("\033[31m[Error] > Al escribir en el archivo:", err, "\033[0m")
			return
		}
		defer f.Close()
		fmt.Println("Se creo el disco con exito.")
	}
}

//--------------------------------FUNCIONES GENERALES------------------------------

func crearCarpetas(path string) {
	parameter := strings.Split(path, "/")
	var pathCrear string
	for i := 0; i < len(parameter); i++ {
		if i == len(parameter)-1 || i == 0 {
			continue
		}
		pathCrear += "/" + parameter[i]
	}
	os.MkdirAll(pathCrear, os.ModePerm)
}

func fileExists(name string) bool {
	if _, err := os.Stat(name); os.IsNotExist(err) {
		return false
	}
	return true
}

func main() {
	http.HandleFunc("/read", func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		var data = []byte(b)
		var tarea task
		err = json.Unmarshal(data, &tarea)

		if err != nil {
			fmt.Printf("Error decodificando: %v\n", err)
		} else {
			fmt.Printf("El nombre: %s\n", tarea.Consola)
		}

		separarPorEspacios(tarea.Consola)
	})

	srv := http.Server{
		Addr: ":8080",
	}

	err := srv.ListenAndServe()

	if err != nil {
		panic(err)
	}
}
