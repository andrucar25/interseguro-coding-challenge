# QA - service-to-service-jwt

## Scope

Reviewed the approved design, Go JWT signer and statistics client, Node JWT configuration and route middleware, tests, environment examples, Docker Compose, Cloud Run Terraform configuration, and README. Product source and configuration were not modified by QA.

Terraform commands were **not executed**: no `terraform fmt`, `terraform validate`, `terraform init`, or `terraform apply` was run, per the user's explicit instruction.

## Automated validation

| Area | Command | Result |
| --- | --- | --- |
| Go formatting | `test -z "$(find . -type f -name '*.go' -exec gofmt -l {} +)"` | Passed |
| Go static analysis | `go vet ./...` | Passed |
| Go tests | `go test ./...` | Passed |
| Go build | `go build ./...` | Passed |
| Node type check | `npm run typecheck` | Passed |
| Node formatting/lint | `npm run check` | Passed |
| Node tests | `npm test` | Passed: 4 files, 52 tests |
| Node build | `npm run build` | Passed |
| Diff whitespace | `git diff --check` | Passed |

The first sandboxed Node test attempt could not bind Supertest's ephemeral local listener (`EPERM`); the same suite was rerun outside that restriction and passed. This was an execution-environment limitation, not a test failure.

Docker Compose and Terraform were inspected only; no containers or Terraform commands were run.

## Requirements and security review

- Go generates a fresh HS256 JWT immediately before every Node statistics request and sends it only as `Authorization: Bearer <token>`.
- The token contains `iss`, `aud`, `iat`, and `exp`; its fixed lifetime is one minute.
- Node applies middleware only to `POST /api/v1/statistics`, before JSON parsing and statistics execution.
- Bearer parsing requires exactly one header and an exact non-empty `Bearer <token>` form.
- `jose.jwtVerify` pins `HS256`, verifies the signature, issuer, audience, and expiration; an additional check requires a finite `exp` claim. This rejects unsigned/`alg=none` tokens.
- All authentication failures use `401` with the safe body `{ "error": "unauthorized" }`; no verification details or token values are returned or logged.
- Go preserves its existing downstream non-2xx mapping, so Node `401` remains a downstream failure rather than a frontend JWT response.
- Secrets are externally configured. Docker Compose maps one required `JWT_SECRET` into both services; examples contain only placeholders. Terraform's `jwt_secret` is marked `sensitive = true` and is absent from `terraform.tfvars.example`.
- The public frontend-to-Go flow remains JWT-free. No user authentication, roles, refresh tokens, sessions, IAM, or database behavior was added.

## Coverage reviewed

Node endpoint tests cover no Authorization header, empty/malformed or wrong-scheme Bearer credentials, invalid signature, expired token, incorrect issuer, incorrect audience, and a valid token continuing to the existing statistics response. Existing validation tests now supply valid service credentials and remain covered.

Go tests verify HS256 signing, signature verification, issuer, audience, `iat`, short `exp`, invalid signer configuration, outbound Authorization-header creation, unchanged `{ q, r }` payload, successful statistics decoding, and existing downstream/timeout behavior.

## Regression checklist

- [x] QR and statistics success contract preserved.
- [x] Node validation and error behavior remains reachable after valid authentication.
- [x] Go-to-Node HTTP request still uses the existing endpoint, body, timeout, response decoding, and downstream error category.
- [x] Node startup and Go startup reject missing/blank JWT configuration through their construction paths.
- [x] Documentation and local Compose configuration describe the service-to-service-only boundary.
- [x] No real secret found in source, README, Docker Compose, examples, or Terraform example variables.

## Findings

No blocking, high, medium, or low severity defects found in the reviewed feature.

## Recommended local smoke test

Before deployment, provide an untracked local `JWT_SECRET` (or shell variable), run `docker compose up --build`, submit a normal request to Go's `/qr` endpoint, and confirm it succeeds. Then call Node's statistics endpoint directly without Authorization and confirm `401 {"error":"unauthorized"}`. This proves the complete local flow without exposing a JWT to the frontend.
