# Development Plan - dockerization

## Summary

Create one independent, multi-stage Docker image per service, built from each service directory:

```text
Client → Go/Fiber → QR → HTTP → Node/Express → statistics → Go/Fiber → Client
```

Docker changes packaging and runtime configuration only; it does not change QR, statistics, or the HTTP contract. The images receive configuration at runtime. Go keeps `NODE_API_URL` mandatory, so a missing downstream address fails clearly at startup instead of silently using a local-only URL.

The design uses the `go-fiber-development`, `express-typescript-development`, and `backend-testing` guidance. `rest-api-design` is intentionally not used: no endpoint or Go-to-Node contract changes.

## Files to Create

- `go-api/Dockerfile`
  - Multi-stage Linux build, with `golang:1.27.1-alpine3.22` as the builder (matching `go.mod`) and `gcr.io/distroless/static-debian12:nonroot` as runtime.
  - Build sequence: copy `go.mod`/`go.sum`; `go mod download`; copy application source; cross-compile a static binary with `CGO_ENABLED=0`, `GOOS=$TARGETOS`, and `GOARCH=$TARGETARCH`; copy only that binary into the runtime image.
  - Intended content:

    ```dockerfile
    # syntax=docker/dockerfile:1
    FROM --platform=$BUILDPLATFORM golang:1.27.1-alpine3.22 AS build
    WORKDIR /src

    ARG TARGETOS
    ARG TARGETARCH

    COPY go.mod go.sum ./
    RUN go mod download

    COPY . .
    RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
        go build -trimpath -ldflags="-s -w" -o /out/go-api .

    FROM gcr.io/distroless/static-debian12:nonroot
    COPY --from=build /out/go-api /go-api

    EXPOSE 8080
    USER nonroot:nonroot
    ENTRYPOINT ["/go-api"]
    ```

- `go-api/.dockerignore`
  - Exclude `.env`, `.env.*`, `.git`, coverage, local Go binaries/test artifacts, logs, temporary files, and editor metadata. Retain `go.mod`, `go.sum`, and all production source. Use `!.env.example` after the `.env.*` rule so the example remains repository documentation (although it is not copied by this Dockerfile).
  - Proposed patterns:

    ```text
    .env
    .env.*
    !.env.example
    .git
    coverage
    *.coverprofile
    bin
    *.exe
    *.test
    *_test.go
    *.log
    tmp
    .DS_Store
    .idea
    .vscode
    ```

- `go-api/.env.example`
  - Documentation-only local convenience file, not read automatically by Go:

    ```dotenv
    PORT=8080
    NODE_API_URL=http://localhost:3000
    ```

- `node-api/Dockerfile`
  - Three stages, all based on the explicit Node version from `.nvmrc`/`engines`: `node:24.16.0-alpine3.22`.
  - `build`: cache `npm ci` behind the lockfiles, then copy production TypeScript source/config and run `npm run build`.
  - `production-dependencies`: independently run `npm ci --omit=dev`, so the runtime has only production packages.
  - `runtime`: copy only production `node_modules` and `dist/`, change ownership during copying, run with the image's built-in `node` user, and execute compiled JavaScript directly.
  - Intended content:

    ```dockerfile
    # syntax=docker/dockerfile:1
    FROM node:24.16.0-alpine3.22 AS build
    WORKDIR /app

    COPY package.json package-lock.json ./
    RUN npm ci

    COPY tsconfig.json ./
    COPY src ./src
    RUN npm run build

    FROM node:24.16.0-alpine3.22 AS production-dependencies
    WORKDIR /app

    COPY package.json package-lock.json ./
    RUN npm ci --omit=dev && npm cache clean --force

    FROM node:24.16.0-alpine3.22 AS runtime
    WORKDIR /app

    COPY --from=production-dependencies --chown=node:node /app/node_modules ./node_modules
    COPY --from=build --chown=node:node /app/dist ./dist

    EXPOSE 8080
    USER node
    CMD ["node", "dist/index.js"]
    ```

- `node-api/.dockerignore`
  - Exclude local dependencies, build/coverage output, actual `.env` files, Git data, logs, temporary files, editor metadata, and test sources not needed for `tsc` production output. Keep lockfiles, source, and `.env.example`.
  - Proposed patterns:

    ```text
    node_modules
    dist
    coverage
    .env
    .env.*
    !.env.example
    .git
    *.log
    npm-debug.log*
    tmp
    .DS_Store
    .idea
    .vscode
    src/**/*.test.ts
    ```

- `node-api/.env.example`
  - Documentation-only local convenience file, not read automatically by Node:

    ```dotenv
    PORT=3000
    ```

## Files to Modify

- `node-api/src/index.ts`
  - Change the fallback from `3000` to `8080` and call `app.listen(port, "0.0.0.0", ...)` explicitly. `PORT` continues to be the only Node environment variable; no `NODE_ENV`, `APP_ENV`, logging variable, or `.env` loader is introduced.
  - This is the only application-code change. It gives the container a Cloud Run-compatible default while callers still choose a different local port by setting `PORT=3000`.

No Go source change is planned. `go-api/main.go` already defaults `PORT` to `8080`, passes `":" + port` to Fiber (all interfaces), and rejects an empty or invalid `NODE_API_URL` through `statistics.New` before it starts listening.

The existing `node-api/.gitignore` already excludes actual Node `.env` files and explicitly permits `.env.example`. This approved scope does not add a Go `.gitignore`; no real Go `.env` file is to be created or committed. Its `.dockerignore` prevents it entering a build context, and repository policy remains that any real local Go `.env` must stay untracked. If a version-controlled Git guard for a future Go `.env` is desired, approve it as a small follow-up rather than expanding this feature's declared file scope.

## Execution Order

1. Inspect the worktree before editing and confirm the runtime observations above remain true.
2. Add both Dockerfiles and their build-context-specific `.dockerignore` files. Do not copy `.env`, credentials, source artifacts, tests, build tools, or dependency manifests into final images except Node's runtime `node_modules`.
3. Add the two documentation-only `.env.example` files; do not create `.env` files and do not add automatic dotenv loading. Developers may export/source their variables by their own local workflow.
4. Make the minimal Node startup change (`PORT` fallback and explicit host). Do not change Go's configuration behavior or either HTTP contract.
5. Run the existing quality gates before image validation.
6. Build each image independently from the monorepo root. Use a `linux/amd64` target for Cloud Run: its runtime contract requires Linux x86_64. The Dockerfile's BuildKit `TARGETOS`/`TARGETARCH` arguments allow Go to cross-compile without emulating its compiler.

   ```bash
   docker build --platform linux/amd64 -t interseguro-go-api ./go-api
   docker build --platform linux/amd64 -t interseguro-node-api ./node-api
   ```

   On an ARM development host, Docker Desktop can run these images through its configured emulation; an equivalent native-architecture image is acceptable for purely local testing, but the image intended for Cloud Run must include `linux/amd64`.

7. Smoke-test Node on its own:

   ```bash
   docker run --rm -p 3000:8080 -e PORT=8080 interseguro-node-api
   curl --fail --request POST http://localhost:3000/api/v1/statistics \
     --header 'Content-Type: application/json' \
     --data '{"q":[[1,0],[0,1]],"r":[[2,3],[0,4]]}'
   ```

8. Smoke-test Go against a Node process running on the host. Start Node locally with `PORT=3000 npm run dev` (or `npm start` after `npm run build`), then run Go with its address supplied only as runtime configuration:

   ```bash
   docker run --rm -p 8080:8080 \
     -e PORT=8080 \
     -e NODE_API_URL=http://host.docker.internal:3000 \
     interseguro-go-api
   ```

   `host.docker.internal` is not put in source or a Dockerfile. On Linux hosts that do not provide it, add `--add-host=host.docker.internal:host-gateway`; on another platform use a reachable host address as the environment variable value.

9. Also smoke-test the two images without Compose, using an explicit throwaway Docker network. This proves the future Compose service DNS value without creating the deferred Compose feature:

   ```bash
   docker network create interseguro-smoke
   docker run --rm -d --name node-api --network interseguro-smoke \
     -e PORT=8080 interseguro-node-api
   docker run --rm -p 8080:8080 --network interseguro-smoke \
     -e PORT=8080 \
     -e NODE_API_URL=http://node-api:8080 \
     interseguro-go-api
   ```

   Exercise `POST /qr` on `http://localhost:8080`; after stopping the Node container, remove the explicit `interseguro-smoke` network. Docker Compose itself, Artifact Registry push, `gcloud`, Terraform, IAM, Cloud Run deployment, health endpoints, and `HEALTHCHECK` remain out of scope.

## Architecture Decisions

### Runtime configuration and ports

| Service | Environment variables | Default/listen behavior | Image `EXPOSE` |
| --- | --- | --- | --- |
| Go/Fiber | `PORT` (optional, default `8080`); `NODE_API_URL` (required absolute HTTP(S) URL) | `":" + PORT`, therefore all interfaces | `8080` |
| Node/Express | `PORT` (optional, default changed to `8080`) | explicit `0.0.0.0:PORT` | `8080` |

`EXPOSE 8080` documents the expected container port; it does not publish a host port. Containers can both use `8080` because they have separate network namespaces. Neither image bakes in `localhost`, a Compose DNS name, or a Cloud Run URL.

This yields the same immutable images with only runtime configuration changed:

- local Go → local Node: `NODE_API_URL=http://localhost:3000`;
- future Compose Go → Node container: `NODE_API_URL=http://node-api:8080`;
- future Cloud Run Go → Cloud Run Node service: `NODE_API_URL=https://<node-cloud-run-service>`.

### Multi-stage and base-image trade-offs

Multi-stage builds keep compilers, TypeScript, `tsx`, Biome, Vitest, Supertest, test source, dev dependencies, and application source out of runtime layers. Copying Go module files and Node lockfiles before application source preserves the expensive dependency-install cache when only source changes. The explicit non-`latest` tags make the selected language/runtime versions reviewable; a later supply-chain policy can pin them by digest without changing the application design.

For Go, recommend distroless static Debian 12 over `scratch` and Alpine runtime. `scratch` is marginally smaller but leaves no CA certificates for a permitted `https://` `NODE_API_URL` and is harder to diagnose. Alpine is more interactive but contains a shell/package base not needed by a statically linked binary. Distroless static supplies the required minimal runtime and CA certificates, is small, and has a purpose-built `nonroot` identity. The trade-off is no shell in the production image, acceptable for this small HTTP service; debugging belongs in the builder or a separate future debug workflow.

For Node, use the official Alpine flavor because this pure TypeScript/JavaScript service needs the Node runtime and production modules, and its official image already provides the `node` non-root user. Adding a separate distroless Node runtime would add less-familiar complexity with limited benefit for this challenge. No OS packages are installed in either image.

### Non-root, lifecycle, security, and state

Go runs as `nonroot:nonroot`; Node runs as its image's `node` user and runtime files are copied with matching ownership. Each container has one foreground process (`/go-api` or `node dist/index.js`); no shell wrapper, `pm2`, `supervisord`, or `systemd` is used.

No Dockerfile `ENV` sets secrets or environment-specific URLs, and neither Dockerfile copies `.env`, GCP credentials, service-account JSON, tokens, or secrets. Actual values are supplied by `docker run`, a later Compose configuration, or later Cloud Run environment/secret configuration. No TLS is configured in containers; Cloud Run terminates TLS. The services remain stateless and do not rely on persistent filesystem data, matching Cloud Run's ephemeral in-memory filesystem.

Cloud Run compatibility is achieved by producing Linux/amd64 images, listening on `0.0.0.0` and the injected `PORT` (with `8080` as sensible non-Cloud default), keeping the main process in the foreground, and avoiding TLS and persistent local state. Cloud Run Services—not Functions—can deploy a compliant container image later; that deployment is deliberately not part of this feature.

Docker Compose is deferred because it is orchestration/runtime wiring, not a property of a single reusable image. The images already support its future `node-api:8080` service address exclusively through `NODE_API_URL`.

## Validation Criteria

Run from each project directory before Docker testing:

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

Then verify the two independent `docker build --platform linux/amd64` commands succeed, Node starts on `PORT=8080` and answers its existing statistics endpoint, Go starts on `PORT=8080` when given a valid `NODE_API_URL`, and Go successfully completes an existing `/qr` request against both a locally run Node service and a separately containerized Node service. Verify a missing/invalid `NODE_API_URL` still stops Go at startup with its current clear configuration error. No new unit-test framework, health endpoint, health check, or Docker-in-Docker automated test is proposed; the focused runtime smoke tests prove the new boundary.

## External Dependencies

No application dependency, service, registry, deployment, or IaC dependency is added.

The Dockerfiles use the existing Go and Node ecosystems plus these externally pulled base images at build time:

- `golang:1.27.1-alpine3.22`
- `gcr.io/distroless/static-debian12:nonroot`
- `node:24.16.0-alpine3.22`

The design follows [Docker build best practices](https://docs.docker.com/build/building/best-practices/), [Docker multi-stage builds](https://docs.docker.com/build/building/multi-stage/), and [Docker platform build arguments](https://docs.docker.com/build/building/variables/). It follows the [Cloud Run container runtime contract](https://cloud.google.com/run/docs/container-contract) for Linux x86_64, `PORT`, `0.0.0.0`, TLS termination, and ephemeral filesystem behavior. Future Cloud Run Services can deploy a compliant image from Artifact Registry; Google recommends Artifact Registry for such images in its [container-image deployment guidance](https://cloud.google.com/run/docs/deploying), but no registry setup or push is included here.

## Blocked Tasks

There is no technical blocker. Implementation is intentionally blocked pending explicit human approval of this `ai-work/dockerization/design.md`.

Approval confirms these decisions:

1. Use `golang:1.27.1-alpine3.22` → `gcr.io/distroless/static-debian12:nonroot` for Go, accepting a shell-less but CA-capable non-root runtime rather than Alpine or `scratch`.
2. Use `node:24.16.0-alpine3.22` in all Node stages, keep the built-in `node` user, and use a separate `npm ci --omit=dev` dependency stage.
3. Change only Node startup to default `PORT=8080` and bind `0.0.0.0`; preserve Go's required `NODE_API_URL` behavior and all HTTP/QR/statistics behavior.
4. Create the two `.env.example` documentation files while not creating or auto-loading real `.env` files.
5. Keep Docker Compose, health checks/endpoints, Artifact Registry, Cloud Run deployment, IAM, Terraform, and CI as later features.
