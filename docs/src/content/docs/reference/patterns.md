---
title: Pattern Annotations
description: Defining string templates through @pattern annotations on constants
---

## Overview

Pattern semantics are defined on constant declarations with the `@pattern` annotation.

The pattern template is the constant value itself. The annotation is a flag and does not take arguments.

## Syntax

```vdl
@pattern
const userEventSubject = "events.users.{userId}.{eventType}"

@pattern
const sessionCacheKey = "cache:session:{sessionId}"
```

`@pattern` applies to `const` declarations only.

## Template Rules

Pattern templates are string literals that can contain placeholders in `{name}` format.

```vdl
@pattern
const orderTopic = "orders.{orderId}.{status}"
```

Each placeholder represents a dynamic segment consumed by generated helpers.

## Generated API Shape

Generators transform a pattern constant into helper utilities that accept placeholder values and return the final string.

TypeScript shape:

```ts
function orderTopic(orderId: string, status: string): string {
  return `orders.${orderId}.${status}`;
}
```

Go shape:

```go
func OrderTopic(orderID string, status string) string {
	return "orders." + orderID + "." + status
}
```

Exact naming and packaging can vary by generator target, but the template contract remains the same.

## Example

```vdl
@pattern
const notificationSubject = "notifications.{tenantId}.{channel}.{eventType}"

@pattern
const productCacheKey = "cache:product:{productId}"
```

This model is useful for queue subjects, routing keys, and cache keys where consistency is required across services.
