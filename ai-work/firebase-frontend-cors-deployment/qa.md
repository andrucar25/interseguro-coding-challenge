# QA Report - firebase-frontend-cors-deployment

## Result

PASS, subject to the intentionally deferred `terraform apply` and browser verification.

The implemented, versioned Terraform change is limited to the approved scope. It supplies the exact Firebase Hosting origin to the public Go Cloud Run service and rolls only the Go image from `v1` to `v2`. The Node service/image is unchanged.

## Implementation Review

- `go-api/main.go` reads `CORS_ALLOWED_ORIGINS` with `os.Getenv` and passes it to `httpapi.New`.
- `go-api/httpapi/app.go` parses one or more origins with `strings.Split(allowedOrigins, ",")` and registers Fiber v3 `middleware/cors` before `POST /qr`.
- The CORS configuration permits `POST` and `OPTIONS`, and permits `Content-Type`; this supports a browser `OPTIONS /qr` preflight followed by `POST /qr` with `Content-Type: application/json`.
- The existing test covers successful preflight (`204`) plus the expected allow-origin, allow-methods, and allow-headers response headers, and verifies that an untrusted POST origin receives no `Access-Control-Allow-Origin` header.
- No Go source change was necessary or made.
- `infrastructure/cloud-run.tf` retains `NODE_API_URL = google_cloud_run_v2_service.node_api.uri` and adds `CORS_ALLOWED_ORIGINS = var.cors_allowed_origins` only to the Go container. It does not set `PORT` or configure CORS on Node.
- `infrastructure/variables.tf` adds the approved non-empty string variable validation.
- `infrastructure/terraform.tfvars.example` uses the safe `https://your-project.web.app` placeholder.
- The ignored local `infrastructure/terraform.tfvars` has the exact allowed origin `https://interseguro-challenge-508112.web.app` (no trailing slash), Go `v2`, and unchanged Node `v1`.

## Automated Results

| Check | Result |
| --- | --- |
| `go test ./...` | PASS |
| `go vet ./...` | PASS |
| `go build ./...` | PASS |
| `go test ./httpapi -run '^TestCORS$' -count=1 -v` | PASS |
| `git diff --check` | PASS |
| `terraform fmt` (implementation result) | PASS |
| `terraform validate` (implementation result) | PASS |
| `terraform plan -var-file=terraform.tfvars` (implementation result) | PASS |
| Go image build/push (implementation result) | PASS: `southamerica-west1-docker.pkg.dev/interseguro-challenge-508112/interseguro-challenge/go-api:v2`, digest `sha256:8ad74cecfe74591ff4ca4c670fb81141e17ee677b6ecc3170d74c933d19d0a78` |

## Terraform Plan Review

The complete reported plan summary is:

```text
Plan: 0 to add, 1 to change, 0 to destroy.
```

Its sole cloud resource change is `google_cloud_run_v2_service.go_api`:

- Go container image: `southamerica-west1-docker.pkg.dev/interseguro-challenge-508112/interseguro-challenge/go-api:v1` to `.../go-api:v2`.
- New Go-container environment variable: `CORS_ALLOWED_ORIGINS = "https://interseguro-challenge-508112.web.app"`.

No Node service/image, Artifact Registry, IAM, service account, URI, Firebase, frontend, or destruction change was reported. This is the expected `0 add / 1 change / 0 destroy` plan.

## QA Environment Note

As an independent follow-up, `terraform fmt -check -diff` passed. A later local `terraform validate` attempt could not start the already-installed Google provider `8.2.0`: Terraform reported `Failed to load plugin schemas` / `Failed to read any lines from plugin's stdout` while launching the `darwin_arm64` provider binary. Because the commands were chained, that follow-up did not execute a second plan. This is an environment/provider-startup failure, not a configuration validation error; the implementation-stage `validate` and plan both completed successfully. Re-run `terraform validate` and the reviewed plan in the operator's working Terraform environment immediately before any apply.

## Problems Found

- No implementation defect found.
- Non-code QA environment issue: subsequent Google provider startup failure described above prevents independent reproduction of Terraform validation/plan in this session. It does not authorize an apply and should be resolved/rechecked before deployment.

## Prioritized Post-Apply Verification

1. Confirm Cloud Run created the Go revision from `go-api:v2` and has `CORS_ALLOWED_ORIGINS` set to exactly `https://interseguro-challenge-508112.web.app`.
2. From the deployed Firebase site, inspect the browser preflight to `OPTIONS /qr`: expect `204`, `Access-Control-Allow-Origin` equal to that exact origin, allowed methods containing `POST, OPTIONS`, and allowed headers containing `Content-Type`.
3. Submit a valid JSON matrix with `POST /qr` and `Content-Type: application/json`; expect the CORS allow-origin response header and successful QR/statistics response.
4. Confirm the Go service continues to call the unchanged Node Cloud Run service and no browser-facing Node CORS configuration was introduced.
5. Send the same POST from an untrusted origin and confirm it does not receive an allow-origin response header.

## Regression Checklist

- [x] Go API behavior and Go-to-Node tests pass locally.
- [x] CORS preflight and trusted/untrusted origin behavior pass in Fiber tests.
- [x] Go source remains unchanged; no unnecessary CORS reimplementation.
- [x] Only approved versioned Terraform files changed.
- [x] Local `terraform.tfvars` remains ignored.
- [x] Go tag is `v2`; Node tag remains `v1`.
- [x] No `terraform apply`, Firebase deployment, Node change, IAM change, or frontend change was performed by this feature.
- [ ] Re-run Terraform validation/plan in a working provider environment before apply.
- [ ] Perform browser end-to-end verification after explicit apply approval.

## Relevant Coverage

`TestCORS` is focused HTTP coverage for the applicable browser boundary: `OPTIONS /qr`, `POST` request method, `Content-Type`, configured-origin allow header, and untrusted-origin denial. Existing Go tests also cover QR behavior, request validation, downstream Node request/response behavior, and downstream error mapping. Cloud Run/Firebase end-to-end behavior necessarily remains unverified until a human-approved apply creates the new Go revision.
