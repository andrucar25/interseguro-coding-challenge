# Interseguro Coding Challenge

Web application that calculates the QR factorization of a matrix and the
statistics of the resulting matrices. The project consists of a public Go API,
an internal Node.js statistics service, and a React frontend.

## Architecture

```text
Client → Frontend (React + Vite) → Go/Fiber → Node/Express
                                   QR         statistics
```

1. The client sends a matrix to the Go API at `POST /qr`.
2. Go validates the matrix and calculates its economy QR factorization (`rows >= columns`).
3. Go signs a JWT and sends the `Q` and `R` matrices to the Node service.
4. Node validates the JWT, calculates the statistics, and returns them to Go.
5. Go responds to the client with `Q`, `R`, and their aggregated statistics.

The challenge statement inconsistently refers to matrix rotation and QR
factorization. This implementation uses QR factorization, the explicit
functional requirement.

## Deployment

The application is deployed on Google Cloud in `southamerica-west1`:

| Component | Hosting | Details |
| --- | --- | --- |
| Frontend | Firebase Hosting | React and Vite static build. |
| Go API | Cloud Run | Public QR API, deployed as the `interseguro-go-api` service. |
| Node API | Cloud Run | Statistics API, deployed as the `interseguro-node-api` service. |
| Backend images | Artifact Registry | Docker images for the Go and Node services are stored in the `interseguro-challenge` repository. |
| Infrastructure | Terraform | Provisions Artifact Registry, Cloud Run services, runtime configuration, and service URLs. |

Terraform defines the Google Cloud resources in `infrastructure/`. It receives
explicitly tagged Go and Node image references from Artifact Registry and
deploys them to Cloud Run. The frontend is built separately and served from
Firebase Hosting.

The current Terraform configuration exposes both Cloud Run services publicly.
The Node endpoint is still protected by the service-to-service JWT described
below; clients should call the Go API rather than Node directly.

## Service-to-service security

JWT protects only the **Go → Node** communication; it is not a login or user /
frontend authentication mechanism.

- Go issues one **HS256** JWT for every request to Node.
- Tokens are valid for **one minute** and include `iss`, `aud`, `iat`, and `exp`.
- Node requires exactly one `Authorization: Bearer <token>` header on
  `POST /api/v1/statistics`.
- Node responds with `401 {"error":"unauthorized"}` for missing, malformed,
  expired, incorrectly signed, non-HS256, wrong-issuer, or wrong-audience tokens.
- Both services must share the same secret. Secrets are supplied through
  environment variables and must never be committed.

## Run locally

### Requirement

Docker and Docker Compose.

1. Create a `.env` file in the repository root; do not add it to Git:

   ```dotenv
   JWT_SECRET=use-a-long-random-local-secret
   # Optional; these are the defaults:
   JWT_ISSUER=interseguro-go-api
   JWT_AUDIENCE=interseguro-node-api
   ```

2. Start all services:

   ```bash
   docker compose up --build
   ```

3. Open <http://localhost:5173>. The backend services are also exposed at:

   - Go API: <http://localhost:8080>
   - Node API: <http://localhost:3000>

To stop the environment, run `docker compose down`.

## Public API

### `POST /qr`

Calculates QR and the statistics for `Q` and `R`.

```bash
curl -X POST http://localhost:8080/qr \
  -H 'Content-Type: application/json' \
  -d '{"matrix":[[1,2],[3,4]]}'
```

The body must contain a non-empty, rectangular matrix of finite numbers, with
at least as many rows as columns.

Successful response (`200`, illustrative values):

```json
{
  "q": [[0.0]],
  "r": [[0.0]],
  "statistics": {
    "maximum": 0.0,
    "minimum": 0.0,
    "sum": 0.0,
    "average": 0.0,
    "hasDiagonalMatrix": true
  }
}
```

The values of `q` and `r` depend on the input matrix. Statistics aggregate all
values in both matrices. `hasDiagonalMatrix` is `true` when either `Q` or `R`
is diagonal.

Main errors:

| Status | Code | When it occurs |
| --- | --- | --- |
| `400` | `invalid_request` | The matrix is invalid, non-rectangular, empty, contains non-finite values, or has more columns than rows. |
| `502` | `downstream_error` | The statistics service could not provide a response. |
| `504` | `downstream_timeout` | The statistics service did not respond in time. |

The statistics endpoint (`POST /api/v1/statistics`) is internal and requires a
service JWT. Clients should use `/qr`.

## Configuration

| Service | Variables |
| --- | --- |
| Go | `PORT`, `NODE_API_URL`, `NODE_API_JWT_SECRET`, `NODE_API_JWT_ISSUER`, `NODE_API_JWT_AUDIENCE`, `CORS_ALLOWED_ORIGINS` |
| Node | `PORT`, `JWT_SECRET`, `JWT_ISSUER`, `JWT_AUDIENCE` |
| Frontend | `VITE_API_URL` |

Docker Compose reads `JWT_SECRET` and maps it to `NODE_API_JWT_SECRET` in Go
and `JWT_SECRET` in Node. It maps `JWT_ISSUER` and `JWT_AUDIENCE` in the same
way; when omitted, they default to `interseguro-go-api` and
`interseguro-node-api`.

To run the services without Docker, use `go-api/.env.example` and
`node-api/.env.example` as references. Use the same secret, issuer, and
audience in both services.

## Validation

Run checks from their respective project directories:

```bash
cd node-api
npm run typecheck
npm run check
npm test

cd ../go-api
go vet ./...
go test ./...
go build ./...
```

The repository includes unit and in-process HTTP tests for the main boundaries.
JWT coverage includes missing, malformed, expired, incorrectly signed,
wrong-issuer, and wrong-audience tokens.

## Infrastructure

`infrastructure/` contains the Terraform configuration for the Google Cloud
deployment. Deployment configuration keeps sensitive values in Terraform or
environment variables, never in source code.
