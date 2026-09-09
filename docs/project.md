# Project

This is a technical challenge that will be a small monorepo with two APIs. Its goal is to accept a matrix operation request, perform QR factorization in Go, and calculate statistics for the resulting matrices in Node.

## Current state

`node-api/` is the only project that may currently exist. Do not assume the other directories or their implementations exist.

## Intended final layout

- `node-api/`: Node.js, Express, TypeScript, Zod, Biome, Vitest, Supertest, `tsx`, and `tsc`; receives factorization matrices and calculates statistics.
- `go-api/`: Go and Fiber; exposes the client-facing API and performs QR matrix factorization.
- `infrastructure/`: Terraform for the Google Cloud deployment.

Each project owns its own source, dependencies, commands, and configuration. Avoid changes across project boundaries unless the approved task requires their integration.
