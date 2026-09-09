---
name: rest-api-design
description: Design or review REST API contracts for this repository. Use when defining or changing endpoints, request/response schemas, validation boundaries, HTTP status codes, error responses, or communication contracts between the Go/Fiber and Node/Express services.
---

# REST API Design

Design REST APIs that are simple, explicit, consistent, and appropriate for the size of this technical challenge.

The user's explicit requirements and the repository's `AGENTS.md` and `docs/` guidance take precedence over this skill.

## Before Designing

Read the smallest relevant context set.

Usually inspect:

- `AGENTS.md`
- `docs/project.md`
- `docs/architecture.md`
- `docs/conventions.md`
- the existing API code and contracts affected by the change

Also read `docs/security.md` when authentication, authorization, service access, credentials, or untrusted external communication is involved.

Do not design an API without first checking existing repository conventions.

## Principles

Prefer the smallest API contract that completely expresses the requirement.

Do not introduce:

- unnecessary resource abstractions
- generic base controllers
- generic response wrappers without a demonstrated need
- versioning complexity that is not required
- GraphQL, RPC, messaging, or event-driven patterns when REST/HTTP satisfies the requirement

Keep transport concerns separate from business or mathematical logic.

HTTP handlers/routes should coordinate:

1. request parsing
2. validation
3. application operation
4. response mapping

They should not contain substantial mathematical or business logic.

## Naming

Use names that describe the domain operation rather than implementation details.

Prefer:

- `matrix`
- `matrices`
- `statistics`
- `qr`
- `maximum`
- `minimum`
- `average`
- `sum`
- `diagonal`

Avoid vague names such as:

- `data`
- `payload`
- `process`
- `resultData`
- `utils`

when a more precise name is available.

Keep naming consistent between Go and Node when both services represent the same concept.

Do not force language-specific types or naming styles to be identical.

## Request Design

Every externally supplied request must have an explicit expected shape.

For matrix inputs, consider at least:

- missing matrix
- empty matrix
- empty rows
- non-rectangular matrices
- invalid numeric values
- dimensions unsupported by the requested operation

Do not silently repair invalid external input unless the specification explicitly requires it.

Validation belongs at the transport boundary.

## Response Design

Return only information needed by the caller.

Avoid exposing internal implementation details.

For successful operations, use predictable and stable JSON structures.

When Go consumes Node, treat the Node response as an external contract even though both services belong to the same repository.

## HTTP Status Codes

Use standard HTTP semantics.

Typical choices include:

- `200` for successful synchronous operations
- `400` for malformed or semantically invalid client input
- `404` only when an actual resource is absent
- `500` for unexpected internal failures
- `502` when the public Go service cannot obtain a valid response from its downstream Node service, when this accurately describes the failure

Do not invent custom status-code semantics.

Use the smallest meaningful error model.

## Error Responses

Error responses should be:

- machine-readable
- concise
- consistent
- safe to expose

Do not leak:

- stack traces
- credentials
- internal URLs unnecessarily
- implementation internals

Prefer an error shape that can remain stable across endpoints.

Do not add elaborate error-code taxonomies unless the project demonstrates a need.

## Cross-Service Contract

For Go → Node communication:

- use HTTP as required by the challenge
- define the request and response contract explicitly
- configure the Node base URL through environment variables
- set reasonable HTTP timeouts
- propagate failures intentionally rather than swallowing them
- validate or safely decode downstream responses
- keep networking code separate from mathematical logic

Do not tightly couple the QR algorithm to the HTTP client.

## OpenAPI

When API contracts become stable enough to document, prefer a version-controlled OpenAPI specification over adding runtime Swagger dependencies solely for documentation.

Do not introduce Swagger UI or code-generation tooling unless explicitly requested or clearly beneficial.

## Final Review

Before completing an API design, verify:

- the endpoint has one clear responsibility
- request and response types are explicit
- validation is defined
- relevant error cases are defined
- HTTP status codes follow normal semantics
- the contract does not leak implementation details
- cross-service naming is coherent
- no unnecessary abstraction was introduced
- the design can be explained succinctly in a technical interview
