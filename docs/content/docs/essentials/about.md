+++
title = "About VDL"
description = "Learn what VDL is, what it provides, and why the core language stays small."
template = "docs.html"
weight = 1
+++

## A Definition Language For Data Contracts

VDL is a schema-first language and toolchain for defining structured data contracts once and generating useful artifacts from them.

The core language models data, documentation, constants, enums, annotations, and composition. API, RPC, event, documentation, schema-generation, and custom internal workflows are built on top through plugins.

## What VDL Provides

VDL is built around three parts.

1. **A compact definition language:** `.vdl` files describe typed data with `type`, `enum`, `const`, `include`, annotations, arrays, maps, inline objects, literals, documentation blocks, and spreads.
2. **Developer tooling:** the CLI formats schemas, compiles them to JSON IR, runs plugin generation, and exposes an LSP for editor diagnostics, completion, hover, definitions, references, rename, document symbols, formatting, and links.
3. **A plugin-first generator runtime:** plugins receive the resolved VDL IR and generate the target artifacts your project needs.

## Why The Core Is Small

The language keeps domain semantics out of the grammar. Instead, annotations describe intent and plugins decide what to do with that intent.

For example:

- `@rpc` can mark a type as an RPC service.
- `@proc` and `@stream` can mark RPC operations.
- `@event("subject.{id}")` can mark event payload contracts.
- `@deprecated` can be interpreted by a JSON Schema plugin.
- Your own annotations can drive your own internal generators.

This lets VDL support many domains without turning the parser into a collection of framework-specific features.

## Common Use Cases

- Generate shared Go and TypeScript data models.
- Export JSON Schema for validation, forms, docs, or external integrations.
- Generate RPC clients, servers, and OpenAPI documents from annotation-based service contracts.
- Generate event subject builders and event catalogs.
- Publish static schema explorers for internal or external documentation.
- Export VDL IR as JSON for runtime metadata and custom tooling.
- Build private plugins that encode your team's framework conventions.

## Project Philosophy

- **Schema as source of truth:** contracts should be defined once and reused everywhere.
- **Small language, extensible semantics:** core syntax stays stable while plugins evolve domain behavior.
- **Readable contracts:** schemas should be understandable during code review without generator knowledge.
- **Strong tooling:** diagnostics, formatting, LSP features, and deterministic IR are core to the workflow.
- **Practical generation:** generated files are regular text files that can be inspected, tested, committed, and integrated into existing stacks.
