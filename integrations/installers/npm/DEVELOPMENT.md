# NPM Installer for VDL

This directory contains the npm package installer for VDL. It provides a seamless installation experience by downloading the appropriate pre-compiled binary for the user's platform.

## How It Works

1. **User installs**: `npm install -g @varavel/vdl`
2. **Postinstall runs**: `install.js` executes automatically
3. **Archive downloads**: Downloads from GitHub releases based on platform/arch
4. **Archive check**: Verify the archive SHA256 hash against the release `manifest.json` pinned inside the package (downloaded from the GitHub release at publish time, never fetched at install time)
5. **Binary extracts**: Extracts to `bin/` directory
6. **Ready to use**: `vdl` command is now available

## Publishing to npm

### Prerequisites

Follow the new trusted publishers mechanism https://docs.npmjs.com/trusted-publishers

### Automatic Publishing

The package is automatically published when you create a new release:

```bash
# Create a GitHub release from the tag
# The CI workflow will:
# 1. Build all binaries
# 2. Upload binaries to GitHub releases
# 3. Download manifest.json from the release into the npm directory
# 4. Publish to npm with the same version (manifest.json ships pinned in the package)
```

### Platform not supported

Currently supported platforms:

- macOS (darwin): x64, arm64
- Linux: x64, arm64
- Windows: x64, arm64

## Version Management

The package version is automatically updated from the git tag during the release workflow:

```bash
# Tag format: v0.1.0
# Package version: 0.1.0 (v prefix removed)
```
