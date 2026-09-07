# syntax=docker/dockerfile:1

FROM node:20-alpine

WORKDIR /app

COPY --chown=node:node package.json package-lock.json* ./
RUN npm install --production=false

COPY --chown=node:node . .
RUN npm run build

RUN npm prune --production

# CIS Docker Benchmark 4.1 準拠: プリセットの非 root ユーザで実行
USER node

EXPOSE 8002

CMD ["node", "dist/index.js"]
