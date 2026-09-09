# Node Statistics Core

## Scope

This feature provides framework-independent statistics for a non-empty collection of numeric matrices. It does not expose an Express endpoint, start a server, communicate with Go, or define an HTTP contract.

## Internal API

The module is available from `node-api/src/statistics/statistics.ts`:

```ts
type Matrix = readonly (readonly number[])[];
type Matrices = readonly Matrix[];

interface MatrixStatistics {
  maximum: number;
  minimum: number;
  sum: number;
  average: number;
  hasDiagonalMatrix: boolean;
}

function calculateStatistics(input: unknown): MatrixStatistics;
```

Example:

```ts
calculateStatistics([
  [[1, 0], [0, 2]],
  [[-3, 4]],
]);
// {
//   maximum: 4,
//   minimum: -3,
//   sum: 4,
//   average: 0.8,
//   hasDiagonalMatrix: true,
// }
```

All scalar values in all matrices contribute equally to `maximum`, `minimum`, `sum`, and `average`. The average is `sum / count`, where `count` is the total number of scalar entries, not the number of rows or matrices.

## Input rules

`calculateStatistics` validates its `unknown` input and throws `TypeError` for:

- a non-array or empty collection;
- an empty or non-array matrix;
- empty rows or non-array rows;
- ragged rows;
- non-numeric, `NaN`, or infinite values.

Valid matrices are non-empty and rectangular. Rectangular matrices are accepted for QR-shaped data; only square matrices can satisfy the diagonal predicate. Inputs are read-only and are not mutated.

## Diagonal semantics

`hasDiagonalMatrix` is `true` when at least one supplied matrix is square and every off-diagonal value is within the relative tolerance `1e-12`.

For each matrix, the threshold is:

```text
1e-12 × max(1, largest absolute matrix value)
```

This tolerance accounts for floating-point round-off in QR results, while scaling with large matrix values. Values exactly at the threshold are accepted. A `1 × 1` matrix is diagonal; rectangular matrices are not diagonal candidates.

## Implementation and tests

The aggregates and diagonal metadata are calculated during one traversal of the scalar values. The operation is `O(total scalar values)` with constant additional working memory, and the input is never modified.

Vitest tests are in `node-api/src/statistics/statistics.test.ts`. They cover positive, negative, decimal, Q/R-shaped, diagonal and non-diagonal matrices, mixed collections, tolerance boundaries, immutability, and invalid or empty inputs. Tests are independent of Express and external services.

Run from `node-api/`:

```bash
npm test
npm run typecheck
npm run check
npm run build
```

