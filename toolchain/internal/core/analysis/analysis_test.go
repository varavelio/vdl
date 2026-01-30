package analysis_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/varavelio/vdl/toolchain/internal/core/analysis"
	"github.com/varavelio/vdl/toolchain/internal/core/vfs"
)

// ============================================================================
// Valid Schema Tests - testdata/valid/*.vdl should have 0 errors
// ============================================================================

func TestValidSchemas(t *testing.T) {
	files := globTestFiles(t, "valid/*.vdl")
	require.NotEmpty(t, files, "No valid test files found")

	for _, file := range files {
		name := filepath.Base(file)
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			program, diagnostics := analyzeTestFile(t, file)

			// Valid files should have no errors
			if len(diagnostics) > 0 {
				t.Errorf("Expected no diagnostics for valid file %s, got:", name)
				for _, d := range diagnostics {
					t.Errorf("  %s", d)
				}
			}

			// Program should always be returned (best-effort)
			require.NotNil(t, program, "Expected program to be non-nil")
		})
	}
}

// ============================================================================
// Error Schema Tests - testdata/errors/**/*.vdl must have @expect comments
// ============================================================================

func TestErrorSchemas(t *testing.T) {
	files := globTestFilesRecursive(t, "errors")
	require.NotEmpty(t, files, "No error test files found")

	for _, file := range files {
		relPath := relativePath(t, file)
		t.Run(relPath, func(t *testing.T) {
			t.Parallel()

			expected := parseExpectedCodes(t, file)
			program, diagnostics := analyzeTestFile(t, file)

			// Best-effort: program should always be returned
			require.NotNil(t, program, "Expected program to be non-nil (best-effort)")

			// Extract actual codes
			actualCodes := extractCodes(diagnostics)

			// Verify all expected codes are present
			for _, expectedCode := range expected.codes {
				if !containsCode(actualCodes, expectedCode) {
					t.Errorf("Expected diagnostic code %s not found in: %v", expectedCode, actualCodes)
				}
			}

			// Verify we got at least the expected number of diagnostics
			if len(diagnostics) < len(expected.codes) {
				t.Errorf("Expected at least %d diagnostics, got %d: %v",
					len(expected.codes), len(diagnostics), actualCodes)
			}
		})
	}
}

// ============================================================================
// Multi-File Tests - testdata/multifile/*/
// ============================================================================

func TestMultiFileSchemas(t *testing.T) {
	t.Run("basic_include", func(t *testing.T) {
		program, diagnostics := analyzeMultiFileTest(t, "testdata/multifile/basic_include", "main.vdl")

		assert.Empty(t, diagnostics, "Expected no diagnostics")
		require.NotNil(t, program)

		// Both types should be available
		assert.Contains(t, program.Types, "BaseEntity", "BaseEntity should be imported")
		assert.Contains(t, program.Types, "User", "User should be defined")

		// User should have spread from BaseEntity
		user := program.Types["User"]
		require.NotNil(t, user)
		assert.Len(t, user.Spreads, 1, "User should have 1 spread")
		assert.Equal(t, "BaseEntity", user.Spreads[0].TypeName)

		// Files should be tracked
		assert.Len(t, program.Files, 2, "Should have 2 files")
	})

	t.Run("rpc_merge", func(t *testing.T) {
		program, diagnostics := analyzeMultiFileTest(t, "testdata/multifile/rpc_merge", "main.vdl")

		assert.Empty(t, diagnostics, "Expected no diagnostics")
		require.NotNil(t, program)

		// RPC should be merged from multiple files
		assert.Contains(t, program.RPCs, "Users", "Users RPC should exist")

		usersRPC := program.RPCs["Users"]
		require.NotNil(t, usersRPC)

		// Should have procs from users_procs.vdl
		assert.Contains(t, usersRPC.Procs, "GetUser", "GetUser proc should exist")
		assert.Contains(t, usersRPC.Procs, "CreateUser", "CreateUser proc should exist")

		// Should have streams from users_streams.vdl
		assert.Contains(t, usersRPC.Streams, "UserUpdates", "UserUpdates stream should exist")

		// Should be declared in 2 files (users_procs.vdl and users_streams.vdl)
		assert.Len(t, usersRPC.DeclaredIn, 2, "Users RPC should be declared in 2 files")
	})

	t.Run("circular_include", func(t *testing.T) {
		program, diagnostics := analyzeMultiFileTest(t, "testdata/multifile/circular_include", "a.vdl")

		// Should detect circular include
		require.NotEmpty(t, diagnostics, "Expected circular include diagnostic")

		hasCircularError := false
		for _, d := range diagnostics {
			if d.Code == analysis.CodeCircularInclude {
				hasCircularError = true
				break
			}
		}
		assert.True(t, hasCircularError, "Expected CodeCircularInclude error")

		// Best-effort: program should still be returned
		require.NotNil(t, program, "Expected program (best-effort)")

		// At least one type should be available
		assert.True(t, len(program.Types) >= 1, "Should have at least one type")
	})

	t.Run("deep_includes", func(t *testing.T) {
		program, diagnostics := analyzeMultiFileTest(t, "testdata/multifile/deep_includes", "main.vdl")

		assert.Empty(t, diagnostics, "Expected no diagnostics")
		require.NotNil(t, program)

		// All types from the include chain should be available
		assert.Contains(t, program.Types, "Level2", "Level2 should be imported")
		assert.Contains(t, program.Types, "Level1", "Level1 should be imported")
		assert.Contains(t, program.Types, "Main", "Main should be defined")

		// Should have 3 files
		assert.Len(t, program.Files, 3, "Should have 3 files")
	})

	t.Run("external_docstrings", func(t *testing.T) {
		program, diagnostics := analyzeMultiFileTest(t, "testdata/multifile/external_docstrings", "main.vdl")

		assert.Empty(t, diagnostics, "Expected no diagnostics")
		require.NotNil(t, program)

		// Type docstring should be resolved from ./docs/user_type.md
		user := program.Types["User"]
		require.NotNil(t, user, "User type should exist")
		require.NotNil(t, user.Docstring, "User docstring should be resolved")
		assert.Contains(t, *user.Docstring, "Represents a user in the system")

		// Field docstring should be resolved from ./docs/user_id.md
		var idField *analysis.FieldSymbol
		for _, f := range user.Fields {
			if f.Name == "id" {
				idField = f
				break
			}
		}
		require.NotNil(t, idField, "id field should exist")
		require.NotNil(t, idField.Docstring, "id field docstring should be resolved")
		assert.Contains(t, *idField.Docstring, "unique identifier for the user")

		// Enum docstring should be resolved from ./docs/status_enum.md
		status := program.Enums["Status"]
		require.NotNil(t, status, "Status enum should exist")
		require.NotNil(t, status.Docstring, "Status docstring should be resolved")
		assert.Contains(t, *status.Docstring, "current status of an entity")

		// Proc docstring should be resolved from ./docs/get_user_proc.md
		usersRPC := program.RPCs["Users"]
		require.NotNil(t, usersRPC, "Users RPC should exist")
		getUser := usersRPC.Procs["GetUser"]
		require.NotNil(t, getUser, "GetUser proc should exist")
		require.NotNil(t, getUser.Docstring, "GetUser proc docstring should be resolved")
		assert.Contains(t, *getUser.Docstring, "Retrieves a user by their unique identifier")

		// Standalone docstring should be resolved from ./docs/api_intro.md
		require.NotEmpty(t, program.StandaloneDocs, "Should have standalone docstrings")
	})

	t.Run("external_docstring_not_found", func(t *testing.T) {
		program, diagnostics := analyzeMultiFileTest(t, "testdata/multifile/external_docstring_not_found", "main.vdl")

		// Should have E003 error for missing external docstring file
		require.NotEmpty(t, diagnostics, "Expected E003 diagnostic for missing external file")

		hasE003 := false
		for _, d := range diagnostics {
			if d.Code == analysis.CodeDocstringFileNotFound {
				hasE003 = true
				assert.Contains(t, d.Message, "nonexistent.md")
				break
			}
		}
		assert.True(t, hasE003, "Expected CodeDocstringFileNotFound (E003) error")

		// Best-effort: program should still be returned
		require.NotNil(t, program, "Expected program (best-effort)")
		assert.Contains(t, program.Types, "User", "User type should still be collected")
	})
}

// ============================================================================
// Symbol Resolution Tests - testdata/resolution/*.vdl
// Verify ResolvedType/ResolvedEnum are populated for LSP navigation
// ============================================================================

func TestSymbolResolution(t *testing.T) {
	t.Run("type_references", func(t *testing.T) {
		program, diagnostics := analyzeTestFile(t, "testdata/resolution/type_references.vdl")
		assert.Empty(t, diagnostics)
		require.NotNil(t, program)

		user := program.Types["User"]
		require.NotNil(t, user)

		// Find the address field
		var addressField *analysis.FieldSymbol
		for _, f := range user.Fields {
			if f.Name == "address" {
				addressField = f
				break
			}
		}
		require.NotNil(t, addressField, "address field not found")
		assert.NotNil(t, addressField.Type.ResolvedType, "ResolvedType should be populated")
		assert.Equal(t, "Address", addressField.Type.ResolvedType.Name)

		// Find the addresses array field
		var addressesField *analysis.FieldSymbol
		for _, f := range user.Fields {
			if f.Name == "addresses" {
				addressesField = f
				break
			}
		}
		require.NotNil(t, addressesField, "addresses field not found")
		assert.NotNil(t, addressesField.Type.ResolvedType, "ResolvedType should be populated for array")
		assert.Equal(t, "Address", addressesField.Type.ResolvedType.Name)

		// Find the addressMap field
		var addressMapField *analysis.FieldSymbol
		for _, f := range user.Fields {
			if f.Name == "addressMap" {
				addressMapField = f
				break
			}
		}
		require.NotNil(t, addressMapField, "addressMap field not found")
		require.NotNil(t, addressMapField.Type.MapValue, "MapValue should be populated")
		assert.NotNil(t, addressMapField.Type.MapValue.ResolvedType, "ResolvedType should be populated for map value")
	})

	t.Run("enum_references", func(t *testing.T) {
		program, diagnostics := analyzeTestFile(t, "testdata/resolution/enum_references.vdl")
		assert.Empty(t, diagnostics)
		require.NotNil(t, program)

		user := program.Types["User"]
		require.NotNil(t, user)

		// Find the status field
		var statusField *analysis.FieldSymbol
		for _, f := range user.Fields {
			if f.Name == "status" {
				statusField = f
				break
			}
		}
		require.NotNil(t, statusField, "status field not found")
		assert.NotNil(t, statusField.Type.ResolvedEnum, "ResolvedEnum should be populated")
		assert.Equal(t, "Status", statusField.Type.ResolvedEnum.Name)
	})
}

// ============================================================================
// Best-Effort Tests - testdata/best_effort/*.vdl
// Verify Program is always returned even with errors
// ============================================================================

func TestBestEffort(t *testing.T) {
	t.Run("partial_errors", func(t *testing.T) {
		program, diagnostics := analyzeTestFile(t, "testdata/best_effort/partial_errors.vdl")

		// Should have errors
		require.NotEmpty(t, diagnostics, "Expected diagnostics for errors")

		// But all symbols should still be collected
		require.NotNil(t, program, "Expected program (best-effort)")
		assert.Contains(t, program.Types, "ValidType")
		assert.Contains(t, program.Types, "InvalidType")
		assert.Contains(t, program.Consts, "VALID_CONST")
		assert.Contains(t, program.Consts, "invalidConst")
		assert.Contains(t, program.Enums, "ValidEnum")
	})

	t.Run("multiple_errors", func(t *testing.T) {
		program, diagnostics := analyzeTestFile(t, "testdata/best_effort/multiple_errors.vdl")

		// Should have multiple errors
		require.True(t, len(diagnostics) >= 3, "Expected at least 3 diagnostics, got %d", len(diagnostics))

		// All symbols should still be collected
		require.NotNil(t, program)
		assert.Contains(t, program.Types, "invalid")
		assert.Contains(t, program.Enums, "lowercase")
		assert.Contains(t, program.Consts, "badName")
	})
}

// ============================================================================
// RPC Merge Tests - verify RPCs are merged correctly
// ============================================================================

func TestRPCMerge(t *testing.T) {
	t.Run("same_file", func(t *testing.T) {
		// Multiple declarations of the same RPC in a single file should be merged
		program, diagnostics := analyzeTestFile(t, "testdata/valid/rpc_merge_same_file.vdl")

		assert.Empty(t, diagnostics, "Expected no diagnostics")
		require.NotNil(t, program)

		// RPC should exist and be merged
		assert.Contains(t, program.RPCs, "Users", "Users RPC should exist")

		usersRPC := program.RPCs["Users"]
		require.NotNil(t, usersRPC)

		// Should have both procs from different declarations
		assert.Contains(t, usersRPC.Procs, "GetUser", "GetUser proc should exist")
		assert.Contains(t, usersRPC.Procs, "CreateUser", "CreateUser proc should exist")
		assert.Len(t, usersRPC.Procs, 2, "Should have exactly 2 procs")

		// Should have stream from third declaration
		assert.Contains(t, usersRPC.Streams, "UserUpdates", "UserUpdates stream should exist")
		assert.Len(t, usersRPC.Streams, 1, "Should have exactly 1 stream")

		// DeclaredIn should have the same file 3 times (once per declaration)
		assert.Len(t, usersRPC.DeclaredIn, 3, "Users RPC should be declared 3 times in the same file")
	})

	t.Run("multiple_files", func(t *testing.T) {
		// Multiple declarations across different files should also be merged
		program, diagnostics := analyzeMultiFileTest(t, "testdata/multifile/rpc_merge", "main.vdl")

		assert.Empty(t, diagnostics, "Expected no diagnostics")
		require.NotNil(t, program)

		usersRPC := program.RPCs["Users"]
		require.NotNil(t, usersRPC)

		// Should have procs from users_procs.vdl
		assert.Contains(t, usersRPC.Procs, "GetUser", "GetUser proc should exist")
		assert.Contains(t, usersRPC.Procs, "CreateUser", "CreateUser proc should exist")

		// Should have streams from users_streams.vdl
		assert.Contains(t, usersRPC.Streams, "UserUpdates", "UserUpdates stream should exist")

		// Should be declared in 2 files
		assert.Len(t, usersRPC.DeclaredIn, 2, "Users RPC should be declared in 2 files")
	})
}

// ============================================================================
// Edge Case Tests - VFS-based tests for special scenarios
// ============================================================================

func TestEdgeCases(t *testing.T) {
	t.Run("file_not_found", func(t *testing.T) {
		fs := vfs.New()

		program, diagnostics := analysis.Analyze(fs, "/nonexistent/file.vdl")

		require.NotNil(t, program, "Expected empty program (best-effort)")
		require.Len(t, diagnostics, 1)
		assert.Equal(t, analysis.CodeFileNotFound, diagnostics[0].Code)
	})
}
