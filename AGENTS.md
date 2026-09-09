# AGENTS.md

This repository contains a small technical challenge that will become a monorepo.

## Repository

Current state: only `node-api/` may exist. The intended final layout also includes `go-api/` and `infrastructure/`; do not create a future directory unless an approved task requires it.

The intended system accepts a client request in Go/Fiber, performs QR factorization, sends the resulting matrices to Node/Express for statistics, and returns the result through the Go API.

## Workflow

The main Codex thread coordinates feature work and waits for each dependent agent result before continuing. The custom agents are in `.codex/agents/`.

For non-trivial features, use this sequential workflow:

```text
scope_explorer
    ↓
planner
    ↓
STOP
    ↓
human reviews ai-work/[feature]/design.md
    ↓
explicit human approval
    ↓
developer
    ↓
documenter
    ↓
qa
```

Never spawn `developer` for a feature until the user explicitly approves its `ai-work/[feature]/design.md`. The main Codex thread must stop after `planner`, summarize or show the design to the user, and must not interpret silence as approval.

- Use `scope_explorer` when requirements need clarification or structured scope analysis. It may be skipped when requirements are completely clear.
- Use `planner` before non-trivial implementation.
- Use `developer` only after explicit human approval of `design.md`.
- Use `documenter` after meaningful features when documentation is useful.
- Use `qa` after implementation when verification is needed.

Do not run dependent stages in parallel. Parallelize only independent work that materially benefits from it.

For trivial, isolated, low-risk changes—such as a typo correction, small rename, simple documentation change, or obvious configuration adjustment—the main Codex thread may make the change directly and run relevant validation. The full workflow or explicit planning is required for new features and non-trivial changes, including statistics, QR factorization, Go-to-Node HTTP communication, Docker, Terraform/GCP, authentication, architecture changes, and service integration.

Feature artifacts belong in `ai-work/[feature]/`: `explorer-output.md` when exploration applies, `design.md`, `feature-doc.md` when documentation applies, and `qa.md` when QA applies. Do not create an empty `ai-work/` directory until a feature needs it.

## Skill Routing

Use repository Skills only when they are relevant to the assigned task. Their practical guidance never expands approved scope. Apply this precedence order:

```text
Explicit user instructions
    ↓
repository instructions in AGENTS.md
    ↓
approved feature scope/design
    ↓
applicable Skills
    ↓
agent default behavior
```

Skills cannot authorize files outside an approved design, replace approved decisions, contradict user instructions, or introduce dependencies or architecture merely because they are mentioned as possibilities.

- `rest-api-design`: use for REST endpoints, request/response or JSON contracts, status codes, error contracts, HTTP-boundary validation, Go → Node contracts, OpenAPI, or other consumer-visible API changes. Do not use it for internal logic that does not change an HTTP contract.
- `go-fiber-development`: use when designing or modifying `go-api/`, including Go/Fiber code, handlers, routes, types, packages, HTTP clients, errors, QR/matrix logic, and Go-specific testing conventions. Do not use it for `node-api/` work.
- `express-typescript-development`: use when designing or modifying `node-api/`, including Express, TypeScript, Zod, statistics, routes, validation, types, errors, Biome, Vitest, Supertest, and Node-specific refactors. Do not use it for `go-api/` work.
- `backend-testing`: use for test design, implementation, or review; numerical QR or statistics validation; HTTP or Go → Node integration tests; regressions; and test edge cases. Do not load it automatically for a trivial change that does not affect tests.

## Minimal Skill Loading

Use the smallest applicable skill set for the task; never load all Skills by default.

- Pure Go QR implementation: `go-fiber-development`; add `backend-testing` only when tests are designed or implemented.
- Go HTTP endpoint: `go-fiber-development` and `rest-api-design`; add `backend-testing` when applicable.
- Node statistics pure logic: `express-typescript-development`; add `backend-testing` when applicable.
- Node statistics endpoint: `express-typescript-development`, `rest-api-design`, and `backend-testing`.
- Go → Node communication: `go-fiber-development`, `rest-api-design`, and `backend-testing`; add `express-typescript-development` only when Node code or its contract is also modified.

## Non-negotiable rules

- Read relevant files and routed context before editing.
- Keep changes focused; do not modify unrelated projects or add unnecessary dependencies.
- Use commands for the project being modified.
- Keep configuration in environment variables; never commit secrets or `.env` files.
- Developer work must follow an explicitly approved design.
- Update repository guidance when a change materially alters commands, architecture, conventions, or security practices.

## Context routing

Read the smallest context set necessary for the task.

- Normal work: `docs/project.md` and `docs/conventions.md`.
- Cross-service flow, design, or deployment: additionally read `docs/architecture.md`.
- Code, validation, or tooling changes: additionally read `docs/commands.md`.
- Credentials, authentication, authorization, external services, IAM, or sensitive infrastructure: additionally read `docs/security.md`.
