---
title: Formatting Guide
description: A guide to correctly format UFO DSL
---

This document specifies the standard formatting rules for the UFO DSL. Consistent formatting enhances code readability, maintainability, and
collaboration. The primary goal is to produce clean, predictable, and
aesthetically pleasing UFO code.

> **⚠️ Reference Only:** All style conventions are automatically enforced by the official UFO formatter. Run it manually with `ufo fmt ./schema.ufo`, or let the built-in LSP formatter (bundled with the VS Code extension and configurable for other editors) format files on save.

## 1. General Principles

- **Encoding:** UTF-8.
- **Line Endings:** Use newline characters (`\n`).
- **Trailing Whitespace:** None.
- **Final Newline:** End non-empty files with one newline.

## 2. Indentation

- Use **2 spaces** per indentation level.
- Do not use tabs.

_Example:_

```ufo
type Example {
  field: string
}
```

## 3. Top-Level Elements

Top-level elements include `namespace`, `import`, `export`, and standalone comments.

- **Namespace:** Must be the first non-comment line of the file (unless it's a barrel file).
- **Default:** Separate each top-level element with one blank line.
- **Exceptions:**
  - **Consecutive Imports:** Do not insert blank lines between consecutive `import` statements.
  - **Consecutive Comments:** Do not insert extra blank lines between consecutive standalone comments.
- **Preservation:** Intentionally placed blank lines in the source are respected.

_Example:_

```ufo
namespace Orders

import "./shared"
import Auth "./users/auth"

// A standalone comment
export type Order {
  field: string
}

export type Item {
  field: int
}
```

## 4. Blocks and Fields

### 4.1 RPC Blocks

`proc` and `stream` definitions must be grouped within an `rpc` block.

```ufo
export rpc Service {
  proc Get { ... }

  stream Watch { ... }
}
```

### 4.2 Fields in a Type

- Each field is placed on its own line.
- **Field Separation:** For simple fields without complex formatting, fields may
  be placed consecutively without blank lines. When a field has a docstring,
  separate it from the preceding field with one blank line for readability.
- **Spread Operator:** The `...` operator should be placed at the beginning of the block or grouped with other spread operators.

_Example:_

```ufo
export type User {
  ...BaseEntity
  ...AuthInfo

  """ The user's name. """
  name: string

  email: string
}
```

### 4.2 Blocks (Type, Input/Output)

- Opening braces (`{`) are on the same line as the declaration header (preceded
  by one space).
- Contents inside non-empty blocks always start on a new, indented line.
- The closing brace (`}`) is placed on its own line, aligned with the opening
  line.
- In procedure and stream bodies, separate the `input`, and `output` blocks with
  one blank line.

## 5. Spacing

- **Colons (`:`):** No space before; one space after (e.g. `field: string`).
- **Commas (`,`):** No space before; one space after.
- **Braces (`{` and `}`):** One space before `{` in declarations; inside blocks,
  use newlines and proper indentation.
- **Brackets (`[]`):** No spaces for array types (e.g. `string[]`); no extra
  interior spacing.
- **Parentheses (`()`):** No extra spaces inside the parentheses.
- **Optional Marker (`?`):** Immediately follows the field name (e.g.
  `email?: string`).

## 6. Comments

Comment content is preserved exactly (including internal whitespace).

- **Standalone Comments:** Use `//` or `/* … */` on their own lines; indent to
  the current block.
- **End-of-Line (EOL) Comments:** Can use either `//` or block style (`/* … */`)
  following code on the same line, with at least one space separating them.

_Example:_

```ufo
version 1 // EOL comment

type Example {
  field: string // Inline comment for field
}
```

## 7. Docstrings

- Place docstrings immediately above the `type`, `proc`, `stream`, or `field` they
  document.
- They are enclosed in triple quotes (`"""`), preserving internal newlines and
  formatting.

_Example:_

```ufo
"""
Docstring for MyType.
Can be multi-line.
"""
type MyType {
  // ...
}
```

### 7.1 Field Docstrings

- For fields, prefer concise, single-line docstrings.
- Place the docstring on the line immediately before the field it documents,
  indented to the same level.
- When multiple fields have docstrings, separate each one with a blank line to
  improve readability.

_Example:_

```ufo
type User {
  """ The user's unique identifier. """
  id: string

  """ The user's email address. Must be unique. """
  email: string

  name?: string // A field without a docstring does not require a blank line before it.
}
```

## 8. Deprecation

The `deprecated` keyword is used to mark types, procedures, or streams as
deprecated.

- Place the `deprecated` keyword on its own line immediately before the element
  definition
- If a docstring exists, place the `deprecated` keyword between the docstring
  and the element definition
- For deprecation with a message, use parentheses with the message in quotes

### 8.1 Basic Deprecation

_Example:_

```ufo
deprecated type MyType {
  // type definition
}

"""
Documentation for MyProc
"""
deprecated proc MyProc {
  // procedure definition
}

deprecated stream MyStream {
  // stream definition
}
```

### 8.2 Deprecation with Message

_Example:_

```ufo
"""
Documentation for MyType
"""
deprecated("Replaced by ImprovedType")
type MyType {
  // type definition
}

deprecated("Use NewStream instead")
stream MyStream {
  // stream definition
}
```

## 9. Naming Conventions

### 9.1 Type, Procedure, and Stream Names

- Use **strict PascalCase** (also known as UpperCamelCase). Each word starts with an uppercase letter with no underscores or consecutive capital letters.
- Acronyms longer than two letters should be treated as regular words (e.g. `HttpRequest`, not `HTTPRequest`).

_Example:_

```ufo
// Correct
type FooBar {
  myField: string
}

// Incorrect
type FooBAR {
  myField: string
}
```

### 9.2 Field Names

- Use **strict camelCase**. The first word is lowercase and each subsequent word starts with an uppercase letter. Do not use underscores or all-caps abbreviations.

_Example:_

```ufo
// Correct
myInput: FooBar
zipCode: string

// Incorrect
MyINPUT: FooBar
zip_code: string
```
