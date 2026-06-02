+++
title = "Start Here"
description = "Learn how to write VDL schemas from first principles."
template = "docs.html"
weight = 1
+++

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

1. [Files and Comments](/docs/language/files/)
2. [Includes](/docs/language/includes/)
3. [Docstrings](/docs/language/docstrings/)
4. [Types](/docs/language/types/)
5. [Field Types](/docs/language/field-types/)
6. [Enums](/docs/language/enums/)
7. [Constants and Literals](/docs/language/constants/)
8. [Annotations](/docs/language/annotations/)
9. [Spreads and References](/docs/language/spreads-references/)
10. [Naming and Validation](/docs/language/naming-validation/)

For exact grammar-level detail, see the [VDL specification](/docs/reference/spec/).
