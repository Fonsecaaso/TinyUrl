import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { map } from 'rxjs/operators';
import { environment } from '../environments/environment';

export interface ApiResponse {
  "short_code": string;
}

export interface ResolveResponse {
  value: string;
}

export interface UserURL {
  id: string;
  url: string;
  created_at: string;
  user_id?: string;
}

export interface UserURLsResponse {
  message: string;
  urls: UserURL[];
}

@Injectable({
  providedIn: 'root'
})
export class UrlShortenerService {
  private apiUrl = environment.apiBaseUrl;

  constructor(private http: HttpClient) {}

  shortenUrl(url: string): Observable<ApiResponse> {
    return this.http.post<ApiResponse>(`${this.apiUrl}/`, { url });
  }

  getUserUrls(): Observable<UserURL[]> {
    return this.http.get<UserURLsResponse>(`${this.apiUrl}/user/urls`).pipe(
      map(response => response.urls)
    );
  }
}