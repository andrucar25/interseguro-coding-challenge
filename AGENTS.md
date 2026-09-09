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
