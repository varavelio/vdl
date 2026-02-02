# AGENTS.md

## 1. Summary

**VDL (Varavel Definition Language)** is a Universal RPC schema and generator tool located in a monorepo. It combines a Go-based core (`toolchain`) with a Svelte 5-based web playground. The Go core compiles to both a native CLI and a WebAssembly (WASM) binary. The WASM binary is consumed by the playground to provide client-side schema generation, validation, formatting, and transformation directly in the browser. The final build embeds the static playground assets back into the Go binary.

## Maintaining this Document

After completing any task, review this file and update it if you made structural changes or discovered patterns worth documenting. Only add information that helps understand how to work with the project. Avoid implementation details, file listings, or trivial changes. This is a general guide, not a changelog.

When updating this document, do so with the context of the entire document in mind; do not simply add new sections at the end, but place them where they make the most sense within the context of the document.

## 2. General Instructions (The Constitution)

- **Context Awareness**: Always respect the monorepo structure. There are distinct environments for Go (`toolchain/`), Node/Svelte (`playground/`), and Editor Extensions (`editors/`).
- **Command Authority**: The root `Taskfile.yml` is the **single source of truth** for orchestration.
  - **Do NOT** run `npm`, `go`, `vite`, or `vsce` commands manually.
  - **ALWAYS** run `task --list-all` from the project root to discover the available commands before starting any task.
  - **Execution**: All commands are run via `task <command>`.
- **Verification**:
  - Always check available verification tasks (lint, test, format) via `task --list-all`.
  - Ensure standard checks pass before finishing work.
- **Dependency Management**:
  - Go: Manage in `toolchain/go.mod`.
  - Node: Manage in `playground/package.json` or root `package.json` for dev tools.
- **Code Style**:
  - Go: Idiomatic, `golangci-lint` strictness.
  - Svelte: Functional components, Svelte 5 runes syntax, Tailwind CSS v4, DaisyUI v5.
- **Terminology**: The core tool is referred to as `toolchain`. The binary is named `vdl`.

## 3. Architecture & Organization

### Root Layout

- `Taskfile.yml`: Orchestrates the entire build pipeline across languages.
- `toolchain/`: The Go backend/core (Compiler, CLI, WASM).
- `playground/`: The Svelte 5 frontend (See `playground/AGENTS.md`).
- `editors/`: Editor plugins (VS Code, Neovim).
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
  - `tests/`: End-to-End (E2E) test suite, organized by target language/format.
    - `golang/`: E2E tests for Go code generation (verifies client/server behavior via `go run`).
    - `typescript/`: E2E tests for TypeScript code generation (verifies client/server behavior via `tsx`).
    - `dart/`: E2E tests for Dart code generation (verifies client/server behavior via `dart run`).
    - `python/`: E2E tests for Python code generation (verifies client/server behavior via `python3`).
    - `jsonschema/`: E2E tests for JSON Schema generation (verifies output against expected JSON).
    - `openapi/`: E2E tests for OpenAPI generation (verifies output against expected YAML/JSON).
    - `playground/`: E2E tests for Playground assets generation.
    - `plugin/`: E2E tests for the Plugin system (verifies Python plugin integration).
    - Each test case in `testdata` contains a schema, config, and verification assets (consumer program or expected output).
  - `dist/`: Build artifacts (e.g., `vdl.wasm`).
- **Integration**: Compiles to `dist/vdl.wasm` which is copied to the playground.

### `playground/` (Frontend)

- **Role**: A visual editor/playground for VDL.
- **Tech**: Svelte 5 (Runes), Vite, TailwindCSS 4, DaisyUI 5, Biome.
- **Integration**:
  - Consumes `vdl.wasm` (via `wasm_exec.js`).
  - Generates Typescript definitions from Go schemas.
  - Builds to static files that are eventually embedded into the Go binary.

### `editors/` (IDE Integrations)

- **Role**: Provides syntax highlighting and LSP support for VDL.
- **Key Directories**:
  - `vscode/`: Visual Studio Code extension (Node.js/TS).
  - `neovim/`: Neovim plugin (Markdown instructions).

### `docs/` (Documentation)

- **Role**: Official documentation and specifications.
- **Key Directories**:
  - `spec/`: The formal VDL specification (`spec.md`, `rpc-request-lifecycle.md`).
  - `public/`: Assets for the documentation site.

### Build Pipeline

1. **WASM Build**: Go code compiles to WASM.
2. **Frontend Build**: Playground consumes WASM and builds static assets.
3. **CLI Build**: Go CLI embeds the static assets and compiles the final binary.

### WASM Communication Protocol

The communication between the Playground (Svelte) and the Core (Go) relies on a strict **JSON-RPC style protocol** defined in VDL itself.

1.  **The Contract (`schemas/wasm.vdl`)**:
    - This is the source of truth. It defines the `WasmInput` and `WasmOutput` data structures and the `WasmFunctionName` enum.

2.  **The Bridge**:
    - **Frontend (`playground/src/lib/wasm/index.ts`)**: The `WasmClient` serializes requests to JSON and calls the global `window.wasmExecuteFunction`.
    - **WASM Entry (`toolchain/cmd/vdlwasm/main.go`)**: This Go program runs in the browser. It attaches the `wasmExecuteFunction` to the Javascript `window` object. It wraps the execution in a Javascript Promise to handle async operations.

3.  **The Execution (`toolchain/internal/wasm/run.go`)**:
    - Receives the JSON string.
    - Unmarshals it into Go structs (generated from `wasm.vdl`).
    - Routes the request based on `FunctionName` (e.g., `Codegen`, `ExpandTypes`, etc) to the internal logic.
    - Returns the response as a JSON string.

## 4. Testing & Quality

### Strategy

- **Unit Tests**: Standard Go tests for internal logic.
- **E2E Tests (`toolchain/tests/*`)**:
  - These are **integration tests** wrapped in Go test runners (`e2e_test.go`) located in each subdirectory.
  - **Mechanism**: The runner builds a temporary `vdl` binary from the current source, then for each case in `testdata/`:
    1.  Runs the generation command via the test harness.
    2.  **Verification**:
        - For **Code Gen**: Executes the consumer program. The consumer code must **panic** or exit with non-zero status on failure.
        - For **Schemas/Docs**: Compares the generated output against a "golden" expected file.
        - For **Plugins**: Executes the plugin and verifies the JSON output or error handling.
- **Frontend Tests**: Component/Logic tests in `playground/`.

### Adding a New E2E Test Case

1.  Navigate to `toolchain/tests/<category>/testdata`.
2.  Create a new directory for your case (e.g., `my_feature`).
3.  Add the required files:
    - `vdl.yaml`: Configuration for generation.
    - `schema.vdl`: The schema to test.
    - **For Code Gen**: `main.go`, `main.ts`, `main.dart`, or `main.py` (or equivalent) to import and verify generated code.
    - **For Schemas**: `expected.json` or `expected.yaml` to match against.

### Commands

Do NOT rely on memory. Run `task --list-all` to find the appropriate testing command (e.g., for running all tests, specific toolchain tests, frontend tests, etc).

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
- **Styles**: Tailwind CSS 4, DaisyUI 5.
- **Linter/Formatter**: Biome.

## 6. Operational Commands

**IMPORTANT**: All operational commands are centralized in `Taskfile.yml` and must be executed from the project root. There is no "npm scripts", "per subfolder Taskfile.yml" or other things, literally the source of truth is the `Taskfile.yml` in the repository root.

1.  **Discover**: Run the following command to see all available tasks and their descriptions:
    ```bash
    task --list-all
    ```
2.  **Execute**: Run a specific task using:
    ```bash
    task <command_name>
    ```

**Do not guess commands.** The `Taskfile.yml` is the only source of truth for building, testing, linting, formatting, and releasing this project. You can list all the commands or read the file but don't guess random commands.
