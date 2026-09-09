---
name: express-typescript-development
description: Implement or review Node.js backend work using Express and TypeScript in node-api. Use for Express routes, validation, statistics logic, TypeScript types, HTTP behavior, error handling, Biome, Vitest, Supertest, or Node-specific refactoring.
---

# Express + TypeScript Development

Build the Node service as a small, strongly typed Express application appropriate for this technical challenge.

The user's explicit requirements and repository instructions take precedence over this skill.

## Required Context

Before modifying `node-api/`, read:

- `AGENTS.md`
- `docs/project.md`
- `docs/conventions.md`
- `docs/commands.md`

Also read:

- `docs/architecture.md` for cross-service or architecture changes
- `docs/security.md` when authentication, credentials, service access, or external communication is involved

Inspect existing Node code and `package.json` before introducing files, patterns, or dependencies.

## Core Philosophy

Prefer:

- strict TypeScript
- small modules
- explicit types
- pure functions for stateless calculations
- simple dependency composition
- Zod at untrusted input boundaries
- Express only at the transport boundary

Avoid:

- unnecessary Clean Architecture layers
- DDD ceremony
- generic repositories
- dependency injection containers
- abstract base services
- class hierarchies
- generic controller frameworks
- patterns copied from NestJS when Express does not require them

Every abstraction should have an explainable reason to exist.

## Object-Oriented Programming

Classes are allowed but not mandatory.

Use a class when it provides actual value through:

- encapsulated state
- dependencies
- lifecycle
- behavior tied naturally to an instance

Prefer pure functions when the operation is stateless.

Statistics such as:

- maximum
- minimum
- average
- sum
- diagonal detection

are naturally suited to pure functions or a small stateless module unless the implementation demonstrates a concrete reason for a class.

Do not use classes merely to make the code appear more architectural.

## Suggested Module Shape

Organize code around capabilities rather than generic technical layers when practical.

For example, a statistics capability may contain its:

- schema
- route
- service/functions
- tests

Do not create empty architecture folders preemptively.

Do not create:

- repository
- entity
- domain
- use-case
- factory
- adapter

layers unless an actual requirement justifies them.

## Express Boundary

Routes and handlers should coordinate:

1. request extraction
2. validation
3. invoking statistics logic
4. mapping errors
5. returning HTTP responses

Do not place substantial statistical logic directly inside a route callback.

Keep Express `Request` and `Response` types out of pure business/mathematical functions.

## Validation

Use Zod for untrusted JSON input.

Validate at the API boundary.

For matrix inputs, explicitly consider:

- required fields
- arrays
- numeric values
- empty matrices
- rectangular shape where required
- appropriate matrix dimensions
- malformed nested arrays

Do not trust the Go caller merely because it belongs to the same repository.

Treat network input as untrusted.

Infer TypeScript types from Zod schemas when this reduces duplication cleanly.

Avoid maintaining two independent definitions of the same request structure without reason.

## TypeScript

Keep strict mode enabled.

Avoid `any`.

If `any` is genuinely unavoidable:

- constrain it as narrowly as possible
- explain why

Prefer:

- `unknown` at untrusted boundaries
- explicit return types where they improve public contracts
- discriminated or explicit result shapes when needed

Avoid excessive type-level complexity for a small service.

## Statistics

Prefer a clear algorithm that can compute related aggregate statistics efficiently.

Where practical, maximum, minimum, sum, and count can be accumulated in a single traversal.

Calculate average from total sum and count.

Do not sacrifice readability for micro-optimizations.

Diagonal detection should have a clearly defined numerical rule, especially if matrices contain floating-point results.

Use a justified tolerance rather than assuming exact floating-point zero when input comes from QR decomposition.

## Error Handling

Use consistent HTTP error responses.

Do not expose:

- stack traces
- environment values
- internal details

Do not catch errors only to silently ignore them.

Keep expected input errors distinct from unexpected server failures.

Do not add a large custom error hierarchy unless it solves a demonstrated problem.

## Biome

Biome is the repository's formatter and linter.

Do not introduce:

- ESLint
- Prettier

Run the configured Biome scripts instead.

Follow existing `biome.json` configuration rather than inventing local exceptions to make code pass.

## Runtime and Build

Use:

- `tsx` for development
- `tsc --noEmit` for type checking
- `tsc` for production build
- Node to execute compiled JavaScript

Do not introduce another transpiler or bundler unless an actual need emerges.

## Testing

Use:

- Vitest for tests
- Supertest for HTTP integration tests

Keep mathematical/statistics tests independent from Express where possible.

Do not test implementation details unnecessarily.

## Comments

Prefer descriptive function and variable names.

Comments should explain non-obvious:

- business interpretation
- numerical tolerance
- architectural trade-offs
- external-service behavior

Do not comment obvious syntax.

## Dependencies

Before adding a new dependency:

1. inspect existing dependencies
2. determine whether Node or the current stack already solves the problem
3. ensure the dependency adds enough value to justify itself

Do not add packages solely for convenience when a few clear lines of code are sufficient.

## Validation

After relevant Node changes run the commands defined in `docs/commands.md`.

Normally this should cover:

```bash
npm run typecheck
npm run check
npm test
npm run build
```

Do not suppress failures.

## Final Review

Before completing Node work, verify:

- TypeScript remains strict
- no unnecessary `any`
- validation occurs at the boundary
- Express code is separated from statistics logic
- stateless operations have not been forced into unnecessary classes
- naming is coherent
- no unnecessary dependencies were added
- Biome passes
- type checking passes
- tests pass
- build succeeds
