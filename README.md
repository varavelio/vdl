<img src="./assets/png/vdl.png" height="100"/>
<h1>Varavel Definition Language</h1>

VDL is the open-source cross-language definition engine for modern stacks. Define your data structures, APIs, contracts, and generate type-safe code for your backend and frontend instantly.

Learn more at [https://varavel.com/vdl](https://varavel.com/vdl).

## Installation

The easiest way to get VDL on your system. Select the method that best fits your workflow.

| Platform          | Method     | Command                                                        |
| ----------------- | ---------- | -------------------------------------------------------------- |
| **Linux / macOS** | Shell      | `curl -fsSL https://get.varavel.com/vdl \| sh`                 |
| **Linux / macOS** | Homebrew   | `brew install varavelio/tap/vdl`                               |
| **Windows**       | PowerShell | `irm https://get.varavel.com/vdl.ps1 \| iex`                   |
| **Any**           | NPM        | `npm install -D @varavel/vdl`                                  |
| **Any**           | Manual     | [Download binaries](https://github.com/varavelio/vdl/releases) |

### Linux & macOS

**Shell script:**

```bash
curl -fsSL https://get.varavel.com/vdl | sh
```

> [!TIP]
> For more installation options using this installer, visit [https://get.varavel.com/vdl](https://get.varavel.com/vdl).
>
> To install a specific version: `curl -fsSL https://get.varavel.com/vdl | VERSION=vx.x.x sh`.

**Homebrew:**

```bash
brew install varavelio/tap/vdl
```

> [!TIP]
> To install the latest experimental release: `brew install varavelio/tap/vdl-next`.
>
> To install a specific version: `brew install varavelio/tap/vdl@x.x.x`.

### Windows

```powershell
irm https://get.varavel.com/vdl.ps1 | iex
```

> [!TIP]
> For more installation options using this installer, visit [https://get.varavel.com/vdl.ps1](https://get.varavel.com/vdl.ps1).
>
> To install a specific version: `$env:VERSION = "vx.x.x"; irm https://get.varavel.com/vdl.ps1 | iex`.

### NPM (Cross platform)

**Local (Recommended):**
Ensures version consistency across your team. Use it from your `package.json` scripts.

```bash
npm install --save-dev @varavel/vdl
```

**Global:**
Makes `vdl` available system-wide.

```bash
npm install --global @varavel/vdl
```

> [!TIP]
> For more details using this package, visit the [npm package page](https://www.npmjs.com/package/@varavel/vdl).
>
> To install the latest experimental release: `npm install --global @varavel/vdl@next`.
>
> To install a specific version: `npm install --global @varavel/vdl@x.x.x`.

## What is VDL?

At its core, VDL is a system built on a **schema as the single source of truth**. You define your API's structure once in a simple, human-readable language, and VDL handles the rest.

The ecosystem is comprised of three fundamental pillars:

1.  **A Definition Language (DSL):** A simple and intuitive language (`.vdl`) for defining data types, procedures (request-response), and streams (real-time communication).
2.  **First-Class Developer Tooling:** A suite of tools that makes working with the DSL a pleasure, including an **LSP** for autocompletion and validation in your editor, automatic formatters, and a **VSCode extension**.
3.  **Multi-Language Code Generators:** A powerful engine that reads your schema and generates fully functional, strongly-typed, and resilient clients and servers, initially for **Go** and **TypeScript**.

The result is a workflow where you can design, implement, document, and consume APIs with unprecedented speed and confidence.

## Core Philosophy

VDL is built on a series of key principles:

- **Developer Experience (DX) First:** Tools should be intuitive and eliminate friction. From the DSL to the generated code and the Playground, everything is designed to be easy to use and to boost productivity.
- **Pragmatism Over Dogma:** We use standard, proven, and accessible technologies like **HTTP, JSON, and Server-Sent Events (SSE)**. We prioritize solutions that work in the real world and are easy to debug, rather than complex binary protocols or overly prescriptive architectures.
- **Simplicity is Power:** We believe that good design eliminates unnecessary complexity. VDL offers the robustness of traditional RPC systems without the overhead and complex configuration.
- **Strong Contracts, Flexible Implementations:** The schema is the law. This guarantees end-to-end type safety. However, the framework is flexible, allowing for the injection of custom HTTP clients, application contexts, and business logic that is completely framework-agnostic.

## Key Features

- **A Human-Centric DSL:** Define your APIs in `.vdl` files that are as easy to read as they are to write. Documentation (in Markdown) lives alongside your code, ensuring it never becomes outdated.
- **Type-Safe, Multi-Language Code Generation:** First-class support for the most modern stacks. V1 includes:
  - **Go Server & Client:** For building high-performance backends.
  - **TypeScript Server & Client:** For seamless integration with the JavaScript/Node.js ecosystem and modern frontends.
  - **Dart Client:** For building mobile and desktop applications with Flutter.
- **"Batteries-Included" Interactive Playground:** Every VDL project can generate a static, self-contained web portal where developers can:
  - Explore all operations and data types.
  - Read the complete Markdown documentation.
  - Execute procedures and subscribe to streams directly from the browser via auto-generated forms.
  - Get ready-to-use `curl` commands and client code snippets.
- **Resilient Clients & Servers by Default:** The generated clients are not simple wrappers. They come with built-in policies for **retries with exponential backoff**, **per-request timeouts**, and **automatic reconnection for streams**, making your applications robust from day one.
- **Interoperability with OpenAPI:** Automatically generate an OpenAPI v3 specification from your `.vdl` schema, enabling integration with the vast ecosystem of existing tools.

## Who is VDL for?

VDL is the ideal tool for teams and developers who:

- Want the **type safety of gRPC** without the complexity of Protobuf and HTTP/2.
- Need clear, well-defined APIs but don't require the **querying flexibility of GraphQL**.
- Are looking for the **simplicity of tRPC** but need to support a **polyglot ecosystem** (beyond just TypeScript).
- Value a fast workflow, exceptional tooling, and the ability to move confidently between the backend and frontend.

In short, if you want to build modern APIs quickly and safely, VDL is built for you.

## License

VDL is **100% open source** and released under MIT license:

### Disclaimer

VDL is provided "AS IS" without warranty of any kind. The authors assume no responsibility for damages or losses from using this software. You are responsible for testing generated code before production use.

See [LICENSE](LICENSE) for details.

---

_VDL is a labor of love by its contributors. We provide it freely in the hope it will be useful, but without any guarantees. Your success is your own responsibility, and your contributions back to the project are always welcome!_
