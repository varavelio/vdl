//nolint:unused
package pieces

import (
	"encoding/json"
	"fmt"
	"io"
)

/** START FROM HERE **/

// -----------------------------------------------------------------------------
// Core Types
// -----------------------------------------------------------------------------

// Response represents the response of a UFO RPC call.
type Response[T any] struct {
	Ok     bool  `json:"ok"`
	Output T     `json:"output,omitempty,omitzero"`
	Error  Error `json:"error,omitzero"`
}

// Write writes the response as a JSON-formatted string to the given writer.
func (r Response[T]) Write(w io.Writer) error {
	return json.NewEncoder(w).Encode(r)
}

// String returns the Response as a JSON-formatted string including all its fields.
func (r Response[T]) String() string {
	b, err := json.Marshal(r)
	if err != nil {
		return fmt.Sprintf("failed to marshal UFO RPC Response: %s", err.Error())
	}
	return string(b)
}

// Bytes returns the Response as a JSON-formatted byte slice.
func (r Response[T]) Bytes() []byte {
	return []byte(r.String())
}

// Error represents a standardized error in the UFO RPC system.
//
// It provides structured information about errors that occur within the system,
// enabling consistent error handling across servers and clients.
//
// Fields:
//   - Message: A human-readable description of the error.
//   - Category: Optional. Categorizes the error by its nature or source (e.g., "ValidationError", "DatabaseError").
//   - Code: Optional. A machine-readable identifier for the specific error condition (e.g., "INVALID_EMAIL").
//   - Details: Optional. Additional information about the error.
//
// The struct implements the error interface.
type Error struct {
	// Message provides a human-readable description of the error.
	//
	// This message can be displayed to end-users or used for logging and debugging purposes.
	//
	// Use Cases:
	//   1. If localization is not implemented, Message can be directly shown to the user to inform them of the issue.
	//   2. Developers can use Message in logs to diagnose problems during development or in production.
	Message string `json:"message"`

	// Category categorizes the error by its nature or source.
	//
	// Examples:
	//   - "ValidationError" for input validation errors.
	//   - "DatabaseError" for errors originating from database operations.
	//   - "AuthenticationError" for authentication-related issues.
	//
	// Use Cases:
	//   1. In middleware, you can use Category to determine how to handle the error.
	//      For instance, you might log "InternalError" types and return a generic message to the client.
	//   2. Clients can inspect the Category to decide whether to prompt the user for action,
	//      such as re-authentication if the Category is "AuthenticationError".
	Category string `json:"category,omitzero"`

	// Code is a machine-readable identifier for the specific error condition.
	//
	// Examples:
	//   - "INVALID_EMAIL" when an email address fails validation.
	//   - "USER_NOT_FOUND" when a requested user does not exist.
	//   - "RATE_LIMIT_EXCEEDED" when a client has made too many requests.
	//
	// Use Cases:
	//   1. Clients can map Codes to localized error messages for internationalization (i18n),
	//      displaying appropriate messages based on the user's language settings.
	//   2. Clients or middleware can implement specific logic based on the Code,
	//      such as retry mechanisms for "TEMPORARY_FAILURE" or showing captcha for "RATE_LIMIT_EXCEEDED".
	Code string `json:"code,omitzero"`

	// Details contains optional additional information about the error.
	//
	// This field can include any relevant data that provides more context about the error.
	// The contents should be serializable to JSON.
	//
	// Use Cases:
	//   1. Providing field-level validation errors, e.g., Details could be:
	//      {"fields": {"email": "Email is invalid", "password": "Password is too short"}}
	//   2. Including diagnostic information such as timestamps, request IDs, or stack traces
	//      (ensure sensitive information is not exposed to clients).
	Details map[string]any `json:"details,omitempty"`
}

// Error implements the error interface, returning the error message.
func (e Error) Error() string {
	return e.Message
}

// String implements the fmt.Stringer interface, returning the error message.
func (e Error) String() string {
	return e.Message
}

// ToJSON returns the Error as a JSON-formatted string including all its fields.
// This is useful for logging and debugging purposes.
//
// Example usage:
//
//	err := Error{
//	  Category: "ValidationError",
//	  Code:     "INVALID_EMAIL",
//	  Message:  "The email address provided is invalid.",
//	  Details:  map[string]any{
//	    "field": "email",
//	  },
//	}
//	log.Println(err.ToJSON())
func (e Error) ToJSON() string {
	b, err := json.Marshal(e)
	if err != nil {
		return fmt.Sprintf(
			`{"message":%q,"error":"Failed to marshal UFO RPC Error: %s"}`,
			e.Message, err.Error(),
		)
	}
	return string(b)
}

// asError converts any error into a Error.
// If the provided error is already a Error, it returns it as is.
// Otherwise, it wraps the error message into a new Error.
//
// This function ensures that all errors conform to the Error structure,
// facilitating consistent error handling across the system.
func asError(err error) Error {
	switch e := err.(type) {
	case Error:
		return e
	case *Error:
		return *e
	default:
		return Error{
			Message: err.Error(),
		}
	}
}

// errorMissingRequiredField creates a new Error for the case
// where a required field is missing in the input.
func errorMissingRequiredField(message string) Error {
	return Error{
		Category: "ValidationError",
		Code:     "MISSING_REQUIRED_FIELD",
		Message:  message,
	}
}
