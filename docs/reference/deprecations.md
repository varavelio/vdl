---
title: Deprecation Annotations
description: Marking declarations and members as deprecated with @deprecated
---

## Overview

Deprecation in VDL is expressed with the `@deprecated` annotation.

The annotation can be used as a flag or with a message argument to provide migration guidance.

## Syntax

```vdl
@deprecated
type LegacyUser {
  id string
}

@deprecated("Use userProfile")
const userSummaryType = "legacy"
```

## Allowed Targets

Deprecation can be attached anywhere annotations are valid in grammar.

- Top-level declarations: `type`, `const`, `enum`
- Type fields
- Named enum members

## Examples

```vdl
@deprecated("Use AccountV2")
type Account {
  id string

  @deprecated("Use primaryEmail")
  email string
}

enum Status {
  Active

  @deprecated("Use Disabled")
  Inactive
}

@deprecated
const legacyTimeoutMs = 12000
```

## Tooling Expectations

Generators and tooling can surface deprecations in generated code, diagnostics, and documentation output.

Deprecation metadata is informational by default and does not remove symbols from schema validity on its own.
