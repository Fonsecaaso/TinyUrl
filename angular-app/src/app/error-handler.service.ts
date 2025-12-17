import { Injectable } from '@angular/core';
import { HttpErrorResponse } from '@angular/common/http';

export interface ErrorDetails {
  message: string;
  userMessage: string;
  code?: string;
  status?: number;
}

@Injectable({
  providedIn: 'root'
})
export class ErrorHandlerService {

  /**
   * Handles HTTP errors and returns user-friendly error details
   */
  handleError(error: HttpErrorResponse): ErrorDetails {
    // Network or client-side error
    if (error.error instanceof ErrorEvent) {
      return {
        message: error.error.message,
        userMessage: 'Network error. Please check your connection and try again.',
        status: 0
      };
    }

    // Backend returned error response
    const errorCode = error.error?.code;
    const errorMessage = error.error?.error;
    const status = error.status;

    // Handle specific error codes
    switch (errorCode) {
      case 'INVALID_CREDENTIALS':
        return {
          message: errorMessage || 'Invalid credentials',
          userMessage: 'Invalid email or password. Please try again.',
          code: errorCode,
          status
        };

      case 'EMAIL_EXISTS':
        return {
          message: errorMessage || 'Email already exists',
          userMessage: 'This email is already registered. Try logging in instead.',
          code: errorCode,
          status
        };

      case 'INVALID_URL':
        return {
          message: errorMessage || 'Invalid URL format',
          userMessage: 'Please enter a valid URL (e.g., https://example.com)',
          code: errorCode,
          status
        };

      case 'INVALID_SHORT_URL':
        return {
          message: errorMessage || 'Invalid short URL',
          userMessage: 'The short URL you entered is not valid.',
          code: errorCode,
          status
        };

      case 'URL_NOT_FOUND':
        return {
          message: errorMessage || 'URL not found',
          userMessage: 'This short URL does not exist or has been deleted.',
          code: errorCode,
          status
        };

      case 'INTERNAL_ERROR':
        return {
          message: errorMessage || 'Internal server error',
          userMessage: 'Something went wrong on our end. Please try again later.',
          code: errorCode,
          status
        };

      default:
        // Handle by HTTP status code if no specific error code
        return this.handleHttpStatus(status, errorMessage);
    }
  }

  /**
   * Handles errors based on HTTP status codes
   */
  private handleHttpStatus(status: number, message?: string): ErrorDetails {
    switch (status) {
      case 400:
        return {
          message: message || 'Bad request',
          userMessage: message || 'Invalid request. Please check your input.',
          status
        };

      case 401:
        return {
          message: message || 'Unauthorized',
          userMessage: 'Your session has expired. Please log in again.',
          status
        };

      case 403:
        return {
          message: message || 'Forbidden',
          userMessage: 'You do not have permission to perform this action.',
          status
        };

      case 404:
        return {
          message: message || 'Not found',
          userMessage: 'The requested resource was not found.',
          status
        };

      case 409:
        return {
          message: message || 'Conflict',
          userMessage: message || 'A conflict occurred. Please try again.',
          status
        };

      case 422:
        return {
          message: message || 'Unprocessable entity',
          userMessage: message || 'Invalid data provided. Please check your input.',
          status
        };

      case 429:
        return {
          message: message || 'Too many requests',
          userMessage: 'Too many requests. Please wait a moment and try again.',
          status
        };

      case 500:
      case 502:
      case 503:
      case 504:
        return {
          message: message || 'Server error',
          userMessage: 'Server error. Please try again later.',
          status
        };

      default:
        return {
          message: message || 'Unknown error',
          userMessage: 'An unexpected error occurred. Please try again.',
          status
        };
    }
  }

  /**
   * Logs error to console (can be extended to send to logging service)
   */
  logError(error: ErrorDetails): void {
    console.error('Error occurred:', error);
    // TODO: Send to logging service (e.g., Sentry, LogRocket)
  }
}
