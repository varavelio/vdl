---
title: Annotations
description: Attach metadata to VDL declarations, fields, and enum members.
---

## What Annotations Are

Annotations attach metadata to schema nodes.

```vdl
@deprecated
type OldUser {
  id string
}
```

The VDL core does not hardcode most domain behavior. Plugins read annotations and decide what they mean.

For example, one plugin may interpret `@event`, another may interpret `@rpc`, and another may interpret `@deprecated`.

## Basic Syntax

An annotation starts with `@` and a name.

```vdl
@internal
type InternalUser {
  id string
}
```

Annotations can also have one argument.

```vdl
@deprecated("Use NewUser instead.")
type OldUser {
  id string
}
```

The argument must be a data literal: string, number, boolean, object, array, constant reference, or enum member reference.

## Multiple Annotations

You can stack annotations.

```vdl
@internal
@owner("platform")
type InternalConfig {
  name string
}
```

Annotations attach to the declaration or member that immediately follows them.

## Declaration Annotations

Annotations can be attached to `type`, `enum`, and `const` declarations.

```vdl
@event("users.created.{userId}")
type UserCreated {
  userId string
  email string
}

@stable
enum Region {
  Europe = "eu"
  America = "us"
}

@config
const serviceName = "billing"
```

## Field Annotations

Fields can have annotations.

```vdl
type User {
  @id
  id string

  @deprecated("Use displayName instead.")
  name? string

  displayName string
}
```

## Enum Member Annotations

Enum members can have annotations.

```vdl
enum Plan {
  Free = "free"

  @deprecated("Use Team instead.")
  Startup = "startup"

  Team = "team"
}
```

## Object Arguments

Object arguments are useful for structured metadata.

```vdl
@meta({ owner "platform" tier "gold" })
type Account {
  id string
}
```

Object entries use VDL literal syntax: key followed by value, no colon, no comma.

## Array Arguments

```vdl
@tags(["public" "billing" "stable"])
type Invoice {
  id string
}
```

Array literal items must be the same kind.

## Annotation Names

Annotation names should use `camelCase`.

```vdl
@generateClient
type PublicApi {
  name string
}
```

Avoid underscores or PascalCase annotation names.

## Annotations Are Plugin Contracts

An annotation has no effect unless some tool or plugin understands it.

This is a feature. It lets teams create domain-specific contracts without changing the VDL grammar.

Examples:

- `@rpc` can mark a type as an RPC service.
- `@proc` can mark a request/response RPC operation.
- `@stream` can mark a server-sent stream operation.
- `@event("subject.{field}")` can mark an event payload and routing subject.
- `@deprecated` can mark generated code or schema output as deprecated.

Always check the plugin documentation for the annotations it supports.
