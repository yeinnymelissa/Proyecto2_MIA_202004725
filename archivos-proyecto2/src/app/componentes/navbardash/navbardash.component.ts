import { Component } from '@angular/core';
import { Router } from '@angular/router';
import { OnInit } from '@angular/core';
import { GeneralService } from 'src/app/general.service';
import swal from'sweetalert2';

@Component({
  selector: 'app-navbardash',
  templateUrl: './navbardash.component.html',
  styleUrls: ['./navbardash.component.css']
})
export class NavbardashComponent implements OnInit{
  constructor(private router:Router, private servicio: GeneralService){
    
  }
  
  ngOnInit(): void {
    
  }

  logout(){
    let datos = {  
      Consola: "logout"
    };
    let stringifiedData = JSON.stringify(datos);
    this.servicio.mandarComando(stringifiedData).subscribe(
      (response:any) =>{
        let usuario:any = localStorage.getItem('user')
        
        if(usuario){
          localStorage.clear()
          this.router.navigate(['/']);
          swal.fire({
            title: response.consola,
            icon: 'success',
          })
        }else{
          this.router.navigate(['/']);
          swal.fire({
            title: response.consola,
            icon: 'error',
          })
        }
      }
    )

  }

  principal(){
    this.router.navigate(['/dashboard']);
  }
}
