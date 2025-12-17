import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router, RouterModule } from '@angular/router';
import type { HttpErrorResponse } from '@angular/common/http';
import type { AuthService } from '../auth.service';
import type { ErrorHandlerService } from '../error-handler.service';

@Component({
  selector: 'app-signup',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule],
  templateUrl: './signup.component.html',
  styleUrl: './signup.component.css'
})
export class SignupComponent {
  name = '';
  email = '';
  password = '';
  confirmPassword = '';
  errorMessage = '';
  isLoading = false;

  constructor(
    private readonly authService: AuthService,
    private readonly router: Router,
    private readonly errorHandler: ErrorHandlerService
  ) {}

  onSubmit(): void {
    if (!this.email || !this.password || !this.confirmPassword) {
      this.errorMessage = 'Please fill in all required fields';
      return;
    }

    if (this.password !== this.confirmPassword) {
      this.errorMessage = 'Passwords do not match';
      return;
    }

    if (this.password.length < 6) {
      this.errorMessage = 'Password must be at least 6 characters';
      return;
    }

    this.isLoading = true;
    this.errorMessage = '';

    const signupData = {
      email: this.email,
      password: this.password,
      ...(this.name && { username: this.name })
    };

    this.authService.signup(signupData)
      .subscribe({
        next: () => {
          this.isLoading = false;
          this.router.navigate(['/dashboard']);
        },
        error: (error: HttpErrorResponse) => {
          this.isLoading = false;
          const errorDetails = this.errorHandler.handleError(error);

          // Special handling for EMAIL_EXISTS error
          if (errorDetails.code === 'EMAIL_EXISTS') {
            this.errorMessage = 'This email is already registered. Try logging in instead or use a different email.';
          } else {
            this.errorMessage = errorDetails.userMessage;
          }

          this.errorHandler.logError(errorDetails);
        }
      });
  }
}
