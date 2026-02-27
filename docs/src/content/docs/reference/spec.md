---
title: VDL Specification
description: Varavel Definition Language (VDL) Specification
---

## Overview

VDL is a schema-first IDL for defining typed contracts that drive code generation across multiple targets. A VDL document is built from a compact core of declarations and extended through annotations.

The language is designed to keep parsing stable while allowing domain semantics to evolve without changing the core grammar.

## Language Model

The parser recognizes three declaration forms at the top level: `type`, `const`, and `enum`. Includes and docstrings are also valid top-level nodes.

Domain semantics are expressed with annotations. An annotation can be attached to declarations, fields, and enum members where supported by grammar.

## Core Syntax

```vdl
include "./shared.vdl"

""" Standalone documentation """

@tag
@meta({ owner "core" })
type User {
  """ Field documentation """
  id string
  email? string
}

@config
const maxPageSize int = 100

enum Status {
  Active
  Disabled
  Archived = "archived"
}
```

## Comments

VDL supports both single-line and multi-line comments.

```vdl
// A single-line comment

/*
  A multi-line comment
  that can span multiple lines
*/

type Example {
  // Comments can appear everywhere
  value string
}
```

Comments are ignored by the parser and have no effect on schema semantics, generated code, or runtime behavior.

## Naming Conventions

VDL enforces naming conventions through formatter and tooling to keep generated code predictable.

| Element         | Convention   | Example                      |
| :-------------- | :----------- | :--------------------------- |
| Types and Enums | `PascalCase` | `UserProfile`, `OrderStatus` |
| Enum Members    | `PascalCase` | `Pending`, `InProgress`      |
| Type Members    | `camelCase`  | `userId`, `createdAt`        |
| Constants       | `camelCase`  | `maxPageSize`, `apiVersion`  |
| Annotations     | `camelCase`  | `rpc`, `internal`            |

## Includes

Includes compose schemas across files.

```vdl
// auth.vdl
type Session {
  token string
  expiresAt datetime
}

// main.vdl
include "./auth.vdl"

type AuthInfo {
  session Session
}
```

Included files are resolved by relative path from the current file. Definitions are loaded into the same compilation context, and each file is processed once.

## Declarations

### Type Declarations

`type` defines reusable structured data.

```vdl
type Product {
  id string
  title string
  price float
  tags? string[]
}
```

Type bodies accept field members (`<name>[?] <FieldType>`), spread members (`...<Reference>`), and standalone docstring members.

Type spreads reference complete type declaration names.

```vdl
type AuditMetadata {
  createdAt datetime
  updatedAt datetime
}

type Article {
  ...AuditMetadata
  title string
  content string
}
```

### Constant Declarations

`const` defines immutable values.

```vdl
const apiVersion = "1.2.0"
const maxRetries = 5
const featureEnabled = true
```

Constants support explicit type or inferred type:

```vdl
const timeoutMs int = 2500
const serviceName = "billing"
```

### Enum Declarations

`enum` defines named finite sets. Members can be implicit string values, explicit strings, or explicit integers, as long as the enum remains type-consistent.

```vdl
enum OrderStatus {
  Pending
  Processing
  Delivered
}

enum Priority {
  Low = 1
  Medium = 2
  High = 3
}
```

Enum bodies support spread entries:

```vdl
enum AllRoles {
  ...StandardRoles
  Admin = "admin"
}
```

Enum spreads reference complete enum declaration names.

Enum members also support docstrings and annotations on named members.

## Type System

### Primitive Types

| Type       | JSON Equivalent | Description                |
| :--------- | :-------------- | :------------------------- |
| `string`   | string          | UTF-8 encoded text         |
| `int`      | integer         | 64-bit signed integer      |
| `float`    | number          | 64-bit floating point      |
| `bool`     | boolean         | Logical value (true/false) |
| `datetime` | string          | ISO 8601 date-time string  |

### Arrays

Arrays are written with `[]` suffix.

```vdl
type Example {
  ids string[]
  matrix int[][]
}
```

### Maps

Maps use bracket syntax.

```vdl
type Metrics {
  counters map[int]
  usersById map[User]
}
```

### Inline Object Types

Inline objects are anonymous structured types.

```vdl
type LocationEnvelope {
  location {
    latitude float
    longitude float
  }
}
```

## Annotations

Annotations are metadata nodes with optional argument values.

```vdl
@flag
@meta({ owner "platform" tier "gold" })
type User {
  @id
  id string
}
```

Annotation syntax:

```vdl
@name
@name(<DataLiteral>)
```

Grammar supports annotation attachment on `type`, `const`, and `enum` declarations, on fields inside type members, and on named enum members.

Multiple annotations are allowed.

## Data Literals

Data literals are used by `const` values and annotation arguments.

### Scalar Literals

```vdl
"text"
42
3.14
true
false
defaultTimeout // Constant reference
Color.Red // Enum member reference
```

### Object Literals

Object entries are space-separated key/value pairs. Commas are not part of the syntax.

```vdl
const appConfig = {
  host "localhost"
  port 8080
  tls true
}
```

Object spreads are valid and reference complete constants by name.

```vdl
const prodConfig = {
  ...baseConfig
  port 443
}
```

### Array Literals

Array elements are also space-separated.

```vdl
const roles = ["admin" "editor" "viewer"]
const retryBackoffMs = [100 250 500 1000]
```

All elements in a data literal array must be of the same type.

## References

References are identifier-based values used in spreads and scalar literals.

```vdl
const defaultStatus = Status.Active
const retryLimit = maxRetries

type Session {
  ...BaseSession
}
```

A reference supports `Name` and `Name.Member` forms.

Spread members use the `Name` form. `Name.Member` is reserved for enum member references in constant values and constant data literals.

## Documentation

### Docstrings

Docstrings can be attached to declarations and members or used as standalone documentation nodes.

```vdl
"""
Represents a user in the system.
"""
type User {
  """ Unique user identifier. """
  id string
}
```

A blank line separates a standalone docstring from the next declaration in the same scope.

Type bodies support standalone docstring members. Enum bodies attach docstrings to the following named member and do not support standalone docstring-only entries.

### External Documentation Files

A docstring containing only a relative `.md` path is treated as an external documentation reference.

```vdl
""" ./docs/welcome.md """

""" ./docs/user.md """
type User {
  id string
}
```

Paths are resolved relative to the `.vdl` file containing the docstring.

## Complete Example

```vdl
include "./common/types.vdl"

""" ./docs/catalog.md """

const apiVersion = "1.0.0"
const defaultStatus = ProductStatus.Draft

enum ProductStatus {
  Draft
  Published
  Archived = "archived"
}

type Product {
  ...Entity
  id string
  name string
  price float
  status ProductStatus
  tags string[]
}

type ProductPage {
  items Product[]
  total int
}

const defaultProduct = {
  ...baseProduct
  name "Sample"
  price 9.99
  status ProductStatus.Draft
}

const sampleTags = ["featured" "popular" "seasonal"]
```

## Related Specifications

- RPC modeling and request lifecycle are documented in [`rpc.md`](./rpc.md).
- Pattern annotation rules are documented in [`patterns.md`](./patterns.md).
- Deprecation annotation rules are documented in [`deprecations.md`](./deprecations.md).
