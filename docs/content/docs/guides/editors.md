+++
title = "Editor Integration"
description = "Set up VDL support in your code editor."
template = "docs.html"
weight = 3
+++

## Overview

VDL works with any editor that supports the Language Server Protocol (LSP). First-class integrations are available for Visual Studio Code and it's forks.

## Visual Studio Code

The official VDL extension provides full language support for Visual Studio Code and any VSCode-based editor (VSCodium, Cursor, Windsurf, and others).

### Installation

Install the extension from either marketplace:

- [Visual Studio Code Marketplace](https://marketplace.visualstudio.com/items?itemName=varavel.vdl)
- [Open VSX Registry](https://open-vsx.org/extension/varavel/vdl)

The extension is plug and play. The only requirement is having the `vdl` binary installed on your system. If you have not installed the CLI yet, follow the [installation guide](/docs/guides/installation/).

Extension source code: [integrations/editors/vscode](https://github.com/varavelio/vdl/tree/main/integrations/editors/vscode)

### Features

Once installed, opening any `.vdl` file gives you:

- Syntax highlighting
- Error highlighting and diagnostics
- Code autocompletion
- Auto-formatting on save
- Go-to-definition, hover info, find references, and rename
- Code snippets for types, fields, enums, constants, annotations, and more

For configuration options, available commands, and troubleshooting, see the extension page on the [VSCode Marketplace](https://marketplace.visualstudio.com/items?itemName=varavel.vdl) or [Open VSX](https://open-vsx.org/extension/varavel/vdl).

## Neovim

VDL's built-in LSP works with Neovim's native LSP client. The language server provides go-to-definition, hover, references, rename, completions, document symbols, and document links for `.vdl` files.

Syntax highlighting is not included through the LSP and may be added in a future release.

The `vdl` binary must be installed and available on your `PATH`. For setup instructions, see the [Neovim integration README](https://github.com/varavelio/vdl/tree/main/editors/neovim).

## Other Editors

The `vdl lsp` command speaks the standard Language Server Protocol over stdin and stdout, so any editor with LSP support can integrate with it. Point your editor's LSP client to `vdl lsp` for `.vdl` files.

VDL's LSP does not provide syntax highlighting on its own. Some editors require a separate TextMate grammar ([vdl.tmLanguage.json](https://github.com/varavelio/vdl/blob/main/editors/vscode/language/vdl.tmLanguage.json)) or Tree-sitter parser for highlighting. Syntax highlighting for editors beyond VSCode is planned for a future release.

To get started, check your editor's documentation on configuring a [custom language server](https://microsoft.github.io/language-server-protocol/implementors/servers/).
