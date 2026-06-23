+++
title = "VPP-2: RPC"
description = "Standard RPC modeling and HTTP/SSE transport behavior for VDL plugins."
template = "docs.html"
path = "docs/protocols/rpc"
weight = 4
+++

**VPP-2: RPC** defines the standard way to model remote procedure calls in VDL and the standard HTTP transport behavior that compatible plugins must implement.

This protocol covers two operation styles.

1. **Procedures:** request-response operations completed within a single HTTP transaction.
2. **Streams:** server-pushed event flows delivered over Server-Sent Events (SSE).

VPP-2 does not add syntax to VDL. It uses normal VDL declarations, inline object types, and annotations. The purpose of the protocol is to give plugin authors a shared contract for generating clients, servers, handlers, SDKs, documentation, adapters, and test tools from the same RPC model.

If a plugin claims compatibility with VPP-2, it must preserve the annotation semantics, operation model, HTTP behavior, SSE behavior, and envelope format defined in this document.

## Why This Protocol Exists

RPC systems often look different across languages, frameworks, and runtimes. One stack may generate controller classes, another may generate function handlers, another may generate typed clients, and another may generate route metadata.

VDL should not hard-code any of those framework choices into the compiler. The compiler should keep producing a language-neutral representation of schemas and annotations. Plugins can then turn that representation into idiomatic code for each target ecosystem.

VPP-2 exists so those plugins still agree on the important parts: how RPC services are modeled, how operations are named, how input and output are represented, how HTTP requests are formed, how errors are returned, and how streams behave over SSE.

The result is a portable RPC contract. A VDL service should feel familiar whether the generated target is Go, TypeScript, Rust, Java, or an internal platform plugin.

VPP-2 is specifically the standard RPC model for HTTP-based transport. It exists because HTTP procedures and SSE streams are a common, practical baseline that many projects need. This does not mean VDL can only model RPC over HTTP.

VDL remains transport-agnostic at the language level. Teams can define their own RPC conventions with normal VDL types, annotations, and plugins when they need another transport or a custom protocol shape. Those custom models are valid VDL workflows, but they are not VPP-2 compatible unless they preserve the HTTP and SSE behavior defined here.

## JSON Transport Model

VPP-2 uses JSON as the standard payload format.

VDL is designed so its data model can be represented as JSON: objects, arrays, maps, strings, numbers, booleans, nullability, constants, enums, and structured inline types all have natural JSON representations. Because of that, JSON is the default wire format for RPC inputs, outputs, and errors.

Compatible plugins may map VDL types into richer language-native types in generated code, but the HTTP payloads defined by this protocol are JSON payloads.

## Annotation Vocabulary

VPP-2 defines three annotations.

1. `@rpc` marks a service container.
2. `@proc` marks a request-response operation.
3. `@stream` marks a server-pushed stream operation.

### `@rpc`

The `@rpc` annotation is attached to a `type` declaration.

That type becomes an RPC service container. The name of the annotated type is the **RPC name** used by the transport protocol.

```vdl
@rpc
type Users {
  // operations go here
}
```

### `@proc`

The `@proc` annotation is attached to a member inside an `@rpc` service container.

That member becomes a procedure. The member name is the **procedure name** used by the transport protocol.

Procedures use request-response HTTP behavior.

### `@stream`

The `@stream` annotation is attached to a member inside an `@rpc` service container.

That member becomes a stream. The member name is the **stream name** used by the transport protocol.

Streams use HTTP to open the subscription and SSE to receive emitted events.

## Schema Shape

RPC services are modeled with normal VDL type declarations and inline object members.

```vdl
@rpc
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

  @stream
  userEvents {
    input {
      userId string
    }
    output {
      eventName string
    }
  }
}
```

The RPC model uses these structural rules.

1. `@rpc` is attached to a `type` declaration.
2. `@proc` and `@stream` are attached to members inside the `@rpc` type.
3. Operation members are regular fields whose type is an inline object.
4. `input` and `output` are regular field names inside the operation object.
5. When present, `input` and `output` must be inline object types.

Because operation bodies, `input`, and `output` are normal VDL object shapes, they can use standard language features supported by object fields, including docstrings and spreads.

## Optional Input and Output

Both `input` and `output` are optional.

Some operations only need to trigger work. Some only need to return data. Some only need to establish a transactional command without input or output data. VPP-2 supports those cases directly.

```vdl
@rpc
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

When either field is present, it must be an inline object. Plugins must not reinterpret scalar fields, arrays, maps, or references named `input` or `output` as valid VPP-2 operation payload definitions.

## Input Validation Scope

VPP-2 requires structural input validation.

A compatible plugin must ensure that incoming input matches the shape defined by the user. Required fields must be present, optional fields may be omitted, and every provided value must match the VDL type declared for that field.

This validation is part of the RPC contract. A VPP-2 compatible implementation must not pass invalid input to user handlers as if it were valid.

VPP-2 does not define application-specific business validation. Rules such as "this user can access this account", "this amount must be lower than the current balance", "this transition is allowed only for administrators", or "this date must be inside an active subscription period" belong to application code, middleware, another protocol, or user-defined validation layers.

In short: VPP-2 guarantees that the input has the structure and types declared in VDL. It does not decide whether that input is meaningful for a specific business domain.

## Service and Operation Names

The HTTP route is derived from VDL names.

1. `RPCName` is the name of the `type` annotated with `@rpc`.
2. `ProcedureName` is the name of the member annotated with `@proc`.
3. `StreamName` is the name of the member annotated with `@stream`.

Procedures use this URL structure.

```text
<baseURL>/<RPCName>/<ProcedureName>
```

Streams use this URL structure.

```text
<baseURL>/<RPCName>/<StreamName>
```

For example, a `getUser` procedure inside an `@rpc type Users` service may be called at this URL when the configured base URL is `https://api.example.com/rpc`.

```text
https://api.example.com/rpc/Users/getUser
```

## Transport Overview

VPP-2 uses HTTP as the standard transport.

Procedures send a JSON request body and receive a single JSON response body.

Streams send a JSON request body to open the subscription and receive a persistent SSE response where each emitted event contains a JSON envelope.

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

For procedures, the envelope is the HTTP response body.

For streams, each SSE `data:` payload is a JSON envelope. Stream error events are normal SSE events whose envelope has `ok: false`.

## Procedures

Procedures represent request-response RPC operations.

A procedure starts when application code invokes a generated procedure client function. The generated client builds an HTTP request, sends JSON input when input exists, and decodes one JSON response envelope.

### HTTP Request

The client sends an HTTP `POST` request.

1. **Method:** `POST`.
2. **URL structure:** `<baseURL>/<RPCName>/<ProcedureName>`.
3. **Headers:** `Content-Type: application/json` and `Accept: application/json`.
4. **Body:** JSON-encoded procedure input when `input` is defined.

Example request body:

```json
{
  "userId": "user-123"
}
```

If the procedure does not define `input`, the operation has no application-level input payload. Generated clients and servers should still preserve the same HTTP method, URL, and response semantics.

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

If the procedure does not define `output`, a successful response still uses the success envelope, but it does not need to include application-level output data.

### Client Handling

The generated client decodes the envelope and returns output or error to application code.

1. It parses the JSON response body.
2. If `ok` is `true`, it returns the decoded output when output exists.
3. If `ok` is `false`, it returns the decoded error.
4. Transport failures may use retry policies such as exponential backoff, depending on client configuration.

## Streams

Streams represent server-pushed RPC operations.

A stream starts when application code subscribes through a generated stream client. The generated client opens an HTTP request with JSON input when input exists. If the server accepts the subscription, the response becomes a persistent SSE connection.

The stream handler may emit multiple output or error envelopes over time.

### HTTP Request

The client opens the stream with an HTTP `POST` request.

1. **Method:** `POST`.
2. **URL structure:** `<baseURL>/<RPCName>/<StreamName>`.
3. **Headers:** `Accept: text/event-stream` and `Content-Type: application/json`.
4. **Body:** JSON-encoded stream input when `input` is defined.

Example request body:

```json
{
  "chatId": "room-42"
}
```

If the stream does not define `input`, the subscription does not require an application-level input payload.

### Server Handling

The server validates input and initializes the SSE channel.

1. **Validation:** invalid input returns a single JSON error response.
2. **Connection setup:** accepted streams return `200 OK`, `Content-Type: text/event-stream`, `Cache-Control: no-cache`, and `Connection: keep-alive`.
3. **Handler execution:** the stream handler runs with an `emit` function.

Middleware hooks are available for cross-cutting behavior in stream flows.

### Event Emission

Handlers call `emit` to push events.

Each emitted event is serialized as a JSON envelope and written as SSE `data`.

An emitted stream error is also an event. It uses the same SSE framing as any other emitted item, but its JSON envelope has `ok: false` and an `error` object instead of successful output.

JSON payloads must be single-line for SSE framing safety. Newline-sensitive formatting must be handled before transmission.

Success event:

```text
data: {"ok":true,"output":{"messageId":"msg-abc","text":"Hello world!"}}
```

Error event:

```text
data: {"ok":false,"error":{"message":"You do not have permission to view this chat."}}
```

If the stream does not define `output`, emitted success envelopes do not need to include application-level output data.

### Keep-Alive Events

Servers send periodic SSE comments to keep connections alive.

1. **Format:** `: ping\n\n`.
2. **Interval:** server-configurable, commonly 30 seconds.
3. **Client behavior:** generated clients ignore ping comments and do not surface them to application code.

### Client Event Loop

The generated client processes stream events continuously.

1. It parses SSE frames.
2. It ignores comment ping frames.
3. It deserializes JSON payloads.
4. It delivers output or error envelopes to application code. Stream errors are delivered from `ok: false` event envelopes; they are not the same thing as transport failures.

### Stream Termination

A stream can end by client cancellation, handler completion, or network interruption.

Generated clients may reconnect automatically with configurable retry behavior, re-sending the original subscription request when applicable.

## Required Behavior for Compatible Plugins

A plugin may call itself VPP-2 compatible only if it follows these rules.

1. It must recognize `@rpc`, `@proc`, and `@stream`.
2. It must treat an `@rpc` type as an RPC service container.
3. It must treat `@proc` members as request-response HTTP procedures.
4. It must treat `@stream` members as SSE stream subscriptions.
5. It must derive route names from the VDL service and operation names.
6. It must use the URL structures defined by this protocol.
7. It must encode operation input as JSON when input exists.
8. It must validate operation input against the VDL-declared shape before handler execution.
9. It must decode operation output and errors through the envelope model.
10. It must preserve the optional nature of `input` and `output`.
11. It must preserve the distinction between procedure responses, stream events, application errors, and transport failures.
12. It must return HTTP `200 OK` for successfully processed procedure requests, including application errors represented by `ok: false`.
13. It must not reinterpret VPP-2 streams as WebSockets, polling, or another transport while claiming VPP-2 wire compatibility.
14. It must not reinterpret VPP-2 procedures as `GET`, query-parameter calls, or non-envelope responses while claiming VPP-2 wire compatibility.

Plugins may provide additional framework integrations, generated helpers, runtime adapters, middleware APIs, testing utilities, or documentation output. Those additions must not change the protocol semantics described here.

## Recommended Output Behavior

VPP-2 compatible plugins should generate target-idiomatic artifacts while preserving the same RPC model.

Depending on the target, useful outputs may include typed clients, server interfaces, handler stubs, route registration code, input and output models, stream subscriber APIs, documentation, mock servers, integration tests, or protocol manifests.

Plugins should also preserve relevant cross-protocol metadata when it applies. For example, if a VPP-2 operation is marked with a global deprecation annotation, generated clients, handlers, and documentation should surface that lifecycle information where the target ecosystem supports it.

## Non-Goals

VPP-2 intentionally does not define every part of an application stack.

This protocol does not define authentication models, authorization policies, business-specific validation rules, retry policy defaults, handler business logic, framework-specific router internals, database access, persistence, OpenAPI output shape, binary transports, WebSocket transports, message broker behavior, or custom non-HTTP RPC transports.

Those concerns may be handled by other protocols, plugin configuration, generated framework integrations, or user code.

## Compatibility Statement

VPP-2 is the standard RPC contract for VDL plugins.

A plugin may call itself **VPP-2 compatible** only if it preserves the annotation vocabulary, schema shape, optional input/output behavior, JSON payload model, HTTP procedure behavior, SSE stream behavior, and envelope semantics defined in this document.
