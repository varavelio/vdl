+++
title = "VPP-3: Policy"
description = "Type-safe authorization checks compiled into native boolean functions."
template = "docs.html"
path = "docs/protocols/policy"
weight = 5
+++

**VPP-3: Policy** defines a standard way to model authorization checks in VDL and compile them into native boolean functions.

The model is intentionally small. A policy is a normal VDL `type` annotated with `@policy`. The fields of that type define the data available to the authorization check, and the annotation defines the boolean expressions that allow or deny the action.

VPP-3 can model Role-Based Access Control (RBAC), Attribute-Based Access Control (ABAC), and combinations of both. In RBAC, access is decided from roles, groups, permissions, or capabilities carried by the principal or context. In ABAC, access is decided from attributes of the principal, resource, and context. Most real systems mix both styles, such as allowing Finance users to approve invoices only when MFA is verified and the invoice is not locked.

VPP-3 does not treat RBAC and ABAC as separate features. Both are expressed as type-checked boolean conditions over the policy inputs.

```vdl
@policy({
  allow "principal.role == 'Admin'"
})
type CanCreateUser {
  principal User
}
```

A VPP-3 compatible plugin validates the policy expressions against the VDL schema before generating code. The generated function returns `true` when access is allowed and `false` when access is denied.

## Why This Protocol Exists

Authorization logic is easy to scatter across handlers, services, jobs, UI flows, and framework hooks. Once that happens, it becomes difficult to audit, test, and keep consistent across languages.

VPP-3 gives VDL projects a shared, type-safe way to describe authorization decisions while still generating idiomatic code for each target language.

The goal is not to create a dynamic policy engine inside VDL. The goal is to define authorization checks in schemas, verify them before code is emitted, and generate simple functions that application code can call anywhere.

This gives policy plugins three important properties.

1. **Static safety:** invalid field references, invalid enum values, unsafe optional field access, and incompatible comparisons are rejected before code generation.
2. **Runtime safety:** generated code or adapters must fail closed when runtime values do not match the expected shape.
3. **Simple integration:** the primary generated API is a boolean function, not a policy runtime that applications must host or configure.

## Core Concepts

VPP-3 uses three policy inputs.

1. **Principal:** the identity attempting to perform something. A principal can be a user, service account, worker, API client, process, or any other authenticated identity represented by a VDL type.
2. **Resource:** the object the principal wants to access or modify. Not every policy needs a resource.
3. **Context:** transient data for the current decision, such as MFA status, tenant, client IP, time window, session state, or request metadata. Not every policy needs context.

The action being authorized is represented by the name of the `@policy` type itself. There is no special `action` variable inside policy expressions.

VPP-3 does not require a special identity field such as `id`. A policy may compare any fields declared by the principal, resource, and context types, as long as the expressions are type-safe.

## The `@policy` Annotation

VPP-3 defines one annotation: `@policy`.

The annotation is attached to a `type` declaration. That annotated type is the authorization check.

```vdl
@policy({
  allow "principal.id == resource.ownerId"
})
type CanReadInvoice {
  principal User
  resource Invoice
}
```

The `@policy` annotation receives one object argument with these fields.

1. `allow` (`string`, required): a non-empty boolean expression that can allow the check.
2. `deny` (`string`, optional): a non-empty boolean expression that can explicitly deny the check.

The `allow` expression is required. Empty strings and strings containing only whitespace are invalid. If a policy should allow everything explicitly, write `allow "true"`.

The `deny` expression is optional. If present, it must be non-empty. If a policy should deny everything explicitly, write `deny "true"`.

## Policy Type Shape

The annotated type defines the inputs available to the policy expression.

A VPP-3 policy type must use these field names.

1. `principal` is required.
2. `resource` is optional.
3. `context` is optional.

No other fields are part of VPP-3 v1.

```vdl
@policy({
  allow "principal.role == 'Admin'"
})
type CanAccessAdminPanel {
  principal User
}
```

```vdl
@policy({
  allow "principal.id == resource.ownerId"
})
type CanReadInvoice {
  principal User
  resource Invoice
}
```

```vdl
@policy({
  allow "principal.role == 'Admin' && context.mfaVerified == true"
})
type CanRunMaintenance {
  principal User
  context SessionContext
}
```

```vdl
@policy({
  allow "principal.id == resource.ownerId && context.mfaVerified == true"
})
type CanApproveInvoice {
  principal User
  resource Invoice
  context SessionContext
}
```

The `principal`, `resource`, and `context` fields must not be optional. If a policy needs a resource or context, it should declare that field as required. If it does not need one, it should omit the field entirely.

The `principal`, `resource`, and `context` fields must reference named VDL `type` declarations. They must not be scalar fields, enum fields, arrays, maps, inline objects, or references to non-object values. If a policy needs scalar input, wrap that value in a small named context or resource type.

The fields should be declared in the order `principal`, `resource`, `context` for readability. Plugins must generate function parameters in that canonical order, omitting roots that are not declared.

## Policy Naming

The annotated type name is the canonical name of the authorization check.

Policy type names should be predicate-style names and should normally start with `Can`.

```text
CanCreateUser
CanApproveInvoice
CanAccessAdminPanel
CanRunMaintenanceJob
CanPublishInvoiceEvent
```

This convention makes policy types read like boolean questions and avoids collisions with request types, event types, resource types, and command payloads.

Plugins must derive generated function names from the annotated type name. They may adapt casing or export style to the target language, but they must not add semantic prefixes or remove suffixes.

Examples:

| Policy type            | TypeScript             | Go                     | Rust                      |
| :--------------------- | :--------------------- | :--------------------- | :------------------------ |
| `CanCreateUser`        | `canCreateUser`        | `CanCreateUser`        | `can_create_user`         |
| `CanApproveInvoice`    | `canApproveInvoice`    | `CanApproveInvoice`    | `can_approve_invoice`     |
| `CanRunMaintenanceJob` | `canRunMaintenanceJob` | `CanRunMaintenanceJob` | `can_run_maintenance_job` |

Plugins must not reject a policy only because its type name does not start with `Can`. The prefix is the idiomatic VDL convention, not a validity requirement.

## Evaluation Semantics

The generated policy function returns a boolean.

1. `true` means the action is allowed.
2. `false` means the action is denied.

VPP-3 uses this evaluation order.

1. Validate runtime values when the target language or adapter cannot guarantee their shape statically.
2. If `deny` exists and evaluates to `true`, return `false`.
3. Evaluate `allow`.
4. If `allow` evaluates to `true`, return `true`.
5. Otherwise, return `false`.

This means explicit deny wins, allow enables access, and the default decision is deny.

In pseudocode:

```ts
function canApproveInvoice(principal, resource, context): boolean {
  if (!runtimeValuesAreValid(principal, resource, context)) {
    return false;
  }

  if (denyExpression(principal, resource, context) === true) {
    return false;
  }

  return allowExpression(principal, resource, context) === true;
}
```

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
}

@policy({
  allow "
    principal.id == resource.ownerId ||
    (
      has(principal.department) &&
      principal.department == 'Finance' &&
      context.mfaVerified == true
    )
  "
  deny "
    principal.disabled == true ||
    resource.status == 'Locked'
  "
})
type CanApproveInvoice {
  principal User
  resource Invoice
  context SessionContext
}
```

This policy allows a user to approve an invoice when the user owns the invoice, or when the user belongs to the Finance department and has verified MFA. It denies the check when the principal is disabled or when the invoice is locked.

Because `deny` is evaluated first, a disabled Finance user cannot approve a locked invoice even if the `allow` expression would otherwise match.

## Condition Language

`allow` and `deny` strings contain VPP-3 condition expressions.

Expressions may span multiple lines because VDL string literals can contain multiline text. Whitespace outside literals is not semantically meaningful.

The condition language is deliberately small so plugins can parse it, type-check it, and compile it to many target languages.

### Roots

Expressions may reference only the roots declared by the policy type.

1. `principal`: available in every valid policy.
2. `resource`: available only when the policy type declares a `resource` field.
3. `context`: available only when the policy type declares a `context` field.

Field access uses dot notation.

```text
principal.id
resource.ownerId
context.mfaVerified
```

Field paths may traverse object fields declared in VDL. They must not use dynamic property access, array indexing, map indexing, method calls, implicit database lookups, or implicit entity lookups.

### Literals

Expressions support these literals.

1. Booleans: `true`, `false`.
2. Strings: single-quoted values such as `'Finance'`.
3. Numbers: `100`, `10.5`.
4. Arrays: comma-separated literal arrays such as `['admin', 'owner']`.

String literals use single quotes so expressions can be embedded naturally inside VDL double-quoted strings. Plugins must support escaping `\'` and `\\` inside expression string literals.

When a plugin receives an `allow` or `deny` expression, VDL has provided the outer double-quoted string content; the VPP-3 expression parser must still interpret escapes inside inner expression string literals. For example, `'Owner\'s Team'` represents `Owner's Team`, and a literal backslash should be written as `\\`.

### Enum Values

Enum fields are compared by value using string literals.

```vdl
enum InvoiceStatus {
  Locked
  Unlocked
}
```

```text
resource.status == 'Locked'
```

When an enum member has an explicit value, the string literal must match that value. When an enum member does not have an explicit value, the string literal must match the member name.

A plugin must validate enum string literals at generation time. If `resource.status` is `InvoiceStatus`, then `resource.status == 'Archived'` is invalid unless `Archived` is a valid value of that enum.

Plugins may generate target-native enum comparisons even though the policy expression uses a string literal.

### Operators

VPP-3 condition expressions support these operators.

1. Logical operators: `&&`, `||`, `!`.
2. Comparison operators: `==`, `!=`, `>`, `>=`, `<`, `<=`.
3. Membership operator: `in`.
4. Grouping: parentheses.

Examples:

```text
principal.id == resource.ownerId
principal.department == 'Finance' && context.mfaVerified == true
principal.role in ['admin', 'owner']
principal.id in resource.allowedUserIds
!(resource.status == 'Locked')
```

The `in` operator requires the right side to be an array expression or an array-typed field. The element type must be compatible with the left side.

Operator precedence is defined as follows.

1. `!`
2. `==`, `!=`, `>`, `>=`, `<`, `<=`, `in`
3. `&&`
4. `||`

Parentheses override precedence.

### Unsupported Expressions

VPP-3 v1 intentionally does not support arbitrary expression features.

Plugins must reject function calls other than `has(...)`, arithmetic, regex matching, string interpolation, array indexing, map indexing, method calls, async operations, imports, dynamic property access, external lookups, and custom user-defined functions.

These restrictions keep policies portable and easy to compile into native code.

## Optional Fields and `has`

Optional VDL fields require explicit presence checks.

The `has(path)` function returns `true` when an optional field is present at runtime.

```text
has(principal.department)
```

A plugin must reject unsafe optional field access.

If `department` is optional, this expression is valid:

```text
has(principal.department) && principal.department == 'Finance'
```

This expression is invalid:

```text
principal.department == 'Finance'
```

`has(...)` only guards access in the expression branch where it is known to be true. For example, in `A && B`, a positive `has(path)` proven in `A` may guard access to `path` in `B`. In `A || B`, each branch must be safe on its own because either branch may evaluate independently.

If a plugin cannot prove optional field access is safe, it must reject the policy before generating code.

## Static Validation

Static validation is the core safety guarantee of VPP-3.

Before generating files, a compatible plugin must parse and validate every `allow` and `deny` expression against the VDL IR.

The plugin must reject the policy if any of these checks fail.

1. `@policy` is attached to something other than a `type` declaration.
2. The policy type does not declare a required `principal` field.
3. `principal`, `resource`, or `context` is marked optional.
4. `principal`, `resource`, or `context` does not reference a named VDL object type.
5. The policy type declares fields other than `principal`, `resource`, and `context`.
6. `allow` is missing, empty, or not a string.
7. `deny` is present but empty or not a string.
8. An expression references a root that the policy type does not declare.
9. A referenced field does not exist.
10. A referenced enum value does not exist.
11. A comparison uses incompatible types.
12. An `in` expression compares incompatible element types.
13. A condition does not evaluate to boolean.
14. An optional field is accessed without a safe `has(...)` guard.
15. An unsupported operator, function, literal, or expression form is used.

If validation fails, the plugin must stop generation and return diagnostics tied to the source `@policy` annotation or policy type. When possible, diagnostics should point to the expression and subexpression that failed validation.

## Runtime Validation and Fail-Closed Behavior

Static validation proves that the policy definition is valid. It does not always prove that runtime values are trustworthy.

Generated code or generated adapters must validate runtime values when the target language, boundary, or integration receives untrusted data. This is especially important for dynamic languages, HTTP handlers, message consumers, plugin boundaries, or any place where values may come from outside the target language type system.

Runtime validation must fail closed.

If `principal`, `resource`, or `context` does not match the expected VDL shape at runtime, the generated authorization function must not return `true`. It should return `false` or use a target-idiomatic checked wrapper that reports the validation error while preserving a deny decision.

## Generated Function Shape

The primary generated API is a pure boolean function.

The generated function name is derived from the annotated policy type name, adapted to the target language naming style.

The parameter order is always `principal`, `resource`, then `context`, omitting roots that are not declared.

Examples:

```vdl
@policy({ allow "principal.role == 'Admin'" })
type CanAccessAdminPanel {
  principal User
}
```

```ts
function canAccessAdminPanel(principal: User): boolean
```

```vdl
@policy({ allow "principal.id == resource.ownerId" })
type CanReadInvoice {
  principal User
  resource Invoice
}
```

```ts
function canReadInvoice(principal: User, resource: Invoice): boolean
```

```vdl
@policy({ allow "principal.role == 'Admin' && context.mfaVerified == true" })
type CanRunMaintenance {
  principal User
  context SessionContext
}
```

```ts
function canRunMaintenance(principal: User, context: SessionContext): boolean
```

```vdl
@policy({ allow "principal.id == resource.ownerId && context.mfaVerified == true" })
type CanApproveInvoice {
  principal User
  resource Invoice
  context SessionContext
}
```

```ts
function canApproveInvoice(
  principal: User,
  resource: Invoice,
  context: SessionContext,
): boolean
```

Plugins must detect generated name collisions and report diagnostics instead of silently overwriting generated symbols.

## Generated Code Requirements

Generated policy code must be deterministic, pure, and local.

The generated decision must not require network calls, database access, local configuration files, runtime expression parsing, or a dynamic policy store.

Plugins may generate helper types, checked wrappers, test utilities, explanation helpers, or diagnostics helpers. Those additions must not change the canonical boolean decision semantics.

Plugins should use VPP-0 generated model types or compatible target-native model types when available.

## Integration With Other Protocols

VPP-3 is transport agnostic.

A generated policy function can protect RPC handlers, event consumers, CLI actions, background jobs, UI actions, feature flags, maintenance tasks, or in-memory workflows. Other protocols may call VPP-3 generated functions, but VPP-3 does not depend on those protocols.

For example, a VPP-2 RPC handler named `createUser` can call a generated `canCreateUser(...)` policy function, but the policy itself remains independent from RPC.

Plugins should preserve cross-protocol metadata when useful. For example, if a policy type is marked as deprecated through the global deprecation standard, generated policy documentation or code comments should surface that lifecycle information where the target supports it.

## Required Behavior for Compatible Plugins

A plugin may call itself VPP-3 compatible only if it follows these rules.

1. It must recognize `@policy` on `type` declarations.
2. It must require a non-optional `principal` field.
3. It must support optional `resource` and `context` fields, and require them to be non-optional when present.
4. It must require `principal`, `resource`, and `context` to reference named VDL object types.
5. It must reject fields other than `principal`, `resource`, and `context` in policy types.
6. It must require non-empty `allow` and support optional non-empty `deny`.
7. It must parse and type-check the condition language before generating code.
8. It must enforce root availability based on the policy type fields.
9. It must validate enum string literals against enum values.
10. It must enforce safe optional field access through `has(...)`.
11. It must evaluate `deny` before `allow`.
12. It must return `false` when `deny` evaluates to `true`.
13. It must return `true` only when `allow` evaluates to `true` and no `deny` matched.
14. It must return `false` when no rule allows the check.
15. It must generate a boolean authorization function derived from the policy type name.
16. It must fail closed when runtime values are invalid.
17. It must not require a runtime policy interpreter to evaluate the canonical decision.

## Recommended Output Behavior

VPP-3 compatible plugins should generate target-idiomatic authorization APIs while preserving the same model.

Useful outputs may include policy functions, modules, test helpers, generated documentation, decision explanation helpers, mock fixtures, or typed adapters for framework integration.

Plugins should make policy logic easy to test. A generated function should be callable directly with principal, resource, and context values without needing to boot an application server.

## Non-Goals

VPP-3 intentionally does not define every part of an authorization system.

This protocol does not define authentication, identity providers, session management, database loading, resource fetching, dynamic policy stores, admin UIs, audit log formats, network transport, RPC integration, event integration, role storage, relationship graph traversal, or application-specific business workflows.

Those concerns may be handled by other protocols, plugin configuration, generated adapters, infrastructure, or application code.

## Compatibility Statement

VPP-3 is the standard type-safe policy contract for VDL plugins.

A plugin may call itself **VPP-3 compatible** only if it preserves the `@policy` annotation model, policy type shape, naming rules, condition language, static validation rules, fail-closed runtime behavior, generated boolean API, and evaluation semantics defined in this document.
