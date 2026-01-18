package formatter

import (
	"embed"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/uforg/uforpc/urpc/internal/util/strutil"
)

//go:embed tests/*.urpc
var testFiles embed.FS

func TestFormatEmptySchema(t *testing.T) {
	input := ""
	expected := ""

	formatted, err := Format("schema.urpc", input)

	require.NoError(t, err)
	require.Equal(t, expected, formatted)
}

func TestFormat(t *testing.T) {
	files, err := testFiles.ReadDir("tests")
	require.NoError(t, err)

	for _, file := range files {
		content, err := testFiles.ReadFile(path.Join("tests", file.Name()))
		require.NoError(t, err)

		separator := "\n// >>>>\n\n"
		input := strutil.GetStrBefore(string(content), separator)
		expected := strutil.GetStrAfter(string(content), separator)

		formatted, err := Format(file.Name(), input)
		require.NoError(t, err, "error formatting %s", file.Name())
		require.Equal(t, expected, formatted, "incorrect formatting for %s", file.Name())
	}
}

// This test is used to debug the formatter using a single file.
// Feel free to modify the prefix to test other files.
func TestFormatOnlyOne(t *testing.T) {
	filePrefix := "0100"

	files, err := testFiles.ReadDir("tests")
	require.NoError(t, err)

	for _, file := range files {
		if !strings.HasPrefix(file.Name(), filePrefix) {
			continue
		}

		content, err := testFiles.ReadFile(path.Join("tests", file.Name()))
		require.NoError(t, err)

		separator := "\n// >>>>\n\n"
		input := strutil.GetStrBefore(string(content), separator)
		expected := strutil.GetStrAfter(string(content), separator)

		formatted, err := Format(file.Name(), input)
		require.NoError(t, err, "error formatting %s", file.Name())
		require.Equal(t, expected, formatted, "incorrect formatting for %s", file.Name())
	}
}
