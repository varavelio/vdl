+++
title = "Enums"
description = "Define finite sets of named values."
template = "docs.html"
weight = 7
+++

## What Enums Are

Enums define a named set of possible values.

```vdl
enum OrderStatus {
  Pending
  Paid
  Shipped
}
```

Use enums when a field should accept only one value from a known list.

```vdl
type Order {
  id string
  status OrderStatus
}
```

## Implicit String Values

When a member has no explicit value, its value is the member name.

```vdl
enum Status {
  Active
  Disabled
}
```

This is equivalent to:

```vdl
enum Status {
  Active = "Active"
  Disabled = "Disabled"
}
```

## Explicit String Values

Use explicit strings when the wire value should differ from the member name.

```vdl
enum ProductStatus {
  Draft = "draft"
  Published = "published"
  Archived = "archived"
}
```

This is common when generated code should use nice member names while serialized data uses lowercase values.

## Integer Enums

Enums can use integer values.

```vdl
enum Priority {
  Low = 1
  Medium = 2
  High = 3
}
```

For integer enums, every member must have an explicit integer value.

Invalid:

```vdl
enum Priority {
  Low = 1
  Medium
  High = 3
}
```

## Do Not Mix Value Kinds

An enum must be consistently string-based or integer-based.

Invalid:

```vdl
enum MixedStatus {
  Active = "active"
  Disabled = 2
}
```

## Unique Names And Values

Member names must be unique.

Invalid:

```vdl
enum Status {
  Active
  Active = "active"
}
```

Member values must also be unique.

Invalid:

```vdl
enum Status {
  Active = "enabled"
  Enabled = "enabled"
}
```

## Enum Member Docstrings

```vdl
enum InvoiceStatus {
  """ The invoice is editable and has not been sent. """
  Draft

  """ The invoice has been paid in full. """
  Paid
}
```

## Enum Member Annotations

```vdl
enum FeatureState {
  @deprecated("Use Enabled instead.")
  Active = "active"

  Enabled = "enabled"
}
```

Annotations on enum members are available to plugins.

## Enum Spreads

Enums can reuse members from another enum with `...Name`.

```vdl
enum StandardRole {
  Viewer = "viewer"
  Editor = "editor"
}

enum WorkspaceRole {
  ...StandardRole
  Owner = "owner"
}
```

Enum spreads must reference another enum and must use the `Name` form, not `Name.Member`.

## Naming

Enum names and member names should use `PascalCase`.

```vdl
enum PaymentStatus {
  Pending
  Completed
  Failed
}
```
