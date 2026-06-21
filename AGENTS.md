# Agent Context for VDL

## Summary

VDL (Varavel Definition Language) is a monorepo centered on a Go toolchain (`toolchain/`) that provides the CLI, parser, analyzer, formatter, and LSP for `.vdl` schemas. The repository also contains docs and external integrations for editors, syntax grammars, and distribution installers.

## Maintaining this Document

After completing any task, review this file and update it if you made structural changes or discovered patterns worth documenting. Only add information that helps understand how to work with the project. Avoid implementation details, file listings, or trivial changes. This is a general guide, not a changelog.

When updating this document, do so with the context of the entire document in mind; do not simply add new sections at the end, but place them where they make the most sense within the context of the document.

## Architecture & Organization

### Root Layout

- `Taskfile.yml`: Central command orchestration for CI, test, lint, format, release, codegen, build, and VS Code extension tasks.
- `toolchain/`: Go codebase for the `vdl` binary (CLI + LSP + formatter + compiler core).
- `schemas/`: VDL schemas for internal contracts (`ir.vdl`, `plugin.vdl`, `plugin_input.vdl`, `plugin_output.vdl`).
- `e2e/`: End-to-end golden tests that build the real `vdl` binary, run `vdl generate` with a small JS IR-dump plugin, and compare generated plugin IR against `output.json` fixtures.
- `docs/`: Zola documentation site using the vendored VaraPress theme.
- `integrations/`: External integrations grouped by purpose (`editors/`, `syntax/`, `installers/`).
- `scripts/`: Release automation (`scripts/release/main.go`).
- `assets/`: Branding/static assets used by docs and packaging.
- `.github/workflows/`: CI, docs deployment, and release workflows.
- `temp/`: Local scratch/output area; treat as non-source, disposable workspace.

### `toolchain/` (Core Go Project)

- **Role**: Implements language tooling and project analysis for VDL.
- **Entry points**:
  - `cmd/vdl/main.go`: CLI entry point (`vdl init`, `vdl format`, `vdl generate`, `vdl compile`, `vdl lsp`, `vdl version`).
- **Key directories**:
  - `internal/core/`: Compiler pipeline (`vfs`, `parser`, `ast`, `analysis`, `ir`).
  - `internal/formatter/`: Lexer-based formatter implementation and golden tests.
  - `internal/lsp/`: Language Server handlers (definition/hover/references/rename/completion/document symbols/links).
  - `internal/codegen/`: Generation entrypoint currently under migration; plugin-oriented direction. `Run` loads `vdl.config.vdl` by analyzing it as VDL, decodes `const config` into `configtypes.VdlConfig`, executes optional global host hooks (`hooks.preGenerate` fail-fast, `hooks.postGenerate` warn-and-continue), resolves plugin sources (local, HTTPS, or GitHub shorthand), caches remote dependencies under the VDL cache directory, and persists remote hashes in `vdl.lock` JSON using `locktypes`.
  - `internal/dirs/`: VDL runtime directory helpers for home/cache/log resolution and log file creation.
  - `internal/util/`: Shared helpers (`cliutil`, `strutil`, `filepathutil`, etc.).
  - `tests/`: Currently only repository notes (`README.md`); legacy E2E suite has been removed.

### `schemas/` (Contracts)

- **Role**: Source-of-truth schema contracts consumed by generation and docs artifacts.
- **Key files**:
  - `config_file.vdl`: `vdl.config.vdl` contract used by the generator runtime.
  - `ir.vdl`: Flattened/resolved IR contract for generators.
  - `lock_file.vdl`: `vdl.lock` contract for cached remote plugin hashes.
  - `plugin.vdl`: Plugin protocol umbrella schema.
  - `plugin_input.vdl`: Input payload passed to plugins.
  - `plugin_output.vdl`: Output payload returned by plugins.
- **Important direction**: New generator behavior should align with plugin contracts instead of hardcoded language generators.

### `docs/` (Documentation)

- **Role**: Public Zola documentation and landing site for VDL, rendered with the VaraPress theme.
- **Key paths**:
  - `zola.toml`: Zola site configuration and VaraPress theme options.
  - `content/_index.md`: Home landing page composed with VaraPress landing shortcodes.
  - `content/docs/_index.md`: Docs root section; uses `template = "docs.html"` and `sort_by = "weight"`.
  - `content/docs/about.md`: Project overview and positioning.
  - `content/docs/language/_index.md`: Language guide section index; child pages are ordered by `weight`.
  - `content/docs/guides/_index.md`: Practical guides section index; child pages are ordered by `weight`.
  - `content/docs/reference/_index.md`: Reference section index; child pages are ordered by `weight`.
  - `themes/varapress/`: Vendored VaraPress theme. Treat as third-party theme code unless the task explicitly requires theme changes.
- **Content conventions**:
  - Use Zola TOML frontmatter delimited by `+++` for all docs pages and sections.
  - Use explicit `weight` values to control sidebar and previous/next ordering.
  - Use clean Zola paths such as `/docs/guides/installation/` for internal links instead of old `.md` links.
  - Compose the landing page with VaraPress shortcodes; use each shortcode's own `container` parameter rather than wrapping standard landing sections.

### `integrations/` (External Integrations)

- **Role**: Contains adapters and packaging that connect VDL to external editors, syntax tooling, package managers, and distribution platforms.
- **Key directories**:
  - `integrations/editors/vscode/`: VS Code extension (`src/extension.js` is runtime entry). The packaged grammar file under `language/vdl.tmLanguage.json` is generated from the shared TextMate grammar before VS Code builds and packaging.
  - `integrations/editors/neovim/`: Neovim integration instructions.
  - `integrations/syntax/textmate/`: Source-of-truth VDL TextMate grammar (`vdl.tmLanguage.json`) and grammar fixtures, tested with `vscode-tmgrammar-test` through the root Taskfile.
  - `integrations/installers/brew/`: Formula generator for Homebrew tap updates.
  - `integrations/installers/npm/`: npm package wrapper/installer for `vdl`.
  - `integrations/installers/unix/`, `integrations/installers/windows/`: shell and PowerShell installers.
  - `integrations/installers/docker/`: Official Docker image definition used by release publishing.

## General Instructions (Constitution)

- **Always re-scan structure first**: run `task --list-all` and inspect relevant directories before changing behavior.
- **Trust current tree, not historical assumptions**: `playground/` and legacy E2E layout are removed; do not reintroduce WASM/playground coupling unless explicitly requested.
- **VDL model alignment**: keep parser/analysis/lsp changes declaration-centric (`include`, docstrings, `type`, `enum`, `const`, annotations, spreads). Avoid legacy RPC/pattern/proc/stream assumptions.
- **Constants are dynamic literals**: `const` declarations do not accept explicit type names.
- **Plugin-first generation**: treat `schemas/plugin*.vdl` as the contract for generator integrations; prefer extending plugin flows rather than adding fixed built-in generators.
- **Shared TextMate grammar**: edit `integrations/syntax/textmate/vdl.tmLanguage.json` as the canonical grammar source. Do not hand-edit the generated VS Code copy; the `task vs:build` automatically copies it.
- **Command policy**: prefer root Taskfile tasks; use direct subproject commands only when a focused Taskfile task does not exist.
- **Formatting/linting split**: Go uses `go fmt` + `vdl format`; JS/TS/Astro/Svelte use Biome; JSON/YAML/Markdown/CSS/HTML-like formats use dprint.

## Testing & Quality

- **Primary test strategy**: package-level Go unit tests (`*_test.go`) in `toolchain/internal/**`.
- **Codegen plugin runtime tests**: `toolchain/internal/codegen/testdata/run_plugin/cases/` uses folder-driven fixtures (`index.js` + `expected.json`) so new plugin execution scenarios can be added without test harness changes.
- **Formatter quality**: relies on golden-style fixtures in `toolchain/internal/formatter/tests/`.
- **LSP quality**: behavior tests live in `toolchain/internal/lsp/*_test.go` and should stay aligned with the current declaration-centric AST/program model.
- **TextMate grammar quality**: TextMate grammar fixtures live in `integrations/syntax/textmate/test/` and run through `task tml:test` using `vscode-tmgrammar-test`. `task vs:test:grammar` is an alias for the shared grammar test.
- **IR golden stability**: IR golden JSON fixtures under `toolchain/internal/core/ir/testdata/*.json` intentionally omit `position`; `toolchain/internal/core/ir/ir_test.go` normalizes generated output via `toolchain/internal/util/testutil/ir.go` (`StripPositionsFromJSON` / `IRJSONEqualNoPos`).
- **E2E IR contract tests**: Root `e2e/cases/*` fixtures use one folder per case. Each case has `input.vdl` and `output.json`, with optional extra files for includes or external docs. The shared fixtures `e2e/vdl.config.vdl` and `e2e/plugin.js` are copied into a temporary project to exercise `vdl generate`, JS plugin execution, and plugin-facing IR output. Positions and absolute entry paths are normalized for stable goldens.
- **E2E note**: `toolchain/tests/` no longer contains the old multi-language E2E harness; do not assume legacy `testdata`-driven E2E structure exists.
- **Verification commands**:
- `task ci` (runs all verifications)
- `task test`
- `task lint`
- `task format`

## Tech Stack & Conventions

- **Go**: 1.26 (`toolchain/go.mod`), with `participle`, `go-arg`, `testify`, and JSON schema tooling.
- **Docs**: Zola site in `docs/` using VaraPress; content is Markdown with TOML frontmatter and is formatted with dprint.
- **Editor extension**: VS Code extension in `integrations/editors/vscode/` with Node/JS and `vscode-languageclient`.
- **Syntax grammar**: Shared TextMate grammar package in `integrations/syntax/textmate/`, consumed by docs highlighting and copied into the VS Code extension during build/package tasks.
- **Monorepo JS tooling**: Biome and dprint are the main cross-project format/lint tools.
- **Code style goals**:
  - Idiomatic Go with clear, small functions and explicit error handling.
  - Keep LSP logic robust under partial/invalid code (best-effort behavior is expected).
  - Keep architecture docs and contracts synchronized when changing schema-level behavior.

## Operational Commands

- **Discover commands**: `task --list-all` (always execute this)
- **Core verification**: `task test`, `task lint`, `task format`
- **E2E IR golden tests**: `task test:e2e`
- **Build**: `task build`
- **Docs build**: `task docs`
- **Dependencies**: `task deps`
- **Codegen workflow**: `task codegen`
- **Install local CLI**: `task tc:install`
- **Syntax grammar**: `task tml:test`
- **VS Code extension**: `task vs:build`, `task vs:test:grammar`, `task vs:package`, `task vs:package:ls`
- **Release pipeline**: `task release`
