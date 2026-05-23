---
title: CLI Commands
description: Complete reference for the VDL command-line interface
---

## Overview

The `vdl` CLI is the primary way to interact with VDL projects. It covers the full lifecycle: creating a project, writing and formatting schemas, generating code through plugins, compiling to inspect the intermediate representation, and running the language server for editor integration.

If you have not installed VDL yet, start with the [installation guide](./installation.md).

## Quick Reference

| Command        | Purpose                                         |
| -------------- | ----------------------------------------------- |
| `vdl init`     | Create a new VDL project with schema and config |
| `vdl format`   | Format `.vdl` files in place                    |
| `vdl generate` | Run code generation from `vdl.config.vdl`       |
| `vdl compile`  | Compile a `.vdl` file and print its IR as JSON  |
| `vdl lsp`      | Start the VDL language server                   |
| `vdl version`  | Show VDL version information                    |

## Global Behavior

A few things work everywhere, no matter which subcommand you use.

### Help

Every command supports `--help` and `-h`:

```bash
vdl --help
vdl init --help
vdl generate --help
```

Help output includes the VDL logo plus a description of the command and its flags.

### Version

Use `--version`, `-v`, or the `version` subcommand to see the installed VDL version:

```bash
vdl --version
vdl version
```

### Default Output

Running `vdl` with no subcommand shows the VDL logo and version.

## `vdl init`

Initialize a new VDL project in a directory.

```bash
vdl init
vdl init ./my-project
```

### What It Creates

Two files are written into the target directory:

| File             | Purpose                                           |
| ---------------- | ------------------------------------------------- |
| `schema.vdl`     | A sample schema with types, constants, and RPCs   |
| `vdl.config.vdl` | A commented config file ready for code generation |

The sample config includes commented-out plugin examples so you can uncomment the ones you need or add more as needed.

### Arguments

| Argument | Required | Description                                                |
| -------- | -------- | ---------------------------------------------------------- |
| `path`   | no       | Target directory. Defaults to the current directory (`.`). |

## `vdl format`

Format `.vdl` files in place.

```bash
vdl format
vdl format ./schemas/**/*.vdl
vdl format ./schemas ./other
vdl format --verbose
```

### How It Works

The formatter uses VDL's own lexer and parser to normalize whitespace, indentation, inline spacing, and blank lines so every file follows a consistent style. Files are overwritten in place.

### Arguments

| Argument    | Required | Description                                                 |
| ----------- | -------- | ----------------------------------------------------------- |
| `patterns`  | no       | Glob patterns or directory paths. Defaults to `./**/*.vdl`. |
| `--verbose` | no       | Print each file path as it is formatted.                    |

### Pattern Behavior

- Plain paths like `./schemas` are treated as directories and expanded to `./schemas/**/*.vdl` automatically.
- Standard glob patterns use double-star syntax (`**`) for recursive matching.
- Files without a `.vdl` extension are skipped.
- Duplicate matches are deduplicated, so the same file is never formatted twice.

### Examples

Format every VDL file in the current project:

```bash
vdl format
```

Format a specific set of files with feedback:

```bash
vdl format ./api/*.vdl ./shared/*.vdl --verbose
```

Format everything inside a directory:

```bash
vdl format ./schemas
```

## `vdl generate`

Run code generation from a `vdl.config.vdl` project.

```bash
vdl generate
vdl generate ./my-project
vdl generate --check
```

### What It Does

The generation pipeline works like this:

1. VDL finds and reads `vdl.config.vdl`.
2. Pre-generation hooks run (if configured). The first failure stops the pipeline.
3. Each configured plugin source is resolved:
   - **Local `.js` files** are loaded directly.
   - **HTTPS URLs** are fetched and cached.
   - **GitHub shorthands** like `owner/vdl-plugin-go@v0.1.3` resolve to a remote `dist/index.js` artifact, which is fetched and cached.
4. Each plugin's configured `.vdl` schema is analyzed and converted into the Intermediate Representation (IR).
5. All plugins execute concurrently in an embedded JavaScript runtime.
6. Output file paths are validated to stay inside their configured `outDir`.
7. Conflicting writes between plugins are detected and fail the run.
8. Output directories are cleaned (default) or merged, and generated files are written.
9. Post-generation hooks run (failures print warnings but do not roll back files).
10. The `vdl.lock` file is updated with remote plugin hashes.

### Arguments

| Argument  | Required | Description                                                         |
| --------- | -------- | ------------------------------------------------------------------- |
| `path`    | no       | Directory or path to `vdl.config.vdl`. Defaults to `.` (cwd).       |
| `--check` | no       | Run the full pipeline but skip writing output files. Useful for CI. |

### Config File Discovery

When `path` points to a directory, VDL looks for a file named exactly `vdl.config.vdl` inside it. When `path` points directly to a file, that file is used. The config file must declare a `const config` with the generation settings.

For all configuration options, see the [project configuration guide](./configuration.md).

### `--check` Mode

Use `--check` in CI to verify that generation would succeed without producing side effects:

```bash
vdl generate --check
```

In check mode, VDL runs the full pipeline—hooks, plugin resolution, schema analysis, plugin execution, and output validation—but skips writing files, updating `vdl.lock` and executing post generation hooks. If the pipeline fails, the command exits with a non-zero code, making it suitable for linting and CI workflows.

### Lock File

Remote plugin artifacts are cached and their content hashes are recorded in `vdl.lock`. Commit this file when your project depends on remote plugins. VDL uses it to detect unexpected changes in cached plugins.

## `vdl compile`

Compile a `.vdl` file and print its Intermediate Representation (IR) as formatted JSON to stdout.

```bash
vdl compile ./schema.vdl
```

### What It Produces

The IR is a machine-readable description of everything VDL understands about your schema: types, enums, constants, annotations, documentation, and relationships. It is the same IR that plugins receive as `input.ir`.

### Arguments

| Argument | Required | Description                         |
| -------- | -------- | ----------------------------------- |
| `file`   | yes      | Path to the `.vdl` file to compile. |

### Use Cases

- Inspect what your schema looks like after analysis, before generation.
- Debug why a plugin produces unexpected output.
- Pipe the IR into custom tooling:

  ```bash
  vdl compile ./schema.vdl | jq '.types | length'
  vdl compile ./schema.vdl > schema-ir.json
  ```

### Error Handling

If the schema has errors, diagnostics are printed to stderr and the command exits with code 1. No partial JSON is emitted.

## `vdl lsp`

Start the VDL language server.

```bash
vdl lsp
vdl lsp --log-path
```

### How It Works

The language server speaks the [Language Server Protocol](https://microsoft.github.io/language-server-protocol/) over stdin and stdout. It provides editor features for `.vdl` files:

- Go-to-definition
- Hover information
- Find references
- Rename
- Completions
- Document symbols
- Document links

You typically do not run `vdl lsp` directly. Your editor plugin starts it and communicates with it over stdio. See the [VS Code extension](https://github.com/varavelio/vdl/blob/main/editors/vscode/README.md) or [Neovim setup](https://github.com/varavelio/vdl/blob/main/editors/neovim/README.md) for editor-specific instructions.

### Arguments

| Argument     | Required | Description                                                        |
| ------------ | -------- | ------------------------------------------------------------------ |
| `--log-path` | no       | Print the path to the LSP log file instead of starting the server. |

### Log Files

The LSP writes detailed diagnostic logs to a file under the VDL logs directory (`~/.vdl/logs/` by default). Use `--log-path` to find the log file location, then inspect it when debugging editor behavior.

```bash
vdl lsp --log-path
# /home/you/.vdl/logs/vdl-lsp.log

cat /home/you/.vdl/logs/vdl-lsp.log
```

## `vdl version`

Show VDL version information.

```bash
vdl version
vdl --version
vdl -v
```

The output includes the VDL logo, version number, and links to the repository.

## The VDL Home Directory

VDL stores caches and logs under a home directory outside your project.

### Location

By default, the VDL home directory is `~/.vdl`. You can override it with the `VDL_HOME` environment variable:

```bash
export VDL_HOME=/custom/path/.vdl
```

If no user home directory is available, VDL falls back to your system temp directory.

### Contents

| Directory | Purpose                                            |
| --------- | -------------------------------------------------- |
| `cache/`  | Cached remote plugin artifacts for reproducibility |
| `logs/`   | Log files written by the language server           |

## Environment Variables

| Variable                  | Purpose                                                                          |
| ------------------------- | -------------------------------------------------------------------------------- |
| `VDL_HOME`                | Override the VDL home directory (default: `~/.vdl`).                             |
| `VDL_INSECURE_ALLOW_HTTP` | Allow `http://` URLs for local plugin development. Set to `true`.                |
| `VDL_SKIP_HOST_HOOKS`     | Skip pre-generation and post-generation hooks. Set to `true`.                    |
| `VDL_CLOUD`               | Skip host hooks (for cloud CI environments where shell hooks are not supported). |

## Common Workflows

### Start A New Project

```bash
vdl init ./my-service
cd ./my-service
# Edit schema.vdl and vdl.config.vdl
vdl generate
```

### Format In CI

```bash
vdl format --verbose
git diff --exit-code || echo "Unformatted VDL files detected"
```

### Check Generation In CI

```bash
vdl generate --check
```

### Explore A Schema

```bash
# See how many types a schema has
vdl compile ./schema.vdl | jq '.types | length'

# List type names
vdl compile ./schema.vdl | jq '.types[].name'

# Inspect enum values
vdl compile ./schema.vdl | jq '.enums[] | {name, values: [.members[].name]}'
```

### Debug Plugin Behavior

```bash
# Compare the IR a plugin will receive
vdl compile ./schema.vdl > debug-ir.json

# Run generation in check mode
vdl generate --check
```

### Repeated Generation Workflow

```bash
# Format first, then generate
vdl format
vdl generate
```

## Troubleshooting

### `vdl generate` cannot find the config file

Make sure the file is named exactly `vdl.config.vdl` and that you are running the command from the correct directory. Pass an explicit path if needed:

```bash
vdl generate ./path/to/vdl.config.vdl
```

### `vdl generate` fails on a remote plugin

Check that the plugin version exists and that the repository is public (or configured with the correct `remotes` block). Remote plugins are cached under the VDL cache directory—clear it with:

```bash
rm -rf ~/.vdl/cache
```

### `vdl format` does not find any files

By default, `vdl format` looks for `./**/*.vdl`. If your VDL files use a different extension or are in an unexpected location, pass explicit patterns:

```bash
vdl format ./my-schemas/**/*.vdl
```

### The language server is not responding

Check the LSP logs:

```bash
vdl lsp --log-path    # find the log file
cat <log-file-path>   # inspect it
```

Make sure your editor extension is configured to launch `vdl lsp` and not a different binary.
