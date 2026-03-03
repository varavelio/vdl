# VDL VSCode Extension

Official VDL extension for Visual Studio Code or any VSCode based editor.

- [VDL Docs](https://vdl.varavel.com)
- [VDL Repository](https://github.com/varavelio/vdl)
- [VDL VSCode Extension Repository](https://github.com/varavelio/vdl/tree/main/editors/vscode)

## Installation

- [Install from the VSCode Marketplace](https://marketplace.visualstudio.com/items?itemName=varavel.vdl)
- [Install from Open VSX Registry](https://open-vsx.org/extension/varavel/vdl)
- [Download the `.vsix` file from the releases page](https://github.com/varavelio/vdl/releases?q=vscode)

## Features

The following features are provided by the extension for `.vdl` files.

- Syntax highlighting
- Error highlighting
- Code autocompletion
- Auto formatting
- Code snippets

## Requirements

This extension requires the VDL (`vdl`) binary on your operating system PATH.

You can download the binary for any OS/Arch in the [VDL Releases page.](https://github.com/varavelio/vdl/releases)

For more information on how to install `vdl`, read the [VDL documentation.](https://vdl.varavel.com)

If the binary is not in your operating system PATH, you can set a custom path in the `vdl.binaryPath` setting.

## Extension Settings

- `vdl.enable`: Enable VDL language support. If disabled, the only feature that will be available is Syntax Highlighting. Default `true`.

- `vdl.binaryPath`: Path to the VDL binary. If not set, the extension will try to find `vdl` or `vdl.exe` in your PATH. Default `<not set>`.

## Commands

- `vdl.init`: Initialize a new `vdl.yaml` and `schema.vdl` files
- `vdl.restart`: Restart Language Server
- `vdl.openLogs`: Open Language Server Logs

## Snippets

The following snippets are available for `.vdl` files:

- `type`: Define a new named type
- `field`: Add a required field to a type
- `field?`: Add an optional field to a type
- `enum`: Define an enumeration with implicit string values
- `enums`: Define an enumeration with explicit string values
- `enumi`: Define an enumeration with explicit integer values
- `const`: Define a constant with an inferred type
- `constt`: Define a constant with an explicit type
- `ann`: Add a flag annotation (`@name`)
- `annarg`: Add an annotation with a scalar argument (`@name("value")`)
- `annobj`: Add an annotation with an object argument
- `map`: Use a map type (`map[ValueType]`)
- `include`: Include another VDL file
- `doc`: Insert a standalone documentation block
- `docfile`: Reference an external markdown documentation file

## Release Notes

Below are release notes for the last 10 versions, you can also see the entire [changelog](https://github.com/varavelio/vdl/blob/main/editors/vscode/CHANGELOG.md).

### v0.1.3 - 2026-02-02

- Fixed syntax highlighting for keywords used as field names (e.g., `input: string`, `output: int`, `type: bool`).

### v0.1.2 - 2026-02-01

- Refactored VDL syntax highlighting with granular declarations and improved pattern matching.

### v0.1.1 - 2026-01-30

- Added `vdl.openLogs` command to easily open the language server log file.

### v0.1.0 - 2026-01-29

- Initial release of VDL for VSCode (formerly UFO RPC).
