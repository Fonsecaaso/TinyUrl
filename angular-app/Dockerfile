# Usa Node.js para rodar o Angular em desenvolvimento
FROM node:20

# Define o diretório de trabalho dentro do container
WORKDIR /app

# Copia os arquivos de dependências primeiro para otimizar o cache
COPY package.json package-lock.json ./

# Instala as dependências do Angular
RUN npm install

# Copia todo o código da aplicação
COPY . .

# Expõe a porta 4200 para acesso ao Angular
EXPOSE 4200

# Comando padrão para iniciar o servidor de desenvolvimento do Angular
CMD ["ng", "serve", "--host", "0.0.0.0", "--port", "4200"]
