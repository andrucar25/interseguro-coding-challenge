# QA Report - frontend-docker-compose

## Result

**PASS** — the implementation conforms to the approved design for the local
frontend Compose service, nginx static runtime, restrictive Go CORS support,
and footer. No defects were found during this review.

## Scope reviewed

- `frontend/Dockerfile` and `frontend/.dockerignore`
- Footer and minimal layout changes in `frontend/src/App.tsx` and
  `frontend/src/styles.css`
- `docker-compose.yaml`
- Go CORS configuration and HTTP CORS tests
- Existing backend behavior for QR and Go-to-Node statistics, as covered by
  the Go suite and live Compose request

The changed files match the approved design. The frontend image is a
multi-stage build: its nginx runtime copies only `/app/dist`, and the Compose
frontend uses the browser-reachable build-time URL `http://localhost:8080`.
The Go service retains the internal `http://node-api:8080` URL and accepts only
the explicit local browser origin in Compose.

## Automated results

| Check | Result |
| --- | --- |
| `frontend: npm run lint` | Pass |
| `frontend: npm run build` (`tsc -b` + Vite) | Pass |
| `go-api: gofmt` check | Pass |
| `go-api: go vet ./...` | Pass |
| `go-api: go test ./...` | Pass (`httpapi`, `qr`, and `statistics`) |
| `go-api: go build ./...` | Pass |
| `docker compose config` | Pass; resolves the three intended services, ports, default network, internal Node URL, frontend build argument, and CORS origin |
| `docker compose build` | Pass; Go, Node, and frontend images build successfully |
| `docker compose ps` | Pass; all three services were running on frontend `5173`, Go `8080`, and Node `3000` |
| Live frontend request | Pass; nginx returned the built application containing the exact requested footer text |
| Live allowed CORS preflight | Pass; `OPTIONS /qr` returned `204`, exact `Access-Control-Allow-Origin: http://localhost:5173`, `POST, OPTIONS`, and `Content-Type` |
| Live allowed QR request | Pass; `POST /qr` returned `200`, the exact allowed-origin header, QR output, and statistics through Go to Node |
| Live untrusted-origin request | Pass; `POST /qr` returned no `Access-Control-Allow-Origin` header |

## Relevant coverage and prioritized cases

The adapted Go HTTP tests cover the highest-risk CORS boundary cases:

1. Allowed browser preflight for `POST /qr`, including requested JSON header.
2. Allowed-origin QR POST returns the exact allow-origin response header.
3. Untrusted-origin QR POST does not receive an allow-origin header.
4. Existing QR success, invalid-input, and downstream-error tests remain
   passing after the application constructor change.

The live Compose request additionally verified the deployed topology:

```text
host browser origin -> Go :8080 -> node-api :8080
```

## Regression checklist

- [x] Node and Go Compose ports/configuration remain unchanged.
- [x] Go continues to resolve Node by Docker-internal service name.
- [x] Frontend is served by nginx on host port `5173`, not Vite dev server.
- [x] Frontend production build is type-checked and lint-clean.
- [x] CORS is explicit, method/header constrained, and has no wildcard or
      credentials behavior.
- [x] Existing QR validation and downstream error mappings remain covered and
      passing.
- [x] Footer text is exact, including accented characters, and has no added
      link/icon/application logic.

## Caveat

No browser automation is available in this environment. The served static
application and exact footer text were verified over HTTP, but visual layout,
form interaction states, duplicate-submission prevention, and the temporary
Go-outage UI error state were not exercised in a real browser during this QA
run. These are appropriate final manual smoke checks before demonstration.
