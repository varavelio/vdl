package analysis_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/varavelio/vdl/toolchain/internal/core/analysis"
	"github.com/varavelio/vdl/toolchain/internal/core/vfs"
)

func TestValidSchemas(t *testing.T) {
	files := globTestFiles(t, "valid/*.vdl")
	require.NotEmpty(t, files, "no valid test files found")

	for _, file := range files {
		name := filepath.Base(file)
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			program, diagnostics := analyzeTestFile(t, file)

			if len(diagnostics) > 0 {
				t.Errorf("expected no diagnostics for valid file %s, got:", name)
				for _, d := range diagnostics {
					t.Errorf("  %s", d)
				}
			}

			require.NotNil(t, program)
		})
	}
}

func TestErrorSchemas(t *testing.T) {
	files := globTestFilesRecursive(t, "errors")
	require.NotEmpty(t, files, "no error test files found")

	for _, file := range files {
		relPath := relativePath(t, file)
		t.Run(relPath, func(t *testing.T) {
			t.Parallel()

			expected := parseExpectedCodes(t, file)
			program, diagnostics := analyzeTestFile(t, file)

			require.NotNil(t, program)
			actualCodes := extractCodes(diagnostics)

			for _, expectedCode := range expected.codes {
				if !containsCode(actualCodes, expectedCode) {
					t.Errorf("expected diagnostic code %s not found in: %v", expectedCode, actualCodes)
				}
			}

			if len(diagnostics) < len(expected.codes) {
				t.Errorf("expected at least %d diagnostics, got %d: %v", len(expected.codes), len(diagnostics), actualCodes)
			}
		})
	}
}

func TestMultiFileSchemas(t *testing.T) {
	t.Run("basic_include", func(t *testing.T) {
		program, diagnostics := analyzeMultiFileTest(t, "testdata/multifile/basic_include", "main.vdl")

		assert.Empty(t, diagnostics)
		require.NotNil(t, program)
		assert.Contains(t, program.Types, "BaseEntity")
		assert.Contains(t, program.Types, "User")

		user := program.Types["User"]
		require.NotNil(t, user)
		require.Len(t, user.Spreads, 1)
		assert.Equal(t, "BaseEntity", user.Spreads[0].Name)

		assert.Len(t, program.Files, 2)
	})

	t.Run("circular_include", func(t *testing.T) {
		program, diagnostics := analyzeMultiFileTest(t, "testdata/multifile/circular_include", "a.vdl")

		require.NotEmpty(t, diagnostics)
		hasCircularError := false
		for _, d := range diagnostics {
			if d.Code == analysis.CodeCircularInclude {
				hasCircularError = true
				break
			}
		}
		assert.True(t, hasCircularError)
		require.NotNil(t, program)
		assert.True(t, len(program.Types) >= 1)
	})

	t.Run("external_docstrings", func(t *testing.T) {
		program, diagnostics := analyzeMultiFileTest(t, "testdata/multifile/external_docstrings", "main.vdl")

		assert.Empty(t, diagnostics)
		require.NotNil(t, program)

		user := program.Types["User"]
		require.NotNil(t, user)
		require.NotNil(t, user.Docstring)
		assert.Contains(t, *user.Docstring, "Represents a user in the system")

		var idField *analysis.FieldSymbol
		for _, f := range user.Fields {
			if f.Name == "id" {
				idField = f
				break
			}
		}
		require.NotNil(t, idField)
		require.NotNil(t, idField.Docstring)
		assert.Contains(t, *idField.Docstring, "unique identifier for the user")

		require.NotEmpty(t, program.StandaloneDocs)
		assert.Contains(t, program.StandaloneDocs[0].Content, "General documentation")
	})

	t.Run("external_docstring_not_found", func(t *testing.T) {
		program, diagnostics := analyzeMultiFileTest(t, "testdata/multifile/external_docstring_not_found", "main.vdl")

		require.NotEmpty(t, diagnostics)
		hasE003 := false
		for _, d := range diagnostics {
			if d.Code == analysis.CodeDocstringFileNotFound {
				hasE003 = true
				assert.Contains(t, d.Message, "nonexistent.md")
				break
			}
		}
		assert.True(t, hasE003)
		require.NotNil(t, program)
	})
}

func TestSymbolResolution(t *testing.T) {
	t.Run("type_references", func(t *testing.T) {
		program, diagnostics := analyzeTestFile(t, "testdata/resolution/type_references.vdl")
		assert.Empty(t, diagnostics)
		require.NotNil(t, program)

		user := program.Types["User"]
		require.NotNil(t, user)

		var profileField *analysis.FieldSymbol
		for _, f := range user.Fields {
			if f.Name == "profile" {
				profileField = f
				break
			}
		}
		require.NotNil(t, profileField)
		require.NotNil(t, profileField.Type)
		assert.NotNil(t, profileField.Type.ResolvedType)
		assert.Equal(t, "Profile", profileField.Type.ResolvedType.Name)
	})

	t.Run("enum_references", func(t *testing.T) {
		program, diagnostics := analyzeTestFile(t, "testdata/resolution/enum_references.vdl")
		assert.Empty(t, diagnostics)
		require.NotNil(t, program)

		user := program.Types["User"]
		require.NotNil(t, user)

		var statusField *analysis.FieldSymbol
		for _, f := range user.Fields {
			if f.Name == "status" {
				statusField = f
				break
			}
		}
		require.NotNil(t, statusField)
		require.NotNil(t, statusField.Type)
		assert.NotNil(t, statusField.Type.ResolvedEnum)
		assert.Equal(t, "Status", statusField.Type.ResolvedEnum.Name)
	})
}

func TestBestEffort(t *testing.T) {
	program, diagnostics := analyzeTestFile(t, "testdata/best_effort/partial_errors.vdl")

	require.NotEmpty(t, diagnostics)
	require.NotNil(t, program)
	assert.Contains(t, program.Types, "ValidType")
	assert.Contains(t, program.Types, "InvalidType")
	assert.Contains(t, program.Consts, "validConst")
	assert.Contains(t, program.Consts, "BAD_CONST")
}

func TestEdgeCases(t *testing.T) {
	t.Run("file_not_found", func(t *testing.T) {
		fs := vfs.New()

		program, diagnostics := analysis.Analyze(fs, "/nonexistent/file.vdl")

		require.NotNil(t, program)
		require.Len(t, diagnostics, 1)
		assert.Equal(t, analysis.CodeFileNotFound, diagnostics[0].Code)
	})
}
