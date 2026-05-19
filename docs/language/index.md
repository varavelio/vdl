---
title: Language Guide
description: Learn how to write VDL schemas from first principles.
---

## What You Will Learn

This guide teaches the VDL language step by step. It is written for readers who are new to VDL and want to understand how to write schemas confidently.

VDL is small on purpose. A schema is made from a few building blocks:

- files and comments
- `include` statements
- docstrings
- `type` declarations
- field types
- `enum` declarations
- `const` declarations
- annotations
- spreads and references

Once you understand these pieces, you can model data contracts, RPC services, events, configuration files, JSON schemas, generated code packages, and custom internal contracts.

## A Tiny Schema

```vdl
"""
Represents a user account.
"""
type User {
  id string
  email string
  displayName? string
  createdAt datetime
}

enum UserStatus {
  Active = "active"
  Suspended = "suspended"
}

const defaultPageSize = 25
```

This file defines three things:

- `User`, an object type with fields
- `UserStatus`, a finite set of allowed values
- `defaultPageSize`, a reusable constant value

## How To Read VDL Syntax

VDL intentionally avoids punctuation that is not needed.

Fields do not use colons:

```vdl
type Product {
  id string
  price float
}
```

Object literal entries do not use colons or commas:

```vdl
const serverConfig = {
  host "localhost"
  port 8080
  tls false
}
```

Array literal items are separated by whitespace:

```vdl
const roles = ["admin" "editor" "viewer"]
```

If you come from JSON, TypeScript, Go, or YAML, this is the main habit to learn: VDL is whitespace-friendly and declaration-oriented.

## Top-Level Declarations

At the top level, a `.vdl` file can contain:

```vdl
include "./shared.vdl"

""" Standalone documentation. """

type User {
  id string
}

enum Status {
  Active
}

const version = "1.0.0"
```

Only `type`, `enum`, and `const` create named symbols. Includes and standalone docstrings help compose and document the schema.

## Recommended Learning Path

Read these pages in order:

1. [Files and Comments](./files.md)
2. [Includes](./includes.md)
3. [Docstrings](./docstrings.md)
4. [Types](./types.md)
5. [Field Types](./field-types.md)
6. [Enums](./enums.md)
7. [Constants and Literals](./constants.md)
8. [Annotations](./annotations.md)
9. [Spreads and References](./spreads-references.md)
10. [Naming and Validation](./naming-validation.md)

For exact grammar-level detail, see the [VDL specification](../reference/spec.md).
