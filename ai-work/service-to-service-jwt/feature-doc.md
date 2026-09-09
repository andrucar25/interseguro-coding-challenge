# Service-to-service JWT

## Scope and flow

JWT protects only the internal statistics request:

```text
Frontend → public Go API → Authorization: Bearer <HS256 JWT> → Node API
```

The frontend does not create, send, or receive a token. Go creates a fresh token immediately before each `POST /api/v1/statistics` request. The `{ q, r }` payload and successful response contracts are unchanged; the token is never returned to the frontend or logged.

The token contains `iss`, `aud`, `iat`, and `exp`, with a one-minute lifetime. Node explicitly allows only `HS256` and verifies the signature, expiration, issuer, and audience before parsing the request body or running statistics. This is service authentication, not user authentication: there are no users, roles, sessions, refresh tokens, or login flow.

## Configuration

Both services must receive the same secret and matching claim values through external configuration:

| Service | Variable | Local default/example | Purpose |
| --- | --- | --- | --- |
| Go | `NODE_API_JWT_SECRET` | `replace-with-a-local-development-secret` | HS256 signing secret |
| Go | `NODE_API_JWT_ISSUER` | `interseguro-go-api` | `iss` claim |
| Go | `NODE_API_JWT_AUDIENCE` | `interseguro-node-api` | `aud` claim |
| Node | `JWT_SECRET` | `replace-with-a-local-development-secret` | HS256 verification secret |
| Node | `JWT_ISSUER` | `interseguro-go-api` | expected issuer |
| Node | `JWT_AUDIENCE` | `interseguro-node-api` | expected audience |

Do not commit a real secret, `.env`, or `terraform.tfvars`. Docker Compose maps one externally supplied root `JWT_SECRET` to the Go and Node variable names. Terraform maps its sensitive `jwt_secret` input to both services; Terraform was not executed as part of this feature.

## Node error contract

`POST /api/v1/statistics` returns the same response for missing, empty, malformed, incorrectly signed, expired, wrong-issuer, and wrong-audience credentials:

```http
401 Unauthorized
Content-Type: application/json

{"error":"unauthorized"}
```

No cryptographic failure details are exposed. A valid token continues to the existing validation and statistics behavior. Node may remain publicly addressable; JWT protects the endpoint from requests that do not possess the shared service credential, but does not replace Cloud Run IAM.

If Node returns `401`, Go preserves its existing downstream error handling and does not expose a JWT-specific response to the frontend.

## Local testing

From the repository root, provide a local-only secret through the shell or an untracked `.env` file, then start the complete flow:

```bash
JWT_SECRET='replace-with-a-local-development-secret' docker compose up --build
```

The frontend remains at `http://localhost:5173`, Go at `http://localhost:8080`, and Node at `http://localhost:3000`. The normal request is still sent to Go:

```bash
curl -X POST http://localhost:8080/qr \
  -H 'Content-Type: application/json' \
  -d '{"matrix":[[1,2],[3,4]]}'
```

Direct calls to Node without a valid service token receive the 401 response above. Automated coverage includes all required invalid-token cases and a valid-token request that reaches statistics, plus Go tests for claims, lifetime, and the outbound Authorization header.

## Dependencies and security boundary

Go uses `github.com/golang-jwt/jwt/v5` for signing and Node uses `jose` for verification. No authentication framework or user store is involved. HS256 means both services hold the shared signing secret, so compromise of either service configuration can mint tokens; keep the secret outside source, images, logs, and versioned configuration, and rotate it externally if compromised.
