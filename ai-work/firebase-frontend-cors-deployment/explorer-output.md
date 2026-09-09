# Scope exploration: `firebase-frontend-cors-deployment`

## What will be built

Prepare the existing Terraform-managed Go Cloud Run service so that the already deployed Firebase frontend can call `POST /qr` from the browser. The deployment configuration must pass the exact Firebase origin
`https://interseguro-challenge-508112.web.app` to the Go container through `CORS_ALLOWED_ORIGINS`, and must point Go Cloud Run at the newly built `go-api:v2` image. Node remains on `node-api:v1`.

The feature should result in a new Go Cloud Run revision while preserving the existing Go service URI. Terraform `apply`, Firebase deployment, and any manual GCP Console changes are explicitly outside this exploration/feature stage.

## Actors

- Firebase Hosting frontend at `https://interseguro-challenge-508112.web.app`.
- Browser, which issues the CORS preflight and then the JSON request.
- Public Go/Fiber Cloud Run service (`google_cloud_run_v2_service.go_api`), which owns CORS and QR factorization.
- Node/Express Cloud Run statistics service, unchanged and still configured as `node-api:v1`.
- Terraform configuration and the local ignored `infrastructure/terraform.tfvars`.
- Artifact Registry image `southamerica-west1-docker.pkg.dev/interseguro-challenge-508112/interseguro-challenge/go-api:v2`.

## Current Go CORS findings

Configuration path:

1. `go-api/main.go` reads `os.Getenv("CORS_ALLOWED_ORIGINS")`.
2. It passes the value to `httpapi.New`.
3. `go-api/httpapi/app.go` parses it with `strings.Split(allowedOrigins, ",")` and supplies the resulting list to Fiber v3's `middleware/cors` as `AllowOrigins`.

The current Fiber configuration is:

- middleware: `github.com/gofiber/fiber/v3/middleware/cors`;
- methods: `POST, OPTIONS`;
- allowed request headers: `Content-Type`;
- allowed origins: the comma-separated values from `CORS_ALLOWED_ORIGINS`.

`POST /qr` is registered after the middleware, so the middleware handles the browser `OPTIONS /qr` preflight before the QR handler. The existing `TestCORS` sends an `OPTIONS /qr` request with `Origin`, `Access-Control-Request-Method: POST`, and `Access-Control-Request-Headers: Content-Type`; it expects `204`, `Access-Control-Allow-Origin`, `Access-Control-Allow-Methods: POST, OPTIONS`, and `Access-Control-Allow-Headers: Content-Type`. The same test sends JSON `POST /qr` requests and verifies that the configured origin receives `Access-Control-Allow-Origin`, while an untrusted origin does not.

Conclusion: for the requested single value, the Go implementation already supports the required browser flow. No `go-api/**` change is required by the stated scope.

## Main flow

1. Terraform supplies `CORS_ALLOWED_ORIGINS` to the Go Cloud Run container with the exact value `https://interseguro-challenge-508112.web.app` (no trailing slash and no wildcard).
2. The Go service starts with `NODE_API_URL` unchanged, pointing to the Terraform-managed Node service URI.
3. The browser sends `OPTIONS /qr` from the Firebase origin, requesting `POST` with `Content-Type`.
4. Fiber CORS returns the successful preflight response for that exact origin.
5. The browser sends `POST /qr` with `Content-Type: application/json`.
6. Go factorizes the matrix and calls Node through the existing statistics client.
7. Go returns QR matrices and statistics to the browser with the allowed origin response header.

## Required configuration changes

The implementation stage should be limited to these configuration files unless inspection during implementation reveals a real blocker:

- `infrastructure/variables.tf`: add the required non-empty string variable `cors_allowed_origins`.
- `infrastructure/cloud-run.tf`: keep the existing `NODE_API_URL` environment entry unchanged and add `CORS_ALLOWED_ORIGINS = var.cors_allowed_origins` only to the Go container. Do not add it to Node and do not add a manual `PORT` entry.
- `infrastructure/terraform.tfvars.example`: add a safe placeholder such as `cors_allowed_origins = "https://your-project.web.app"`.
- ignored local `infrastructure/terraform.tfvars`: set the real Firebase origin and change only `go_image` from `.../go-api:v1` to `.../go-api:v2`; preserve `node_image = .../node-api:v1`. This file is ignored by Git and must not become versioned.

The Go image to prepare is:

`southamerica-west1-docker.pkg.dev/interseguro-challenge-508112/interseguro-challenge/go-api:v2`

The requested build/push command is an external deployment-side operation and is not part of this exploration. It must not reuse or overwrite `v1`. No Node image rebuild is needed.

## Edge cases and risks

- The origin comparison is exact. A trailing slash, `*`, or an added cloud `localhost` value would not satisfy the requirement and should not be introduced.
- `strings.Split` supports multiple comma-separated origins, but does not trim whitespace around each item. The requested production value is a single origin with no whitespace, so this is not a blocker for this feature. Do not change Go merely for style; if future multi-origin values are needed, parsing/trimming would require a separate explicit decision and tests.
- An empty or unset `CORS_ALLOWED_ORIGINS` would not allow the Firebase origin. Terraform validation should reject an empty/whitespace-only variable before deployment.
- The preflight only advertises `POST` and `OPTIONS` and `Content-Type`; this matches the frontend QR request. Adding methods or headers is not required by the current flow.
- A successful CORS preflight does not guarantee the downstream statistics call succeeds. Go still depends on the existing `NODE_API_URL` and Node availability; Node code, IAM, and private service changes are out of scope.
- Terraform state or pre-existing cloud drift could produce plan changes beyond the expected Go service update. The implementation stage must stop and report unexpected resources or changes rather than applying them.
- The expected plan (`0 add, 1 change, 0 destroy`) assumes the current state matches this repository, the image exists in Artifact Registry, and no unrelated drift exists. This cannot be confirmed during exploration without running the explicitly deferred plan.
- The local `terraform.tfvars` may be absent in another checkout or may contain user-specific values. It is ignored and must be handled locally without committing it.
- Firebase files and frontend source are already deployed and must not be modified.

## Dependencies

- Existing Go CORS implementation and its `TestCORS` coverage.
- Fiber v3 CORS middleware behavior.
- Terraform Google provider resources for the two existing Cloud Run services.
- Artifact Registry permissions and Docker/buildx availability for publishing `go-api:v2`.
- Existing Node Cloud Run service URI exposed through `google_cloud_run_v2_service.node_api.uri`.
- A valid Terraform state and credentials for later validation/plan/deployment by the main workflow.

## Out of scope

- `terraform apply`.
- Firebase deploy, Firebase configuration, frontend source, or frontend rebuild.
- Any Node source or Node Terraform/image change; Node must remain `node-api:v1`.
- IAM, service-to-service authentication, private Node Cloud Run, JWT, GitHub Actions, or Artifact Registry resource changes.
- CORS reimplementation or Go changes without a demonstrated blocker.
- Adding wildcard origins, trailing-slash variants, or production `localhost` origins.
- Creating future repository directories or changing unrelated documentation/configuration.

## Resolved questions

- The production origin is exactly `https://interseguro-challenge-508112.web.app`.
- The Go image tag is `v2`; `v1` must not be reused. Node remains `v1`.
- The origin configuration is a string because the current Go API accepts a string and splits comma-separated origins.
- The existing CORS behavior covers `OPTIONS`, `POST`, JSON `Content-Type`, and the `/qr` route; Go source should remain unchanged.
- Terraform should update only the existing Go Cloud Run service for this feature, subject to plan verification.

## Unresolved blockers / decisions for the next stage

- No functional blocker was found. The planner/developer stage must decide whether to trim comma-separated origin entries; current scope does not require that change and the exact single-origin deployment is safe.
- The actual build/push result, `terraform validate`, and full `terraform plan` remain unverified by this exploration. The workflow must run them later and stop if the plan is not limited to `google_cloud_run_v2_service.go_api` with `0 add / 1 change / 0 destroy`.
- Whether the ignored local `terraform.tfvars` should be edited depends on its presence and local workflow policy; if not edited, the operator must add the exact production-origin and `v2` entries manually before planning.
