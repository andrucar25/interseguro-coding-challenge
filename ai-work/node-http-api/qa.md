# QA Report - node-http-api

## Scope reviewed

Reviewed the approved `node-http-api` design, implementation, HTTP contract, validation boundary, existing statistics unit tests, and the Node validation commands. No source, dependency, configuration, Go, infrastructure, or deployment files were modified by QA.

## Result

The implemented HTTP feature conforms to the approved API design and has no functional defect found in its endpoint, validation, core separation, or numerical behavior. One medium-severity tooling regression/blocker prevents the documented validation matrix from being entirely green after a build.

## Contract and implementation review

- `POST /api/v1/statistics` is mounted by the exported Express `app`, while `index.ts` only starts `listen`; Supertest imports the app without an explicit server bootstrap.
- Valid `{ q, r }` input returns `200` JSON with the unchanged core shape: `maximum`, `minimum`, `sum`, `average`, and `hasDiagonalMatrix`.
- The successful HTTP test verifies the documented `Q`/`R` example and exact expected values: `4`, `0`, `11`, `1.375`, and `true`.
- Zod is the network boundary. The strict schema accepts only non-empty rectangular matrices with non-empty rows and finite numeric scalars. It rejects an absent/scalar body, missing `q`/`r`, unknown top-level fields, malformed nested arrays, empty matrices/rows, ragged rows, strings, and non-finite JSON numeric input.
- Malformed JSON is mapped by final error middleware to `400` with `{ "error": "invalid_request" }`. Schema failures use the same response. The fallback returns `500` with `{ "error": "internal_error" }`; neither response exposes validation details, stack traces, paths, or environment data.
- The route passes `[q, r]` to `calculateStatistics` exactly once. It contains no aggregate calculation, rounding, or diagonal calculation. The existing core remains independent from Express and is still the sole owner of the `1e-12` relative diagonal tolerance.
- Existing unit tests remain separate from HTTP tests and cover aggregates across both matrices, negatives, decimals, rectangular matrices, invalid core input, diagonal detection, and tolerance boundary behavior. HTTP tests focus on transport and contract cases rather than repeating the mathematical suite.
- `helmet` and `express.json()` are used as planned. No CORS middleware, dotenv loading, health route, Go changes, or new dependencies were introduced.

## Automated validation

Commands were run from `node-api/`.

| Command | Result | Evidence |
| --- | --- | --- |
| `npm run typecheck` | Passed | `tsc --noEmit` completed with exit code 0. |
| `npm run check` | Failed after a generated `dist/` is present | `biome check .` inspected generated `dist/**/*.js` and reported 12 formatting/import-organization violations. Source TypeScript is not named in these failures. |
| `npm test` | Passed | Vitest reported 4 files / 62 tests passed. The 31 source tests are duplicated by their compiled copies in `dist/`. Supertest required the permitted local ephemeral listener. |
| `npm run build` | Passed | `rm -rf dist && tsc` completed with exit code 0. |

## Findings

### Medium — generated `dist/` makes `npm run check` fail and duplicates tests

**Evidence:** `npm run build` emits JavaScript and compiled test files under `node-api/dist/`. Because `npm run check` is `biome check .`, Biome checks that compiler output and fails on its formatting/import organization (12 errors in the generated files). Vitest also discovers `dist/app.test.js` and `dist/statistics/statistics.test.js`, resulting in 4 test files / 62 tests rather than the 2 source test files / 31 tests.

**Impact:** the required command set is not reproducibly all green after `npm run build`; test totals are misleading and each source test executes twice when `dist/` exists.

**Scope disposition:** this is a pre-existing/tooling configuration concern that the approved design explicitly deferred: it did not authorize changes to Biome, Vitest, TypeScript output, or package scripts. QA did not fix it. A follow-up tooling task should exclude generated `dist/` from Biome and Vitest discovery (or otherwise avoid emitting test files to runtime output), then demonstrate `typecheck`, `check`, `test`, and `build` in a stable sequence.

No other functional, security, API-contract, validation, numerical, or source-separation issue was found.

## Relevant coverage

- HTTP success: status, JSON content type, exact stable success object, and aggregate values across Q and R.
- HTTP invalid input: malformed JSON, absent body, absent `q`, absent `r`, empty matrix/row, non-array input, malformed nesting, ragged matrix, non-numeric scalar, unexpected property, and non-finite numeric JSON input.
- HTTP error contract: `400` JSON and exact `{ "error": "invalid_request" }` response across expected invalid-input cases.
- Core unit behavior: all aggregate fields, actual scalar count for average, rectangular input, diagonal/non-diagonal cases, diagonal tolerance threshold, input immutability, and defensive invalid-core-input errors.

The approved design does not require artificial dependency injection to force the fallback `500` path. Its safe final error response was inspected, but that path is not directly HTTP-tested.

## Prioritized follow-up tests

1. In a future tooling task, assert that the configured test command executes source tests exactly once after a build.
2. If the application later gains a natural failure seam, add one focused HTTP test that verifies an unexpected handler/core exception produces `500` and exactly `{ "error": "internal_error" }` without exposing internals.
3. When Go-to-Node communication is implemented, add a narrowly scoped cross-service contract test; it is explicitly out of scope here.

## Regression checklist

- [x] No Go, Docker, Terraform, GCP, JWT, CI/CD, or Go-to-Node integration was added.
- [x] App composition is independent from server startup.
- [x] Zod validates all accepted network payloads before core execution.
- [x] Expected input failures are `400`; unexpected failures are safely mapped to `500`.
- [x] Successful contract preserves core field names and JSON number precision without arbitrary rounding.
- [x] Diagonal tolerance remains exclusively in the statistics core.
- [x] HTTP tests use Vitest and Supertest without explicit port startup.
- [x] TypeScript strict typecheck and build pass.
- [ ] Biome check remains green after build (blocked by generated `dist/` inspection).
- [ ] Test command executes each test once after build (blocked by `dist/` test discovery).
