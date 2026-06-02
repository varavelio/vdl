---
title: Formatting Guide
description: Standard formatting conventions for VDL
---

This guide defines how to keep VDL files easy to read, stable in diffs, and predictable across teams.

> **Note:** These conventions are enforced by the official formatter. Run `vdl format ./schema.vdl` or use editor format-on-save through the VDL LSP.

## 1. General Principles

- Format for consistency, not personal style.
- Prefer the smallest readable layout.
- Keep related items close; separate unrelated items with one blank line.
- Use UTF-8, LF (`\n`), no trailing whitespace, and a final newline.

## 2. Indentation

- Use **2 spaces** per indentation level.
- Do not use tabs.

```vdl
type Example {
  field string
}
```

## 3. Top-Level Elements

Top-level elements are `include`, `type`, `const`, `enum`, comments, and standalone docstrings.

- Group `include` statements at the top with no blank lines inside the group.
- Use one blank line between top-level declarations.
- Avoid multiple consecutive blank lines.

```vdl
include "./common.vdl"
include "./auth.vdl"

const maxPageSize = 100

type Order {
  id string
  total float
}

enum OrderStatus {
  Pending
  Paid
}
```

## 4. Blocks and Fields

### 4.1 Block Structure

- Put `{` on the same line as the declaration, with one space before it.
- Put block members on their own lines.
- Align `}` with the declaration start.

```vdl
type User {
  id string
  name string
}
```

### 4.2 Field Members

- Keep one member per line.
- Prefer no blank lines between plain fields.
- Add one blank line around field docstrings when it improves scanning.
- Place spreads before regular fields when possible.

```vdl
type User {
  ...AuditMetadata

  """ The user's display name. """
  name string

  """ The user's email address. """
  email string

  age? int
}
```

### 4.3 Annotation Placement

Annotations carry domain semantics. Keep placement strict and vertical for readability.

- Put annotations directly above the declaration or member they target.
- Use one annotation per line.
- Keep docstring above annotations when both are present.
- Keep a single blank line between logical groups of annotated members.

```vdl
"""
User operations.
"""
@rpc
type Users {
  @proc
  GetUser {
    input {
      userId string
    }

    output {
      id string
      email string
    }
  }

  @stream
  UserEvents {
    input {
      userId string
    }
    output {
      event string
      at datetime
    }
  }
}
```

## 5. Spacing

| Context                | Rule                              | Example                   |
| :--------------------- | :-------------------------------- | :------------------------ |
| Field syntax           | `<name>[?] <type>`                | `email? string`           |
| Assignment             | One space around `=`              | `const retries = 3`       |
| Braces (`{}`)          | One space before `{`              | `type User {`             |
| Arrays (`[]`)          | No spaces inside or before suffix | `string[]`, `int[][]`     |
| Maps (`map[...]`)      | No spaces inside brackets         | `map[int]`, `map[User]`   |
| Annotation call        | No space before `(`               | `@meta({ owner "core" })` |
| Object literal entries | Space-separated key/value         | `{ host "localhost" }`    |
| Array literal entries  | Space-separated elements          | `["a" "b" "c"]`           |
| Enum member assignment | One space around `=`              | `Archived = "archived"`   |

## 6. Comments

VDL supports two comment styles, and formatter preserves comment content exactly as written.

- **Single-line:** `// ...`
- **Multi-line:** `/* ... */`
- Comments can appear at top level or inside blocks.

```vdl
// Top-level single-line comment

/*
  Top-level multi-line comment
*/

type Example {
  // Inside block
  field string
}
```

## 7. Docstrings

- Place docstrings immediately above the element they document.
- Enclose in triple quotes (`"""`), preserving internal newlines and formatting.
- Prefer concise, purpose-first text.
- Use a blank line after a standalone docstring node.

```vdl
"""
Represents a user in the system.
"""
type User {
  """ The unique identifier. """
  id string

  """ The user's full name. """
  name string
}
```

## 8. Naming Conventions

The formatter automatically enforces the following naming conventions:

| Element         | Convention   | Example                      |
| :-------------- | :----------- | :--------------------------- |
| Types and Enums | `PascalCase` | `UserProfile`, `OrderStatus` |
| Enum Members    | `PascalCase` | `Pending`, `InProgress`      |
| Type Members    | `camelCase`  | `userId`, `createdAt`        |
| Constants       | `camelCase`  | `maxPageSize`, `apiVersion`  |
| Annotations     | `camelCase`  | `rpc`, `proc`, `deprecated`  |

### Acronym Handling

Acronyms longer than two letters are treated as regular words:

```vdl
// Correct
type HttpRequest { ... }
type JsonParser { ... }

// Incorrect
type HTTPRequest { ... }
type JSONParser { ... }
```
