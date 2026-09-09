# Dockerization

The Go and Node APIs are packaged as independent multi-stage images. Build
each image from its service directory (the commands below are run from the
repository root):

```bash
docker build --platform linux/amd64 -t interseguro-go-api ./go-api
docker build --platform linux/amd64 -t interseguro-node-api ./node-api
```

The `linux/amd64` target is the image contract intended for Cloud Run. The Go
build stage uses the requested target OS and architecture to produce a static
binary; the final Go image is distroless and runs as `nonroot`. The Node image
contains only compiled `dist/` output and production dependencies and runs as
the image-provided `node` user. Docker Compose, registry publishing, Cloud Run
deployment, IAM, and health checks are intentionally deferred.

## Runtime configuration

Both services listen on all interfaces and expose container port `8080`.
Values are injected at runtime; neither image contains environment-specific
URLs or secrets.

| Service | Variables | Container behavior |
| --- | --- | --- |
| Go/Fiber | `PORT` (optional, default `8080`), `NODE_API_URL` (required absolute `http://` or `https://` URL) | Listens on `:${PORT}` and calls Node through `NODE_API_URL`. |
| Node/Express | `PORT` (optional, default `8080`) | Listens on `0.0.0.0:${PORT}`. |

The example files [`go-api/.env.example`](/Users/andres/code/challenges/interseguro-challenge/go-api/.env.example)
and [`node-api/.env.example`](/Users/andres/code/challenges/interseguro-challenge/node-api/.env.example)
document local values only. They are not loaded automatically. Actual `.env`
files are excluded from both Docker build contexts.

## Running the images

Run Node independently and publish its container port to a local port:

```bash
docker run --rm -p 3000:8080 \
  -e PORT=8080 \
  interseguro-node-api
```

With Node running, Go can call it through a host-reachable address:

```bash
docker run --rm -p 8080:8080 \
  -e PORT=8080 \
  -e NODE_API_URL=http://host.docker.internal:3000 \
  interseguro-go-api
```

On Linux, add `--add-host=host.docker.internal:host-gateway` if that hostname
is not provided by the local Docker installation. Alternatively, set
`NODE_API_URL` to another reachable host address.

For an image-to-image smoke test, use a temporary user-created Docker network
and the Node container name as DNS address:

```bash
docker network create interseguro-smoke
docker run --rm -d --name node-api --network interseguro-smoke \
  -e PORT=8080 interseguro-node-api
docker run --rm -p 8080:8080 --network interseguro-smoke \
  -e PORT=8080 \
  -e NODE_API_URL=http://node-api:8080 \
  interseguro-go-api
```

Exercise the existing `POST /qr` endpoint at `http://localhost:8080` while Go
is running. Stop the containers and remove `interseguro-smoke` when finished.

## Validation status

The repository quality commands remain unchanged: Go formatting/vet/tests/
build from `go-api/`, and Node typecheck/check/tests/build from `node-api/`.
Docker image builds and container smoke tests could not be completed in this
environment because the Docker daemon was unavailable. They should be run on a
host with Docker/BuildKit enabled before publishing or deploying either image.

