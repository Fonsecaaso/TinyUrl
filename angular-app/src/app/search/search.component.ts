import { Component, inject, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { HttpClient } from '@angular/common/http';

@Component({
  selector: 'app-search',
  standalone: true,
  template: `<p>Redirecionando...</p>`, // Exibe um aviso enquanto redireciona
})
export class SearchComponent implements OnInit {
  private route = inject(ActivatedRoute);
  private http = inject(HttpClient);

  ngOnInit() {
    const id = this.route.snapshot.paramMap.get('id'); // Captura o ID da URL
    if (id) {
      this.http.get<{ value: string }>(`http://localhost:80/${id}`).subscribe({
        next: (response) => {
          let url = response.value;
          
          // Se a URL não começar com http ou https, adicionamos https://
          if (!/^https?:\/\//i.test(url)) {
            url = `https://${url}`;
          }

          // Redireciona para a URL correta
          window.location.href = url;
        },
        error: () => console.error('Erro ao buscar os dados')
      });
    }
  }
}
