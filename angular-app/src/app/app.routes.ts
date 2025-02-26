import { Routes } from '@angular/router';
import { provideRouter } from '@angular/router';
import { UrlShortenerComponent } from './url-shortener/url-shortener.component'; // Página principal
import { SearchComponent } from './search/search.component';

export const routes: Routes = [
  { path: '', component: UrlShortenerComponent },
  { path: ':id', component: SearchComponent },
];

export const appRouter = provideRouter(routes); // Configuração do roteamento
