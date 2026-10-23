---
title: "VPP-2: HTTP-RPC"
description: "Standard HTTP-RPC modeling and transport behavior for VDL plugins."
path: "docs/protocols/http-rpc"
weight: 4
---

**VPP-2: HTTP-RPC** defines the standard way to model HTTP remote procedure calls in VDL and the standard HTTP transport behavior that compatible plugins must implement.

This protocol covers one operation style.

1. **Procedures:** request-response operations completed within a single HTTP transaction.

VPP-2 does not add syntax to VDL. It uses normal VDL declarations, inline object types, and annotations. The purpose of the protocol is to give plugin authors a shared contract for generating clients, servers, handlers, SDKs, documentation, adapters, and test tools from the same RPC model.

If a plugin claims compatibility with VPP-2, it must preserve the annotation semantics, operation model, HTTP behavior, and envelope format defined in this document.

## Why This Protocol Exists

RPC systems often look different across languages, frameworks, and runtimes. One stack may generate controller classes, another may generate function handlers, another may generate typed clients, and another may generate route metadata.

VDL should not hard-code any of those framework choices into the compiler. The compiler should keep producing a language-neutral representation of schemas and annotations. Plugins can then turn that representation into idiomatic code for each target ecosystem.

VPP-2 exists so those plugins still agree on the important parts: how RPC services are modeled, how operations are named, how input and output are represented, how HTTP requests are formed, and how errors are returned.

The result is a portable RPC contract. A VDL service should feel familiar whether the generated target is Go, TypeScript, Rust, Java, or an internal platform plugin.

VPP-2 is specifically the standard RPC model for HTTP-based transport. It exists because HTTP procedures are a common, practical baseline that many projects need. This does not mean VDL can only model RPC over HTTP.

VDL remains transport-agnostic at the language level. Teams can define their own RPC conventions with normal VDL types, annotations, and plugins when they need another transport or a custom protocol shape. Those custom models are valid VDL workflows, but they are not VPP-2 compatible unless they preserve the HTTP behavior defined here. Future transports (for example stdin/stdout) will define their own service annotation and their own VPP; they may reuse the generic `@proc` operation shape.

## JSON Transport Model

VPP-2 uses JSON as the standard payload format.

VDL is designed so its data model can be represented as JSON: objects, arrays, maps, strings, numbers, booleans, nullability, constants, enums, and structured inline types all have natural JSON representations. Because of that, JSON is the default wire format for RPC inputs, outputs, and errors.

Compatible plugins may map VDL types into richer language-native types in generated code, but the HTTP payloads defined by this protocol are JSON payloads.

A VPP-2 plugin only processes `@proc` members inside `@httpRpc` service containers. It silently ignores everything without `@httpRpc`, including orphan `@proc` members, with no warnings. Those shapes stay free for ad-hoc workflows or future transports. This keeps VPPs decoupled.

## Annotation Vocabulary

VPP-2 defines two annotations.

1. `@httpRpc` marks an HTTP-RPC service container.
2. `@proc` marks a request-response operation. `@proc` is intentionally generic so future RPC transports can reuse the same operation shape with their own service annotation.

### `@httpRpc`

The `@httpRpc` annotation is attached to a `type` declaration.

That type becomes an HTTP-RPC service container. The name of the annotated type is the **RPC name** used by the transport protocol.

```vdl
@httpRpc
type Users {
  // operations go here
}
```

`@httpRpc` takes no arguments. A VPP-2 compatible plugin must reject `@httpRpc` with any argument. Arguments are reserved for future protocol revisions.

### `@proc`

The `@proc` annotation is attached to a member inside an `@httpRpc` service container.

That member becomes a procedure. The member name is the **procedure name** used by the transport protocol.

Procedures use request-response HTTP behavior.

`@proc` takes no arguments. A VPP-2 compatible plugin must reject `@proc` with any argument. Arguments are reserved for future protocol revisions.

The `@proc` member itself must be an inline object. A VPP-2 compatible plugin must reject `@proc` on anything else with an error.

## Schema Shape

RPC services are modeled with normal VDL type declarations and inline object members.

```vdl
@httpRpc
type Users {
  @proc
  getUser {
    input {
      userId string
    }
    output {
      id string
      email string
    }
  }
}
```

The RPC model uses these structural rules.

1. `@httpRpc` is attached to a `type` declaration.
2. `@proc` is attached to members inside the `@httpRpc` type.
3. Operation members are regular fields whose type is an inline object.
4. `input` and `output` are regular field names inside the operation object.
5. When present, `input` and `output` must be inline object types.

Because operation bodies, `input`, and `output` are normal VDL object shapes, they can use standard language features supported by object fields, including docstrings and spreads.

## Optional Input and Output

Both `input` and `output` are optional.

Some operations only need to trigger work. Some only need to return data. Some only need to establish a transactional command without input or output data. VPP-2 supports those cases directly.

```vdl
@httpRpc
type Maintenance {
  @proc
  ping {}

  @proc
  rotateKeys {
    output {
      rotatedAt datetime
    }
  }

  @proc
  clearCache {
    input {
      namespace string
    }
  }
}
```

If `input` is omitted, the operation does not require an application-level input payload.

If `output` is omitted, a successful operation does not return application-level output data.

When either field is present, it must be an inline object. A VPP-2 compatible plugin must reject any other shape with an error. Plugins must not reinterpret scalar fields, arrays, maps, or references named `input` or `output` as valid VPP-2 operation payload definitions.

## Input Validation Scope

VPP-2 requires structural input validation.

A compatible plugin must ensure that incoming input matches the shape defined by the user. Required fields must be present, optional fields may be omitted, and every provided value must match the VDL type declared for that field.

This validation is part of the RPC contract. A VPP-2 compatible implementation must not pass invalid input to user handlers as if it were valid.

VPP-2 does not define application-specific business validation. Rules such as "this user can access this account", "this amount must be lower than the current balance", "this transition is allowed only for administrators", or "this date must be inside an active subscription period" belong to application code, middleware, another protocol, or user-defined validation layers.

In short: VPP-2 guarantees that the input has the structure and types declared in VDL. It does not decide whether that input is meaningful for a specific business domain.

## Service and Operation Names

The HTTP route is derived from VDL names. Plugins must apply `lowerFirst` to the `RPCName` segment (`lowerFirst` lowercases the first Unicode letter and leaves the rest untouched).

1. `RPCName` is the name of the `type` annotated with `@httpRpc`.
2. `ProcedureName` is the name of the member annotated with `@proc`.

Procedures use this URL structure.

```text
<baseURL>/<lowerFirst(RPCName)>/<ProcedureName>
```

`ProcedureName` is already `camelCase` in valid schemas, so it is used verbatim.

Name casing is enforced by the VDL analyzer before VPP-2 runs, so VPP-2 assumes valid names and never redefines them. Join rule: strip trailing slashes from `baseURL`, then join as `baseURL + "/" + lowerFirst(RPCName) + "/" + ProcedureName`. An empty `baseURL` yields `/<lowerFirst(RPCName)>/<ProcedureName>`.

For example, a `getUser` procedure inside an `@httpRpc type Users` service may be called at this URL when the configured base URL is `https://api.example.com/rpc`.

```text
https://api.example.com/rpc/users/getUser
```

## Transport Overview

VPP-2 uses HTTP as the standard transport.

Procedures send a JSON request body and receive a single JSON response body.

Application-level success or failure is represented by envelopes (the JSON response).

When the server successfully processes the RPC transport flow, the HTTP status is `200 OK` even if the application result is an error represented by `ok: false`.

Non-`200` HTTP status codes are reserved for transport-level failures or situations where the RPC transport itself could not be completed. Examples include connection errors, timeouts, malformed transport requests, unavailable servers, framework-level failures, or infrastructure errors. Generated clients must account for this distinction: `ok: false` is an application error envelope, while a non-`200` status is a transport failure.

Middleware may run before, during, or after operation handling. Authentication, authorization, validation, logging, metrics, tracing, rate limiting, and similar concerns may be implemented by generated or user-provided middleware. Middleware must not change the wire contract defined by this protocol.

## Envelope Model

VPP-2 represents application-level operation results with JSON envelopes.

A success envelope has `ok: true` and may include `output` when the operation defines output data.

```json
{
  "ok": true,
  "output": {
    "id": "user-123",
    "email": "john.doe@example.com"
  }
}
```

A failure envelope has `ok: false` and includes an `error` object.

```json
{
  "ok": false,
  "error": {
    "message": "User not found.",
    "category": "NotFound",
    "code": "USER_NOT_FOUND",
    "details": {
      "userId": "user-123"
    }
  }
}
```

The `error` object uses these fields.

1. `message` (`string`): human-readable description.
2. `category` (`string`, optional): error class.
3. `code` (`string`, optional): machine-readable code.
4. `details` (`object`, optional): structured context.

The envelope is the HTTP response body.

## Procedures

Procedures represent request-response RPC operations.

A procedure starts when application code invokes a generated procedure client function. The generated client builds an HTTP request, sends JSON input when input exists, and decodes one JSON response envelope.

### HTTP Request

The client sends an HTTP `POST` request.

1. **Method:** `POST`.
2. **URL structure:** `<baseURL>/<lowerFirst(RPCName)>/<ProcedureName>` joined with the canonical rule above.
3. **Headers:** always `Content-Type: application/json` and `Accept: application/json`.
4. **Body:** always a JSON object. The encoded `input` when defined, or `{}` when the operation defines no `input`.

Example request body:

```json
{
  "userId": "user-123"
}
```

Empty input body:

```json
{}
```

Servers must also accept the fallback form with no body and no `Content-Type` as empty input. Clients always send the symmetric form.

### Server Handling

The server processes a procedure request in three stages.

1. **Routing:** the URL path is mapped to an RPC service and procedure handler.
2. **Deserialization and validation:** the JSON body is decoded and validated against the operation input when input exists.
3. **Handler execution:** the user-defined handler executes with validated input, or without input when the operation does not define input.

Middleware hooks may run custom logic such as authentication, authorization, validation, logging, metrics, tracing, or rate limiting.

### HTTP Response

The server returns one HTTP response.

1. **Status code:** `200 OK` for successfully processed requests, even when `ok` is `false` and the envelope contains an application error.
2. **Headers:** `Content-Type: application/json`.
3. **Body:** a success or failure envelope.

Application-level success or failure is represented by the response envelope. Clients must inspect the envelope to distinguish successful output from application errors; they must not treat HTTP `200 OK` as application success by itself.

Success envelope:

```json
{
  "ok": true,
  "output": {
    "id": "user-123",
    "email": "john.doe@example.com"
  }
}
```

Failure envelope:

```json
{
  "ok": false,
  "error": {
    "message": "User not found.",
    "category": "NotFound",
    "code": "USER_NOT_FOUND",
    "details": {
      "userId": "user-123"
    }
  }
}
```

If the procedure does not define `output`, a successful response is exactly `{"ok":true}` with no `output` key.

### Client Handling

The generated client decodes the envelope and returns output or error to application code.

1. It parses the JSON response body.
2. If `ok` is `true`, it returns the decoded output when output exists.
3. If `ok` is `false`, it returns the decoded error.
4. Transport failures may use retry policies such as exponential backoff, depending on client configuration.

## Required Behavior for Compatible Plugins

A plugin may call itself VPP-2 compatible only if it follows these rules.

1. It must recognize `@httpRpc` and `@proc`.
2. It must treat an `@httpRpc` type as an HTTP-RPC service container.
3. It must treat `@proc` members inside `@httpRpc` containers as request-response HTTP procedures.
4. It must silently ignore everything without `@httpRpc`, with no warnings.
5. It must reject `@httpRpc` with any argument.
6. It must reject `@proc` with any argument.
7. It must reject `@proc` on anything other than an inline object.
8. It must reject non-inline-object `input` or `output` when present.
9. It must derive route names from the VDL service and operation names, applying `lowerFirst` to the `RPCName` segment and using operation names verbatim.
10. It must use the canonical URL join rule defined by this protocol.
11. It must always send a JSON object body with `Content-Type: application/json`, using `{}` when no `input` exists.
12. It must accept missing body and missing `Content-Type` as empty input on the server.
13. It must respond exactly `{"ok":true}` with no `output` key when no `output` exists.
14. It must validate operation input against the VDL-declared shape before handler execution.
15. It must decode operation output and errors through the envelope model.
16. It must preserve the distinction between procedure responses, application errors, and transport failures.
17. It must return HTTP `200 OK` for successfully processed procedure requests, including application errors represented by `ok: false`.
18. It must not reinterpret VPP-2 procedures as `GET`, query-parameter calls, or non-envelope responses while claiming VPP-2 wire compatibility.

Plugins may provide additional framework integrations, generated helpers, runtime adapters, middleware APIs, testing utilities, or documentation output. Those additions must not change the protocol semantics described here.

## Recommended Output Behavior

VPP-2 compatible plugins should generate target-idiomatic artifacts while preserving the same RPC model.

Depending on the target, useful outputs may include typed clients, server interfaces, handler stubs, route registration code, input and output models, documentation, mock servers, integration tests, or protocol manifests.

Plugins should also preserve relevant cross-protocol metadata when it applies. For example, if a VPP-2 operation is marked with a global deprecation annotation, generated clients, handlers, and documentation should surface that lifecycle information where the target ecosystem supports it.

Existing plugin examples:

- https://github.com/varavelio/vdl-plugin-http-rpc-go
- https://github.com/varavelio/vdl-plugin-http-rpc-ts
- https://github.com/varavelio/vdl-plugin-http-rpc-openapi

## Non-Goals

VPP-2 intentionally does not define every part of an application stack.

This protocol does not define authentication models, authorization policies, business-specific validation rules, retry policy defaults, handler business logic, framework-specific router internals, database access, persistence, OpenAPI output shape, binary transports, WebSocket transports, server-pushed streams, message broker behavior, or custom non-HTTP RPC transports.

Other transports (for example stdin/stdout) are out of scope. They will define their own service annotation and their own VPP, and may reuse the generic `@proc` operation shape.

Those concerns may be handled by other protocols, plugin configuration, generated framework integrations, or user code.

## Conformance Vectors

A VPP-2 compatible implementation must satisfy these vectors (names are just examples):

1. `ping {}` sends `{}` and receives `{"ok":true}`.
2. `clearCache {input}` sends `{"namespace":"..."}` and receives `{"ok":true}`.
3. `rotateKeys {output}` sends `{}` and receives `{"ok":true,"output":{"rotatedAt":"..."}}`.
4. `getUser` success returns `{"ok":true,"output":{...}}`.
5. `getUser` application error returns HTTP `200` with `{"ok":false,"error":{...}}`.
6. Transport failures use non-`200` statuses and are never confused with `ok:false`.
7. `ping` with no body and no `Content-Type` is accepted as empty input.

## Notes for Plugin Authors

> Remember that [the SDK](https://vdl-plugin-sdk.varavel.com) is here to help you, it includes many utilities for developing VDL plugins.

Validate the IR before you generate anything. The SDK `assertValidIrForHttpRpc` helper function checks service containers, procedure members, and input/output shapes, and reports problems with source positions your users can act on.

```ts
import { definePlugin } from "@varavel/vdl-plugin-sdk";
import { assertValidIrForHttpRpc } from "@varavel/vdl-plugin-sdk/utils/rpc";

export const generate = definePlugin((input) => {
  assertValidIrForHttpRpc(input.ir);
  // The IR matches VPP-2 from here on.
  // Build your clients, handlers, or docs next.
});
```

Run validation first, then emit code. That keeps route derivation, envelope handling, and error reporting consistent with the contract above. The full helper reference lives in [RPC](https://vdl-plugin-sdk.varavel.com/docs/api/rpc/).

## Compatibility Statement

VPP-2 is the standard HTTP-RPC contract for VDL plugins.

A plugin may call itself **VPP-2 compatible** only if it preserves the annotation vocabulary, schema shape, optional input/output behavior, JSON payload model, HTTP procedure behavior, and envelope semantics defined in this document.
