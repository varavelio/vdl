# Agent Context for VDL

## Summary

VDL (Varavel Definition Language) is a monorepo centered on a Go toolchain (`toolchain/`) that provides the CLI, parser, analyzer, formatter, and LSP for `.vdl` schemas. The repository also contains docs, editor integrations, and installers.

## Maintaining this Document

After completing any task, review this file and update it if you made structural changes or discovered patterns worth documenting. Only add information that helps understand how to work with the project. Avoid implementation details, file listings, or trivial changes. This is a general guide, not a changelog.

When updating this document, do so with the context of the entire document in mind; do not simply add new sections at the end, but place them where they make the most sense within the context of the document.

## Architecture & Organization

### Root Layout

- `Taskfile.yml`: Central command orchestration for CI, test, lint, format, release, docs, and VS Code extension tasks.
- `toolchain/`: Go codebase for the `vdl` binary (CLI + LSP + formatter + compiler core).
- `schemas/`: VDL schemas for internal contracts (`ir.vdl`, `plugin.vdl`, `pluginInput.vdl`, `pluginOutput.vdl`).
- `docs/`: Astro/Starlight documentation site.
- `editors/`: Editor integrations (`vscode/`, `neovim/`).
- `installers/`: Distribution installers (`brew/`, `npm/`, `unix/`, `windows/`).
- `scripts/`: Release automation (`scripts/release/main.go`).
- `assets/`: Branding/static assets used by docs and packaging.
- `.github/workflows/`: CI, docs deployment, and release workflows.
- `temp/`: Local scratch/output area; treat as non-source, disposable workspace.

### `toolchain/` (Core Go Project)

- **Role**: Implements language tooling and project analysis for VDL.
- **Entry points**:
  - `cmd/vdl/main.go`: CLI entry point (`vdl init`, `vdl format`, `vdl generate`, `vdl lsp`, `vdl version`).
- **Key directories**:
  - `internal/core/`: Compiler pipeline (`vfs`, `parser`, `ast`, `analysis`, `ir`).
  - `internal/formatter/`: Lexer-based formatter implementation and golden tests.
  - `internal/lsp/`: Language Server handlers (definition/hover/references/rename/completion/document symbols/links).
  - `internal/codegen/`: Generation entrypoint currently under migration; plugin-oriented direction. `Run` loads `vdl.config.vdl` by analyzing it as VDL, builds IR, decodes `const config` into `configtypes.VdlConfig`, resolves plugin sources (local, HTTPS, or GitHub shorthand), caches remote dependencies under the VDL cache directory, and persists remote hashes in `vdl.lock` JSON using `locktypes`.
  - `internal/dirs/`: VDL runtime directory helpers for home/cache/log resolution and log file creation.
  - `internal/util/`: Shared helpers (`cliutil`, `strutil`, `filepathutil`, etc.).
  - `tests/`: Currently only repository notes (`README.md`); legacy E2E suite has been removed.

### `schemas/` (Contracts)

- **Role**: Source-of-truth schema contracts consumed by generation and docs artifacts.
- **Key files**:
  - `ir.vdl`: Flattened/resolved IR contract for generators.
  - `plugin.vdl`: Plugin protocol umbrella schema.
  - `pluginInput.vdl`: Input payload sent to plugins over stdin.
  - `pluginOutput.vdl`: Output payload returned by plugins over stdout.
- **Important direction**: New generator behavior should align with plugin contracts instead of hardcoded language generators.

### `docs/` (Documentation Site)

- **Role**: Public documentation and schema artifact publishing.
- **Key paths**:
  - `src/content/docs/`: Markdown/MDX documentation source.
  - `public/schemas/v0/`: Published schema artifacts.

### `editors/` (IDE Integrations)

- `editors/vscode/`: VS Code extension (`src/extension.js` is runtime entry).
- `editors/neovim/`: Neovim integration instructions.

### `installers/` (Distribution)

- `installers/brew/`: Formula generator for Homebrew tap updates.
- `installers/npm/`: npm package wrapper/installer for `vdl`.
- `installers/unix/`, `installers/windows/`: shell and PowerShell installers.

## General Instructions (Constitution)

- **Always re-scan structure first**: run `task --list-all` and inspect relevant directories before changing behavior.
- **Trust current tree, not historical assumptions**: `playground/` and legacy E2E layout are removed; do not reintroduce WASM/playground coupling unless explicitly requested.
- **VDL model alignment**: keep parser/analysis/lsp changes declaration-centric (`include`, docstrings, `type`, `enum`, `const`, annotations, spreads). Avoid legacy RPC/pattern/proc/stream assumptions.
- **Plugin-first generation**: treat `schemas/plugin*.vdl` as the contract for generator integrations; prefer extending plugin flows rather than adding fixed built-in generators.
- **Command policy**: prefer root Taskfile tasks; use direct subproject commands only when a focused Taskfile task does not exist.
- **Formatting/linting split**: Go uses `go fmt` + `vdl format`; JS/TS/Astro/Svelte use Biome; JSON/YAML/Markdown/CSS/HTML-like formats use dprint.

## Testing & Quality

- **Primary test strategy**: package-level Go unit tests (`*_test.go`) in `toolchain/internal/**`.
- **Codegen plugin runtime tests**: `toolchain/internal/codegen/testdata/run_plugin/cases/` uses folder-driven fixtures (`index.js` + `expected.json`) so new plugin execution scenarios can be added without test harness changes.
- **Formatter quality**: relies on golden-style fixtures in `toolchain/internal/formatter/tests/`.
- **LSP quality**: behavior tests live in `toolchain/internal/lsp/*_test.go` and should stay aligned with the current declaration-centric AST/program model.
- **E2E note**: `toolchain/tests/` no longer contains the old multi-language E2E harness; do not assume legacy `testdata`-driven E2E structure exists.
- **Verification commands**:
  - `task test`
  - `task lint`
  - `task format`

## Tech Stack & Conventions

- **Go**: 1.26 (`toolchain/go.mod`), with `participle`, `go-arg`, `testify`, and JSON schema tooling.
- **Docs**: Astro 5 + Starlight + Tailwind CSS 4 (`docs/`).
- **Editor extension**: VS Code extension in Node/JS with `vscode-languageclient`.
- **Monorepo JS tooling**: Biome and dprint are the main cross-project format/lint tools.
- **Code style goals**:
  - Idiomatic Go with clear, small functions and explicit error handling.
  - Keep LSP logic robust under partial/invalid code (best-effort behavior is expected).
  - Keep architecture docs and contracts synchronized when changing schema-level behavior.

## Operational Commands

- **Discover commands**: `task --list-all`
- **Core verification**: `task test`, `task lint`, `task format`
- **Build**: `task build`
- **Dependencies**: `task deps`
- **Codegen workflow**: `task codegen`
- **Install local CLI**: `task tc:install`
- **Docs**: `task dc:dev`, `task dc:build`
- **VS Code extension**: `task vs:dev`, `task vs:build`, `task vs:package`, `task vs:package:ls`
- **Release pipeline**: `task release`

### Current Caveats

- `task tc:build:wasm` exists in `Taskfile.yml`, but the corresponding `toolchain/cmd/vdlwasm` entrypoint is currently absent.
