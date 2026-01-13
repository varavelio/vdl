---
title: Formatting Guide
description: Standard formatting conventions for the UFO IDL
---

This document specifies the standard formatting rules for the UFO IDL. Consistent formatting enhances code readability, maintainability, and collaboration.

> **Note:** All style conventions are automatically enforced by the official UFO formatter. Run it manually with `ufo fmt ./schema.ufo`, or let the built-in LSP formatter (bundled with the VS Code extension) format files on save.

## 1. General Principles

- **Encoding:** UTF-8.
- **Line Endings:** Use newline characters (`\n`).
- **Trailing Whitespace:** None.
- **Final Newline:** End non-empty files with one newline.

## 2. Indentation

- Use **2 spaces** per indentation level.
- Do not use tabs.

```ufo
type Example {
  field: string
}
```

## 3. Top-Level Elements

Top-level elements include `import`, `type`, `enum`, `const`, `pattern`, `rpc`, and standalone docstrings or comments.

- **Imports:** Group consecutive `import` statements together without blank lines between them.
- **Separation:** Separate each top-level element with one blank line.
- **Preservation:** Intentionally placed blank lines in the source are respected.

```ufo
import "./common.ufo"
import "./auth.ufo"

const MAX_PAGE_SIZE = 100

type Order {
  id: string
  total: float
}

rpc Orders {
  proc GetOrder { ... }
}
```

## 4. Blocks and Fields

### 4.1 Block Structure

- Opening braces (`{`) are on the same line as the declaration, preceded by one space.
- Contents inside non-empty blocks start on a new, indented line.
- The closing brace (`}`) is placed on its own line, aligned with the declaration.

```ufo
type User {
  id: string
  name: string
}
```

### 4.2 RPC Blocks

Procedures (`proc`) and streams (`stream`) must be defined inside an `rpc` block. Separate each endpoint with one blank line.

```ufo
rpc Service {
  proc Get {
    input {
      id: string
    }
    output {
      data: string
    }
  }

  stream Watch {
    input {
      filter: string
    }
    output {
      event: string
    }
  }
}
```

### 4.3 Fields

- Each field is placed on its own line.
- Fields without docstrings can be placed consecutively without blank lines.
- When a field has a docstring, separate it from the preceding field with one blank line.
- The spread operator (`...`) should be placed at the beginning of the block.

```ufo
type User {
  ...AuditMetadata

  """ The user's display name. """
  name: string

  """ The user's email address. """
  email: string

  age?: int
}
```

### 4.4 Input and Output Blocks

In procedures and streams, separate the `input` and `output` blocks with one blank line.

```ufo
proc CreateUser {
  input {
    name: string
    email: string
  }

  output {
    userId: string
    createdAt: datetime
  }
}
```

## 5. Spacing

| Context               | Rule                               | Example          |
| :-------------------- | :--------------------------------- | :--------------- |
| Colons (`:`)          | No space before, one space after   | `field: string`  |
| Braces (`{`)          | One space before in declarations   | `type User {`    |
| Brackets (`[]`)       | No spaces for array types          | `string[]`       |
| Optional marker (`?`) | Immediately follows the field name | `email?: string` |
| Map syntax            | No extra spaces                    | `map<User>`      |

## 6. Comments

Comment content is preserved exactly, including internal whitespace.

- **Standalone Comments:** Use `//` or `/* ... */` on their own lines, indented to the current block level.
- **End-of-Line Comments:** Place after code on the same line, with at least one space separating them.

```ufo
// This is a standalone comment
type Example {
  field: string  // End-of-line comment
}
```

## 7. Docstrings

- Place docstrings immediately above the element they document.
- Enclose in triple quotes (`"""`), preserving internal newlines and formatting.
- For fields, prefer concise, single-line docstrings.

```ufo
"""
Represents a user in the system.
"""
type User {
  """ The unique identifier. """
  id: string

  """ The user's full name. """
  name: string
}
```

## 8. Deprecation

The `deprecated` keyword marks elements as deprecated. Place it on its own line immediately before the element definition. If a docstring exists, place `deprecated` between the docstring and the element.

```ufo
deprecated type LegacyUser {
  // ...
}

"""
Documentation for this service.
"""
deprecated rpc OldService {
  deprecated proc FetchData {
    // ...
  }
}

deprecated("Use NewType instead")
type OldType {
  // ...
}
```

## 9. Naming Conventions

The formatter automatically enforces the following naming conventions:

| Element      | Convention         | Example                 |
| :----------- | :----------------- | :---------------------- |
| Types        | `PascalCase`       | `UserProfile`           |
| Enums        | `PascalCase`       | `OrderStatus`           |
| Enum Members | `PascalCase`       | `Pending`, `InProgress` |
| RPC Services | `PascalCase`       | `UserService`           |
| Procedures   | `PascalCase`       | `GetUser`               |
| Streams      | `PascalCase`       | `NewMessages`           |
| Patterns     | `PascalCase`       | `UserEventSubject`      |
| Fields       | `camelCase`        | `userId`, `createdAt`   |
| Constants    | `UPPER_SNAKE_CASE` | `MAX_PAGE_SIZE`         |

### Acronym Handling

Acronyms longer than two letters are treated as regular words:

```ufo
// Correct
type HttpRequest { ... }
type JsonParser { ... }

// Incorrect
type HTTPRequest { ... }
type JSONParser { ... }
```
