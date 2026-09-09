# Frontend Docker Compose

The local stack now includes the React/Vite frontend alongside the existing Go
and Node APIs:

```text
Browser :5173 → frontend/nginx :80 → Go API :8080 → Node API :8080
```

Run the complete stack from the repository root:

```bash
docker compose up --build
```

The services are published locally at:

| Service | Host address | Container address |
| --- | --- | --- |
| Frontend | `http://localhost:5173` | nginx `:80` |
| Go API | `http://localhost:8080` | `:8080` |
| Node API | `http://localhost:3000` | `:8080` |

## Frontend image and API URL

`frontend/Dockerfile` uses a multi-stage build. Node installs the locked
dependencies and runs `npm run build`; nginx serves only the resulting
`dist/` files. The image does not run the Vite development server and does not
include source files or `node_modules` in its runtime stage.

The browser URL is a build-time value, not a Docker-network URL:

```yaml
frontend:
  build:
    context: ./frontend
    args:
      VITE_API_URL: "http://localhost:8080"
```

`VITE_API_URL` must be reachable by the browser. Do not set it to
`http://go-api:8080`; that hostname exists only inside the Compose network.

## CORS configuration

Go reads `CORS_ALLOWED_ORIGINS` at startup and applies it to the `/qr` HTTP
handler. Compose sets the local frontend origin explicitly:

```text
CORS_ALLOWED_ORIGINS=http://localhost:5173
```

The middleware allows only `POST` and `OPTIONS` with the `Content-Type` header.
Preflight and QR responses from the configured origin receive the exact
`Access-Control-Allow-Origin` value; other origins do not receive that header.
No wildcard origin or credentials are enabled. For another frontend host,
provide its comma-separated origins through the Go environment.

The existing API contract is unchanged:

```http
POST http://localhost:8080/qr
Content-Type: application/json
```

```json
{"matrix":[[1,2],[3,4],[5,6]]}
```

Go still reaches Node using the internal Compose address
`NODE_API_URL=http://node-api:8080`; the browser never uses that address.

## Local checks

Run frontend checks from `frontend/`:

```bash
npm run lint
npm run build
```

Run Go checks from `go-api/`:

```bash
go vet ./...
go test ./...
go build ./...
```

Validate the resolved Compose configuration before starting the stack:

```bash
docker compose config
docker compose build
```

`depends_on` expresses startup order only; it is not a readiness check. The
frontend footer is static UI content and requires no additional route or nginx
configuration.
