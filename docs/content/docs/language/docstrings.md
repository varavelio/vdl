+++
title = "Docstrings"
description = "Document VDL schemas with inline and external Markdown docs."
template = "docs.html"
weight = 4
+++

## What Docstrings Are

Docstrings are triple-quoted documentation blocks.

```vdl
"""
Represents a user account in the product.
"""
type User {
  id string
}
```

Unlike comments, docstrings are part of the schema model. Plugins can use them to generate docs, comments, JSON Schema descriptions, explorers, and metadata.

## Declaration Docstrings

Place a docstring immediately before a `type`, `enum`, or `const` to attach it to that declaration.

```vdl
""" A product that can be listed in the catalog. """
type Product {
  id string
  name string
}

""" Lifecycle state of a product. """
enum ProductStatus {
  Draft
  Published
}

""" Current public API version. """
const apiVersion = "1.0.0"
```

## Field Docstrings

Place a docstring immediately before a field to document that field.

```vdl
type User {
  """ Stable unique identifier. """
  id string

  """ Email address used for account notifications. """
  email string
}
```

Field docstrings also work inside inline objects.

```vdl
type Checkout {
  payment {
    """ Amount charged in cents. """
    amountCents int
  }
}
```

## Enum Member Docstrings

Enum members can have docstrings too.

```vdl
enum InvoiceStatus {
  """ The invoice has been created but not sent. """
  Draft

  """ The invoice has been paid in full. """
  Paid
}
```

The docstring attaches to the named member that follows it.

## Standalone Docstrings

A docstring at the top level can stand alone as schema documentation.

```vdl
"""
# Billing Schema

This schema contains shared billing contracts used by services and jobs.
"""

type Invoice {
  id string
}
```

A blank line separates a standalone docstring from the next declaration.

## Organizing Field Docstrings

Inside object type bodies, a docstring attaches to the field that follows it. This can also be used to make groups of fields easier to scan, but remember that the docstring belongs to the next field.

```vdl
type Product {
  """
  Identifiers
  """
  id string
  sku string

  """
  Public catalog data
  """
  name string
  description? string
}
```

In this example, `""" Identifiers """` documents `id`, and `""" Public catalog data """` documents `name`.

Use grouped wording sparingly. Most fields should be documented directly with field-specific descriptions.

## External Markdown Files

If a docstring contains only a relative `.md` path, VDL treats it as a reference to an external Markdown file.

```vdl
""" ./docs/product.md """
type Product {
  id string
  name string
}
```

The path is resolved relative to the `.vdl` file that contains the docstring.

Standalone external docs are also supported:

```vdl
""" ./docs/overview.md """

type Account {
  id string
}
```

## Good Docstring Style

- Explain what a contract means, not just its name.
- Mention units when fields are numeric.
- Mention formats when strings carry structured values.
- Keep generated-code readers in mind.

Good:

```vdl
type Payment {
  """ Amount charged in the smallest currency unit, such as cents. """
  amountMinor int
}
```

Less useful:

```vdl
type Payment {
  """ Amount. """
  amountMinor int
}
```
