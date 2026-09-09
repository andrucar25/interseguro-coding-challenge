# Agent Context System

`AGENTS.md` is the canonical repository guidance. The workflow agents in `agents/` coordinate feature work and write their artifacts under `/ai-work/[feature]/`.

## Workflow

1. Explorer validates requirements and writes `explorer-output.md`.
2. Planner creates `design.md`.
3. A human explicitly approves `design.md`.
4. Developer implements only that approved design and records results in it.
5. Documenter writes `feature-doc.md` when useful.
6. QA writes `qa.md`.

The Orchestrator maintains `orchestrator-state.md` and supports resuming a feature. No approval is required between the other stages unless a blocker requires a decision.

## Context files

- `context/project.md`: challenge purpose, current state, and intended layout.
- `context/architecture.md`: service flow, QR assumption, and deployment direction.
- `context/commands.md`: project-specific validation commands.
- `context/conventions.md`: implementation conventions.
- `context/security.md`: proportionate security guidance.

Use the routing in `AGENTS.md`; read only the context relevant to the task beyond the shared project and convention files.
