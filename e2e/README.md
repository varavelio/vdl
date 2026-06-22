# E2E IR Golden Tests

These tests validate the real `vdl generate` flow end to end. The harness builds the `vdl` binary, creates a temporary VDL project for each case, runs generation with a small JavaScript plugin, and compares the plugin-facing IR against a golden `output.json` file.

## Flow

1. `e2e_test.go` builds the real CLI from `toolchain/cmd/vdl`.
2. Each folder under `cases/` is copied into a temporary project.
3. The shared `vdl.config.vdl` fixture is copied into the temporary project root.
4. The shared `plugin.js` fixture is copied into the temporary project root.
5. The test runs `vdl generate <temp-project>`.
6. The plugin writes the normalized IR to `gen/output.json`.
7. The test compares `gen/output.json` with the case's `output.json`.

## Case Layout

Each case must include:

```txt
cases/<name>/
  input.vdl
  output.json
```

Cases may include extra files when needed, such as included `.vdl` files or external Markdown docs referenced by docstrings.

## Coverage

The case suite is intended to cover the documented valid language surface and the full generator-facing IR shape:

- declarations: `type`, `enum`, `const`, includes, and standalone docs
- type refs: primitives, aliases, custom types, enums, arrays, multidimensional arrays, maps, nested maps, inline objects, optional fields, and recursive optional references
- spreads: object type spreads, inline object spreads, enum spreads, and object literal spreads
- literals: strings, ints, floats, booleans, arrays, objects, empty arrays, empty objects, constant references, enum member references, and deeply nested literals
- annotations: declaration, field, enum member, argument-less, scalar arguments, array arguments, object arguments, constant references, and enum references
- docs: attached docstrings, standalone docstrings, external Markdown docs from entrypoint files, and external Markdown docs from included files
- parser edge cases: line comments, block comments, inline comments, compact whitespace, deep nesting, and deterministic top-level ordering

## Stable Goldens

The plugin removes every nested `position` field and rewrites `entryPoint` to `input.vdl`. This keeps goldens stable across machines and temporary directories while still validating the generator-facing IR shape.

## Run tests

Run:

```sh
task test:e2e
```

## Updating Goldens

Run:

```sh
go test -C ./e2e ./... -update
```
