# AI Agent Structure

This directory contains the English agent workflow and project guidance for Forge TAZ.

`AGENTS.md` is the canonical entry point. This directory is split by purpose:

- `agents/` contains workflow agent definitions.
- `context/` contains project knowledge and operating rules.

## Workflow Agents

- `agents/orchestrator.md` - coordinates the full feature workflow and human approvals.
- `agents/explorer.md` - clarifies feature scope from source documents.
- `agents/planner.md` - converts approved scope into an implementation design.
- `agents/developer.md` - executes an approved design without changing architecture ad hoc.
- `agents/documenter.md` - writes feature documentation after implementation.
- `agents/qa.md` - analyzes the implementation and produces test cases.

## Project Context

- `context/project.md` - repository overview and ownership map.
- `context/architecture.md` - runtime architecture and data flow.
- `context/form-development.md` - required pattern for adding new document-generation request forms.
- `context/forge.md` - Forge CLI, manifest, deploy, install, and tunnel rules.
- `context/commands.md` - local commands and validation.
- `context/conventions.md` - coding and documentation conventions.
- `context/security.md` - permission, API, and data-safety guidance.

## Reading Rule

Every workflow agent must consider `AGENTS.md` first, then this README, then only the context files selected by the routing below.

## Context Routing

Use the smallest context set that covers the requested work. Do not load every file under `context/` by default.

For all feature planning or implementation work, start with:

- `context/project.md`
- `context/architecture.md`
- `context/conventions.md`
- `context/commands.md`

For new or changed document-generation request forms, also read:

- `context/form-development.md`

For Forge platform changes, also read:

- `context/forge.md`

Forge platform changes include manifest modules, function keys, triggers, scopes, permissions, external egress, Forge CLI workflows, deployment, installation, and tunneling.

For security-sensitive changes, also read:

- `context/security.md`

Security-sensitive changes include secrets, credentials, customer data, external APIs, resolver payloads, attachment handling, Jira field hydration, Assets object references, request participants, and any data persisted through Forge or Jira.

If a task touches multiple categories, combine the selected context files. If the selected set is insufficient after inspecting the source, read the additional specific context file needed and explain why in the agent output.
