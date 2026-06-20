# Change Log

## v0.1.5 - 2026-03-09

- Added workspace-local VDL binary discovery through `node_modules/.bin`, including parent directories up to the filesystem root.
- Prioritized project-local VDL installations before `GOBIN` and `PATH` to better match monorepo and dev dependency workflows.
- Unified binary execution so `vdl`, `vdl.exe`, `vdl.cmd`, and `vdl.bat` work consistently for commands and the language server.

## v0.1.4 - 2026-03-03

- Rewrote syntax highlighting grammar to align with current VDL spec and remove legacy rules.
- Rewrote snippets to match current VDL syntax.

## v0.1.3 - 2026-02-02

- Fixed syntax highlighting for keywords used as field names (e.g., `input: string`, `output: int`, `type: bool`).

## v0.1.2 - 2026-02-01

- Refactored VDL syntax highlighting with granular declarations and improved pattern matching.

## v0.1.1 - 2026-01-30

- Added `vdl.openLogs` command to easily open the language server log file.

## v0.1.0 - 2026-01-29

- Initial release of VDL for VSCode (formerly UFO RPC).
