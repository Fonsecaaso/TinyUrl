# Build Angular
FROM node:20 AS builder

WORKDIR /app

COPY package*.json ./
RUN npm install

COPY . .
RUN npm run build

# Serve com NGINX
FROM nginx:alpine

# Copia o build correto
COPY --from=builder /app/dist/angular-app/browser /usr/share/nginx/html

# Sobrescreve a config padrão do NGINX com seu arquivo
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
