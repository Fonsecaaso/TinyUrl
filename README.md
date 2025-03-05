# TinyUrl

Este encurtador de URL é um serviço rápido e confiável desenvolvido com Go para back-end, Angular para front-end e Redis para cache de alto desempenho. Ele permite que os usuários encurtem URLs longos, tornando-os mais fáceis de compartilhar e gerenciar. O aplicativo fornece redirecionamentos rápidos e escalabilidade, garantindo latência mínima e desempenho ideal.

## Como Executar

Para rodar o projeto utilizando o Docker, execute o seguinte comando:

```bash
docker-compose up --build
```

## Arquitetura Atual

O código atual consiste em uma aplicação em angular, um server em go, um cache redis e um banco postgres.

![image](https://github.com/user-attachments/assets/beb05f45-9caa-402b-ba46-37d4f24c2193)


## Fases do projeto

### ✅ 1. Aplicações golang e angular com funcionalidades básicas

- Um server backend feito em angular com as funcionalidades de inserir e ler tupla no redis.
- E uma aplicação em angular com as funcionalidades de encurtar url via formulário e com a url retornada da criação redirecionar para o endereço inicial. Ambas aplicações orquestradas via docker-compose.


### 2. Autenticação, url personalizada, expiração de urls (default, personalizada e inteligente - de acordo com # de acessos)

- Teremos criação de usuários, com login e senha, token jwt no frontend
- Na página do usuário ele pode personalizar a url encurtada, pode listar as urls que já criou
- Setar um tempo de expiração das urls no redis, permitir que o usuário configure esse tempo, e de acordo com a frequência de acessos às urls determinar quais urls serão deletadas (i.e, urls com mais de 48h sem acesso são deletadas ou as 30% urls menos acessadas são deletadas a cada 24h)

### 3. Observabilidade

- Prometeus e Grafana: Dashboard com métricas de consumo de mémoria e CPU, gráfico com tempo de resposta da api com percentis
- ElasticSearch e Kibana: Estatística de acessoas às urls, frequência, geografia
- OpenTelemetry: Tracing para determinar gargalos no fluxo do código

