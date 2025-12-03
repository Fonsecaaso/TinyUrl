import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { AuthService } from '../auth.service';
import { UrlShortenerService, UserURL } from '../url-shortener.service';
import { environment } from '../../environments/environment';

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './dashboard.component.html',
  styleUrl: './dashboard.component.css'
})
export class DashboardComponent implements OnInit {
  urls: UserURL[] = [];
  newUrl: string = '';
  isLoading: boolean = false;
  isCreating: boolean = false;
  errorMessage: string = '';
  successMessage: string = '';
  newlyCreatedUrl: string = ''; // Store the newly created short URL
  newlyCreatedOriginal: string = ''; // Store the original URL
  showNewUrlResult: boolean = false;
  isCopied: boolean = false;
  appDomain = environment.appDomain;

  constructor(
    private authService: AuthService,
    private urlService: UrlShortenerService,
    private router: Router
  ) {}

  ngOnInit(): void {
    this.loadUserUrls();
  }

  loadUserUrls(): void {
    this.isLoading = true;
    this.errorMessage = '';

    this.urlService.getUserUrls()
      .subscribe({
        next: (urls) => {
          this.urls = urls;
          this.isLoading = false;
        },
        error: (error) => {
          this.isLoading = false;
          this.errorMessage = 'Error loading your URLs';
          console.error('Error loading URLs:', error);
        }
      });
  }

  createShortUrl(): void {
    if (!this.newUrl) {
      this.errorMessage = 'Please enter a URL';
      return;
    }

    // Basic URL validation
    const urlPattern = /^(https?:\/\/)?([\da-z\.-]+)\.([a-z\.]{2,6})([\/\w \.-]*)*\/?$/;
    if (!urlPattern.test(this.newUrl)) {
      this.errorMessage = 'Please enter a valid URL';
      return;
    }

    // Add http:// if it doesn't have a protocol
    let urlToShorten = this.newUrl;
    if (!urlToShorten.startsWith('http://') && !urlToShorten.startsWith('https://')) {
      urlToShorten = 'http://' + urlToShorten;
    }

    this.isCreating = true;
    this.errorMessage = '';
    this.showNewUrlResult = false;

    this.urlService.shortenUrl(urlToShorten)
      .subscribe({
        next: (response) => {
          this.isCreating = false;
          this.newlyCreatedUrl = response.short_code;
          this.newlyCreatedOriginal = urlToShorten;
          this.showNewUrlResult = true;
          this.isCopied = false;
          this.newUrl = '';
          this.loadUserUrls(); // Reload the list
        },
        error: (error) => {
          this.isCreating = false;
          this.errorMessage = error.error?.error || 'Error creating short link';
        }
      });
  }

  getNewShortUrl(): string {
    return `https://${this.appDomain}/${this.newlyCreatedUrl}`;
  }

  getTruncatedUrl(url: string): string {
    return url.length > 50 ? url.substring(0, 50) + '...' : url;
  }

  copyNewUrlToClipboard(text: string): void {
    if (navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(text).then(() => {
        this.isCopied = true;
        setTimeout(() => {
          this.isCopied = false;
        }, 3000);
      });
    } else {
      // Fallback for older browsers
      const textarea = document.createElement('textarea');
      textarea.value = text;
      textarea.style.position = 'fixed';
      textarea.style.opacity = '0';
      document.body.appendChild(textarea);
      textarea.select();
      document.execCommand('copy');
      document.body.removeChild(textarea);
      this.isCopied = true;
      setTimeout(() => {
        this.isCopied = false;
      }, 3000);
    }
  }

  getShortUrl(shortCode: string): string {
    return `https://${this.appDomain}/${shortCode}`;
  }

  copyToClipboard(shortCode: string): void {
    const url = this.getShortUrl(shortCode);

    if (navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(url).then(() => {
        this.showCopyFeedback(shortCode);
      });
    } else {
      // Fallback for older browsers
      const textarea = document.createElement('textarea');
      textarea.value = url;
      textarea.style.position = 'fixed';
      textarea.style.opacity = '0';
      document.body.appendChild(textarea);
      textarea.select();
      document.execCommand('copy');
      document.body.removeChild(textarea);
      this.showCopyFeedback(shortCode);
    }
  }

  private showCopyFeedback(shortCode: string): void {
    const element = document.getElementById(`copy-btn-${shortCode}`);
    if (element) {
      const originalText = element.textContent;
      element.textContent = 'Copied!';
      setTimeout(() => {
        element.textContent = originalText;
      }, 2000);
    }
  }

  formatDate(dateString: string): string {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: 'numeric',
      minute: '2-digit',
      hour12: true
    });
  }

  logout(): void {
    this.authService.logout();
    this.router.navigate(['/']);
  }
}
