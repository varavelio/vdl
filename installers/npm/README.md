# Varavel Definition Language (VDL)

Open-source cross-language definition engine for modern stacks. Define your data structures, APIs, contracts, and generate type-safe code for your backend and frontend instantly.

## Installation

```bash
npm install -g @varavel/vdl
```

Or as a dev dependency:

```bash
npm install --save-dev @varavel/vdl
```

## Usage

After installation, the `vdl` command will be available:

```bash
vdl --help # See available commands
```

## Programmatic Usage

You can also use this package programmatically in Node.js:

```javascript
const vdl = require("@varavel/vdl");
const { execSync } = require("child_process");

// Get the path to the binary
console.log(vdl.binaryPath);

// Execute VDL commands
execSync(`"${vdl.binaryPath}" version`, { stdio: "inherit" });
```

## Supported Platforms

- **macOS**: x64, ARM64 (Apple Silicon)
- **Linux**: x64, ARM64
- **Windows**: x64, ARM64

## How It Works

This package downloads the appropriate pre-compiled VDL binary for your platform during installation. The binary is fetched from the [official GitHub releases](https://github.com/varavelio/vdl/releases).

## Documentation

Visit https://varavel.com/vdl for full documentation.

## License

MIT
