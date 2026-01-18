package transform

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/varavelio/vdl/urpc/internal/urpc/ast"
	"github.com/varavelio/vdl/urpc/internal/urpc/formatter"
	"github.com/varavelio/vdl/urpc/internal/urpc/parser"
)

func TestExtractType(t *testing.T) {
	require := require.New(t)

	input := `
		version 1

		""" User type with basic fields """
		type User {
			""" Unique identifier """
			id: string
			name: string
			email: string
			""" User age """
			age: int
		}

		""" Post type with nested objects """
		type Post {
			id: string
			""" Post title """
			title: string
			content: string
			author: User
			tags: string[]
			metadata: {
				""" Creation timestamp """
				createdAt: datetime
				updatedAt: datetime
				views: int
			}
		}

	""" Comment type """
	type Comment {
		id: string
		text: string
		author: User
	}

	""" Legacy type - do not use """
	deprecated("Use NewProfile instead")
	type OldProfile {
		userId: string
		bio: string
	}
`

	schema, err := parser.ParserInstance.ParseString("test.urpc", input)
	require.NoError(err, "failed to parse input schema")

	t.Run("ExtractUser", func(t *testing.T) {
		expected := `
			version 1

			""" User type with basic fields """
			type User {
				""" Unique identifier """
				id: string
				name: string
				email: string
				""" User age """
				age: int
			}
		`

		typeDecl, err := ExtractType(schema, "User")
		require.NoError(err, "failed to extract User type")

		// Create a minimal schema with just the extracted type
		extractedSchema := &ast.Schema{
			Children: []*ast.SchemaChild{
				{Version: &ast.Version{Number: 1}},
				{Type: typeDecl},
			},
		}

		gotStr := formatter.FormatSchema(extractedSchema)
		expectedStr, err := formatter.Format("", expected)
		require.NoError(err, "failed to format expected schema")
		require.Equal(expectedStr, gotStr, "extracted User type does not match expected")
	})

	t.Run("ExtractPost", func(t *testing.T) {
		expected := `
			version 1

			""" Post type with nested objects """
			type Post {
				id: string
				""" Post title """
				title: string
				content: string
				author: User
				tags: string[]
				metadata: {
					""" Creation timestamp """
					createdAt: datetime
					updatedAt: datetime
					views: int
				}
			}
		`

		typeDecl, err := ExtractType(schema, "Post")
		require.NoError(err, "failed to extract Post type")

		extractedSchema := &ast.Schema{
			Children: []*ast.SchemaChild{
				{Version: &ast.Version{Number: 1}},
				{Type: typeDecl},
			},
		}

		gotStr := formatter.FormatSchema(extractedSchema)
		expectedStr, err := formatter.Format("", expected)
		require.NoError(err, "failed to format expected schema")
		require.Equal(expectedStr, gotStr, "extracted Post type does not match expected")
	})

	t.Run("ExtractComment", func(t *testing.T) {
		expected := `
			version 1

			""" Comment type """
			type Comment {
				id: string
				text: string
				author: User
			}
		`

		typeDecl, err := ExtractType(schema, "Comment")
		require.NoError(err, "failed to extract Comment type")

		extractedSchema := &ast.Schema{
			Children: []*ast.SchemaChild{
				{Version: &ast.Version{Number: 1}},
				{Type: typeDecl},
			},
		}

		gotStr := formatter.FormatSchema(extractedSchema)
		expectedStr, err := formatter.Format("", expected)
		require.NoError(err, "failed to format expected schema")
		require.Equal(expectedStr, gotStr, "extracted Comment type does not match expected")
	})

	t.Run("TypeNotFound", func(t *testing.T) {
		_, err := ExtractType(schema, "NonExistent")
		require.Error(err, "expected error for non-existent type")
		require.Contains(err.Error(), "not found", "error message should indicate type was not found")
	})

	t.Run("ExtractDeprecatedType", func(t *testing.T) {
		expected := `
			version 1

			""" Legacy type - do not use """
			deprecated("Use NewProfile instead")
			type OldProfile {
				userId: string
				bio: string
			}
		`

		typeDecl, err := ExtractType(schema, "OldProfile")
		require.NoError(err, "failed to extract OldProfile type")
		require.NotNil(typeDecl.Deprecated, "deprecated field should not be nil")
		require.NotNil(typeDecl.Deprecated.Message, "deprecated message should not be nil")
		require.Equal("Use NewProfile instead", *typeDecl.Deprecated.Message, "deprecated message should match")

		extractedSchema := &ast.Schema{
			Children: []*ast.SchemaChild{
				{Version: &ast.Version{Number: 1}},
				{Type: typeDecl},
			},
		}

		gotStr := formatter.FormatSchema(extractedSchema)
		expectedStr, err := formatter.Format("", expected)
		require.NoError(err, "failed to format expected schema")
		require.Equal(expectedStr, gotStr, "extracted OldProfile type does not match expected")
	})
}

func TestExtractProc(t *testing.T) {
	require := require.New(t)

	input := `
		version 1

		type User {
			id: string
			name: string
		}

		""" Get user by ID """
		proc GetUser {
			input {
				""" User ID to fetch """
				userId: string
			}

			output {
				user: User
				found: bool
			}
		}

		""" Create a new user """
		proc CreateUser {
			input {
				name: string
				email: string
				profile: {
					""" User bio """
					bio: string
					avatar: string
				}
			}

			output {
				""" Created user """
				user: User
				success: bool
			}
		}

	""" Delete user """
	proc DeleteUser {
		input {
			userId: string
		}

		output {
			success: bool
		}
	}

	deprecated
	proc LegacyGetUser {
		input {
			id: string
		}

		output {
			user: User
		}
	}
`

	schema, err := parser.ParserInstance.ParseString("test.urpc", input)
	require.NoError(err, "failed to parse input schema")

	t.Run("ExtractGetUser", func(t *testing.T) {
		expected := `
			version 1

			""" Get user by ID """
			proc GetUser {
				input {
					""" User ID to fetch """
					userId: string
				}

				output {
					user: User
					found: bool
				}
			}
		`

		procDecl, err := ExtractProc(schema, "GetUser")
		require.NoError(err, "failed to extract GetUser proc")

		extractedSchema := &ast.Schema{
			Children: []*ast.SchemaChild{
				{Version: &ast.Version{Number: 1}},
				{Proc: procDecl},
			},
		}

		gotStr := formatter.FormatSchema(extractedSchema)
		expectedStr, err := formatter.Format("", expected)
		require.NoError(err, "failed to format expected schema")
		require.Equal(expectedStr, gotStr, "extracted GetUser proc does not match expected")
	})

	t.Run("ExtractCreateUser", func(t *testing.T) {
		expected := `
			version 1

			""" Create a new user """
			proc CreateUser {
				input {
					name: string
					email: string
					profile: {
						""" User bio """
						bio: string
						avatar: string
					}
				}

				output {
					""" Created user """
					user: User
					success: bool
				}
			}
		`

		procDecl, err := ExtractProc(schema, "CreateUser")
		require.NoError(err, "failed to extract CreateUser proc")

		extractedSchema := &ast.Schema{
			Children: []*ast.SchemaChild{
				{Version: &ast.Version{Number: 1}},
				{Proc: procDecl},
			},
		}

		gotStr := formatter.FormatSchema(extractedSchema)
		expectedStr, err := formatter.Format("", expected)
		require.NoError(err, "failed to format expected schema")
		require.Equal(expectedStr, gotStr, "extracted CreateUser proc does not match expected")
	})

	t.Run("ExtractDeleteUser", func(t *testing.T) {
		expected := `
			version 1

			""" Delete user """
			proc DeleteUser {
				input {
					userId: string
				}

				output {
					success: bool
				}
			}
		`

		procDecl, err := ExtractProc(schema, "DeleteUser")
		require.NoError(err, "failed to extract DeleteUser proc")

		extractedSchema := &ast.Schema{
			Children: []*ast.SchemaChild{
				{Version: &ast.Version{Number: 1}},
				{Proc: procDecl},
			},
		}

		gotStr := formatter.FormatSchema(extractedSchema)
		expectedStr, err := formatter.Format("", expected)
		require.NoError(err, "failed to format expected schema")
		require.Equal(expectedStr, gotStr, "extracted DeleteUser proc does not match expected")
	})

	t.Run("ProcNotFound", func(t *testing.T) {
		_, err := ExtractProc(schema, "NonExistent")
		require.Error(err, "expected error for non-existent proc")
		require.Contains(err.Error(), "not found", "error message should indicate proc was not found")
	})

	t.Run("ExtractDeprecatedProc", func(t *testing.T) {
		expected := `
			version 1

			deprecated
			proc LegacyGetUser {
				input {
					id: string
				}

				output {
					user: User
				}
			}
		`

		procDecl, err := ExtractProc(schema, "LegacyGetUser")
		require.NoError(err, "failed to extract LegacyGetUser proc")
		require.NotNil(procDecl.Deprecated, "deprecated field should not be nil")

		extractedSchema := &ast.Schema{
			Children: []*ast.SchemaChild{
				{Version: &ast.Version{Number: 1}},
				{Proc: procDecl},
			},
		}

		gotStr := formatter.FormatSchema(extractedSchema)
		expectedStr, err := formatter.Format("", expected)
		require.NoError(err, "failed to format expected schema")
		require.Equal(expectedStr, gotStr, "extracted LegacyGetUser proc does not match expected")
	})
}

func TestExtractStream(t *testing.T) {
	require := require.New(t)

	input := `
		version 1

		type Message {
			id: string
			text: string
		}

		""" Chat stream for real-time messaging """
		stream ChatStream {
			input {
				""" Message content """
				message: string
				userId: string
			}

			output {
				message: Message
				""" Timestamp """
				timestamp: datetime
			}
		}

		""" Data synchronization stream """
		stream DataSyncStream {
			input {
				action: string
				data: {
					""" Record ID """
					id: string
					payload: string
				}
			}

			output {
				""" Sync status """
				status: string
				changes: int
			}
		}

	""" Log stream """
	stream LogStream {
		output {
			level: string
			message: string
		}
	}

	deprecated("Replaced by ChatStream v2")
	stream OldChatStream {
		input {
			roomId: string
		}

		output {
			msg: string
		}
	}
`

	schema, err := parser.ParserInstance.ParseString("test.urpc", input)
	require.NoError(err, "failed to parse input schema")

	t.Run("ExtractChatStream", func(t *testing.T) {
		expected := `
			version 1

			""" Chat stream for real-time messaging """
			stream ChatStream {
				input {
					""" Message content """
					message: string
					userId: string
				}

				output {
					message: Message
					""" Timestamp """
					timestamp: datetime
				}
			}
		`

		streamDecl, err := ExtractStream(schema, "ChatStream")
		require.NoError(err, "failed to extract ChatStream")

		extractedSchema := &ast.Schema{
			Children: []*ast.SchemaChild{
				{Version: &ast.Version{Number: 1}},
				{Stream: streamDecl},
			},
		}

		gotStr := formatter.FormatSchema(extractedSchema)
		expectedStr, err := formatter.Format("", expected)
		require.NoError(err, "failed to format expected schema")
		require.Equal(expectedStr, gotStr, "extracted ChatStream does not match expected")
	})

	t.Run("ExtractDataSyncStream", func(t *testing.T) {
		expected := `
			version 1

			""" Data synchronization stream """
			stream DataSyncStream {
				input {
					action: string
					data: {
						""" Record ID """
						id: string
						payload: string
					}
				}

				output {
					""" Sync status """
					status: string
					changes: int
				}
			}
		`

		streamDecl, err := ExtractStream(schema, "DataSyncStream")
		require.NoError(err, "failed to extract DataSyncStream")

		extractedSchema := &ast.Schema{
			Children: []*ast.SchemaChild{
				{Version: &ast.Version{Number: 1}},
				{Stream: streamDecl},
			},
		}

		gotStr := formatter.FormatSchema(extractedSchema)
		expectedStr, err := formatter.Format("", expected)
		require.NoError(err, "failed to format expected schema")
		require.Equal(expectedStr, gotStr, "extracted DataSyncStream does not match expected")
	})

	t.Run("ExtractLogStream", func(t *testing.T) {
		expected := `
			version 1

			""" Log stream """
			stream LogStream {
				output {
					level: string
					message: string
				}
			}
		`

		streamDecl, err := ExtractStream(schema, "LogStream")
		require.NoError(err, "failed to extract LogStream")

		extractedSchema := &ast.Schema{
			Children: []*ast.SchemaChild{
				{Version: &ast.Version{Number: 1}},
				{Stream: streamDecl},
			},
		}

		gotStr := formatter.FormatSchema(extractedSchema)
		expectedStr, err := formatter.Format("", expected)
		require.NoError(err, "failed to format expected schema")
		require.Equal(expectedStr, gotStr, "extracted LogStream does not match expected")
	})

	t.Run("StreamNotFound", func(t *testing.T) {
		_, err := ExtractStream(schema, "NonExistent")
		require.Error(err, "expected error for non-existent stream")
		require.Contains(err.Error(), "not found", "error message should indicate stream was not found")
	})

	t.Run("ExtractDeprecatedStream", func(t *testing.T) {
		expected := `
			version 1

			deprecated("Replaced by ChatStream v2")
			stream OldChatStream {
				input {
					roomId: string
				}

				output {
					msg: string
				}
			}
		`

		streamDecl, err := ExtractStream(schema, "OldChatStream")
		require.NoError(err, "failed to extract OldChatStream")
		require.NotNil(streamDecl.Deprecated, "deprecated field should not be nil")
		require.NotNil(streamDecl.Deprecated.Message, "deprecated message should not be nil")
		require.Equal("Replaced by ChatStream v2", *streamDecl.Deprecated.Message, "deprecated message should match")

		extractedSchema := &ast.Schema{
			Children: []*ast.SchemaChild{
				{Version: &ast.Version{Number: 1}},
				{Stream: streamDecl},
			},
		}

		gotStr := formatter.FormatSchema(extractedSchema)
		expectedStr, err := formatter.Format("", expected)
		require.NoError(err, "failed to format expected schema")
		require.Equal(expectedStr, gotStr, "extracted OldChatStream does not match expected")
	})
}

func TestExtractTypeStr(t *testing.T) {
	require := require.New(t)

	input := `
		version 1

		type User {
			id: string
			name: string
		}

		type Post {
			title: string
		}
	`

	result, err := ExtractTypeStr("test.urpc", input, "User")
	require.NoError(err, "failed to extract type from string")
	require.NotEmpty(result, "result should not be empty")

	// Verify that only User type is in the result
	require.Contains(result, "type User", "should contain User type")
	require.Contains(result, "id: string", "should contain User fields")
	require.NotContains(result, "type Post", "should not contain Post type")

	// Verify the result is valid URPC
	expectedStr, err := formatter.Format("", result)
	require.NoError(err, "result should be valid URPC")
	require.Equal(expectedStr, result, "result should be properly formatted")
}

func TestExtractTypeStr_NotFound(t *testing.T) {
	require := require.New(t)

	input := `
		version 1

		type User {
			id: string
		}
	`

	_, err := ExtractTypeStr("test.urpc", input, "NonExistent")
	require.Error(err, "should error when type not found")
	require.Contains(err.Error(), "not found", "error should mention not found")
}

func TestExtractTypeStr_EmptyInput(t *testing.T) {
	require := require.New(t)

	_, err := ExtractTypeStr("test.urpc", "", "User")
	require.Error(err, "should error on empty input")
	require.Contains(err.Error(), "empty", "error should mention empty")
}

func TestExtractTypeStr_Deprecated(t *testing.T) {
	require := require.New(t)

	input := `
		version 1

		deprecated("Use NewUser instead")
		type OldUser {
			id: string
		}
	`

	result, err := ExtractTypeStr("test.urpc", input, "OldUser")
	require.NoError(err, "failed to extract deprecated type")
	require.Contains(result, "deprecated", "should contain deprecated keyword")
	require.Contains(result, "Use NewUser instead", "should contain deprecation message")
}

func TestExtractTypeStr_PreservesDocstrings(t *testing.T) {
	require := require.New(t)

	input := `
		version 1

		""" User documentation """
		type User {
			""" User ID """
			id: string

			""" User name """
			name: string
		}
	`

	result, err := ExtractTypeStr("test.urpc", input, "User")
	require.NoError(err, "failed to extract type")

	// Verify docstrings are preserved
	require.Contains(result, `""" User documentation """`, "should preserve type docstring")
	require.Contains(result, `""" User ID """`, "should preserve field docstrings")
	require.Contains(result, `""" User name """`, "should preserve field docstrings")
}

func TestExtractProcStr(t *testing.T) {
	require := require.New(t)

	input := `
		version 1

		proc GetUser {
			input {
				userId: string
			}

			output {
				name: string
			}
		}

		proc CreateUser {
			input {
				name: string
			}

			output {
				id: string
			}
		}
	`

	result, err := ExtractProcStr("test.urpc", input, "GetUser")
	require.NoError(err, "failed to extract proc from string")
	require.NotEmpty(result, "result should not be empty")

	// Verify that only GetUser proc is in the result
	require.Contains(result, "proc GetUser", "should contain GetUser proc")
	require.Contains(result, "userId: string", "should contain GetUser input fields")
	require.NotContains(result, "proc CreateUser", "should not contain CreateUser proc")

	// Verify the result is valid URPC
	expectedStr, err := formatter.Format("", result)
	require.NoError(err, "result should be valid URPC")
	require.Equal(expectedStr, result, "result should be properly formatted")
}

func TestExtractProcStr_NotFound(t *testing.T) {
	require := require.New(t)

	input := `
		version 1

		proc GetUser {
			output {
				name: string
			}
		}
	`

	_, err := ExtractProcStr("test.urpc", input, "NonExistent")
	require.Error(err, "should error when proc not found")
	require.Contains(err.Error(), "not found", "error should mention not found")
}

func TestExtractProcStr_Deprecated(t *testing.T) {
	require := require.New(t)

	input := `
		version 1

		deprecated
		proc OldGetUser {
			output {
				name: string
			}
		}
	`

	result, err := ExtractProcStr("test.urpc", input, "OldGetUser")
	require.NoError(err, "failed to extract deprecated proc")
	require.Contains(result, "deprecated", "should contain deprecated keyword")
}

func TestExtractStreamStr(t *testing.T) {
	require := require.New(t)

	input := `
		version 1

		stream ChatStream {
			input {
				roomId: string
			}

			output {
				message: string
			}
		}

		stream LogStream {
			output {
				level: string
			}
		}
	`

	result, err := ExtractStreamStr("test.urpc", input, "ChatStream")
	require.NoError(err, "failed to extract stream from string")
	require.NotEmpty(result, "result should not be empty")

	// Verify that only ChatStream is in the result
	require.Contains(result, "stream ChatStream", "should contain ChatStream")
	require.Contains(result, "roomId: string", "should contain ChatStream input fields")
	require.NotContains(result, "stream LogStream", "should not contain LogStream")

	// Verify the result is valid URPC
	expectedStr, err := formatter.Format("", result)
	require.NoError(err, "result should be valid URPC")
	require.Equal(expectedStr, result, "result should be properly formatted")
}

func TestExtractStreamStr_NotFound(t *testing.T) {
	require := require.New(t)

	input := `
		version 1

		stream ChatStream {
			output {
				message: string
			}
		}
	`

	_, err := ExtractStreamStr("test.urpc", input, "NonExistent")
	require.Error(err, "should error when stream not found")
	require.Contains(err.Error(), "not found", "error should mention not found")
}

func TestExtractStreamStr_Deprecated(t *testing.T) {
	require := require.New(t)

	input := `
		version 1

		deprecated("Use ChatStream v2")
		stream OldChatStream {
			output {
				msg: string
			}
		}
	`

	result, err := ExtractStreamStr("test.urpc", input, "OldChatStream")
	require.NoError(err, "failed to extract deprecated stream")
	require.Contains(result, "deprecated", "should contain deprecated keyword")
	require.Contains(result, "Use ChatStream v2", "should contain deprecation message")
}

func TestIntegration_ExtractAndExpand(t *testing.T) {
	require := require.New(t)

	input := `
		version 1

		type User {
			id: string
			name: string
		}

		type Post {
			title: string
			author: User
		}

		proc GetPost {
			input {
				postId: string
			}

			output {
				post: Post
			}
		}
	`

	// Extract the proc
	extracted, err := ExtractProcStr("test.urpc", input, "GetPost")
	require.NoError(err, "failed to extract proc")

	// Now expand types in the extracted proc (it should have User type reference)
	// But since we only extracted the proc without the type definitions,
	// the expand should keep the named references as they are
	expanded, err := ExpandTypesStr("test.urpc", extracted)
	require.NoError(err, "failed to expand extracted proc")
	require.NotEmpty(expanded, "expanded result should not be empty")

	// Since there's no User type definition in the extracted schema,
	// Post should remain as a named reference
	require.Contains(expanded, "post: Post", "Post should remain as named reference")
}
