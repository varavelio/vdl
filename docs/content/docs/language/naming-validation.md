+++
title = "Naming and Validation"
description = "Learn the naming conventions and validation rules VDL applies."
template = "docs.html"
weight = 11
+++

## Why Naming Matters

VDL schemas are used to generate code in different languages. Consistent names make generated output predictable and pleasant to use.

The analyzer reports diagnostics when names do not follow the expected convention.

## Naming Conventions

| Element      | Convention   | Examples                     |
| ------------ | ------------ | ---------------------------- |
| Types        | `PascalCase` | `User`, `OrderItem`          |
| Enums        | `PascalCase` | `OrderStatus`, `Region`      |
| Enum members | `PascalCase` | `Pending`, `InProgress`      |
| Fields       | `camelCase`  | `userId`, `createdAt`        |
| Constants    | `camelCase`  | `apiVersion`, `maxPageSize`  |
| Annotations  | `camelCase`  | `deprecated`, `generateCode` |

## PascalCase

PascalCase starts with an uppercase letter and does not use underscores.

Good:

```vdl
type UserProfile {
  displayName string
}

enum OrderStatus {
  Pending
}
```

Avoid:

```vdl
type user_profile {
  displayName string
}
```

## camelCase

camelCase starts with a lowercase letter and does not use underscores.

Good:

```vdl
type UserProfile {
  displayName string
  createdAt datetime
}

const maxPageSize = 100
```

Avoid:

```vdl
type UserProfile {
  display_name string
}

const MaxPageSize = 100
```

## File Names

Regular schema files should match:

```text
[a-z0-9_]+.vdl
```

Examples:

```text
schema.vdl
shared_types.vdl
events_v1.vdl
```

The configuration file must be named:

```text
vdl.config.vdl
```

## Global Name Uniqueness

Top-level type, enum, and constant names share one global namespace.

Invalid:

```vdl
type Status string

enum Status {
  Active
}
```

Use unique names across declarations.

## Duplicate Fields

Fields must be unique within each object or inline object.

Invalid:

```vdl
type User {
  id string
  id int
}
```

## Undefined References

Every type reference must point to a declared type, enum, or primitive.

Invalid:

```vdl
type User {
  address Address
}
```

Fix it by declaring `Address` or including the file that declares it.

```vdl
type Address {
  city string
}

type User {
  address Address
}
```

## Cycles

VDL rejects required type cycles.

Invalid:

```vdl
type Parent {
  child Child
}

type Child {
  parent Parent
}
```

Break the cycle with an optional field.

```vdl
type Parent {
  child Child
}

type Child {
  parent? Parent
}
```

## Practical Checklist

Before generating code, check that:

- file names are lowercase with underscores when needed
- types and enums use PascalCase
- fields and constants use camelCase
- all referenced types and enums exist
- object fields are unique
- enum member names and values are unique
- required type references do not form cycles
- spreads do not conflict or form cycles
