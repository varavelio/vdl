package transform

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/varavelio/vdl/toolchain/internal/core/ast"
	"github.com/varavelio/vdl/toolchain/internal/core/parser"
	"github.com/varavelio/vdl/toolchain/internal/formatter"
)

func TestExtractType(t *testing.T) {
	require := require.New(t)

	input := `
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

	schema, err := parser.ParserInstance.ParseString("test.vdl", input)
	require.NoError(err, "failed to parse input schema")

	t.Run("ExtractUser", func(t *testing.T) {
		expected := `
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
		require.Equal("Use NewProfile instead", string(*typeDecl.Deprecated.Message), "deprecated message should match")

		extractedSchema := &ast.Schema{
			Children: []*ast.SchemaChild{
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
		type User {
			id: string
			name: string
		}

		rpc UserService {
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
		}
	`

	schema, err := parser.ParserInstance.ParseString("test.vdl", input)
	require.NoError(err, "failed to parse input schema")

	t.Run("ExtractGetUser", func(t *testing.T) {
		expected := `
			rpc UserService {
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
			}
		`

		procDecl, err := ExtractProc(schema, "UserService", "GetUser")
		require.NoError(err, "failed to extract GetUser proc")

		extractedSchema := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					RPC: &ast.RPCDecl{
						Name: "UserService",
						Children: []*ast.RPCChild{
							{Proc: procDecl},
						},
					},
				},
			},
		}

		gotStr := formatter.FormatSchema(extractedSchema)
		expectedStr, err := formatter.Format("", expected)
		require.NoError(err, "failed to format expected schema")
		require.Equal(expectedStr, gotStr, "extracted GetUser proc does not match expected")
	})

	t.Run("ExtractCreateUser", func(t *testing.T) {
		expected := `
			rpc UserService {
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
			}
		`

		procDecl, err := ExtractProc(schema, "UserService", "CreateUser")
		require.NoError(err, "failed to extract CreateUser proc")

		extractedSchema := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					RPC: &ast.RPCDecl{
						Name: "UserService",
						Children: []*ast.RPCChild{
							{Proc: procDecl},
						},
					},
				},
			},
		}

		gotStr := formatter.FormatSchema(extractedSchema)
		expectedStr, err := formatter.Format("", expected)
		require.NoError(err, "failed to format expected schema")
		require.Equal(expectedStr, gotStr, "extracted CreateUser proc does not match expected")
	})

	t.Run("ExtractDeleteUser", func(t *testing.T) {
		expected := `
			rpc UserService {
				""" Delete user """
				proc DeleteUser {
					input {
						userId: string
					}

					output {
						success: bool
					}
				}
			}
		`

		procDecl, err := ExtractProc(schema, "UserService", "DeleteUser")
		require.NoError(err, "failed to extract DeleteUser proc")

		extractedSchema := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					RPC: &ast.RPCDecl{
						Name: "UserService",
						Children: []*ast.RPCChild{
							{Proc: procDecl},
						},
					},
				},
			},
		}

		gotStr := formatter.FormatSchema(extractedSchema)
		expectedStr, err := formatter.Format("", expected)
		require.NoError(err, "failed to format expected schema")
		require.Equal(expectedStr, gotStr, "extracted DeleteUser proc does not match expected")
	})

	t.Run("ProcNotFound", func(t *testing.T) {
		_, err := ExtractProc(schema, "UserService", "NonExistent")
		require.Error(err, "expected error for non-existent proc")
		require.Contains(err.Error(), "not found", "error message should indicate proc was not found")
	})

	t.Run("RPCNotFound", func(t *testing.T) {
		_, err := ExtractProc(schema, "NonExistentService", "GetUser")
		require.Error(err, "expected error for non-existent RPC")
		require.Contains(err.Error(), "not found", "error message should indicate RPC was not found")
	})

	t.Run("ExtractDeprecatedProc", func(t *testing.T) {
		expected := `
			rpc UserService {
				deprecated
				proc LegacyGetUser {
					input {
						id: string
					}

					output {
						user: User
					}
				}
			}
		`

		procDecl, err := ExtractProc(schema, "UserService", "LegacyGetUser")
		require.NoError(err, "failed to extract LegacyGetUser proc")
		require.NotNil(procDecl.Deprecated, "deprecated field should not be nil")

		extractedSchema := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					RPC: &ast.RPCDecl{
						Name: "UserService",
						Children: []*ast.RPCChild{
							{Proc: procDecl},
						},
					},
				},
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
		type Message {
			id: string
			text: string
		}

		rpc MessagingService {
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
		}
	`

	schema, err := parser.ParserInstance.ParseString("test.vdl", input)
	require.NoError(err, "failed to parse input schema")

	t.Run("ExtractChatStream", func(t *testing.T) {
		expected := `
			rpc MessagingService {
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
			}
		`

		streamDecl, err := ExtractStream(schema, "MessagingService", "ChatStream")
		require.NoError(err, "failed to extract ChatStream")

		extractedSchema := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					RPC: &ast.RPCDecl{
						Name: "MessagingService",
						Children: []*ast.RPCChild{
							{Stream: streamDecl},
						},
					},
				},
			},
		}

		gotStr := formatter.FormatSchema(extractedSchema)
		expectedStr, err := formatter.Format("", expected)
		require.NoError(err, "failed to format expected schema")
		require.Equal(expectedStr, gotStr, "extracted ChatStream does not match expected")
	})

	t.Run("ExtractDataSyncStream", func(t *testing.T) {
		expected := `
			rpc MessagingService {
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
			}
		`

		streamDecl, err := ExtractStream(schema, "MessagingService", "DataSyncStream")
		require.NoError(err, "failed to extract DataSyncStream")

		extractedSchema := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					RPC: &ast.RPCDecl{
						Name: "MessagingService",
						Children: []*ast.RPCChild{
							{Stream: streamDecl},
						},
					},
				},
			},
		}

		gotStr := formatter.FormatSchema(extractedSchema)
		expectedStr, err := formatter.Format("", expected)
		require.NoError(err, "failed to format expected schema")
		require.Equal(expectedStr, gotStr, "extracted DataSyncStream does not match expected")
	})

	t.Run("ExtractLogStream", func(t *testing.T) {
		expected := `
			rpc MessagingService {
				""" Log stream """
				stream LogStream {
					output {
						level: string
						message: string
					}
				}
			}
		`

		streamDecl, err := ExtractStream(schema, "MessagingService", "LogStream")
		require.NoError(err, "failed to extract LogStream")

		extractedSchema := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					RPC: &ast.RPCDecl{
						Name: "MessagingService",
						Children: []*ast.RPCChild{
							{Stream: streamDecl},
						},
					},
				},
			},
		}

		gotStr := formatter.FormatSchema(extractedSchema)
		expectedStr, err := formatter.Format("", expected)
		require.NoError(err, "failed to format expected schema")
		require.Equal(expectedStr, gotStr, "extracted LogStream does not match expected")
	})

	t.Run("StreamNotFound", func(t *testing.T) {
		_, err := ExtractStream(schema, "MessagingService", "NonExistent")
		require.Error(err, "expected error for non-existent stream")
		require.Contains(err.Error(), "not found", "error message should indicate stream was not found")
	})

	t.Run("ExtractDeprecatedStream", func(t *testing.T) {
		expected := `
			rpc MessagingService {
				deprecated("Replaced by ChatStream v2")
				stream OldChatStream {
					input {
						roomId: string
					}

					output {
						msg: string
					}
				}
			}
		`

		streamDecl, err := ExtractStream(schema, "MessagingService", "OldChatStream")
		require.NoError(err, "failed to extract OldChatStream")
		require.NotNil(streamDecl.Deprecated, "deprecated field should not be nil")
		require.NotNil(streamDecl.Deprecated.Message, "deprecated message should not be nil")
		require.Equal("Replaced by ChatStream v2", string(*streamDecl.Deprecated.Message), "deprecated message should match")

		extractedSchema := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					RPC: &ast.RPCDecl{
						Name: "MessagingService",
						Children: []*ast.RPCChild{
							{Stream: streamDecl},
						},
					},
				},
			},
		}

		gotStr := formatter.FormatSchema(extractedSchema)
		expectedStr, err := formatter.Format("", expected)
		require.NoError(err, "failed to format expected schema")
		require.Equal(expectedStr, gotStr, "extracted OldChatStream does not match expected")
	})
}

func TestExtractRPC(t *testing.T) {
	require := require.New(t)

	input := `
		""" User service for managing users """
		rpc UserService {
			proc GetUser {
				input { id: string }
				output { name: string }
			}

			stream UserUpdates {
				output { userId: string }
			}
		}

		rpc AdminService {
			proc DeleteAll {
				output { count: int }
			}
		}
	`

	schema, err := parser.ParserInstance.ParseString("test.vdl", input)
	require.NoError(err, "failed to parse input schema")

	t.Run("ExtractUserService", func(t *testing.T) {
		rpcDecl, err := ExtractRPC(schema, "UserService")
		require.NoError(err, "failed to extract UserService")
		require.Equal("UserService", rpcDecl.Name)
		require.NotNil(rpcDecl.Docstring, "docstring should be present")
		require.Len(rpcDecl.GetProcs(), 1, "should have 1 proc")
		require.Len(rpcDecl.GetStreams(), 1, "should have 1 stream")
	})

	t.Run("ExtractAdminService", func(t *testing.T) {
		rpcDecl, err := ExtractRPC(schema, "AdminService")
		require.NoError(err, "failed to extract AdminService")
		require.Equal("AdminService", rpcDecl.Name)
		require.Len(rpcDecl.GetProcs(), 1, "should have 1 proc")
	})

	t.Run("RPCNotFound", func(t *testing.T) {
		_, err := ExtractRPC(schema, "NonExistent")
		require.Error(err, "expected error for non-existent RPC")
		require.Contains(err.Error(), "not found")
	})
}

func TestExtractConst(t *testing.T) {
	require := require.New(t)

	input := `
		""" Maximum page size """
		const MAX_PAGE_SIZE = 100

		const API_VERSION = "v1.0.0"

		const TAX_RATE = 0.21

		const ENABLED = true

		deprecated("Use NEW_LIMIT instead")
		const OLD_LIMIT = 50
	`

	schema, err := parser.ParserInstance.ParseString("test.vdl", input)
	require.NoError(err, "failed to parse input schema")

	t.Run("ExtractIntConst", func(t *testing.T) {
		constDecl, err := ExtractConst(schema, "MAX_PAGE_SIZE")
		require.NoError(err, "failed to extract MAX_PAGE_SIZE")
		require.Equal("MAX_PAGE_SIZE", constDecl.Name)
		require.NotNil(constDecl.Value.Int, "should be an int const")
		require.Equal("100", *constDecl.Value.Int)
	})

	t.Run("ExtractStringConst", func(t *testing.T) {
		constDecl, err := ExtractConst(schema, "API_VERSION")
		require.NoError(err, "failed to extract API_VERSION")
		require.Equal("API_VERSION", constDecl.Name)
		require.NotNil(constDecl.Value.Str, "should be a string const")
		require.Equal("v1.0.0", string(*constDecl.Value.Str))
	})

	t.Run("ExtractFloatConst", func(t *testing.T) {
		constDecl, err := ExtractConst(schema, "TAX_RATE")
		require.NoError(err, "failed to extract TAX_RATE")
		require.Equal("TAX_RATE", constDecl.Name)
		require.NotNil(constDecl.Value.Float, "should be a float const")
	})

	t.Run("ExtractBoolConst", func(t *testing.T) {
		constDecl, err := ExtractConst(schema, "ENABLED")
		require.NoError(err, "failed to extract ENABLED")
		require.Equal("ENABLED", constDecl.Name)
		require.True(constDecl.Value.True, "should be true")
	})

	t.Run("ExtractDeprecatedConst", func(t *testing.T) {
		constDecl, err := ExtractConst(schema, "OLD_LIMIT")
		require.NoError(err, "failed to extract OLD_LIMIT")
		require.NotNil(constDecl.Deprecated, "should be deprecated")
	})

	t.Run("ConstNotFound", func(t *testing.T) {
		_, err := ExtractConst(schema, "NonExistent")
		require.Error(err, "expected error for non-existent const")
		require.Contains(err.Error(), "not found")
	})
}

func TestExtractEnum(t *testing.T) {
	require := require.New(t)

	input := `
		""" Order status """
		enum OrderStatus {
			Pending
			Processing
			Shipped
			Delivered
		}

		enum Priority {
			Low = 1
			Medium = 2
			High = 3
		}

		deprecated("Use NewStatus instead")
		enum OldStatus {
			Active
			Inactive
		}
	`

	schema, err := parser.ParserInstance.ParseString("test.vdl", input)
	require.NoError(err, "failed to parse input schema")

	t.Run("ExtractOrderStatus", func(t *testing.T) {
		enumDecl, err := ExtractEnum(schema, "OrderStatus")
		require.NoError(err, "failed to extract OrderStatus")
		require.Equal("OrderStatus", enumDecl.Name)
		require.NotNil(enumDecl.Docstring, "should have docstring")
		require.Len(enumDecl.Members, 4, "should have 4 members")
	})

	t.Run("ExtractPriority", func(t *testing.T) {
		enumDecl, err := ExtractEnum(schema, "Priority")
		require.NoError(err, "failed to extract Priority")
		require.Equal("Priority", enumDecl.Name)
		require.Len(enumDecl.Members, 3, "should have 3 members")
	})

	t.Run("ExtractDeprecatedEnum", func(t *testing.T) {
		enumDecl, err := ExtractEnum(schema, "OldStatus")
		require.NoError(err, "failed to extract OldStatus")
		require.NotNil(enumDecl.Deprecated, "should be deprecated")
	})

	t.Run("EnumNotFound", func(t *testing.T) {
		_, err := ExtractEnum(schema, "NonExistent")
		require.Error(err, "expected error for non-existent enum")
		require.Contains(err.Error(), "not found")
	})
}

func TestExtractPattern(t *testing.T) {
	require := require.New(t)

	input := `
		""" User event subject """
		pattern UserEventSubject = "events.users.{userId}.{eventType}"

		pattern CacheKey = "cache:{type}:{id}"

		deprecated("Use NewPattern instead")
		pattern OldPattern = "old.{value}"
	`

	schema, err := parser.ParserInstance.ParseString("test.vdl", input)
	require.NoError(err, "failed to parse input schema")

	t.Run("ExtractUserEventSubject", func(t *testing.T) {
		patternDecl, err := ExtractPattern(schema, "UserEventSubject")
		require.NoError(err, "failed to extract UserEventSubject")
		require.Equal("UserEventSubject", patternDecl.Name)
		require.Equal("events.users.{userId}.{eventType}", string(patternDecl.Value))
		require.NotNil(patternDecl.Docstring, "should have docstring")
	})

	t.Run("ExtractCacheKey", func(t *testing.T) {
		patternDecl, err := ExtractPattern(schema, "CacheKey")
		require.NoError(err, "failed to extract CacheKey")
		require.Equal("CacheKey", patternDecl.Name)
		require.Equal("cache:{type}:{id}", string(patternDecl.Value))
	})

	t.Run("ExtractDeprecatedPattern", func(t *testing.T) {
		patternDecl, err := ExtractPattern(schema, "OldPattern")
		require.NoError(err, "failed to extract OldPattern")
		require.NotNil(patternDecl.Deprecated, "should be deprecated")
	})

	t.Run("PatternNotFound", func(t *testing.T) {
		_, err := ExtractPattern(schema, "NonExistent")
		require.Error(err, "expected error for non-existent pattern")
		require.Contains(err.Error(), "not found")
	})
}

func TestExtractTypeStr(t *testing.T) {
	require := require.New(t)

	input := `
		type User {
			id: string
			name: string
		}

		type Post {
			title: string
		}
	`

	result, err := ExtractTypeStr("test.vdl", input, "User")
	require.NoError(err, "failed to extract type from string")
	require.NotEmpty(result, "result should not be empty")

	// Verify that only User type is in the result
	require.Contains(result, "type User", "should contain User type")
	require.Contains(result, "id: string", "should contain User fields")
	require.NotContains(result, "type Post", "should not contain Post type")

	// Verify the result is valid VDL
	expectedStr, err := formatter.Format("", result)
	require.NoError(err, "result should be valid VDL")
	require.Equal(expectedStr, result, "result should be properly formatted")
}

func TestExtractTypeStr_NotFound(t *testing.T) {
	require := require.New(t)

	input := `
		type User {
			id: string
		}
	`

	_, err := ExtractTypeStr("test.vdl", input, "NonExistent")
	require.Error(err, "should error when type not found")
	require.Contains(err.Error(), "not found", "error should mention not found")
}

func TestExtractTypeStr_EmptyInput(t *testing.T) {
	require := require.New(t)

	_, err := ExtractTypeStr("test.vdl", "", "User")
	require.Error(err, "should error on empty input")
	require.Contains(err.Error(), "empty", "error should mention empty")
}

func TestExtractTypeStr_Deprecated(t *testing.T) {
	require := require.New(t)

	input := `
		deprecated("Use NewUser instead")
		type OldUser {
			id: string
		}
	`

	result, err := ExtractTypeStr("test.vdl", input, "OldUser")
	require.NoError(err, "failed to extract deprecated type")
	require.Contains(result, "deprecated", "should contain deprecated keyword")
	require.Contains(result, "Use NewUser instead", "should contain deprecation message")
}

func TestExtractTypeStr_PreservesDocstrings(t *testing.T) {
	require := require.New(t)

	input := `
		""" User documentation """
		type User {
			""" User ID """
			id: string

			""" User name """
			name: string
		}
	`

	result, err := ExtractTypeStr("test.vdl", input, "User")
	require.NoError(err, "failed to extract type")

	// Verify docstrings are preserved
	require.Contains(result, `""" User documentation """`, "should preserve type docstring")
	require.Contains(result, `""" User ID """`, "should preserve field docstrings")
	require.Contains(result, `""" User name """`, "should preserve field docstrings")
}

func TestExtractProcStr(t *testing.T) {
	require := require.New(t)

	input := `
		rpc TestService {
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
		}
	`

	result, err := ExtractProcStr("test.vdl", input, "TestService", "GetUser")
	require.NoError(err, "failed to extract proc from string")
	require.NotEmpty(result, "result should not be empty")

	// Verify that only GetUser proc is in the result
	require.Contains(result, "proc GetUser", "should contain GetUser proc")
	require.Contains(result, "userId: string", "should contain GetUser input fields")
	require.NotContains(result, "proc CreateUser", "should not contain CreateUser proc")

	// Verify the result is valid VDL
	expectedStr, err := formatter.Format("", result)
	require.NoError(err, "result should be valid VDL")
	require.Equal(expectedStr, result, "result should be properly formatted")
}

func TestExtractProcStr_NotFound(t *testing.T) {
	require := require.New(t)

	input := `
		rpc TestService {
			proc GetUser {
				output {
					name: string
				}
			}
		}
	`

	_, err := ExtractProcStr("test.vdl", input, "TestService", "NonExistent")
	require.Error(err, "should error when proc not found")
	require.Contains(err.Error(), "not found", "error should mention not found")
}

func TestExtractProcStr_Deprecated(t *testing.T) {
	require := require.New(t)

	input := `
		rpc TestService {
			deprecated
			proc OldGetUser {
				output {
					name: string
				}
			}
		}
	`

	result, err := ExtractProcStr("test.vdl", input, "TestService", "OldGetUser")
	require.NoError(err, "failed to extract deprecated proc")
	require.Contains(result, "deprecated", "should contain deprecated keyword")
}

func TestExtractStreamStr(t *testing.T) {
	require := require.New(t)

	input := `
		rpc TestService {
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
		}
	`

	result, err := ExtractStreamStr("test.vdl", input, "TestService", "ChatStream")
	require.NoError(err, "failed to extract stream from string")
	require.NotEmpty(result, "result should not be empty")

	// Verify that only ChatStream is in the result
	require.Contains(result, "stream ChatStream", "should contain ChatStream")
	require.Contains(result, "roomId: string", "should contain ChatStream input fields")
	require.NotContains(result, "stream LogStream", "should not contain LogStream")

	// Verify the result is valid VDL
	expectedStr, err := formatter.Format("", result)
	require.NoError(err, "result should be valid VDL")
	require.Equal(expectedStr, result, "result should be properly formatted")
}

func TestExtractStreamStr_NotFound(t *testing.T) {
	require := require.New(t)

	input := `
		rpc TestService {
			stream ChatStream {
				output {
					message: string
				}
			}
		}
	`

	_, err := ExtractStreamStr("test.vdl", input, "TestService", "NonExistent")
	require.Error(err, "should error when stream not found")
	require.Contains(err.Error(), "not found", "error should mention not found")
}

func TestExtractStreamStr_Deprecated(t *testing.T) {
	require := require.New(t)

	input := `
		rpc TestService {
			deprecated("Use ChatStream v2")
			stream OldChatStream {
				output {
					msg: string
				}
			}
		}
	`

	result, err := ExtractStreamStr("test.vdl", input, "TestService", "OldChatStream")
	require.NoError(err, "failed to extract deprecated stream")
	require.Contains(result, "deprecated", "should contain deprecated keyword")
	require.Contains(result, "Use ChatStream v2", "should contain deprecation message")
}

func TestExtractRPCStr(t *testing.T) {
	require := require.New(t)

	input := `
		rpc UserService {
			proc GetUser {
				input { id: string }
				output { name: string }
			}
		}

		rpc AdminService {
			proc DeleteAll {
				output { count: int }
			}
		}
	`

	result, err := ExtractRPCStr("test.vdl", input, "UserService")
	require.NoError(err, "failed to extract RPC from string")
	require.NotEmpty(result, "result should not be empty")
	require.Contains(result, "rpc UserService", "should contain UserService")
	require.Contains(result, "proc GetUser", "should contain GetUser proc")
	require.NotContains(result, "rpc AdminService", "should not contain AdminService")
}

func TestExtractConstStr(t *testing.T) {
	require := require.New(t)

	input := `
		const MAX_SIZE = 100
		const MIN_SIZE = 10
	`

	result, err := ExtractConstStr("test.vdl", input, "MAX_SIZE")
	require.NoError(err, "failed to extract const from string")
	require.NotEmpty(result, "result should not be empty")
	require.Contains(result, "MAX_SIZE", "should contain MAX_SIZE")
	require.Contains(result, "100", "should contain value")
	require.NotContains(result, "MIN_SIZE", "should not contain MIN_SIZE")
}

func TestExtractEnumStr(t *testing.T) {
	require := require.New(t)

	input := `
		enum Status {
			Active
			Inactive
		}

		enum Priority {
			Low
			High
		}
	`

	result, err := ExtractEnumStr("test.vdl", input, "Status")
	require.NoError(err, "failed to extract enum from string")
	require.NotEmpty(result, "result should not be empty")
	require.Contains(result, "enum Status", "should contain Status enum")
	require.Contains(result, "Active", "should contain Active member")
	require.NotContains(result, "enum Priority", "should not contain Priority enum")
}

func TestExtractPatternStr(t *testing.T) {
	require := require.New(t)

	input := `
		pattern UserEvent = "users.{id}.events"
		pattern AdminEvent = "admin.{action}"
	`

	result, err := ExtractPatternStr("test.vdl", input, "UserEvent")
	require.NoError(err, "failed to extract pattern from string")
	require.NotEmpty(result, "result should not be empty")
	require.Contains(result, "pattern UserEvent", "should contain UserEvent pattern")
	require.Contains(result, "users.{id}.events", "should contain pattern value")
	require.NotContains(result, "pattern AdminEvent", "should not contain AdminEvent pattern")
}

func TestIntegration_ExtractAndExpand(t *testing.T) {
	require := require.New(t)

	input := `
		type User {
			id: string
			name: string
		}

		type Post {
			title: string
			author: User
		}

		rpc TestService {
			proc GetPost {
				input {
					postId: string
				}

				output {
					post: Post
				}
			}
		}
	`

	// Extract the proc
	extracted, err := ExtractProcStr("test.vdl", input, "TestService", "GetPost")
	require.NoError(err, "failed to extract proc")

	// Now expand types in the extracted proc (it should have User type reference)
	// But since we only extracted the proc without the type definitions,
	// the expand should keep the named references as they are
	expanded, err := ExpandTypesStr("test.vdl", extracted)
	require.NoError(err, "failed to expand extracted proc")
	require.NotEmpty(expanded, "expanded result should not be empty")

	// Since there's no User type definition in the extracted schema,
	// Post should remain as a named reference
	require.Contains(expanded, "post: Post", "Post should remain as named reference")
}
