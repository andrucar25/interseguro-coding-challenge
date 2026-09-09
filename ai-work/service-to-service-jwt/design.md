# Development Plan - service-to-service-jwt

## Summary

Protect only the internal `Go API -> Node API` call with a short-lived shared-secret JWT.  The public browser flow remains unchanged:

```text
Frontend
   ↓
Go API (public; performs QR)
   ↓ Authorization: Bearer <HS256 JWT>
Node API (statistics)
```

Go creates a new token immediately before each `POST /api/v1/statistics` request. Node verifies it before it reaches the existing statistics controller. The frontend neither sends nor receives a JWT, and the existing `{ q, r }` request body and successful responses remain unchanged.

The token is signed with HS256 and a shared secret obtained only from environment configuration. It contains no user or authorization data:

| Claim | Value |
| --- | --- |
| `iss` | configured Go issuer, normally `interseguro-go-api` |
| `aud` | configured Node audience, normally `interseguro-node-api` |
| `iat` | current issuance time |
| `exp` | issuance time plus one minute |

One minute is intentionally short for an immediate internal HTTP request, while avoiding a refresh mechanism. Node will explicitly allow only HS256, require a valid signature, expiration, issuer, and audience. Missing, empty, malformed, expired, wrongly signed, wrong-issuer, and wrong-audience Bearer credentials all return `401 {"error":"unauthorized"}` without cryptographic diagnostics.

## Files to Create

- `go-api/external/statistics/token.go` — compact HS256 token signer owned by the Go-to-Node client boundary. It will validate its supplied secret/issuer/audience at construction, mint the four required claims with a one-minute lifetime, and never log or return the secret/token in an error.
- `go-api/external/statistics/token_test.go` — focused Go tests that parse generated tokens with the test secret and assert algorithm, issuer, audience, `iat`, and a bounded short expiration without asserting an exact clock instant.
- `node-api/src/auth/service-token.ts` — typed runtime JWT configuration plus a small Express middleware factory that performs strict Bearer extraction and `jose` HS256 verification before calling `next()`.
- `node-api/src/auth/service-token.test.ts` — focused middleware/config tests only if they improve coverage beyond the endpoint tests; otherwise do not create this file and keep all required authorization cases in `src/app.test.ts`.

The exact Node test-file choice is intentionally left to the implementer after composing the middleware: the required cases must exist once, not be duplicated across unit and Supertest suites.

## Files to Modify

### Go API

- `go-api/go.mod` and `go-api/go.sum` — add the current compatible `github.com/golang-jwt/jwt/v5` release after confirming its version at implementation time. Go 1.27.1 has no standard JWT implementation; this maintained library avoids manually implementing signing/parsing.
- `go-api/config/config.go` — add `NODE_API_JWT_SECRET`, `NODE_API_JWT_ISSUER`, and `NODE_API_JWT_AUDIENCE` to `Config`; retain all existing values and defaults. Do not add a token-lifetime environment variable for this fixed challenge policy.
- `go-api/config/config_test.go` — preserve current cases and verify the three new environment values are loaded.
- `go-api/main.go` — pass the three Go JWT settings to the existing statistics client during startup. Empty/invalid signing settings must fail startup with a safe configuration error; no credential values appear in logs.
- `go-api/external/statistics/client.go` — compose the signer into the existing reusable client, mint a token immediately before creating/executing the existing request, and set exactly `Authorization: Bearer <token>` alongside the unchanged JSON body and content type. Its current non-2xx mapping remains unchanged, so Node `401` continues to be categorized as `ErrDownstream` and Go continues to return its existing downstream error response.
- `go-api/external/statistics/client_test.go` — retain existing contract/error tests and add an `httptest` fake Node assertion that Authorization is present, is Bearer-shaped, and verifies as HS256 with the expected issuer/audience/short expiration. Include a successful real client flow to prove QR-to-statistics behavior is not broken. Do not compare a token string or exact timestamps.
- `go-api/.env.example` — document the three required Go JWT variables with placeholder/local-development values only, never a deployable secret.

### Node API

- `node-api/package.json` and `node-api/package-lock.json` — add the current compatible `jose` release after checking the existing Node 24/TypeScript 7 dependency set. `jose` has native TypeScript support and ESM support, so no separate `@types` package is needed; it supports explicit algorithm allow-lists and avoids a manual verifier.
- `node-api/src/app.ts` — compose the protected statistics router with validated JWT configuration. Prefer an explicit app/router factory that lets tests supply a test-only verifier configuration while `index.ts` builds the production instance once. Do not read secrets per request or expose them through error handlers.
- `node-api/src/index.ts` — load/validate `JWT_SECRET`, `JWT_ISSUER`, and `JWT_AUDIENCE` before listening; fail fast with a generic missing/invalid JWT configuration message without printing values.
- `node-api/src/statistics/statistics.router.ts` — attach the JWT middleware specifically to `POST /api/v1/statistics` before `calculateStatisticsResponse`. Do not change the endpoint path, JSON body validation, statistics implementation, or success body.
- `node-api/src/app.test.ts` — update existing successful and validation requests to use a valid test token, then add Supertest coverage for no Authorization, empty/malformed Bearer input, invalid signature, expiration, issuer, audience, and a valid JWT which reaches the current statistics response. Each authentication failure asserts exactly `401` and `{ error: "unauthorized" }`.
- `node-api/.env.example` — document `JWT_SECRET`, `JWT_ISSUER`, and `JWT_AUDIENCE` with placeholders/local-development-only values; do not include a real shared secret.

### Local deployment, cloud configuration, and documentation

- `docker-compose.yaml` — inject the same externally supplied Compose secret into `NODE_API_JWT_SECRET` for Go and `JWT_SECRET` for Node, and inject matching issuer/audience values with each service's own variable names. Use Compose variable substitution (for example a required `JWT_SECRET` supplied by a local untracked `.env` or shell), not an image/Dockerfile literal. The frontend remains unchanged.
- `infrastructure/variables.tf` — introduce `jwt_secret` as a non-empty `string` marked `sensitive = true`, plus non-secret, non-empty issuer/audience variables. Do not add Secret Manager, IAM, service accounts, or Cloud Run identity changes.
- `infrastructure/cloud-run.tf` — set the three `NODE_API_JWT_*` values on Go and matching `JWT_*` values on Node. Both receive the identical secret variable; issuer/audience must match. Keep Node publicly reachable at the Cloud Run layer as currently configured, relying on endpoint JWT verification for this feature.
- `infrastructure/terraform.tfvars.example` — add only the non-secret issuer/audience examples if useful; never add `jwt_secret`. Operators provide the sensitive value via an untracked `terraform.tfvars`, `TF_VAR_jwt_secret`, or equivalent secure local/CI injection.
- `README.md` — replace the optional-JWT-not-implemented statement with a short description of service-to-service protection, the flow above, the required configuration names, and the fact that public Go remains intentionally available for the Firebase demo. State explicitly that this is not frontend/user authentication.

No Dockerfile, frontend source, QR logic, statistics logic, input-validation schema, API paths, success contracts, Terraform IAM, or Terraform Secret Manager resource is changed.

## Execution Order

1. Reconfirm the dependency versions in the existing manifests/lockfiles, then add only `github.com/golang-jwt/jwt/v5` to Go and `jose` to Node. Do not add Passport, user/session packages, or crypto primitives.
2. Implement Go configuration and the small signer, using `jwt.NewWithClaims(jwt.SigningMethodHS256, ...)` and a one-minute `exp`. Reject empty signing configuration at construction/startup. Keep signing time injectable or bounded in tests so they do not depend on an exact wall clock.
3. Extend the existing Go statistics client to use the signer per outbound request and write its Authorization-header and claims tests. Preserve payload encoding, timeout, response decoding, and `ErrDownstream` behavior.
4. Implement Node's typed config and strict middleware. It must accept only one Authorization header matching `Bearer` followed by a non-empty token, call `jwtVerify` with the shared secret and `{ algorithms: ["HS256"], issuer, audience }`, and send the same safe 401 body for every authentication failure. Only successful verification calls `next()`.
5. Wire the middleware only to `POST /api/v1/statistics`, then update/add the required Node HTTP tests. A valid test token is generated with the same `jose` library and test configuration, not pasted as a static opaque token.
6. Update environment examples, Compose, Terraform injection, and README. Ensure every example secret is a non-production placeholder, the Terraform secret input is sensitive, and no secret is put in Dockerfiles or `terraform.tfvars.example`.
7. Format and run the Go, Node, and Terraform checks in **Validation Criteria**. Do not apply Terraform, deploy Cloud Run, or commit a real `.env`/`terraform.tfvars` file.

## Architecture Decisions

### JWT flow and boundary

The JWT belongs solely to the Go HTTP client and Node route middleware. QR remains a pure Go concern; statistics remains the existing Node controller/service concern. This gives the feature one clear security boundary and makes it easy to remove or replace later without contaminating frontend or mathematical code.

```text
POST /qr from browser (no JWT)
  -> Go factorizes matrix
  -> Go signs a new HS256 token: iss, aud, iat, exp (+60s)
  -> Go POSTs unchanged {q,r} with Authorization: Bearer token
  -> Node verifies allowed algorithm, signature, exp, issuer, audience
  -> Node calculates statistics only after verification
  -> Go returns existing consolidated result (no token)
```

HS256 is selected because the approved architecture explicitly uses one shared secret and two controlled services. Asymmetric keys, key IDs/rotation, refresh tokens, sessions, databases, roles, and user claims add no value to this challenge. The trade-off is that both services possess signing-capable key material; this is acceptable for the deliberately small shared-secret design and must be called out in the security risks.

### Strict Bearer and error behavior

Node treats Authorization as an authentication boundary, before body validation/statistics execution. It must reject absent headers, empty values, wrong scheme/casing format if not exactly accepted by the parser, multiple/ambiguous header values, and malformed token strings. Authentication failures intentionally collapse to the same response:

```http
401 Unauthorized
Content-Type: application/json

{"error":"unauthorized"}
```

No token parsing, signature, claims, or secret details are returned. `alg=none` is rejected by cryptographic verification plus the explicit HS256 allow-list. Node does not accept a decoded-but-unverified payload.

Go does not transform an unauthorized Node response into a frontend JWT error. The existing statistics client already classifies all downstream non-2xx statuses as `ErrDownstream`, and the handler maps that to its existing `502 downstream_error` response. This preserves the public contract and avoids exposing a service-to-service security diagnostic.

### Configuration contract

| Scope | Environment variable | Required | Purpose |
| --- | --- | --- | --- |
| Go | `NODE_API_JWT_SECRET` | yes | HS256 shared secret used to sign Node-bound tokens |
| Go | `NODE_API_JWT_ISSUER` | yes | `iss` signed into each token |
| Go | `NODE_API_JWT_AUDIENCE` | yes | `aud` signed into each token |
| Node | `JWT_SECRET` | yes | HS256 shared secret used to verify signature |
| Node | `JWT_ISSUER` | yes | required token issuer |
| Node | `JWT_AUDIENCE` | yes | required token audience |

Terraform maps one sensitive `jwt_secret` input into the Go/Node secret names and maps shared non-secret issuer/audience inputs into the corresponding names. Docker Compose does the same with a locally provided `JWT_SECRET`. Different names at the service boundary make each service's intent clear while keeping the deployment mapping explicit.

### Test design

Node Supertest cases must prove all requested boundary behavior:

- no Authorization header -> 401 unauthorized;
- empty or malformed Bearer header -> 401 unauthorized;
- token signed with a different secret -> 401 unauthorized;
- validly signed expired token -> 401 unauthorized;
- validly signed token with a different issuer -> 401 unauthorized;
- validly signed token with a different audience -> 401 unauthorized;
- valid HS256 token with matching claims -> existing `200` statistics response.

Go tests must prove the signer emits HS256 and the expected claims, with `exp` after `iat` and within an intentionally tolerant one-minute range. The existing `httptest` Node server test will verify the outbound Authorization header and then return statistics, proving the full Go client flow remains functional. Tests use generated tokens and bounded time windows rather than serialized token strings or exact timestamps.

## Validation Criteria

The feature is complete only when:

- The normal browser-to-Go `/qr` request stays JWT-free, and neither Go responses nor logs expose a JWT.
- Go sends the unchanged `{q,r}` payload plus `Authorization: Bearer <HS256 token>` on every Node statistics request; that token has only required standard claims plus any JWT-library-required metadata, with a one-minute lifetime.
- Node executes no statistics route/controller work before a valid token is verified; all required invalid-token categories yield exactly 401 and the safe `unauthorized` body, while a valid token preserves the existing success contract.
- Empty/malformed JWT configuration causes a safe startup/configuration failure rather than unsigned/partially verified requests.
- All secret values are external, no real secret is committed, the Terraform secret variable is sensitive, and `terraform.tfvars.example`, Dockerfiles, source, README, and logs contain no secret.
- Existing QR, validation, statistics, and downstream-error behavior remains intact, including Go's current mapping of Node 401 to downstream failure.

Run these commands after implementation:

```bash
cd go-api
find . -type f -name '*.go' -exec gofmt -w {} +
test -z "$(find . -type f -name '*.go' -exec gofmt -l {} +)"
go vet ./...
go test ./...
go build ./...

cd ../node-api
npm run typecheck
npm run check
npm test
npm run build

cd ../infrastructure
terraform fmt -check
terraform validate
```

`terraform validate` may require the provider modules already initialized by the operator. The implementation must not run `terraform apply` as part of this feature.

## External Dependencies

- Go: add `github.com/golang-jwt/jwt/v5`, selecting its current maintained v5 release compatible with the inspected Go 1.27.1 module. It provides well-known claim types and HS256 signing/parsing without manual cryptographic code.
- Node: add `jose`, selecting its current release compatible with the inspected Node 24.x ESM/TypeScript 7 application. It provides its own TypeScript declarations and explicit verification constraints, avoiding `@types` maintenance drift.

No authentication framework, Passport, database client, Google auth library, Secret Manager dependency, or new runtime framework is required.

## Blocked Tasks

Deployment requires the operator to supply a real secret outside version control; this plan does not create, print, or commit one. Terraform commands were intentionally not run at the user's request, so Terraform formatting and provider validation remain for the operator.

## Execution Results

Completed:

- Added `github.com/golang-jwt/jwt/v5` v5.3.1 and `jose` v6.2.12 after checking their current compatible releases.
- Implemented short-lived HS256 token signing in Go, required Go configuration, and an Authorization header on every statistics request.
- Implemented strict Bearer parsing and HS256/signature/expiration/issuer/audience verification before Node parses or calculates statistics.
- Added the required Go and Node JWT coverage, plus environment examples, Compose wiring, Cloud Run environment injection, Terraform inputs, and README documentation.

Validation completed:

```text
go-api: gofmt check, go vet ./..., go test ./..., go build ./... (passed)
node-api: npm run typecheck, npm run check, npm test, npm run build (passed)
```

Blocked by explicit user instruction:

```text
infrastructure: Terraform fmt/validate/init/apply were not run.
```

## Security Risks

- HS256 makes the shared secret both a verification and signing key. A compromise of either service/configuration can mint valid tokens. Keep it out of source, logs, images, examples containing real values, and versioned Terraform variables; rotate it externally if compromised.
- JWT does not make a publicly reachable Node service private. It protects this endpoint from unauthenticated application requests, but Cloud Run IAM/private ingress is explicitly out of scope.
- A one-minute token limits replay exposure but does not eliminate it within that window. No nonce/jti store is added because it would require state and is outside the small challenge scope.
- Correct validation depends on algorithm pinning and mandatory `exp`, issuer, audience, and signature checks. The design explicitly prevents `alg=none` and decoded-but-unverified claims.
- A mismatch between Go and Node secret/issuer/audience causes Node 401 and Go's existing 502 downstream error. Fail-fast configuration validation and Compose/Terraform mapping reduce, but cannot eliminate, deployment misconfiguration.
