# TinyUrl

URL shortener developed with **Go** (backend), **Angular** (frontend), **Redis** (cache), and **PostgreSQL** (database). The service offers fast redirects, scalability, and high performance with minimal latency.

🌐 **Live production**: [fonsecaaso.com](http://fonsecaaso.com)

### Production Screenshot
 <img src="https://github.com/user-attachments/assets/183ef8b4-bb7d-40ee-8b4e-9748324fdec6" width="800" />


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

