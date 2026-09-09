# Development Plan - gcp-terraform-infrastructure

## Summary

Provision the first Cloud Run deployment phase with Terraform, preserving all existing Go, Node, QR, statistics, and HTTP behavior. The deployment is deliberately a temporary public-to-public topology so the complete container, registry, Cloud Run, and Go-to-Node HTTP path can be validated before service-to-service IAM is introduced.

```text
Internet
   |
   v
Cloud Run Service: interseguro-go-api (public)
   |
   | HTTPS request to NODE_API_URL (no Google ID token in this phase)
   v
Cloud Run Service: interseguro-node-api (public)
```

Both services run explicitly tagged images built outside Terraform and stored in one regional Artifact Registry Docker repository. Cloud Run resolves a tag to a digest for each immutable revision:

```text
southamerica-west1-docker.pkg.dev/interseguro-challenge-508112/interseguro-challenge/
├── go-api:v1
└── node-api:v1
```

The Cloud Run Go container already obtains `PORT` from the environment and requires an absolute `NODE_API_URL`. The Node container already binds to `0.0.0.0` and reads `PORT`. Terraform will only set `NODE_API_URL` on Go to the computed URI of the Node Cloud Run service. It will not set or hardcode `PORT`, and no application source change is planned.

The implementation uses Cloud Run *Services* (`google_cloud_run_v2_service`), not Cloud Run Functions. Terraform owns GCP API enablement, the repository, the two services, their runtime configuration, and outputs. Docker owns build/push; Terraform will have no `null_resource`, `local-exec`, Docker provider, build, push, compilation, or imperative deploy script.

## Files to Create

- `infrastructure/versions.tf`
  - Set `required_version = ">= 1.16.1, < 2.0.0"` and require `hashicorp/google` with `version = "~> 8.1"`.
  - The current stable provider verified during planning is `8.1.0`; this permits compatible future 8.x minor/patch releases but excludes the next major version. Terraform 1.16.1 satisfies the declared Terraform constraint.

- `infrastructure/providers.tf`
  - Configure the `google` provider with `project = var.project_id` and `region = var.region` only. It uses the existing Application Default Credentials; it does not accept a service-account JSON key or credentials file.

- `infrastructure/variables.tf`
  - Declare the small public configuration surface: `project_id`, `region`, `artifact_registry_repository`, `go_service_name`, `node_service_name`, `go_image`, and `node_image`.
  - Validate non-empty values and require explicit image tags rather than `:latest`. Image variables remain complete Artifact Registry references so Terraform never assembles or conceals deployment artifacts.

- `infrastructure/services.tf`
  - Enable exactly `run.googleapis.com` and `artifactregistry.googleapis.com` through a `for_each` set of `google_project_service` resources, with `disable_on_destroy = false`.
  - Do not declare `serviceusage.googleapis.com`: it is enabled by default and is the control-plane API Terraform needs to enable the two product APIs. If it was manually disabled, the operator must re-enable it before Terraform can manage project services.
  - Do not enable Cloud Build, Cloud Functions, Compute Engine, IAM, VPC Access, Secret Manager, or any speculative API.

- `infrastructure/artifact-registry.tf`
  - Define one regional standard Docker `google_artifact_registry_repository`, dependent on Artifact Registry API enablement, with `location = var.region`, `repository_id = var.artifact_registry_repository`, and `format = "DOCKER"`.
  - No cleanup policy, KMS key, replication, IAM binding, or second repository is needed in this phase.

- `infrastructure/cloud-run.tf`
  - Define `google_cloud_run_v2_service.node_api` and `google_cloud_run_v2_service.go_api`, each dependent on the required API/repository resources.
  - Put `NODE_API_URL = google_cloud_run_v2_service.node_api.uri` in Go's container `env` block. This output reference establishes Node-before-Go deployment ordering naturally.

- `infrastructure/outputs.tf`
  - Output `artifact_registry_repository_url`, `go_service_url`, and `node_service_url`; values are non-secret.

- `infrastructure/terraform.tfvars.example`
  - Provide the exact non-secret bootstrap/manual values and tagged image-reference pattern, without creating a real `terraform.tfvars`.

- `.gitignore`
  - Add repository-level Terraform state exclusions: `.terraform/`, `*.tfstate`, `*.tfstate.*`, and `terraform.tfvars`.
  - Do not ignore `infrastructure/.terraform.lock.hcl`; it must be committed after `terraform init` pins the selected provider checksums.

## Files to Modify

- `README.md`
  - Add a concise infrastructure deployment section after implementation: prerequisites, bootstrap exception, Docker credential-helper setup, Linux/amd64 build-and-push commands, full Terraform apply, outputs, and the `/qr` Postman request. This is operational documentation, not CI/CD documentation.

No file below will be modified: `go-api/**`, `node-api/**`, Dockerfiles, Docker Compose, API contracts, dependency manifests, `docs/architecture.md`, or `AGENTS.md`. The existing applications already satisfy the Cloud Run runtime requirements relevant to this phase.

## Execution Order

### 1. Implement and check Terraform configuration (after approval)

1. Create only the infrastructure files and root `.gitignore` entry above. Do not create Terraform modules, a remote backend, or GCP credentials files.
2. Run from `infrastructure/`:

   ```bash
   terraform fmt -check
   terraform init
   terraform validate
   terraform plan -var-file=terraform.tfvars
   ```

3. Commit `.terraform.lock.hcl`; do not commit `.terraform/`, any state file, or `terraform.tfvars`.

`terraform.tfvars` is local and must be created from `terraform.tfvars.example` before the first plan. It can already contain the final Artifact Registry image paths with `:v1`, even though the repository/images do not yet exist. Planning accepts those string references; only an attempt to create Cloud Run before pushing would fail.

### 2. Bootstrap only the APIs and Artifact Registry

The image/service dependency is intentionally resolved in a one-time targeted bootstrap:

```text
Artifact Registry repository must exist
    before images can be pushed
    before Cloud Run services can use those image URLs.
```

After reviewing the normal plan, execute this exceptional bootstrap command from `infrastructure/`:

```bash
terraform apply -var-file=terraform.tfvars \
  -target='google_project_service.required["run.googleapis.com"]' \
  -target='google_project_service.required["artifactregistry.googleapis.com"]' \
  -target=google_artifact_registry_repository.containers
```

Targeting is not the normal Terraform workflow. It is used only once to break the Artifact Registry/image/Cloud Run bootstrap cycle. All subsequent changes, including the immediately following Cloud Run deployment, use a full `terraform plan` followed by a full `terraform apply`, with no `-target`.

### 3. Configure Artifact Registry Docker authentication

With the repository created, configure Docker's `gcloud` credential helper from the monorepo root or any shell on the same machine:

```bash
gcloud auth configure-docker southamerica-west1-docker.pkg.dev
```

This registers the `gcloud` credential helper in Docker's configuration; Docker uses the active `gcloud` account for Artifact Registry authentication. This is distinct from Application Default Credentials, which the Google Terraform Provider uses. Do not use saved passwords, manually piped access tokens, or service-account JSON keys.

### 4. Build and push the two Linux/amd64 images

The inspected local Docker client is 29.4.0 with Buildx v0.33.0, which supports this direct `buildx build --platform ... --push` workflow. On the Apple Silicon workstation, `--platform linux/amd64` is mandatory so Cloud Run receives a Linux x86_64 image rather than a local ARM-only image.

From the monorepo root, use bootstrap tag `v1` for the first manual deployment:

```bash
docker buildx build \
  --platform linux/amd64 \
  --tag southamerica-west1-docker.pkg.dev/interseguro-challenge-508112/interseguro-challenge/node-api:v1 \
  --push \
  ./node-api

docker buildx build \
  --platform linux/amd64 \
  --tag southamerica-west1-docker.pkg.dev/interseguro-challenge-508112/interseguro-challenge/go-api:v1 \
  --push \
  ./go-api
```

`v1` is selected rather than Git SHA because this is the explicit bootstrap/manual deployment and there is no CI/CD yet. It is clear and explicit, unlike `latest`. Later manual revisions must use a new explicit tag (for example `v2`); a future CI/CD feature can switch the same variables to Git-SHA tags without changing Terraform structure. For stronger immutability later, variables may be supplied as `@sha256:...` image digests after push, but that is not required for this first execution.

Verify the pushed artifacts before service deployment:

```bash
gcloud artifacts docker images list \
  southamerica-west1-docker.pkg.dev/interseguro-challenge-508112/interseguro-challenge/node-api \
  --include-tags

gcloud artifacts docker images list \
  southamerica-west1-docker.pkg.dev/interseguro-challenge-508112/interseguro-challenge/go-api \
  --include-tags
```

### 5. Execute a normal full apply

Keep the same complete tagged image references in local `infrastructure/terraform.tfvars`, then run:

```bash
cd infrastructure
terraform plan -var-file=terraform.tfvars
terraform apply -var-file=terraform.tfvars
```

This full apply creates Node first, reads its generated `uri`, injects it as `NODE_API_URL` into Go, then deploys Go. No `PORT`, `NODE_API_AUDIENCE`, token, service account, IAM member, VPC, secret, database, load balancer, or gateway is configured.

### 6. Verify deployment and request flow

Read the outputs:

```bash
terraform output artifact_registry_repository_url
terraform output node_service_url
terraform output go_service_url
```

Directly verify that the temporarily public Node service answers its existing endpoint. Replace `NODE_SERVICE_URL` with `terraform output -raw node_service_url` if not exporting it in the shell:

```bash
curl --fail --request POST "NODE_SERVICE_URL/api/v1/statistics" \
  --header 'Content-Type: application/json' \
  --data '{"q":[[1,0],[0,1]],"r":[[2,3],[0,4]]}'
```

In Postman, issue the end-to-end request:

```text
POST GO_SERVICE_URL/qr
Content-Type: application/json

{"matrix":[[1,2],[3,4]]}
```

Expect HTTP 200 and JSON containing `q`, `r`, and `statistics` with `maximum`, `minimum`, `sum`, `average`, and `hasDiagonalMatrix`. This proves:

```text
Postman → public Go Cloud Run → QR → public Node Cloud Run → statistics → Go → Postman
```

For basic operational debugging, list and inspect services:

```bash
gcloud run services list \
  --project=interseguro-challenge-508112 \
  --region=southamerica-west1

gcloud run services describe interseguro-node-api \
  --project=interseguro-challenge-508112 \
  --region=southamerica-west1

gcloud run services describe interseguro-go-api \
  --project=interseguro-challenge-508112 \
  --region=southamerica-west1
```

If a container fails to start or Go returns a downstream error, inspect revision logs for the relevant service:

```bash
gcloud logging read \
  'resource.type="cloud_run_revision" AND resource.labels.service_name="interseguro-node-api"' \
  --project=interseguro-challenge-508112 \
  --limit=50 \
  --format='value(textPayload)'

gcloud logging read \
  'resource.type="cloud_run_revision" AND resource.labels.service_name="interseguro-go-api"' \
  --project=interseguro-challenge-508112 \
  --limit=50 \
  --format='value(textPayload)'
```

## Architecture Decisions

### GCP resources, names, and required APIs

| Terraform resource | Proposed name/value | Purpose |
| --- | --- | --- |
| `google_project_service.required["run.googleapis.com"]` | Cloud Run Admin API | Create and manage Cloud Run Services. |
| `google_project_service.required["artifactregistry.googleapis.com"]` | Artifact Registry API | Create the Docker repository and host images. |
| `google_artifact_registry_repository.containers` | `interseguro-challenge` in `southamerica-west1` | One standard regional Docker repository for both services. |
| `google_cloud_run_v2_service.node_api` | `interseguro-node-api` | Public temporary downstream statistics service. |
| `google_cloud_run_v2_service.go_api` | `interseguro-go-api` | Public client-facing QR API. |

Only `run.googleapis.com` and `artifactregistry.googleapis.com` are product APIs required by this design. The Service Usage API is normally enabled for a project and is a prerequisite control-plane API for `google_project_service`, not an additional workload API to manage. The deployer must already have the relevant project permissions (notably Service Usage administration, Artifact Registry administration/write access, and Cloud Run deployment permissions); this plan does not create broad IAM grants to obtain them.

One repository is preferable here because both small images share a lifecycle, region, project, and manual deployment procedure. Separate repositories add names, policy surface, and bootstrap work without a current isolation, retention, or ownership requirement.

### Provider, variables, outputs, and local state

Example local variables:

```hcl
project_id                   = "interseguro-challenge-508112"
region                       = "southamerica-west1"
artifact_registry_repository = "interseguro-challenge"
go_service_name              = "interseguro-go-api"
node_service_name            = "interseguro-node-api"
go_image                     = "southamerica-west1-docker.pkg.dev/interseguro-challenge-508112/interseguro-challenge/go-api:v1"
node_image                   = "southamerica-west1-docker.pkg.dev/interseguro-challenge-508112/interseguro-challenge/node-api:v1"
```

The outputs are:

| Output | Value |
| --- | --- |
| `artifact_registry_repository_url` | `southamerica-west1-docker.pkg.dev/interseguro-challenge-508112/interseguro-challenge` constructed from the managed repository. |
| `node_service_url` | `google_cloud_run_v2_service.node_api.uri`. |
| `go_service_url` | `google_cloud_run_v2_service.go_api.uri`. |

Use the default local Terraform state for this individual challenge. There is no GCS backend, HCP Terraform, remote state, workspace scheme, or backend block. State and tfvars are ignored; `.terraform.lock.hcl` is versioned for reproducible provider selection.

### Cloud Run configuration

Both services use the same deliberately modest runtime baseline:

| Setting | Go | Node | Rationale |
| --- | --- | --- | --- |
| Location | `southamerica-west1` | `southamerica-west1` | Matches the repository and requested region. |
| Image | `var.go_image` | `var.node_image` | Full, explicitly tagged Artifact Registry references; no `latest`. |
| Public access | `invoker_iam_disabled = true` | `invoker_iam_disabled = true` | Current Cloud Run/Terraform-supported way to disable the Invoker IAM check; no legacy `allUsers` binding. |
| Ingress | `INGRESS_TRAFFIC_ALL` | `INGRESS_TRAFFIC_ALL` | Allows public Internet validation. |
| Container environment | `NODE_API_URL = node_api.uri` | none | Go derives the Node endpoint from Terraform; Cloud Run supplies `PORT` to both. |
| CPU limit | `1` vCPU | `1` vCPU | Small, uncomplicated allocation suitable for Go QR and Node runtime startup. |
| Memory limit | `512Mi` | `512Mi` | Small supported baseline with 1 vCPU; avoids forcing a first-generation execution environment merely to use 256Mi. |
| CPU allocation | `cpu_idle = true` | `cpu_idle = true` | Request-based/default CPU billing; explicitly preserve the Cloud Run default when resource limits are declared. |
| Minimum instances | `0` | `0` | Scales to zero and avoids idle-instance cost. |
| Maximum instances | `3` | `3` | A small accidental-cost guard appropriate to the challenge; it can be increased later if real traffic demonstrates a need. |
| Deletion protection | `deletion_protection = false` | `deletion_protection = false` | Both resources must be destroyable with Terraform in this technical-challenge environment. |
| Service account resources | none | none | No custom runtime identity or IAM work is needed while invocation is unauthenticated. |
| Other networking | none | none | No VPC, load balancer, API Gateway, or private networking. |

Both `google_cloud_run_v2_service` resources explicitly set `deletion_protection = false`, so this technical-challenge environment can be destroyed with Terraform. `template.scaling { min_instance_count = 0; max_instance_count = 3 }` is revision-level autoscaling. It retains Cloud Run automatic scaling, not manual scaling. Container concurrency, startup CPU boost, probes, and execution-environment overrides remain provider/Cloud Run defaults because the existing containers are simple stateless HTTP processes and no evidence requires tuning them.

The services are intentionally public for initial deployment validation. In particular, Node is public temporarily so Go can make normal HTTPS calls to its service URI without a Google ID token. This is not treated as a final security posture.

### Future feature — `cloud-run-private-node` (not part of this plan)

After validating the complete deployment, a separately approved hardening feature will:

- make Node private by re-enabling the Invoker IAM check;
- create or select a Go runtime service identity;
- grant only that identity `roles/run.invoker` on Node;
- have Go obtain and send a Google ID token for Node's audience;
- verify unauthenticated Node requests are denied.

This phase must not pre-create that identity, add `roles/run.invoker`, set `NODE_API_AUDIENCE`, request Google ID tokens, or attempt private Node access.

## Validation Criteria

The implementation is complete only when all of the following hold:

1. Existing application quality gates still pass without functional source changes:

   ```bash
   cd go-api
   test -z "$(find . -type f -name '*.go' -exec gofmt -l {} +)"
   go vet ./...
   go test ./...
   go build ./...

   cd ../node-api
   npm run typecheck
   npm run check
   npm test
   npm run build
   ```

2. From `infrastructure/`, `terraform fmt -check`, `terraform init`, `terraform validate`, and a normal `terraform plan -var-file=terraform.tfvars` succeed. Planner work does not run `terraform apply`.
3. The single exceptional target apply creates only the required APIs and Artifact Registry repository; it does not build/push images or create Cloud Run services.
4. `gcloud auth configure-docker southamerica-west1-docker.pkg.dev` succeeds and both Buildx commands build and push Linux/amd64 tagged images.
5. The normal, non-targeted Terraform plan/apply creates both Cloud Run Services. `terraform output` returns a repository URL and two HTTPS service URLs.
6. A direct unauthenticated `POST` to Node `/api/v1/statistics` returns its existing successful statistics response.
7. The Postman request to Go `/qr` returns 200 with QR matrices and statistics, proving the Go-to-Node Cloud Run HTTP call.
8. Both Cloud Run services show `min_instance_count = 0`, `max_instance_count = 3`, 1 CPU, `512Mi`, and public invoker IAM disabled. No service-to-service IAM binding, custom service account, Google ID token, VPC, load balancer, secret, database, or CI/CD resource appears in the Terraform plan/state.

Official references used for implementation decisions:

- [Terraform Google provider registry](https://registry.terraform.io/providers/hashicorp/google/latest/docs) — current stable provider `8.1.0` at planning time.
- [Cloud Run public access](https://cloud.google.com/run/docs/authenticating/public) — `invoker_iam_disabled = true` on `google_cloud_run_v2_service`.
- [Cloud Run service deployment](https://cloud.google.com/run/docs/deploying) — Cloud Run Service image use and Artifact Registry image reference format.
- [Cloud Run minimum instances](https://cloud.google.com/run/docs/configuring/min-instances), [CPU](https://cloud.google.com/run/docs/configuring/services/cpu), and [memory limits](https://cloud.google.com/run/docs/configuring/services/memory-limits) — Terraform scaling and resource syntax/limits.
- [Artifact Registry Terraform resources](https://cloud.google.com/artifact-registry/docs/repositories/terraform) and [Docker push/pull authentication](https://cloud.google.com/artifact-registry/docs/docker/pushing-and-pulling) — regional Docker repository and `gcloud auth configure-docker` workflow.

## External Dependencies

- Existing local Terraform `1.16.1`.
- Existing Google Cloud CLI `583.0.0`, configured for project `interseguro-challenge-508112` and Application Default Credentials. No service-account key is used.
- Existing Docker client `29.4.0` and Docker Buildx `v0.33.0`. A reachable Docker daemon remains required when executing the approved build/push commands; the planner did not start builds.
- Google Cloud project permissions for the person running Terraform/builds: ability to enable the two APIs, create/manage the Artifact Registry repository, push images, and deploy Cloud Run services. Same-project Cloud Run/Artifact Registry avoids cross-project image-reader IAM configuration.
- Internet access to pull Docker base images and Terraform provider packages during the approved implementation/deployment workflow.

No new application library, Terraform module, CI/CD provider, state backend, secret manager, database, network component, service account, or IAM binding is introduced.

## Execution Results

The user explicitly approved this design for implementation on 2026-09-09.

### Completed files

- Created `infrastructure/versions.tf`, `providers.tf`, `variables.tf`, `services.tf`, `artifact-registry.tf`, `cloud-run.tf`, `outputs.tf`, and `terraform.tfvars.example`.
- Created root `.gitignore` Terraform local-state exclusions without ignoring `.terraform.lock.hcl`.
- Updated `README.md` with the approved manual bootstrap, Linux/amd64 image push, full apply, outputs, and Postman request workflow.

### Validation

- Passed: `git diff --check` static whitespace check.
- Deliberately deferred: `terraform fmt -check`, `terraform init`, `terraform validate`, and `terraform plan -var-file=terraform.tfvars`.
- Deliberately not run: any Terraform apply, `gcloud`/Google command, Docker build/push, or infrastructure-changing command, per the user's request to review created files first.
- No `terraform.tfvars` or `.terraform.lock.hcl` was created. The lock file will be generated and committed only after a future `terraform init`.

### Blocked work

Runtime validation and deployment remain pending user review and authorization to run the deferred commands.

### Out-of-scope suggestions

- None.

## Approved Decisions

The approved implementation follows these decisions:

1. Use a single `interseguro-challenge` Docker repository in `southamerica-west1`, with explicit `v1` bootstrap image tags and full image references supplied through `go_image`/`node_image`.
2. Use Terraform `>= 1.16.1, < 2.0.0` and `hashicorp/google ~> 8.1`, commit the generated lock file, and keep local state with the specified root `.gitignore` rules.
3. Enable only Cloud Run and Artifact Registry product APIs through `google_project_service`; rely on the default Service Usage API rather than managing unrelated APIs.
4. Use the one-time targeted bootstrap apply exclusively for APIs/repository, followed by Docker credential-helper setup, Linux/amd64 Buildx pushes, and a normal complete Terraform apply.
5. Deploy both `interseguro-go-api` and `interseguro-node-api` publicly with `invoker_iam_disabled = true` and normal HTTPS, explicitly accepting temporary public Node access while validating the deployment.
6. Use `NODE_API_URL = google_cloud_run_v2_service.node_api.uri`, with no `NODE_API_AUDIENCE`, token, `roles/run.invoker`, custom service account, or private networking.
7. Use 1 vCPU, `512Mi`, request-based CPU allocation, min instances `0`, and max instances `3` for each service.
8. Make only concise README operational-documentation updates; do not implement CI/CD or the future private-Node hardening in this feature.
