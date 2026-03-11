FROM node:22-bookworm-slim AS base
WORKDIR /app

COPY package.json ./
COPY package-lock.json* pnpm-lock.yaml* yarn.lock* ./

RUN npm ci

COPY . .

EXPOSE 3000

CMD ["sh", "-c", "node server.js"]
