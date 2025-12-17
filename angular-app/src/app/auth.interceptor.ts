import type { HttpInterceptorFn, HttpErrorResponse } from '@angular/common/http';
import { inject } from '@angular/core';
import { Router } from '@angular/router';
import { catchError, throwError } from 'rxjs';
import { AuthService } from './auth.service';

export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const authService = inject(AuthService);
  const router = inject(Router);
  const token = authService.getToken();

  // Clone the request and add the authorization header if token exists
  const clonedRequest = token
    ? req.clone({
        headers: req.headers.set('Authorization', `Bearer ${token}`)
      })
    : req;

  return next(clonedRequest).pipe(
    catchError((error: HttpErrorResponse) => {
      // Handle 401 Unauthorized errors globally
      if (error.status === 401) {
        // Don't redirect if already on login/signup pages or if it's a login/signup request
        const isAuthPage = req.url.includes('/login') || req.url.includes('/signup');
        const currentUrl = router.url;
        const isOnAuthPage = currentUrl === '/login' || currentUrl === '/signup';

        if (!isAuthPage && !isOnAuthPage) {
          console.warn('401 Unauthorized - Token expired or invalid');
          authService.logout();
          router.navigate(['/login'], {
            queryParams: { reason: 'unauthorized' }
          });
        }
      }

      return throwError(() => error);
    })
  );
};
