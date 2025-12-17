import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router, RouterModule, ActivatedRoute } from '@angular/router';
import type { HttpErrorResponse } from '@angular/common/http';
import type { AuthService } from '../auth.service';
import type { ErrorHandlerService } from '../error-handler.service';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule],
  templateUrl: './login.component.html',
  styleUrl: './login.component.css'
})
export class LoginComponent implements OnInit {
  email: string = '';
  password: string = '';
  errorMessage: string = '';
  infoMessage: string = '';
  isLoading: boolean = false;

  constructor(
    private authService: AuthService,
    private router: Router,
    private route: ActivatedRoute,
    private errorHandler: ErrorHandlerService
  ) {}

  ngOnInit(): void {
    // Check for query parameters to show info messages
    this.route.queryParams.subscribe(params => {
      if (params['reason'] === 'session_expired') {
        this.infoMessage = 'Your session has expired. Please log in again.';
      } else if (params['reason'] === 'unauthorized') {
        this.infoMessage = 'Your session has expired. Please log in again.';
      }
    });
  }

  onSubmit(): void {
    if (!this.email || !this.password) {
      this.errorMessage = 'Please fill in all fields';
      return;
    }

    this.isLoading = true;
    this.errorMessage = '';
    this.infoMessage = '';

    this.authService.login({ email: this.email, password: this.password })
      .subscribe({
        next: () => {
          this.isLoading = false;
          this.router.navigate(['/dashboard']);
        },
        error: (error: HttpErrorResponse) => {
          this.isLoading = false;
          const errorDetails = this.errorHandler.handleError(error);
          this.errorMessage = errorDetails.userMessage;
          this.errorHandler.logError(errorDetails);
        }
      });
  }
}
