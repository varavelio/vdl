# VDL E2E Tests

This directory contains End-to-End (E2E) tests for the VDL toolchain, organized by the target output (language or format).

## Structure

Tests are categorized by the target they verify:

- `golang/`: Verifies Go code generation.
- `typescript/`: Verifies TypeScript code generation.
- `dart/`: Verifies Dart code generation.
- `python/`: Verifies Python code generation.
- `jsonschema/`: Verifies JSON Schema generation.
- `openapi/`: Verifies OpenAPI (Swagger) generation.
- `playground/`: Verifies the WebAssembly/Playground asset generation.
- `plugin/`: Verifies the plugin system integration.

## How it works

Each category directory contains its own `e2e_test.go` runner and a `testdata/` folder.

1. **Test Runner**: The Go test runner builds a temporary `vdl` binary from the current source code.
2. **Execution**: It iterates through folders in `testdata/` and runs `vdl generate`.
3. **Verification**:
   - **Code Generation (Golang, TypeScript, Dart, Python)**: It executes the generated code using the respective runtime/tool:
     - Go: `go run .`
     - TypeScript: `tsx main.ts`
     - Dart: `dart run --enable-asserts main.dart`
     - Python: `python3 main.py`
       The test passes if the program exits successfully (0) and fails if it panics or errors.
   - **Schema/Docs (e.g., OpenAPI, JSONSchema)**: It compares the generated output file against a "golden" expected file (e.g., `expected.json`).
   - **Plugins**: It executes the plugin and verifies the JSON output or standard output/error.

## Adding a Test Case

Read one or two tests to see how it works but in general this is the process:

1. Go to the relevant subdirectory (e.g., `golang/testdata` or `openapi/testdata`).
2. Create a new folder for your case.
3. Add `vdl.yaml` and `schema.vdl`.
4. Add verification files:
   - For **Code**: A `main.go`, `main.ts`, `main.dart`, or `main.py` that uses the generated code and panics/fails on error.
   - For **Schemas**: An `expected.json` or `expected.yaml` file to match against.
