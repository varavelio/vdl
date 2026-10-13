---
title: "VPP-3: Policy"
description: "Type-safe authorization checks compiled into native boolean functions."
path: "docs/protocols/policy"
weight: 5
---

**VPP-3: Policy** is the standard way to model authorization in VDL.

A policy is a normal VDL `type` annotated with `@policy`. The type declares the inputs. The annotation declares two boolean expressions: `allow` and `deny`. A compatible plugin checks them before generating code and emits a pure boolean function.

```vdl
@policy({
  allow "principal.role == 'Admin'"
})
type CanCreateUser {
  principal User
}
```

VPP-3 covers Role-Based Access Control (RBAC), Attribute-Based Access Control (ABAC), and mixes of both. It is a starting base, not a full authorization platform. Complex graph traversal, database lookups, and external queries stay in application code.

## Why It Exists

Authorization spreads across handlers, jobs, UI flows, and framework hooks. VPP-3 keeps the common decision in one typed place, verified once, generated per language.

1. **Static safety:** bad fields, bad enums, bad types, and unsafe optional access fail before code is emitted.
2. **Runtime safety:** bad runtime shapes fail closed to `false`.
3. **Simple integration:** the output is a function, not a runtime to host.

## Core Concepts

A policy has up to three inputs.

1. **Principal:** who acts. User, service account, worker, API client. Always required.
2. **Resource:** what is accessed. Optional. Omit it when not needed.
3. **Context:** transient decision data. MFA status, tenant, time, session. Optional.

The action is the policy type name itself. There is no `action` variable inside expressions.

```vdl
@policy({ allow "principal.id == resource.ownerId" })
type CanReadInvoice {
  principal User
  resource Invoice
}
```

## The `@policy` Annotation

`@policy` goes on a `type` declaration. It takes one object with:

1. `allow` (`string`, required): non-empty boolean expression.
2. `deny` (`string`, optional): non-empty boolean expression.
3. `version` (`int`, optional): language version. Defaults to `1`. Unknown versions must be rejected.

Use `allow "true"` to allow everything explicitly. Use `deny "true"` to deny everything explicitly. Empty or whitespace-only strings are invalid.

```vdl
@policy({
  version 1
  allow "principal.role == 'Admin' && context.mfaVerified == true"
  deny "principal.disabled == true"
})
type CanRunMaintenance {
  principal User
  context SessionContext
}
```

## Policy Type Shape

Only these fields are part of VPP-3 v1:

1. `principal`: required.
2. `resource`: optional.
3. `context`: optional.

Rules:

- All three must be required when present. Never `principal?`.
- All three must reference named VDL object types. No scalars, enums, arrays, maps, or inline objects. Wrap scalar input in a small type.
- Declare them in `principal`, `resource`, `context` order. Generators emit parameters in that order, omitting missing roots.

## Policy Naming

Name policies like boolean questions, normally starting with `Can`:

```text
CanCreateUser
CanApproveInvoice
CanAccessAdminPanel
```

Generators derive function names from the type name with target casing (`canCreateUser`, `CanCreateUser`, `can_create_user`). The `Can` prefix is convention, not validity. Never reject a policy only for missing it. Report collisions instead of overwriting.

## Evaluation Semantics

The generated function returns `true` for allow, `false` for deny.

1. Validate runtime shapes. Mismatch returns `false`.
2. If `deny` exists and is `true`, return `false`.
3. Return `allow`.
4. Default is `false`.

Deny wins. No rule means deny. `&&` and `||` short-circuit left to right.

## Policy Language

VPP-3 expressions use a small language called the **Policy Language**. It lives inside `allow` and `deny` strings. It has no statements, no assignments, and no variables in v1.

### Keywords

```text
true false in has principal resource context
```

Everything else is an identifier, literal, operator, or punctuation. Future versions only add keywords additively. Unknown syntax must be rejected.

### Lexical Rules

- Whitespace outside literals is ignored, including newlines.
- Identifiers: `[A-Za-z_][A-Za-z0-9_]*`.
- Strings: single quotes only, such as `'Finance'`. Support only `\'` and `\\` escapes.
- Numbers: `100`, `10.5`. No hex, no exponents in v1.
- Booleans: `true`, `false`.
- Arrays: comma-separated, such as `['admin', 'owner']`. Unlike VDL data literals, which separate items with whitespace, Policy Language arrays require commas because items are full expressions and whitespace is insignificant.

### Syntax

```ebnf
expr    := orExpr ;
orExpr  := andExpr ( "||" andExpr )* ;
andExpr := cmpExpr ( "&&" cmpExpr )* ;
cmpExpr := unary ( ( "==" | "!=" | ">" | ">=" | "<" | "<=" | "in" ) unary )? ;
unary   := "!" unary | primary ;
primary := hasCall | paren | ref | lit | array ;
hasCall := "has" "(" ref ")" ;
paren   := "(" expr ")" ;
ref     := root ( "." ident )+ ;
root    := "principal" | "resource" | "context" ;
lit     := string | number | "true" | "false" ;
array   := "[" [ expr ( "," expr )* ] "]" ;
```

Precedence, highest first: `!`, then `== != > >= < <= in`, then `&&`, then `||`. Parentheses override.

### Roots and Paths

An expression may use only roots declared by its policy type. `resource` and `context` are unavailable when omitted.

```text
principal.id
resource.ownerId
context.mfaVerified
```

Paths walk named object fields with dot notation. No array indexing, no map indexing, no method calls, no dynamic access, no database lookups.

### Types

No implicit conversion. `1 == '1'` is invalid. `int` versus `float` is invalid in v1. Every root expression must evaluate to `bool`.

- `== !=`: `string`, `int`, `float`, `bool`, `datetime`, `enum`.
- `> >= < <=`: `int`, `float`, `datetime`, `string`. Strings use UTF-8 byte order.
- `in`: scalar left, array right with compatible element type. Empty array is `false`.
- `!`: `bool` only.
- Objects and arrays are never directly comparable except through `in`.

### Arrays and Objects

Arrays hold scalars only: string, number, boolean, enum value, or reference to one of those. All items must share one element type. Nested arrays and objects inside arrays are invalid in v1.

```text
principal.role in ['admin', 'owner']
principal.id in resource.allowedIds
```

There are no object literals in v1. Compare objects field by field instead of comparing them directly:

```text
principal.tenantId == resource.tenantId
```

VDL uses whitespace in data arrays such as `["admin" "editor"]`. The Policy Language uses commas because array items are expressions, not plain values. Generators always emit commas for target languages.

### Enum Values

Compare enums by value with a string literal:

```vdl
enum InvoiceStatus {
  Locked
  Unlocked
}
```

```text
resource.status == 'Locked'
```

The literal must match the member value, or the member name when no explicit value exists. `resource.status == 'Archived'` is invalid without that member. Generators may emit native enum comparisons.

### Datetime

Compare `datetime` fields against `datetime` fields only. There are no datetime literals in v1. Pass time through `context`, such as `context.now`:

```text
resource.expiresAt > context.now
```

No `now()` calls, no duration arithmetic. Compute windows before calling the function.

### Optional Fields and `has`

`has(ref)` is `true` when the full chain is present. `has(principal.profile.department)` means `profile` is present and `department` is present. Generators expand the chain per target (`!= nil`, `!== undefined`, `is_some()`).

One `has` guards the whole path. Never require `has(p) && has(p.x) && has(p.x.y)` by hand.

Equality is null-safe to keep the common case short:

- Missing `== value` is `false`.
- Missing `!= value` is `true`.

The generator still emits the presence check. You just do not write it.

Ordering and membership are strict and require an explicit guard:

```text
has(principal.department) && principal.department == 'Finance'
has(resource.score) && resource.score > 100
has(principal.id) && principal.id in resource.allowedIds
```

Guard scope follows short-circuit flow. A `has` on the left of `&&` guards the right side. Each `||` branch must be safe alone. `!has(p.x)` does not guard `p.x`.

Presence-only checks need no other condition:

```text
has(principal.department)
```

### Unsupported in v1

Reject arithmetic, regex, string interpolation, indexing, method calls, async, imports, custom functions, and any function other than `has(...)`. This keeps every target emitter to a direct mapping with explicit parentheses.

## Static Validation

Validate before emitting. Reject the policy when any check fails:

1. `@policy` is not on a `type`.
2. `principal` is missing or optional.
3. `resource` or `context` is optional when present.
4. Any root is not a named object type.
5. Extra fields beyond `principal`, `resource`, `context`.
6. `allow` missing, empty, or not a string.
7. `deny` present but empty or not a string.
8. Unknown `version`.
9. Unknown root, field, or enum value.
10. Incompatible comparison or `in` element mismatch.
11. Non-`bool` root result.
12. Unsafe optional access.
13. Unsupported syntax.

Stop generation and point diagnostics at the annotation, expression, and subexpression when possible.

## Runtime Behavior

Static checks do not make runtime data trusted. Validate shapes at runtime when data crosses a dynamic boundary such as HTTP, queues, or plugin input. Mismatch returns `false`, or a checked wrapper reports the error while keeping deny.

## Generated Code

Emit one pure boolean function per policy. Deterministic, local, no network, no database, no config files, no expression parsing at runtime.

```ts
function canReadInvoice(principal: User, resource: Invoice): boolean
```

Use VPP-0 models when available. Helpers, test fixtures, and explanation utilities are allowed when they preserve the canonical decision.

## Complete Example

```vdl
type User {
  id string
  name string
  role string
  department? string
  disabled bool
}

enum InvoiceStatus {
  Locked
  Unlocked
}

type Invoice {
  ownerId string
  status InvoiceStatus
}

type SessionContext {
  mfaVerified bool
  now datetime
}

@policy({
  version 1
  allow "principal.id == resource.ownerId || has(principal.department) && principal.department == 'Finance' && context.mfaVerified == true"
  deny "principal.disabled == true || resource.status == 'Locked'"
})
type CanApproveInvoice {
  principal User
  resource Invoice
  context SessionContext
}
```

Finance users with MFA pass. Owners pass. Disabled users and locked invoices fail because `deny` runs first.

## Integration

VPP-3 is transport agnostic. RPC handlers, event consumers, CLI actions, jobs, and UI actions can call the same function. Other protocols may use VPP-3 functions, but VPP-3 never depends on them. Preserve `VPP-1` deprecation metadata where the target supports it.

## Non-Goals

VPP-3 does not define authentication, identity providers, sessions, database loading, dynamic stores, admin UIs, audit formats, transport, role storage, graph traversal, or business workflows.

## Compatibility Statement

A plugin is **VPP-3 compatible** only when it preserves the annotation model, type shape, naming rules, Policy Language, validation, fail-closed behavior, boolean API, and deny-first semantics defined here.
