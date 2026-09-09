# Interseguro Coding Challenge

This technical challenge accepts a matrix, calculates its QR factorization, and returns statistics for the resulting matrices.

## Challenge coverage

- Go/Fiber entry API and Node.js/Express statistics API.
- Economy QR factorization for rectangular matrices (`rows >= columns`).
- Synchronous Go-to-Node HTTP communication: Go sends `Q` and `R`; Node calculates maximum, minimum, average, total sum, and diagonal-matrix detection.
- Input validation and error handling.
- Unit tests and in-process HTTP endpoint tests for the main logic and API boundaries.
- Docker for both backends and Docker Compose for the frontend, Go, and Node services.
- GCP infrastructure with Terraform, Artifact Registry, and Cloud Run Services.
- Optional React, Vite, and TypeScript frontend deployed through Firebase Hosting.
- CORS configured through environment variables.

The challenge statement refers inconsistently to *matrix rotation* and *QR factorization*. This solution uses QR factorization as an informed technical decision because it is the challenge's explicit functional requirement.

## Architecture

```text
Browser
   ↓
React + Vite
Firebase Hosting
   ↓
Go + Fiber
Cloud Run
   ↓ HTTP
Node.js + Express
Cloud Run
```

Go receives the matrix and performs QR factorization, then sends `Q` and `R` to Node. Node calculates the statistics, and Go returns the consolidated response to the frontend.

## Technologies and decisions

- Go + Fiber is the public API and orchestration layer.
- Node.js + Express is responsible for statistics.
- The APIs communicate through synchronous HTTP.
- Docker images are independent per service, and Compose runs the complete local flow.
- Configuration is environment-based rather than embedded in code.
- Cloud Run Services are used instead of Cloud Run Functions, with Terraform for infrastructure as code.

## Deployment

The application is deployed on Google Cloud. The public frontend is available at:

<https://interseguro-challenge-508112.web.app>

Retrieve the current Go and Node service URLs from Terraform state:

```bash
cd infrastructure
terraform output -raw go_service_url
terraform output -raw node_service_url
```

Both Cloud Run Services are public during the initial end-to-end validation phase. Restricting Node with IAM is a possible later hardening measure, not an outstanding challenge requirement.

## Local execution

```bash
docker compose up --build
```

- Frontend: <http://localhost:5173>
- Go: <http://localhost:8080>
- Node: <http://localhost:3000>

You can also verify the API directly:

```bash
curl -X POST http://localhost:8080/qr \
  -H 'Content-Type: application/json' \
  -d '{"matrix":[[1,2],[3,4]]}'
```

## Environment configuration

The same application runs in different environments by changing external configuration:

| Component | Variables |
| --- | --- |
| Go | `PORT`, `NODE_API_URL`, `CORS_ALLOWED_ORIGINS` |
| Node | `PORT` |
| Frontend | `VITE_API_URL` |

## Quality

The repository contains unit tests plus in-process HTTP tests for the main endpoints. It does not currently include a full cross-service integration test that runs the Go API against the real Node API.

```bash
cd node-api
npm run typecheck && npm run check && npm test

cd ../go-api
go vet ./... && go test ./... && go build ./...
```

## Optional challenge features

Implemented: frontend; unit and in-process HTTP endpoint tests.

Not implemented: JWT authentication (optional in the challenge).
