import { Component } from '@angular/core';
import { NgIf } from '@angular/common'; // Importando *ngIf
import { FormsModule } from '@angular/forms'; // Importando ngModel
import { UrlShortenerService } from '../url-shortener.service';

@Component({
  selector: 'app-url-shortener',
  standalone: true,
  templateUrl: './url-shortener.component.html',
  styleUrls: ['./url-shortener.component.css'],
  imports: [NgIf, FormsModule], // Importando diretivas necessárias
})
export class UrlShortenerComponent {
  urlInput: string = '';
  shortenedUrl: string = '';

  constructor(private urlShortenerService: UrlShortenerService) {}

  shorten() {
    if (!this.urlInput.trim()) return;
  
    this.urlShortenerService.shortenUrl(this.urlInput).subscribe({
      next: (response) => {
        this.shortenedUrl = response["chave gerada"]; // Pegando a chave da resposta da API
      },
      error: (err) => console.error('Erro ao encurtar URL', err),
    });
  }
}
