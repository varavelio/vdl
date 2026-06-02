+++
title = "Constants and Literals"
description = "Define reusable dynamic literal values with const declarations."
template = "docs.html"
weight = 8
+++

## What Constants Are

Constants define reusable literal values.

```vdl
const apiVersion = "1.0.0"
const maxRetries = 3
const featureEnabled = true
```

Constants do not have explicit type annotations. VDL infers their value kind from the literal.

Correct:

```vdl
const timeoutMs = 2500
```

Invalid:

```text
const timeoutMs int = 2500
```

## Scalar Literals

VDL supports these scalar literal forms:

```vdl
const textValue = "hello"
const intValue = 42
const floatValue = 3.14
const trueValue = true
const falseValue = false
```

Strings use double quotes.

## Object Literals

Object literals use braces and space-separated key/value entries.

```vdl
const serverConfig = {
  host "localhost"
  port 8080
  tls false
}
```

There are no colons and no commas.

Object keys must be unique inside the same object.

## Nested Object Literals

```vdl
const appConfig = {
  server {
    host "localhost"
    port 8080
  }
  logging {
    level "info"
  }
}
```

## Array Literals

Array items are separated by whitespace.

```vdl
const roles = ["admin" "editor" "viewer"]
const retryBackoffMs = [100 250 500]
```

All items in an array literal must have the same value kind.

Invalid:

```vdl
const mixedValues = ["admin" 1 true]
```

## Constant References

Constants can reference other constants.

```vdl
const defaultTimeoutMs = 5000
const requestTimeoutMs = defaultTimeoutMs
```

References use the constant name directly.

## Enum Member References

Constants can reference enum members with `EnumName.MemberName`.

```vdl
enum ProductStatus {
  Draft = "draft"
  Published = "published"
}

const defaultStatus = ProductStatus.Draft
```

The referenced enum and member must exist.

## Object Spreads

Object literals can spread another constant object.

```vdl
const baseConfig = {
  host "localhost"
  port 8080
}

const productionConfig = {
  ...baseConfig
  host "api.example.com"
}
```

Only constants whose value is an object literal can be used in object spreads.

Spreads use the `Name` form. `Name.Member` is not valid for object spreads.

## Constants In Annotations

Because annotation arguments are data literals, constants can also help share annotation values.

```vdl
const owner = "platform"

@meta({ team owner })
type ServiceConfig {
  name string
}
```

Whether a plugin understands a particular annotation is plugin-specific.

## Naming

Constant names should use `camelCase`.

```vdl
const apiVersion = "1.0.0"
const maxPageSize = 100
```
