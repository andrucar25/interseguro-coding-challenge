---
name: go-fiber-development
description: Implement or review Go backend work using Fiber in the go-api project. Use for Go source code, Fiber routes and handlers, HTTP clients, matrix or QR-related Go logic, Go tests, package structure, error handling, or Go-specific refactoring.
---

# Go + Fiber Development

Implement Go code idiomatically rather than translating Java, C#, NestJS, or TypeScript patterns directly into Go.

The user's explicit requirements and repository instructions take precedence over this skill.

## Required Context

Before modifying `go-api/`, read:

- `AGENTS.md`
- `docs/project.md`
- `docs/conventions.md`
- `docs/commands.md`

Also read:

- `docs/architecture.md` for request flow, integration, or architecture work
- `docs/security.md` for service authentication, credentials, external communication, or sensitive configuration

Inspect the existing Go source before proposing new packages or abstractions.

## Core Philosophy

Prefer:

- simple packages
- structs
- methods
- functions
- composition
- small interfaces
- explicit errors
- standard library functionality

Avoid trying to reproduce traditional class-based OOP.

Go does not need:

- abstract base classes
- inheritance hierarchies
- service inheritance
- repository interfaces for every implementation
- factories for simple constructors
- dependency injection containers
- interfaces for every struct

Introduce an abstraction only when it solves an actual problem.

## Object-Oriented Design in Go

Use structs when data and dependencies naturally belong together.

Example use cases include:

- an HTTP handler with dependencies
- a statistics HTTP client containing a base URL and `http.Client`
- configuration values that travel together

Attach methods to structs when behavior naturally belongs to that type.

Use standalone functions for stateless operations, especially mathematical transformations.

Prefer composition over inheritance.

## Interfaces

Keep interfaces small and consumer-oriented.

Do not create:

`QRServiceInterface`

solely because a `QRService` struct exists.

An interface is appropriate when there is a real abstraction boundary, such as allowing an HTTP handler to substitute a downstream statistics client during tests.

Prefer interfaces containing the minimum methods required by the consumer.

Go types satisfy interfaces implicitly; do not emulate `implements`.

## Package Design

Use short domain-oriented package names.

Prefer names such as:

- `qr`
- `statistics`
- `httpapi` when a package distinction is genuinely necessary

Avoid catch-all packages such as:

- `utils`
- `common`
- `helpers`
- `misc`

Do not create deeply nested package structures for this challenge.

Packages should reflect real responsibilities rather than architectural ceremony.

## Fiber Boundary

Fiber belongs to the HTTP transport layer.

Handlers should primarily:

1. obtain request data
2. validate or decode it
3. invoke application/mathematical functionality
4. map errors
5. produce HTTP responses

Do not implement QR decomposition directly inside a Fiber handler.

Do not spread Fiber-specific types through mathematical packages.

Mathematical code should be testable without starting Fiber.

## Data Types

Use domain types when they make code clearer.

For matrices, prefer an explicit type when it improves readability and method/function signatures.

Do not create large object models for simple matrix operations.

Be deliberate about `float64`, since QR decomposition naturally involves non-integer results even when the input contains integers.

## Error Handling

Handle errors explicitly.

Do not:

- ignore returned errors
- use panic for normal request failures
- obscure errors behind generic messages internally

Wrap errors with useful context when appropriate.

Expose safe HTTP errors at the boundary while preserving enough internal context for debugging.

## Context and HTTP

When making downstream HTTP calls:

- use `context.Context`
- propagate request cancellation where practical
- configure an `http.Client` with a timeout
- close response bodies
- handle non-success HTTP responses explicitly
- avoid creating a new HTTP client for every request

Prefer the standard library `net/http` client unless another dependency has a demonstrated benefit.

Fiber being the server framework does not require using a third-party HTTP client.

## Mathematical Code

Keep matrix and QR logic independent from HTTP.

Prefer functions that make inputs and outputs explicit.

Handle numerical edge cases intentionally.

Never compare floating-point results for mathematical equality when tolerance is required.

When checking properties affected by floating-point computation, define and justify an epsilon/tolerance.

Do not optimize prematurely, but avoid unnecessary repeated traversals or allocations when an equally clear solution is available.

## Comments and Documentation

Prefer self-explanatory names and small functions.

Comments should explain:

- non-obvious numerical decisions
- unusual Go behavior
- architectural reasons
- important trade-offs

Do not comment obvious code.

Exported Go identifiers should have concise Go doc comments when appropriate.

## Dependencies

Prefer the standard library.

Before adding a dependency, determine whether:

1. the standard library already solves the problem clearly
2. the dependency substantially improves correctness or maintainability
3. the dependency can be defended during the interview

Do not add libraries merely to imitate patterns used in Node.

## Validation

After modifying Go code, run the applicable repository commands.

First apply `gofmt`, then verify that it reports no unformatted files, and
then run the remaining applicable Go validations. Normally include:

```bash
find . -type f -name '*.go' -exec gofmt -w {} +
test -z "$(find . -type f -name '*.go' -exec gofmt -l {} +)"
go vet ./...
go test ./...
go build ./...
```

Use the exact commands documented in `docs/commands.md` when they differ.

Do not hide failing validation.

## Final Review

Before completing Go work, verify:

- the code is idiomatic Go
- Fiber is restricted to the transport boundary
- mathematical logic is independently testable
- errors are explicit
- interfaces exist only where useful
- package structure remains small
- standard library was preferred where reasonable
- no Java/TypeScript architecture was copied unnecessarily
- formatting, vet, tests, and build succeed
