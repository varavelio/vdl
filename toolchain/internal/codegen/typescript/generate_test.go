package typescript

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/analysis"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
	"github.com/varavelio/vdl/toolchain/internal/core/vfs"
)

func TestGenerator_Name(t *testing.T) {
	g := New(&config.TypeScriptConfig{})
	assert.Equal(t, "typescript", g.Name())
}

func parseAndBuildIR(t *testing.T, content string) *ir.Schema {
	fs := vfs.New()
	path := "/test.vdl"
	fs.WriteFileCache(path, []byte(content))

	program, diags := analysis.Analyze(fs, path)
	require.Empty(t, diags, "analysis failed")

	return ir.FromProgram(program)
}

func TestGenerator_Generate_Empty(t *testing.T) {
	g := New(&config.TypeScriptConfig{
		CommonConfig: config.CommonConfig{
			Output: "api.ts",
		},
		ClientConfig: config.ClientConfig{
			GenClient: true,
		},
	})

	schema := parseAndBuildIR(t, "")

	files, err := g.Generate(context.Background(), schema)
	require.NoError(t, err)
	require.Len(t, files, 1)
	assert.Equal(t, "api.ts", files[0].RelativePath)
	// Core types should still be included
	assert.Contains(t, string(files[0].Content), "export type Response<T>")
}

func TestGenerator_Generate_WithTypes(t *testing.T) {
	g := New(&config.TypeScriptConfig{
		CommonConfig: config.CommonConfig{
			Output: "api.ts",
		},
		ClientConfig: config.ClientConfig{
			GenClient: true,
		},
	})

	vdl := `
		type User {
			// Represents a user in the system.
			id: string
			email: string
			age?: int
		}
	`
	schema := parseAndBuildIR(t, vdl)

	files, err := g.Generate(context.Background(), schema)
	require.NoError(t, err)
	require.Len(t, files, 1)

	content := string(files[0].Content)
	assert.Contains(t, content, "export type User = {")
	assert.Contains(t, content, "id: string")
	assert.Contains(t, content, "email: string")
	assert.Contains(t, content, "age?: number")
}

func TestGenerator_Generate_WithEnums(t *testing.T) {
	g := New(&config.TypeScriptConfig{
		CommonConfig: config.CommonConfig{
			Output: "api.ts",
		},
	})

	vdl := `
		enum OrderStatus {
			Pending = "pending"
			Shipped = "shipped"
			Delivered = "delivered"
		}

		enum Priority {
			Low = 1
			Medium = 2
			High = 3
		}
	`
	schema := parseAndBuildIR(t, vdl)

	files, err := g.Generate(context.Background(), schema)
	require.NoError(t, err)
	require.Len(t, files, 1)

	content := string(files[0].Content)

	// String enum
	assert.Contains(t, content, `export type OrderStatus = "pending" | "shipped" | "delivered";`)
	assert.Contains(t, content, `OrderStatusValues`)
	assert.Contains(t, content, `Pending: "pending"`)
	assert.Contains(t, content, `OrderStatusList`)
	assert.Contains(t, content, `function isOrderStatus(value: unknown): value is OrderStatus`)

	// Int enum
	assert.Contains(t, content, `export type Priority = 1 | 2 | 3;`)
	assert.Contains(t, content, `PriorityValues`)
	assert.Contains(t, content, `Low: 1`)
	assert.Contains(t, content, `PriorityList`)
}

func TestGenerator_Generate_WithConstants(t *testing.T) {
	g := New(&config.TypeScriptConfig{
		CommonConfig: config.CommonConfig{
			Output: "api.ts",
		},
	})

	vdl := `
		const MAX_PAGE_SIZE = 100
		const API_VERSION = "1.0.0"
		const DEFAULT_RATE = 0.21
		const ENABLED = true
	`
	schema := parseAndBuildIR(t, vdl)

	files, err := g.Generate(context.Background(), schema)
	require.NoError(t, err)
	require.Len(t, files, 1)

	content := string(files[0].Content)
	assert.Contains(t, content, "export const MAX_PAGE_SIZE: number = 100;")
	assert.Contains(t, content, `export const API_VERSION: string = "1.0.0";`)
	assert.Contains(t, content, "export const DEFAULT_RATE: number = 0.21;")
	assert.Contains(t, content, "export const ENABLED: boolean = true;")
}

func TestGenerator_Generate_WithPatterns(t *testing.T) {
	g := New(&config.TypeScriptConfig{
		CommonConfig: config.CommonConfig{
			Output: "api.ts",
		},
	})

	vdl := `
		pattern UserEventSubject = "events.users.{userId}.{eventType}"
		pattern CacheKey = "cache:{key}"
	`
	schema := parseAndBuildIR(t, vdl)

	files, err := g.Generate(context.Background(), schema)
	require.NoError(t, err)
	require.Len(t, files, 1)

	content := string(files[0].Content)
	assert.Contains(t, content, "export function UserEventSubject(userId: string, eventType: string): string")
	assert.Contains(t, content, "return `events.users.${userId}.${eventType}`")
	assert.Contains(t, content, "export function CacheKey(key: string): string")
	assert.Contains(t, content, "return `cache:${key}`")
}

func TestGenerator_Generate_WithProcedures(t *testing.T) {
	g := New(&config.TypeScriptConfig{
		CommonConfig: config.CommonConfig{
			Output: "api.ts",
		},
		ClientConfig: config.ClientConfig{
			GenClient: true,
		},
	})

	vdl := `
		rpc Users {
			proc GetUser {
				input {
					userId: string
				}
				output {
					id: string
					name: string
				}
			}
		}
	`
	schema := parseAndBuildIR(t, vdl)

	files, err := g.Generate(context.Background(), schema)
	require.NoError(t, err)
	require.Len(t, files, 1)

	content := string(files[0].Content)

	// Check procedure types (with RPC prefix)
	assert.Contains(t, content, "export type UsersGetUserInput = {")
	assert.Contains(t, content, "export type UsersGetUserOutput = {")
	assert.Contains(t, content, "export type UsersGetUserResponse = Response<UsersGetUserOutput>")

	// Check procedure names list
	assert.Contains(t, content, `"Users/GetUser"`)

	// Check client implementation
	assert.Contains(t, content, "class builderUsersGetUser")
	assert.Contains(t, content, "async execute(input: UsersGetUserInput): Promise<UsersGetUserOutput>")
}

func TestGenerator_Generate_WithStreams(t *testing.T) {
	g := New(&config.TypeScriptConfig{
		CommonConfig: config.CommonConfig{
			Output: "api.ts",
		},
		ClientConfig: config.ClientConfig{
			GenClient: true,
		},
	})

	vdl := `
		rpc Chat {
			stream Messages {
				input {
					roomId: string
				}
				output {
					message: string
					timestamp: datetime
				}
			}
		}
	`
	schema := parseAndBuildIR(t, vdl)

	files, err := g.Generate(context.Background(), schema)
	require.NoError(t, err)
	require.Len(t, files, 1)

	content := string(files[0].Content)

	// Check stream types (with RPC prefix)
	assert.Contains(t, content, "export type ChatMessagesInput = {")
	assert.Contains(t, content, "export type ChatMessagesOutput = {")
	assert.Contains(t, content, "export type ChatMessagesResponse = Response<ChatMessagesOutput>")

	// Check stream names list
	assert.Contains(t, content, `"Chat/Messages"`)

	// Check client implementation
	assert.Contains(t, content, "class builderChatMessagesStream")
	assert.Contains(t, content, "execute(input: ChatMessagesInput)")
}

func TestGenerator_Generate_WithComplexTypes(t *testing.T) {
	g := New(&config.TypeScriptConfig{
		CommonConfig: config.CommonConfig{
			Output: "api.ts",
		},
	})

	vdl := `
		type User {
			id: string
		}

		type Product {
			tags: string[]
			matrix: int[][]
			metadata: map<string>
			owner: User
			address: {
				city: string
				zip: string
			}
		}
	`
	schema := parseAndBuildIR(t, vdl)

	files, err := g.Generate(context.Background(), schema)
	require.NoError(t, err)
	require.Len(t, files, 1)

	content := string(files[0].Content)

	// Arrays
	assert.Contains(t, content, "tags: string[]")

	// Multi-dimensional arrays
	assert.Contains(t, content, "matrix: number[][]")

	// Maps
	assert.Contains(t, content, "metadata: Record<string, string>")

	// Custom type
	assert.Contains(t, content, "owner: User")

	// Inline object - should generate a separate type
	assert.Contains(t, content, "export type ProductAddress = {")
	assert.Contains(t, content, "city: string")
	assert.Contains(t, content, "address: ProductAddress")
}

func TestTypeRefToTS(t *testing.T) {
	tests := []struct {
		name   string
		tr     ir.TypeRef
		parent string
		want   string
	}{
		{
			name: "string primitive",
			tr:   ir.TypeRef{Kind: ir.TypeKindPrimitive, Primitive: ir.PrimitiveString},
			want: "string",
		},
		{
			name: "int primitive",
			tr:   ir.TypeRef{Kind: ir.TypeKindPrimitive, Primitive: ir.PrimitiveInt},
			want: "number",
		},
		{
			name: "float primitive",
			tr:   ir.TypeRef{Kind: ir.TypeKindPrimitive, Primitive: ir.PrimitiveFloat},
			want: "number",
		},
		{
			name: "bool primitive",
			tr:   ir.TypeRef{Kind: ir.TypeKindPrimitive, Primitive: ir.PrimitiveBool},
			want: "boolean",
		},
		{
			name: "datetime primitive",
			tr:   ir.TypeRef{Kind: ir.TypeKindPrimitive, Primitive: ir.PrimitiveDatetime},
			want: "Date",
		},
		{
			name: "custom type",
			tr:   ir.TypeRef{Kind: ir.TypeKindType, Type: "User"},
			want: "User",
		},
		{
			name: "enum",
			tr:   ir.TypeRef{Kind: ir.TypeKindEnum, Enum: "OrderStatus"},
			want: "OrderStatus",
		},
		{
			name: "1D array of strings",
			tr: ir.TypeRef{
				Kind:            ir.TypeKindArray,
				ArrayDimensions: 1,
				ArrayItem:       &ir.TypeRef{Kind: ir.TypeKindPrimitive, Primitive: ir.PrimitiveString},
			},
			want: "string[]",
		},
		{
			name: "2D array of ints",
			tr: ir.TypeRef{
				Kind:            ir.TypeKindArray,
				ArrayDimensions: 2,
				ArrayItem:       &ir.TypeRef{Kind: ir.TypeKindPrimitive, Primitive: ir.PrimitiveInt},
			},
			want: "number[][]",
		},
		{
			name: "map of strings",
			tr: ir.TypeRef{
				Kind:     ir.TypeKindMap,
				MapValue: &ir.TypeRef{Kind: ir.TypeKindPrimitive, Primitive: ir.PrimitiveString},
			},
			want: "Record<string, string>",
		},
		{
			name: "map of custom types",
			tr: ir.TypeRef{
				Kind:     ir.TypeKindMap,
				MapValue: &ir.TypeRef{Kind: ir.TypeKindType, Type: "User"},
			},
			want: "Record<string, User>",
		},
		{
			name:   "inline object",
			tr:     ir.TypeRef{Kind: ir.TypeKindObject},
			parent: "UserAddress",
			want:   "UserAddress",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := typeRefToTS(tt.parent, tt.tr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConvertPatternToTemplateLiteral(t *testing.T) {
	tests := []struct {
		name         string
		template     string
		placeholders []string
		want         string
	}{
		{
			name:         "simple pattern",
			template:     "events.{userId}",
			placeholders: []string{"userId"},
			want:         "`events.${userId}`",
		},
		{
			name:         "multiple placeholders",
			template:     "events.users.{userId}.{eventType}",
			placeholders: []string{"userId", "eventType"},
			want:         "`events.users.${userId}.${eventType}`",
		},
		{
			name:         "placeholder at start",
			template:     "{prefix}.suffix",
			placeholders: []string{"prefix"},
			want:         "`${prefix}.suffix`",
		},
		{
			name:         "placeholder at end",
			template:     "prefix.{suffix}",
			placeholders: []string{"suffix"},
			want:         "`prefix.${suffix}`",
		},
		{
			name:         "no placeholders",
			template:     "static.path",
			placeholders: []string{},
			want:         "`static.path`",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertPatternToTemplateLiteral(tt.template, tt.placeholders)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGenerator_Generate_NoClient(t *testing.T) {
	g := New(&config.TypeScriptConfig{
		CommonConfig: config.CommonConfig{
			Output: "api.ts",
		},
		ClientConfig: config.ClientConfig{
			GenClient: false,
		},
	})

	vdl := `
		rpc Users {
			proc GetUser {
				input {
					userId: string
				}
				output {
					id: string
					name: string
				}
			}
		}
	`
	schema := parseAndBuildIR(t, vdl)

	files, err := g.Generate(context.Background(), schema)
	require.NoError(t, err)
	require.Len(t, files, 1)

	content := string(files[0].Content)

	// Procedure types should still be generated
	assert.Contains(t, content, "export type UsersGetUserInput")
	assert.Contains(t, content, "export type UsersGetUserOutput")

	// But client code should NOT be generated
	assert.NotContains(t, content, "class ClientBuilder")
	assert.NotContains(t, content, "class builderUsersGetUser")
}

func TestGenerator_Generate_WithDeprecation(t *testing.T) {
	g := New(&config.TypeScriptConfig{
		CommonConfig: config.CommonConfig{
			Output: "api.ts",
		},
	})

	vdl := `
		deprecated("Use NewUser instead")
		type LegacyUser {
			// Old user type
			id: string
		}

		deprecated
		enum OldStatus {
			Active
		}
	`
	schema := parseAndBuildIR(t, vdl)

	files, err := g.Generate(context.Background(), schema)
	require.NoError(t, err)

	content := string(files[0].Content)
	assert.Contains(t, content, "@deprecated Use NewUser instead")
	assert.Contains(t, content, "@deprecated")
}
