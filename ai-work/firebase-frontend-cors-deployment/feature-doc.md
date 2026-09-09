# Firebase frontend CORS deployment

## Purpose

Configure the existing Go Cloud Run service to accept browser requests from the deployed Firebase Hosting frontend. The production origin is exactly:

`https://interseguro-challenge-508112.web.app`

There is no trailing slash, wildcard origin, or production `localhost` entry.

## CORS behavior

The Go service reads `CORS_ALLOWED_ORIGINS` in `go-api/main.go` and passes it to `httpapi.New`. `go-api/httpapi/app.go` splits the string on commas and supplies the resulting origins to Fiber v3's `middleware/cors`.

The middleware allows `POST` and `OPTIONS` methods and the `Content-Type` request header. It is registered before `POST /qr`, so a browser preflight to `OPTIONS /qr` is handled before the QR endpoint. The existing `TestCORS` covers a successful JSON preflight and the subsequent `POST /qr`, including the `Access-Control-Allow-Origin` response header for the configured origin.

No Go source change was required.

## Deployment configuration

Terraform adds the required non-empty string variable `cors_allowed_origins` and injects it only into the Go Cloud Run container as:

```hcl
env {
  name  = "CORS_ALLOWED_ORIGINS"
  value = var.cors_allowed_origins
}
```

The existing `NODE_API_URL = google_cloud_run_v2_service.node_api.uri` configuration is unchanged. Node does not receive the CORS variable and remains on:

`southamerica-west1-docker.pkg.dev/interseguro-challenge-508112/interseguro-challenge/node-api:v1`

The Go image is updated to:

`southamerica-west1-docker.pkg.dev/interseguro-challenge-508112/interseguro-challenge/go-api:v2`

The versioned example uses `https://your-project.web.app`; the ignored local `terraform.tfvars` uses the exact production Firebase origin and the Go `v2` image.

## Verification and deployment boundary

The intended verification commands are `terraform fmt`, `terraform validate`, and `terraform plan -var-file=terraform.tfvars` from `infrastructure/`. The acceptable plan is `0 to add, 1 to change, 0 to destroy`, with only `google_cloud_run_v2_service.go_api` changing (Go image `v1` to `v2` and the CORS environment variable). Any Node, IAM, Artifact Registry, service-account, URI, or destroy change requires stopping for review.

`terraform apply`, Firebase deployment, manual GCP Console changes, and post-apply browser verification remain intentionally deferred. The Go image build/push, if not authorized or completed by the execution workflow, must be run manually with the approved `go-api:v2` Buildx command before applying the plan.

## Scope record

Versioned implementation files:

- `infrastructure/variables.tf`
- `infrastructure/cloud-run.tf`
- `infrastructure/terraform.tfvars.example`

Local-only change:

- ignored `infrastructure/terraform.tfvars` (`cors_allowed_origins`, Go `v2`; Node remains `v1`)

Out of scope: Node code, frontend source or Firebase configuration, IAM, JWT, private Node Cloud Run, Artifact Registry resources, CI/CD, and Go CORS reimplementation.
