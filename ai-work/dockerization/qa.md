# QA Report — dockerization

## Result

**Pass with an environment limitation.** The implementation matches the approved
design on static review and all available application quality gates pass. Docker
image builds and container smoke tests could not run because the configured Docker
daemon is unavailable in this environment.

## Scope reviewed

- `go-api/Dockerfile`, `.dockerignore`, and `.env.example`
- `node-api/Dockerfile`, `.dockerignore`, `.env.example`, and `src/index.ts`
- Existing Go and Node tests relevant to the unchanged API behavior

Static review confirms:

- Go uses the approved multi-stage Linux build, static binary, distroless
  `nonroot` runtime, port `8080`, and no baked runtime configuration.
- Node uses the approved build, production-dependency, and non-root runtime
  stages; only production dependencies and compiled output are copied to runtime.
- Both build contexts exclude real `.env` files, build output, local dependencies,
  VCS data, and test artifacts as designed. The documented `.env.example` files
  remain included.
- Node defaults to port `8080` and explicitly binds `0.0.0.0`. Go continues to
  default to `8080` and rejects an absent or invalid `NODE_API_URL` before serving.
- No endpoint, request/response, QR, statistics, or Go-to-Node contract change was
  introduced.

## Automated validation

| Area | Command | Result |
| --- | --- | --- |
| Go formatting | `test -z "$(find . -type f -name '*.go' -exec gofmt -l {} +)"` | Pass |
| Go static analysis | `go vet ./...` | Pass |
| Go tests | `go test ./...` | Pass: `httpapi`, `qr`, and `statistics`; root has no tests |
| Go build | `go build ./...` | Pass |
| Node type check | `npm run typecheck` | Pass |
| Node lint/format | `npm run check` | Pass |
| Node tests | `npm test` | Pass: 2 files, 31 tests |
| Node build | `npm run build` | Pass |

The first sandboxed Node test run failed because Supertest could not bind its
temporary `0.0.0.0` listener (`EPERM`). Re-running with host-level permission
passed all 31 tests; this is an environment restriction, not an application
failure.

## Docker validation limitation

`docker version` could not connect to the configured OrbStack socket:

```text
failed to connect to the docker API at unix:///Users/andres/.orbstack/run/docker.sock:
connect: no such file or directory
```

Consequently, the following required validations remain unexecuted:

- `docker build --platform linux/amd64 -t interseguro-go-api ./go-api`
- `docker build --platform linux/amd64 -t interseguro-node-api ./node-api`
- Node standalone container smoke test
- Go container against local Node smoke test
- Two-container Docker-network `/qr` smoke test
- Go startup failure check with missing/invalid `NODE_API_URL`

## Findings

No implementation defect found. The only blocker is the unavailable Docker daemon,
which prevents runtime image verification in this environment.

## Prioritized follow-up tests

1. With Docker running, build both images for `linux/amd64` and confirm their
   final processes run as the configured non-root users.
2. Start Node with `PORT=8080`; POST valid matrices to
   `/api/v1/statistics` and verify the existing statistics response.
3. Run Go with `NODE_API_URL=http://host.docker.internal:3000` against a local
   Node instance; POST `/qr` and verify the existing end-to-end result.
4. Run both images on an isolated Docker network using
   `NODE_API_URL=http://node-api:8080`; repeat the `/qr` request.
5. Start the Go image without `NODE_API_URL` and with an invalid value; verify
   that it exits before listening and logs the established configuration error.

## Regression checklist

- [x] Go formatting, vetting, unit/integration tests, and build pass.
- [x] Node type check, Biome check, HTTP/statistics tests, and build pass.
- [x] Existing HTTP paths and Go-to-Node statistics path remain unchanged.
- [x] Runtime ports and environment-variable behavior match the approved design.
- [x] No real `.env` file is included in either Docker build context by rule.
- [ ] Docker image build and runtime smoke tests (blocked by unavailable daemon).

## Coverage

No coverage command or threshold is configured. Existing validation covers Go QR,
HTTP-handler, and statistics-client packages, plus 31 Node statistics-route and
validation tests. Container layers, runtime ownership, platform manifests, and
inter-container networking require the deferred Docker smoke tests above.
