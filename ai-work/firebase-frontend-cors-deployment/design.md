# Development Plan - firebase-frontend-cors-deployment

## Summary

Prepare the existing Go Cloud Run service for browser calls from the already-published Firebase Hosting origin `https://interseguro-challenge-508112.web.app`. Terraform will pass that origin to the existing Go CORS configuration and update only the Go image from `v1` to `v2`; Node remains on `v1`.

The existing Go implementation needs no change. `go-api/main.go` reads `CORS_ALLOWED_ORIGINS` with `os.Getenv` and passes it to `httpapi.New`. `go-api/httpapi/app.go` splits the string on commas and configures Fiber v3's `middleware/cors` with that list, `POST` and `OPTIONS` methods, and `Content-Type` as the allowed request header. The middleware is registered before `POST /qr`; the existing `TestCORS` proves that a browser preflight to `/qr` receives a `204` response and the allow-origin, allow-methods, and allow-headers values, and that the following JSON `POST /qr` receives the allow-origin header for a configured origin. A single exact origin without whitespace is within the current parser's supported behavior, so trimming multi-origin inputs is not needed for this feature.

No Firebase, frontend, Node, IAM, Artifact Registry resource, or Go-to-Node contract changes are included. Terraform apply remains explicitly deferred for human review after the plan.

## Files to Create

None during implementation. This design is the planning artifact.

## Files to Modify

- `infrastructure/variables.tf`: add the required string variable `cors_allowed_origins`, using the repository's existing non-empty `trimspace` validation convention.
- `infrastructure/cloud-run.tf`: in the `google_cloud_run_v2_service.go_api` container, preserve the current `NODE_API_URL = google_cloud_run_v2_service.node_api.uri` block unchanged and add `CORS_ALLOWED_ORIGINS = var.cors_allowed_origins`. Do not add this environment variable to the Node container and do not configure `PORT`.
- `infrastructure/terraform.tfvars.example`: add `cors_allowed_origins = "https://your-project.web.app"` as a safe, versioned example. Keep the example's image tags otherwise unchanged unless the approved implementation deliberately documents the v2 rollout example.
- `infrastructure/terraform.tfvars` (ignored, local-only): set `cors_allowed_origins = "https://interseguro-challenge-508112.web.app"`; change only `go_image` from `southamerica-west1-docker.pkg.dev/interseguro-challenge-508112/interseguro-challenge/go-api:v1` to `southamerica-west1-docker.pkg.dev/interseguro-challenge-508112/interseguro-challenge/go-api:v2`; retain `node_image = "southamerica-west1-docker.pkg.dev/interseguro-challenge-508112/interseguro-challenge/node-api:v1"`. Confirm it remains ignored and never stage it.

No `go-api/**` files will be modified unless a new inspection identifies a concrete incompatibility with `OPTIONS`, `POST`, or `Content-Type: application/json`; none is currently present. No additional versioned documentation is needed because this is an environment-specific deployment value already represented by Terraform's variable and example configuration.

## Execution Order

1. Reconfirm the existing Go CORS path and its test coverage before editing: `main.go` reads `CORS_ALLOWED_ORIGINS`; `httpapi.New` configures Fiber CORS; `POST /qr` remains after the middleware. Confirm the configured methods are `POST, OPTIONS`, the allowed header is `Content-Type`, and the exact Firebase origin is neither suffixed with `/` nor replaced with `*`.
2. Edit only the three versioned Terraform configuration files listed above. Add the non-empty `cors_allowed_origins` variable and wire it only into the Go Cloud Run container.
3. Update the ignored local `infrastructure/terraform.tfvars` if it is present and permitted by the local workflow. If it is unavailable or local-file edits are not appropriate, leave it untouched and have the operator add these exact lines before planning:

   ```hcl
   cors_allowed_origins = "https://interseguro-challenge-508112.web.app"
   go_image             = "southamerica-west1-docker.pkg.dev/interseguro-challenge-508112/interseguro-challenge/go-api:v2"
   ```

   Do not alter the local `node_image` value.
4. Build and publish the new Go image from the repository root, after the configuration changes are ready and only with credentials/permissions that allow the explicitly requested push:

   ```bash
   docker buildx build \
     --platform linux/amd64 \
     --tag southamerica-west1-docker.pkg.dev/interseguro-challenge-508112/interseguro-challenge/go-api:v2 \
     --push \
     ./go-api
   ```

   Do not use `latest`, overwrite `v1`, or build/push Node. If the execution workflow cannot perform an external registry push automatically, report this exact command for manual execution and do not claim the image was published.
5. Because Go source is unchanged, do not make a Go-only change merely to run quality gates. If a real CORS blocker forces a Go edit, run from `go-api/`: `find . -type f -name '*.go' -exec gofmt -w {} +`, `test -z "$(find . -type f -name '*.go' -exec gofmt -l {} +)"`, `go vet ./...`, `go test ./...`, and `go build ./...` (plus an existing configured lint command, if one exists).
6. From `infrastructure/`, run `terraform fmt`, `terraform validate`, and `terraform plan -var-file=terraform.tfvars`. Inspect the full plan before any deployment action.
7. Continue only when the plan reports `0 to add, 1 to change, 0 to destroy` and the only changed cloud resource is `google_cloud_run_v2_service.go_api`, with the Go image changing from `v1` to `v2` and the Go container gaining `CORS_ALLOWED_ORIGINS`. Stop and report any other resource/change, including Node, Artifact Registry, IAM, service account, URI, or destroy changes.
8. Do not run `terraform apply`. Stop for explicit human review of the successful plan and later approval.

## Architecture Decisions

- Keep CORS at the public Go/Fiber boundary, where it already belongs; Node receives no browser traffic configuration change.
- Use the existing string configuration contract rather than introducing a Terraform collection or Go parser redesign. Fiber already accepts the comma-split values, and this release supplies one exact origin.
- Permit exactly `https://interseguro-challenge-508112.web.app` in the local production tfvars: no trailing slash, wildcard, or cloud `localhost` entry. Docker Compose/local development remains independently configured.
- Keep `NODE_API_URL` Terraform-derived from `google_cloud_run_v2_service.node_api.uri`, preserving the stable Go service URI and existing Go-to-Node flow.
- Publish a new immutable-in-practice tag `go-api:v2` because CORS support was added after `v1` was built. Retain `node-api:v1` because Node has no change.
- Rely on the existing focused Fiber CORS test rather than adding a duplicate test: it covers the required preflight and allowed/untrusted-origin POST behavior. The final browser verification occurs after a separately approved apply.

## Validation Criteria

- Source inspection confirms `CORS_ALLOWED_ORIGINS` is read in `go-api/main.go`, passed to `httpapi.New`, comma-split for Fiber's CORS middleware, and applied before `POST /qr`.
- Existing CORS behavior supports `OPTIONS /qr` preflight and `POST /qr` with `Content-Type: application/json`; allowed methods are `POST, OPTIONS` and allowed request headers are `Content-Type`.
- The published Go image, if the push is executed, is exactly `southamerica-west1-docker.pkg.dev/interseguro-challenge-508112/interseguro-challenge/go-api:v2`. Node remains exactly `southamerica-west1-docker.pkg.dev/interseguro-challenge-508112/interseguro-challenge/node-api:v1`.
- `terraform fmt` completes, `terraform validate` succeeds, and `terraform plan -var-file=terraform.tfvars` is reviewed in full.
- The accepted plan is exactly `0 add / 1 change / 0 destroy`; its sole resource change is `google_cloud_run_v2_service.go_api` and is limited in substance to the Go image revision and `CORS_ALLOWED_ORIGINS` environment value. Any deviation blocks deployment pending review.
- No `terraform apply`, Firebase deployment/configuration change, manual GCP Console change, Node change, IAM change, JWT, CI/CD, or frontend rebuild is performed.

## External Dependencies

- Docker Buildx, registry authentication, and permission to push the Go `v2` image to Artifact Registry.
- Terraform installation, initialized state/backend, Google Cloud credentials, and access sufficient for `validate`/`plan`.
- Existing Cloud Run Go and Node services, with the Node URI available through Terraform and the Node image retained at `v1`.
- The deployed Firebase Hosting site at the exact configured origin.

## Blocked Tasks

- `terraform apply` is intentionally blocked until a human reviews the resulting plan and explicitly approves deployment.
- Final production browser verification (`OPTIONS /qr`, then JSON `POST /qr`, QR factorization, Node statistics, and response) is blocked until that approved apply creates the Go Cloud Run revision.
- Image publishing is blocked only if local Docker/Artifact Registry credentials or the execution workflow do not authorize the push; in that case the exact build/push command above must be run manually before planning an image update that references `v2`.
