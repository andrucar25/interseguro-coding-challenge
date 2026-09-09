---
name: backend-testing
description: Design, implement, or review backend tests for the Go/Fiber and Node/Express services. Use for unit tests, HTTP integration tests, matrix edge cases, QR numerical validation, statistics validation, Go-to-Node contract testing, regression checks, or deciding the appropriate testing level.
---

# Backend Testing

Create focused tests that provide confidence in correctness without maximizing test count.

Use the smallest testing level that can prove the behavior.

The user's explicit requirements and repository instructions take precedence over this skill.

## Required Context

Before testing, inspect:

- `AGENTS.md`
- `docs/project.md`
- `docs/architecture.md` when integration is involved
- `docs/conventions.md`
- `docs/commands.md`
- the implementation being tested

Understand the expected behavior before writing tests.

Do not derive expected behavior solely from the current implementation.

## Testing Strategy

Prefer three useful levels:

### Pure Logic Tests

Use for mathematical and statistical behavior.

These should be:

- fast
- deterministic
- independent from HTTP
- independent from cloud infrastructure

Examples include:

- QR decomposition properties
- maximum
- minimum
- average
- sum
- diagonal detection

### HTTP Tests

Use to verify:

- endpoint behavior
- validation
- status codes
- serialization
- error responses

For Node, use Supertest where appropriate.

For Go/Fiber, use Fiber's supported request testing mechanisms and Go's standard `testing` package.

### Cross-Service Tests

Use selectively for Go → Node communication.

Verify the contract between services without turning every test into a full end-to-end deployment.

Mock or substitute the downstream boundary when testing Go behavior in isolation.

Add a real local integration test only when it provides meaningful confidence beyond isolated tests.

## Matrix Test Cases

Select cases that reveal defects rather than merely increasing coverage.

Consider:

- simple square matrices
- rectangular matrices
- single-row or single-column matrices when supported
- zeros
- negative numbers
- decimal values
- repeated values
- empty input
- malformed/non-rectangular input
- matrices invalid for the requested operation
- numerically sensitive cases where relevant

Use examples with manually verifiable expected results whenever possible.

## Statistics Tests

Verify at least the meaningful behavior for:

- maximum
- minimum
- total sum
- average
- diagonal detection

Include negative and decimal values where they reveal errors.

Ensure average uses the number of actual scalar values, not rows or matrices.

When statistics are defined across multiple matrices, explicitly verify that all intended matrices are included.

## QR Tests

Do not validate QR decomposition only by comparing Q and R against one exact representation.

QR decompositions can differ in signs while remaining mathematically valid.

Prefer validating mathematical properties such as:

- matrix dimensions are correct
- `QᵀQ` is approximately the identity matrix
- `R` has the expected triangular property
- `Q × R` reconstructs the original matrix within an appropriate tolerance

Use numerical tolerances for floating-point comparisons.

Document or centralize the tolerance when practical.

Include invalid or degenerate input behavior according to the approved design.

## Diagonal Matrices and Floating Point

When diagonal detection is applied to QR output, do not assume mathematically zero values will always be represented as exact `0`.

Use the tolerance defined by the implementation/design.

Tests should include values:

- clearly within tolerance
- clearly outside tolerance

to make the intended behavior explicit.

## Error Cases

Test important error paths, including where applicable:

- invalid JSON
- missing matrix
- invalid matrix shape
- invalid values
- unsupported matrix dimensions
- downstream Node service failure
- timeout
- malformed downstream response

Do not exhaustively test framework internals.

## Go Tests

Prefer Go's standard `testing` package.

Use table-driven tests when several cases exercise the same behavior cleanly.

Do not add an external testing library unless it materially improves the suite.

Keep tests idiomatic and readable.

## Node Tests

Use Vitest.

Use Supertest for Express HTTP integration tests.

Test pure statistics functions directly without routing them through HTTP.

Avoid snapshots for simple JSON or numerical behavior when explicit assertions communicate intent better.

## Efficiency

Test meaningful boundaries instead of duplicating equivalent cases.

A smaller suite of strong tests is preferable to a large shallow suite.

Tests should:

- fail for a real behavioral regression
- explain the expected contract
- remain deterministic
- avoid external cloud dependencies during normal local/CI execution

## Validation

After changing tests, run the relevant project validation commands from `docs/commands.md`.

For Node this normally includes:

```bash
npm test
npm run typecheck
npm run check
npm run build
```

For Go this normally includes:

```bash
go test ./...
go vet ./...
go build ./...
```

Run formatting as required by repository conventions.

## Final Review

Before completing testing work, verify:

- tests assert requirements rather than implementation details
- mathematical tests use appropriate floating-point tolerance
- QR is verified through mathematical properties
- HTTP tests cover meaningful validation and errors
- pure logic tests do not require the web framework
- cross-service tests are used only where they add value
- the suite remains understandable and maintainable
