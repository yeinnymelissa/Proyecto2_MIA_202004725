import { Injectable } from '@angular/core';
import { Observable, lastValueFrom } from 'rxjs';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { catchError } from 'rxjs/operators';


@Injectable({
  providedIn: 'root'
})
export class GeneralService {

  constructor(private http:HttpClient) { }

  mandarComando(comando:any){
    let httpOptions = {
      headers: new HttpHeaders({
        'Content-type':'application/json'
      })
    }
    return this.http.post<any>("/read", comando, httpOptions).pipe(
      catchError(e => {console.log(e)
        return ""
      })
    );
  }

  public async mandarComandoEsperar(comando: any) : Promise<any>{
    let httpOptions = {
      headers: new HttpHeaders({
        'Content-type':'application/json'
      })
    }
    let res = await lastValueFrom(this.http.post<any>("/read", comando))
    return res
  }

}
