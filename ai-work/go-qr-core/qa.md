# QA Report — go-qr-core

## Scope reviewed

Reviewed the approved pure-Go QR core in `go-api/qr`. This review covers mathematical correctness, input and numerical-error behavior, unit-test quality, package boundaries, and Go validation. It does not cover HTTP, Fiber, Node integration, Docker, or deployment because those are deliberately outside this feature.

## Automated validation

Commands were run from `go-api/`:

| Command | Result |
| --- | --- |
| `gofmt -d qr/qr.go qr/qr_test.go` | Passed; no formatting diff. |
| `go vet ./...` | Passed; no findings. |
| `go test ./...` | Passed: `ok github.com/andrucar25/interseguro-coding-challenge/go-api/qr` (cached). |
| `go build ./...` | Passed. |

## Findings

No QA problems found.

The implementation matches the approved design:

- `qr.Factorize` is stateless pure mathematical code with explicit `Matrix` inputs and results; it imports only `errors` and `math`.
- It has no Fiber, HTTP, Node, configuration, dependency-injection, interface, or external-dependency coupling.
- It validates empty, empty-row, non-rectangular, wide, and non-finite inputs through documented sentinel errors and returns nil factors on failure.
- It copies caller input before mutation and returns economy factors for `m >= n`: `Q` is `m x n`, `R` is `n x n`.
- Householder reflectors use stable norm accumulation, a cancellation-avoiding sign choice, exact clearing below the diagonal, and non-finite intermediate/output guards.
- Tests are deterministic, standard-library unit tests. They verify dimensions, `Q^T Q ≈ I`, `Q R ≈ A`, upper-triangular `R`, input preservation, rank-deficient/all-zero behavior, negative/decimal values, and all approved invalid-input cases. Their centralized relative/absolute tolerance is `1e-10`.

## Regression checklist

- Valid square, tall, and single-column matrices retain economy dimensions and reconstruct the input within tolerance.
- Rank-deficient and all-zero valid matrices succeed without a rank-threshold error.
- No caller-owned input matrix is mutated.
- Invalid shape and non-finite inputs return their matching sentinel error with nil factors.
- Factor matrices remain finite, orthonormal (within tolerance), and upper triangular where required.

## Future test consideration (out of scope)

If the QR core later accepts substantially larger or ill-conditioned production matrices, add a near-linearly-dependent and extreme-but-finite magnitude property test to characterize numerical behavior at that scale. No change is required for the approved small-core feature.
