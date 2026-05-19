---
title: Installing VDL
description: Complete installation options for the VDL CLI
---

## Overview

Choose the installation method that best fits your operating system and workflow.

| Platform          | Method     | Command                                                        |
| ----------------- | ---------- | -------------------------------------------------------------- |
| **Linux / macOS** | Shell      | <code>curl -fsSL https://get.varavel.com/vdl &#124; sh</code>  |
| **Linux / macOS** | Homebrew   | `brew install varavelio/tap/vdl`                               |
| **Windows**       | PowerShell | <code>irm https://get.varavel.com/vdl.ps1 &#124; iex</code>    |
| **Any**           | NPM        | `npm install -D @varavel/vdl`                                  |
| **Any**           | Manual     | [Download binaries](https://github.com/varavelio/vdl/releases) |

After installation, verify that the CLI is available:

```bash
vdl --version
```

## Linux and macOS

### Shell Installer

The shell installer is the quickest way to install VDL on Linux and macOS.

```bash
curl -fsSL https://get.varavel.com/vdl | sh
```

For more installation options using this installer, visit [https://get.varavel.com/vdl](https://get.varavel.com/vdl).

Install a specific version:

```bash
curl -fsSL https://get.varavel.com/vdl | VERSION=vx.x.x sh
```

Replace `vx.x.x` with the release tag you want to install, for example `v0.1.0`.

### Homebrew

Use Homebrew if you prefer managing CLI tools through taps.

```bash
brew install varavelio/tap/vdl
```

Install the latest experimental release:

```bash
brew install varavelio/tap/vdl-next
```

Install a specific version:

```bash
brew install varavelio/tap/vdl@x.x.x
```

Replace `x.x.x` with the version you want to install.

## Windows

Use the PowerShell installer on Windows.

```powershell
irm https://get.varavel.com/vdl.ps1 | iex
```

For more installation options using this installer, visit [https://get.varavel.com/vdl.ps1](https://get.varavel.com/vdl.ps1).

Install a specific version:

```powershell
$env:VERSION = "vx.x.x"; irm https://get.varavel.com/vdl.ps1 | iex
```

Replace `vx.x.x` with the release tag you want to install, for example `v0.1.0`.

## NPM

The npm package is cross-platform and works well when VDL should be pinned per project.

### Local Project Install

This is the recommended npm workflow for teams because it keeps every developer and CI job on the same VDL version.

```bash
npm install --save-dev @varavel/vdl
```

Then call VDL from package scripts, for example:

```json
{
  "scripts": {
    "vdl:generate": "vdl generate",
    "vdl:format": "vdl format"
  }
}
```

### Global Install

Use a global install when you want `vdl` available system-wide.

```bash
npm install --global @varavel/vdl
```

Install the latest experimental release:

```bash
npm install --global @varavel/vdl@next
```

Install a specific version:

```bash
npm install --global @varavel/vdl@x.x.x
```

For package details, visit the [npm package page](https://www.npmjs.com/package/@varavel/vdl).

## Manual Downloads

You can download prebuilt binaries from the [VDL releases page](https://github.com/varavelio/vdl/releases).

Manual installation is useful when you need to:

- install VDL in an environment without package managers
- mirror binaries internally
- pin a binary in a custom CI image
- inspect release assets before installing

Download the archive for your operating system and architecture, extract it, and place the `vdl` binary somewhere on your `PATH`.

## Choosing A Method

- Use the **shell installer** for the fastest Linux/macOS setup.
- Use **Homebrew** if you already manage developer tools with Homebrew.
- Use the **PowerShell installer** on Windows.
- Use **local npm install** for project-pinned VDL versions in Node-based repositories.
- Use **manual downloads** for custom distribution, offline environments, or CI images.
