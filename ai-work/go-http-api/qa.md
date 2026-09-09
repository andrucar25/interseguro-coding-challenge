# QA Report — go-http-api

## Scope reviewed

Reviewed the approved `design.md` against `go-api/main.go`, `go-api/httpapi/`,
and the existing pure `go-api/qr/` package. No source, module manifest,
configuration, documentation, or dependency changes were made during QA.

## Automated validation

Run from `go-api/`:

| Check | Result |
| --- | --- |
| `test -z "$(find . -type f -name '*.go' -exec gofmt -l {} +)"` | Pass |
| `go vet ./...` | Pass |
| `go test -count=1 ./...` | Pass (`httpapi`, `qr`; root has no test files) |
| `go build ./...` | Pass |

An additional, non-required `go test -count=1 -cover ./...` failed before
executing tests with `cannot find package` / `internal/coverage/cfile: package
testmain: cannot find package`. The required test command passes, so this is a
Go toolchain coverage-instrumentation issue in the current environment, not a
feature test failure. Coverage percentages are therefore unavailable.

## Contract and behavior review

- `POST /qr` is the only registered route. It returns `200` with `q` and `r`
  for a valid square matrix.
- The HTTP test verifies output dimensions, `Q^TQ ≈ I`, `Q × R ≈ A`, and that
  `R` is upper triangular using a centralized tolerance; it does not depend on
  QR sign convention.
- Malformed JSON, missing matrix, empty matrix, empty row, non-rectangular
  rows, a wide matrix, non-numeric values, and an overflowing numeric literal
  are table-tested. Each returns `400` with the defined
  `{ "error": "invalid_request", "message": ... }` shape.
- The handler delegates the matrix constraints and factorization to
  `qr.Factorize`; it contains no QR mathematical logic. Fiber types remain in
  `httpapi`, while `qr` remains framework independent.
- The handler maps all specified QR input sentinel errors to `400` and all
  other errors, including `ErrNumericalFailure`, to the approved safe `500`
  response. Existing QR tests continue to cover valid dimensions and invalid
  core inputs, and passed.
- `main.go` uses `PORT` with a `3000` fallback and introduces no Node client,
  statistics, infrastructure, authentication, or other out-of-scope work.

## Findings

No functional, API-contract, regression, or scope violations found in the
implemented files. The files correspond to the approved implementation list:
`go-api/main.go`, `go-api/httpapi/app.go`, `go-api/httpapi/qr_handler.go`, and
`go-api/httpapi/qr_handler_test.go`. No Go module changes are present.

The `500` mapping cannot be reached through ordinary finite JSON input and has
no direct test seam, which matches the approved decision to avoid a QR mock or
interface. Its branch is straightforward and statically reviewed; this is not
a release blocker.

## Regression checklist

- [x] Valid QR request returns the expected JSON field names and dimensions.
- [x] Invalid JSON and invalid matrix inputs return safe `400` errors.
- [x] QR core property and invalid-input tests pass independently of Fiber.
- [x] Fiber contains transport concerns only.
- [x] Required formatting, vet, test, and build checks pass.
- [x] No Go-to-Node communication or other explicitly deferred scope was added.

