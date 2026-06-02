+++
title = "Includes"
description = "Split VDL schemas across files with include statements."
template = "docs.html"
weight = 3
+++

## Why Includes Exist

As schemas grow, it is useful to split them into focused files.

For example:

```text
schema.vdl
shared.vdl
events.vdl
```

Use `include` to bring definitions from another `.vdl` file into the same compilation context.

## Basic Syntax

```vdl
include "./shared.vdl"

type UserSession {
  token string
  account Account
}
```

The included file can define types, enums, constants, or documentation used by the current file.

Example `shared.vdl`:

```vdl
type Account {
  id string
  email string
}
```

Example `schema.vdl`:

```vdl
include "./shared.vdl"

type Session {
  account Account
  token string
}
```

`Session.account` can use `Account` because `schema.vdl` includes `shared.vdl`.

## Relative Paths

Include paths are resolved relative to the file that contains the `include` statement.

```vdl
include "../shared/types.vdl"
include "./events/user_events.vdl"
```

This makes each file self-contained: moving the entry point does not change how nested includes are resolved.

## Include Graphs

Includes are recursive. If `schema.vdl` includes `orders.vdl`, and `orders.vdl` includes `money.vdl`, all three files are analyzed together.

```vdl
include "./orders.vdl"

type Checkout {
  order Order
}
```

Each file is processed once. Circular includes are reported as diagnostics.

## File Name Rules

Regular included schema files should have names matching this pattern:

```text
[a-z0-9_]+.vdl
```

Good names:

```text
shared.vdl
order_events.vdl
api_v1.vdl
```

Avoid names like:

```text
Shared.vdl
order-events.vdl
schema.backup.vdl
```

## Config Files Are Not Included

`vdl.config.vdl` is the project generation configuration file. It cannot be included as a normal schema file.

Keep language schemas and generation configuration separate:

```text
schema.vdl        # Your contracts
vdl.config.vdl    # Plugin generation configuration
```

## Practical Advice

- Put shared primitives, aliases, and base objects in files like `shared.vdl`.
- Put event contracts in files like `events.vdl` or `user_events.vdl`.
- Keep the main entry file small and readable.
- Prefer shallow include graphs when possible.
