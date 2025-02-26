import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

interface ApiResponse {
  "chave gerada": string; // Ajuste conforme a resposta da API
}

interface ResolveResponse {
  value: string; // A API retorna um JSON com um campo "value"
}

@Injectable({
  providedIn: 'root'
})
export class UrlShortenerService {
  private apiUrl = 'http://localhost:8080/'; // URL da sua API

  constructor(private http: HttpClient) {}

  shortenUrl(url: string): Observable<ApiResponse> {
    return this.http.post<ApiResponse>(`${this.apiUrl}${url}`, {}); // Chamada para API
  }
}
