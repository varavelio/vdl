---
title: UFO DSL Specification (.ufo)
description: The official specification for the UFO Domain-Specific Language (v1)
---

## 1. Overview

The UFO DSL (`.ufo`) is a minimalist language for defining type-safe APIs, data contracts, and services. It is designed to be **simple to read, easy to write**, and to serve as the single source of truth for your entire stack.

From a `.ufo` schema, you can generate client SDKs, server stubs, and documentation for any language.

## 2. Hello World

```ufo
namespace Greeter

service GreeterService {
    func SayHello(name: string): string
}
```

## 3. General Syntax

### 3.1 Namespace

Every file must declare a namespace at the top. This namespace groups your types and services.

```ufo
namespace My.App
```

### 3.2 Imports

You can import other `.ufo` files or directories.

```ufo
import "./common"         // Imports everything from the 'common' folder
import User "./users"     // Imports with an alias (e.g. User.Profile)
```

### 3.3 Visibility

By default, everything is **private** to the file. Use `export` to expose types or services to other files.

```ufo
type InternalConfig { ... }  // Private
export type User { ... }     // Public
```

## 4. Services (RPC)

Services define the operations your API supports.

```ufo
export service UserService {
    // A standard Request/Response function
    func GetUser(id: string): User

    // A real-time stream (Server-Sent Events / WebSockets)
    stream OnUserUpdated(id: string): User
}
```

### 4.1 Functions (`func`)

Functions take arguments and return a single response.

```ufo
// func Name(args...): ReturnType
func CreatePost(title: string, content: string): Post
```

### 4.2 Streams (`stream`)

Streams allow the server to push data to the client over time.

```ufo
// stream Name(args...): YieldType
stream SubscribeToTicker(symbol: string): float
```

## 5. Data Types

### 5.1 Types (Structs)

Types define the shape of your data.

```ufo
export type User {
    id: string
    name: string
    age?: int        // Optional field
    tags: string[]   // Array of strings
}
```

### 5.2 Composition (Spread)

Reuse fields from other types using `...Type`.

```ufo
type Timestamped {
    createdAt: datetime
    updatedAt: datetime
}

export type Product {
    ...Timestamped   // Inherits createdAt, updatedAt
    id: string
    price: float
}
```

### 5.3 Enums

Enums define a fixed set of constants.

```ufo
export enum Role {
    ADMIN
    USER
    GUEST
}

// Enums with explicit integer values
export enum Status: int {
    ACTIVE = 1
    INACTIVE = 0
}
```

### 5.4 Primitive Types

| Type       | Description                              |
| :--------- | :--------------------------------------- |
| `string`   | UTF-8 text                               |
| `int`      | 64-bit integer                           |
| `float`    | 64-bit float                             |
| `bool`     | `true` or `false`                        |
| `datetime` | ISO 8601 date string                     |
| `map<K,V>` | Key-Value map (e.g., `map<string, int>`) |

## 6. Project Structure

UFO encourages a clean, predictable project structure.

### 6.1 The Namespace Rule

All files in the same folder **must** belong to the same namespace. This keeps your project organization (folders) and logical organization (namespaces) in sync.

```text
src/
  users/
    user.ufo      -> namespace Users
    profile.ufo   -> namespace Users
  orders/
    order.ufo     -> namespace Orders
```

### 6.2 Barrel Files

A file without a `namespace` is a "barrel" (usually `index.ufo`). It simply exports other files to create a clean public API for a folder.

```ufo
// src/index.ufo
export "./users"
export "./orders"
```

## 7. Advanced Features

### 7.1 Constants

Define shared constants for your application.

```ufo
export const MAX_RETRIES = 3
export const API_VERSION = "v1"
```

### 7.2 Patterns

Define string templates for event subjects or topic names.

```ufo
export pattern UserEvents = "users.{id}.events"
```

## 8. Configuration (`ufo.yaml`)

The `ufo.yaml` file at your project root configures the code generator.

```yaml
version: 1
# Entry point for your schema
schema: "./src/index.ufo"

# Generators
generates:
  # Generate Go Server definitions
  "./internal/gen":
    plugin: "go-server"

  # Generate TypeScript Client
  "./frontend/src/client":
    plugin: "ts-client"
```
