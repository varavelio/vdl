+++
title = "VPP-1: Deprecations"
description = "Global deprecation metadata for every VDL plugin and protocol."
template = "docs.html"
path = "docs/protocols/deprecations"
weight = 3
+++

**VPP-1: Deprecations** defines the standard meaning of the `@deprecated` annotation across the entire VDL ecosystem.

This protocol is different from protocols that describe a family of generators. VPP-1 does not exist so people can build "deprecation plugins". It exists so every plugin, protocol, SDK, code generator, documentation generator, editor tool, and internal workflow can understand deprecation metadata in the same way.

If a plugin claims compatibility with VPP-1, it must preserve the semantics defined in this document.

## Why This Protocol Exists

Deprecation is a universal concern.

Schemas evolve. Fields get replaced. Enum members stop being recommended. Constants become legacy. Protocol-specific operations may remain supported while users migrate to newer shapes. Generated APIs often need to keep old symbols available without encouraging new code to use them.

Without a shared standard, every plugin could invent its own way to mark deprecation. One plugin might use `@deprecated`, another might use `@legacy`, another might require structured metadata, and another might ignore the concept entirely. That would make VDL projects harder to understand and harder to compose.

VPP-1 keeps the meaning simple and global.

When users write `@deprecated`, every VPP-1 compatible plugin should understand the same intent: this symbol is still part of the schema, but it should be treated as deprecated wherever the generated target can express that idea.

## The `@deprecated` Annotation

Deprecation is expressed with the `@deprecated` annotation.

It can be used as a flag.

```vdl
@deprecated
type LegacyUser {
  id string
}
```

It can also receive one optional string argument.

```vdl
@deprecated("Use UserProfile instead.")
type UserSummary {
  id string
  displayName string
}
```

The presence of `@deprecated` is what marks the symbol as deprecated. The optional string is only additional human-readable guidance.

## Accepted Forms

VPP-1 compatible plugins must recognize these forms.

```vdl
@deprecated
```

```vdl
@deprecated("Use another symbol instead.")
```

The first form means the symbol is deprecated with no message.

The second form means the symbol is deprecated with a message.

## Message Semantics

The optional string argument is a human-facing note.

It can contain a reason, migration hint, replacement suggestion, timeline note, warning, or any other text that helps a person understand what to do next.

```vdl
@deprecated("Use primaryEmail instead.")
email string
```

VPP-1 does not assign machine-readable meaning to the message. Plugins should not parse the message to infer replacement symbols, dates, severity levels, owners, or migration rules.

Future protocols may define structured deprecation metadata, but VPP-1 only standardizes the flag and the optional string message.

## Invalid or Future Arguments

If `@deprecated` receives an argument that is not a string, the symbol is still deprecated because the annotation is present.

However, the non-string argument must not be treated as the deprecation message.

```vdl
@deprecated({ reason "Use AccountV2" })
type AccountV1 {
  id string
}
```

In this example, `AccountV1` is deprecated, but the object argument is ignored for VPP-1 message purposes.

Plugins may emit a warning for unsupported deprecation arguments when that fits their diagnostic model, but they must not reinterpret a non-string value as the VPP-1 message.

This rule keeps current behavior predictable while leaving room for the ecosystem to extend deprecation metadata in a later protocol revision.

## Allowed Targets

The `@deprecated` annotation can be used wherever VDL annotations are valid and where a consuming plugin can preserve or surface the meaning.

Common targets include these constructs.

1. Top-level `type` declarations.
2. Top-level `enum` declarations.
3. Top-level `const` declarations.
4. Type fields.
5. Named enum members.
6. Any future or protocol-specific construct that supports annotations through normal VDL metadata.

VPP-1 is intentionally global. It is not limited to any specific protocol family, target language, transport, runtime, or generator category.

## Schema Semantics

Deprecation is metadata.

It does not make a declaration invalid. It does not remove a symbol from the schema. It does not require a plugin to stop generating that symbol. It does not change the type, value, validation rules, or reference behavior of the deprecated construct.

A deprecated symbol remains part of the VDL program unless another rule, plugin, or configuration explicitly removes or rejects it.

The purpose of `@deprecated` is to communicate lifecycle status, not to change schema correctness.

## Required Behavior for Compatible Plugins

A plugin may call itself VPP-1 compatible only if it follows these rules.

1. It must recognize `@deprecated` as a deprecation marker.
2. It must recognize `@deprecated("message")` as a deprecation marker with a human-readable message.
3. It must treat the annotation as the source of deprecation, regardless of whether a valid message is present.
4. It must ignore non-string arguments as VPP-1 messages.
5. It must not make a schema invalid only because a symbol is deprecated.
6. It must preserve, expose, or reflect deprecation metadata whenever doing so is reasonable for its output or tool behavior.

The exact output is target-specific. A CLI tool, language generator, editor extension, documentation generator, runtime adapter, and internal company plugin may all expose deprecation differently. The semantic interpretation must remain the same.

## Recommended Output Behavior

Plugins should surface deprecation information in the most idiomatic way available to their target.

Generated source code may use language-level deprecation comments, attributes, decorators, annotations, doc comments, metadata tables, generated catalogs, or framework-specific markers.

Generated documentation should make deprecated symbols visible and should show the optional message when one exists.

Editor tooling may show hover metadata, completion labels, diagnostics, warnings, or visual markers for deprecated symbols.

Runtime or framework integrations may forward deprecation metadata into generated registries, manifests, schemas, or descriptors when those artifacts have a natural place for it.

Plugins should avoid hiding deprecated symbols by default. Deprecation tells users to migrate away from a symbol, it does not mean the symbol has already disappeared.

## Examples

Deprecating a type:

```vdl
@deprecated("Use AccountV2 instead.")
type Account {
  id string
}
```

Deprecating a field:

```vdl
type User {
  id string

  @deprecated("Use displayName instead.")
  name string
}
```

Deprecating an enum member:

```vdl
enum Status {
  Active

  @deprecated("Use Disabled instead.")
  Inactive
}
```

Deprecating a constant:

```vdl
@deprecated
const legacyTimeoutMs = 12000
```

Deprecating a protocol-specific construct:

```vdl
@rpc
type ServiceContract {
  @proc
  @deprecated("Use createUserV2 instead.")
  createUser {
    input {
      email string
    }
    output {
      id string
    }
  }
}
```

The final example intentionally avoids relying on one specific protocol. The rule is the same anywhere annotations are valid: if a VPP-1 compatible plugin understands the annotated construct, it must preserve the deprecation meaning.

## Compatibility Statement

VPP-1 is intentionally small because it is meant to be universal.

The standard exists so every part of the VDL ecosystem can share one deprecation vocabulary. A deprecated model, field, enum member, constant, generated API, documented symbol, or protocol-specific construct should communicate the same lifecycle status everywhere.

A plugin may call itself **VPP-1 compatible** only if it respects that shared meaning.
