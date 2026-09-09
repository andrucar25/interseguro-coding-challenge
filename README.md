# Interseguro Coding Challenge

Two small APIs process a matrix request end to end:

1. The Go/Fiber API factorizes the input matrix into **Q** and **R**.
2. It sends both matrices to the Node/Express API.
3. The Node service calculates aggregate statistics and returns them with the QR result.

## Requirements

- Docker and Docker Compose (recommended), or
- Node.js 24 and Go 1.27 for local development.

## Run with Docker

From the repository root:

```bash
docker compose up --build
```

The public Go API is available at `http://localhost:8080`. The Node statistics API is published at `http://localhost:3000` for local inspection; the Go service reaches it through the Compose network.

Stop the services with `Ctrl+C`, or run `docker compose down` from another terminal.

## Run locally

Open two terminals.

```bash
# Terminal 1: statistics service
cd node-api
npm ci
cp .env.example .env
npm run dev
```

```bash
# Terminal 2: public QR API
cd go-api
cp .env.example .env
set -a && source .env && set +a
go run .
```

The Node API listens on port `3000` and the Go API listens on port `8080` by default. Configure them with `PORT`; set `NODE_API_URL` in `go-api/.env` to the Node service base URL.

## Use the API

Send a non-empty, rectangular matrix with no more columns than rows:

```bash
curl -X POST http://localhost:8080/qr \
  -H 'Content-Type: application/json' \
  -d '{"matrix":[[1,2],[3,4]]}'
```

The response contains `q`, `r`, and `statistics`. Statistics include `maximum`, `minimum`, `sum`, `average`, and `hasDiagonalMatrix` across both result matrices.

## Development checks

```bash
cd node-api
npm run typecheck && npm run check && npm test

cd ../go-api
go vet ./... && go test ./... && go build ./...
```

## Deploy to Google Cloud

This manual bootstrap deploys both APIs as public Cloud Run services in
`southamerica-west1`. The Node service is intentionally public for this first
deployment phase; a later hardening feature can make it private with
service-to-service IAM.

Prerequisites: Terraform 1.16.1 or newer, the Google Cloud CLI authenticated
with Application Default Credentials and deployment permissions for
`interseguro-challenge-508112`, Docker with Buildx, and a running Docker daemon.
Copy `infrastructure/terraform.tfvars.example` to the ignored local
`infrastructure/terraform.tfvars` before planning. Do not commit it. Terraform
does not build or push images, so the first deployment has a one-time bootstrap
exception to create the APIs and Artifact Registry repository first.

```bash
cd infrastructure
terraform fmt -check
terraform init
terraform validate
terraform plan -var-file=terraform.tfvars

terraform apply -var-file=terraform.tfvars \
  -target='google_project_service.required["run.googleapis.com"]' \
  -target='google_project_service.required["artifactregistry.googleapis.com"]' \
  -target=google_artifact_registry_repository.containers
```

Configure Docker to use the active `gcloud` account for Artifact Registry, then
build and push Linux/amd64 images from the repository root. Use a new explicit
tag for later manual revisions; do not use `latest`.

```bash
gcloud auth configure-docker southamerica-west1-docker.pkg.dev

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

After the images are available, run a normal full apply and read the service
URLs:

```bash
cd infrastructure
terraform plan -var-file=terraform.tfvars
terraform apply -var-file=terraform.tfvars
terraform output artifact_registry_repository_url
terraform output node_service_url
terraform output go_service_url
```

In Postman, send this request to `POST GO_SERVICE_URL/qr` (replace
`GO_SERVICE_URL` with `terraform output -raw go_service_url`):

```json
{"matrix":[[1,2],[3,4]]}
```

It should return HTTP 200 with `q`, `r`, and `statistics`.
