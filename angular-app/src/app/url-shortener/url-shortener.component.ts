import { Component } from '@angular/core';
import { NgIf } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { UrlShortenerService } from '../url-shortener.service';

@Component({
  selector: 'app-url-shortener',
  standalone: true,
  templateUrl: './url-shortener.component.html',
  styleUrls: ['./url-shortener.component.css'],
  imports: [NgIf, FormsModule],
})
export class UrlShortenerComponent {
  urlInput: string = '';
  shortenedUrl: string = '';
  isLoading: boolean = false;
  errorMessage: string = '';
  hasError: boolean = false;
  isCopied: boolean = false;

  constructor(private urlShortenerService: UrlShortenerService) {}

  shorten() {
    if (!this.validateUrl()) return;
    
    this.isLoading = true;
    this.clearError();
  
    this.urlShortenerService.shortenUrl(this.urlInput).subscribe({
      next: (response) => {
        this.shortenedUrl = response["short_code"];
        this.isLoading = false;
        this.isCopied = false;
      },
      error: (err) => {
        console.error('Erro ao encurtar URL', err);
        this.showError('Erro ao encurtar a URL. Tente novamente.');
        this.isLoading = false;
      },
    });
  }

  validateUrl(): boolean {
    const urlInput = this.urlInput.trim();
    
    if (!urlInput) {
      this.showError('Por favor, insira uma URL.');
      return false;
    }

    // Verifica se é uma URL válida (aceita com ou sem protocolo)
    const urlPattern = /^(https?:\/\/)?([\w-]+\.)+[\w-]+(\/[^\s]*)?$/;
    if (!urlPattern.test(urlInput)) {
      this.showError('Por favor, insira uma URL válida (ex: google.com)');
      return false;
    }

    // Adiciona http:// se não tiver protocolo
    if (!urlInput.startsWith('http://') && !urlInput.startsWith('https://')) {
      this.urlInput = 'http://' + urlInput;
    }

    return true;
  }

  showError(message: string) {
    this.errorMessage = message;
    this.hasError = true;
  }

  clearError() {
    this.errorMessage = '';
    this.hasError = false;
  }

  getFullShortenedUrl(): string {
    return `localhost:4200/${this.shortenedUrl}`;
  }

  getTruncatedUrl(url: string): string {
    return url.length > 50 ? url.substring(0, 50) + '...' : url;
  }

  async copyToClipboard(text: string) {
    try {
      await navigator.clipboard.writeText(text);
      this.isCopied = true;
      
      // Reset após 3 segundos
      setTimeout(() => {
        this.isCopied = false;
      }, 3000);
    } catch (err) {
      console.error('Erro ao copiar para a área de transferência', err);
      // Fallback para navegadores mais antigos
      const textArea = document.createElement('textarea');
      textArea.value = text;
      document.body.appendChild(textArea);
      textArea.select();
      document.execCommand('copy');
      document.body.removeChild(textArea);
      this.isCopied = true;
      
      setTimeout(() => {
        this.isCopied = false;
      }, 3000);
    }
  }
}
