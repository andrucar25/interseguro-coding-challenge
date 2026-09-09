# Development Plan - node-http-api

## Summary

Expose the existing Node statistics core through one synchronous Express endpoint:

`POST /api/v1/statistics`

The endpoint receives the QR matrices separately as `q` and `r`, validates untrusted HTTP input with Zod, passes them to the existing `calculateStatistics` function as its existing collection input (`[q, r]`), and returns that function's result unchanged. This preserves the core's domain names and, in particular, its existing relative-tolerance rule for diagonal matrices. No QR logic, Go client, Go source, deployment, authentication, or infrastructure is in scope.

The existing core names are clear and coherent, so the external success contract deliberately retains them rather than introducing an output mapping:

```json
{
  "maximum": 4,
  "minimum": 0,
  "sum": 11,
  "average": 1.375,
  "hasDiagonalMatrix": true
}
```

The only transport adaptation is from named QR inputs to the core's generic matrix collection. It is intentionally kept in the HTTP handler, rather than changing the core API or duplicating calculation logic.

## Files to Create

- `node-api/src/app.ts` — creates and exports the configured Express application without calling `listen`, allowing Supertest to exercise it in-process. It installs JSON parsing, mounts the statistics router, and installs the small final error handler.
- `node-api/src/statistics/statistics-request.schema.ts` — Zod schema and inferred request type for `{ q, r }`; it is the HTTP boundary's single definition of the accepted payload.
- `node-api/src/statistics/statistics.router.ts` — small Express router/handler for `POST /api/v1/statistics` (mounted at the application level). It parses with the schema, calls `calculateStatistics([q, r])`, and serializes the core result.
- `node-api/src/app.test.ts` — Vitest + Supertest HTTP contract tests that import the app directly, so no real TCP port is opened.

## Files to Modify

- `node-api/src/index.ts` — replace the current setup message with the runtime entry point: import the app and call `app.listen` using `PORT` from the environment, with a documented local default. This module must not contain routes or statistics logic.

The following remain unchanged:

- `node-api/src/statistics/statistics.ts` — remains the sole implementation of statistics and diagonal tolerance.
- `node-api/src/statistics/statistics.test.ts` — remains the unit-level mathematical suite; HTTP tests supplement it and do not replay all math cases.
- `node-api/package.json`, lockfile, TypeScript/Biome configuration, and dependencies — Express, Zod, Vitest, and Supertest are already installed, so no dependency or tooling change is required.

## Execution Order

1. Add the Zod request schema. Require a strict object with `q` and `r`; each is a non-empty matrix of non-empty rows of finite numbers, with equal row lengths inside that matrix.
2. Add the statistics router/handler. On a successful `safeParse`, call the existing `calculateStatistics([q, r])` once and return its object with `res.status(200).json(...)`. Do not calculate, round, or test diagonal elements in this handler.
3. Add the app composition module. Use `express.json()` before the router; distinguish malformed JSON parser errors from unexpected errors in the final error middleware.
4. Convert `index.ts` into the server bootstrap that imports the app and calls `listen`. Keep `PORT` configuration at this bootstrap boundary; it is not needed by tests.
5. Add focused Supertest cases, preserving the existing pure-core tests.
6. Run the Node validation commands in the Validation Criteria section from `node-api/`.

## Architecture Decisions

### Minimal application shape

The request path is:

```text
HTTP request -> Express app -> statistics router/handler -> Zod schema
             -> existing calculateStatistics([q, r]) -> JSON response
```

`app.ts` owns Express composition and exports an app; `index.ts` owns process startup and `listen`. This is the smallest useful separation for port-free HTTP tests and avoids server lifecycle code in tests. A small router/handler and a colocated schema make the transport boundary explicit without adding controllers, services, repositories, entities, use cases, dependency injection, factories, or generic response abstractions.

### Endpoint and contracts

**Request** — `POST /api/v1/statistics`, with `Content-Type: application/json`:

```json
{
  "q": [
    [1, 0],
    [0, 1]
  ],
  "r": [
    [2, 3],
    [0, 4]
  ]
}
```

`q` and `r` are named after the QR outputs, which makes the later Go caller and an interview explanation direct. The statistics are computed across every scalar in both matrices, and `hasDiagonalMatrix` is true when either matrix meets the core's diagonal definition.

**Successful response** — `200 OK`, JSON, with the flat existing `MatrixStatistics` shape:

```json
{
  "maximum": 4,
  "minimum": 0,
  "sum": 11,
  "average": 1.375,
  "hasDiagonalMatrix": true
}
```

The value order in JSON is not contractual; the five field names and their JSON number/boolean types are. There is no arbitrary rounding: JavaScript number values from the core are serialized as JSON numbers.

**Error response** — every planned error response is JSON with this small, stable shape:

```json
{ "error": "invalid_request" }
```

for invalid client input, and:

```json
{ "error": "internal_error" }
```

for unexpected server failures. The stable machine-readable value is sufficient for the later Go caller and deliberately omits validation paths, stacks, server paths, and implementation details.

### Status codes

- `200 OK`: both matrices pass validation and the core returns statistics.
- `400 Bad Request`: malformed JSON, an absent body, or a payload that fails the request schema.
- `500 Internal Server Error`: an unexpected error after parsing/validation. Log the underlying error only through the service's normal server-side mechanism if one is later adopted; never include it in this response.

No `404` is introduced because this operation does not address a resource, and no expanded error taxonomy is needed.

### Validation boundary

Zod must reject with `400` all of the following before calling the core:

- missing or non-object body, including an absent body and a JSON scalar/array;
- syntactically invalid JSON (handled by the JSON parser's error path);
- missing `q` or `r`, or unexpected top-level fields (strict object);
- `q` or `r` that is not an array, an empty matrix, a row that is not an array, or an empty row;
- scalar elements that are not numbers or are non-finite (`NaN`, `Infinity`, `-Infinity`);
- ragged/non-rectangular rows within either matrix; and
- malformed nested structures at any level.

Both matrices are validated independently. The core accepts non-empty rectangular matrices of any dimensions and calculates statistics independently of matrix multiplication, so there is no core-relevant cross-dimension constraint to impose between `q` and `r`. The HTTP layer must not invent QR-specific compatibility rules that belong to the future Go QR producer.

The core keeps its current defensive input checks for direct callers; Zod is the authoritative network boundary. The planned implementation must not copy the aggregate loop or diagonal check into Express.

### Floating point and diagonal matrices

Do not round incoming QR `float64` values or successful output values. `calculateStatistics` already owns diagonal detection using its `1e-12` relative tolerance scaled by the matrix's largest absolute value. The HTTP handler only conveys the matrices and returns the result, so it cannot accidentally diverge from that numerical rule.

### Middleware and operational scope

- Use `express.json()` because the endpoint accepts JSON.
- Use the already installed `helmet` as a small, justified HTTP security baseline for an exposed Express process; it does not create an architectural layer or affect the core.
- Do **not** use `cors`: Go-to-Node service calls are not browser requests, and future private Cloud Run communication does not need CORS. It can be considered only if a browser consumer is explicitly added.
- Do **not** load `dotenv` for this feature. The only runtime setting is conventional `PORT`, which `index.ts` reads from `process.env` with a local default; loading `.env` is unnecessary and no `.env` file is added.
- Do **not** add `/health` now. Docker, Cloud Run, and observability are deferred; a health route without an active deployment need is out of the feature's focused scope.

## Validation Criteria

Keep existing unit tests as the proof of aggregate mathematics and add HTTP-focused tests only for the boundary:

- success: valid `q` and `r` return `200`, `application/json`, the exact five-field success shape, and manually verifiable statistics (including that either diagonal matrix makes `hasDiagonalMatrix` true);
- validation: malformed JSON and an absent body; missing `q` and missing `r`; empty matrices and empty rows; non-array/nested malformed structures; ragged rows; strings or other non-numeric values; and non-finite values when constructible at the request boundary;
- contract: invalid payloads return `400`, JSON, and exactly `{ error: "invalid_request" }`; an unexpected error is mapped by the final handler to `500` with `{ error: "internal_error" }` and no leaked implementation detail. The suite need not create artificial dependency injection solely to force the latter path.

Do not duplicate all positive/negative/decimal/tolerance mathematics through HTTP. Supertest imports `app` directly and never starts a port. Existing statistics unit tests remain independent of Express.

From `node-api/`, run the actual documented commands:

```bash
npm run typecheck
npm run check
npm test
npm run build
```

## External Dependencies

No new dependencies. The necessary packages are already present: `express`, `zod`, `helmet`, `vitest`, and `supertest`. `cors` and `dotenv` remain installed but are intentionally unused by this feature. TypeScript strict mode, Biome, `tsx`, and `tsc` remain as configured.

## Blocked Tasks

None for planning. Implementation must wait for explicit human approval of this design. The following are explicitly deferred: Go-to-Node HTTP communication and client configuration, all `go-api/` changes, Docker, Terraform/GCP/Cloud Run, JWT, and CI/CD.

## Execution Results

Implemented after explicit human approval.

### Completed files

- Created `node-api/src/app.ts` with the exportable Express app, `helmet`, JSON parsing, the statistics route, and safe final error responses.
- Created `node-api/src/statistics/statistics-request.schema.ts` with the strict Zod request schema for non-empty rectangular finite-number matrices.
- Created `node-api/src/statistics/statistics.router.ts`; it invokes `calculateStatistics([q, r])` without reimplementing aggregate or diagonal logic.
- Created `node-api/src/app.test.ts` with focused Vitest/Supertest HTTP contract tests.
- Modified `node-api/src/index.ts` into the `PORT`-based server bootstrap.

`node-api/src/statistics/statistics.ts` and its unit tests were not changed.

### Validation results

From `node-api/`:

- `npm run typecheck` — passed.
- `npm run check` — passed against the source tree after clearing a pre-existing generated `dist/` directory. The current Biome configuration checks generated JavaScript in `dist/`, while TypeScript emits it with formatting Biome flags; no configuration change was made because it is outside this approved feature.
- `npm test` — passed: 2 test files and 31 tests. The sandbox blocks Supertest's ephemeral local listener (`EPERM`), so the successful run was performed with the required local-listening permission.
- `npm run build` — passed.

### Blocked files and out-of-scope suggestions

No approved implementation files were blocked. No changes were made to Go integration, infrastructure, authentication, CI/CD, CORS, dotenv loading, or a health endpoint. A future tooling task may choose to exclude generated `dist/` output from Biome checks so that `npm run check` is independent of whether a build has run.
