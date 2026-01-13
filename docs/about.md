---
title: About UFO RPC
description: About this project
---

## Modern RPC framework that puts developer experience first.

UFO RPC is a complete and modern API development ecosystem, designed for building robust, strongly-typed, and maintainable network services. It was born from the need for a tool that combines the simplicity and readability of REST APIs with the type safety and performance of more complex RPC frameworks.

## What is UFO RPC?

At its core, UFO RPC is a system built on a **schema as the single source of truth**. You define your API's structure once in a simple, human-readable language, and UFO RPC handles the rest.

The ecosystem is comprised of three fundamental pillars:

1.  **A Definition Language (DSL):** A simple and intuitive language (`.ufo`) for defining data types, procedures (request-response), and streams (real-time communication), as well as enums, constants, and messaging patterns.
2.  **First-Class Developer Tooling:** A suite of tools that makes working with the DSL a pleasure, including an **LSP** for autocompletion and validation in your editor, automatic formatters, and a **VSCode extension**.
3.  **Multi-Language Code Generators:** A powerful engine that reads your schema and generates fully functional, strongly-typed, and resilient clients and servers, initially for **Go** and **TypeScript**.

The result is a workflow where you can design, implement, document, and consume APIs with unprecedented speed and confidence.

## Core Philosophy

UFO RPC is built on a series of key principles:

- **Developer Experience (DX) First:** Tools should be intuitive and eliminate friction. From the DSL to the generated code and the Playground, everything is designed to be easy to use and to boost productivity.
- **Pragmatism Over Dogma:** We use standard, proven, and accessible technologies like **HTTP, JSON, and Server-Sent Events (SSE)**. We prioritize solutions that work in the real world and are easy to debug, rather than complex binary protocols or overly prescriptive architectures.
- **Simplicity is Power:** We believe that good design eliminates unnecessary complexity. UFO RPC offers the robustness of traditional RPC systems without the overhead and complex configuration.
- **Strong Contracts, Flexible Implementations:** The schema is the law. This guarantees end-to-end type safety. However, the framework is flexible, allowing for the injection of custom HTTP clients, application contexts, and business logic that is completely framework-agnostic.

## Key Features

- **A Human-Centric DSL:** Define your APIs in `.ufo` files that are as easy to read as they are to write. Documentation (in Markdown) lives alongside your code, ensuring it never becomes outdated.
- **Type-Safe, Multi-Language Code Generation:** First-class support for the most modern stacks. V1 includes:
  - **Go Server & Client:** For building high-performance backends.
  - **TypeScript Server & Client:** For seamless integration with the JavaScript/Node.js ecosystem and modern frontends.
  - **Dart Client:** For building mobile and desktop applications with Flutter.
- **"Batteries-Included" Interactive Playground:** Every UFO project can generate a static, self-contained web portal where developers can:
  - Explore all operations and data types.
  - Read the complete Markdown documentation.
  - Execute procedures and subscribe to streams directly from the browser via auto-generated forms.
  - Get ready-to-use `curl` commands and client code snippets.
- **Resilient Clients & Servers by Default:** The generated clients are not simple wrappers. They come with built-in policies for **retries with exponential backoff**, **per-request timeouts**, and **automatic reconnection for streams**, making your applications robust from day one.
- **Interoperability with OpenAPI:** Automatically generate an OpenAPI v3 specification from your `.ufo` schema, enabling integration with the vast ecosystem of existing tools.

## Who is UFO RPC for?

UFO RPC is the ideal tool for teams and developers who:

- Want the **type safety of gRPC** without the complexity of Protobuf and HTTP/2.
- Need clear, well-defined APIs but don't require the **querying flexibility of GraphQL**.
- Are looking for the **simplicity of tRPC** but need to support a **polyglot ecosystem** (beyond just TypeScript).
- Value a fast workflow, exceptional tooling, and the ability to move confidently between the backend and frontend.

In short, if you want to build modern APIs quickly and safely, UFO RPC is built for you.
