import { Component } from '@angular/core';
import { Router } from '@angular/router';
import { HttpClient, HttpClientModule } from '@angular/common/http';
import { GeneralService } from 'src/app/general.service';
import { delay} from "rxjs/operators";

@Component({
  selector: 'app-carga',
  templateUrl: './carga.component.html',
  styleUrls: ['./carga.component.css']
})
export class CargaComponent {

  comandos:any;
  fileName:any;
  consola:String = ""
  comandosTmp: any[]

  constructor(private router:Router, private http: HttpClient, private servicio: GeneralService){
    this.comandosTmp = []
  }

  onFileSelected(event:any) {

    const file:File = event.target.files[0];

    if (file) {
      let reader = new FileReader()
      reader.onload = function(event){
        let contenido = event.target?.result
        const elemento: any = document.getElementById("editorComandos")
        if(elemento != null){
          elemento.value = contenido
          console.log(elemento.value)
        }
      }
      reader.readAsText(file)
    }
  }

  ejecutar(){
    this.comandosTmp = []
    const elemento: any = document.getElementById("editorComandos")
    if(elemento != null){
      this.consola = ""
      this.comandos = elemento.value
      let datos = {  
        Consola: this.comandos
        };
        
      let stringifiedData = JSON.stringify(datos);
      this.servicio.mandarComando(stringifiedData).subscribe(
        (response:any) =>{
          this.consola += response.consola
          console.log(this.consola)
          const consol: any = document.getElementById("consola")
          consol.value = this.consola
        }
      )

    }
  }

}
