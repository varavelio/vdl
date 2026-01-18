package strutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEscapeQuotes(t *testing.T) {
	require.Equal(t, EscapeQuotes(`"Hello, world!"`), `\"Hello, world!\"`)
	require.Equal(t, EscapeQuotes(`Hello, world!`), `Hello, world!`)
	require.Equal(t, EscapeQuotes(`Hello, "world"!`), `Hello, \"world\"!`)
	require.Equal(t, EscapeQuotes(`Hello, "world"!"`), `Hello, \"world\"!\"`)
	require.Equal(t, EscapeQuotes(`Hello, \world!`), `Hello, \\world!`)
	require.Equal(t, EscapeQuotes(`Hello, \"world"!`), `Hello, \\\"world\"!`)
}
