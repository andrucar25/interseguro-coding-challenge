# Development Plan - go-http-api

## Summary

Add the first client-facing Go/Fiber surface for the existing pure `qr` package. The service will start a Fiber application and expose `POST /qr`. It will accept one JSON matrix, validate it at the HTTP boundary through the existing QR input contract, execute `qr.Factorize`, and return the economy factor matrices. The HTTP layer will contain parsing, validation/error mapping, and JSON serialization only; it will not duplicate QR math.

This feature intentionally ends after QR output is returned. It does not call the Node API, calculate statistics, add a downstream client, or introduce any cross-service configuration.

## Files to Create

- `go-api/main.go`
  - Provide the executable entry point.
  - Construct the Fiber app through `httpapi.New()` and listen on `":" + port`, using `PORT` when set and `3000` as the local default. Keep startup wiring limited to this file; do not add configuration packages.

- `go-api/httpapi/app.go`
  - Define the small `httpapi` transport package and export `New() *fiber.App`.
  - Register the sole route, `POST /qr`, and no version, health, statistics, or Node-related routes.

- `go-api/httpapi/qr_handler.go`
  - Define unexported request, success-response, and error-response structs plus the handler registered by `New`.
  - Decode the JSON request, invoke `qr.Factorize`, and map only safe errors to HTTP responses. Fiber types stay confined to this package.

- `go-api/httpapi/qr_handler_test.go`
  - Add standard-library Go/Fiber HTTP tests against the application returned by `httpapi.New()`. The tests will exercise the route as a client would; they will not start a network listener or mock QR because QR is already a stateless function.

## Files to Modify

- None. `go-api/go.mod` and `go-api/go.sum` already declare Fiber and require no dependency changes. `go-api/qr/` remains the unchanged pure mathematical implementation.

## Execution Order

1. Create `httpapi.New` and register `POST /qr` in the Fiber app.
2. Add the QR request/response types and handler. Parse JSON, call the existing `qr.Factorize`, and map its errors according to the contract below.
3. Add `main.go` to read `PORT`, construct the app, and call `Listen`.
4. Add focused HTTP tests, then format and validate from `go-api/`:

   ```bash
   find . -type f -name '*.go' -exec gofmt -w {} +
   test -z "$(find . -type f -name '*.go' -exec gofmt -l {} +)"
   go vet ./...
   go test ./...
   go build ./...
   ```

## Architecture Decisions

- **Route and method:** The public operation is `POST /qr`. QR factorization consumes a JSON matrix and performs computation, so a POST operation is clearer than encoding arbitrary matrix data in a URL. There is one route because only one client-facing operation is in scope.

- **Request contract:** Require an `application/json` body shaped as:

  ```json
  {
    "matrix": [[12, -51, 4], [6, 167, -68], [-4, 24, -41]]
  }
  ```

  `matrix` must decode as an array of numeric row arrays. It must contain at least one row and each row must contain at least one number; every row must have the same number of columns; every value must be finite; and the matrix must be tall or square (`rows >= columns`). These are precisely the constraints accepted by the existing economy QR API. No artificial maximum row/column count is added in this feature. Omitted, `null`, or empty `matrix` is invalid; JSON syntax errors, wrong JSON types, and numeric values that cannot decode to a finite `float64` are also invalid. Extra JSON fields are not part of the contract and may be ignored by normal struct decoding.

- **Success contract:** On valid input and successful factorization, return `200 OK` with:

  ```json
  {
    "q": [[-0.8571428571, 0.3942857143], [-0.4285714286, -0.9028571429], [0.2857142857, -0.1714285714]],
    "r": [[-14, -21], [0, -175]]
  }
  ```

  `q` and `r` are JSON arrays of `float64` values directly produced by `qr.Factorize`. For an `m x n` input, `m >= n >= 1`, `q` is `m x n` and `r` is `n x n`; callers must not depend on one exact QR sign convention, only on those dimensions and the factorization properties. The response contains no statistics or downstream metadata.

- **Error contract:** Return concise, stable, safe JSON errors:

  ```json
  {"error":"invalid_request","message":"matrix rows must have equal lengths"}
  ```

  Return `400 Bad Request` with `error: "invalid_request"` for malformed JSON, a missing/non-array matrix, and `qr.ErrEmptyMatrix`, `qr.ErrEmptyRow`, `qr.ErrNonRectangular`, `qr.ErrWideMatrix`, or `qr.ErrNonFinite`. For QR sentinel errors, expose their existing messages as `message`; malformed/type-decoding failures use the generic message `"request body must contain a valid matrix"` so parser internals are not exposed. Return `500 Internal Server Error` as `{"error":"internal_error","message":"unable to factorize matrix"}` for `qr.ErrNumericalFailure` or any unexpected error. Do not return a partial `q` or `r` on error.

- **Validation boundary:** Let Fiber decode JSON into the local request type, then call `qr.Factorize` as the single authoritative semantic validator and calculator. This preserves one source of truth for matrix shape/finite-number rules and ensures direct Go callers and HTTP callers behave consistently. The handler maps known QR errors with `errors.Is`; it does not replicate matrix loops or numerical logic.

- **Server structure:** Use one small transport package rather than controllers, service interfaces, a dependency container, or a generic response framework. `main` owns only process startup, `httpapi` owns Fiber, and `qr` remains framework-independent. Fiber is already a direct module dependency, so no module changes are needed.

- **Port configuration:** Read only `PORT`, a conventional deployment setting, with `3000` as a development fallback. Node base URLs, timeouts, credentials, and any Node HTTP client are explicitly deferred because no downstream call belongs to this approved scope.

- **Focused HTTP tests:** Use `httptest.NewRequest` and Fiber's supported in-process `app.Test` mechanism. Assert a valid tall or square matrix returns `200`, JSON content with expected `Q`/`R` shapes, and factorization properties (with a floating-point tolerance) rather than hard-coding one exact QR representation. Table-test meaningful invalid requests: malformed JSON, omitted/empty matrix, empty row, ragged rows, a wide matrix, and an invalid numeric/type payload. Assert `400` and the stable error JSON. A test seam/interface for QR is unnecessary: the existing calculation is fast and deterministic. The existing pure QR package tests continue to own comprehensive numerical-property coverage.

## Validation Criteria

- `POST /qr` accepts a finite, non-empty rectangular square or tall matrix and returns `200` with `q` of shape `m x n` and `r` of shape `n x n`.
- An HTTP test verifies `QᵀQ ≈ I`, `Q × R ≈ A`, and the upper-triangular property of `R` using a centralized tolerance, without relying on QR signs.
- Malformed JSON; omitted, empty, or empty-row matrices; ragged matrices; wide matrices; and non-numeric/non-finite values are rejected with `400` and the defined safe error format.
- A numerical or unexpected QR failure maps to the defined `500` response and never leaks internal details or a partial result.
- The server starts from `main.go`, honors `PORT` when present, and uses local port `3000` otherwise.
- From `go-api/`, `gofmt`, its no-unformatted-files check, `go vet ./...`, `go test ./...`, and `go build ./...` all pass.
- Tests are deterministic, use no live listener, Node service, cloud resource, environment secret, or new test framework.

## External Dependencies

None. Fiber is already declared in `go-api/go.mod`; the server, validation, and tests use Fiber plus the Go standard library. Do not add a validation library, numerical library, HTTP client, mock library, OpenAPI runtime, or Node dependency.

## Blocked Tasks

None. The plan deliberately excludes Node HTTP communication and statistics; those require a separate approved design after this feature is implemented.
