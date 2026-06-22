+++
title = "Field Types"
description = "Learn primitives, arrays, maps, custom types, enums, and inline object field types."
template = "docs.html"
weight = 6
+++

## Field Type Overview

Every field has a type.

```vdl
type Example {
  name string
  count int
}
```

VDL supports:

- primitive types
- custom types
- enums
- arrays
- maps
- inline objects

## Primitive Types

| Type       | JSON Equivalent | Description                |
| ---------- | --------------- | -------------------------- |
| `string`   | string          | UTF-8 encoded text         |
| `int`      | integer         | 64-bit signed integer      |
| `float`    | number          | 64-bit floating point      |
| `bool`     | boolean         | Logical value (true/false) |
| `datetime` | string          | RFC 3339 date-time string  |

> **Note:** `datetime` values must be RFC 3339 formatted strings. RFC 3339 is a strict subset of ISO 8601 chosen to ensure consistent parsing across implementations.

```vdl
type PrimitiveExample {
  name string
  age int
  score float
  active bool
  createdAt datetime
}
```

## Custom Types

Fields can reference other declared types.

```vdl
type Address {
  city string
  country string
}

type User {
  id string
  address Address
}
```

## Enum Fields

Fields can reference enums.

```vdl
enum OrderStatus {
  Pending
  Paid
  Shipped
}

type Order {
  id string
  status OrderStatus
}
```

## Arrays

Arrays use `[]` after the element type.

```vdl
type Collection {
  tags string[]
  scores int[]
  orders Order[]
}

type Order {
  id string
}
```

Multidimensional arrays repeat the suffix:

```vdl
type MatrixData {
  values float[][]
}
```

## Maps

Maps use `map[T]`, where `T` is the value type.

```vdl
type Metrics {
  counters map[int]
  labels map[string]
}
```

Maps can contain custom types:

```vdl
type User {
  id string
  email string
}

type UserDirectory {
  usersById map[User]
}
```

Maps can also contain arrays or nested maps:

```vdl
type ComplexMapExample {
  tagsByGroup map[string[]]
  matrixByName map[float[][]]
}
```

## Inline Objects

Inline objects are anonymous structured field types.

```vdl
type SearchResult {
  id string
  metadata {
    score float
    source string
  }
}
```

Use inline objects when the nested shape does not need a reusable name.

Use a named type when the shape is reused or important enough to document on its own.

## Optional Field Types

Optionality belongs to the field name, not the type.

```vdl
type User {
  displayName? string
  tags? string[]
  profile? {
    bio? string
  }
}
```

The `?` always appears after the field name.

## Recursive Types

VDL can model recursive structures when a cycle is broken by an optional field.

```vdl
type TreeNode {
  id string
  children? TreeNode[]
}
```

Direct required cycles are rejected because they cannot be represented as finite data.

Invalid:

```vdl
type A {
  b B
}

type B {
  a A
}
```

Make one side optional to model a nullable or absent link.
