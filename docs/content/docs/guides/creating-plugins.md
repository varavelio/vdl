+++
title = "Creating Plugins"
description = "A practical guide to writing VDL plugins."
template = "docs.html"
weight = 6
+++

> **Note:** You do not need Node.js or any JavaScript runtime installed on your machine. VDL executes plugin JavaScript code through [Goja](https://github.com/dop251/goja), an embedded ECMAScript runtime built into the VDL binary.

## The Short Version

A VDL plugin is just a JavaScript file that exports a `generate(input)` function.

```js
exports.generate = (input) => {
  return {
    files: [
      {
        path: "summary.txt",
        content: `This schema has ${input.ir.types.length} types.\n`,
      },
    ],
  };
};
```

Then reference it from `vdl.config.vdl`:

```vdl
const config = {
  version 1
  plugins [
    {
      src "./plugins/summary.js"
      schema "./schema.vdl"
      outDir "./gen"
    }
  ]
}
```

Run:

```bash
vdl generate
```

VDL compiles `schema.vdl`, calls your plugin function, and writes the returned files under `./gen`.

## How Plugins Work

The generation flow is simple.

1. VDL reads `vdl.config.vdl`.
2. VDL resolves each configured plugin source.
3. VDL analyzes the configured `schema` file.
4. VDL converts the valid schema into the generator-facing IR.
5. VDL executes the plugin's `generate(input)` function in a JavaScript runtime.
6. The plugin returns generated files or structured errors.
7. VDL validates output paths and writes files inside `outDir`.

The plugin input is defined by the canonical [`schemas/plugin_input.vdl`](https://github.com/varavelio/vdl/blob/main/schemas/plugin_input.vdl) contract and has this shape:

```ts
type PluginInput = {
  version: string;
  ir: IrSchema;
  options: Record<string, string>;
};
```

The plugin output is defined by the canonical [`schemas/plugin_output.vdl`](https://github.com/varavelio/vdl/blob/main/schemas/plugin_output.vdl) contract and has this shape:

```ts
type PluginOutput = {
  files?: Array<{
    path: string;
    content: string;
  }>;
  errors?: Array<{
    message: string;
    position?: {
      file: string;
      line: number;
      column: number;
    };
  }>;
};
```

`input.version` is the VDL version. `input.ir` is the fully resolved Intermediate Representation of the schema. `input.options` contains the key-value options from the plugin block in `vdl.config.vdl`.

Options are strings in the config schema. If you need booleans, numbers, lists, or enums, parse and validate them in your plugin.

## Minimal JavaScript Plugin

Create `plugins/models-list.js`:

```js
exports.generate = (input) => {
  const lines = [];

  lines.push("# VDL Types");
  lines.push("");

  for (const typeDef of input.ir.types) {
    lines.push(`- ${typeDef.name}`);
  }

  return {
    files: [
      {
        path: "models.md",
        content: `${lines.join("\n")}\n`,
      },
    ],
  };
};
```

Configure it:

```vdl
const config = {
  version 1
  plugins [
    {
      src "./plugins/models-list.js"
      schema "./schema.vdl"
      outDir "./gen/docs"
    }
  ]
}
```

Run:

```bash
vdl generate
```

Result:

```text
gen/docs/models.md
```

## Export Formats

VDL accepts any of these forms:

```js
exports.generate = (input) => ({ files: [] });
```

```js
module.exports.generate = (input) => ({ files: [] });
```

## Important Runtime Rules

- The plugin runs as JavaScript, not as a full Node.js process.
- Keep the release artifact self-contained; do not rely on runtime filesystem or npm package loading.
- `console.log`, `console.info`, `console.warn`, and `console.error` are available for plugin logs.
- Returned file paths must be relative and must stay inside `outDir`.
- If the plugin returns `errors`, VDL stops and does not write generated files.
- Throwing an unexpected error fails the plugin run.
- Remote plugins should be pinned to a release tag or full commit hash for reproducibility.

## Plugin Sources

`src` in `vdl.config.vdl` can point to several kinds of plugin artifact.

| Source kind      | Example                                   | Notes                                     |
| ---------------- | ----------------------------------------- | ----------------------------------------- |
| Local `.js` file | `./plugins/my-plugin.js`                  | Must start with `.` or `/`.               |
| HTTPS `.js` URL  | `https://example.com/vdl-plugin/index.js` | Must point to a `.js` file.               |
| GitHub shorthand | `varavelio/vdl-plugin-go@v0.1.3`          | Resolves to `dist/index.js` in that repo. |

GitHub shorthand repositories must be named with the `vdl-plugin-` prefix. VDL caches remote plugin files and records hashes in `vdl.lock`.

## Authenticated Remote Plugins

Private plugin hosts can be configured through `remotes` in `vdl.config.vdl`. VDL matches each plugin URL against the most specific configured remote host and reads credentials from environment variables.

```vdl
const config = {
  version 1
  remotes [
    {
      host "github.com/my-org/private-vdl-plugin"
      auth {
        github {
          tokenEnv "GITHUB_TOKEN"
        }
      }
    }
  ]
  plugins [
    {
      src "my-org/private-vdl-plugin@v1.0.0"
      schema "./schema.vdl"
      outDir "./gen/private"
    }
  ]
}
```

Supported auth styles are GitHub token, custom header, bearer token, and basic auth.

## Using The Plugin SDK

Plain JavaScript is enough for small plugins. For serious plugins, use [`@varavel/vdl-plugin-sdk`](https://github.com/varavelio/vdl-plugin-sdk).

The SDK gives you:

- typed `definePlugin(...)` authoring API
- typed access to the VDL IR and plugin input/output contracts
- structured error helpers such as `PluginError`, `fail`, and `assert`
- utility imports for strings, arrays, RPC validation, and other common generator tasks
- testing builders for realistic IR fixtures
- a `vdl-plugin` CLI for type-checking and bundling plugins
- TypeScript configuration presets for app and test code

Minimal SDK plugin:

```ts
import { definePlugin } from "@varavel/vdl-plugin-sdk";

export const generate = definePlugin((input) => {
  return {
    files: [
      {
        path: "hello.txt",
        content: `Hello from VDL ${input.version}\n`,
      },
    ],
  };
});
```

> **How exports work:** SDK plugins use `export const generate`, not `exports.generate`. The `vdl-plugin build` command converts the ESM export into `exports.generate` automatically when bundling into `dist/index.js`. The final artifact that VDL loads always uses the CommonJS `exports.generate` form.

Check and build:

```bash
npx vdl-plugin check
npx vdl-plugin build
```

`check` runs TypeScript without emitting files. `build` bundles the plugin into `dist/index.js`, which is the artifact VDL consumes for GitHub shorthand plugins.

## Starting From The Template

The fastest way to start a production-ready plugin is [`varavelio/vdl-plugin-template`](https://github.com/varavelio/vdl-plugin-template).

The template includes:

- `src/index.ts` with a minimal `definePlugin(...)` entrypoint
- the VDL plugin SDK
- TypeScript config
- `@varavel/gen` for structured text/code generation
- Biome and dprint for linting and formatting
- Vitest for tests
- a devcontainer
- a GitHub Actions CI workflow
- `dist/index.js` as the distributable plugin bundle

Basic workflow:

```bash
npm install
npm run check
npm run build
npm run test
```

When releasing a GitHub-hosted plugin:

1. Build with `npm run build`.
2. Commit source files and `dist/index.js`.
3. Create a release tag such as `v0.1.0`.
4. Users can reference `owner/vdl-plugin-name@v0.1.0` from `vdl.config.vdl`.

VDL plugins do not need to be published to npm. The GitHub release plus committed `dist/index.js` is enough for VDL to fetch the plugin.

## Error Handling

Return structured errors when the user's schema or options are invalid.

```ts
import { definePlugin, fail } from "@varavel/vdl-plugin-sdk";

export const generate = definePlugin((input) => {
  const root = input.options.root;

  if (!root) {
    fail('Missing required option "root".');
  }

  const typeDef = input.ir.types.find((item) => item.name === root);

  if (!typeDef) {
    fail(`Unknown root type "${root}".`);
  }

  return {
    files: [{ path: "root.txt", content: `${typeDef.name}\n` }],
  };
});
```

For schema-specific errors, attach a source position when possible. VDL will print the diagnostic with file, line, and column information.

## Testing Plugins

The SDK testing entry point helps you build plugin inputs without manually writing a full IR object.

```ts
import {
  field,
  objectType,
  pluginInput,
  primitiveType,
  schema,
  typeDef,
} from "@varavel/vdl-plugin-sdk/testing";

const input = pluginInput({
  options: { root: "User" },
  ir: schema({
    types: [
      typeDef("User", objectType([field("id", primitiveType("string"))])),
    ],
  }),
});
```

Pass the input to your plugin handler and assert generated files or returned errors.

## Practical Checklist

- Validate plugin options before generating output.
- Validate annotation models before assuming a shape such as RPC or events.
- Generate deterministic output so diffs stay clean.
- Keep generated file paths relative to `outDir`.
- Return structured errors for user mistakes.
- Use the SDK for TypeScript, typed IR access, bundling, and tests.
- Commit `dist/index.js` for GitHub-hosted plugins.
- Pin plugin versions in consuming projects.
