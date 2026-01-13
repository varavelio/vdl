---
title: Request Lifecycle
description: Lifecycle of a single request in UFO RPC
---

This document outlines the end-to-end data flow for both procedure calls and stream subscriptions in UFO RPC. It details the process from the client's initial request to the server's final response, including URL structure, JSON payloads, and error handling. This specification is language-agnostic and applies to all official UFO RPC code generators.

---

## Procedures (Request-Response)

Procedures follow a standard request-response model over HTTP. The entire lifecycle is completed within a single HTTP transaction.

### 1. Client Invocation

The developer uses the generated client to call a procedure (e.g., `CreateUser`). The client library is responsible for constructing the HTTP request.

### 2. HTTP Request

The client sends an HTTP `POST` request to the server.

- **Method:** `POST`
- **URL Structure:** The URL is formed by appending the procedure name to the base URL.
  - Format: `<baseURL>/<ProcedureName>`
  - Example: `https://api.example.com/urpc/CreateUser`
- **Headers:**
  - `Content-Type: application/json`
  - `Accept: application/json`
- **Body:** The request body contains the JSON-encoded input for the procedure.
  ```json
  {
    "name": "John Doe",
    "email": "john.doe@example.com"
  }
  ```

### 3. Server Handling

The server receives the request and performs the following steps:

1.  **Routing:** It maps the URL path (`/CreateUser`) to the corresponding procedure handler.
2.  **Deserialization & Validation:** It decodes the JSON body and performs built-in validation (e.g., checking for required fields). If this fails, it immediately responds with a validation error.
3.  **Handler Execution:** The server invokes the user-defined business logic for the procedure, passing the validated input.

> **Note:** UFO RPC provides a hook system that allows developers to run custom code at various points in the lifecycle for tasks like authentication, custom input validation, logging, metrics, etc.

### 4. Server Response

The user's handler returns data on success or error information on failure. The server then constructs the final JSON payload and serializes it to a string.

### 5. HTTP Response

The server sends a single HTTP response back to the client.

- **Status Code:** `200 OK`. The HTTP status is always 200 for successfully processed requests, even if the application logic resulted in an error. The success or failure is indicated within the JSON payload via the `ok` field.
- **Headers:** `Content-Type: application/json`
- **Body:** A JSON payload containing the result of the operation.

  **On Success:**

  ```json
  {
    "ok": true,
    "output": {
      "userId": "user-123",
      "status": "created"
    }
  }
  ```

  **On Failure:**

  ```json
  {
    "ok": false,
    "error": {
      "message": "A user with this email already exists.",
      "category": "ValidationError",
      "code": "EMAIL_ALREADY_EXISTS",
      "details": {
        "field": "email"
      }
    }
  }
  ```

### 6. Client Handling

The client library receives the HTTP response.

1.  **Deserialization:** It decodes the JSON body.
2.  **Result Unwrapping:**
    - If `ok` is `true`, it returns the content of the `output` field to the application code.
    - If `ok` is `false`, it returns the content of the `error` field.
3.  **Resilience:** For transport-level failures or 5xx server errors, the client automatically handles retries with exponential backoff according to its configuration.

---

## Streams (Server-Sent Events)

Streams use Server-Sent Events (SSE) to maintain a persistent connection, allowing the server to push multiple events to the client over time.

### 1. Client Subscription

The developer subscribes to a stream (e.g., `NewMessage`) using the generated client.

### 2. HTTP Request (Connection)

The client initiates the connection with a single HTTP `POST` request.

- **Method:** `POST`
- **URL Structure:** `<baseURL>/<StreamName>`
  - Example: `https://api.example.com/urpc/NewMessage`
- **Headers:**
  - `Accept: text/event-stream`
  - `Content-Type: application/json`
- **Body:** The JSON-encoded input for the stream subscription.
  ```json
  {
    "chatId": "room-42"
  }
  ```

### 3. Server Handling (Connection)

The server receives the request and establishes the persistent connection.

1.  **Validation:** It validates the input just like a procedure. An error here terminates the connection attempt with a single JSON error response.
2.  **Connection Upgrade:** If validation passes, the server sends back HTTP headers to establish the SSE stream. The connection is now open and long-lived.
    - **Status Code:** `200 OK`
    - **Headers:**
      - `Content-Type: text/event-stream`
      - `Cache-Control: no-cache`
      - `Connection: keep-alive`
3.  **Handler Execution:** The server invokes the user-defined stream handler, providing it with an `emit` function.

> **Note:** Just like with procedures, a hook system is available for streams to run custom code for authentication, validation, logging, and other cross-cutting concerns.

### 4. Event Emission

The server-side handler logic can now call the `emit` function at any time to push data to the client.

- **SSE Formatting:** The server formats the output into a standard JSON payload and sends it as an SSE `data` event.
- **Data Transmission:** The formatted event is written to the open HTTP connection.

  **Success Event:**

  ```
  data: {"ok":true,"output":{"messageId":"msg-abc","text":"Hello world!"}}

  ```

  _(Note the required blank line after the data line)_

  **Error Event (for stream-specific errors):**

  ```
  data: {"ok":false,"error":{"message":"You do not have permission to view this chat."}}

  ```

### 5. Client Handling (Event Loop)

The client library maintains the open connection and listens for incoming events.

1.  **Event Parsing:** As data arrives, the client parses the SSE `data:` payload.
2.  **Deserialization:** It decodes the JSON from the data field.
3.  **Delivery:** It delivers the content of the `output` or `error` field to the application code, typically through a channel or callback.

### 6. Stream Termination

The connection can be closed in several ways:

- **Client-side:** The developer cancels the context, which closes the connection.
- **Server-side:** The stream handler function returns, signaling the end of the stream.
- **Network Error:** The connection is lost.

**Resilience:** If the connection is lost unexpectedly, the client automatically attempts to reconnect with exponential backoff, re-submitting the initial request.
