---
title: Files and Comments
description: How VDL files are structured and how comments work.
---

## File Extension

VDL schemas are written in files ending with `.vdl`.

```text
schema.vdl
auth.vdl
events.vdl
```

Regular schema files should use lowercase names with numbers or underscores only:

```text
user.vdl
order_events.vdl
api_v1.vdl
```

The special file name `vdl.config.vdl` is reserved for project generation configuration.

## File Contents

A VDL file can contain:

- `include` statements
- standalone docstrings
- `type` declarations
- `enum` declarations
- `const` declarations

Example:

```vdl
include "./shared.vdl"

""" Documentation for this schema. """

type User {
  id string
}

enum UserStatus {
  Active
  Suspended
}

const apiVersion = "1.0.0"
```

## Single-Line Comments

Use `//` for a comment that runs to the end of the line.

```vdl
// User-facing account data.
type Account {
  id string // Stable account identifier.
  email string
}
```

Comments are ignored by the compiler. They are useful for readers, but they do not appear in the generated IR as documentation.

Use docstrings when you want documentation to be part of the schema model.

## Block Comments

Use `/* ... */` for multi-line comments.

```vdl
/*
  This schema is shared by the billing and support systems.
  Keep field names stable because generated clients depend on them.
*/
type Customer {
  id string
  email string
}
```

Block comments can appear anywhere whitespace can appear.

## Comments vs Docstrings

Use comments for notes to humans reading the source file:

```vdl
// Internal note: this name must match the external system.
type ExternalAccount {
  id string
}
```

Use docstrings for documentation that plugins should see:

```vdl
"""
Public account information returned to clients.
"""
type PublicAccount {
  id string
  displayName string
}
```

Plugins receive docstrings through the VDL IR. Comments are discarded.

## Whitespace

VDL is flexible with whitespace. These are equivalent:

```vdl
type User { id string email string }
```

```vdl
type User {
  id string
  email string
}
```

Prefer the second style. It is easier to read and matches the formatter.
