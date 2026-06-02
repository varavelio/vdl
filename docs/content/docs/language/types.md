+++
title = "Types"
description = "Define object types, aliases, arrays, maps, and inline objects."
template = "docs.html"
weight = 5
+++

## What Types Are

`type` declarations define reusable data shapes.

The most common form is an object type:

```vdl
type User {
  id string
  email string
}
```

But a type can also be an alias for another type expression:

```vdl
type UserId string
type Tags string[]
type Metadata map[string]
```

## Object Types

Object types use braces and contain fields.

```vdl
type Product {
  id string
  name string
  price float
  active bool
}
```

Each field is written as:

```text
fieldName TypeName
```

There are no colons and no commas.

## Optional Fields

Add `?` after the field name to make a field optional.

```vdl
type UserProfile {
  userId string
  displayName? string
  avatarUrl? string
}
```

Optional means the field may be absent in generated representations, depending on the target plugin.

## Type Aliases

A type can name any valid field type.

```vdl
type UserId string
type UserIds string[]
type UserMetadata map[string]
```

Aliases are useful when you want semantic names without repeating raw primitives everywhere.

```vdl
type EmailAddress string

type User {
  id string
  email EmailAddress
}
```

## Object Types Can Reference Other Types

```vdl
type Address {
  line1 string
  line2? string
  city string
  country string
}

type Customer {
  id string
  billingAddress Address
  shippingAddress? Address
}
```

References must point to declared types or enums.

## Inline Objects

Use an inline object when a nested shape is only useful in one place.

```vdl
type CheckoutSession {
  id string
  payment {
    provider string
    amountCents int
    currency string
  }
}
```

Inline objects can contain fields, nested inline objects, and spreads.

## Empty Objects

An object type can be empty, but it is rarely useful.

```vdl
type EmptyObject {
}
```

Prefer adding fields when the contract has meaningful data.

## Field Names Must Be Unique

Within an object or inline object, each field name must be unique.

Good:

```vdl
type Product {
  id string
  sku string
}
```

Invalid:

```vdl
type Product {
  id string
  id int
}
```

## Naming

Type names should use `PascalCase`:

```vdl
type UserProfile {
  displayName string
}
```

Field names should use `camelCase`:

```vdl
type UserProfile {
  displayName string
  createdAt datetime
}
```

The analyzer reports diagnostics when names do not follow these conventions.
