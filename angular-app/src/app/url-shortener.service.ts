import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../environments/environment';

interface ApiResponse {
  "short_code": string; // Ajuste conforme a resposta da API
}

interface ResolveResponse {
  value: string; // A API retorna um JSON com um campo "value"
}

@Injectable({
  providedIn: 'root'
})
export class UrlShortenerService {
  private apiUrl = environment.apiBaseUrl;

  constructor(private http: HttpClient) {}

  shortenUrl(url: string): Observable<ApiResponse> {
    return this.http.post<ApiResponse>(this.apiUrl, { url }); // Envia a URL no body
  }
}