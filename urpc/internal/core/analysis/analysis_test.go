package analysis_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uforg/uforpc/urpc/internal/core/analysis"
	"github.com/uforg/uforpc/urpc/internal/core/parser"
	"github.com/uforg/uforpc/urpc/internal/core/vfs"
)

func TestAnalyzeSchema_ValidSchema(t *testing.T) {
	schema, err := parser.ParserInstance.ParseString("test.ufo", `
		type User {
			id: string
			name: string
			email?: string
		}

		enum Status {
			Active
			Inactive
		}

		const MAX_USERS = 100

		pattern UserEvent = "users.{userId}.{event}"

		rpc Users {
			proc GetUser {
				input {
					id: string
				}
				output {
					user: User
				}
			}

			stream UserUpdates {
				input {
					userId: string
				}
				output {
					status: Status
				}
			}
		}
	`)
	require.NoError(t, err)

	program, diagnostics := analysis.AnalyzeSchema(schema, "test.ufo")

	assert.Empty(t, diagnostics, "Expected no diagnostics")
	require.NotNil(t, program, "Expected program to be non-nil")

	// Verify types
	assert.Contains(t, program.Types, "User")
	userType := program.Types["User"]
	assert.Equal(t, "User", userType.Name)
	assert.Len(t, userType.Fields, 3)

	// Verify enums
	assert.Contains(t, program.Enums, "Status")
	statusEnum := program.Enums["Status"]
	assert.Equal(t, "Status", statusEnum.Name)
	assert.Len(t, statusEnum.Members, 2)

	// Verify constants
	assert.Contains(t, program.Consts, "MAX_USERS")
	maxUsers := program.Consts["MAX_USERS"]
	assert.Equal(t, "MAX_USERS", maxUsers.Name)
	assert.Equal(t, "100", maxUsers.Value)

	// Verify patterns
	assert.Contains(t, program.Patterns, "UserEvent")
	pattern := program.Patterns["UserEvent"]
	assert.Equal(t, "UserEvent", pattern.Name)
	assert.Equal(t, []string{"userId", "event"}, pattern.Placeholders)

	// Verify RPCs
	assert.Contains(t, program.RPCs, "Users")
	rpc := program.RPCs["Users"]
	assert.Contains(t, rpc.Procs, "GetUser")
	assert.Contains(t, rpc.Streams, "UserUpdates")
}

func TestAnalyzeSchema_DuplicateType(t *testing.T) {
	schema, err := parser.ParserInstance.ParseString("test.ufo", `
		type User {
			id: string
		}

		type User {
			name: string
		}
	`)
	require.NoError(t, err)

	program, diagnostics := analysis.AnalyzeSchema(schema, "test.ufo")

	// Best-effort: program is returned even with errors
	require.NotNil(t, program, "Expected program to be non-nil (best-effort)")
	require.Len(t, diagnostics, 1)
	assert.Equal(t, analysis.CodeDuplicateType, diagnostics[0].Code)
	assert.Contains(t, diagnostics[0].Message, "User")

	// The first User type should still be in the program
	assert.Contains(t, program.Types, "User")
}

func TestAnalyzeSchema_UndefinedType(t *testing.T) {
	schema, err := parser.ParserInstance.ParseString("test.ufo", `
		type User {
			profile: Profile
		}
	`)
	require.NoError(t, err)

	program, diagnostics := analysis.AnalyzeSchema(schema, "test.ufo")

	// Best-effort: program is returned even with errors
	require.NotNil(t, program, "Expected program to be non-nil (best-effort)")
	require.Len(t, diagnostics, 1)
	assert.Equal(t, analysis.CodeTypeNotDeclared, diagnostics[0].Code)
	assert.Contains(t, diagnostics[0].Message, "Profile")

	// User type should still be in the program
	assert.Contains(t, program.Types, "User")
}

func TestAnalyzeSchema_NamingConventions(t *testing.T) {
	// Test PascalCase for types
	t.Run("Type must be PascalCase", func(t *testing.T) {
		schema, err := parser.ParserInstance.ParseString("test.ufo", `
			type user {
				id: string
			}
		`)
		require.NoError(t, err)

		program, diagnostics := analysis.AnalyzeSchema(schema, "test.ufo")
		require.NotNil(t, program, "Expected program (best-effort)")
		require.Len(t, diagnostics, 1)
		assert.Equal(t, analysis.CodeNotPascalCase, diagnostics[0].Code)
		// Type is still collected despite naming error
		assert.Contains(t, program.Types, "user")
	})

	// Test camelCase for fields
	t.Run("Field must be camelCase", func(t *testing.T) {
		schema, err := parser.ParserInstance.ParseString("test.ufo", `
			type User {
				UserId: string
			}
		`)
		require.NoError(t, err)

		program, diagnostics := analysis.AnalyzeSchema(schema, "test.ufo")
		require.NotNil(t, program, "Expected program (best-effort)")
		require.Len(t, diagnostics, 1)
		assert.Equal(t, analysis.CodeNotCamelCase, diagnostics[0].Code)
	})

	// Test UPPER_SNAKE_CASE for constants
	t.Run("Constant must be UPPER_SNAKE_CASE", func(t *testing.T) {
		schema, err := parser.ParserInstance.ParseString("test.ufo", `
			const maxUsers = 100
		`)
		require.NoError(t, err)

		program, diagnostics := analysis.AnalyzeSchema(schema, "test.ufo")
		require.NotNil(t, program, "Expected program (best-effort)")
		require.Len(t, diagnostics, 1)
		assert.Equal(t, analysis.CodeNotUpperSnakeCase, diagnostics[0].Code)
		// Constant is still collected
		assert.Contains(t, program.Consts, "maxUsers")
	})
}

func TestAnalyzeSchema_SpreadValidation(t *testing.T) {
	// Test spread with undefined type
	t.Run("Spread references undefined type", func(t *testing.T) {
		schema, err := parser.ParserInstance.ParseString("test.ufo", `
			type User {
				...BaseEntity
				name: string
			}
		`)
		require.NoError(t, err)

		program, diagnostics := analysis.AnalyzeSchema(schema, "test.ufo")
		require.NotNil(t, program, "Expected program (best-effort)")
		require.Len(t, diagnostics, 1)
		assert.Equal(t, analysis.CodeSpreadTypeNotFound, diagnostics[0].Code)
		// User type is still collected
		assert.Contains(t, program.Types, "User")
	})

	// Test valid spread
	t.Run("Valid spread", func(t *testing.T) {
		schema, err := parser.ParserInstance.ParseString("test.ufo", `
			type BaseEntity {
				id: string
				createdAt: datetime
			}

			type User {
				...BaseEntity
				name: string
			}
		`)
		require.NoError(t, err)

		program, diagnostics := analysis.AnalyzeSchema(schema, "test.ufo")
		assert.Empty(t, diagnostics)
		assert.NotNil(t, program)
	})
}

func TestAnalyzeSchema_EnumValidation(t *testing.T) {
	// Test duplicate enum value
	t.Run("Duplicate enum value", func(t *testing.T) {
		schema, err := parser.ParserInstance.ParseString("test.ufo", `
			enum Status {
				Active = "active"
				Pending = "active"
			}
		`)
		require.NoError(t, err)

		program, diagnostics := analysis.AnalyzeSchema(schema, "test.ufo")
		require.NotNil(t, program, "Expected program (best-effort)")
		require.Len(t, diagnostics, 1)
		assert.Equal(t, analysis.CodeEnumDuplicateValue, diagnostics[0].Code)
		// Enum is still collected
		assert.Contains(t, program.Enums, "Status")
	})
}

func TestAnalyzeSchema_RPCValidation(t *testing.T) {
	// Test duplicate procedure
	t.Run("Duplicate procedure", func(t *testing.T) {
		schema, err := parser.ParserInstance.ParseString("test.ufo", `
			rpc Users {
				proc GetUser {
					input { id: string }
					output { name: string }
				}
				proc GetUser {
					input { email: string }
					output { name: string }
				}
			}
		`)
		require.NoError(t, err)

		program, diagnostics := analysis.AnalyzeSchema(schema, "test.ufo")
		require.NotNil(t, program, "Expected program (best-effort)")
		require.Len(t, diagnostics, 1)
		assert.Equal(t, analysis.CodeDuplicateProc, diagnostics[0].Code)
		// RPC is still collected with one of the procs
		assert.Contains(t, program.RPCs, "Users")
		assert.Contains(t, program.RPCs["Users"].Procs, "GetUser")
	})
}

func TestAnalyze_WithIncludes(t *testing.T) {
	fs := vfs.New()

	// Create files in memory
	fs.WriteFileCache("/project/common.ufo", []byte(`
		type BaseEntity {
			id: string
			createdAt: datetime
		}
	`))

	fs.WriteFileCache("/project/main.ufo", []byte(`
		include "./common.ufo"

		type User {
			...BaseEntity
			name: string
		}
	`))

	program, diagnostics := analysis.Analyze(fs, "/project/main.ufo")

	assert.Empty(t, diagnostics, "Expected no diagnostics")
	require.NotNil(t, program, "Expected program to be non-nil")

	// Both types should be available
	assert.Contains(t, program.Types, "BaseEntity")
	assert.Contains(t, program.Types, "User")

	// Files should be tracked
	assert.Len(t, program.Files, 2)
}

func TestAnalyze_CircularInclude(t *testing.T) {
	fs := vfs.New()

	// Create circular includes
	fs.WriteFileCache("/project/a.ufo", []byte(`
		include "./b.ufo"
		type A { id: string }
	`))

	fs.WriteFileCache("/project/b.ufo", []byte(`
		include "./a.ufo"
		type B { id: string }
	`))

	program, diagnostics := analysis.Analyze(fs, "/project/a.ufo")

	// Best-effort: program is returned even with circular include errors
	require.NotNil(t, program, "Expected program (best-effort)")
	require.NotEmpty(t, diagnostics)
	assert.Equal(t, analysis.CodeCircularInclude, diagnostics[0].Code)

	// At least one file should have been processed
	assert.NotEmpty(t, program.Files)
}

func TestAnalyze_FileNotFound(t *testing.T) {
	fs := vfs.New()

	program, diagnostics := analysis.Analyze(fs, "/nonexistent/file.ufo")

	// Best-effort: returns empty program when entry point not found
	require.NotNil(t, program, "Expected empty program (best-effort)")
	require.Len(t, diagnostics, 1)
	assert.Equal(t, analysis.CodeFileNotFound, diagnostics[0].Code)
}

func TestAnalyze_BestEffort_PartialErrors(t *testing.T) {
	// Test that valid symbols are collected even when some have errors
	schema, err := parser.ParserInstance.ParseString("test.ufo", `
		type ValidType {
			id: string
			name: string
		}

		type InvalidType {
			profile: NonExistentType
		}

		const VALID_CONST = 100
		const invalidConst = 200

		enum ValidEnum {
			Active
			Inactive
		}
	`)
	require.NoError(t, err)

	program, diagnostics := analysis.AnalyzeSchema(schema, "test.ufo")

	// Should have errors
	require.NotEmpty(t, diagnostics, "Expected diagnostics for errors")

	// But all symbols should still be collected
	require.NotNil(t, program, "Expected program (best-effort)")
	assert.Contains(t, program.Types, "ValidType")
	assert.Contains(t, program.Types, "InvalidType")
	assert.Contains(t, program.Consts, "VALID_CONST")
	assert.Contains(t, program.Consts, "invalidConst")
	assert.Contains(t, program.Enums, "ValidEnum")
}
