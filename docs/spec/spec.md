---
title: UFO RPC IDL Specification
description: Full specification of the UFO RPC Interface Definition Language
---

## Overview

UFO is a modern **IDL (Interface Definition Language)** designed for Schema-First development. It provides a declarative syntax for defining RPC services, data structures, and contracts with strong typing that UFO RPC can interpret and generate code for.

The primary goal of URPC is to offer an intuitive, human-readable format that ensures the best possible developer experience (DX) while maintaining type safety.

This IDL serves as the single source of truth for your projects, from which you can generate type-safe code for multiple programming languages.

## UFO Syntax

This is the syntax for the IDL.

```ufo
include "./foo.ufo"

// <comment>

/*
  <multiline comment>
*/

""" <Standalone documentation> """

""" <Type documentation> """
type <CustomTypeName> {
  """ <Field documentation> """
  <field>[?]: <Type>
}

""" <Constant documentation> """
const <ConstantName> = <Value>

""" <Enum documentation> """
enum <EnumName> {
  <EnumMember>[ = <EnumValue>]
  <EnumMember>[ = <EnumValue>]
}

""" <Pattern documentation> """
pattern <PatternName> = "<PatternValue>"

""" <RPC documentation> """
rpc <RPCName> {
  """ <RPC Standalone documentation> """

  """ <Procedure documentation> """
  proc <ProcedureName> {
    input {
      """ <Field documentation> """
      <field>[?]: <PrimitiveType> | <CustomType>
    }

    output {
      """ <Field documentation> """
      <field>[?]: <PrimitiveType> | <CustomType>
    }
  }

  """ <Stream documentation> """
  stream <StreamName> {
    input {
      """ <Field documentation> """
      <field>[?]: <PrimitiveType> | <CustomType>
    }

    output {
      """ <Field documentation> """
      <field>[?]: <PrimitiveType> | <CustomType>
    }
  }
}
```

## Naming Conventions

UFO RPC enforces consistent naming conventions to ensure code generated across all target languages is idiomatic and predictable. The built-in formatter will automatically apply these styles to your schema.

| Element                                         | Convention         | Example                  |
| :---------------------------------------------- | :----------------- | :----------------------- |
| Types, Enums, RPC services, Procedures, Streams | `PascalCase`       | `UserProfile`, `GetUser` |
| Fields (in types, input, and output blocks)     | `camelCase`        | `userId`, `createdAt`    |
| Constants                                       | `UPPER_SNAKE_CASE` | `MAX_PAGE_SIZE`          |
| Patterns                                        | `PascalCase`       | `UserEventSubject`       |
| Enum Members                                    | `PascalCase`       | `Pending`, `InProgress`  |

> **Note:** The formatter will automatically correct casing when you run it on your `.ufo` files, so you can focus on the logic while the tooling ensures consistency.

## Includes

To maintain clean and maintainable projects, UFO RPC allows you to split your schemas into multiple files. This modular approach helps you organize your types and procedures by domain, making them easier to navigate and reuse across different schemas.

### How to use Includes

You can include other `.ufo` files using the `include` keyword, typically at the top of your file.

```ufo
// auth.ufo
type Session {
  token: string
  expiresAt: datetime
}

// main.ufo
include "./auth.ufo"

type AuthInfo {
  session: Session
}
```

### Core Principles

- **Flat Context:** When a file is included, all its definitions (types, enums, constants, etc.) are merged into the global context. You can think of it as copying the content of the included file into the current file.
- **Relative Paths:** Includes always use relative paths (e.g., `./common.ufo`) starting from the current file's directory.
- **Idempotent Processing:** Each file is processed only once. If your project structure leads to the same file being included multiple times, the compiler simply skips files it has already processed, preventing duplication.

This system empowers you to build a robust library of common types while keeping your service-specific logic focused and uncluttered.

## Data Types

Data types are the core components of your schema. They define the precise structure of information exchanged between the client and server, ensuring consistency and type safety across your entire application.

### Primitive Types

The UFO IDL provides several built-in primitive types that map directly to standard JSON types while maintaining strong typing.

| Type       | JSON Equivalent | Description                                 |
| :--------- | :-------------- | :------------------------------------------ |
| `string`   | string          | UTF-8 encoded text string.                  |
| `int`      | integer         | 64-bit signed integer.                      |
| `float`    | number          | 64-bit floating point number.               |
| `bool`     | boolean         | A logical value: `true` or `false`.         |
| `datetime` | string          | An ISO 8601 formatted date and time string. |

### Data Structures

You can combine primitive and custom types into more complex structures to represent your data accurately.

#### Arrays

Represent an ordered collection of elements. All elements in an array must share the same type.

```ufo
// Syntax: <Type>[]
string[]      // A list of strings
User[]        // A list of User objects
```

#### Maps

Represent a collection of key-value pairs where keys are always strings. Maps are useful for lookups and dynamic dictionaries.

```ufo
// Syntax: map<<ValueType>>
map<int>      // Example: { "active": 1, "pending": 5 }
map<User>     // Example: { "user_123": { ... } }
```

#### Inline Types (Anonymous Objects)

Define a structure "on the fly" without naming it. This is useful for small, localized data structures that don't need to be reused elsewhere.

```ufo
{
  latitude: float
  longitude: float
}
```

### Custom Types

For reusable data structures, you can define named `type` blocks. These serve as the blueprint for your application's domain models.

```ufo
"""
Represents a user in the system.
"""
type User {
  id: string
  username: string
  email: string
}
```

#### Type Reuse: Composition & Destructuring

UFO RPC provides two powerful ways to share fields between types, allowing you to build complex models from simpler ones while avoiding duplication.

**1. Composition (Nesting)**
Include one type as a property of another. This creates a clear hierarchy and relationship between objects.

```ufo
type AuditMetadata {
  createdAt: datetime
  updatedAt: datetime
}

type Article {
  title: string
  content: string
  metadata: AuditMetadata  // Nested relationship
}
```

**2. Destructuring (Spreading)**
Merge the fields of one type directly into another using the `...` operator. This is ideal for "inheriting" fields from a base structure.

```ufo
type Article {
  ...AuditMetadata        // Fields are flattened into Article
  title: string
  content: string
}

// Equivalent to:
// type Article {
//   createdAt: datetime
//   updatedAt: datetime
//   title: string
//   content: string
// }
```

You can destructure multiple types in a single definition:

```ufo
type FullEntity {
  ...AuditMetadata
  ...OwnershipInfo
  name: string
}
```

> **Important:** Field names must be unique across the entire type. If two destructured types share a field name, or if you define a field that already exists in a destructured type, the compiler will raise an error. You cannot override fields from destructured types.

#### Field Modifiers

- **Required by Default:** All fields are mandatory. The compiler ensures that these fields are present during communication.
- **Optional Fields:** Use the `?` suffix to mark a field as optional.
  ```ufo
  type Profile {
    bio?: string  // This field can be omitted or null
  }
  ```

#### Documentation

Adding documentation to your types and fields is highly recommended. These comments are preserved by the compiler and used to generate readable documentation for API consumers.

```ufo
type Product {
  """ The unique identifier for the SKU. """
  sku: string

  """ Optional marketing description. """
  description?: string
}
```

## Constants

Constants allow you to define fixed values that can be referenced throughout your schema and in the generated code. They are useful for configuration values, limits, or any other static data that should be shared across your application.

```ufo
"""
Optional documentation for the constant.
"""
const <ConstantName> = <Value>
```

Constants support the following value types:

- **Strings:** `const API_VERSION = "v1"`
- **Integers:** `const MAX_PAGE_SIZE = 100`
- **Floats:** `const DEFAULT_TAX_RATE = 0.21`
- **Booleans:** `const FEATURE_FLAG_ENABLED = true`

```ufo
""" The maximum number of items allowed per request. """
const MAX_ITEMS = 50

""" The current API version string. """
const VERSION = "2.1.0"
```

## Enumerations

Enums define a set of named, discrete values. They are ideal for representing a fixed list of options, such as statuses, categories, or modes. UFO RPC supports two types of enums: **string enums** and **integer enums**.

```ufo
"""
Optional documentation for the enum.
"""
enum <EnumName> {
  <Member1>
  <Member2>
}
```

### Enum Type Inference

The type of an enum is inferred from the **first member's value**:

- If the first member has no explicit value or is assigned a string, the enum is a **string enum**.
- If the first member is assigned an integer, the enum is an **integer enum**.

All members within an enum must be of the same type. Mixing string and integer values in a single enum will result in a compiler error.

### String Enums

String enums are the default. If no value is assigned, the member name itself is used as the value. You can also assign explicit string values.

```ufo
// Implicit values (member name is used as the value)
enum OrderStatus {
  Pending
  Processing
  Shipped
  Delivered
  Cancelled
}

// Explicit string values
enum HttpMethod {
  Get = "GET"
  Post = "POST"
  Put = "PUT"
  Delete = "DELETE"
}
```

### Integer Enums

If the first member is assigned an integer value, the enum becomes an integer enum. **All members must have explicit integer values**; there is no auto-increment behavior.

```ufo
enum Priority {
  Low = 1
  Medium = 2
  High = 3
  Critical = 10
}
```

## Patterns

Patterns are template strings that generate helper functions for constructing dynamic string values at runtime. They are particularly useful for defining message queue topics, cache keys, routing paths, or any other string that requires interpolation.

```ufo
"""
Optional documentation for the pattern.
"""
pattern <PatternName> = "<template_string>"
```

### Syntax

A pattern template uses `{placeholder}` syntax for dynamic segments. Each placeholder becomes a `string` parameter in the generated function.

```ufo
""" Generates a NATS subject for user-specific events. """
pattern UserEventSubject = "events.users.{userId}.{eventType}"

""" Generates a Redis cache key for a session. """
pattern SessionCacheKey = "cache:session:{sessionId}"
```

### Generated Code

The compiler transforms each pattern into a function that accepts the placeholders as arguments and returns the constructed string. For example, the `UserEventSubject` pattern above would generate something similar to:

```typescript
// TypeScript
function UserEventSubject(userId: string, eventType: string): string {
  return `events.users.${userId}.${eventType}`;
}
```

```go
// Go
func UserEventSubject(userId string, eventType string) string {
  return "events.users." + userId + "." + eventType
}
```

This makes patterns a powerful tool for ensuring consistency across your codebase when working with message brokers like NATS, Kafka, or RabbitMQ, or when defining structured cache keys for Redis or Memcached.

## RPC Services

An `rpc` block acts as a logical container for your API's communication endpoints. It allows you to group related **Procedures** and **Streams** under a single named service, providing better organization and a clearer structure for your generated clients and server implementation.

```ufo
"""
Optional documentation for the entire service.
"""
rpc <RPCName> {
  // Procedure and Stream definitions go here
}
```

### RPC Merging Across Files

To facilitate large-scale project organization, UFO RPC supports **RPC merging**. If the same `rpc` block name is declared in multiple files (for example, via includes), the compiler will automatically merge their contents into a single, unified service.

This allows you to split a large service definition across multiple files by domain or feature:

```ufo
// users_procs.ufo
rpc Users {
  proc GetUser { ... }
  proc CreateUser { ... }
}

// users_streams.ufo
rpc Users {
  stream UserStatusUpdates { ... }
}

// main.ufo
include "./users_procs.ufo"
include "./users_streams.ufo"

// The "Users" RPC now contains GetUser, CreateUser, and UserStatusUpdates.
```

> **Important:** While RPC blocks are merged, duplicate procedure or stream names within the same RPC will cause a compiler error. Each endpoint name must be unique within its service.

### Procedures (`proc`)

Procedures are the standard way to define request-response interactions. They represent discrete actions that a client can trigger on the server. They must be defined inside an `rpc` block.

```ufo
rpc <RPCName> {
  """
  Describes the purpose of this procedure.
  """
  proc <ProcedureName> {
    input {
      """ Field-level documentation. """
      <field>: <Type>
    }

    output {
      """ Field-level documentation. """
      <field>: <Type>
    }
  }
}
```

### Streams (`stream`)

Streams enable real-time, unidirectional communication from the server to the client using Server-Sent Events (SSE). They are designed for scenarios where the server needs to push updates as they happen. They must be defined inside an `rpc` block.

```ufo
rpc <RPCName> {
  """
  Describes the nature of the events being streamed.
  """
  stream <StreamName> {
    input {
      """ Subscription parameters. """
      <field>: <Type>
    }

    output {
      """ Event data structure. """
      <field>: <Type>
    }
  }
}
```

### Input and Output Blocks

The `input` and `output` blocks in procedures and streams behave exactly like **inline types**. This means they support all the same features, including field documentation and **destructuring** with the `...` operator.

This is particularly useful for sharing common request or response fields across multiple endpoints:

```ufo
type PaginationParams {
  page: int
  limit: int
}

type PaginatedResponse {
  totalItems: int
  totalPages: int
}

rpc Articles {
  proc ListArticles {
    input {
      ...PaginationParams
      filterByAuthor?: string
    }

    output {
      ...PaginatedResponse
      items: Article[]
    }
  }
}
```

### Service Example

Grouping related functionality makes your schema easier to maintain:

```ufo
rpc Messaging {
  """ Sends a new message to a specific channel. """
  proc SendMessage {
    input {
      channelId: string
      text: string
    }
    output {
      messageId: string
      sentAt: datetime
    }
  }

  """ Real-time feed of messages for a channel. """
  stream NewMessages {
    input {
      channelId: string
    }
    output {
      sender: string
      text: string
      timestamp: datetime
    }
  }
}
```

## Documentation

### Docstrings

Docstrings can be used in two ways: associated with specific elements or as standalone documentation.

**1. Associated Docstrings**

These are placed immediately before an element definition and provide specific documentation for that element. Associated docstrings can be used with: `type`, `rpc`, `proc`, `stream`, `enum`, `const`, `pattern`, and individual fields.

```ufo
"""
This is documentation for MyType.
"""
type MyType {
  """ This is documentation for myField. """
  myField: string
}
```

**2. Standalone Docstrings**

These provide general documentation for the schema (or an RPC block) and are not associated with any specific element. To create a standalone docstring, ensure there is at least one blank line between the docstring and any following element.

Standalone docstrings can be placed at the schema level (outside of any block) or inside an `rpc` block. When inside an RPC, they become part of that service's documentation, which is useful for adding section headers or contextual notes for a group of endpoints.

```ufo
"""
# Welcome
This is general documentation for the entire schema.
"""

type MyType {
  // ...
}

rpc MyService {
  """
  # User Management
  The following procedures handle user lifecycle operations.
  """

  proc CreateUser { ... }
  proc DeleteUser { ... }
  proc CreateSession { ... }
}
```

### Multi-line Docstrings and Indentation

Docstrings support Markdown syntax, allowing you to format your documentation with headings, lists, code blocks, and more.

Since docstrings can contain Markdown, whitespace is significant for formatting constructs like lists or code blocks. To prevent conflicts with URPC's own syntax indentation, UFO RPC automatically normalizes multi-line docstrings.

The leading whitespace from the first non-empty line is considered the baseline indentation. This baseline is then removed from every line in the docstring. This process preserves the _relative_ indentation, ensuring that Markdown formatting remains intact regardless of how the docstring block is indented in the source file.

_Example:_

In the following docstring, the first line has 4 spaces of indentation, which will be removed from all lines.

```ufo
type MyType {
  """
    This is a multi-line docstring.

    The list below will be rendered correctly:

    - Level 1
      - Level 2
  """
  field: string
}
```

The resulting content for rendering will be:

```markdown
This is a multi-line docstring.

The list below will be rendered correctly:

- Level 1
  - Level 2
```

Remember to keep your documentation up to date with your schema changes.

### External Documentation Files

For extensive documentation, you can reference external Markdown files instead of writing content inline. When a docstring contains only a valid path to a `.md` file, the compiler will read the content of that file and use it as the documentation.

**Important:** External file paths must always be **relative to the `.ufo` file** that references them. If the specified file does not exist, the compiler will raise an error.

```ufo
// Standalone documentation from external files
""" ./docs/welcome.md """
""" ./docs/authentication.md """

rpc Users {
  // Associated documentation from an external file
  """ ./docs/create-user.md """
  proc CreateUser {
    // ...
  }
}
```

This approach helps maintain clean and focused schema files while allowing for detailed, long-form documentation in separate files. Remember to keep external documentation files up to date with your schema changes.

## Deprecation

URPC provides a mechanism to mark elements as deprecated, signaling to API consumers that a feature should no longer be used in new code and may be removed in a future version. This applies to `type`, `rpc`, `proc`, `stream`, `enum`, `const`, and `pattern` definitions.

### Basic Deprecation

To mark an element as deprecated without a specific message, use the `deprecated` keyword directly before its definition:

```ufo
deprecated type LegacyUser {
  // ...
}

deprecated rpc OldService {
  // ...
}

deprecated enum OldStatus {
  // ...
}

deprecated const OLD_LIMIT = 100

deprecated pattern OldQueueName = "legacy.{id}"

rpc MyService {
  deprecated proc FetchData {
    // ...
  }

  deprecated stream OldStream {
    // ...
  }
}
```

### Deprecation with Message

To provide additional context, such as a migration path or a removal timeline, include a message in parentheses:

```ufo
deprecated("Use UserV2 instead")
type LegacyUser {
  // ...
}

deprecated("This service will be removed in v3.0. Migrate to NewService.")
rpc OldService {
  // ...
}
```

### Placement

The `deprecated` keyword must be placed between any associated docstring and the element definition:

```ufo
"""
Original documentation for the type.
"""
deprecated("Use NewType instead")
type MyType {
  // type definition
}
```

### Effects

Deprecated elements will:

- Be visually flagged in the generated playground and documentation.
- Generate warning comments in the output code to alert developers.
- Continue to function normally in the generated code; deprecation is purely informational.

## Complete Example

The following example demonstrates a comprehensive schema that uses all the features of the UFO IDL, including includes, constants, enums, patterns, types with composition and destructuring, and RPC services with procedures and streams.

```ufo
include "./foo.ufo"

// ============================================================================
// External Documentation
// ============================================================================
""" ./docs/welcome.md """
""" ./docs/authentication.md """

// ============================================================================
// Constants
// ============================================================================
""" Maximum number of items returned in a single page. """
const MAX_PAGE_SIZE = 100

""" Current API version. """
const API_VERSION = "1.0.0"

// ============================================================================
// Enumerations
// ============================================================================
""" Represents the status of an order in the system. """
enum OrderStatus {
  Pending
  Processing
  Shipped
  Delivered
  Cancelled
}

""" Priority levels for support tickets. """
enum Priority {
  Low = 1
  Medium = 2
  High = 3
  Critical = 10
}

// ============================================================================
// Patterns
// ============================================================================
""" Generates a NATS subject for product-related events. """
pattern ProductEventSubject = "events.products.{productId}.{eventType}"

""" Generates a Redis cache key for a user session. """
pattern SessionCacheKey = "cache:session:{sessionId}"

// ============================================================================
// Shared Types
// ============================================================================
"""
Common fields for all auditable entities.
"""
type AuditMetadata {
  id: string
  createdAt: datetime
  updatedAt: datetime
}

"""
Standard pagination parameters for list requests.
"""
type PaginationParams {
  page: int
  limit: int
}

"""
Standard pagination response metadata.
"""
type PaginatedResponse {
  totalItems: int
  totalPages: int
  currentPage: int
}

// ============================================================================
// Domain Types
// ============================================================================
"""
Represents a product in the catalog.
"""
type Product {
  ...AuditMetadata

  """ The name of the product. """
  name: string

  """ The price of the product in USD. """
  price: float

  """ The current order status for this product listing. """
  status: OrderStatus

  """ The date when the product will be available. """
  availabilityDate: datetime

  """ A list of tags for categorization. """
  tags?: string[]
}

"""
Represents a customer review for a product.
"""
type Review {
  """ The rating, from 1 to 5. """
  rating: int

  """ The customer's written feedback. """
  comment: string

  """ The ID of the user who wrote the review. """
  userId: string
}

// ============================================================================
// RPC Services
// ============================================================================
"""
Catalog Service

Provides operations for managing products and browsing the catalog.
"""
rpc Catalog {
  """
  # Product Lifecycle
  Endpoints for creating and managing products.
  """

  """
  Creates a new product in the system.
  """
  proc CreateProduct {
    input {
      product: Product
    }

    output {
      success: bool
      productId: string
    }
  }

  """
  Retrieves a product by its ID, including its reviews.
  """
  proc GetProduct {
    input {
      productId: string
    }

    output {
      product: Product
      reviews: Review[]
    }
  }

  """
  Lists all products with pagination support.
  """
  proc ListProducts {
    input {
      ...PaginationParams
      filterByStatus?: OrderStatus
    }

    output {
      ...PaginatedResponse
      items: Product[]
    }
  }
}

"""
Chat Service

Provides real-time messaging capabilities.
"""
rpc Chat {
  """
  Sends a message to a chat room.
  """
  proc SendMessage {
    input {
      """ The ID of the chat room. """
      chatId: string

      """ The content of the message. """
      message: string
    }

    output {
      """ The ID of the created message. """
      messageId: string

      """ The server timestamp of when the message was recorded. """
      timestamp: datetime
    }
  }

  """
  Subscribes to new messages in a specific chat room.
  """
  stream NewMessage {
    input {
      chatId: string
    }

    output {
      id: string
      message: string
      userId: string
      timestamp: datetime
    }
  }
}
```

## Limitations

The UFO IDL is designed to be simple and focused. As such, there are a few constraints to be aware of:

1.  **Reserved Keywords:** All language keywords (e.g., `type`, `rpc`, `proc`, `stream`, `enum`, `const`, `pattern`, `input`, `output`, `include`, `deprecated`, etc.) cannot be used as identifiers for your types, fields, or services.
2.  **Validation Logic:** The compiler handles type checking and ensures required fields are present. Any additional business validation logic (e.g., "rating must be between 1 and 5") must be implemented in your application code.
