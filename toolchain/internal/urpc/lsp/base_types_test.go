package lsp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIntOrString(t *testing.T) {
	t.Run("Unmarshal int", func(t *testing.T) {
		var i IntOrString
		err := json.Unmarshal([]byte("123"), &i)
		require.NoError(t, err)
		require.Equal(t, IntOrString("123"), i)
	})

	t.Run("Unmarshal string", func(t *testing.T) {
		var i IntOrString
		err := json.Unmarshal([]byte("\"abc\""), &i)
		require.NoError(t, err)
		require.Equal(t, IntOrString("abc"), i)
	})

	t.Run("Marshal int", func(t *testing.T) {
		i := IntOrString("123")
		b, err := json.Marshal(i)
		require.NoError(t, err)
		require.Equal(t, []byte("123"), b)
	})

	t.Run("Marshal string", func(t *testing.T) {
		i := IntOrString("abc")
		b, err := json.Marshal(i)
		require.NoError(t, err)
		require.Equal(t, []byte("\"abc\""), b)
	})
}
