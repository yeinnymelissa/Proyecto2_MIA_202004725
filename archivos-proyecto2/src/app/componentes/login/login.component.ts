import { Component } from '@angular/core';
import { GeneralService } from 'src/app/general.service';
import { MessageService } from 'primeng/api';
import { Router } from '@angular/router';
import swal from'sweetalert2';

@Component({
  selector: 'app-login',
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.css'],
  providers: [MessageService]
})
export class LoginComponent {

  user: any
  pwd: any
  idParti: any
  constructor(private servicio: GeneralService, private messageService: MessageService, private router:Router){
  }

  login(){
    const formu:any = document.getElementById("formu")
    let valido = formu.reportValidity()
    if(valido){
      let datos = {  
      Consola: "login >user="+this.user+" >pwd="+this.pwd+" >id="+this.idParti
      };
      
    let stringifiedData = JSON.stringify(datos);
    this.servicio.mandarComando(stringifiedData).subscribe(
      (response:any) =>{
        console.log(response)
        if(response.login == true  && response.user.confi == true){
          localStorage.setItem('user', JSON.stringify(this.user));
          localStorage.setItem('pwd', JSON.stringify(this.pwd));
          localStorage.setItem('idParti', JSON.stringify(this.idParti));
          this.router.navigate(['/dashboard']);
          swal.fire({
            title: 'Inicio de sesión exitoso',
            text: 'Bienvenido '+ this.user,
            icon: 'success',
          })
        }else{
          swal.fire({
            title: 'No se pudo iniciar sesión',
            text: response.consola ,
            icon: 'error',
          })
        }
      }
    )
    }
  }
}
