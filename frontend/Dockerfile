# syntax=docker/dockerfile:1
FROM node:24.16.0-alpine3.22 AS build
WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY index.html tsconfig.json tsconfig.app.json tsconfig.node.json vite.config.ts ./
COPY src ./src

ARG VITE_API_URL
ENV VITE_API_URL=${VITE_API_URL}
RUN npm run build

FROM nginx:1.28.0-alpine
COPY --from=build /app/dist /usr/share/nginx/html

EXPOSE 80
