# TinyUrl

Encurtador de URLs desenvolvido com **Go** (backend), **Angular** (frontend), **Redis** (cache) e **PostgreSQL** (banco de dados). O serviço oferece redirecionamentos rápidos, escalabilidade e alta performance com latência mínima.

🌐 **Acesse em produção**: [fonsecaaso.com](http://fonsecaaso.com)

## 🚀 Como Executar

### Requisitos
- Docker
- Docker Compose

### Executando o Projeto

Para rodar o projeto localmente:

```bash
docker-compose up --build
```

A aplicação estará disponível em: `http://localhost:4200`

## 🏗️ Arquitetura

A arquitetura consiste em:
- **Frontend**: Aplicação Angular
- **Load Balancer**: Nginx Gateway
- **Backend**: 2 servidores Go (escaláveis)
- **Cache**: Redis
- **Banco de Dados**: PostgreSQL

![image](https://github.com/user-attachments/assets/24835408-6913-4130-a013-3a02f004b895)

## 📦 Deploy Manual

### Backend (Go Server)

1. Build da imagem:
```bash
cd go-server
docker build --platform linux/x86_64 -t tiny-url .
```

2. Autenticação no AWS ECR:
```bash
aws ecr get-login-password --region us-east-1 --profile personal-account | \
  docker login --username AWS --password-stdin 173941740239.dkr.ecr.us-east-1.amazonaws.com
```

3. Tag e push da imagem:
```bash
docker tag tiny-url:latest 173941740239.dkr.ecr.us-east-1.amazonaws.com/tiny-url:latest
docker push 173941740239.dkr.ecr.us-east-1.amazonaws.com/tiny-url:latest
```

**Nota**: Certifique-se de que o Application Load Balancer, Task Definition e Target Group estão configurados no AWS ECS antes de criar o serviço.

### Frontend (Angular)

1. Build da imagem:
```bash
cd angular-app
docker build --platform linux/x86_64 -t tiny-url-frontend .
```

2. Autenticação no AWS ECR:
```bash
aws ecr get-login-password --region us-east-1 --profile personal-account | \
  docker login --username AWS --password-stdin 173941740239.dkr.ecr.us-east-1.amazonaws.com
```

3. Tag e push da imagem:
```bash
docker tag tiny-url-frontend:latest 173941740239.dkr.ecr.us-east-1.amazonaws.com/tiny-url-frontend:latest
docker push 173941740239.dkr.ecr.us-east-1.amazonaws.com/tiny-url-frontend:latest
```



## 🛣️ Roadmap do Projeto

### ✅ Fase 1: MVP - Funcionalidades Básicas

- ✅ Backend em Go com operações básicas de CRUD
- ✅ Frontend Angular com formulário de encurtamento
- ✅ Redirecionamento automático de URLs encurtadas
- ✅ Integração com Redis para cache
- ✅ Orquestração via Docker Compose
- ✅ Rate limiting para proteção da API
- ✅ Deploy em produção na AWS (disponível em [fonsecaaso.com](http://fonsecaaso.com))

### 🔄 Fase 2: CI/CD

- [ ] Pipeline de integração contínua
- [ ] Testes automatizados (unitários e integração)
- [ ] Deploy automatizado para produção
- [ ] Versionamento automático de releases
- [ ] Rollback automatizado em caso de falhas

### 📊 Fase 3: Observabilidade

**Prometheus + Grafana**:
- [ ] Métricas de consumo de CPU e memória
- [ ] Tempo de resposta da API (percentis p50, p95, p99)

**Elasticsearch + Kibana**:
- [ ] Estatísticas de acessos às URLs
- [ ] Análise de frequência e geografia

**OpenTelemetry**:
- [ ] Tracing distribuído para identificação de gargalos

### 🚀 Fase 4: Novas Features

**Autenticação e Gerenciamento**:
- [ ] Sistema de autenticação (login e cadastro)
- [ ] Autenticação JWT no frontend
- [ ] Dashboard do usuário com histórico de URLs

**Analytics e Personalização**:
- [ ] Analytics de uso das URLs (cliques, origem geográfica, dispositivos)
- [ ] URLs personalizadas pelo usuário
- [ ] URLs de uso único (single-use URLs)
- [ ] Expiração de URLs configurável
- [ ] Limpeza automática baseada em:
  - URLs sem acesso por 48h
  - Remoção das 30% URLs menos acessadas (diariamente)


