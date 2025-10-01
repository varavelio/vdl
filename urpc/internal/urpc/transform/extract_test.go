package transform

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/uforg/uforpc/urpc/internal/urpc/ast"
	"github.com/uforg/uforpc/urpc/internal/urpc/formatter"
	"github.com/uforg/uforpc/urpc/internal/urpc/parser"
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
