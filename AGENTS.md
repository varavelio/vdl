# AGENTS.md

## 1. Summary

**UfoRPC** is a Universal RPC schema and generator tool located in a monorepo. It combines a Go-based core (`urpc`) with a Svelte 5-based web playground. The Go core compiles to both a native CLI and a WebAssembly (WASM) binary. The WASM binary is consumed by the playground to provide client-side schema generation and validation directly in the browser. The final build embeds the static playground assets back into the Go binary.

## 2. General Instructions (The Constitution)

- **Context Awareness**: Always respect the monorepo structure. There are distinct environments for Go (`urpc/`) and Node/Svelte (`playground/`).
- **Command Authority**: The root `Taskfile.yml` is the single source of truth for orchestration. Do not run `npm` or `go` commands manually if a `task` command exists for it.
- **Verification**:
  - Always run `task fmt` to handle multi-language formatting (Prettier + Go Fmt).
  - Always run `task lint` to check both Go (golangci-lint) and Svelte/TS (Biome).
  - Always run `task test` to verify integrity.
- **Dependency Management**:
  - Go: Manage in `urpc/go.mod`.
  - Node: Manage in `playground/package.json` or root `package.json` for dev tools.
- **Code Style**:
  - Go: Idiomatic, `golangci-lint` strictness.
  - Svelte: Functional components, Svelte 5 runes syntax (if applicable/modern), Tailwind CSS v4.

## 3. Architecture & Organization

### Root Layout

- `Taskfile.yml`: Orchestrates the entire build pipeline across languages.
- `urpc/`: The Go backend/core.
- `playground/`: The Svelte 5 frontend.

### `urpc/` (Go Core)

- **Role**: Contains the business logic for parsing and generating RPC schemas.
- **Key Directories**:
  - `cmd/urpc`: Main CLI entry point (native).
  - `cmd/urpcwasm`: Entry point for WASM compilation (browser target).
  - `internal/`: Private library code.
  - `dist/`: Build artifacts.
- **Integration**: Compiles to `dist/urpc.wasm` which is copied to the playground.

### `playground/` (Frontend)

- **Role**: A visual editor/playground for UfoRPC.
- **Tech**: Svelte 5, Vite, TailwindCSS 4, Biome.
- **Integration**:
  - Consumes `urpc.wasm` (via `wasm_exec.js`).
  - Generates Typescript definitions from Go schemas via `npm run genschema`.
  - Builds to static files that are eventually embedded into the Go binary.

### Build Pipeline (Circular Dependency Resolved via Steps)

1. **WASM Build**: Go code compiles to WASM (`task build:wasm`).
2. **Frontend Build**: Playground consumes WASM and builds static assets (`npm run build`).
3. **CLI Build**: Go CLI embeds the static assets and compiles the final binary (`task build:urpc`).

## 4. Testing & Quality

- **Strategy**:
  - Unit tests for Go (`go test ./...`).
  - Component/Logic tests for Frontend (`vitest`).
- **Commands**: Run `task test` at the root to execute all suites.
- **Linting**: Run `task lint` (Triggers `golangci-lint` for Go and `biome` for JS/TS).

## 5. Tech Stack & Conventions

### Go

- **Version**: 1.24+
- **Key Libs**: `go-arg`, `participle`, `ufogenkit`.
- **Patterns**: Standard Go project layout (`cmd`, `internal`).

### Frontend

- **Framework**: Svelte 5 (Runes preferred).
- **Bundler**: Vite.
- **Styles**: Tailwind CSS 4.
- **Linter/Formatter**: Biome.

## 6. Operational Commands

**Run these from the project root:**

- **Setup/Install**: `task deps` (Installs Node modules and Go mods).
- **Build All**: `task build` (Handles the WASM -> Frontend -> CLI pipeline).
- **Test All**: `task test`.
- **Lint All**: `task lint`.
- **Format All**: `task fmt`.
