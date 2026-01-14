package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/uforg/uforpc/urpc/internal/urpc/ast"
	"github.com/uforg/uforpc/urpc/internal/util/testutil"
)

// Helper function to create a pointer to a string
func ptr(s string) *string {
	return &s
}

// Helper function to create a pointer to a QuotedString
func qptr(s string) *ast.QuotedString {
	q := ast.QuotedString(s)
	return &q
}

////////////////
// INCLUDES   //
////////////////

func TestParserInclude(t *testing.T) {
	t.Run("Basic include", func(t *testing.T) {
		input := `include "./foo.ufo"`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					Include: &ast.Include{
						Path: "./foo.ufo",
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Multiple includes", func(t *testing.T) {
		input := `
			include "./foo.ufo"
			include "./bar.ufo"
			include "../common/types.ufo"
		`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{Include: &ast.Include{Path: "./foo.ufo"}},
				{Include: &ast.Include{Path: "./bar.ufo"}},
				{Include: &ast.Include{Path: "../common/types.ufo"}},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})
}

////////////////
// CONSTANTS  //
////////////////

func TestParserConstDecl(t *testing.T) {
	t.Run("String constant", func(t *testing.T) {
		input := `const API_VERSION = "1.0.0"`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					Const: &ast.ConstDecl{
						Name: "API_VERSION",
						Value: &ast.ConstValue{
							Str: qptr("1.0.0"),
						},
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Integer constant", func(t *testing.T) {
		input := `const MAX_PAGE_SIZE = 100`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					Const: &ast.ConstDecl{
						Name: "MAX_PAGE_SIZE",
						Value: &ast.ConstValue{
							Int: ptr("100"),
						},
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Float constant", func(t *testing.T) {
		input := `const DEFAULT_TAX_RATE = 0.21`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					Const: &ast.ConstDecl{
						Name: "DEFAULT_TAX_RATE",
						Value: &ast.ConstValue{
							Float: ptr("0.21"),
						},
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Boolean true constant", func(t *testing.T) {
		input := `const FEATURE_FLAG_ENABLED = true`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					Const: &ast.ConstDecl{
						Name: "FEATURE_FLAG_ENABLED",
						Value: &ast.ConstValue{
							True: true,
						},
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Boolean false constant", func(t *testing.T) {
		input := `const DEBUG_MODE = false`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					Const: &ast.ConstDecl{
						Name: "DEBUG_MODE",
						Value: &ast.ConstValue{
							False: true,
						},
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Constant with docstring", func(t *testing.T) {
		input := `
			""" The maximum number of items allowed per request. """
			const MAX_ITEMS = 50
		`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					Const: &ast.ConstDecl{
						Docstring: &ast.Docstring{
							Value: " The maximum number of items allowed per request. ",
						},
						Name: "MAX_ITEMS",
						Value: &ast.ConstValue{
							Int: ptr("50"),
						},
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Deprecated constant", func(t *testing.T) {
		input := `deprecated const OLD_LIMIT = 100`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					Const: &ast.ConstDecl{
						Deprecated: &ast.Deprecated{},
						Name:       "OLD_LIMIT",
						Value: &ast.ConstValue{
							Int: ptr("100"),
						},
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Deprecated constant with message", func(t *testing.T) {
		input := `deprecated("Use NEW_LIMIT instead") const OLD_LIMIT = 100`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					Const: &ast.ConstDecl{
						Deprecated: &ast.Deprecated{
							Message: qptr("Use NEW_LIMIT instead"),
						},
						Name: "OLD_LIMIT",
						Value: &ast.ConstValue{
							Int: ptr("100"),
						},
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})
}

////////////////
// ENUMS      //
////////////////

func TestParserEnumDecl(t *testing.T) {
	t.Run("String enum with implicit values", func(t *testing.T) {
		input := `
			enum OrderStatus {
				Pending
				Processing
				Shipped
			}
		`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					Enum: &ast.EnumDecl{
						Name: "OrderStatus",
						Members: []*ast.EnumMember{
							{Name: "Pending"},
							{Name: "Processing"},
							{Name: "Shipped"},
						},
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("String enum with explicit values", func(t *testing.T) {
		input := `
			enum HttpMethod {
				Get = "GET"
				Post = "POST"
			}
		`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					Enum: &ast.EnumDecl{
						Name: "HttpMethod",
						Members: []*ast.EnumMember{
							{Name: "Get", Value: &ast.EnumValue{Str: qptr("GET")}},
							{Name: "Post", Value: &ast.EnumValue{Str: qptr("POST")}},
						},
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Integer enum", func(t *testing.T) {
		input := `
			enum Priority {
				Low = 1
				High = 10
			}
		`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					Enum: &ast.EnumDecl{
						Name: "Priority",
						Members: []*ast.EnumMember{
							{Name: "Low", Value: &ast.EnumValue{Int: ptr("1")}},
							{Name: "High", Value: &ast.EnumValue{Int: ptr("10")}},
						},
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Enum with docstring", func(t *testing.T) {
		input := `
			""" Order status enum """
			enum OrderStatus {
				Pending
			}
		`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					Enum: &ast.EnumDecl{
						Docstring: &ast.Docstring{Value: " Order status enum "},
						Name:      "OrderStatus",
						Members: []*ast.EnumMember{
							{Name: "Pending"},
						},
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Deprecated enum", func(t *testing.T) {
		input := `
			deprecated enum OldStatus {
				Active
			}
		`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					Enum: &ast.EnumDecl{
						Deprecated: &ast.Deprecated{},
						Name:       "OldStatus",
						Members: []*ast.EnumMember{
							{Name: "Active"},
						},
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})
}

////////////////
// PATTERNS   //
////////////////

func TestParserPatternDecl(t *testing.T) {
	t.Run("Basic pattern", func(t *testing.T) {
		input := `pattern UserEventSubject = "events.users.{userId}"`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					Pattern: &ast.PatternDecl{
						Name:  "UserEventSubject",
						Value: "events.users.{userId}",
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Pattern with docstring", func(t *testing.T) {
		input := `
			""" Redis cache key pattern """
			pattern SessionCacheKey = "cache:session:{sessionId}"
		`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					Pattern: &ast.PatternDecl{
						Docstring: &ast.Docstring{Value: " Redis cache key pattern "},
						Name:      "SessionCacheKey",
						Value:     "cache:session:{sessionId}",
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Deprecated pattern", func(t *testing.T) {
		input := `deprecated pattern OldQueueName = "legacy.{id}"`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					Pattern: &ast.PatternDecl{
						Deprecated: &ast.Deprecated{},
						Name:       "OldQueueName",
						Value:      "legacy.{id}",
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})
}

////////////////
// TYPES      //
////////////////

func TestParserTypeDecl(t *testing.T) {
	t.Run("Minimum type declaration", func(t *testing.T) {
		input := `
			type MyType {
				field: string
			}
		`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					Type: &ast.TypeDecl{
						Name: "MyType",
						Children: []*ast.TypeDeclChild{
							{
								Field: &ast.Field{
									Name: "field",
									Type: ast.FieldType{
										Base: &ast.FieldTypeBase{Named: ptr("string")},
									},
								},
							},
						},
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Type with docstring", func(t *testing.T) {
		input := `
			""" My type description """
			type MyType {
				field: string
			}
		`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					Type: &ast.TypeDecl{
						Docstring: &ast.Docstring{Value: " My type description "},
						Name:      "MyType",
						Children: []*ast.TypeDeclChild{
							{
								Field: &ast.Field{
									Name: "field",
									Type: ast.FieldType{
										Base: &ast.FieldTypeBase{Named: ptr("string")},
									},
								},
							},
						},
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Deprecated type", func(t *testing.T) {
		input := `
			deprecated type MyType {
				field: string
			}
		`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					Type: &ast.TypeDecl{
						Deprecated: &ast.Deprecated{},
						Name:       "MyType",
						Children: []*ast.TypeDeclChild{
							{
								Field: &ast.Field{
									Name: "field",
									Type: ast.FieldType{
										Base: &ast.FieldTypeBase{Named: ptr("string")},
									},
								},
							},
						},
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Deprecated type with message", func(t *testing.T) {
		input := `
			deprecated("Use NewType instead")
			type MyType {
				field: string
			}
		`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					Type: &ast.TypeDecl{
						Deprecated: &ast.Deprecated{
							Message: qptr("Use NewType instead"),
						},
						Name: "MyType",
						Children: []*ast.TypeDeclChild{
							{
								Field: &ast.Field{
									Name: "field",
									Type: ast.FieldType{
										Base: &ast.FieldTypeBase{Named: ptr("string")},
									},
								},
							},
						},
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Type with all primitive types", func(t *testing.T) {
		input := `
			type MyType {
				f1: string
				f2: int
				f3: float
				f4: bool
				f5: datetime
			}
		`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					Type: &ast.TypeDecl{
						Name: "MyType",
						Children: []*ast.TypeDeclChild{
							{Field: &ast.Field{Name: "f1", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}}}},
							{Field: &ast.Field{Name: "f2", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("int")}}}},
							{Field: &ast.Field{Name: "f3", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("float")}}}},
							{Field: &ast.Field{Name: "f4", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("bool")}}}},
							{Field: &ast.Field{Name: "f5", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("datetime")}}}},
						},
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Type with optional fields", func(t *testing.T) {
		input := `
			type MyType {
				required: string
				optional?: string
			}
		`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					Type: &ast.TypeDecl{
						Name: "MyType",
						Children: []*ast.TypeDeclChild{
							{Field: &ast.Field{Name: "required", Optional: false, Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}}}},
							{Field: &ast.Field{Name: "optional", Optional: true, Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}}}},
						},
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Type with array fields", func(t *testing.T) {
		input := `
			type MyType {
				tags: string[]
				scores: int[]
			}
		`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					Type: &ast.TypeDecl{
						Name: "MyType",
						Children: []*ast.TypeDeclChild{
							{Field: &ast.Field{Name: "tags", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}, Dimensions: 1}}},
							{Field: &ast.Field{Name: "scores", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("int")}, Dimensions: 1}}},
						},
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Type with nested inline object", func(t *testing.T) {
		input := `
			type MyType {
				location: {
					lat: float
					lng: float
				}
			}
		`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					Type: &ast.TypeDecl{
						Name: "MyType",
						Children: []*ast.TypeDeclChild{
							{
								Field: &ast.Field{
									Name: "location",
									Type: ast.FieldType{
										Base: &ast.FieldTypeBase{
											Object: &ast.FieldTypeObject{
												Children: []*ast.TypeDeclChild{
													{Field: &ast.Field{Name: "lat", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("float")}}}},
													{Field: &ast.Field{Name: "lng", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("float")}}}},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Type with map fields", func(t *testing.T) {
		input := `
			type MyType {
				counts: map<int>
				users: map<User>
			}
		`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					Type: &ast.TypeDecl{
						Name: "MyType",
						Children: []*ast.TypeDeclChild{
							{
								Field: &ast.Field{
									Name: "counts",
									Type: ast.FieldType{
										Base: &ast.FieldTypeBase{
											Map: &ast.FieldTypeMap{
												ValueType: &ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("int")}},
											},
										},
									},
								},
							},
							{
								Field: &ast.Field{
									Name: "users",
									Type: ast.FieldType{
										Base: &ast.FieldTypeBase{
											Map: &ast.FieldTypeMap{
												ValueType: &ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("User")}},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Type with spread operator", func(t *testing.T) {
		input := `
			type Article {
				...AuditMetadata
				title: string
			}
		`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					Type: &ast.TypeDecl{
						Name: "Article",
						Children: []*ast.TypeDeclChild{
							{Spread: &ast.Spread{TypeName: "AuditMetadata"}},
							{Field: &ast.Field{Name: "title", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}}}},
						},
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Type with field docstrings", func(t *testing.T) {
		input := `
			type Product {
				""" The SKU identifier """
				sku: string
			}
		`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					Type: &ast.TypeDecl{
						Name: "Product",
						Children: []*ast.TypeDeclChild{
							{
								Field: &ast.Field{
									Docstring: &ast.Docstring{Value: " The SKU identifier "},
									Name:      "sku",
									Type:      ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}},
								},
							},
						},
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})
}

////////////////
// RPC BLOCKS //
////////////////

func TestParserRPCDecl(t *testing.T) {
	t.Run("Empty RPC block", func(t *testing.T) {
		input := `rpc MyService {}`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					RPC: &ast.RPCDecl{
						Name:     "MyService",
						Children: nil,
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("RPC with docstring", func(t *testing.T) {
		input := `
			""" My service description """
			rpc MyService {}
		`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					RPC: &ast.RPCDecl{
						Docstring: &ast.Docstring{Value: " My service description "},
						Name:      "MyService",
						Children:  nil,
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Deprecated RPC", func(t *testing.T) {
		input := `deprecated rpc OldService {}`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					RPC: &ast.RPCDecl{
						Deprecated: &ast.Deprecated{},
						Name:       "OldService",
						Children:   nil,
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("RPC with procedure", func(t *testing.T) {
		input := `
			rpc MyService {
				proc GetUser {
					input {
						userId: string
					}
					output {
						user: User
					}
				}
			}
		`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					RPC: &ast.RPCDecl{
						Name: "MyService",
						Children: []*ast.RPCChild{
							{
								Proc: &ast.ProcDecl{
									Name: "GetUser",
									Children: []*ast.ProcOrStreamDeclChild{
										{
											Input: &ast.ProcOrStreamDeclChildInput{
												Children: []*ast.InputOutputChild{
													{Field: &ast.Field{Name: "userId", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}}}},
												},
											},
										},
										{
											Output: &ast.ProcOrStreamDeclChildOutput{
												Children: []*ast.InputOutputChild{
													{Field: &ast.Field{Name: "user", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("User")}}}},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("RPC with stream", func(t *testing.T) {
		input := `
			rpc MyService {
				stream NewMessages {
					input {
						channelId: string
					}
					output {
						message: string
					}
				}
			}
		`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					RPC: &ast.RPCDecl{
						Name: "MyService",
						Children: []*ast.RPCChild{
							{
								Stream: &ast.StreamDecl{
									Name: "NewMessages",
									Children: []*ast.ProcOrStreamDeclChild{
										{
											Input: &ast.ProcOrStreamDeclChildInput{
												Children: []*ast.InputOutputChild{
													{Field: &ast.Field{Name: "channelId", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}}}},
												},
											},
										},
										{
											Output: &ast.ProcOrStreamDeclChildOutput{
												Children: []*ast.InputOutputChild{
													{Field: &ast.Field{Name: "message", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}}}},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Deprecated proc inside RPC", func(t *testing.T) {
		input := `
			rpc MyService {
				deprecated proc OldProc {}
			}
		`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					RPC: &ast.RPCDecl{
						Name: "MyService",
						Children: []*ast.RPCChild{
							{
								Proc: &ast.ProcDecl{
									Deprecated: &ast.Deprecated{},
									Name:       "OldProc",
									Children:   nil,
								},
							},
						},
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Proc with spread in input/output", func(t *testing.T) {
		input := `
			rpc Articles {
				proc ListArticles {
					input {
						...PaginationParams
						filter?: string
					}
					output {
						items: Article[]
					}
				}
			}
		`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					RPC: &ast.RPCDecl{
						Name: "Articles",
						Children: []*ast.RPCChild{
							{
								Proc: &ast.ProcDecl{
									Name: "ListArticles",
									Children: []*ast.ProcOrStreamDeclChild{
										{
											Input: &ast.ProcOrStreamDeclChildInput{
												Children: []*ast.InputOutputChild{
													{Spread: &ast.Spread{TypeName: "PaginationParams"}},
													{Field: &ast.Field{Name: "filter", Optional: true, Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}}}},
												},
											},
										},
										{
											Output: &ast.ProcOrStreamDeclChildOutput{
												Children: []*ast.InputOutputChild{
													{Field: &ast.Field{Name: "items", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("Article")}, Dimensions: 1}}},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})
}

////////////////
// COMMENTS   //
////////////////

func TestParserComments(t *testing.T) {
	t.Run("Top-level comments", func(t *testing.T) {
		input := `
			// This is a comment
			type MyType {
				field: string
			}
		`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{Comment: &ast.Comment{Simple: ptr("// This is a comment")}},
				{
					Type: &ast.TypeDecl{
						Name: "MyType",
						Children: []*ast.TypeDeclChild{
							{Field: &ast.Field{Name: "field", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}}}},
						},
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Block comments", func(t *testing.T) {
		input := `
			/* This is a block comment */
			type MyType {
				field: string
			}
		`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{Comment: &ast.Comment{Block: ptr("/* This is a block comment */")}},
				{
					Type: &ast.TypeDecl{
						Name: "MyType",
						Children: []*ast.TypeDeclChild{
							{Field: &ast.Field{Name: "field", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}}}},
						},
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Comments inside type", func(t *testing.T) {
		input := `
			type MyType {
				// Before field
				field: string
				/* After field */
			}
		`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					Type: &ast.TypeDecl{
						Name: "MyType",
						Children: []*ast.TypeDeclChild{
							{Comment: &ast.Comment{Simple: ptr("// Before field")}},
							{Field: &ast.Field{Name: "field", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}}}},
							{Comment: &ast.Comment{Block: ptr("/* After field */")}},
						},
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})
}

////////////////
// DOCSTRINGS //
////////////////

func TestParserDocstrings(t *testing.T) {
	t.Run("Standalone docstring", func(t *testing.T) {
		input := `""" This is a standalone docstring. """`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{Docstring: &ast.Docstring{Value: " This is a standalone docstring. "}},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Multiple standalone docstrings", func(t *testing.T) {
		input := `
			""" First docstring """
			""" Second docstring """
		`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{Docstring: &ast.Docstring{Value: " First docstring "}},
				{Docstring: &ast.Docstring{Value: " Second docstring "}},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Docstring followed by blank line becomes standalone", func(t *testing.T) {
		input := `
			""" This is standalone """

			type MyType {
				field: string
			}
		`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{Docstring: &ast.Docstring{Value: " This is standalone "}},
				{
					Type: &ast.TypeDecl{
						Name: "MyType",
						Children: []*ast.TypeDeclChild{
							{Field: &ast.Field{Name: "field", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}}}},
						},
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Docstring in RPC becomes standalone", func(t *testing.T) {
		input := `
			rpc MyService {
				""" Standalone doc inside RPC """

				proc DoSomething {}
			}
		`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					RPC: &ast.RPCDecl{
						Name: "MyService",
						Children: []*ast.RPCChild{
							{Docstring: &ast.Docstring{Value: " Standalone doc inside RPC "}},
							{Proc: &ast.ProcDecl{Name: "DoSomething", Children: nil}},
						},
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})
}

////////////////
// EDGE CASES //
////////////////

func TestParserEdgeCases(t *testing.T) {
	t.Run("Empty schema", func(t *testing.T) {
		input := ``
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: nil,
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Whitespace only", func(t *testing.T) {
		input := `   
		
		`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: nil,
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Empty type", func(t *testing.T) {
		input := `type EmptyType {}`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{Type: &ast.TypeDecl{Name: "EmptyType", Children: nil}},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Empty enum", func(t *testing.T) {
		input := `enum EmptyEnum {}`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{Enum: &ast.EnumDecl{Name: "EmptyEnum", Members: nil}},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Deeply nested inline objects", func(t *testing.T) {
		input := `
			type Deep {
				level1: {
					level2: {
						value: string
					}
				}
			}
		`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					Type: &ast.TypeDecl{
						Name: "Deep",
						Children: []*ast.TypeDeclChild{
							{
								Field: &ast.Field{
									Name: "level1",
									Type: ast.FieldType{
										Base: &ast.FieldTypeBase{
											Object: &ast.FieldTypeObject{
												Children: []*ast.TypeDeclChild{
													{
														Field: &ast.Field{
															Name: "level2",
															Type: ast.FieldType{
																Base: &ast.FieldTypeBase{
																	Object: &ast.FieldTypeObject{
																		Children: []*ast.TypeDeclChild{
																			{Field: &ast.Field{Name: "value", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}}}},
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Array of inline objects", func(t *testing.T) {
		input := `
			type MyType {
				items: {
					name: string
				}[]
			}
		`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					Type: &ast.TypeDecl{
						Name: "MyType",
						Children: []*ast.TypeDeclChild{
							{
								Field: &ast.Field{
									Name: "items",
									Type: ast.FieldType{
										Base: &ast.FieldTypeBase{
											Object: &ast.FieldTypeObject{
												Children: []*ast.TypeDeclChild{
													{Field: &ast.Field{Name: "name", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}}}},
												},
											},
										},
										Dimensions: 1,
									},
								},
							},
						},
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Map of arrays", func(t *testing.T) {
		input := `
			type MyType {
				data: map<string[]>
			}
		`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					Type: &ast.TypeDecl{
						Name: "MyType",
						Children: []*ast.TypeDeclChild{
							{
								Field: &ast.Field{
									Name: "data",
									Type: ast.FieldType{
										Base: &ast.FieldTypeBase{
											Map: &ast.FieldTypeMap{
												ValueType: &ast.FieldType{
													Base:       &ast.FieldTypeBase{Named: ptr("string")},
													Dimensions: 1,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Custom type reference", func(t *testing.T) {
		input := `
			type Profile {
				user: User
			}
		`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					Type: &ast.TypeDecl{
						Name: "Profile",
						Children: []*ast.TypeDeclChild{
							{Field: &ast.Field{Name: "user", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("User")}}}},
						},
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})
}

////////////////
// MULTI-DIMENSIONAL ARRAYS //
////////////////

func TestParserMultiDimensionalArrays(t *testing.T) {
	t.Run("2D array of primitives", func(t *testing.T) {
		input := `
			type Matrix {
				data: int[][]
			}
		`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					Type: &ast.TypeDecl{
						Name: "Matrix",
						Children: []*ast.TypeDeclChild{
							{
								Field: &ast.Field{
									Name: "data",
									Type: ast.FieldType{
										Base:       &ast.FieldTypeBase{Named: ptr("int")},
										Dimensions: 2,
									},
								},
							},
						},
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("3D array of strings", func(t *testing.T) {
		input := `
			type Tensor {
				values: string[][][]
			}
		`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					Type: &ast.TypeDecl{
						Name: "Tensor",
						Children: []*ast.TypeDeclChild{
							{
								Field: &ast.Field{
									Name: "values",
									Type: ast.FieldType{
										Base:       &ast.FieldTypeBase{Named: ptr("string")},
										Dimensions: 3,
									},
								},
							},
						},
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("2D array of custom types", func(t *testing.T) {
		input := `
			type Grid {
				cells: Cell[][]
			}
		`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					Type: &ast.TypeDecl{
						Name: "Grid",
						Children: []*ast.TypeDeclChild{
							{
								Field: &ast.Field{
									Name: "cells",
									Type: ast.FieldType{
										Base:       &ast.FieldTypeBase{Named: ptr("Cell")},
										Dimensions: 2,
									},
								},
							},
						},
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Mixed array dimensions in same type", func(t *testing.T) {
		input := `
			type Container {
				single: int
				oneDim: int[]
				twoDim: int[][]
				threeDim: int[][][]
			}
		`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					Type: &ast.TypeDecl{
						Name: "Container",
						Children: []*ast.TypeDeclChild{
							{Field: &ast.Field{Name: "single", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("int")}}}},
							{Field: &ast.Field{Name: "oneDim", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("int")}, Dimensions: 1}}},
							{Field: &ast.Field{Name: "twoDim", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("int")}, Dimensions: 2}}},
							{Field: &ast.Field{Name: "threeDim", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("int")}, Dimensions: 3}}},
						},
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("2D array of inline objects", func(t *testing.T) {
		input := `
			type Board {
				squares: {
					value: int
					color: string
				}[][]
			}
		`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					Type: &ast.TypeDecl{
						Name: "Board",
						Children: []*ast.TypeDeclChild{
							{
								Field: &ast.Field{
									Name: "squares",
									Type: ast.FieldType{
										Base: &ast.FieldTypeBase{
											Object: &ast.FieldTypeObject{
												Children: []*ast.TypeDeclChild{
													{Field: &ast.Field{Name: "value", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("int")}}}},
													{Field: &ast.Field{Name: "color", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}}}},
												},
											},
										},
										Dimensions: 2,
									},
								},
							},
						},
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Map with multi-dimensional array value", func(t *testing.T) {
		input := `
			type Cache {
				matrices: map<int[][]>[][][]
			}
		`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					Type: &ast.TypeDecl{
						Name: "Cache",
						Children: []*ast.TypeDeclChild{
							{
								Field: &ast.Field{
									Name: "matrices",
									Type: ast.FieldType{
										Base: &ast.FieldTypeBase{
											Map: &ast.FieldTypeMap{
												ValueType: &ast.FieldType{
													Base:       &ast.FieldTypeBase{Named: ptr("int")},
													Dimensions: 2,
												},
											},
										},
										Dimensions: 3,
									},
								},
							},
						},
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Optional multi-dimensional array", func(t *testing.T) {
		input := `
			type OptionalMatrix {
				data?: float[][]
			}
		`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					Type: &ast.TypeDecl{
						Name: "OptionalMatrix",
						Children: []*ast.TypeDeclChild{
							{
								Field: &ast.Field{
									Name:     "data",
									Optional: true,
									Type: ast.FieldType{
										Base:       &ast.FieldTypeBase{Named: ptr("float")},
										Dimensions: 2,
									},
								},
							},
						},
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("RPC proc with multi-dimensional arrays", func(t *testing.T) {
		input := `
			rpc MatrixService {
				proc Multiply {
					input {
						a: int[][]
						b: int[][]
					}
					output {
						result: int[][]
					}
				}
			}
		`
		parsed, err := ParserInstance.ParseString("schema.ufo", input)
		require.NoError(t, err)

		expected := &ast.Schema{
			Children: []*ast.SchemaChild{
				{
					RPC: &ast.RPCDecl{
						Name: "MatrixService",
						Children: []*ast.RPCChild{
							{
								Proc: &ast.ProcDecl{
									Name: "Multiply",
									Children: []*ast.ProcOrStreamDeclChild{
										{
											Input: &ast.ProcOrStreamDeclChildInput{
												Children: []*ast.InputOutputChild{
													{Field: &ast.Field{Name: "a", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("int")}, Dimensions: 2}}},
													{Field: &ast.Field{Name: "b", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("int")}, Dimensions: 2}}},
												},
											},
										},
										{
											Output: &ast.ProcOrStreamDeclChildOutput{
												Children: []*ast.InputOutputChild{
													{Field: &ast.Field{Name: "result", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("int")}, Dimensions: 2}}},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})
}

//////////////////////
// COMPLETE SCHEMA  //
//////////////////////

// TestParserCompleteSchema is a comprehensive test that covers ALL syntax features
// of the UFO RPC IDL specification. It tests:
// - Includes (multiple)
// - Comments (simple and block)
// - Standalone docstrings (at schema level and inside RPC)
// - Constants (string, int, float, bool true/false, with docstrings, deprecated, deprecated with message)
// - Enums (string implicit, string explicit, integer, with docstrings, deprecated)
// - Patterns (with docstrings, deprecated)
// - Types (all primitives: string/int/float/bool/datetime, arrays, maps, inline objects, spreads, optional fields, field docstrings, deprecated)
// - RPC (with docstrings, deprecated)
// - Proc (with input/output, field docstrings, spreads in input/output, deprecated)
// - Stream (with input/output, field docstrings, deprecated)
func TestParserCompleteSchema(t *testing.T) {
	input := `
		include "./common.ufo"
		include "./auth.ufo"

		/*
		  This is a block comment at the schema level.
		  It can span multiple lines.
		*/

		""" This is a standalone docstring at schema level. """

		"""
		This is a multi-line standalone docstring.
		It can contain markdown formatting.
		"""

		""" String constant documentation. """
		const API_VERSION = "1.0.0"

		""" Integer constant documentation. """
		const MAX_PAGE_SIZE = 100

		""" Float constant documentation. """
		const DEFAULT_TAX_RATE = 0.21

		""" Boolean true constant. """
		const FEATURE_ENABLED = true

		""" Boolean false constant. """
		const MAINTENANCE_MODE = false

		deprecated
		const OLD_LIMIT = 50

		deprecated("Use NEW_TIMEOUT instead")
		const OLD_TIMEOUT = 30

		""" String enum with implicit values. """
		enum OrderStatus {
			Pending
			Processing
			Shipped // inline comment
			Delivered
			Cancelled
		}

		""" String enum with explicit values. """
		enum HttpMethod {
			Get = "GET"
			Post = "POST"
			Put = "PUT"
			Delete = "DELETE"
		}

		""" Integer enum. """
		enum Priority {
			Low = 1
			Medium = 2
			High = 3
			Critical = 10
		}

		deprecated
		enum OldStatus {
			Active
			Inactive
		}

		deprecated("Use NewCategory instead")
		enum LegacyCategory {
			TypeA = "A"
			TypeB = "B"
		}

		""" NATS subject for product events. """
		pattern ProductEventSubject = "events.products.{productId}.{eventType}"

		""" Redis cache key for sessions. """
		pattern SessionCacheKey = "cache:session:{sessionId}"

		deprecated
		pattern OldQueueName = "legacy.{id}"

		deprecated("Use NewPattern instead")
		pattern DeprecatedPattern = "old.{value}"

		""" Base audit metadata for all entities. """
		type AuditMetadata {
			id: string
			createdAt: datetime
			updatedAt: datetime
		}

		""" Pagination parameters for list requests. """
		type PaginationParams {
			page: int
			limit: int
		}

		""" Pagination response metadata. """
		type PaginatedResponse {
			totalItems: int
			totalPages: int
			currentPage: int
		}

		"""
		Comprehensive type demonstrating all field types.
		"""
		type CompleteExample {
			// Spread at the beginning
			...AuditMetadata

			""" String field documentation. """
			name: string

			""" Integer field. """
			count: int

			""" Float field. """
			price: float

			""" Boolean field. """
			isActive: bool

			""" Datetime field. """
			scheduledAt: datetime

			""" Optional string field. """
			nickname?: string

			""" Array of strings. """
			tags: string[]

			""" Optional array. """
			categories?: string[]

			""" Array of custom types. """
			items: OrderStatus[]

			""" Map with string values. """
			metadata: map<string>

			""" Map with integer values. """
			scores: map<int>

			""" Map with custom type values. """
			priorities: map<Priority>

			""" Inline object field. """
			location: {
				latitude: float
				longitude: float
			}

			""" Optional inline object. """
			address?: {
				street: string
				city: string
				country: string
			}

			""" Array of inline objects. """
			coordinates: {
				x: int
				y: int
			}[]

			""" 2D Matrix of integers. """
			matrix: int[][]

			""" 3D Tensor of floats. """
			tensor: float[][][]

			""" 2D array of inline objects. """
			grid: {
				row: int
				col: int
			}[][]

			""" Deeply nested inline object. """
			config: {
				display: {
					theme: string
					fontSize: int
				}
				notifications: {
					email: bool
					push: bool
				}
			}
		}

		deprecated
		type LegacyUser {
			username: string
		}

		deprecated("Use UserV2 instead")
		type OldUser {
			name: string
			email: string
		}

		"""
		Catalog Service

		Provides operations for managing products and browsing the catalog.
		This is a multi-line docstring for the RPC.
		"""
		rpc Catalog {
			"""
			# Product Lifecycle
			This is a standalone docstring inside RPC.
			It documents a section of related procedures.
			"""

			""" Creates a new product in the system. """
			proc CreateProduct {
				input {
					""" The product to create. """
					product: CompleteExample
				}

				output {
					""" Whether the operation succeeded. """
					success: bool
					""" The ID of the created product. """
					productId: string
				}
			}

			""" Retrieves a product by ID with its reviews. """
			proc GetProduct {
				input {
					productId: string
				}

				output {
					product: CompleteExample
					reviews: {
						rating: int
						comment: string
						userId: string
					}[]
				}
			}

			""" Lists products with pagination and filtering. """
			proc ListProducts {
				input {
					// Spread in input block
					...PaginationParams
					filterByStatus?: OrderStatus
				}

				output {
					// Spread in output block
					...PaginatedResponse
					items: CompleteExample[]
				}
			}

			deprecated
			proc OldGetProduct {
				input {
					id: string
				}
				output {
					name: string
				}
			}

			deprecated("Use GetProductV2 instead")
			proc DeprecatedGetProduct {
				input {
					legacyId: string
				}
				output {
					data: string
				}
			}

			// A comment between proc and stream

			""" Subscribes to product updates. """
			stream ProductUpdates {
				input {
					""" The product ID to watch. """
					productId: string
					""" Optional filter by event type. """
					eventType?: string
				}

				output {
					""" The type of update. """
					updateType: string
					""" The updated product data. """
					product: CompleteExample
					""" Server timestamp. """
					timestamp: datetime
				}
			}

			deprecated
			stream OldStream {
				input {
					id: string
				}
				output {
					data: string
				}
			}

			deprecated("Use NewStream instead")
			stream LegacyStream {
				input {
					legacyId: string
				}
				output {
					result: bool
				}
			}
		}

		"""
		Chat Service

		Real-time messaging capabilities with procedures and streams.
		"""
		rpc Chat {
			""" Sends a message to a chat room. """
			proc SendMessage {
				input {
					chatId: string
					message: string
					attachments?: {
						url: string
						mimeType: string
					}[]
				}

				output {
					messageId: string
					timestamp: datetime
				}
			}

			""" Subscribes to new messages in a chat room. """
			stream NewMessage {
				input {
					chatId: string
					sinceTimestamp?: datetime
				}

				output {
					id: string
					senderId: string
					message: string
					timestamp: datetime
				}
			}
		}

		deprecated
		rpc OldService {
			proc DoSomething {
				input {
					value: string
				}
				output {
					result: bool
				}
			}
		}

		deprecated("Use NewAPI instead")
		rpc LegacyAPI {
			proc LegacyCall {
				input {
					param: string
				}
				output {
					response: string
				}
			}
		}
	`
	parsed, err := ParserInstance.ParseString("schema.ufo", input)
	require.NoError(t, err)

	// Note: EnumValue.Int is stored as *string (the lexer returns the token value as string)
	// So we just use ptr("1") for integer enum values

	expected := &ast.Schema{
		Children: []*ast.SchemaChild{
			// ==================== INCLUDES ====================
			{Include: &ast.Include{Path: "./common.ufo"}},
			{Include: &ast.Include{Path: "./auth.ufo"}},

			// ==================== BLOCK COMMENT ====================
			{Comment: &ast.Comment{Block: ptr(`/*
		  This is a block comment at the schema level.
		  It can span multiple lines.
		*/`)}},

			// ==================== STANDALONE DOCSTRINGS ====================
			{Docstring: &ast.Docstring{Value: " This is a standalone docstring at schema level. "}},
			{Docstring: &ast.Docstring{Value: `
		This is a multi-line standalone docstring.
		It can contain markdown formatting.
		`}},

			// ==================== CONSTANTS ====================
			// String constant
			{
				Const: &ast.ConstDecl{
					Docstring: &ast.Docstring{Value: " String constant documentation. "},
					Name:      "API_VERSION",
					Value:     &ast.ConstValue{Str: qptr("1.0.0")},
				},
			},
			// Integer constant
			{
				Const: &ast.ConstDecl{
					Docstring: &ast.Docstring{Value: " Integer constant documentation. "},
					Name:      "MAX_PAGE_SIZE",
					Value:     &ast.ConstValue{Int: ptr("100")},
				},
			},
			// Float constant
			{
				Const: &ast.ConstDecl{
					Docstring: &ast.Docstring{Value: " Float constant documentation. "},
					Name:      "DEFAULT_TAX_RATE",
					Value:     &ast.ConstValue{Float: ptr("0.21")},
				},
			},
			// Boolean true constant
			{
				Const: &ast.ConstDecl{
					Docstring: &ast.Docstring{Value: " Boolean true constant. "},
					Name:      "FEATURE_ENABLED",
					Value:     &ast.ConstValue{True: true},
				},
			},
			// Boolean false constant
			{
				Const: &ast.ConstDecl{
					Docstring: &ast.Docstring{Value: " Boolean false constant. "},
					Name:      "MAINTENANCE_MODE",
					Value:     &ast.ConstValue{False: true},
				},
			},
			// Deprecated constant (no message)
			{
				Const: &ast.ConstDecl{
					Deprecated: &ast.Deprecated{},
					Name:       "OLD_LIMIT",
					Value:      &ast.ConstValue{Int: ptr("50")},
				},
			},
			// Deprecated constant (with message)
			{
				Const: &ast.ConstDecl{
					Deprecated: &ast.Deprecated{Message: qptr("Use NEW_TIMEOUT instead")},
					Name:       "OLD_TIMEOUT",
					Value:      &ast.ConstValue{Int: ptr("30")},
				},
			},

			// ==================== ENUMERATIONS ====================
			// String enum with implicit values
			{
				Enum: &ast.EnumDecl{
					Docstring: &ast.Docstring{Value: " String enum with implicit values. "},
					Name:      "OrderStatus",
					Members: []*ast.EnumMember{
						{Name: "Pending"},
						{Name: "Processing"},
						{Name: "Shipped"},
						{Comment: &ast.Comment{Simple: ptr("// inline comment")}},
						{Name: "Delivered"},
						{Name: "Cancelled"},
					},
				},
			},
			// String enum with explicit values
			{
				Enum: &ast.EnumDecl{
					Docstring: &ast.Docstring{Value: " String enum with explicit values. "},
					Name:      "HttpMethod",
					Members: []*ast.EnumMember{
						{Name: "Get", Value: &ast.EnumValue{Str: qptr("GET")}},
						{Name: "Post", Value: &ast.EnumValue{Str: qptr("POST")}},
						{Name: "Put", Value: &ast.EnumValue{Str: qptr("PUT")}},
						{Name: "Delete", Value: &ast.EnumValue{Str: qptr("DELETE")}},
					},
				},
			},
			// Integer enum
			{
				Enum: &ast.EnumDecl{
					Docstring: &ast.Docstring{Value: " Integer enum. "},
					Name:      "Priority",
					Members: []*ast.EnumMember{
						{Name: "Low", Value: &ast.EnumValue{Int: ptr("1")}},
						{Name: "Medium", Value: &ast.EnumValue{Int: ptr("2")}},
						{Name: "High", Value: &ast.EnumValue{Int: ptr("3")}},
						{Name: "Critical", Value: &ast.EnumValue{Int: ptr("10")}},
					},
				},
			},
			// Deprecated enum (no message)
			{
				Enum: &ast.EnumDecl{
					Deprecated: &ast.Deprecated{},
					Name:       "OldStatus",
					Members: []*ast.EnumMember{
						{Name: "Active"},
						{Name: "Inactive"},
					},
				},
			},
			// Deprecated enum (with message)
			{
				Enum: &ast.EnumDecl{
					Deprecated: &ast.Deprecated{Message: qptr("Use NewCategory instead")},
					Name:       "LegacyCategory",
					Members: []*ast.EnumMember{
						{Name: "TypeA", Value: &ast.EnumValue{Str: qptr("A")}},
						{Name: "TypeB", Value: &ast.EnumValue{Str: qptr("B")}},
					},
				},
			},

			// ==================== PATTERNS ====================
			// Pattern with docstring
			{
				Pattern: &ast.PatternDecl{
					Docstring: &ast.Docstring{Value: " NATS subject for product events. "},
					Name:      "ProductEventSubject",
					Value:     "events.products.{productId}.{eventType}",
				},
			},
			// Another pattern with docstring
			{
				Pattern: &ast.PatternDecl{
					Docstring: &ast.Docstring{Value: " Redis cache key for sessions. "},
					Name:      "SessionCacheKey",
					Value:     "cache:session:{sessionId}",
				},
			},
			// Deprecated pattern (no message)
			{
				Pattern: &ast.PatternDecl{
					Deprecated: &ast.Deprecated{},
					Name:       "OldQueueName",
					Value:      "legacy.{id}",
				},
			},
			// Deprecated pattern (with message)
			{
				Pattern: &ast.PatternDecl{
					Deprecated: &ast.Deprecated{Message: qptr("Use NewPattern instead")},
					Name:       "DeprecatedPattern",
					Value:      "old.{value}",
				},
			},

			// ==================== TYPES ====================
			// AuditMetadata type
			{
				Type: &ast.TypeDecl{
					Docstring: &ast.Docstring{Value: " Base audit metadata for all entities. "},
					Name:      "AuditMetadata",
					Children: []*ast.TypeDeclChild{
						{Field: &ast.Field{Name: "id", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}}}},
						{Field: &ast.Field{Name: "createdAt", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("datetime")}}}},
						{Field: &ast.Field{Name: "updatedAt", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("datetime")}}}},
					},
				},
			},
			// PaginationParams type
			{
				Type: &ast.TypeDecl{
					Docstring: &ast.Docstring{Value: " Pagination parameters for list requests. "},
					Name:      "PaginationParams",
					Children: []*ast.TypeDeclChild{
						{Field: &ast.Field{Name: "page", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("int")}}}},
						{Field: &ast.Field{Name: "limit", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("int")}}}},
					},
				},
			},
			// PaginatedResponse type
			{
				Type: &ast.TypeDecl{
					Docstring: &ast.Docstring{Value: " Pagination response metadata. "},
					Name:      "PaginatedResponse",
					Children: []*ast.TypeDeclChild{
						{Field: &ast.Field{Name: "totalItems", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("int")}}}},
						{Field: &ast.Field{Name: "totalPages", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("int")}}}},
						{Field: &ast.Field{Name: "currentPage", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("int")}}}},
					},
				},
			},
			// CompleteExample type - demonstrating all field types
			{
				Type: &ast.TypeDecl{
					Docstring: &ast.Docstring{Value: `
		Comprehensive type demonstrating all field types.
		`},
					Name: "CompleteExample",
					Children: []*ast.TypeDeclChild{
						// Comment
						{Comment: &ast.Comment{Simple: ptr("// Spread at the beginning")}},
						// Spread
						{Spread: &ast.Spread{TypeName: "AuditMetadata"}},
						// String field with docstring
						{Field: &ast.Field{
							Docstring: &ast.Docstring{Value: " String field documentation. "},
							Name:      "name",
							Type:      ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}},
						}},
						// Integer field
						{Field: &ast.Field{
							Docstring: &ast.Docstring{Value: " Integer field. "},
							Name:      "count",
							Type:      ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("int")}},
						}},
						// Float field
						{Field: &ast.Field{
							Docstring: &ast.Docstring{Value: " Float field. "},
							Name:      "price",
							Type:      ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("float")}},
						}},
						// Boolean field
						{Field: &ast.Field{
							Docstring: &ast.Docstring{Value: " Boolean field. "},
							Name:      "isActive",
							Type:      ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("bool")}},
						}},
						// Datetime field
						{Field: &ast.Field{
							Docstring: &ast.Docstring{Value: " Datetime field. "},
							Name:      "scheduledAt",
							Type:      ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("datetime")}},
						}},
						// Optional string field
						{Field: &ast.Field{
							Docstring: &ast.Docstring{Value: " Optional string field. "},
							Name:      "nickname",
							Optional:  true,
							Type:      ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}},
						}},
						// Array of strings
						{Field: &ast.Field{
							Docstring: &ast.Docstring{Value: " Array of strings. "},
							Name:      "tags",
							Type:      ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}, Dimensions: 1},
						}},
						// Optional array
						{Field: &ast.Field{
							Docstring: &ast.Docstring{Value: " Optional array. "},
							Name:      "categories",
							Optional:  true,
							Type:      ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}, Dimensions: 1},
						}},
						// Array of custom types
						{Field: &ast.Field{
							Docstring: &ast.Docstring{Value: " Array of custom types. "},
							Name:      "items",
							Type:      ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("OrderStatus")}, Dimensions: 1},
						}},
						// Map with string values
						{Field: &ast.Field{
							Docstring: &ast.Docstring{Value: " Map with string values. "},
							Name:      "metadata",
							Type: ast.FieldType{Base: &ast.FieldTypeBase{Map: &ast.FieldTypeMap{
								ValueType: &ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}},
							}}},
						}},
						// Map with integer values
						{Field: &ast.Field{
							Docstring: &ast.Docstring{Value: " Map with integer values. "},
							Name:      "scores",
							Type: ast.FieldType{Base: &ast.FieldTypeBase{Map: &ast.FieldTypeMap{
								ValueType: &ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("int")}},
							}}},
						}},
						// Map with custom type values
						{Field: &ast.Field{
							Docstring: &ast.Docstring{Value: " Map with custom type values. "},
							Name:      "priorities",
							Type: ast.FieldType{Base: &ast.FieldTypeBase{Map: &ast.FieldTypeMap{
								ValueType: &ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("Priority")}},
							}}},
						}},
						// Inline object field
						{Field: &ast.Field{
							Docstring: &ast.Docstring{Value: " Inline object field. "},
							Name:      "location",
							Type: ast.FieldType{Base: &ast.FieldTypeBase{Object: &ast.FieldTypeObject{
								Children: []*ast.TypeDeclChild{
									{Field: &ast.Field{Name: "latitude", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("float")}}}},
									{Field: &ast.Field{Name: "longitude", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("float")}}}},
								},
							}}},
						}},
						// Optional inline object
						{Field: &ast.Field{
							Docstring: &ast.Docstring{Value: " Optional inline object. "},
							Name:      "address",
							Optional:  true,
							Type: ast.FieldType{Base: &ast.FieldTypeBase{Object: &ast.FieldTypeObject{
								Children: []*ast.TypeDeclChild{
									{Field: &ast.Field{Name: "street", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}}}},
									{Field: &ast.Field{Name: "city", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}}}},
									{Field: &ast.Field{Name: "country", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}}}},
								},
							}}},
						}},
						// Array of inline objects
						{Field: &ast.Field{
							Docstring: &ast.Docstring{Value: " Array of inline objects. "},
							Name:      "coordinates",
							Type: ast.FieldType{
								Base: &ast.FieldTypeBase{Object: &ast.FieldTypeObject{
									Children: []*ast.TypeDeclChild{
										{Field: &ast.Field{Name: "x", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("int")}}}},
										{Field: &ast.Field{Name: "y", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("int")}}}},
									},
								}},
								Dimensions: 1,
							},
						}},
						// 2D matrix
						{Field: &ast.Field{
							Docstring: &ast.Docstring{Value: " 2D Matrix of integers. "},
							Name:      "matrix",
							Type:      ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("int")}, Dimensions: 2},
						}},
						// 3D tensor
						{Field: &ast.Field{
							Docstring: &ast.Docstring{Value: " 3D Tensor of floats. "},
							Name:      "tensor",
							Type:      ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("float")}, Dimensions: 3},
						}},
						// 2D array of inline objects
						{Field: &ast.Field{
							Docstring: &ast.Docstring{Value: " 2D array of inline objects. "},
							Name:      "grid",
							Type: ast.FieldType{
								Base: &ast.FieldTypeBase{Object: &ast.FieldTypeObject{
									Children: []*ast.TypeDeclChild{
										{Field: &ast.Field{Name: "row", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("int")}}}},
										{Field: &ast.Field{Name: "col", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("int")}}}},
									},
								}},
								Dimensions: 2,
							},
						}},
						// Deeply nested inline object
						{Field: &ast.Field{
							Docstring: &ast.Docstring{Value: " Deeply nested inline object. "},
							Name:      "config",
							Type: ast.FieldType{Base: &ast.FieldTypeBase{Object: &ast.FieldTypeObject{
								Children: []*ast.TypeDeclChild{
									{Field: &ast.Field{
										Name: "display",
										Type: ast.FieldType{Base: &ast.FieldTypeBase{Object: &ast.FieldTypeObject{
											Children: []*ast.TypeDeclChild{
												{Field: &ast.Field{Name: "theme", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}}}},
												{Field: &ast.Field{Name: "fontSize", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("int")}}}},
											},
										}}},
									}},
									{Field: &ast.Field{
										Name: "notifications",
										Type: ast.FieldType{Base: &ast.FieldTypeBase{Object: &ast.FieldTypeObject{
											Children: []*ast.TypeDeclChild{
												{Field: &ast.Field{Name: "email", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("bool")}}}},
												{Field: &ast.Field{Name: "push", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("bool")}}}},
											},
										}}},
									}},
								},
							}}},
						}},
					},
				},
			},
			// Deprecated type (no message)
			{
				Type: &ast.TypeDecl{
					Deprecated: &ast.Deprecated{},
					Name:       "LegacyUser",
					Children: []*ast.TypeDeclChild{
						{Field: &ast.Field{Name: "username", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}}}},
					},
				},
			},
			// Deprecated type (with message)
			{
				Type: &ast.TypeDecl{
					Deprecated: &ast.Deprecated{Message: qptr("Use UserV2 instead")},
					Name:       "OldUser",
					Children: []*ast.TypeDeclChild{
						{Field: &ast.Field{Name: "name", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}}}},
						{Field: &ast.Field{Name: "email", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}}}},
					},
				},
			},

			// ==================== RPC SERVICES ====================
			// Catalog RPC
			{
				RPC: &ast.RPCDecl{
					Docstring: &ast.Docstring{Value: `
		Catalog Service

		Provides operations for managing products and browsing the catalog.
		This is a multi-line docstring for the RPC.
		`},
					Name: "Catalog",
					Children: []*ast.RPCChild{
						// Standalone docstring inside RPC
						{Docstring: &ast.Docstring{Value: `
			# Product Lifecycle
			This is a standalone docstring inside RPC.
			It documents a section of related procedures.
			`}},
						// CreateProduct proc
						{
							Proc: &ast.ProcDecl{
								Docstring: &ast.Docstring{Value: " Creates a new product in the system. "},
								Name:      "CreateProduct",
								Children: []*ast.ProcOrStreamDeclChild{
									{
										Input: &ast.ProcOrStreamDeclChildInput{
											Children: []*ast.InputOutputChild{
												{Field: &ast.Field{
													Docstring: &ast.Docstring{Value: " The product to create. "},
													Name:      "product",
													Type:      ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("CompleteExample")}},
												}},
											},
										},
									},
									{
										Output: &ast.ProcOrStreamDeclChildOutput{
											Children: []*ast.InputOutputChild{
												{Field: &ast.Field{
													Docstring: &ast.Docstring{Value: " Whether the operation succeeded. "},
													Name:      "success",
													Type:      ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("bool")}},
												}},
												{Field: &ast.Field{
													Docstring: &ast.Docstring{Value: " The ID of the created product. "},
													Name:      "productId",
													Type:      ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}},
												}},
											},
										},
									},
								},
							},
						},
						// GetProduct proc
						{
							Proc: &ast.ProcDecl{
								Docstring: &ast.Docstring{Value: " Retrieves a product by ID with its reviews. "},
								Name:      "GetProduct",
								Children: []*ast.ProcOrStreamDeclChild{
									{
										Input: &ast.ProcOrStreamDeclChildInput{
											Children: []*ast.InputOutputChild{
												{Field: &ast.Field{
													Name: "productId",
													Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}},
												}},
											},
										},
									},
									{
										Output: &ast.ProcOrStreamDeclChildOutput{
											Children: []*ast.InputOutputChild{
												{Field: &ast.Field{
													Name: "product",
													Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("CompleteExample")}},
												}},
												{Field: &ast.Field{
													Name: "reviews",
													Type: ast.FieldType{
														Base: &ast.FieldTypeBase{Object: &ast.FieldTypeObject{
															Children: []*ast.TypeDeclChild{
																{Field: &ast.Field{Name: "rating", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("int")}}}},
																{Field: &ast.Field{Name: "comment", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}}}},
																{Field: &ast.Field{Name: "userId", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}}}},
															},
														}},
														Dimensions: 1,
													},
												}},
											},
										},
									},
								},
							},
						},
						// ListProducts proc with spreads
						{
							Proc: &ast.ProcDecl{
								Docstring: &ast.Docstring{Value: " Lists products with pagination and filtering. "},
								Name:      "ListProducts",
								Children: []*ast.ProcOrStreamDeclChild{
									{
										Input: &ast.ProcOrStreamDeclChildInput{
											Children: []*ast.InputOutputChild{
												{Comment: &ast.Comment{Simple: ptr("// Spread in input block")}},
												{Spread: &ast.Spread{TypeName: "PaginationParams"}},
												{Field: &ast.Field{
													Name:     "filterByStatus",
													Optional: true,
													Type:     ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("OrderStatus")}},
												}},
											},
										},
									},
									{
										Output: &ast.ProcOrStreamDeclChildOutput{
											Children: []*ast.InputOutputChild{
												{Comment: &ast.Comment{Simple: ptr("// Spread in output block")}},
												{Spread: &ast.Spread{TypeName: "PaginatedResponse"}},
												{Field: &ast.Field{
													Name: "items",
													Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("CompleteExample")}, Dimensions: 1},
												}},
											},
										},
									},
								},
							},
						},
						// Deprecated proc (no message)
						{
							Proc: &ast.ProcDecl{
								Deprecated: &ast.Deprecated{},
								Name:       "OldGetProduct",
								Children: []*ast.ProcOrStreamDeclChild{
									{
										Input: &ast.ProcOrStreamDeclChildInput{
											Children: []*ast.InputOutputChild{
												{Field: &ast.Field{Name: "id", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}}}},
											},
										},
									},
									{
										Output: &ast.ProcOrStreamDeclChildOutput{
											Children: []*ast.InputOutputChild{
												{Field: &ast.Field{Name: "name", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}}}},
											},
										},
									},
								},
							},
						},
						// Deprecated proc (with message)
						{
							Proc: &ast.ProcDecl{
								Deprecated: &ast.Deprecated{Message: qptr("Use GetProductV2 instead")},
								Name:       "DeprecatedGetProduct",
								Children: []*ast.ProcOrStreamDeclChild{
									{
										Input: &ast.ProcOrStreamDeclChildInput{
											Children: []*ast.InputOutputChild{
												{Field: &ast.Field{Name: "legacyId", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}}}},
											},
										},
									},
									{
										Output: &ast.ProcOrStreamDeclChildOutput{
											Children: []*ast.InputOutputChild{
												{Field: &ast.Field{Name: "data", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}}}},
											},
										},
									},
								},
							},
						},
						// Comment between proc and stream
						{Comment: &ast.Comment{Simple: ptr("// A comment between proc and stream")}},
						// ProductUpdates stream
						{
							Stream: &ast.StreamDecl{
								Docstring: &ast.Docstring{Value: " Subscribes to product updates. "},
								Name:      "ProductUpdates",
								Children: []*ast.ProcOrStreamDeclChild{
									{
										Input: &ast.ProcOrStreamDeclChildInput{
											Children: []*ast.InputOutputChild{
												{Field: &ast.Field{
													Docstring: &ast.Docstring{Value: " The product ID to watch. "},
													Name:      "productId",
													Type:      ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}},
												}},
												{Field: &ast.Field{
													Docstring: &ast.Docstring{Value: " Optional filter by event type. "},
													Name:      "eventType",
													Optional:  true,
													Type:      ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}},
												}},
											},
										},
									},
									{
										Output: &ast.ProcOrStreamDeclChildOutput{
											Children: []*ast.InputOutputChild{
												{Field: &ast.Field{
													Docstring: &ast.Docstring{Value: " The type of update. "},
													Name:      "updateType",
													Type:      ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}},
												}},
												{Field: &ast.Field{
													Docstring: &ast.Docstring{Value: " The updated product data. "},
													Name:      "product",
													Type:      ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("CompleteExample")}},
												}},
												{Field: &ast.Field{
													Docstring: &ast.Docstring{Value: " Server timestamp. "},
													Name:      "timestamp",
													Type:      ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("datetime")}},
												}},
											},
										},
									},
								},
							},
						},
						// Deprecated stream (no message)
						{
							Stream: &ast.StreamDecl{
								Deprecated: &ast.Deprecated{},
								Name:       "OldStream",
								Children: []*ast.ProcOrStreamDeclChild{
									{
										Input: &ast.ProcOrStreamDeclChildInput{
											Children: []*ast.InputOutputChild{
												{Field: &ast.Field{Name: "id", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}}}},
											},
										},
									},
									{
										Output: &ast.ProcOrStreamDeclChildOutput{
											Children: []*ast.InputOutputChild{
												{Field: &ast.Field{Name: "data", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}}}},
											},
										},
									},
								},
							},
						},
						// Deprecated stream (with message)
						{
							Stream: &ast.StreamDecl{
								Deprecated: &ast.Deprecated{Message: qptr("Use NewStream instead")},
								Name:       "LegacyStream",
								Children: []*ast.ProcOrStreamDeclChild{
									{
										Input: &ast.ProcOrStreamDeclChildInput{
											Children: []*ast.InputOutputChild{
												{Field: &ast.Field{Name: "legacyId", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}}}},
											},
										},
									},
									{
										Output: &ast.ProcOrStreamDeclChildOutput{
											Children: []*ast.InputOutputChild{
												{Field: &ast.Field{Name: "result", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("bool")}}}},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			// Chat RPC
			{
				RPC: &ast.RPCDecl{
					Docstring: &ast.Docstring{Value: `
		Chat Service

		Real-time messaging capabilities with procedures and streams.
		`},
					Name: "Chat",
					Children: []*ast.RPCChild{
						// SendMessage proc
						{
							Proc: &ast.ProcDecl{
								Docstring: &ast.Docstring{Value: " Sends a message to a chat room. "},
								Name:      "SendMessage",
								Children: []*ast.ProcOrStreamDeclChild{
									{
										Input: &ast.ProcOrStreamDeclChildInput{
											Children: []*ast.InputOutputChild{
												{Field: &ast.Field{Name: "chatId", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}}}},
												{Field: &ast.Field{Name: "message", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}}}},
												{Field: &ast.Field{
													Name:     "attachments",
													Optional: true,
													Type: ast.FieldType{
														Base: &ast.FieldTypeBase{Object: &ast.FieldTypeObject{
															Children: []*ast.TypeDeclChild{
																{Field: &ast.Field{Name: "url", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}}}},
																{Field: &ast.Field{Name: "mimeType", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}}}},
															},
														}},
														Dimensions: 1,
													},
												}},
											},
										},
									},
									{
										Output: &ast.ProcOrStreamDeclChildOutput{
											Children: []*ast.InputOutputChild{
												{Field: &ast.Field{Name: "messageId", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}}}},
												{Field: &ast.Field{Name: "timestamp", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("datetime")}}}},
											},
										},
									},
								},
							},
						},
						// NewMessage stream
						{
							Stream: &ast.StreamDecl{
								Docstring: &ast.Docstring{Value: " Subscribes to new messages in a chat room. "},
								Name:      "NewMessage",
								Children: []*ast.ProcOrStreamDeclChild{
									{
										Input: &ast.ProcOrStreamDeclChildInput{
											Children: []*ast.InputOutputChild{
												{Field: &ast.Field{Name: "chatId", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}}}},
												{Field: &ast.Field{Name: "sinceTimestamp", Optional: true, Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("datetime")}}}},
											},
										},
									},
									{
										Output: &ast.ProcOrStreamDeclChildOutput{
											Children: []*ast.InputOutputChild{
												{Field: &ast.Field{Name: "id", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}}}},
												{Field: &ast.Field{Name: "senderId", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}}}},
												{Field: &ast.Field{Name: "message", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}}}},
												{Field: &ast.Field{Name: "timestamp", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("datetime")}}}},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			// Deprecated RPC (no message)
			{
				RPC: &ast.RPCDecl{
					Deprecated: &ast.Deprecated{},
					Name:       "OldService",
					Children: []*ast.RPCChild{
						{
							Proc: &ast.ProcDecl{
								Name: "DoSomething",
								Children: []*ast.ProcOrStreamDeclChild{
									{
										Input: &ast.ProcOrStreamDeclChildInput{
											Children: []*ast.InputOutputChild{
												{Field: &ast.Field{Name: "value", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}}}},
											},
										},
									},
									{
										Output: &ast.ProcOrStreamDeclChildOutput{
											Children: []*ast.InputOutputChild{
												{Field: &ast.Field{Name: "result", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("bool")}}}},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			// Deprecated RPC (with message)
			{
				RPC: &ast.RPCDecl{
					Deprecated: &ast.Deprecated{Message: qptr("Use NewAPI instead")},
					Name:       "LegacyAPI",
					Children: []*ast.RPCChild{
						{
							Proc: &ast.ProcDecl{
								Name: "LegacyCall",
								Children: []*ast.ProcOrStreamDeclChild{
									{
										Input: &ast.ProcOrStreamDeclChildInput{
											Children: []*ast.InputOutputChild{
												{Field: &ast.Field{Name: "param", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}}}},
											},
										},
									},
									{
										Output: &ast.ProcOrStreamDeclChildOutput{
											Children: []*ast.InputOutputChild{
												{Field: &ast.Field{Name: "response", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}}}},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	testutil.ASTEqualNoPos(t, expected, parsed)
}
