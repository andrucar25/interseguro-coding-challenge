# Development Plan - go-node-integration

## Summary

Integrate the existing public Go/Fiber `POST /qr` operation with the existing
Node statistics HTTP API.  Go will retain ownership of parsing the original
matrix and calling the pure `qr.Factorize` core.  After a successful
factorization it will use one reusable standard-library HTTP client to send
the resulting matrices to Node, then combine the existing QR result with the
statistics returned by Node.

The inspected Node contract is already sufficient; this feature changes no
Node source, dependency, configuration, or test files.  The actual downstream
contract is:

```text
POST {NODE_API_URL}/api/v1/statistics
Content-Type: application/json

{"q": [[...]], "r": [[...]]}
```

with `200 OK`:

```json
{
  "maximum": 4,
  "minimum": 0,
  "sum": 11,
  "average": 1.375,
  "hasDiagonalMatrix": true
}
```

Node validates both matrices as non-empty, rectangular matrices of finite
numbers and responds with `400 {"error":"invalid_request"}` for invalid
payloads.  Its existing statistics core calculates aggregates over every
scalar of both matrices and applies its diagonal tolerance.  Since Go's QR
core always produces compatible finite, non-empty `Q` and `R`, no contract
incompatibility was found and no Node modification is proposed.

The public `POST /qr` success response will preserve the existing top-level
`q` and `r` fields and add the Node result under `statistics`:

```json
{
  "q": [[-0.8571428571, 0.3942857143], [-0.4285714286, -0.9028571429], [0.2857142857, -0.1714285714]],
  "r": [[-14, -21], [0, -175]],
  "statistics": {
    "maximum": 0.3942857143,
    "minimum": -175,
    "sum": -210.7642857143,
    "average": -23.4182539683,
    "hasDiagonalMatrix": false
  }
}
```

This is the smallest coherent final result: it preserves the previously
returned QR matrices and makes the downstream result explicit.  It is an
additive response change; reviewers should confirm that adding `statistics`
to the existing response is acceptable for any strict existing client.

## Files to Create

- `go-api/statistics/client.go` — a small `statistics.Client` that owns the
  validated Node base URL and a reusable `*http.Client`.  It will define the
  Node request/result JSON types, send `POST /api/v1/statistics`, and expose a
  context-aware `Calculate` method for `Q` and `R`.
- `go-api/statistics/client_test.go` — standard-library `httptest` unit tests
  for the real Go HTTP client, using local in-process fake Node handlers.

## Files to Modify

- `go-api/main.go` — read `NODE_API_URL` at startup, fail fast with a concise
  configuration error when it is absent or invalid, create the single timeout
  bounded `http.Client` and `statistics.Client`, and pass it into the Fiber
  app.  `PORT` behavior remains unchanged.
- `go-api/httpapi/app.go` — make app construction accept the small downstream
  client dependency and register the same `POST /qr` route.  It must not
  construct HTTP clients or read environment variables.
- `go-api/httpapi/qr_handler.go` — after the existing QR operation succeeds,
  call the injected statistics client, map its categorized errors, and return
  the additive final response.  Existing parsing and QR-error behavior remain
  intact.
- `go-api/httpapi/qr_handler_test.go` — adapt existing app construction and
  retain the current QR and invalid-input tests; add focused end-to-end
  handler tests against an `httptest` fake Node service.

No `node-api/` files are modified.  In particular, `statistics.ts` remains
the sole statistics/diagonal implementation and the existing Express endpoint
remains the contract consumed by Go.

## Execution Order

1. Add the `statistics` package with explicit request/response structs,
   result validation, and a constructor that validates a non-empty absolute
   HTTP(S) `NODE_API_URL`.  Join the fixed endpoint path
   `/api/v1/statistics` to the configured base URL without embedding a host,
   port, localhost value, or production URL in code.
2. Configure exactly one `http.Client` during Go process startup with a
   `5 * time.Second` timeout and give it to `statistics.Client`.  Read
   `NODE_API_URL` through `os.Getenv` only in `main`; do not add a configuration
   framework or an `.env` file.
3. Change `httpapi.New` to receive the client dependency.  Define only the
   one consumer-oriented interface needed by the handler: a method equivalent
   to `Calculate(context.Context, qr.Matrix, qr.Matrix) (statistics.Result,
   error)`.  `statistics.Client` satisfies it implicitly and a test fake can
   satisfy it without a live network.  Do not add QR interfaces, factories,
   repositories, services, or a dependency-injection container.
4. Update the QR handler to bind the original request and invoke
   `qr.Factorize` exactly as it does now.  Only once both matrices exist, call
   the statistics dependency and serialize `{q, r, statistics}`.  The QR core
   neither imports nor knows about HTTP or Node.
5. Add the client unit tests and Fiber/mock-Node integration tests described
   below, preserving all existing QR-core and Go HTTP tests.
6. From `go-api/`, run the exact repository validation commands in
   **Validation Criteria**.  Do not run Node validation because Node is not
   changed.

## Architecture Decisions

### HTTP client design

`statistics.Client` is a compact stateful struct because base URL and the
reused `*http.Client` naturally belong together.  Its `Calculate(ctx, q, r)`
method will:

1. JSON-encode `{ "q": Q, "r": R }` using `encoding/json`.
2. Create a `net/http` request with `http.NewRequestWithContext`, method
   `POST`, `Content-Type: application/json`, and the configured endpoint.
3. Execute it with the injected shared client, always close the response body,
   and read any non-success body only through a small bounded reader for
   diagnostic context.
4. Treat every Node non-2xx status as a downstream failure; do not attempt to
   expose Node's body to the public client.  In particular, an observed Node
   `400` means Go generated a payload its declared downstream contract
   rejected, so it is a Go-to-Node contract/upstream failure rather than a
   reason to misleadingly reclassify the original client matrix as bad.
5. Decode exactly one JSON object into an internal response shape.  Require
   all five fields, require finite numeric aggregate values, require the
   boolean field, and reject trailing JSON values.  Missing fields, wrong
   types, non-finite values, and malformed/trailing JSON are invalid downstream
   responses.
6. Return errors wrapped with useful internal context (operation/status), but
   retain only simple sentinel/category information needed by the handler:
   timeout versus another downstream failure.  No error hierarchy is needed.

The `5s` timeout applies to the complete outbound request and gives an
explicit, appropriate bound for a local, small numerical operation while
remaining configurable in source later if operational evidence requires it.
It is deliberately a named Go constant for this feature, not another runtime
setting.  The sole environment variable is therefore:

| Variable | Required | Meaning |
| --- | --- | --- |
| `NODE_API_URL` | yes | Base HTTP(S) URL of the Node service; local, Compose, and Cloud Run deployments change only this value. |

No authentication is implemented.  Future Cloud Run IAM authentication can be
added at the outbound transport/request-authentication boundary without
changing QR or statistics logic.

### Context propagation

The client method accepts `context.Context`, and the Fiber handler passes
`c.RequestCtx()` to `http.NewRequestWithContext` while the call is executing.
This is the Fiber v3/fasthttp request context and allows cancellation on server
shutdown.  Fiber's default `c.Context()` is a background context, and
fasthttp's request context does not signal an individual client disconnect;
therefore this framework does not provide full disconnect cancellation without
additional plumbing.  The explicit five-second outbound timeout is the
reliable per-request bound.  This is the practical idiomatic propagation
available in the current framework, requires no custom context abstraction,
and keeps the client independently usable from non-Fiber callers.

### Public success and error contract

`POST /qr` continues to accept the same request and to return the existing QR
validation responses:

- QR input parse/semantic errors (`ErrEmptyMatrix`, `ErrEmptyRow`,
  `ErrNonRectangular`, `ErrWideMatrix`, and `ErrNonFinite`) remain `400` with
  the existing `invalid_request` response shape and message.
- A QR numerical or unexpected Go error remains `500` with the existing safe
  `internal_error` response.
- Successful QR plus successful Node statistics returns `200` and the
  top-level `q`, `r`, and `statistics` shape shown in the Summary.

Downstream errors use the existing small Go error JSON shape and avoid Node
URLs, response bodies, matrices, stack traces, and configuration values:

| Condition | Public status | Public body | Rationale |
| --- | --- | --- | --- |
| Connection/DNS/TLS/refusal, Node `4xx`/`5xx`, malformed or schema-invalid Node JSON | `502 Bad Gateway` | `{"error":"downstream_error","message":"unable to obtain statistics"}` | Go is the public orchestrator and did not obtain a valid upstream result; Node `4xx` is not incorrectly converted to `500` or blamed on the original client. |
| Client outbound timeout (including a context deadline) | `504 Gateway Timeout` | `{"error":"downstream_timeout","message":"statistics service timed out"}` | The upstream did not respond within Go's explicit gateway timeout. |
| Unexpected Go handler/client construction error not categorized as downstream | `500 Internal Server Error` | existing `internal_error` body | Internal failure, kept distinct from an upstream failure. |

The error body details above are a review decision: they extend the existing
two-field error convention with two stable downstream error labels, without a
large taxonomy.  Operational logs are not introduced in this feature; in
particular, matrices and `NODE_API_URL` are not logged.

### Testing strategy

Tests use Go's `testing`, `net/http/httptest`, and Fiber's in-process
`app.Test`; they do not require a Node process, a port configured for either
service, Docker, or cloud resources.

- `statistics/client_test.go` will verify a successful fake Node response,
  including POST path, JSON `{q,r}`, content type, and decoded result; a
  non-2xx response (include Node `400`); invalid/trailing JSON and a response
  missing/wrong required fields; and a delayed handler or failing round trip
  that becomes the timeout/network-error category.  Tests will use a short
  test-only client timeout and avoid timing-sensitive sleeps where a context
  deadline can deterministically trigger the case.
- `httpapi/qr_handler_test.go` will keep the mathematical QR assertions and
  invalid-client-input table.  It will add the representative flow
  `POST /qr` -> actual QR core -> fake Node `httptest.Server` -> `200` final
  response, asserting `q`/`r` properties plus the exact statistics field
  forwarded from Node.  It will also assert handler mappings for unavailable
  downstream (`502`), timeout (`504`), and malformed downstream JSON (`502`).
  A small fake satisfying the one handler interface is appropriate for mapping
  tests; the real `statistics.Client` tests own HTTP serialization/decoding.
- Existing `go-api/qr` unit tests and previous `POST /qr` validation tests
  continue to run unchanged in purpose.  This feature does not require a real
  cross-process Go+Node test; it may be added later as an optional local smoke
  test only if it provides value beyond this contract-tested boundary.

### Out of scope

Do not add Dockerfiles, Docker Compose, Terraform, GCP/Cloud Run resources,
Artifact Registry, service accounts, IAM, JWT, CI/CD, a frontend, a logging
framework, a configuration framework, or new Go dependencies.

## Validation Criteria

- A valid `POST /qr` factorizes through the existing QR core, calls the Node
  endpoint with its exact established `{q,r}` contract, and returns `200` with
  QR matrices plus all five Node statistics fields.
- The Go statistics client reuses the startup-created standard `http.Client`,
  has the explicit five-second timeout, sends JSON with the expected path and
  content type, closes every response body, rejects non-2xx responses, and
  rejects invalid, incomplete, non-finite, or trailing downstream JSON.
- Existing `400` QR-input behavior and existing `500` QR-failure behavior are
  preserved.  Downstream availability/malformed-response failures map to
  `502`; a timeout maps to `504`; Node `4xx` does not become a blind `500`.
- The tests are deterministic and isolated from a real Node service while
  exercising client serialization and the full Go handler/QR/mock-Node flow.
- From `go-api/`, run:

  ```bash
  find . -type f -name '*.go' -exec gofmt -w {} +
  test -z "$(find . -type f -name '*.go' -exec gofmt -l {} +)"
  go vet ./...
  go test ./...
  go build ./...
  ```

## External Dependencies

No new dependencies.  The Go implementation uses the existing Fiber module
and the standard library: `context`, `encoding/json`, `errors`, `fmt`,
`io`, `math`, `net/http`, `net/url`, `os`, `strings`, and `time` as needed.
Node's existing Express/Zod statistics contract is reused unchanged.

## Blocked Tasks

There are no technical blockers.  Implementation is blocked intentionally by
the repository workflow until a human explicitly approves this design.

Review before approval:

1. Confirm the additive public response shape `{q, r, statistics}` is the
   desired compatibility choice for existing `/qr` consumers.
2. Confirm `NODE_API_URL` is required at Go startup and that fail-fast startup
   is preferred over accepting traffic that can never complete integration.
3. Confirm the proposed five-second gateway timeout and the `502`/`504`
   downstream mapping.
4. Confirm the decision to treat a Node `400` as `502`, since Go created the
   QR matrices; this surfaces a downstream contract fault rather than calling
   the original client input invalid.
5. Confirm the Fiber v3 context limitation is acceptable for this scope: the
   outbound call is bound by timeout and server-shutdown context cancellation,
   but not individual client-disconnect cancellation.

Implementation must not begin, and no Developer agent may be invoked, until
explicit human approval of this `design.md`.

## Execution Results

Human approval was received before implementation. The approved scope was
completed without modifying `node-api/` or adding dependencies.

### Completed Files

- Created `go-api/statistics/client.go` and `go-api/statistics/client_test.go`.
- Modified `go-api/main.go`, `go-api/httpapi/app.go`,
  `go-api/httpapi/qr_handler.go`, and `go-api/httpapi/qr_handler_test.go`.
- The statistics client validates an absolute HTTP(S) `NODE_API_URL`, uses the
  shared startup-created `net/http.Client` with the approved five-second
  timeout, calls `POST /api/v1/statistics`, closes response bodies, validates
  the complete downstream JSON response, and categorizes timeout versus other
  downstream failures.
- The Fiber handler retains the QR core call, then delegates statistics to the
  injected client and returns the approved additive `{q, r, statistics}` body.
- Tests cover the Node request contract, successful client and handler flows,
  response-body closure, Node non-2xx, malformed/trailing/incomplete/wrong
  downstream JSON, deadline timeout, network failure, preserved QR validation,
  and `502`/`504` handler mappings.

### Validation

From `go-api/`, all approved validation commands passed:

```text
find . -type f -name '*.go' -exec gofmt -w {} +
test -z "$(find . -type f -name '*.go' -exec gofmt -l {} +)"
go vet ./...
go test ./...
go build ./...
```

Node validation was intentionally not run because no Node file changed.

### Blockers and Out-of-Scope Suggestions

There are no implementation blockers. The pre-approval workflow blocker is
resolved. Docker, Compose, Terraform, GCP/Cloud Run, IAM, JWT, CI/CD,
authentication, and Node source changes remain out of scope and were not
implemented.
