---
title: Event Specification
description: Modeling asynchronous events, routing subjects, and strict payloads with the @event annotation
---

## Overview

Event-driven architectures in VDL are modeled using the `@event` annotation.

The `@event` annotation binds a dynamic routing subject to a strongly-typed data structure (the payload). This establishes a universal, transport-agnostic contract that can be seamlessly implemented across message brokers like NATS, RabbitMQ, Kafka, HTTP Webhooks, etc.

## Syntax

The `@event` annotation takes a string argument representing the routing subject, and it must be attached to an `object type` declaration.

```vdl
@event("auth.user_created.{userId}")
type UserCreatedEvent {
  userId string
  email string
  timestamp datetime
}
```

## Template and Validation Rules

The routing subject is a string literal that can contain dynamic placeholders in `{fieldName}` format.

To ensure events are fully self-contained and agnostic to the underlying network transport, VDL enforces strict schema validation on event placeholders:

1. **Explicit Field Binding:** Every placeholder used in the `@event` subject **must** exactly match a field defined within the annotated `type`.
2. **Type Safety:** If a subject references `{tenantId}`, the `tenantId` field must exist in the type definition so its data type is statically known.

```vdl
// VALID: 'orderId' exists in the payload
@event("orders.status_changed.{orderId}")
type OrderStatusChanged {
  orderId string
  status string
}

// INVALID: 'region' is used in the subject but missing from the payload
@event("logistics.{region}.delivery_failed")
type DeliveryFailed {
  deliveryId string
}
```

This strict validation guarantees that when an event is persisted to a database or forwarded via a protocol without native routing headers (like HTTP), the payload retains 100% of the context required to identify it.

## Generated API Shape

Generators transform an `@event` declaration into primary components: the payload data structure and the subject formatting utilities. Depending on the target, generators may also emit a centralized event catalog.

### TypeScript Shape

```ts
// 1. The Payload Type
export interface UserCreatedEvent {
  userId: string;
  email: string;
  timestamp: string;
}

// 2. The Subject Formatter
export function buildUserCreatedEventSubject(userId: string): string {
  return `auth.user_created.${userId}`;
}
```

### Go Shape

```go
// 1. The Payload Struct
type UserCreatedEvent struct {
    UserId    string    `json:"userId"`
    Email     string    `json:"email"`
    Timestamp time.Time `json:"timestamp"`
}

// 2. The Subject Formatter
func BuildUserCreatedEventSubject(userId string) string {
    return "auth.user_created." + userId
}
```

### Event Catalogs

Plugins should aggregate `@event` declarations into a centralized catalog (e.g., a map or dictionary) to facilitate runtime event routing. The catalog metadata must expose the literal routing string strictly under the `Subject` key.

```go
// Conceptual Catalog Shape
type EventMetadata struct {
    Name    string // e.g., "UserCreatedEvent"
    Subject string // e.g., "auth.user_created.{userId}"
}
```
