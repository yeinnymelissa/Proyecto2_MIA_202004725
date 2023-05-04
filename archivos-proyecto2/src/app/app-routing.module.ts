import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { LoginComponent } from './componentes/login/login.component';
import { CargaComponent } from './componentes/carga/carga.component';
import { DashboardComponent } from './componentes/dashboard/dashboard.component';
import { InicioComponent } from './componentes/inicio/inicio.component';
import { ReportesComponent } from './componentes/reportes/reportes.component';

const routes: Routes = [
  { path: '', component: InicioComponent }, 
  { path: 'login', component: LoginComponent},
  { path: 'dashboard', component: DashboardComponent},
  { path: 'reportes', component: ReportesComponent}
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
