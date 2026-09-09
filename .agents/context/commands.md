# Commands

Run commands from the project directory they apply to.

## Node API

The current `node-api/package.json` provides:

```bash
cd node-api
npm run dev          # tsx watch src/index.ts
npm run typecheck    # tsc --noEmit
npm run check        # Biome lint, formatting, and assists check
npm run check:fix    # apply Biome safe fixes
npm run format       # format with Biome
npm run format:check # verify formatting
npm test             # Vitest, one run
npm run test:watch   # Vitest watch mode
npm run build        # compile TypeScript to dist/
npm start            # run compiled JavaScript
```

## Future Go API

When `go-api/` exists, expected validation conventions are `gofmt` (or `go fmt`), `go vet`, `go test`, and `go build`. These commands have not been tested in this repository.

## Future infrastructure

When Terraform exists, use `terraform fmt` and `terraform validate`. These commands have not been tested in this repository.
