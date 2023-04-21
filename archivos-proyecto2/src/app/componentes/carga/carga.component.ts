import { Component } from '@angular/core';
import { Router } from '@angular/router';

@Component({
  selector: 'app-carga',
  templateUrl: './carga.component.html',
  styleUrls: ['./carga.component.css']
})
export class CargaComponent {

  comandos:any;
  fileName:any;

  constructor(private router:Router){
    
  }

  onFileSelected(event:any) {

    const file:File = event.target.files[0];

    if (file) {
      let reader = new FileReader()
      reader.onload = function(event){
        let contenido = event.target?.result
        const elemento: any = document.getElementById("editorComandos")
        if(elemento != null){
          elemento.innerText = contenido
        }
      }
      reader.readAsText(file)
    }
  }

}
