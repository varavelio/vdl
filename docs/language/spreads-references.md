---
title: Spreads and References
description: Reuse fields, enum members, and literal values safely.
---

## References

A reference points to a named VDL symbol.

References appear in several places:

- field types
- spreads
- constant values
- annotation arguments

## Type References

Fields can reference declared types.

```vdl
type UserId string

type User {
  id UserId
}
```

Fields can also reference enums.

```vdl
enum Status {
  Active
  Disabled
}

type User {
  status Status
}
```

## Constant References

Constants can reference other constants by name.

```vdl
const defaultLimit = 100
const pageSize = defaultLimit
```

## Enum Member References

Use `EnumName.MemberName` to reference an enum member in literal values.

```vdl
enum Status {
  Active = "active"
  Disabled = "disabled"
}

const defaultStatus = Status.Active
```

## Type Spreads

Type spreads copy fields from another object type.

```vdl
type AuditFields {
  createdAt datetime
  updatedAt datetime
}

type Article {
  ...AuditFields
  id string
  title string
}
```

After analysis, plugins see `Article` as if it had all fields directly.

Only object types can be spread into object types.

Invalid:

```vdl
type UserId string

type User {
  ...UserId
  name string
}
```

## Inline Object Spreads

Inline objects can also spread object types.

```vdl
type Money {
  amount int
  currency string
}

type Order {
  id string
  payment {
    ...Money
    provider string
  }
}
```

## Field Conflicts

Spreads cannot introduce a field that already exists in the same object.

Invalid:

```vdl
type BaseUser {
  id string
}

type User {
  ...BaseUser
  id string
}
```

Rename one field or avoid the spread.

## Enum Spreads

Enum spreads copy members from another enum.

```vdl
enum StandardRole {
  Viewer = "viewer"
  Editor = "editor"
}

enum ProjectRole {
  ...StandardRole
  Owner = "owner"
}
```

Enum spreads must reference enums, not enum members.

Invalid:

```vdl
enum Role {
  Viewer = "viewer"
}

enum ProjectRole {
  ...Role.Viewer
  Owner = "owner"
}
```

## Object Literal Spreads

Object literals can spread constant objects.

```vdl
const baseConfig = {
  retries 3
  timeoutMs 5000
}

const productionConfig = {
  ...baseConfig
  timeoutMs 10000
}
```

Only object constants can be used this way.

## Spread Cycles

Spreads cannot form cycles.

Invalid:

```vdl
type A {
  ...B
  a string
}

type B {
  ...A
  b string
}
```

Keep spread hierarchies simple and acyclic.

## Name Forms

VDL supports two reference forms:

| Form          | Used for                                      |
| ------------- | --------------------------------------------- |
| `Name`        | Types, constants, spreads                     |
| `Name.Member` | Enum member references in literal expressions |

Spreads must use `Name`, not `Name.Member`.
