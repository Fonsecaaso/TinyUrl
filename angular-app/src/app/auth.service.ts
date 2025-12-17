import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { BehaviorSubject, Observable, tap, interval, Subscription } from 'rxjs';
import { Router } from '@angular/router';
import { environment } from '../environments/environment';

export interface LoginRequest {
  email: string;
  password: string;
}

export interface SignupRequest {
  email: string;
  password: string;
  username?: string;
}

export interface AuthResponse {
  token: string;
}

export interface User {
  id: string;
  email: string;
  name?: string;
}

interface TokenPayload {
  user_id: string;
  email?: string;
  name?: string;
  exp?: number;
  iat?: number;
}

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private readonly TOKEN_KEY = 'auth_token';
  private readonly API_URL = environment.apiBaseUrl;
  private readonly TOKEN_CHECK_INTERVAL = 60000; // Check every 60 seconds

  private currentUserSubject = new BehaviorSubject<User | null>(null);
  public currentUser$ = this.currentUserSubject.asObservable();

  private tokenExpirationTimer?: Subscription;

  constructor(
    private http: HttpClient,
    private router: Router
  ) {
    // Check if user is already logged in on service initialization
    this.checkAuthStatus();
    this.startTokenExpirationCheck();
  }

  private checkAuthStatus(): void {
    const token = this.getToken();
    if (token && this.isTokenValid(token)) {
      this.currentUserSubject.next(this.decodeToken(token));
    } else if (token) {
      // Token exists but is expired
      this.handleExpiredToken();
    }
  }

  /**
   * Starts periodic check for token expiration
   */
  private startTokenExpirationCheck(): void {
    // Check token expiration every minute
    this.tokenExpirationTimer = interval(this.TOKEN_CHECK_INTERVAL).subscribe(() => {
      const token = this.getToken();
      if (token && !this.isTokenValid(token)) {
        this.handleExpiredToken();
      }
    });
  }

  /**
   * Handles expired token by logging out user and redirecting to login
   */
  private handleExpiredToken(): void {
    console.warn('Token expired. Logging out user...');
    this.logout();
    this.router.navigate(['/login'], {
      queryParams: { reason: 'session_expired' }
    });
  }

  /**
   * Checks if token is valid (not expired)
   */
  private isTokenValid(token: string): boolean {
    try {
      const payload = this.decodeTokenPayload(token);
      if (!payload) return false;

      const exp = payload.exp;
      if (exp) {
        const expirationDate = new Date(exp * 1000);
        const now = new Date();
        // Add 5 second buffer to account for clock skew
        return expirationDate.getTime() > now.getTime() + 5000;
      }
      return true;
    } catch {
      return false;
    }
  }

  /**
   * Decodes token and returns payload
   */
  private decodeTokenPayload(token: string): TokenPayload | null {
    try {
      const payload = JSON.parse(atob(token.split('.')[1]));
      return payload as TokenPayload;
    } catch {
      return null;
    }
  }

  signup(request: SignupRequest): Observable<AuthResponse> {
    return this.http.post<AuthResponse>(`${this.API_URL}/signup`, request)
      .pipe(
        tap(response => {
          this.setToken(response.token);
          this.currentUserSubject.next(this.decodeToken(response.token));
        })
      );
  }

  login(request: LoginRequest): Observable<AuthResponse> {
    return this.http.post<AuthResponse>(`${this.API_URL}/login`, request)
      .pipe(
        tap(response => {
          this.setToken(response.token);
          this.currentUserSubject.next(this.decodeToken(response.token));
        })
      );
  }

  logout(): void {
    this.removeToken();
    this.currentUserSubject.next(null);

    // Clean up timer
    if (this.tokenExpirationTimer) {
      this.tokenExpirationTimer.unsubscribe();
    }
  }

  isAuthenticated(): boolean {
    const token = this.getToken();
    if (!token) {
      return false;
    }

    return this.isTokenValid(token);
  }

  getToken(): string | null {
    return localStorage.getItem(this.TOKEN_KEY);
  }

  private setToken(token: string): void {
    localStorage.setItem(this.TOKEN_KEY, token);
  }

  private removeToken(): void {
    localStorage.removeItem(this.TOKEN_KEY);
  }

  private decodeToken(token: string): User | null {
    try {
      const payload = JSON.parse(atob(token.split('.')[1]));
      return {
        id: payload.user_id || '',
        email: payload.email || '',
        name: payload.name || ''
      };
    } catch {
      return null;
    }
  }

  getCurrentUser(): User | null {
    return this.currentUserSubject.value;
  }
}
