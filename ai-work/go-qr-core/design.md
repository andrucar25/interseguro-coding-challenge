# Development Plan - go-qr-core

## Summary

Add a small, pure Go QR-factorization package under `go-api/`. It will accept a finite, non-empty, rectangular `float64` matrix with at least as many rows as columns and return its economy (thin) QR factorization. The package will have no imports of Fiber, HTTP, Node, configuration, or service-layer code.

The current `go-api/` contains only `go.mod` and `go.sum`; it has no Go source to extend or preserve. No `explorer-output.md` exists for this feature.

## Files to Create

- `go-api/qr/qr.go`
  - Define `package qr` and the matrix type `type Matrix [][]float64`.
  - Expose the stateless API:

    ```go
    func Factorize(input Matrix) (q Matrix, r Matrix, err error)
    ```

  - Expose documented sentinel errors so callers can distinguish invalid input from numerical failure with `errors.Is`:

    ```go
    var (
        ErrEmptyMatrix
        ErrEmptyRow
        ErrNonRectangular
        ErrWideMatrix
        ErrNonFinite
        ErrNumericalFailure
    )
    ```

  - Validate the complete input before calculation: at least one row, at least one column, equal row lengths, `rows >= columns`, and only finite values. `Factorize` must not mutate the caller's matrix and must return `nil, nil, err` on failure.
  - Implement economy QR with Householder reflections using only the standard library (`math` and `errors` as needed). Copy the input into the working `R`; construct only the requested `m x n` columns of `Q` instead of materializing a full `m x m` matrix. Apply stored reflectors in reverse order to form `Q`.
  - Use stable norm accumulation (`math.Hypot`) and the conventional reflector sign choice that avoids cancellation. Explicitly clear mathematically eliminated entries below `R`'s diagonal. Detect non-finite intermediate/output values and return `ErrNumericalFailure` rather than returning invalid matrices.
  - Treat an exactly zero Householder tail as a no-op. Do not impose an arbitrary rank threshold: rank-deficient and all-zero valid matrices still have a valid QR factorization, with zero (or numerically near-zero) diagonal entries in `R`.

- `go-api/qr/qr_test.go`
  - Add pure, deterministic standard-library `testing` tests for the public behavior. Keep test-only matrix multiplication, transpose, comparison, and shape helpers in this file; do not create a production utility package.

## Files to Modify

- None. In particular, leave `go-api/go.mod` and `go-api/go.sum` unchanged: the feature needs no dependency beyond the Go standard library, and the existing Fiber dependency is outside this feature's responsibility.

## Execution Order

1. Create `go-api/qr` with the `Matrix` type, documented error values, input validation, and copying/allocation helpers private to the package.
2. Implement `Factorize` with Householder reflections, including its finite-result guard and economy output dimensions.
3. Add property-oriented unit tests in `go-api/qr/qr_test.go`, plus table-driven invalid-input tests.
4. Format and validate only from `go-api/`:

   ```bash
   gofmt -w qr/qr.go qr/qr_test.go
   go vet ./...
   go test ./...
   go build ./...
   ```

## Architecture Decisions

- **Package and API:** Use one domain package, `qr`, rather than a service, repository, interface, or generic helper layer. QR factorization has no state or dependencies, so a pure function with explicit input/output/error values is the idiomatic Go boundary. `Matrix` gives signatures a clear domain meaning without creating an object model.

- **Algorithm:** Use Householder reflections rather than classical Gram-Schmidt. Householder QR is a compact standard-library implementation and is materially more numerically stable for nearly dependent columns. Modified Gram-Schmidt would be simpler to narrate but has weaker orthogonality in numerically sensitive cases; it offers no offsetting benefit here.

- **Supported rectangular shape and dimensions:** Support tall and square matrices only: for input `A` of shape `m x n`, where `m >= n >= 1`, return economy factors `Q` of shape `m x n` and `R` of shape `n x n`. `Q` has orthonormal columns and `R` is upper triangular, so `A ≈ Q R`. This avoids allocating an unused full `m x m` orthogonal matrix. Reject wide matrices (`m < n`) with `ErrWideMatrix`; a wide-matrix/full-QR contract would require different dimensions (`Q m x m`, `R m x n`) and is intentionally outside this small core API.

- **Degenerate but valid input:** Rank deficiency, repeated/dependent columns, and the zero matrix are supported, not reported as input errors. A QR decomposition remains defined; tests will assert reconstruction and orthogonality rather than require positive diagonal values in `R`. An exact all-zero reflector tail is skipped so it does not divide by zero.

- **Numerical behavior:** Use `float64` throughout. Validate `NaN` and `±Inf` at the boundary. Avoid a magic rank epsilon in the implementation because it would silently redefine finite input as zero. Use stable norm construction and cancellation-avoiding reflector signs; if calculation itself produces a non-finite value, fail explicitly with `ErrNumericalFailure`. Floating-point equality is never used to judge matrix properties.

- **Ownership:** Copy the input before transformations and allocate independent `Q`/`R` outputs. This makes the exported operation functionally safe for callers while retaining the simple slice-of-slices representation.

- **Deliberate exclusions:** Do not add Fiber routes/handlers/server setup, HTTP clients, Node communication, configuration, Docker, GCP, JWT, interfaces, mocks, or additional dependencies. No REST contract is being designed.

## Validation Criteria

- Valid square, tall rectangular, single-column, and decimal/negative-valued matrices return the exact economy dimensions and no error.
- For every valid result, test the mathematical properties rather than a single exact representation:
  - `QᵀQ ≈ I_n`;
  - `Q × R ≈ A`;
  - every `R[i][j]` below the diagonal (`i > j`) is approximately zero;
  - each row of `Q` has `n` elements and each row of `R` has `n` elements, with `len(Q) == m` and `len(R) == n`.
- Centralize a test tolerance such as `1e-10` and compare scalars with a combined absolute/relative rule, `abs(got-want) <= tolerance * max(1, abs(got), abs(want))`. This is strict enough for the small deterministic matrices while accommodating normal `float64` rounding and QR sign differences.
- Include a rank-deficient matrix (dependent columns) and an all-zero matrix; both must succeed and satisfy the same factorization/orthogonality/dimension properties. Include a nontrivial tall example with negative and decimal values to exercise numerical arithmetic.
- Table-test errors for nil/empty input, an empty first row, an empty later row, ragged rows, a wide matrix, and matrices containing `NaN`, `+Inf`, and `-Inf`. Assert the relevant sentinel via `errors.Is` and assert no factor matrices are returned on error.
- Run `gofmt`, `go vet ./...`, `go test ./...`, and `go build ./...` from `go-api/`. The tests must not start Fiber, use HTTP, contact Node, or require external infrastructure.

## External Dependencies

None. The implementation and tests use the Go standard library only. Do not add numerical, web, test-framework, or mocking dependencies.

## Blocked Tasks

None. The feature is ready for explicit human approval before implementation.

## Execution Results

- Completed: `go-api/qr/qr.go` implements the approved pure Householder economy QR API, validation, finite-value guards, and explicit sentinel errors.
- Completed: `go-api/qr/qr_test.go` adds deterministic standard-library property tests, input-ownership coverage, and table-driven invalid-input coverage.
- Blocked files: none.
- Validation from `go-api/`: `gofmt -w qr/qr.go qr/qr_test.go`, `go vet ./...`, `go test ./...`, and `go build ./...` all passed. `go test ./...` reported `ok github.com/andrucar25/interseguro-coding-challenge/go-api/qr`.
- Out-of-scope suggestions: none.
