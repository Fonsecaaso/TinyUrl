# TinyUrl

URL shortener developed with **Go** (backend), **Angular** (frontend), **Redis** (cache), and **PostgreSQL** (database). The service offers fast redirects, scalability, and high performance with minimal latency.

🌐 **Live production**: [fonsecaaso.com](http://fonsecaaso.com)

### Production Screenshot
<img width="1952" height="1394" alt="image" src="https://github.com/user-attachments/assets/4fa12fbe-9a1d-46dd-9aa8-628f689cdb4c" />


## 🚀 How to Run

### Requirements
- Docker
- Docker Compose

### Running the Project

To run the project locally:

```bash
docker-compose up --build
```

The application will be available at: `http://localhost:4200`

## 🏗️ Architecture

The architecture consists of:
- **Frontend**: Angular application
- **Load Balancer**: Nginx Gateway
- **Backend**: 2 Go servers (scalable)
- **Cache**: Redis
- **Database**: PostgreSQL

![image](https://github.com/user-attachments/assets/24835408-6913-4130-a013-3a02f004b895)

## 📦 Manual Deployment

### Backend (Go Server)

1. Build the image:
```bash
cd go-server
docker build --platform linux/x86_64 -t tiny-url .
```

2. Authenticate with AWS ECR:
```bash
aws ecr get-login-password --region us-east-1 --profile personal-account | \
  docker login --username AWS --password-stdin 173941740239.dkr.ecr.us-east-1.amazonaws.com
```

3. Tag and push the image:
```bash
docker tag tiny-url:latest 173941740239.dkr.ecr.us-east-1.amazonaws.com/tiny-url:latest
docker push 173941740239.dkr.ecr.us-east-1.amazonaws.com/tiny-url:latest
```

**Note**: Ensure that the Application Load Balancer, Task Definition, and Target Group are configured in AWS ECS before creating the service.

### Frontend (Angular)

1. Build the image:
```bash
cd angular-app
docker build --platform linux/x86_64 -t tiny-url-frontend .
```

2. Authenticate with AWS ECR:
```bash
aws ecr get-login-password --region us-east-1 --profile personal-account | \
  docker login --username AWS --password-stdin 173941740239.dkr.ecr.us-east-1.amazonaws.com
```

3. Tag and push the image:
```bash
docker tag tiny-url-frontend:latest 173941740239.dkr.ecr.us-east-1.amazonaws.com/tiny-url-frontend:latest
docker push 173941740239.dkr.ecr.us-east-1.amazonaws.com/tiny-url-frontend:latest
```



## 🛣️ Project Roadmap

### ✅ Phase 1: MVP - Basic Features

- ✅ Go backend with basic CRUD operations
- ✅ Angular frontend with shortening form
- ✅ Automatic redirect for shortened URLs
- ✅ Redis integration for caching
- ✅ Orchestration via Docker Compose
- ✅ Rate limiting for API protection
- ✅ Production deployment on AWS (available at [fonsecaaso.com](http://fonsecaaso.com))

### 🔄 Phase 2: CI/CD

- [ ] Continuous integration pipeline
- [ ] Automated tests (unit and integration)
- [ ] Automated production deployment
- [ ] Automatic release versioning
- [ ] Automated rollback on failures

### 📊 Phase 3: Observability

**Prometheus + Grafana**:
- [ ] CPU and memory consumption metrics
- [ ] API response time metrics (p50, p95, p99 percentiles)

**Elasticsearch + Kibana**:
- [ ] URL access statistics
- [ ] Frequency and geographic analysis

**OpenTelemetry**:
- [ ] Distributed tracing for bottleneck identification

### 🚀 Phase 4: New Features

**Authentication and Management**:
- [ ] Authentication system (login and signup)
- [ ] JWT authentication in frontend
- [ ] User dashboard with URL history

**Analytics and Customization**:
- [ ] URL usage analytics (clicks, geographic origin, devices)
- [ ] Custom URLs by user
- [ ] Single-use URLs
- [ ] Configurable URL expiration
- [ ] Automatic cleanup based on:
  - URLs without access for 48h
  - Daily removal of the 30% least accessed URLs


