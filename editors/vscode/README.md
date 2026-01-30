# VDL VSCode Extension

Official VDL extension for Visual Studio Code or any VSCode based editor.

- [VDL Docs](https://vdl.varavel.com)
- [VDL Repository](https://github.com/varavelio/vdl)
- [VDL VSCode Extension Repository](https://github.com/varavelio/vdl/tree/main/editors/vscode)

## Installation

- [Install from the VSCode Marketplace](https://marketplace.visualstudio.com/items?itemName=varavel.vdl)
- [Install from Open VSX Registry](https://open-vsx.org/extension/varavel/vdl)
- [Download the `.vsix` file from the releases page](https://github.com/varavelio/vdl/releases)

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

## Snippets

The following snippets are available for `.vdl` files:

- `type`: Define a new custom data type
- `field`: Add a field to a type
- `enum`: Define an enumeration (String or Integer)
- `const`: Define a global constant
- `pattern`: Define a string interpolation pattern
- `map`: Define a map type (keys are always strings)
- `rpc`: Define a new RPC service block
- `proc`: Define a request-response procedure inside an RPC
- `stream`: Define a unidirectional stream inside an RPC
- `include`: Include another VDL file
- `deprecated`: Mark the next element as deprecated
- `doc`: Insert a standalone documentation block
- `docfile`: Reference an external markdown documentation file

## Release Notes

Below are release notes for the last 10 versions, you can also see the entire [changelog](./CHANGELOG.md).

### 0.1.0

Initial release of VDL for VSCode (formerly UFO RPC).
