+++
title = "Project Configuration"
description = "Configure VDL generation with vdl.config.vdl."
template = "docs.html"
weight = 4
+++

## What `vdl.config.vdl` Does

`vdl.config.vdl` tells `vdl generate` which plugins to run, which schema each plugin should read, where generated files should be written, and which plugin-specific options should be passed in.

The file is itself written in VDL. VDL reads a constant named `config` from this file.

```vdl
const config = {
  version 1
  plugins [
    {
      src "varavelio/vdl-plugin-ts@v0.1.4"
      schema "./schema.vdl"
      outDir "./gen/ts"
    }
  ]
}
```

The file must be named exactly:

```text
vdl.config.vdl
```

## Minimal Configuration

```vdl
const config = {
  version 1
  plugins [
    {
      src "varavelio/vdl-plugin-json-schema@v0.1.0"
      schema "./schema.vdl"
      outDir "./gen/json-schema"
    }
  ]
}
```

This configuration:

- uses config format version `1`
- runs one plugin
- analyzes `./schema.vdl`
- writes generated files under `./gen/json-schema`

## Top-Level Fields

| Field         | Required | Description                                                     |
| ------------- | -------- | --------------------------------------------------------------- |
| `version`     | yes      | Configuration format version. Currently use `1`.                |
| `cleanOutDir` | no       | Whether VDL cleans output directories before writing new files. |
| `plugins`     | no       | List of plugin runs to execute.                                 |
| `remotes`     | no       | Authentication settings for private plugin hosts.               |
| `hooks`       | no       | Host shell commands run before or after generation.             |

## `version`

Set `version` to `1`.

```vdl
const config = {
  version 1
}
```

VDL rejects unsupported config versions.

## `plugins`

`plugins` is an array of plugin configurations.

```vdl
const config = {
  version 1
  plugins [
    {
      src "varavelio/vdl-plugin-go@v0.1.3"
      schema "./schema.vdl"
      outDir "./gen/go"
      options {
        package "contracts"
      }
    }
  ]
}
```

Each plugin entry has these fields:

| Field            | Required | Description                                                         |
| ---------------- | -------- | ------------------------------------------------------------------- |
| `src`            | yes      | Plugin JavaScript source.                                           |
| `schema`         | yes      | Path to the `.vdl` schema this plugin should process.               |
| `outDir`         | yes      | Directory where plugin output files should be written.              |
| `generateHeader` | no       | Whether VDL adds generated-file header comments. Default is `true`. |
| `options`        | no       | Plugin-specific string options passed through as-is.                |

## Plugin Sources

`src` can use one of three forms.

### GitHub Shorthand

```vdl
src "varavelio/vdl-plugin-go@v0.1.3"
```

This resolves to:

```text
https://raw.githubusercontent.com/varavelio/vdl-plugin-go/v0.1.3/dist/index.js
```

Rules:

- use `owner/repo@ref`
- the repository name must start with `vdl-plugin-`
- the plugin artifact must exist at `dist/index.js`
- prefer release tags or full commit hashes for reproducibility

### Local JavaScript File

```vdl
src "./plugins/my-plugin.js"
```

Local plugin paths must point to `.js` files and must start with `.` or `/`.

Use this while developing a plugin locally or for private in-repo generators.

### Remote HTTPS URL

```vdl
src "https://example.com/plugins/my-plugin/index.js"
```

Remote plugin URLs must point to `.js` files. HTTPS is required by default.

For local development, insecure HTTP can be enabled with `VDL_INSECURE_ALLOW_HTTP=true`.

## `schema`

`schema` points to the VDL file the plugin should compile.

```vdl
schema "./schema.vdl"
```

The path is resolved relative to `vdl.config.vdl` and must end with `.vdl`.

Different plugins can process the same schema:

```vdl
const config = {
  version 1
  plugins [
    {
      src "varavelio/vdl-plugin-go@v0.1.3"
      schema "./schema.vdl"
      outDir "./gen/go"
    }
    {
      src "varavelio/vdl-plugin-json-schema@v0.1.0"
      schema "./schema.vdl"
      outDir "./gen/json-schema"
    }
  ]
}
```

## `outDir`

`outDir` is the root directory for files returned by a plugin.

```vdl
outDir "./gen/go"
```

Plugins return relative file paths like `types.go` or `models/user.ts`. VDL writes them under `outDir`.

VDL validates generated paths so plugins cannot write outside `outDir`.

If two plugin outputs try to write the same final file path in one generation run, generation fails.

## `options`

`options` is a map of string values passed to the plugin.

```vdl
options {
  package "contracts"
  strict "true"
  importExtension "js"
}
```

All option values are strings. If a plugin option behaves like a boolean, pass values such as `"true"` or `"false"` according to that plugin's documentation.

Options are plugin-specific. Always check the plugin README or [available plugins guide](/docs/guides/plugins/).

## `generateHeader`

By default, VDL adds generated-file headers when writing plugin outputs.

```vdl
const config = {
  version 1
  plugins [
    {
      src "varavelio/vdl-plugin-go@v0.1.3"
      schema "./schema.vdl"
      outDir "./gen/go"
      generateHeader false
    }
  ]
}
```

Use `generateHeader false` only when a target format cannot safely contain comments or when the plugin already handles headers.

## `cleanOutDir`

By default, VDL replaces configured output directories with the newly generated files.

```vdl
const config = {
  version 1
  cleanOutDir true
  plugins [
    {
      src "varavelio/vdl-plugin-ts@v0.1.4"
      schema "./schema.vdl"
      outDir "./gen/ts"
    }
  ]
}
```

Set `cleanOutDir false` to merge generated files into existing directories without deleting files that are not part of the current output.

```vdl
const config = {
  version 1
  cleanOutDir false
  plugins [
    {
      src "./plugins/custom.js"
      schema "./schema.vdl"
      outDir "./gen/custom"
    }
  ]
}
```

Use merge mode carefully. Old generated files may remain if a plugin stops producing them.

## `remotes`

Use `remotes` when a plugin is hosted behind authentication.

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

`host` should not include `https://` or `http://`.

When multiple remotes match a plugin URL, VDL uses the most specific host.

Supported authentication forms:

- GitHub token
- custom header
- bearer token
- basic auth

Each remote auth entry should define exactly one authentication method.

### GitHub Token

```vdl
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
```

VDL reads the token from the named environment variable and sends it as a bearer token.

### Custom Header

```vdl
remotes [
  {
    host "plugins.example.com"
    auth {
      header {
        nameEnv "PLUGIN_HEADER_NAME"
        valueEnv "PLUGIN_HEADER_VALUE"
      }
    }
  }
]
```

### Bearer Token

```vdl
remotes [
  {
    host "plugins.example.com"
    auth {
      bearer {
        tokenEnv "PLUGIN_TOKEN"
      }
    }
  }
]
```

### Basic Auth

```vdl
remotes [
  {
    host "plugins.example.com"
    auth {
      basic {
        userEnv "PLUGIN_USER"
        passEnv "PLUGIN_PASS"
      }
    }
  }
]
```

## `hooks`

Hooks run shell commands on the host machine.

```vdl
const config = {
  version 1
  hooks {
    preGenerate ["npm run check"]
    postGenerate ["npm run format"]
  }
  plugins [
    {
      src "varavelio/vdl-plugin-ts@v0.1.4"
      schema "./schema.vdl"
      outDir "./gen/ts"
    }
  ]
}
```

Hook behavior:

- `preGenerate` commands run before plugins. The first failure stops generation.
- `postGenerate` commands run after files are written. Failures are printed as warnings and do not roll back generated files.
- commands run from the directory containing `vdl.config.vdl`
- hooks are skipped when `VDL_SKIP_HOST_HOOKS` or `VDL_CLOUD` is truthy

## Lock File

Remote plugin downloads are cached and recorded in `vdl.lock`.

The lock file stores hashes for remote plugin artifacts so VDL can detect unexpected changes.

Commit `vdl.lock` when your project depends on remote plugins.

## Complete Example

```vdl
const config = {
  version 1
  cleanOutDir true

  hooks {
    preGenerate ["npm run check"]
    postGenerate ["npm run format"]
  }

  plugins [
    {
      src "varavelio/vdl-plugin-go@v0.1.3"
      schema "./schema.vdl"
      outDir "./gen/go"
      options {
        package "contracts"
        strict "true"
      }
    }
    {
      src "varavelio/vdl-plugin-json-schema@v0.1.0"
      schema "./schema.vdl"
      outDir "./gen/json-schema"
      options {
        outFile "schema.json"
        root "User"
      }
    }
  ]
}
```

## Common Patterns

Generate data models and documentation from the same schema:

```vdl
const config = {
  version 1
  plugins [
    {
      src "varavelio/vdl-plugin-ts@v0.1.4"
      schema "./schema.vdl"
      outDir "./generated/types"
    }
    {
      src "varavelio/vdl-plugin-explorer@v0.1.1"
      schema "./schema.vdl"
      outDir "./generated/docs"
    }
  ]
}
```

Generate Go models and Go RPC code into separate packages:

```vdl
const config = {
  version 1
  plugins [
    {
      src "varavelio/vdl-plugin-go@v0.1.3"
      schema "./schema.vdl"
      outDir "./internal/contracts"
      options {
        package "contracts"
      }
    }
    {
      src "varavelio/vdl-plugin-rpc-go@v0.1.2"
      schema "./schema.vdl"
      outDir "./internal/rpcclient"
      options {
        target "client"
        package "rpcclient"
        typesImport "example.com/project/internal/contracts"
      }
    }
  ]
}
```

## Troubleshooting

If `vdl generate` cannot find the config file, make sure the file is named exactly `vdl.config.vdl`.

If a plugin source fails to resolve, check whether `src` is a local `.js` path, an HTTPS `.js` URL, or a valid GitHub shorthand.

If a private plugin fails to download, check the matching `remotes` entry and the required environment variables.

If generated files disappear, check `cleanOutDir`. The default behavior is to replace output directories.
