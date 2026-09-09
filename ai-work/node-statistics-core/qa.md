# QA Report - node-statistics-core

## Result

**Pass with one pre-existing tooling issue outside the approved feature scope.** The statistics core meets the approved design on inspection and its source-level validation passes. No HTTP endpoint, Express server wiring, Go communication, or other excluded integration code was introduced.

## Scope and implementation review

- `node-api/src/statistics/statistics.ts` is a small, stateless TypeScript module with no framework, transport, or external-service imports.
- It accepts `unknown`, validates a non-empty collection of non-empty rectangular matrices, and rejects non-arrays, invalid rows, non-numbers, `NaN`, and infinities with `TypeError`.
- The nested scalar traversal updates maximum, minimum, sum, count, matrix scale, and off-diagonal magnitude together. `average` is derived from `sum / count`, so it uses the actual number of scalar values.
- Diagonal classification matches the approved rule: square matrices only, with `abs(offDiagonal) <= 1e-12 * max(1, largestAbsoluteValue)`. A `1 x 1` matrix is consequently diagonal, while rectangular matrices are not candidates.
- Inputs are only read; the feature adds no dependencies or configuration changes.

## Automated validation

Executed from `node-api/`:

| Command | Result | Notes |
| --- | --- | --- |
| `npm run typecheck` | Pass | `tsc --noEmit` completed successfully with strict TypeScript. |
| `npm run check` | Blocked by existing tooling issue | Biome checks generated `dist/` files because `biome.json` has `vcs.useIgnoreFile: false`; it reports 5 format/import-assist findings in compiled JavaScript only. |
| `npm test` | Pass | Vitest: 2 files, 36 tests passed. The statistics suite contributes 18 deterministic pure-logic cases. |
| `npm run build` | Pass | TypeScript compilation to `dist/` completed successfully. |
| `npx biome check src` | Pass | Confirms the three source files, including the new module and test, satisfy Biome. |

`npm run check` was observed both with an existing `dist/` directory and again after `npm run build`; it fails for the same generated artifacts. `dist/` is gitignored, but Biome is explicitly configured not to honor ignore files. This is not a defect in the feature source and no configuration change was made because it is outside the approved design. A separate approved tooling task should either exclude `dist/` from Biome or revise the validation order/command.

## Test coverage review

The direct Vitest tests cover positive, negative, decimal, and Q/R-shaped rectangular inputs; maximum, minimum, sum, and average; diagonal, non-diagonal, mixed collections, `1 x 1`, and relative-tolerance boundary behavior; immutability; and empty/malformed/non-finite input failures. They do not depend on Express, a listener, Go, or infrastructure, as required.

No coverage-instrumentation command is configured, so no percentage coverage was generated.

## Findings and risks

1. **P2 — `npm run check` is not reproducibly green after build.** Generated `dist/` files are included by Biome despite `.gitignore`; the command therefore fails in the normal validation sequence. This predates/is external to the feature behavior, but it prevents the requested full project check from passing. Do not fix it as part of this feature without separate approval.
2. **No functional or numerical defects found** in the approved statistics-core scope.

## Regression checklist

- [x] Strict TypeScript and no `any` in the feature module.
- [x] Pure calculation is independent of Express, HTTP, Go, and runtime configuration.
- [x] Aggregates include every scalar across every matrix and average uses scalar count.
- [x] Validation rejects the designed invalid and empty cases without returning partial results.
- [x] Diagonal tolerance is documented, scale-aware, and tested at/above its boundary.
- [x] Inputs are not mutated.
- [x] No dependencies, application configuration, or excluded-scope integration code changed.
