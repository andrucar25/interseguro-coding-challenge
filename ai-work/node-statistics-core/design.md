# Development Plan - node-statistics-core

## Summary

Add one small, framework-independent TypeScript statistics module under `node-api/src/statistics/`. It will validate a collection of finite, non-empty, rectangular numeric matrices and calculate their maximum, minimum, sum, average, and whether at least one matrix is diagonal. It will not import Express, Zod, configuration, or any future Go/HTTP integration code.

The module is deliberately a pure-function boundary: it neither mutates its inputs nor keeps state. The current `node-api/` only has the setup `src/index.ts`; that file remains untouched because starting a server is explicitly out of scope.

## Files to Create

- `node-api/src/statistics/statistics.ts`
  - Define the exported read-only matrix aliases:

    ```ts
    export type Matrix = readonly (readonly number[])[];
    export type Matrices = readonly Matrix[];
    ```

  - Define and export the single result shape:

    ```ts
    export interface MatrixStatistics {
      maximum: number;
      minimum: number;
      sum: number;
      average: number;
      hasDiagonalMatrix: boolean;
    }
    ```

  - Export the pure entry point:

    ```ts
    export function calculateStatistics(input: unknown): MatrixStatistics;
    ```

    Its `unknown` input intentionally allows a future transport boundary to pass decoded JSON without `any`; it performs the core's own runtime validation before treating the input as `Matrices`.
  - Keep validation and traversal helpers private. Reject with a concise `TypeError` when the input is not a non-empty array of matrices, a matrix or row is empty, rows are ragged, a value is not a number, or a number is `NaN`, `Infinity`, or `-Infinity`. No custom error hierarchy is needed for this pure module; a future HTTP feature can map this expected error at its own boundary.
  - Traverse every scalar exactly once. During that traversal, accumulate `maximum`, `minimum`, `sum`, and the real scalar `count`; calculate `average` only as `sum / count` after the traversal. Per matrix, also retain its largest absolute element and largest absolute off-diagonal element, so diagonal detection does not require a second value traversal.
  - Treat only non-empty square matrices as diagonal candidates. A `1 x 1` matrix is diagonal; a rectangular matrix is not. A square matrix is diagonal when every off-diagonal value is numerically zero according to the tolerance decision below. Set `hasDiagonalMatrix` if any supplied matrix meets that predicate.

- `node-api/src/statistics/statistics.test.ts`
  - Add direct, deterministic Vitest unit tests for `calculateStatistics`. The tests import no Express application and require no listener, environment configuration, Go service, or infrastructure.

## Files to Modify

- None. Leave `node-api/src/index.ts`, `package.json`, `package-lock.json`, `tsconfig.json`, and `biome.json` unchanged. Vitest, TypeScript strict mode, and Biome are already present, and this feature requires no dependency or configuration change.

## Execution Order

1. Create `statistics.ts` with the read-only public types, result type, documented tolerance constant, private structural/numeric validation, and the pure calculation function.
2. Implement the single scalar traversal and per-matrix diagonal metadata, then produce the result only after a validated non-zero value count exists.
3. Create focused Vitest tests in `statistics.test.ts` using manually verifiable expected values and explicit floating-point assertions where necessary.
4. Validate from `node-api/`:

   ```bash
   npm run typecheck
   npm run check
   npm test
   npm run build
   ```

## Architecture Decisions

- **One stateless module:** Statistics are calculations with no dependencies or lifecycle. A function plus explicit types is clearer than a class, service, repository, entity, use case, factory, interface layer, or DI container. Keep both the validation needed to protect the core and the calculation in this capability module; do not create empty architectural folders.

- **Input and validation contract:** The function accepts `unknown` and rejects invalid runtime values rather than relying only on compile-time types. Valid input is a non-empty collection; each matrix has at least one row and column, is rectangular, and contains only finite JavaScript numbers. Rectangular matrices are valid because future QR output includes a generally rectangular `Q`; square shape is required only for the mathematical diagonal predicate. Empty collections, empty matrices, empty rows, ragged matrices, non-array structures, non-numbers, `NaN`, and infinities are invalid. Inputs are read only and never modified. No Zod schema is introduced: there is no HTTP/untrusted transport in scope, and explicit guards are sufficient for this small core.

- **Result names and aggregate semantics:** Return `maximum`, `minimum`, `sum`, `average`, and `hasDiagonalMatrix`. All scalar entries across all matrices participate equally in the four aggregates. `sum` is the requested total sum. `average` uses the total scalar count, not the number of rows or matrices, so matrices of different dimensions contribute correctly. The implementation initializes aggregate values from the first validated scalar (or equivalent sentinels after collection validation), avoiding incorrect assumptions that values are positive or that zero is present.

- **Single-pass aggregation:** In one nested traversal, update maximum, minimum, sum, count, the per-matrix maximum absolute magnitude, and the per-matrix maximum off-diagonal magnitude. This is `O(total scalar values)` time and `O(1)` additional working memory apart from loop variables. It preserves clarity and fulfills the one-pass requirement for maximum, minimum, sum, and count; average is derived once from sum/count.

- **Diagonal definition and floating-point tolerance:** QR results are `float64` values serialized as JavaScript `number`s, so checking off-diagonal values with exact `=== 0` would incorrectly reject numerical round-off. Define one documented module constant, `DIAGONAL_RELATIVE_TOLERANCE = 1e-12`. For each square matrix, compute `scale = max(1, largest absolute value in that matrix)` and regard it as diagonal when `largest absolute off-diagonal value <= DIAGONAL_RELATIVE_TOLERANCE * scale`. The floor of `1` makes the rule meaningful for zero/small matrices; scaling makes it appropriate when QR factors have large magnitudes. This is a pragmatic classification tolerance, not a claim that near-zero values are mathematically exact zero. Values exactly at the threshold are accepted; values clearly above it are not. A fixed exact-zero rule is rejected because it is brittle for floating-point QR output, while a fixed absolute epsilon alone is not scale-aware.

- **Deliberate exclusions:** Do not add an Express app, endpoint, route, controller, HTTP validation schema, response contract, Go client, server startup, environment variables, Docker, GCP, Terraform, JWT, OpenAPI, mocks, or cross-service tests. Those require separately approved future work. `rest-api-design` is intentionally not used because this plan creates no HTTP contract.

## Validation Criteria

- Direct Vitest tests cover a positive-valued collection and a negative-valued collection, asserting each aggregate independently: maximum, minimum, sum, and average.
- A decimal-value test verifies the aggregate results with `toBeCloseTo` (or an equivalent explicit numerical tolerance) so ordinary JavaScript floating-point representation does not make the test brittle.
- A combined `Q` and `R`-shaped collection (for example, a rectangular `Q` plus square `R`) proves all values in both matrices contribute to maximum, minimum, sum, and the average denominator; it also demonstrates rectangular matrices are accepted.
- A diagonal square matrix, including a `1 x 1` case where useful, yields `hasDiagonalMatrix: true`; a square matrix with a clearly non-zero off-diagonal value yields `false` when no other input matrix is diagonal.
- A mixed collection verifies the result is `true` when one matrix is diagonal even if another is not.
- Floating-point diagonal tests use a known scale and assert both sides of the rule: an off-diagonal magnitude at or below `1e-12 * max(1, scale)` is accepted, and one clearly above it is rejected. Include a large-scale matrix to prove the tolerance is relative rather than a hidden fixed absolute epsilon.
- Table-driven invalid-input tests cover an empty collection, an empty matrix, an empty row, ragged rows, a non-array input, a non-numeric scalar, `NaN`, and both positive and negative infinity. Each must throw and never return partial statistics.
- `npm run typecheck`, `npm run check`, `npm test`, and `npm run build` pass from `node-api/`. Tests remain independent of Express and external services.

## External Dependencies

None. Use the existing TypeScript runtime and Vitest development dependency only. Do not add Zod usage, numerical libraries, test libraries, HTTP libraries, or configuration packages for this pure core feature.

## Blocked Tasks

None. The plan is ready for human review and explicit approval before implementation.
