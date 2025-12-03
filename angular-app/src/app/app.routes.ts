import { Routes } from '@angular/router';
import { provideRouter } from '@angular/router';
import { UrlShortenerComponent } from './url-shortener/url-shortener.component'; // Página principal
import { SearchComponent } from './search/search.component';
import { LoginComponent } from './login/login.component';
import { SignupComponent } from './signup/signup.component';
import { DashboardComponent } from './dashboard/dashboard.component';
import { authGuard } from './auth.guard';

export const routes: Routes = [
  { path: '', component: UrlShortenerComponent },
  { path: 'login', component: LoginComponent },
  { path: 'signup', component: SignupComponent },
  { path: 'dashboard', component: DashboardComponent, canActivate: [authGuard] },
  { path: ':id', component: SearchComponent },
];

export const appRouter = provideRouter(routes); // Configuração do roteamento
