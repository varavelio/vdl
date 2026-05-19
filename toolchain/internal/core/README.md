# Compiler Core

This directory contains the **core language logic** for VDL. It is responsible for reading, parsing, validating, and transforming `.vdl` source files into a format ready for code generation.

The core is designed as a strict pipeline. No external tools (LSP, CLI, Generators) should interact with the raw source code, they must consume the artifacts produced here.

## Architecture & Data Flow

The compilation process moves linearly through these four packages:

```mermaid
flowchart LR
  VFS[("VFS (Disk/RAM)")] --> Parser
  Parser --> AST
  AST --> Analysis
  Analysis -- "Resolve Imports" --> VFS
  Analysis --> IR
  IR --> Generators["Plugin Runtime / Generators"]
```

## Package Overview

### `vfs` (Virtual File System)

**"The I/O Abstraction Layer."**

- **Responsibility:** Provides a unified, thread-safe way to read source files. It acts as the bridge between the Operating System and the Compiler.

**Key Features:**

- **Memory Caching:** Caches disk reads to avoid repeated syscalls during analysis.
- **Dirty Buffers (LSP):** Supports in-memory file overlays (`WriteFileCache`), allowing the compiler to analyze unsaved changes in the editor before they exist on disk.
- **Cache Invalidation:** Supports removing cached files (`RemoveFileCache`) when files are saved or closed in the editor.
- **Path Resolution:** Handles canonical absolute paths (`Resolve`), ensuring consistent file identity across different operating systems.

### `ast` (Abstract Syntax Tree)

**"How the code is written."**

- **Responsibility:** Defines the data structures that represent the raw syntactic structure of a `.vdl` file.
- **Contents:** Node definitions for the current declaration model, including `Schema`, `Include`, `Docstring`, `TypeDecl`, `EnumDecl`, `ConstDecl`, annotations, fields, spreads, references, and data literals.

**Characteristics:**

- **Dirty:** Contains syntax noise like quotes (`"string"`), positions, and comments.
- **Hierarchical:** Represents the exact nesting of the source file.
- **Passive:** Just data structs. No logic.
- **Position Tracking:** Every node embeds `Positions` (start/end line and column) for LSP features and error reporting.

**Primitive Types:** `string`, `int`, `float`, `bool`, `datetime`.

### `parser`

**"From Text to Tree."**

- **Responsibility:** Converts raw bytes (source code) into an `ast.Schema`.
- **Technology:** Built on the [participle](https://github.com/alecthomas/participle) parser-combinator library.

**Contents:**

- **Lexer:** Tokenizes input using regex-based rules. Keywords have priority over identifiers.
- **Parser:** Validates **syntax** only (e.g., matching braces, valid keywords). Uses lookahead to distinguish attached from standalone docstrings.

- **Input:** `.vdl` file content (bytes).
- **Output:** `*ast.Schema` or Syntax Errors (`parser.Error`).

### `analysis`

**"The Semantic Brain."**

- **Responsibility:** Understands and validates the **meaning** of the code across the entire project. This is the engine that powers the **LSP**.

**Contents:**

- **Resolver:** Parses the entry point, recursively resolves `include` statements, detects circular dependencies, and inlines external markdown docstrings.
- **Symbol Table:** Collects and registers all types, enums, constants, and standalone documentation blocks into a global namespace.
- **Validators:** Modular validation passes for:
  - Naming conventions (PascalCase and camelCase)
  - Type references (undefined types, maps, arrays, and inline objects)
  - Enum consistency (mixed types, duplicate values)
  - Constant literals and references
  - Field and object literal uniqueness
  - Spread validity for types and enums
  - Cycle detection (circular type dependencies)
  - Global uniqueness (cross-category name collisions)
  - File naming conventions
- **Diagnostics:** Errors are reported as `Diagnostic` with file, position, error code (e.g., `E001`), and message.

**Key Design:** Uses a **best-effort strategy** — always returns a `Program` that is as complete as possible, even when errors are found. This enables LSP features (hover, go-to-definition, completions) to remain functional in files with errors.

- **Input:** Entry point file path.
- **Output:** `*analysis.Program` (A validated graph of symbols and files) + `[]Diagnostic`.

### `ir` (Intermediate Representation)

**"The Blueprint for Generation."**

- **Responsibility:** Transforms the complex analysis graph into a clean, flat model optimized for code generators.

**Design Principles:**

- **Resolved Shape:** Generators consume normalized type references and resolved literal values rather than raw AST nodes.
- **Aggressive Flattening:** Type and enum spreads are expanded so generators see final field/member lists.
- **Deterministic Order:** Top-level collections are sorted alphabetically for reproducible output.
- **Source Positions:** IR keeps source positions so plugins can report useful diagnostics.
- **Serializable:** Designed for JSON export, useful for `vdl compile`, plugins, and tests.

**Contents:**

- **Flattening:** Spreads are resolved, fields are copied into the final struct.
- **Doc Normalization:** Docstrings are trimmed and dedented for consistent formatting.

- **Input:** `*analysis.Program`.
- **Output:** `*irtypes.IrSchema` (the generator-facing model passed to plugins and emitted by `vdl compile`).

---

## Quick Reference

| Feature       | `ast`                     | `analysis`          | `ir`                      |
| ------------- | ------------------------- | ------------------- | ------------------------- |
| **Scope**     | Single File               | Entire Project      | Optimized for Output      |
| **Includes**  | Raw Strings (`"./a.vdl"`) | Resolved Graph      | Merged/Invisible          |
| **Spreads**   | Reference (`...Base`)     | Resolved Pointer    | Copied Fields (Flattened) |
| **Docs**      | Raw (`""" ./doc.md """`)  | Validated/Inlined   | Normalized Content        |
| **Positions** | Full (line, column)       | Full (for LSP)      | Preserved for diagnostics |
| **Errors**    | Syntax Errors             | Diagnostics (codes) | None (assumes valid)      |
| **Used By**   | Parser, Formatter         | LSP, IR Builder     | Plugin runtime/generators |
