# AGENTS.md

## 1. Summary

**VDL (Varavel Definition Language)** is a Universal RPC schema and generator tool located in a monorepo. It combines a Go-based core (`toolchain`) with a Svelte 5-based web playground. The Go core compiles to both a native CLI and a WebAssembly (WASM) binary. The WASM binary is consumed by the playground to provide client-side schema generation, validation, formatting, and transformation directly in the browser. The final build embeds the static playground assets back into the Go binary.

## Maintaining this Document

After completing any task, review this file and update it if you made structural changes or discovered patterns worth documenting. Only add information that helps understand how to work with the project. Avoid implementation details, file listings, or trivial changes. This is a general guide, not a changelog.

When updating this document, do so with the context of the entire document in mind; do not simply add new sections at the end, but place them where they make the most sense within the context of the document.

## 2. General Instructions (The Constitution)

- **Context Awareness**: Always respect the monorepo structure. There are distinct environments for Go (`toolchain/`) and Node/Svelte (`playground/`).
- **Command Authority**: The root `Taskfile.yml` is the single source of truth for orchestration. Do not run `npm` or `go` commands manually if a `task` command exists for it.
- **Verification**:
  - Always run `task fmt` to handle multi-language formatting (Prettier + Go Fmt).
  - Always run `task lint` to check both Go (golangci-lint) and Svelte/TS (Biome).
  - Always run `task test` to verify integrity.
- **Dependency Management**:
  - Go: Manage in `toolchain/go.mod`.
  - Node: Manage in `playground/package.json` or root `package.json` for dev tools.
- **Code Style**:
  - Go: Idiomatic, `golangci-lint` strictness.
  - Svelte: Functional components, Svelte 5 runes syntax, Tailwind CSS v4.
- **Terminology**: The core tool is referred to as `toolchain`. The binary is named `vdl`.

## 3. Architecture & Organization

### Root Layout

- `Taskfile.yml`: Orchestrates the entire build pipeline across languages.
- `toolchain/`: The Go backend/core (Compiler, CLI, WASM).
- `playground/`: The Svelte 5 frontend.
- `docs/`: Documentation and specifications (contains the VDL spec).
- `scripts/`: Build and maintenance scripts.
- `assets/`: Static assets.

### `toolchain/` (Go Core)

- **Role**: Contains the business logic for parsing, analyzing, transforming, and generating RPC schemas.
- **Key Directories**:
  - `cmd/`: Entry points.
    - `vdl/`: Main CLI entry point (native).
    - `vdlwasm/`: Entry point for WASM compilation (browser target).
  - `internal/`: Private library code.
    - `core/`: The Compiler Core Pipeline.
      - `ast/`: Abstract Syntax Tree definitions. **Crucial**: The AST structure groups `proc` and `stream` definitions inside `rpc` blocks.
      - `parser/`: Lexical analysis and parsing.
      - `analysis/`: Semantic analysis and symbol resolution.
      - `ir/`: Intermediate Representation for generators.
      - `vfs/`: Virtual File System.
    - `transform/`: AST Transformations (Used by Playground/LSP).
      - `expand.go`: Handles type flattening (spreads) and circular reference protection.
      - `extract.go`: Logic to extract specific AST nodes (Types, RPCs, Procs, Streams) as standalone strings.
    - `formatter/`: Source code formatting logic (`vdl fmt`).
    - `lsp/`: Language Server Protocol implementation.
    - `codegen/`: Code Generators (Go, TS, Dart, etc.).
    - `cmd/schemagen/`: Internal tool to generate JSON schemas for IR and Config.
    - `util/`: Shared Utilities (strings, paths, debug).
  - `dist/`: Build artifacts (e.g., `vdl.wasm`).
- **Integration**: Compiles to `dist/vdl.wasm` which is copied to the playground.

### `playground/` (Frontend)

- **Role**: A visual editor/playground for VDL.
- **Tech**: Svelte 5 (Runes), Vite, TailwindCSS 4, Biome.
- **Integration**:
  - Consumes `vdl.wasm` (via `wasm_exec.js`).
  - Generates Typescript definitions from Go schemas.
  - Builds to static files that are eventually embedded into the Go binary.

### Build Pipeline

1. **WASM Build**: Go code compiles to WASM (`task build:wasm`) -> `toolchain/dist/vdl.wasm`.
2. **Frontend Build**: Playground consumes WASM and builds static assets (`npm run build`).
3. **CLI Build**: Go CLI embeds the static assets and compiles the final binary (`task build:vdl`).

## 4. Testing & Quality

- **Strategy**:
  - Unit tests for Go (`go test ./...`).
  - Component/Logic tests for Frontend (`vitest`).
- **Commands**: Run `task test` at the root to execute all suites.
- **Linting**: Run `task lint` (Triggers `golangci-lint` for Go and `biome` for JS/TS).

## 5. Tech Stack & Conventions

### Go

- **Version**: 1.25+
- **Key Libs**: `go-arg`, `participle` (parser), `testify`, `jsonschema`.
- **Patterns**: Standard Go project layout (`cmd`, `internal`).
- **AST Handling**:
  - When working with the AST, remember that `Proc` and `Stream` are children of `RPC`.
  - Use `transform` package for AST manipulations intended for display or refactoring.

### Frontend

- **Framework**: Svelte 5 (Runes preferred).
- **Bundler**: Vite.
- **Styles**: Tailwind CSS 4.
- **Linter/Formatter**: Biome.

## 6. Operational Commands

**Run these from the project root:**

- **Run all checks**: `task ci`
- **Setup/Install**: `task deps` (Installs Node modules and Go mods).
- **Build All**: `task build` (Handles the WASM -> Frontend -> CLI pipeline).
- **Test All**: `task test`.
- **Lint All**: `task lint`.
- **Format All**: `task fmt`.
