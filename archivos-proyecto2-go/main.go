package main

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

//----------------------------STRUCTS-----------------------------
type task struct {
	Consola string `json:"Consola"`
}

type Particion struct {
	Part_status byte
	Part_type   byte
	Part_fit    byte
	Part_start  int32
	Part_s      int32
	Part_name   [16]byte
}

type Ebr struct {
	Part_status byte
	Part_fit    byte
	Part_start  int32
	Part_s      int32
	Part_next   int32
	Part_name   [16]byte
}

type Mbr struct {
	Mbr_tamano         int32
	Mbr_fecha_creacion [16]byte
	Mbr_dsk_signature  int32
	Dsk_fit            byte
	Particiones        [4]Particion
}

type MountFormat struct {
	Id         int
	IdDisco    int
	IdMount    string
	Name       string
	Path       string
	DiscoName  string
	Part_fit   byte
	Part_start int
	Part_s     int
	Part_next  int
	Part_type  byte
	S_mtime    time.Time
}

type SuperBloque struct {
	S_filesystem_type   int32
	S_inodes_count      int32
	S_blocks_count      int32
	S_free_blocks_count int32
	S_free_inodes_count int32
	S_mtime             [16]byte
	S_mnt_count         int32
	S_magic             int32
	S_inode_size        int32
	S_block_size        int32
	S_firts_ino         int32
	S_first_blo         int32
	S_bm_inode_start    int32
	S_bm_block_start    int32
	S_inode_start       int32
	S_block_start       int32
}

type Inodo struct {
	I_uid   int32
	I_gid   int32
	I_size  int32
	I_atime [16]byte
	I_ctime [16]byte
	I_mtime [16]byte
	I_block [16]int32
	I_type  byte
	I_perm  int32
}

type Content struct {
	B_name  [12]byte
	B_inodo int32
}

type BloqueCarpetas struct {
	B_content [4]Content
}

type BloqueArchivos struct {
	B_content [64]byte
}

var mountMap = make(map[string]MountFormat)

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
	} else if tokens[0] == "rmdisk" {
		rmdisk(tokens)
	} else if tokens[0] == "fdisk" {
		fdisk(tokens)
	} else if tokens[0] == "mount" {
		mount(tokens)
	} else if tokens[0] == "mkfs" {
		mkfs(tokens)
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

//------------------------------MKDISK----------------------------

func mkdisk(tokens []string) {
	size := 0
	unit := ""
	path := ""
	fit := ""
	var fitD byte
	error := false
	for i := 1; i < len(tokens); i++ {
		parameter := strings.Split(tokens[i], "=")
		parameter[0] = strings.ToLower(parameter[0])
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
		var m Mbr
		m.Mbr_tamano = int32(auxSize)
		t := time.Now()
		tiempo := strconv.Itoa(t.Year()) + "-" + strconv.Itoa(int(t.UTC().Month())) + "-" + strconv.Itoa(t.Day()) + " " + strconv.Itoa(t.Hour()) + ":" + strconv.Itoa(t.Minute())
		var array [16]byte
		longitud := len(array)
		for i := 0; i < longitud; i++ {
			if i < len(tiempo) {
				array[i] = tiempo[i]
			} else {
				array[i] = ' '
			}
		}
		m.Mbr_fecha_creacion = array
		m.Dsk_fit = fitD
		m.Mbr_dsk_signature = int32(rand.Intn(100000))
		var nada string = ""
		for i := 0; i < 4; i++ {
			m.Particiones[i].Part_start = -1
			m.Particiones[i].Part_type = 'X'
			m.Particiones[i].Part_fit = 'X'
			m.Particiones[i].Part_status = 'X'
			m.Particiones[i].Part_s = -1
			copy(m.Particiones[i].Part_name[:], []byte(nada))
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

func rmdisk(tokens []string) {
	path := ""
	error := false
	for i := 1; i < len(tokens); i++ {
		parameter := make([]string, 0)
		sp := strings.Split(tokens[i], "=")
		for _, tokenPar := range sp {
			parameter = append(parameter, tokenPar)
		}
		if parameter[0] == ">path" {
			path = parameter[1]
		} else {
			fmt.Println("No se reconoce el comando.")
			error = true
		}
	}
	if path == "" {
		fmt.Println("El path es obligatorio para realizar el rmdisk.")
		error = true
	}
	if error == false {
		estaBien := fileExists(path)
		if estaBien == true {
			err := os.Remove(path)
			if err != nil {
				fmt.Println("Error al borrar archivo!.")
			} else {
				fmt.Println("El archivo se borro con exito!")
			}
		} else {
			fmt.Println("El archivo con el path: ", path, ", no existe.")
		}
	}
}

func fdisk(tokens []string) {
	size := 0
	unit := ""
	path := ""
	name := ""
	typeVal := ""
	fit := ""
	errorVal := false
	for i := 1; i < len(tokens); i++ {
		parameter := strings.Split(tokens[i], "=")
		parameter[0] = strings.ToLower(parameter[0])
		if parameter[0] == ">size" {
			size, _ = strconv.Atoi(parameter[1])
			if size <= 0 {
				fmt.Println("El size de la particion debe ser mayor a 0.")
				errorVal = true
			}
		} else if parameter[0] == ">unit" {
			unit = strings.ToLower(parameter[1])
		} else if parameter[0] == ">path" {
			path = parameter[1]
		} else if parameter[0] == ">name" {
			name = parameter[1]
		} else if parameter[0] == ">type" {
			typeVal = strings.ToLower(parameter[1])
		} else if parameter[0] == ">fit" {
			fit = strings.ToLower(parameter[1])
		} else {
			fmt.Println("No se reconoce el comando.")
			errorVal = true
		}
	}

	if size <= 0 || path == "" || name == "" {
		fmt.Println("El size, el path y el nombre son obligatorios.")
		errorVal = true
	}
	aux := 0
	if unit == "b" {
		aux = size
	} else if unit == "k" {
		aux = size * 1024
	} else if unit == "m" {
		aux = size * 1024 * 1024
	} else if unit == "" {
		aux = size * 1024
		unit = "k"
	} else {
		fmt.Println("La unidades no son validas.")
		errorVal = true
	}

	tipoChar := ' '

	if typeVal == "p" {
		tipoChar = 'P'
	} else if typeVal == "e" {
		tipoChar = 'E'
	} else if typeVal == "l" {
		tipoChar = 'L'
	} else if typeVal == "" {
		typeVal = "p"
		tipoChar = 'P'
	} else {
		fmt.Println("El tipo de particion no es valido.")
		errorVal = true
	}
	fitChar := ' '
	if fit == "bf" {
		fitChar = 'B'
	} else if fit == "ff" {
		fitChar = 'F'
	} else if fit == "wf" {
		fitChar = 'W'
	} else if fit == "" {
		fit = "wf"
		fitChar = 'W'
	} else {
		fmt.Println("El tipo de particion no es valido.")
		errorVal = true
	}

	if errorVal == false {

		var par Particion
		par.Part_s = int32(aux)
		var arrayName [16]byte
		longitud := len(arrayName)
		for i := 0; i < longitud; i++ {
			if i < len(name) {
				arrayName[i] = name[i]
			} else {
				arrayName[i] = '\x00'
			}
		}
		par.Part_name = arrayName

		f, err := os.OpenFile(path, os.O_RDWR, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		var m Mbr
		if err := binary.Read(f, binary.LittleEndian, &m); err != nil {
			log.Fatal(err)
		}

		fitMbr := m.Dsk_fit
		hayExtendida := false
		tamanoExtendida := 0
		startExtendida := 0

		for i := 0; i < 4; i++ {
			if m.Particiones[i].Part_type == 'E' {
				hayExtendida = true
				tamanoExtendida = int(m.Particiones[i].Part_s)
				startExtendida = int(m.Particiones[i].Part_start)
				break
			}
		}

		if typeVal == "e" && hayExtendida == true {
			fmt.Println("No se puede crear mas de una particion extendida.")
		} else if typeVal == "l" && hayExtendida == false {
			fmt.Println("No se puede crear una particion logica si no hay una extendida.")
		} else {
			hayEspacio := false
			for i := 0; i < 4; i++ {
				if m.Particiones[i].Part_name[0] == '\x00' {
					hayEspacio = true
					break
				}
			}
			totalDisco := 0
			for i := 0; i < 4; i++ {
				if m.Particiones[i].Part_s != -1 {
					totalDisco += int(m.Particiones[i].Part_s)
				}
			}
			huboErrorParti := false

			if hayEspacio == false && tipoChar != 'L' {
				fmt.Println("Se han creado el maximo de particiones permitidas.")
				return
			} else {
				if fitMbr == 'F' && tipoChar != 'L' {
					for i := 0; i < 4; i++ {
						if m.Particiones[i].Part_name[0] == '\x00' {
							if m.Particiones[i].Part_s == -1 {
								if i == 0 {
									var info Mbr
									const infoSize = unsafe.Sizeof(info)
									auxParti := int(infoSize) + aux
									totalDisco += aux
									if auxParti <= int(m.Mbr_tamano) && totalDisco <= int(m.Mbr_tamano) {
										var arrayNameAux [16]byte
										longitud := len(arrayNameAux)
										for i := 0; i < longitud; i++ {
											if i < len(name) {
												arrayNameAux[i] = name[i]
											} else {
												arrayNameAux[i] = '\x00'
											}
										}
										m.Particiones[i].Part_name = arrayNameAux
										m.Particiones[i].Part_s = int32(aux)
										m.Particiones[i].Part_type = byte(tipoChar)
										m.Particiones[i].Part_status = 0
										m.Particiones[i].Part_fit = byte(fitChar)
										m.Particiones[i].Part_start = int32(infoSize)
										f.Seek(0, 0)
										binary.Write(f, binary.LittleEndian, &m)
										fmt.Println("Se creo la particion con exito.")
										huboErrorParti = false
										break
									} else {
										huboErrorParti = true
										totalDisco -= aux
									}
								} else {
									auxParti := int(m.Particiones[i-1].Part_start) + int(m.Particiones[i-1].Part_s) + aux
									totalDisco += aux
									if auxParti <= int(m.Mbr_tamano) && totalDisco <= int(m.Mbr_tamano) {
										var arrayNameAux [16]byte
										longitud := len(arrayNameAux)
										for i := 0; i < longitud; i++ {
											if i < len(name) {
												arrayNameAux[i] = name[i]
											} else {
												arrayNameAux[i] = '\x00'
											}
										}
										m.Particiones[i].Part_name = arrayNameAux
										m.Particiones[i].Part_s = int32(aux)
										m.Particiones[i].Part_type = byte(tipoChar)
										m.Particiones[i].Part_status = 0
										m.Particiones[i].Part_fit = byte(fitChar)
										m.Particiones[i].Part_start = m.Particiones[i-1].Part_start + m.Particiones[i-1].Part_s
										f.Seek(0, 0)
										binary.Write(f, binary.LittleEndian, &m)
										fmt.Println("Se creo la particion con exito.")
										huboErrorParti = false
										break
									} else {
										huboErrorParti = true
										totalDisco -= aux
									}
								}
							} else {
								var auxTot int
								auxTot = totalDisco - (int(m.Particiones[i].Part_s) - aux)
								if aux <= int(m.Particiones[i].Part_s) && auxTot <= int(m.Mbr_tamano) {
									var arrayNameAux [16]byte
									longitud := len(arrayNameAux)
									for i := 0; i < longitud; i++ {
										if i < len(name) {
											arrayNameAux[i] = name[i]
										} else {
											arrayNameAux[i] = '\x00'
										}
									}
									m.Particiones[i].Part_name = arrayNameAux
									m.Particiones[i].Part_s = int32(aux)
									m.Particiones[i].Part_type = byte(tipoChar)
									m.Particiones[i].Part_status = 0
									m.Particiones[i].Part_fit = byte(fitChar)
									f.Seek(0, 0)
									binary.Write(f, binary.LittleEndian, &m)
									fmt.Println("Se creo la particion con exito.")
									huboErrorParti = false
									break
								} else {
									huboErrorParti = true
								}
							}
						}
					}
				} else if fitMbr == 'F' && tipoChar == 'L' {
					contador := 0
					var auxEbr byte
					_, err := f.Seek(int64(startExtendida), 0)
					if err != nil {
						panic(err)
					}
					err = binary.Read(f, binary.LittleEndian, &auxEbr)
					if err != nil {
						panic(err)
					}
					tamanoUsado := 0
					for auxEbr != '\x00' {
						var e Ebr
						if auxEbr == '1' {
							contador++
							err = binary.Read(f, binary.LittleEndian, &e)
							if err != nil {
								panic(err)
							}
							tamanoUsado += int(e.Part_s)
							fmt.Println(e.Part_next)
							if e.Part_next != -1 {
								_, err = f.Seek(int64(e.Part_next), 0)
								if err != nil {
									panic(err)
								}
								err = binary.Read(f, binary.LittleEndian, &auxEbr)
								if err != nil {
									panic(err)
								}
							} else {
								break
							}
						}
					}
					startPart := 0
					_, err = f.Seek(int64(startExtendida), 0)
					if err != nil {
						panic(err)
					}
					err = binary.Read(f, binary.LittleEndian, &auxEbr)
					if err != nil {
						panic(err)
					}
					noPuedoAgregar := true
					for auxEbr != '\x00' {
						var e Ebr
						if auxEbr == '1' {
							err = binary.Read(f, binary.LittleEndian, &e)
							if err != nil {
								panic(err)
							}
							if e.Part_name[0] == '\x00' {
								auxUsado := tamanoUsado - (int(e.Part_s) - aux)
								if aux <= int(e.Part_s) && auxUsado <= tamanoExtendida {
									_, err = f.Seek(int64(binary.Size(e))*-1, 1)
									if err != nil {
										panic(err)
									}
									var arrayNameAux [16]byte
									longitud := len(arrayNameAux)
									for i := 0; i < longitud; i++ {
										if i < len(name) {
											arrayNameAux[i] = name[i]
										} else {
											arrayNameAux[i] = '\x00'
										}
									}
									e.Part_name = arrayNameAux
									e.Part_s = int32(aux)
									e.Part_status = 0
									e.Part_fit = byte(fitChar)
									err = binary.Write(f, binary.LittleEndian, &e)
									if err != nil {
										panic(err)
									}
									noPuedoAgregar = false
									fmt.Println("Se creo la particion con exito.")
									huboErrorParti = false
									break
								}
							}
							startPart = int(e.Part_start)
							fmt.Println(e.Part_next)
							if e.Part_next != -1 {
								_, err = f.Seek(int64(e.Part_next), 0)
								if err != nil {
									panic(err)
								}
								err = binary.Read(f, binary.LittleEndian, &auxEbr)
								if err != nil {
									panic(err)
								}
							} else {
								break
							}
						}
					}

					if noPuedoAgregar {
						if contador > 0 {
							auxUsado := tamanoUsado + aux
							if auxUsado <= tamanoExtendida {
								_, err := f.Seek(int64(startPart+int(unsafe.Sizeof(byte(0)))), 0)
								if err != nil {
									panic(err)
								}
								eb := Ebr{}
								err = binary.Read(f, binary.LittleEndian, &eb)
								if err != nil {
									panic(err)
								}
								eb.Part_next = eb.Part_start + eb.Part_s + int32(unsafe.Sizeof(byte(0)))
								_, err = f.Seek(int64(startPart+int(unsafe.Sizeof(byte(0)))), 0)
								if err != nil {
									panic(err)
								}
								err = binary.Write(f, binary.LittleEndian, &eb)
								if err != nil {
									panic(err)
								}
								aux2 := byte('1')
								e := Ebr{}
								copy(e.Part_name[:], []byte(name)[:])
								e.Part_s = int32(aux)
								e.Part_status = 0
								e.Part_fit = byte(fitChar)
								e.Part_start = eb.Part_start + eb.Part_s + int32(unsafe.Sizeof(byte(0)))
								e.Part_next = -1
								start := eb.Part_start + eb.Part_s + int32(unsafe.Sizeof(byte(0)))
								_, err = f.Seek(int64(start), 0)
								if err != nil {
									panic(err)
								}
								err = binary.Write(f, binary.LittleEndian, &aux2)
								if err != nil {
									panic(err)
								}
								err = binary.Write(f, binary.LittleEndian, &e)
								if err != nil {
									panic(err)
								}
								fmt.Println("Se creo la particion con exito.")
								huboErrorParti = false
							} else {
								fmt.Println("No hay espacio en la particion extendida para crear la particion logica.")
								huboErrorParti = true
							}
						} else {
							auxUsado := tamanoUsado + aux
							if auxUsado <= tamanoExtendida {
								_, err := f.Seek(int64(startExtendida), 0)
								if err != nil {
									panic(err)
								}
								aux2 := byte('1')
								e := Ebr{}
								copy(e.Part_name[:], []byte(name)[:])
								e.Part_s = int32(aux)
								e.Part_status = 0
								e.Part_fit = byte(fitChar)
								e.Part_start = int32(startExtendida)
								e.Part_next = -1
								err = binary.Write(f, binary.LittleEndian, &aux2)
								if err != nil {
									panic(err)
								}
								err = binary.Write(f, binary.LittleEndian, &e)
								if err != nil {
									panic(err)
								}
								fmt.Println("Se creo la particion con exito.")
								huboErrorParti = false
							} else {
								fmt.Println("No hay espacio en la particion extendida para crear la particion logica.")
								huboErrorParti = true
							}
						}
					}
				} else if fitMbr == 'W' && tipoChar != 'L' {
					var elPeor int
					elPeor = -1
					var valPeor int = 0
					valPeor = 0
					for i := 0; i < 4; i++ {
						if m.Particiones[i].Part_s != -1 && m.Particiones[i].Part_name[0] == '\x00' {
							if aux <= int(m.Particiones[i].Part_s) && int(m.Particiones[i].Part_s) > valPeor {
								valPeor = int(m.Particiones[i].Part_s)
								elPeor = i
							}
						}
					}
					if elPeor != -1 {
						for i := 0; i < 4; i++ {
							if i == elPeor {
								var arrayNameAux [16]byte
								longitud := len(arrayNameAux)
								for i := 0; i < longitud; i++ {
									if i < len(name) {
										arrayNameAux[i] = name[i]
									} else {
										arrayNameAux[i] = '\x00'
									}
								}
								m.Particiones[i].Part_name = arrayNameAux
								m.Particiones[i].Part_s = int32(aux)
								m.Particiones[i].Part_type = byte(tipoChar)
								m.Particiones[i].Part_status = 0
								m.Particiones[i].Part_fit = byte(fitChar)
								f.Seek(0, 0)
								binary.Write(f, binary.LittleEndian, &m)
								fmt.Println("Se creo la particion con exito.")
								huboErrorParti = false
								break
							}
						}
					} else {
						for i := 0; i < 4; i++ {
							if m.Particiones[i].Part_s == -1 && m.Particiones[i].Part_name[0] == '\x00' {
								var info Mbr
								const infoSize = unsafe.Sizeof(info)
								if i == 0 {
									auxParti := int(infoSize) + aux
									totalDisco += aux
									if auxParti <= int(m.Mbr_tamano) && totalDisco <= int(m.Mbr_tamano) {
										var arrayNameAux [16]byte
										longitud := len(arrayNameAux)
										for i := 0; i < longitud; i++ {
											if i < len(name) {
												arrayNameAux[i] = name[i]
											} else {
												arrayNameAux[i] = '\x00'
											}
										}
										m.Particiones[i].Part_name = arrayNameAux
										m.Particiones[i].Part_s = int32(aux)
										m.Particiones[i].Part_type = byte(tipoChar)
										m.Particiones[i].Part_status = 0
										m.Particiones[i].Part_fit = byte(fitChar)
										m.Particiones[i].Part_start = int32(infoSize)
										f.Seek(0, 0)
										binary.Write(f, binary.LittleEndian, &m)
										fmt.Println("Se creo la particion con exito.")
										huboErrorParti = false
										break
									} else {
										huboErrorParti = true
										totalDisco -= aux
									}
								} else {
									auxParti := int(m.Particiones[i-1].Part_start+m.Particiones[i-1].Part_s) + aux
									totalDisco += aux
									if auxParti <= int(m.Mbr_tamano) && totalDisco <= int(m.Mbr_tamano) {
										var arrayNameAux [16]byte
										longitud := len(arrayNameAux)
										for i := 0; i < longitud; i++ {
											if i < len(name) {
												arrayNameAux[i] = name[i]
											} else {
												arrayNameAux[i] = '\x00'
											}
										}
										m.Particiones[i].Part_name = arrayNameAux
										m.Particiones[i].Part_s = int32(aux)
										m.Particiones[i].Part_type = byte(tipoChar)
										m.Particiones[i].Part_status = 0
										m.Particiones[i].Part_fit = byte(fitChar)
										m.Particiones[i].Part_start = m.Particiones[i-1].Part_start + m.Particiones[i-1].Part_s
										f.Seek(0, 0)
										binary.Write(f, binary.LittleEndian, &m)
										fmt.Println("Se creo la particion con exito.")
										huboErrorParti = false
										break
									} else {
										huboErrorParti = true
										totalDisco -= aux
									}
								}
							}
						}
					}

				} else if fitMbr == 'W' && tipoChar == 'L' {
					contador := 0
					startPart := 0
					var auxEbr byte
					_, err := f.Seek(int64(startExtendida), 0)
					if err != nil {
						panic(err)
					}
					err = binary.Read(f, binary.LittleEndian, &auxEbr)
					if err != nil {
						panic(err)
					}
					tamanoUsado := 0
					for auxEbr != '\x00' {
						var e Ebr
						if auxEbr == '1' {
							contador++
							err = binary.Read(f, binary.LittleEndian, &e)
							if err != nil {
								panic(err)
							}
							tamanoUsado += int(e.Part_s)
							if e.Part_next != -1 {
								_, err = f.Seek(int64(e.Part_next), 0)
								if err != nil {
									panic(err)
								}
								err = binary.Read(f, binary.LittleEndian, &auxEbr)
								if err != nil {
									panic(err)
								}
							} else {
								break
							}
						}
					}
					_, err = f.Seek(int64(startExtendida), 0)
					if err != nil {
						panic(err)
					}
					err = binary.Read(f, binary.LittleEndian, &auxEbr)
					if err != nil {
						panic(err)
					}
					noPuedoAgregar := true
					cont := 0
					esPeor := -1
					valPeor := 0
					for auxEbr != '\x00' {
						var e Ebr
						if auxEbr == '1' {
							err = binary.Read(f, binary.LittleEndian, &e)
							if err != nil {
								panic(err)
							}
							if e.Part_name[0] == '\x00' {
								if aux <= int(e.Part_s) && int(e.Part_s) > valPeor {
									esPeor = cont
									valPeor = int(e.Part_s)
								}
							}
							startPart = int(e.Part_start)
							if e.Part_next != -1 {
								_, err = f.Seek(int64(e.Part_next), 0)
								if err != nil {
									panic(err)
								}
								err = binary.Read(f, binary.LittleEndian, &auxEbr)
								if err != nil {
									panic(err)
								}
							} else {
								break
							}
						}
						cont++
					}
					cont = 0
					if esPeor != -1 {
						_, err = f.Seek(int64(startExtendida), 0)
						if err != nil {
							panic(err)
						}
						err = binary.Read(f, binary.LittleEndian, &auxEbr)
						if err != nil {
							panic(err)
						}
						for auxEbr != '\x00' {
							var e Ebr
							if auxEbr == '1' {
								err = binary.Read(f, binary.LittleEndian, &e)
								if err != nil {
									panic(err)
								}
								if cont == esPeor {
									_, err = f.Seek(int64(binary.Size(e))*-1, 1)
									if err != nil {
										panic(err)
									}
									var arrayNameAux [16]byte
									longitud := len(arrayNameAux)
									for i := 0; i < longitud; i++ {
										if i < len(name) {
											arrayNameAux[i] = name[i]
										} else {
											arrayNameAux[i] = '\x00'
										}
									}
									e.Part_name = arrayNameAux
									e.Part_s = int32(aux)
									e.Part_status = 0
									e.Part_fit = byte(fitChar)
									err = binary.Write(f, binary.LittleEndian, &e)
									if err != nil {
										panic(err)
									}
									noPuedoAgregar = false
									fmt.Println("Se creo la particion con exito.")
									huboErrorParti = false
									break
								}
								if e.Part_next != -1 {
									_, err = f.Seek(int64(e.Part_next), 0)
									if err != nil {
										panic(err)
									}
									err = binary.Read(f, binary.LittleEndian, &auxEbr)
									if err != nil {
										panic(err)
									}
								} else {
									break
								}
							}
							cont++
						}
					}

					if noPuedoAgregar {
						if contador > 0 {
							auxUsado := tamanoUsado + aux
							if auxUsado <= tamanoExtendida {
								_, err := f.Seek(int64(startPart+int(unsafe.Sizeof(byte(0)))), 0)
								if err != nil {
									panic(err)
								}
								var eb Ebr
								err = binary.Read(f, binary.LittleEndian, &eb)
								if err != nil {
									panic(err)
								}
								eb.Part_next = eb.Part_start + eb.Part_s + int32(unsafe.Sizeof(byte(0)))
								_, err = f.Seek(int64(startPart+int(unsafe.Sizeof(byte(0)))), 0)
								if err != nil {
									panic(err)
								}
								err = binary.Write(f, binary.LittleEndian, &eb)
								if err != nil {
									panic(err)
								}
								aux2 := byte('1')
								var e Ebr
								var arrayNameAux [16]byte
								longitud := len(arrayNameAux)
								for i := 0; i < longitud; i++ {
									if i < len(name) {
										arrayNameAux[i] = name[i]
									} else {
										arrayNameAux[i] = '\x00'
									}
								}
								e.Part_name = arrayNameAux
								e.Part_s = int32(aux)
								e.Part_status = 0
								e.Part_fit = byte(fitChar)
								e.Part_start = eb.Part_start + eb.Part_s + int32(unsafe.Sizeof(byte(0)))
								e.Part_next = -1
								_, err = f.Seek(int64(e.Part_start), 0)
								if err != nil {
									panic(err)
								}
								err = binary.Write(f, binary.LittleEndian, &aux2)
								if err != nil {
									panic(err)
								}
								err = binary.Write(f, binary.LittleEndian, &e)
								if err != nil {
									panic(err)
								}
								fmt.Println("Se creo la particion con exito.")
								huboErrorParti = false
							} else {
								fmt.Println("No hay espacio en la particion extendida para crear la particion logica.")
								huboErrorParti = true
							}
						} else {
							auxUsado := tamanoUsado + aux
							if auxUsado <= tamanoExtendida {
								_, err := f.Seek(int64(startExtendida), 0)
								if err != nil {
									panic(err)
								}
								aux2 := byte('1')
								var e Ebr
								var arrayNameAux [16]byte
								longitud := len(arrayNameAux)
								for i := 0; i < longitud; i++ {
									if i < len(name) {
										arrayNameAux[i] = name[i]
									} else {
										arrayNameAux[i] = '\x00'
									}
								}
								e.Part_name = arrayNameAux
								e.Part_s = int32(aux)
								e.Part_status = 0
								e.Part_fit = byte(fitChar)
								e.Part_start = int32(startExtendida)
								e.Part_next = -1
								err = binary.Write(f, binary.LittleEndian, &aux2)
								if err != nil {
									panic(err)
								}
								err = binary.Write(f, binary.LittleEndian, &e)
								if err != nil {
									panic(err)
								}
								fmt.Println("Se creo la particion con exito.")
								huboErrorParti = false
							} else {
								fmt.Println("No hay espacio en la particion extendida para crear la particion logica.")
								huboErrorParti = true
							}
						}

					}

				} else if fitMbr == 'B' && tipoChar != 'L' {
					elMejor := -1
					valMejor := 2147483647
					for i := 0; i < 4; i++ {
						if m.Particiones[i].Part_s != -1 && m.Particiones[i].Part_name[0] == '\x00' {
							if aux <= int(m.Particiones[i].Part_s) && int(m.Particiones[i].Part_s) < valMejor {
								valMejor = int(m.Particiones[i].Part_s)
								elMejor = i
							}
						}
					}
					if elMejor != -1 {
						for i := 0; i < 4; i++ {
							if i == elMejor {
								var info Mbr
								const infoSize = unsafe.Sizeof(info)
								var arrayNameAux [16]byte
								longitud := len(arrayNameAux)
								for i := 0; i < longitud; i++ {
									if i < len(name) {
										arrayNameAux[i] = name[i]
									} else {
										arrayNameAux[i] = '\x00'
									}
								}
								m.Particiones[i].Part_name = arrayNameAux
								m.Particiones[i].Part_s = int32(aux)
								m.Particiones[i].Part_type = byte(tipoChar)
								m.Particiones[i].Part_status = 0
								m.Particiones[i].Part_fit = byte(fitChar)
								f.Seek(0, 0)
								binary.Write(f, binary.LittleEndian, &m)
								fmt.Println("Se creo la particion con exito.")
								huboErrorParti = false
								break
							}
						}
					} else {
						for i := 0; i < 4; i++ {
							if m.Particiones[i].Part_s == -1 && m.Particiones[i].Part_name[0] == '\x00' {
								if i == 0 {
									var info Mbr
									const infoSize = unsafe.Sizeof(info)
									auxParti := int(infoSize) + aux
									totalDisco += aux
									if auxParti <= int(m.Mbr_tamano) && totalDisco <= int(m.Mbr_tamano) {
										var arrayNameAux [16]byte
										longitud := len(arrayNameAux)
										for i := 0; i < longitud; i++ {
											if i < len(name) {
												arrayNameAux[i] = name[i]
											} else {
												arrayNameAux[i] = '\x00'
											}
										}
										m.Particiones[i].Part_name = arrayNameAux
										m.Particiones[i].Part_s = int32(aux)
										m.Particiones[i].Part_type = byte(tipoChar)
										m.Particiones[i].Part_status = 0
										m.Particiones[i].Part_fit = byte(fitChar)
										m.Particiones[i].Part_start = int32(infoSize)
										f.Seek(0, 0)
										binary.Write(f, binary.LittleEndian, &m)
										fmt.Println("Se creo la particion con exito.")
										huboErrorParti = false
										break
									} else {
										huboErrorParti = true
										totalDisco -= aux
									}
								} else {
									auxParti := int(m.Particiones[i-1].Part_start+m.Particiones[i-1].Part_s) + aux
									totalDisco += aux
									if auxParti <= int(m.Mbr_tamano) && totalDisco <= int(m.Mbr_tamano) {
										var arrayNameAux [16]byte
										longitud := len(arrayNameAux)
										for i := 0; i < longitud; i++ {
											if i < len(name) {
												arrayNameAux[i] = name[i]
											} else {
												arrayNameAux[i] = '\x00'
											}
										}
										m.Particiones[i].Part_name = arrayNameAux
										m.Particiones[i].Part_s = int32(aux)
										m.Particiones[i].Part_type = byte(tipoChar)
										m.Particiones[i].Part_status = 0
										m.Particiones[i].Part_fit = byte(fitChar)
										m.Particiones[i].Part_start = m.Particiones[i-1].Part_start + m.Particiones[i-1].Part_s

										f.Seek(0, 0)
										binary.Write(f, binary.LittleEndian, &m)
										fmt.Println("Se creo la particion con exito.")
										huboErrorParti = false
										break
									} else {
										huboErrorParti = true
										totalDisco -= aux
									}
								}
							}
						}
					}

				} else if fitMbr == 'B' && tipoChar == 'L' {
					contador := 0
					var auxEbr byte
					_, err := f.Seek(int64(startExtendida), 0)
					if err != nil {
						panic(err)
					}
					err = binary.Read(f, binary.LittleEndian, &auxEbr)
					if err != nil {
						panic(err)
					}
					tamanoUsado := 0
					for auxEbr != '\x00' {
						var e Ebr
						if auxEbr == '1' {
							contador++
							err = binary.Read(f, binary.LittleEndian, &e)
							if err != nil {
								panic(err)
							}
							tamanoUsado += int(e.Part_s)
							if e.Part_next != -1 {
								_, err = f.Seek(int64(e.Part_next), 0)
								if err != nil {
									panic(err)
								}
								err = binary.Read(f, binary.LittleEndian, &auxEbr)
								if err != nil {
									panic(err)
								}
							} else {
								break
							}
						}
					}
					_, err = f.Seek(int64(startExtendida), 0)
					if err != nil {
						panic(err)
					}
					err = binary.Read(f, binary.LittleEndian, &auxEbr)
					if err != nil {
						panic(err)
					}
					noPuedoAgregar := true
					cont := 0
					esMejor := -1
					valMejor := 2147483647
					startPart := 0
					for auxEbr != '\x00' {
						var e Ebr
						if auxEbr == '1' {
							err = binary.Read(f, binary.LittleEndian, &e)
							if err != nil {
								panic(err)
							}
							if e.Part_name[0] == '\x00' {
								if aux <= int(e.Part_s) && int(e.Part_s) < valMejor {
									esMejor = cont
									valMejor = int(e.Part_s)
								}
							}
							startPart = int(e.Part_start)
							if e.Part_next != -1 {
								_, err = f.Seek(int64(e.Part_next), 0)
								if err != nil {
									panic(err)
								}
								err = binary.Read(f, binary.LittleEndian, &auxEbr)
								if err != nil {
									panic(err)
								}
							} else {
								break
							}
						}
						cont++
					}
					cont = 0
					if esMejor != -1 {
						_, err = f.Seek(int64(startExtendida), 0)
						if err != nil {
							panic(err)
						}
						err = binary.Read(f, binary.LittleEndian, &auxEbr)
						if err != nil {
							panic(err)
						}
						for auxEbr != '\x00' {
							var e Ebr
							if auxEbr == '1' {
								err = binary.Read(f, binary.LittleEndian, &e)
								if err != nil {
									panic(err)
								}
								if cont == esMejor {
									_, err = f.Seek(int64(binary.Size(e))*-1, 1)
									if err != nil {
										panic(err)
									}
									var e Ebr
									var arrayNameAux [16]byte
									longitud := len(arrayNameAux)
									for i := 0; i < longitud; i++ {
										if i < len(name) {
											arrayNameAux[i] = name[i]
										} else {
											arrayNameAux[i] = '\x00'
										}
									}
									e.Part_name = arrayNameAux
									e.Part_s = int32(aux)
									e.Part_status = 0
									e.Part_fit = byte(fitChar)
									err = binary.Write(f, binary.LittleEndian, &e)
									if err != nil {
										panic(err)
									}
									noPuedoAgregar = false
									fmt.Println("Se creo la particion con exito.")
									huboErrorParti = false
									break
								}
								if e.Part_next != -1 {
									_, err = f.Seek(int64(e.Part_next), 0)
									if err != nil {
										panic(err)
									}
									err = binary.Read(f, binary.LittleEndian, &auxEbr)
									if err != nil {
										panic(err)
									}
								} else {
									break
								}
							}
							cont++
						}
					}

					if noPuedoAgregar {
						if contador > 0 {
							auxUsado := tamanoUsado + aux
							if auxUsado <= tamanoExtendida {
								_, err := f.Seek(int64(startPart+int(unsafe.Sizeof(byte(0)))), 0)
								if err != nil {
									panic(err)
								}
								var eb Ebr
								err = binary.Read(f, binary.LittleEndian, &eb)
								if err != nil {
									panic(err)
								}
								eb.Part_next = eb.Part_start + eb.Part_s + int32(unsafe.Sizeof(byte(0)))
								_, err = f.Seek(int64(startPart+int(unsafe.Sizeof(byte(0)))), 0)
								if err != nil {
									panic(err)
								}
								err = binary.Write(f, binary.LittleEndian, &eb)
								if err != nil {
									panic(err)
								}
								aux2 := byte('1')
								var e Ebr
								var arrayNameAux [16]byte
								longitud := len(arrayNameAux)
								for i := 0; i < longitud; i++ {
									if i < len(name) {
										arrayNameAux[i] = name[i]
									} else {
										arrayNameAux[i] = '\x00'
									}
								}
								e.Part_name = arrayNameAux
								e.Part_s = int32(aux)
								e.Part_status = 0
								e.Part_fit = byte(fitChar)
								e.Part_start = eb.Part_start + eb.Part_s + int32(unsafe.Sizeof(byte(0)))
								e.Part_next = -1
								_, err = f.Seek(int64(e.Part_start), 0)
								if err != nil {
									panic(err)
								}
								err = binary.Write(f, binary.LittleEndian, &aux2)
								if err != nil {
									panic(err)
								}
								err = binary.Write(f, binary.LittleEndian, &e)
								if err != nil {
									panic(err)
								}
								fmt.Println("Se creo la particion con exito.")
								huboErrorParti = false
							} else {
								fmt.Println("No hay espacio en la particion extendida para crear la particion logica.")
								huboErrorParti = true
							}
						} else {
							auxUsado := tamanoUsado + aux
							if auxUsado <= tamanoExtendida {
								_, err := f.Seek(int64(startExtendida), 0)
								if err != nil {
									panic(err)
								}
								aux2 := byte('1')
								var e Ebr
								var arrayNameAux [16]byte
								longitud := len(arrayNameAux)
								for i := 0; i < longitud; i++ {
									if i < len(name) {
										arrayNameAux[i] = name[i]
									} else {
										arrayNameAux[i] = '\x00'
									}
								}
								e.Part_name = arrayNameAux
								e.Part_s = int32(aux)
								e.Part_status = 0
								e.Part_fit = byte(fitChar)
								e.Part_start = int32(startExtendida)
								e.Part_next = -1
								err = binary.Write(f, binary.LittleEndian, &aux2)
								if err != nil {
									panic(err)
								}
								err = binary.Write(f, binary.LittleEndian, &e)
								if err != nil {
									panic(err)
								}
								fmt.Println("Se creo la particion con exito.")
								huboErrorParti = false
							} else {
								fmt.Println("No hay espacio en la particion extendida para crear la particion logica.")
								huboErrorParti = true
							}
						}
					}

				}

				if huboErrorParti == true {
					fmt.Println("No se pudo realizar la particion.")
				}
			}
		}
	}
}

func mount(tokens []string) {
	name := ""
	path := ""
	errorVal := false
	for i := 1; i < len(tokens); i++ {
		parameter := strings.Split(tokens[i], "=")
		if parameter[0] == ">name" {
			name = parameter[1]
		} else if parameter[0] == ">path" {
			path = parameter[1]
		} else {
			fmt.Println("No se reconoce el comando.")
			errorVal = true
		}
	}
	if name == "" || path == "" {
		fmt.Println("El nombre y el path son obligatorios.")
		errorVal = true
	}
	if errorVal == false {
		if fileExists(path) == true {
			encontre := false
			pathVec := strings.Split(path, "/")
			nameDisco := ""
			for i := 0; i < len(pathVec); i++ {
				if i == len(pathVec)-1 {
					nameVec := strings.Split(pathVec[i], ".")
					nameDisco = nameVec[0]
				}
			}
			yaExiste := false
			for _, mountItem := range mountMap {
				var mf MountFormat
				mf = mountItem
				if strings.ToLower(mf.Name) == strings.ToLower(name) && nameDisco == mf.DiscoName {
					yaExiste = true
				}
			}
			if yaExiste == false {
				var part_fit byte
				var part_start int32
				var part_s int32
				var part_next int32
				var part_type byte
				f, err := os.OpenFile(path, os.O_RDWR, 0644)
				if err != nil {
					fmt.Println("Ocurrio un error al abrir el archivo")
					return
				}
				defer f.Close()
				var mb Mbr
				startExtendida := 0
				f.Seek(0, 0)
				err = binary.Read(f, binary.LittleEndian, &mb)
				if err != nil {
					fmt.Println("Ocurrio un error al leer el archivo")
					return
				}
				for i := 0; i < 4; i++ {
					if mb.Particiones[i].Part_type == 'E' {
						startExtendida = int(mb.Particiones[i].Part_start)
					}
				}
				var m Mbr
				f.Seek(0, 0)
				err = binary.Read(f, binary.LittleEndian, &m)
				if err != nil {
					fmt.Println("Ocurrio un error al leer el archivo")
					return
				}

				var arrayName [16]byte
				longitud := len(arrayName)
				for i := 0; i < longitud; i++ {
					if i < len(name) {
						arrayName[i] = name[i]
					} else {
						arrayName[i] = '\x00'
					}
				}
				for i := 0; i < 4; i++ {
					if m.Particiones[i].Part_name == arrayName {
						if m.Particiones[i].Part_type == 'E' {
							fmt.Println("No se puede montar una particion extendida.")
						} else {
							m.Particiones[i].Part_status = 1
							part_fit = m.Particiones[i].Part_fit
							part_start = m.Particiones[i].Part_start
							part_s = m.Particiones[i].Part_s
							part_next = -1
							part_type = m.Particiones[i].Part_type
							encontre = true
							f.Seek(0, 0)
							err = binary.Write(f, binary.LittleEndian, &m)
							if err != nil {
								fmt.Println("Ocurrio un error al escribir en el archivo")
								return
							}
						}
						break
					}
				}
				if encontre == false {
					var auxEbr byte
					f.Seek(int64(startExtendida), 0)
					err = binary.Read(f, binary.LittleEndian, &auxEbr)
					if err != nil {
						fmt.Println("Ocurrio un error al leer el archivo")
						return
					}
					for auxEbr != '\x00' {
						var e Ebr
						if auxEbr == '1' {
							err = binary.Read(f, binary.LittleEndian, &e)
							if err != nil {
								fmt.Println("Ocurrio un error al leer el archivo")
								return
							}
							if e.Part_name == arrayName {
								f.Seek(int64(e.Part_start)+int64(unsafe.Sizeof(byte(0))), 0)
								e.Part_status = 1
								part_fit = e.Part_fit
								part_start = e.Part_start
								part_s = e.Part_s
								part_next = e.Part_next
								part_type = 'L'
								err = binary.Write(f, binary.LittleEndian, &e)
								if err != nil {
									fmt.Println("Ocurrio un error al escribir en el archivo")
									return
								}
								encontre = true
								break
							}
							f.Seek(int64(e.Part_next), 0)
							err = binary.Read(f, binary.LittleEndian, &auxEbr)
							if err != nil {
								fmt.Println("Ocurrio un error al leer el archivo")
								return
							}
						}
					}
				}
				if encontre == false {
					fmt.Println("La particion con el nombre: " + name + ", no existe.")
				}

				if encontre == true {
					conta := 0
					idDisco := 0
					encontreDisco := false
					for _, mountItem := range mountMap {
						mf := mountItem
						if mf.Id > conta && nameDisco == mf.DiscoName {
							conta = mf.Id
						}

						if mf.Path == path {
							idDisco = mf.IdDisco
							encontreDisco = true
						}
					}

					if encontreDisco == false {
						for _, mountItem := range mountMap {
							mf := mountItem
							if mf.IdDisco > idDisco {
								idDisco = mf.IdDisco
							}
						}
						idDisco += 1
					}
					conta += 1
					nameMount := "25" + strconv.Itoa(idDisco) + obtenerLetra(conta)
					mountF := MountFormat{
						Id:         conta,
						IdDisco:    idDisco,
						DiscoName:  nameDisco,
						Path:       path,
						Name:       name,
						IdMount:    nameMount,
						Part_next:  int(part_next),
						Part_fit:   part_fit,
						Part_start: int(part_start),
						Part_s:     int(part_s),
						Part_type:  part_type,
						S_mtime:    time.Now(),
					}
					mountMap[nameMount] = mountF
					fmt.Println("Se monto la particion con exito.")
					for _, entry := range mountMap {
						fmt.Printf("{%d, %s, %s}\n", entry.Id, entry.Name, entry.IdMount)
					}
				} else {
					fmt.Println("No se pudo montar la particion.")
				}
			}
		}
	}
}

func mkfs(tokens []string) {
	typeVal := ""
	idVal := ""
	errorVal := false
	for i := 1; i < len(tokens); i++ {
		parameter := strings.Split(tokens[i], "=")
		if parameter[0] == ">id" {
			idVal = parameter[1]
		} else if parameter[0] == ">type" {
			typeVal = parameter[1]
		} else {
			fmt.Println("No se reconoce el comando.")
			errorVal = true
		}
	}
	if idVal == "" {
		fmt.Println("El id es obligatorio.")
		errorVal = true
	}
	if typeVal == "full" || typeVal == "" {
	} else {
		fmt.Println("El tipo debe ser full.")
		errorVal = true
	}
	if errorVal == false {
		var encontre bool = false
		var mf MountFormat
		for idMount, mountItem := range mountMap {
			if idMount == idVal {
				mf = mountItem
				encontre = true
			}
		}
		if encontre == true {
			f, err := os.OpenFile(mf.Path, os.O_RDWR, 0644)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()
			var infoEbr Ebr
			const infoSizeEbr = unsafe.Sizeof(infoEbr)
			var infoInodo Inodo
			const infoSizeInodo = unsafe.Sizeof(infoInodo)
			var infoBloqueArchi BloqueArchivos
			const infoSizeBloqueArchi = unsafe.Sizeof(infoBloqueArchi)
			var infoSuperBloque BloqueArchivos
			const infoSizeSuperBloque = unsafe.Sizeof(infoSuperBloque)
			if mf.Part_type == 'L' {
				_, err = f.Seek(int64(mf.Part_start)+int64(infoSizeEbr), 0)
				if err != nil {
					log.Fatal(err)
				}
				var buff byte = 0
				for i := 0; i < int(mf.Part_s)-int(infoSizeEbr); i++ {
					_, err = f.Write([]byte{buff})
					if err != nil {
						log.Fatal(err)
					}
				}
				n := int(math.Floor(float64(mf.Part_s)-float64(infoSizeEbr)-float64(infoSizeEbr)) / (4 + float64(infoSizeInodo) + (3 * float64(infoSizeBloqueArchi))))
				var sb SuperBloque
				sb.S_filesystem_type = 2
				sb.S_inodes_count = int32(n)
				sb.S_blocks_count = int32(3 * n)
				sb.S_free_blocks_count = int32(3 * n)
				sb.S_free_inodes_count = int32(n)
				t := time.Now()
				tiempo := strconv.Itoa(t.Year()) + "-" + strconv.Itoa(int(t.UTC().Month())) + "-" + strconv.Itoa(t.Day()) + " " + strconv.Itoa(t.Hour()) + ":" + strconv.Itoa(t.Minute())
				var array [16]byte
				longitud := len(array)
				for i := 0; i < longitud; i++ {
					if i < len(tiempo) {
						array[i] = tiempo[i]
					} else {
						array[i] = ' '
					}
				}
				sb.S_mtime = array
				sb.S_mnt_count = 1
				sb.S_magic = 0xEF53
				sb.S_inode_size = int32(infoSizeInodo)
				sb.S_block_size = int32(infoSizeBloqueArchi)
				sb.S_firts_ino = 2
				sb.S_first_blo = 2
				sb.S_bm_inode_start = int32(mf.Part_start) + int32(infoSizeEbr) + int32(infoSizeSuperBloque)
				sb.S_bm_block_start = sb.S_bm_inode_start + int32(n)
				sb.S_inode_start = sb.S_bm_block_start + int32(3*n)
				sb.S_block_start = sb.S_inode_start + int32(n)*int32(infoSizeInodo)
				_, err = f.Seek(int64(mf.Part_start)+int64(infoSizeEbr), 0)
				if err != nil {
					log.Fatal(err)
				}
				err = binary.Write(f, binary.LittleEndian, &sb)
				if err != nil {
					panic(err)
				}

				mbinodos := make([]byte, n)
				mbbloques := make([]byte, 3*n)
				for i := 2; i < n; i++ {
					mbinodos[i] = '0'
				}
				mbinodos[0] = '1'
				mbinodos[1] = '1'
				_, err := f.Seek(int64(mf.Part_start)+int64(infoSizeEbr)+int64(infoSizeSuperBloque), 0)
				if err != nil {
					panic(err)
				}
				err = binary.Write(f, binary.LittleEndian, &mbinodos)
				if err != nil {
					panic(err)
				}

				for i := 2; i < 3*n; i++ {
					mbbloques[i] = '0'
				}
				mbbloques[0] = '1'
				mbbloques[1] = '1'
				_, err = f.Seek(int64(mf.Part_start)+int64(infoSizeEbr)+int64(infoSizeSuperBloque)+int64(n), 0)
				if err != nil {
					panic(err)
				}

				err = binary.Write(f, binary.LittleEndian, &mbbloques)
				if err != nil {
					panic(err)
				}
				var raiz Inodo
				raiz.I_uid = 1
				raiz.I_gid = 1
				raiz.I_size = 0
				t = time.Now()
				tiempo = strconv.Itoa(t.Year()) + "-" + strconv.Itoa(int(t.UTC().Month())) + "-" + strconv.Itoa(t.Day()) + " " + strconv.Itoa(t.Hour()) + ":" + strconv.Itoa(t.Minute())
				longitud = len(array)
				for i := 0; i < longitud; i++ {
					if i < len(tiempo) {
						array[i] = tiempo[i]
					} else {
						array[i] = ' '
					}
				}
				raiz.I_atime = array
				raiz.I_ctime = array
				raiz.I_mtime = array
				for i := 0; i < 15; i++ {
					raiz.I_block[i] = -1
				}
				raiz.I_type = '0'
				raiz.I_perm = 0777
				raiz.I_block[0] = 0
				var bcRoot BloqueCarpetas
				var contenidoR Content

				contenidoR.B_name[0] = '.'
				contenidoR.B_inodo = 0
				bcRoot.B_content[0] = contenidoR
				contenidoR.B_name[0] = '.'
				contenidoR.B_name[1] = '.'
				contenidoR.B_inodo = 0
				bcRoot.B_content[1] = contenidoR
				contenidoR.B_name[0] = 'u'
				contenidoR.B_name[1] = 's'
				contenidoR.B_name[2] = 'e'
				contenidoR.B_name[3] = 'r'
				contenidoR.B_name[4] = 's'
				contenidoR.B_name[5] = '.'
				contenidoR.B_name[6] = 't'
				contenidoR.B_name[7] = 'x'
				contenidoR.B_name[8] = 't'
				contenidoR.B_inodo = 1
				bcRoot.B_content[2] = contenidoR
				var arrayVacio [12]byte
				contenidoR.B_name = arrayVacio
				contenidoR.B_inodo = -1
				bcRoot.B_content[3] = contenidoR
				var auxsuperbloque SuperBloque
				_, err = f.Seek(int64(mf.Part_start)+int64(infoSizeEbr), 0)
				if err != nil {
					panic(err)
				}
				err = binary.Read(f, binary.LittleEndian, &auxsuperbloque)
				if err != nil {
					panic(err)
				}
				_, err = f.Seek(int64(auxsuperbloque.S_inode_start), 0)
				if err != nil {
					panic(err)
				}
				err = binary.Write(f, binary.LittleEndian, &raiz)
				if err != nil {
					panic(err)
				}
				auxsuperbloque.S_free_inodes_count--
				_, err = f.Seek(int64(auxsuperbloque.S_block_start), 0)
				if err != nil {
					panic(err)
				}
				err = binary.Write(f, binary.LittleEndian, &bcRoot)
				if err != nil {
					panic(err)
				}
				auxsuperbloque.S_free_blocks_count--

				archivoUsers := "1,G,root\n1,U,root,root,123\n"
				var archivousuarios Inodo
				archivousuarios.I_gid = 1
				archivousuarios.I_size = int32(len(archivoUsers))
				archivousuarios.I_uid = 1
				t = time.Now()
				tiempo = strconv.Itoa(t.Year()) + "-" + strconv.Itoa(int(t.UTC().Month())) + "-" + strconv.Itoa(t.Day()) + " " + strconv.Itoa(t.Hour()) + ":" + strconv.Itoa(t.Minute())
				longitud = len(array)
				for i := 0; i < longitud; i++ {
					if i < len(tiempo) {
						array[i] = tiempo[i]
					} else {
						array[i] = '\x00'
					}
				}
				archivousuarios.I_atime = array
				archivousuarios.I_ctime = array
				archivousuarios.I_mtime = array
				for i := 0; i < 16; i++ {
					archivousuarios.I_block[i] = -1
				}
				archivousuarios.I_block[0] = 1
				archivousuarios.I_type = '1'
				archivousuarios.I_perm = 0664
				var bloquearchivos BloqueArchivos
				var arrayContent [64]byte
				longitud = len(arrayContent)
				for i := 0; i < longitud; i++ {
					if i < len(archivoUsers) {
						arrayContent[i] = archivoUsers[i]
					} else {
						arrayContent[i] = '\x00'
					}
				}
				bloquearchivos.B_content = arrayContent
				f, _ := os.OpenFile(mf.Path, os.O_WRONLY, 0644)
				defer f.Close()
				_, err = f.Seek(int64(auxsuperbloque.S_inode_start)+int64(infoSizeInodo), 0)
				if err != nil {
					panic(err)
				}
				err = binary.Write(f, binary.LittleEndian, &archivousuarios)
				if err != nil {
					panic(err)
				}
				auxsuperbloque.S_free_inodes_count--
				_, err = f.Seek(int64(auxsuperbloque.S_block_start)+int64(infoSizeBloqueArchi), 0)
				if err != nil {
					panic(err)
				}
				err = binary.Write(f, binary.LittleEndian, &bloquearchivos)
				if err != nil {
					panic(err)
				}
				auxsuperbloque.S_free_blocks_count--
				_, err = f.Seek(int64(mf.Part_start)+int64(infoSizeEbr), 0)
				if err != nil {
					panic(err)
				}
				err = binary.Write(f, binary.LittleEndian, &auxsuperbloque)
				if err != nil {
					panic(err)
				}
				fmt.Println("Particion formateada con exito.")
			}
		}
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

func obtenerLetra(num int) string {
	if num == 1 {
		return "A"
	} else if num == 2 {
		return "B"
	} else if num == 3 {
		return "C"
	} else if num == 4 {
		return "D"
	} else if num == 5 {
		return "E"
	} else if num == 6 {
		return "F"
	} else if num == 7 {
		return "G"
	} else if num == 8 {
		return "H"
	} else if num == 9 {
		return "I"
	} else if num == 10 {
		return "J"
	} else if num == 11 {
		return "K"
	} else if num == 12 {
		return "L"
	} else if num == 13 {
		return "M"
	} else if num == 14 {
		return "N"
	} else if num == 15 {
		return "O"
	} else if num == 16 {
		return "P"
	} else if num == 17 {
		return "Q"
	} else if num == 18 {
		return "R"
	} else if num == 19 {
		return "S"
	} else if num == 20 {
		return "T"
	} else if num == 21 {
		return "U"
	} else if num == 22 {
		return "V"
	} else if num == 23 {
		return "W"
	} else if num == 24 {
		return "X"
	} else if num == 25 {
		return "Y"
	} else if num == 26 {
		return "Z"
	} else {
		return "XX"
	}
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
