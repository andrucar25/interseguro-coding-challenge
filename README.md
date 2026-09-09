# Interseguro Coding Challenge

Two small APIs process a matrix request end to end:

1. The Go/Fiber API factorizes the input matrix into **Q** and **R**.
2. It sends both matrices to the Node/Express API.
3. The Node service calculates aggregate statistics and returns them with the QR result.

## Requirements

- Docker and Docker Compose (recommended), or
- Node.js 24 and Go 1.27 for local development.

## Run with Docker

From the repository root:

```bash
docker compose up --build
```

This command builds and starts the complete application: the React frontend, the
Go/Fiber QR API, and the Node/Express statistics API.

To try the application, copy this URL into your browser:

```
http://localhost:5173
```

The public Go API is available at `http://localhost:8080`. The Node statistics API is published at `http://localhost:3000` for local inspection; the Go service reaches it through the Compose network.

Stop the services with `Ctrl+C`, or run `docker compose down` from another terminal.

## Run locally

Open two terminals.

```bash
# Terminal 1: statistics service
cd node-api
npm ci
cp .env.example .env
npm run dev
```

```bash
# Terminal 2: public QR API
cd go-api
cp .env.example .env
set -a && source .env && set +a
go run .
```

The Node API listens on port `3000` and the Go API listens on port `8080` by default. Configure them with `PORT`; set `NODE_API_URL` in `go-api/.env` to the Node service base URL.

## Use the API

Send a non-empty, rectangular matrix with no more columns than rows:

```bash
curl -X POST http://localhost:8080/qr \
  -H 'Content-Type: application/json' \
  -d '{"matrix":[[1,2],[3,4]]}'
```

The response contains `q`, `r`, and `statistics`. Statistics include `maximum`, `minimum`, `sum`, `average`, and `hasDiagonalMatrix` across both result matrices.

## Development checks

```bash
cd node-api
npm run typecheck && npm run check && npm test

cd ../go-api
go vet ./... && go test ./... && go build ./...
```

## Deployment

The application was deployed to Google Cloud Platform (GCP) using Terraform.
