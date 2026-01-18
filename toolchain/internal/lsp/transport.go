package lsp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

// scannerSplitFunc is a custom split function for the scanner that splits the input
// into valid LSP JSON-RPC messages.
//
// If the input has no header it will ignore that chunk and continue reading the input.
//
// If the input has no sufficient data to form a valid LSP JSON-RPC message, it returns
// 0, nil, nil to indicate that the scanner should continue reading the input.
func scannerSplitFunc(data []byte, _ bool) (advance int, token []byte, err error) {
	if !bytes.HasPrefix(data, []byte("Content-Length: ")) {
		return len(data), nil, nil
	}

	delimiter := []byte("\r\n\r\n")
	header, content, found := bytes.Cut(data, delimiter)
	if !found {
		return 0, nil, nil
	}

	rawContentLength := bytes.TrimPrefix(header, []byte("Content-Length: "))
	rawContentLength = bytes.TrimSpace(rawContentLength)
	contentLength, err := strconv.Atoi(string(rawContentLength))
	if err != nil {
		return 0, nil, fmt.Errorf("invalid Content-Length, should be an integer: %s", err)
	}

	if len(content) < contentLength {
		return 0, nil, nil
	}
	content = content[:contentLength]

	totalLength := len(header) + len(delimiter) + len(content)
	return totalLength, content, nil
}

// encode encodes the given data into a valid LSP JSON-RPC message and returns
// the encoded message as a byte slice.
func encode(data any) ([]byte, error) {
	marshaled, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	contentLength := len(marshaled)
	content := string(marshaled)

	return fmt.Appendf(nil, "Content-Length: %d\r\n\r\n%s", contentLength, content), nil
}

// decode decodes the given data into the given value. It expects the data to be
// a valid LSP JSON-RPC message.
//
// If the data contains a header part (Content-Length: ...\r\n\r\n), it will be removed.
func decode(data []byte, v any) error {
	if bytes.HasPrefix(data, []byte("Content-Length: ")) {
		delimiter := []byte("\r\n\r\n")
		_, content, found := bytes.Cut(data, delimiter)
		if !found {
			return fmt.Errorf("invalid LSP JSON-RPC message")
		}
		data = content
	}

	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}
	return nil
}

// decodeToMap decodes the given data into a map[string]any. It expects the data to be
// a valid LSP JSON-RPC message.
//
// If the data contains a header part (Content-Length: ...\r\n\r\n), it will be removed.
func decodeToMap(data []byte) (map[string]any, error) {
	var v map[string]any
	if err := decode(data, &v); err != nil {
		return nil, fmt.Errorf("failed to decode data: %w", err)
	}
	return v, nil
}
