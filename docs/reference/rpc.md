---
title: RPC Specification
description: Modeling RPC services with annotations and full request lifecycle behavior
---

## Overview

RPC services in VDL are modeled with annotations on core language declarations.

`@rpc` marks a service container on a `type` declaration. `@proc` and `@stream` mark service operations on type members.

## Schema Modeling

```vdl
@rpc
type Users {
  @proc
  GetUser {
    input {
      userId string
    }
    output {
      id string
      email string
    }
  }

  @stream
  UserEvents {
    input {
      userId string
    }
    output {
      event string
      at datetime
    }
  }
}
```

The RPC model uses these structural rules.

- `@rpc` is attached to a `type` declaration.
- `@proc` and `@stream` are attached to members inside the `@rpc` type.
- Operation members are regular fields with inline object types.
- `input` and `output` are regular field names whose types are inline objects.

Because `input` and `output` are inline object fields, they support standard language features, including spreads and docstrings.

## Request Lifecycle

This section defines the end-to-end data flow for both procedure calls and stream subscriptions. It covers URL structure, JSON payloads, connection behavior, and error handling. The behavior is language-agnostic and applies to official generators.

## Procedures (Request-Response)

Procedures follow a request-response model over HTTP and complete within a single HTTP transaction.

### 1. Client Invocation

Application code invokes a generated procedure client function. The generated client builds the HTTP request.

### 2. HTTP Request

The client sends an HTTP `POST` request.

- **Method:** `POST`
- **URL Structure:** `<baseURL>/<RPCName>/<ProcedureName>`
  - Example: `https://api.example.com/rpc/Users/GetUser`
- **Headers:**
  - `Content-Type: application/json`
  - `Accept: application/json`
- **Body:** JSON-encoded procedure input.

  ```json
  {
    "userId": "user-123"
  }
  ```

### 3. Server Handling

The server processes a procedure request in three stages.

1. **Routing:** The URL path is mapped to an RPC service and procedure handler.
2. **Deserialization and Validation:** The JSON body is decoded and validated.
3. **Handler Execution:** The user-defined handler executes with validated input.

VDL middleware hooks can run custom logic such as authentication, authorization, validation, logging, and metrics.

### 4. Server Response

The handler returns either output data or error data. The server serializes the result into a JSON envelope.

### 5. HTTP Response

The server returns one HTTP response.

- **Status Code:** `200 OK` for successfully processed requests. Application-level success or failure is represented by the response envelope.
- **Headers:** `Content-Type: application/json`
- **Body:**

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

`error` fields:

- `message` (`string`): human-readable description
- `category` (`string`, optional): error class
- `code` (`string`, optional): machine-readable code
- `details` (`object`, optional): structured context

### 6. Client Handling

The generated client decodes the envelope and returns output or error to application code.

1. **Deserialization:** parse JSON body.
2. **Envelope handling:**
   - `ok: true` returns `output`.
   - `ok: false` returns `error`.
3. **Resilience:** transport failures can use retry policies such as exponential backoff, depending on client configuration.

## Streams (Server-Sent Events)

Streams use Server-Sent Events (SSE) with a persistent connection so servers can push multiple events over time.

### 1. Client Subscription

Application code subscribes through the generated stream client.

### 2. HTTP Request (Connection)

The client opens the stream with a `POST` request.

- **Method:** `POST`
- **URL Structure:** `<baseURL>/<RPCName>/<StreamName>`
  - Example: `https://api.example.com/rpc/Chat/NewMessage`
- **Headers:**
  - `Accept: text/event-stream`
  - `Content-Type: application/json`
- **Body:** JSON-encoded stream input.

  ```json
  {
    "chatId": "room-42"
  }
  ```

### 3. Server Handling (Connection)

The server validates input and initializes the SSE channel.

1. **Validation:** invalid input returns a single JSON error response.
2. **Connection setup:**
   - Status `200 OK`
   - `Content-Type: text/event-stream`
   - `Cache-Control: no-cache`
   - `Connection: keep-alive`
3. **Handler execution:** stream handler runs with an `emit` function.

Middleware hooks are available for cross-cutting behavior in stream flows.

### 4. Event Emission

Handlers call `emit` to push events. Each emitted event is serialized and written as SSE `data`.

- JSON payloads must be single-line for SSE framing safety.
- Newline-sensitive formatting is handled before transmission.

Success event:

```
data: {"ok":true,"output":{"messageId":"msg-abc","text":"Hello world!"}}
```

Error event:

```
data: {"ok":false,"error":{"message":"You do not have permission to view this chat."}}
```

### 5. Keep-Alive (Ping Events)

Servers send periodic SSE comments to keep connections alive.

- **Format:** `: ping\n\n`
- **Interval:** server-configurable (commonly 30 seconds)
- **Client behavior:** ignored by generated clients and not surfaced to application code

### 6. Client Handling (Event Loop)

The client processes stream events continuously.

1. Parse SSE frames.
2. Ignore comment ping frames.
3. Deserialize JSON payloads.
4. Deliver `output` or `error` envelopes to application code.

### 7. Stream Termination

A stream can end by client cancellation, handler completion, or network interruption.

Generated clients may reconnect automatically with configurable retry behavior, re-sending the original subscription request when applicable.
