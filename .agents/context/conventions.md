# Conventions

## Global

- Prefer simple solutions and avoid premature abstractions.
- Read before editing and keep service boundaries explicit.
- Do not modify unrelated projects.
- Use environment variables for configuration; never hardcode environment-specific URLs or credentials.
- Add comments for non-obvious reasoning, not to restate code.

## Node API

- Use strict TypeScript, Express, and Zod for external input validation where appropriate.
- Use Biome only for linting and formatting; do not add ESLint or Prettier.
- Use Vitest and Supertest for HTTP integration tests.
- Use `tsx` in development and `tsc` for type checking and builds.
- Avoid `any` unless unavoidable and justified.

## Go API

- Use idiomatic Go, Fiber, explicit error handling, and the standard library when sufficient.
- Format Go code with `gofmt`.

Do not introduce Clean Architecture, DDD, repository patterns, dependency injection, or comparable structural complexity without a demonstrated need.
