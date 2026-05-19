<p align="center">
  <img
    src="https://cdn.jsdelivr.net/gh/varavelio/vdl@9cb843/assets/png/vdl.png"
    alt="VDL logo"
    width="150"
  />
</p>

<h1 align="center">Varavel Definition Language</h1>

<p align="center">
  VDL is an open-source, type-safe, multi-language, and easily extensible schema definition language and code generation toolchain.
</p>

<p align="center">
  <a href="https://github.com/varavelio/vdl/actions">
    <img src="https://github.com/varavelio/vdl/actions/workflows/ci.yaml/badge.svg" alt="CI status"/>
  </a>
  <a href="https://github.com/varavelio/vdl/releases/latest">
    <img src="https://img.shields.io/github/release/varavelio/vdl.svg" alt="Release Version"/>
  </a>
  <a href="https://github.com/varavelio/vdl">
    <img src="https://img.shields.io/github/stars/varavelio/vdl?style=flat&label=github+stars" alt="GitHub Stars"/>
  </a>
  <a href="https://github.com/varavelio/vdl/blob/main/LICENSE">
    <img src="https://img.shields.io/github/license/varavelio/vdl.svg" alt="License"/>
  </a>
</p>

<p align="center">
  <a href="https://varavel.com">
    <img src="https://cdn.jsdelivr.net/gh/varavelio/brand@1.0.1/dist/badges/project.svg" alt="A Varavel project"/>
  </a>
  <a href="https://vdl.varavel.com">
    <img src="https://cdn.jsdelivr.net/gh/varavelio/brand@1.0.1/dist/badges/documentation.svg" alt="VDL Documentation"/>
  </a>
</p>

Easily define typed data models, constants, enums, documentation, RPC APIs, and domain-specific contracts once in `.vdl` files, then use plugins to generate type-safe code and the supporting artifacts your stack needs.

The core language focuses on structured data. APIs, RPC services, events, schemas, documentation, and custom internal contracts are modeled through annotations and generated through plugins.

Learn more at [https://vdl.varavel.com](https://vdl.varavel.com).

## Installation

Select the method that best fits your workflow.

| Platform          | Method     | Command                                                        |
| ----------------- | ---------- | -------------------------------------------------------------- |
| **Linux / macOS** | Shell      | <code>curl -fsSL https://get.varavel.com/vdl &#124; sh</code>  |
| **Linux / macOS** | Homebrew   | `brew install varavelio/tap/vdl`                               |
| **Windows**       | PowerShell | <code>irm https://get.varavel.com/vdl.ps1 &#124; iex</code>    |
| **Any**           | NPM        | `npm install -D @varavel/vdl`                                  |
| **Any**           | Manual     | [Download binaries](https://github.com/varavelio/vdl/releases) |

For version pinning, experimental releases, npm usage, and manual installs, see the [complete installation guide](https://vdl.varavel.com/guides/installation/).

## What is VDL?

VDL is a schema-first language for describing data contracts in a compact, readable format. The core language is intentionally small: `include`, documentation blocks, `type`, `enum`, `const`, annotations, arrays, maps, inline objects, literals, and spreads.

Domain semantics are added with annotations and plugins. For example, `@rpc` can describe RPC services, `@event` can describe event routing contracts, and a custom annotation can be interpreted by your own plugin.

```vdl
include "./shared.vdl"

""" A customer account shared across systems. """
type Account {
  id string
  email string
  createdAt datetime
  tags? string[]
}

enum AccountStatus {
  Active = "active"
  Suspended = "suspended"
}

const apiVersion = "2026-05"
```

The compiler resolves includes, validates semantics, expands spreads, resolves constants and annotations, and produces a deterministic Intermediate Representation (IR). Plugins consume that IR and generate code, schemas, documentation, catalogs, or any other text artifact.

## Core Capabilities

- **General data modeling:** define reusable types, aliases, inline objects, arrays, maps, optional fields, enums, and constants.
- **Documentation next to contracts:** attach Markdown docstrings to declarations, fields, enum members, and top-level docs.
- **Annotations as extension points:** keep the language small while allowing plugins to interpret domain-specific metadata such as RPC operations, events, deprecations, IDs, ownership, or framework hints.
- **Project-aware analysis:** resolve `include` graphs, validate references, catch duplicate declarations, detect invalid spreads and cycles, and report diagnostics with source positions.
- **Stable IR for generators:** plugins receive a flattened, resolved, sorted schema representation instead of raw source text.
- **Developer tooling:** use the CLI formatter, JSON IR compiler, and LSP for editor diagnostics, completion, hover, definitions, references, rename, document symbols, formatting, and document links.
- **Plugin-first generation:** official and custom JavaScript plugins generate Go, TypeScript, JSON Schema, OpenAPI, event catalogs, schema explorers, metadata exports, and more.

## Quick Start

Create a project:

```bash
# This creates a `schema.vdl` and `vdl.config.vdl` in the current directory.
vdl init
```

Generate outputs with the plugins configured in `vdl.config.vdl`:

```bash
vdl generate
```

Format VDL files:

```bash
vdl format
```

Compile a schema to IR JSON:

```bash
vdl compile ./schema.vdl
```

Start the language server:

```bash
vdl lsp
```

## Plugin System

VDL generation is handled by plugins. A plugin is just a JavaScript file that exports a `generate(input)` function. VDL passes `{ version, ir, options }` into the function and writes the returned virtual files into the configured output directory.

Example `vdl.config.vdl`:

```vdl
const config = {
  version 1
  plugins [
    {
      src "varavelio/vdl-plugin-go@v0.1.3"
      schema "./schema.vdl"
      outDir "./gen/go"
      options {
        package "contracts"
      }
    }
    {
      src "varavelio/vdl-plugin-json-schema@v0.1.0"
      schema "./schema.vdl"
      outDir "./gen/json-schema"
      options {
        root "Account"
      }
    }
  ]
}
```

Plugin sources can be local `.js` files, HTTPS `.js` URLs, or GitHub shorthand references like `owner/vdl-plugin-name@v0.1.0`. Remote plugins are cached and recorded in `vdl.lock` for reproducibility.

Useful docs:

- [Available plugins](https://vdl.varavel.com/guides/plugins/)
- [Creating plugins](https://vdl.varavel.com/guides/creating-plugins/)
- [Language guide](https://vdl.varavel.com/language/)
- [Project configuration](https://vdl.varavel.com/guides/configuration/)
- [VDL language specification](https://vdl.varavel.com/reference/spec/)
- [RPC annotation model](https://vdl.varavel.com/reference/rpc/)
- [Event annotation model](https://vdl.varavel.com/reference/events/)

## Who is VDL for?

VDL is useful when you want one contract to drive many outputs:

- shared backend/frontend data models
- generated Go or TypeScript contract packages
- JSON Schema for validators, forms, or external integrations
- OpenAPI documents for annotation-based RPC services
- event subject builders and event catalogs
- static schema explorers and metadata exports
- custom internal generators for your own frameworks and conventions

## License

VDL is **100% open source** and released under MIT license.

### Disclaimer

VDL is provided "AS IS" without warranty of any kind. The authors assume no responsibility for damages or losses from using this software. You are responsible for testing generated code before production use.

See [LICENSE](LICENSE) for details.

---

_VDL is a labor of love by its contributors. We provide it freely in the hope it will be useful, but without any guarantees. Your success is your own responsibility, and your contributions back to the project are always welcome!_
