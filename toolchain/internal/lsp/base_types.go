package lsp

import (
	"encoding/json"
	"fmt"
	"strconv"
)

const (
	ErrorCodeInvalidParams = -32602
)

var (
	DefaultMessage = Message{
		JSONRPC: "2.0",
	}
)

// Message is a general message as defined by JSON-RPC. The language server protocol
// always uses “2.0” as the jsonrpc version.
type Message struct {
	// The JSON-RPC version. Always "2.0".
	JSONRPC string `json:"jsonrpc"`
	// The request id. Only for messages.
	ID IntOrString `json:"id,omitempty,omitzero"`
	// The method to be invoked. For messages and notifications.
	Method string `json:"method,omitempty,omitzero"`
}

// IntOrString is a type that can be either an int or a string when unmarshalled from JSON.
type IntOrString string

// UnmarshalJSON implements the json.Unmarshaler interface.
// The value is parsed as an int if possible, otherwise it is parsed as a string.
func (i *IntOrString) UnmarshalJSON(b []byte) error {
	var n int
	if err := json.Unmarshal(b, &n); err == nil {
		*i = IntOrString(fmt.Appendf(nil, "%d", n))
		return nil
	}

	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		*i = IntOrString(s)
		return nil
	}

	return fmt.Errorf("IntOrString: %s is not an int or a string", string(b))
}

// MarshalJSON implements the json.Marshaler interface.
// If the value is an int, it is marshaled as an int. Otherwise, it is marshaled as a string.
func (i IntOrString) MarshalJSON() ([]byte, error) {
	if n, err := strconv.Atoi(string(i)); err == nil {
		return json.Marshal(n)
	}

	return json.Marshal(string(i))
}

// RequestMessage describes a request between the client and the server. Every processed
// request must send a response back to the sender of the request.
type RequestMessage struct {
	Message
}

// ResponseMessage sent as a result of a request.
type ResponseMessage struct {
	Message
	// The request id.
	ID IntOrString `json:"id"`
	// The error object.
	Error ResponseError `json:"error,omitzero"`
}

// ResponseError is an error that occurred while processing a request.
type ResponseError struct {
	// A number indicating the error type that occurred.
	Code int `json:"code"`
	// A string providing a short description of the error.
	Message string `json:"message"`
	// Additional error data. Can be omitted.
	Data map[string]any `json:"data"`
}

// NotificationMessage describes a notification between the client and the server.
type NotificationMessage struct {
	Message
}
