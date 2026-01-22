package typescript

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
)

func TestGenerator_Name(t *testing.T) {
	g := New(&config.TypeScriptConfig{})
	assert.Equal(t, "typescript", g.Name())
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

	schema := &ir.Schema{
		Types:     []ir.Type{},
		Enums:     []ir.Enum{},
		Constants: []ir.Constant{},
		Patterns:  []ir.Pattern{},
		RPCs:      []ir.RPC{},
	}

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

	schema := &ir.Schema{
		Types: []ir.Type{
			{
				Name: "User",
				Doc:  "Represents a user in the system.",
				Fields: []ir.Field{
					{
						Name: "id",
						Type: ir.TypeRef{Kind: ir.TypeKindPrimitive, Primitive: ir.PrimitiveString},
					},
					{
						Name: "email",
						Type: ir.TypeRef{Kind: ir.TypeKindPrimitive, Primitive: ir.PrimitiveString},
					},
					{
						Name:     "age",
						Optional: true,
						Type:     ir.TypeRef{Kind: ir.TypeKindPrimitive, Primitive: ir.PrimitiveInt},
					},
				},
			},
		},
		RPCs: []ir.RPC{},
	}

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

	schema := &ir.Schema{
		Enums: []ir.Enum{
			{
				Name:      "OrderStatus",
				ValueType: ir.EnumValueTypeString,
				Members: []ir.EnumMember{
					{Name: "Pending", Value: "pending"},
					{Name: "Shipped", Value: "shipped"},
					{Name: "Delivered", Value: "delivered"},
				},
			},
			{
				Name:      "Priority",
				ValueType: ir.EnumValueTypeInt,
				Members: []ir.EnumMember{
					{Name: "Low", Value: "1"},
					{Name: "Medium", Value: "2"},
					{Name: "High", Value: "3"},
				},
			},
		},
	}

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

	schema := &ir.Schema{
		Constants: []ir.Constant{
			{
				Name:      "MAX_PAGE_SIZE",
				ValueType: ir.ConstValueTypeInt,
				Value:     "100",
			},
			{
				Name:      "API_VERSION",
				ValueType: ir.ConstValueTypeString,
				Value:     "1.0.0",
			},
			{
				Name:      "DEFAULT_RATE",
				ValueType: ir.ConstValueTypeFloat,
				Value:     "0.21",
			},
			{
				Name:      "ENABLED",
				ValueType: ir.ConstValueTypeBool,
				Value:     "true",
			},
		},
	}

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

	schema := &ir.Schema{
		Patterns: []ir.Pattern{
			{
				Name:         "UserEventSubject",
				Template:     "events.users.{userId}.{eventType}",
				Placeholders: []string{"userId", "eventType"},
			},
			{
				Name:         "CacheKey",
				Template:     "cache:{key}",
				Placeholders: []string{"key"},
			},
		},
	}

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

	schema := &ir.Schema{
		RPCs: []ir.RPC{
			{
				Name: "Users",
				Procs: []ir.Procedure{
					{
						Name: "GetUser",
						Input: []ir.Field{
							{
								Name: "userId",
								Type: ir.TypeRef{Kind: ir.TypeKindPrimitive, Primitive: ir.PrimitiveString},
							},
						},
						Output: []ir.Field{
							{
								Name: "id",
								Type: ir.TypeRef{Kind: ir.TypeKindPrimitive, Primitive: ir.PrimitiveString},
							},
							{
								Name: "name",
								Type: ir.TypeRef{Kind: ir.TypeKindPrimitive, Primitive: ir.PrimitiveString},
							},
						},
					},
				},
			},
		},
	}

	files, err := g.Generate(context.Background(), schema)
	require.NoError(t, err)
	require.Len(t, files, 1)

	content := string(files[0].Content)

	// Check procedure types (with RPC prefix)
	assert.Contains(t, content, "export type UsersGetUserInput = {")
	assert.Contains(t, content, "export type UsersGetUserOutput = {")
	assert.Contains(t, content, "export type UsersGetUserResponse = Response<UsersGetUserOutput>")

	// Check procedure names list
	assert.Contains(t, content, `"users/getUser"`)

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

	schema := &ir.Schema{
		RPCs: []ir.RPC{
			{
				Name: "Chat",
				Streams: []ir.Stream{
					{
						Name: "Messages",
						Input: []ir.Field{
							{
								Name: "roomId",
								Type: ir.TypeRef{Kind: ir.TypeKindPrimitive, Primitive: ir.PrimitiveString},
							},
						},
						Output: []ir.Field{
							{
								Name: "message",
								Type: ir.TypeRef{Kind: ir.TypeKindPrimitive, Primitive: ir.PrimitiveString},
							},
							{
								Name: "timestamp",
								Type: ir.TypeRef{Kind: ir.TypeKindPrimitive, Primitive: ir.PrimitiveDatetime},
							},
						},
					},
				},
			},
		},
	}

	files, err := g.Generate(context.Background(), schema)
	require.NoError(t, err)
	require.Len(t, files, 1)

	content := string(files[0].Content)

	// Check stream types (with RPC prefix)
	assert.Contains(t, content, "export type ChatMessagesInput = {")
	assert.Contains(t, content, "export type ChatMessagesOutput = {")
	assert.Contains(t, content, "export type ChatMessagesResponse = Response<ChatMessagesOutput>")

	// Check stream names list
	assert.Contains(t, content, `"chat/messages"`)

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

	schema := &ir.Schema{
		Types: []ir.Type{
			{
				Name: "Product",
				Fields: []ir.Field{
					// Array
					{
						Name: "tags",
						Type: ir.TypeRef{
							Kind:            ir.TypeKindArray,
							ArrayDimensions: 1,
							ArrayItem:       &ir.TypeRef{Kind: ir.TypeKindPrimitive, Primitive: ir.PrimitiveString},
						},
					},
					// Multi-dimensional array
					{
						Name: "matrix",
						Type: ir.TypeRef{
							Kind:            ir.TypeKindArray,
							ArrayDimensions: 2,
							ArrayItem:       &ir.TypeRef{Kind: ir.TypeKindPrimitive, Primitive: ir.PrimitiveInt},
						},
					},
					// Map
					{
						Name: "metadata",
						Type: ir.TypeRef{
							Kind:     ir.TypeKindMap,
							MapValue: &ir.TypeRef{Kind: ir.TypeKindPrimitive, Primitive: ir.PrimitiveString},
						},
					},
					// Custom type reference
					{
						Name: "owner",
						Type: ir.TypeRef{Kind: ir.TypeKindType, Type: "User"},
					},
					// Inline object
					{
						Name: "address",
						Type: ir.TypeRef{
							Kind: ir.TypeKindObject,
							Object: &ir.InlineObject{
								Fields: []ir.Field{
									{
										Name: "city",
										Type: ir.TypeRef{Kind: ir.TypeKindPrimitive, Primitive: ir.PrimitiveString},
									},
									{
										Name: "zip",
										Type: ir.TypeRef{Kind: ir.TypeKindPrimitive, Primitive: ir.PrimitiveString},
									},
								},
							},
						},
					},
				},
			},
		},
	}

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

func TestFlattenSchema(t *testing.T) {
	schema := &ir.Schema{
		RPCs: []ir.RPC{
			{
				Name: "Users",
				Procs: []ir.Procedure{
					{Name: "GetUser"},
					{Name: "CreateUser"},
				},
				Streams: []ir.Stream{
					{Name: "UserUpdates"},
				},
			},
			{
				Name: "Products",
				Procs: []ir.Procedure{
					{Name: "ListProducts"},
				},
			},
		},
	}

	flat := flattenSchema(schema)

	// Check procedures
	assert.Len(t, flat.Procedures, 3)
	assert.Equal(t, "Users", flat.Procedures[0].RPCName)
	assert.Equal(t, "GetUser", flat.Procedures[0].Procedure.Name)
	assert.Equal(t, "Users", flat.Procedures[1].RPCName)
	assert.Equal(t, "CreateUser", flat.Procedures[1].Procedure.Name)
	assert.Equal(t, "Products", flat.Procedures[2].RPCName)
	assert.Equal(t, "ListProducts", flat.Procedures[2].Procedure.Name)

	// Check streams
	assert.Len(t, flat.Streams, 1)
	assert.Equal(t, "Users", flat.Streams[0].RPCName)
	assert.Equal(t, "UserUpdates", flat.Streams[0].Stream.Name)
}

func TestFullProcName(t *testing.T) {
	assert.Equal(t, "UsersGetUser", fullProcName("Users", "GetUser"))
	assert.Equal(t, "ProductsList", fullProcName("Products", "List"))
}

func TestFullStreamName(t *testing.T) {
	assert.Equal(t, "ChatMessages", fullStreamName("Chat", "Messages"))
	assert.Equal(t, "UsersUpdates", fullStreamName("Users", "Updates"))
}

func TestRpcProcPath(t *testing.T) {
	assert.Equal(t, "users/getUser", rpcProcPath("Users", "GetUser"))
	assert.Equal(t, "products/list", rpcProcPath("Products", "List"))
}

func TestRpcStreamPath(t *testing.T) {
	assert.Equal(t, "chat/messages", rpcStreamPath("Chat", "Messages"))
	assert.Equal(t, "users/updates", rpcStreamPath("Users", "Updates"))
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

	schema := &ir.Schema{
		RPCs: []ir.RPC{
			{
				Name: "Users",
				Procs: []ir.Procedure{
					{Name: "GetUser"},
				},
			},
		},
	}

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

	schema := &ir.Schema{
		Types: []ir.Type{
			{
				Name:       "LegacyUser",
				Doc:        "Old user type",
				Deprecated: &ir.Deprecation{Message: "Use NewUser instead"},
				Fields: []ir.Field{
					{Name: "id", Type: ir.TypeRef{Kind: ir.TypeKindPrimitive, Primitive: ir.PrimitiveString}},
				},
			},
		},
		Enums: []ir.Enum{
			{
				Name:       "OldStatus",
				Deprecated: &ir.Deprecation{},
				ValueType:  ir.EnumValueTypeString,
				Members:    []ir.EnumMember{{Name: "Active", Value: "active"}},
			},
		},
	}

	files, err := g.Generate(context.Background(), schema)
	require.NoError(t, err)

	content := string(files[0].Content)
	assert.Contains(t, content, "@deprecated Use NewUser instead")
	assert.Contains(t, content, "@deprecated")
}
