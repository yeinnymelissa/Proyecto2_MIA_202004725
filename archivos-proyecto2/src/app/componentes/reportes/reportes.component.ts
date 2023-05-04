import { Component } from '@angular/core';
import { graphviz }  from 'd3-graphviz';
import { GeneralService } from 'src/app/general.service';

@Component({
  selector: 'app-reportes',
  templateUrl: './reportes.component.html',
  styleUrls: ['./reportes.component.css']
})
export class ReportesComponent {

  constructor(private servicio: GeneralService){
    
  }

  reporteDisk(){
    let id:any = localStorage.getItem('idParti')
    let datos = {  
      Consola: "rep >id="+id+" >Path=/home/yeinny/Documentos/MIA/Disco4.jpg >name=disk"
      };
      
    let stringifiedData = JSON.stringify(datos);
    this.servicio.mandarComando(stringifiedData).subscribe(
      (response:any) =>{
        console.log(response)

        graphviz('#disk').renderDot(response.reporte);
      }
    )
  }

  reporteSuper(){
    let id:any = localStorage.getItem('idParti')
    let datos = {  
      Consola: "rep >id="+id+" >Path=/home/yeinny/Documentos/MIA/Disco4.jpg >name=sb"
      };
      
    let stringifiedData = JSON.stringify(datos);
    this.servicio.mandarComando(stringifiedData).subscribe(
      (response:any) =>{
        console.log(response)

        graphviz('#super').renderDot(response.reporte);
      }
    )
  }
}
