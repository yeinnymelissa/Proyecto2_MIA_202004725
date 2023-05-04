import { NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';
import { FormsModule } from '@angular/forms';

import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { NavbarComponent } from './componentes/navbar/navbar.component';
import { CargaComponent } from './componentes/carga/carga.component';
import { LoginComponent } from './componentes/login/login.component';
import { HttpClientModule} from '@angular/common/http';
import { DashboardComponent } from './componentes/dashboard/dashboard.component';
import { MessagesModule } from 'primeng/messages';
import { MessageModule } from 'primeng/message';
import { NavbardashComponent } from './componentes/navbardash/navbardash.component';
import { InicioComponent } from './componentes/inicio/inicio.component';
import { ReportesComponent } from './componentes/reportes/reportes.component';

@NgModule({
  declarations: [
    AppComponent,
    NavbarComponent,
    CargaComponent,
    LoginComponent,
    DashboardComponent,
    NavbardashComponent,
    InicioComponent,
    ReportesComponent
  ],
  imports: [
    BrowserModule,
    AppRoutingModule,
    FormsModule,
    HttpClientModule,
    MessagesModule,
    MessageModule
  ],
  providers: [],
  bootstrap: [AppComponent]
})
export class AppModule { }
