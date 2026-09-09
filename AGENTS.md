# AGENTS.md

This repository contains a small technical challenge that will become a monorepo. Read [.agents/README.md](.agents/README.md) before changing repository files.

## Repository

Current state: only `node-api/` may exist. The intended final layout also includes `go-api/` and `infrastructure/`; do not create a future directory unless an approved task requires it.

The intended system accepts a client request in Go/Fiber, performs QR factorization, sends the resulting matrices to Node/Express for statistics, and returns the result through the Go API.

## Workflow

The specialized workflow agents are in `.agents/agents/`: Orchestrator, Explorer, Planner, Developer, Documenter, and QA. The required sequence is Explorer → Planner → explicit human approval of `/ai-work/[feature]/design.md` → Developer → Documenter → QA. Only planning-to-implementation needs mandatory human approval.

## Non-negotiable rules

- Read relevant files and routed context before editing.
- Keep changes focused; do not modify unrelated projects or add unnecessary dependencies.
- Use commands for the project being modified.
- Keep configuration in environment variables; never commit secrets or `.env` files.
- Do not implement work outside an explicitly approved design when acting as Developer.
- Update repository guidance when a change materially alters commands, architecture, conventions, or security practices.

## Context routing

- All work: `.agents/context/project.md` and `.agents/context/conventions.md`.
- Cross-service flow or deployment: additionally read `architecture.md`.
- Commands, validation, or tooling: additionally read `commands.md`.
- External input, credentials, or service access: additionally read `security.md`.
