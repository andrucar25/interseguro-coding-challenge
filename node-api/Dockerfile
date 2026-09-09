# syntax=docker/dockerfile:1
FROM node:24.16.0-alpine3.22 AS build
WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY tsconfig.json ./
COPY src ./src
RUN npm run build

FROM node:24.16.0-alpine3.22 AS production-dependencies
WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci --omit=dev && npm cache clean --force

FROM node:24.16.0-alpine3.22 AS runtime
WORKDIR /app

COPY --from=production-dependencies --chown=node:node /app/node_modules ./node_modules
COPY --from=build --chown=node:node /app/dist ./dist

EXPOSE 8080
USER node
CMD ["node", "dist/index.js"]
