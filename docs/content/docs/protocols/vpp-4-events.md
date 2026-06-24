+++
title = "VPP-4: Events"
description = "Transport-agnostic event payloads and routing subjects with the @event annotation."
template = "docs.html"
path = "docs/protocols/events"
weight = 6
+++

**VPP-4: Events** defines a standard way to model asynchronous event payloads and routing subjects in VDL.

The model is intentionally transport agnostic. An event can be published through NATS, RabbitMQ, Kafka, HTTP webhooks, a database outbox, an internal queue, or any other event delivery system. VPP-4 only defines the schema contract: the payload type, the routing subject template, and the rules that connect them.

```vdl
@event("auth.userCreated.{userId}")
type UserCreatedEvent {
  userId string
  email string
  timestamp datetime
}
```

A VPP-4 compatible plugin turns this declaration into a typed payload representation and subject-building utilities for the target language.

## Why This Protocol Exists

Event-driven systems often depend on routing strings: subjects, topics, routing keys, webhook paths, stream names, or other delivery identifiers. Those strings are easy to mistype and easy to separate from the payload data they describe.

VPP-4 keeps the routing subject and payload schema together.

If a subject contains a dynamic placeholder such as `{userId}`, the payload must contain a matching field named `userId`. This makes events self-contained: when an event is logged, persisted, replayed, forwarded, or delivered through a transport without rich routing metadata, the payload still contains the values needed to identify and route it.

## The `@event` Annotation

VPP-4 defines one annotation: `@event`.

The annotation is attached to a `type` declaration and receives one string argument: the routing subject template.

```vdl
@event("orders.statusChanged.{orderId}")
type OrderStatusChanged {
  orderId string
  status string
}
```

The annotated type is the event payload. Its fields define the data that publishers emit and subscribers consume.

The subject template is a literal string. Static segments are copied as-is. Dynamic segments are written as placeholders in `{fieldName}` form.

## Subject Templates

A subject template is the routing string attached to an event declaration.

```text
auth.userCreated.{userId}
```

The template may contain zero or more placeholders. Each placeholder references a field in the annotated payload type.

Examples:

```vdl
@event("audit.login")
type UserLoginAudit {
  userId string
  at datetime
}

@event("orders.{orderId}.statusChanged")
type OrderStatusChanged {
  orderId string
  status string
}

@event("tenant.{tenantId}.users.{userId}.created")
type TenantUserCreated {
  tenantId string
  userId string
  email string
}
```

A template with no placeholders is valid. In that case, the subject builder takes no placeholder arguments and returns the static subject string.

## Placeholder Rules

Placeholders make routing dynamic, so they must be statically validated.

A VPP-4 compatible plugin must enforce these rules.

1. Every placeholder must exactly match a field declared on the annotated event type.
2. Placeholder fields must be top-level fields of the payload type.
3. Placeholder fields must be primitive scalar fields.
4. Nested placeholder paths are not allowed.
5. Object, array, map, and other non-primitive placeholder fields are not allowed.
6. The same placeholder may appear multiple times in the subject template.

Valid example:

```vdl
@event("orders.statusChanged.{orderId}")
type OrderStatusChanged {
  orderId string
  status string
}
```

Invalid example, because `region` is not declared in the payload:

```vdl
@event("logistics.{region}.delivery_failed")
type DeliveryFailed {
  deliveryId string
}
```

Valid example, because repeated placeholders reuse the same field value:

```vdl
@event("audit.{tenantId}.users.{tenantId}.created")
type UserAuditCreated {
  tenantId string
}
```

Invalid example, because nested placeholder paths are not allowed:

```vdl
@event("accounts.{user.id}.created")
type AccountCreated {
  user {
    id string
  }
}
```

The plugin must reject invalid placeholders before generating code.

## Payload Semantics

The annotated type is the event payload.

VPP-4 does not define a separate envelope. The event payload is exactly the data structure represented by the annotated VDL type.

Because placeholder values must also exist in the payload, events remain self-contained. Consumers can inspect the payload without relying on transport-specific routing metadata being preserved forever.

VPP-4 does not require every payload field to appear in the subject. Only fields needed for routing or partitioning should be included in the template.

## Subject Builder Semantics

Plugins must generate a way to build the final subject string for each event.

A subject builder always returns a `string`.

If a placeholder field is not a string, the subject builder must convert that primitive value to a string when composing the final subject.

When a placeholder appears more than once, the builder should accept that field once and reuse the same value for every occurrence. For deterministic APIs, placeholder parameters should follow the order of first appearance in the template.

Example template:

```text
tenant.{tenantId}.users.{userId}.created
```

Conceptual generated builder:

```ts
function buildTenantUserCreatedSubject(tenantId: string, userId: string): string {
  return `tenant.${tenantId}.users.${userId}.created`;
}
```

## Generated API Shape

Generators transform an `@event` declaration into two primary artifacts.

1. A payload type for the target language.
2. Subject formatting utilities for the event subject template.

Depending on the target, plugins may also emit catalogs, registries, documentation, publish helpers, subscribe helpers, test fixtures, or broker-specific adapters.

### TypeScript Example

```ts
export interface UserCreatedEvent {
  userId: string;
  email: string;
  timestamp: string;
}

export function buildUserCreatedEventSubject(userId: string): string {
  return `auth.userCreated.${userId}`;
}
```

### Go Example

```go
type UserCreatedEvent struct {
    UserId    string    `json:"userId"`
    Email     string    `json:"email"`
    Timestamp time.Time `json:"timestamp"`
}

func BuildUserCreatedEventSubject(userId string) string {
    return "auth.userCreated." + userId
}
```

These examples are illustrative. Each plugin should generate code that is idiomatic for its target while preserving the same payload and subject semantics.

## Event Catalogs

Plugins should aggregate event declarations into a centralized event catalog when that is useful for the target ecosystem.

Catalogs help applications inspect available events, route incoming messages, generate documentation, configure subscribers, or connect generated code to framework-specific registries.

If a plugin emits event catalog metadata, the literal subject template must be exposed under the `Subject` key.

Conceptual catalog shape:

```go
type EventMetadata struct {
    Name    string // e.g. "UserCreatedEvent"
    Subject string // e.g. "auth.userCreated.{userId}"
}
```

The `Subject` value is the template from the VDL annotation, not a formatted runtime subject.

## Required Behavior for Compatible Plugins

A plugin may call itself VPP-4 compatible only if it follows these rules.

1. It must recognize `@event` on `type` declarations.
2. It must require the `@event` argument to be a string subject template.
3. It must treat the annotated type as the event payload.
4. It must validate every placeholder against the payload type.
5. It must reject placeholders that reference missing fields.
6. It must reject nested placeholder paths.
7. It must reject non-primitive placeholder fields.
8. It must allow repeated placeholders and reuse the same field value.
9. It must generate or expose subject formatting behavior that returns a string.
10. It must preserve the literal subject template for generated metadata and catalogs.

## Recommended Output Behavior

VPP-4 compatible plugins should generate target-idiomatic event APIs while preserving the same event model.

Useful outputs may include payload models, subject builders, event catalogs, publish helpers, subscriber helpers, broker adapters, generated documentation, mock event fixtures, or routing metadata.

Plugins should preserve relevant cross-protocol metadata when useful. For example, if an event payload type is marked as deprecated through the global deprecation standard, generated payload types, docs, and catalogs should surface that lifecycle information where the target supports it.

## Non-Goals

VPP-4 intentionally does not define every part of an event platform.

This protocol does not define broker selection, delivery guarantees, retry policies, dead-letter handling, consumer groups, partitions, transactions, schemas for broker configuration, authentication, authorization, persistence, serialization formats beyond the generated payload model, or runtime publish/subscribe behavior.

Those concerns may be handled by other protocols, other plugins, generated adapters, infrastructure, or application code.

## Compatibility Statement

VPP-4 is the standard event contract for VDL plugins.

A plugin may call itself **VPP-4 compatible** only if it preserves the `@event` annotation model, payload semantics, subject template rules, placeholder validation, subject builder behavior, and catalog metadata requirements defined in this document.
