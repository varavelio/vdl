package lsp

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScannerSplitFunc(t *testing.T) {
	t.Run("full message", func(t *testing.T) {
		encoded := []byte("Content-Length: 26\r\n\r\n{\"jsonrpc\":\"2.0\",\"id\":\"1\"}")
		advance, token, err := scannerSplitFunc(encoded, false)
		require.NoError(t, err)
		require.Equal(t, len(encoded), advance)
		require.Equal(t, []byte("{\"jsonrpc\":\"2.0\",\"id\":\"1\"}"), token)
	})

	t.Run("partial message", func(t *testing.T) {
		encoded := []byte("Content-Length: 26\r\n\r\n{")
		advance, token, err := scannerSplitFunc(encoded, false)
		require.NoError(t, err)
		require.Equal(t, 0, advance)
		require.Nil(t, token)
	})

	t.Run("invalid header", func(t *testing.T) {
		encoded := []byte("Content-Length: invalid\r\n\r\n{")
		advance, token, err := scannerSplitFunc(encoded, false)
		require.Error(t, err)
		require.Equal(t, 0, advance)
		require.Nil(t, token)
	})

	t.Run("with extra data", func(t *testing.T) {
		input := []byte("Content-Length: 26\r\n\r\n{\"jsonrpc\":\"2.0\",\"id\":\"1\"}")
		extra := []byte("extra")
		encoded := append(input, extra...)
		advance, token, err := scannerSplitFunc(encoded, false)
		require.NoError(t, err)
		require.Equal(t, len(input), advance)
		require.Equal(t, []byte("{\"jsonrpc\":\"2.0\",\"id\":\"1\"}"), token)
	})
}

func TestEncode(t *testing.T) {
	request := ResponseMessage{
		Message: DefaultMessage,
		ID:      IntOrString("1"),
	}
	encoded, err := encode(request)

	require.NoError(t, err)
	require.Equal(t, "Content-Length: 24\r\n\r\n{\"jsonrpc\":\"2.0\",\"id\":1}", string(encoded))
}

func TestDecode(t *testing.T) {
	t.Run("with header", func(t *testing.T) {
		encoded := []byte("Content-Length: 26\r\n\r\n{\"jsonrpc\":\"2.0\",\"id\":\"1\"}")
		expected := ResponseMessage{
			Message: DefaultMessage,
			ID:      IntOrString("1"),
		}

		var decoded ResponseMessage
		require.NoError(t, decode(encoded, &decoded))
		require.Equal(t, expected, decoded)
	})

	t.Run("without header", func(t *testing.T) {
		encoded := []byte("{\"jsonrpc\":\"2.0\",\"id\":\"1\"}")
		expected := ResponseMessage{
			Message: DefaultMessage,
			ID:      IntOrString("1"),
		}

		var decoded ResponseMessage
		require.NoError(t, decode(encoded, &decoded))
		require.Equal(t, expected, decoded)
	})
}
