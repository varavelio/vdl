package dart

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
)

func TestGenerator_Name(t *testing.T) {
	g := New(&config.DartConfig{})
	assert.Equal(t, "dart", g.Name())
}

func TestGenerator_Generate_Empty(t *testing.T) {
	g := New(&config.DartConfig{
		CommonConfig: config.CommonConfig{
			Output: "output",
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
	require.Len(t, files, 1) // client.dart

	// Find client.dart
	var clientContent string
	for _, f := range files {
		if f.RelativePath == "client.dart" {
			clientContent = string(f.Content)
			break
		}
	}
	require.NotEmpty(t, clientContent)
}

func TestGenerator_Generate_WithTypes(t *testing.T) {
	g := New(&config.DartConfig{
		CommonConfig: config.CommonConfig{
			Output: "output",
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

	var clientContent string
	for _, f := range files {
		if f.RelativePath == "client.dart" {
			clientContent = string(f.Content)
			break
		}
	}

	assert.Contains(t, clientContent, "class User {")
	assert.Contains(t, clientContent, "final String id;")
	assert.Contains(t, clientContent, "final String email;")
	assert.Contains(t, clientContent, "final int? age;")
	assert.Contains(t, clientContent, "factory User.fromJson")
	assert.Contains(t, clientContent, "Map<String, dynamic> toJson()")
}

func TestGenerator_Generate_WithEnums(t *testing.T) {
	g := New(&config.DartConfig{
		CommonConfig: config.CommonConfig{
			Output: "output",
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

	var clientContent string
	for _, f := range files {
		if f.RelativePath == "client.dart" {
			clientContent = string(f.Content)
			break
		}
	}

	// String enum
	assert.Contains(t, clientContent, "enum OrderStatus {")
	assert.Contains(t, clientContent, "Pending('pending')")
	assert.Contains(t, clientContent, "Shipped('shipped')")
	assert.Contains(t, clientContent, "final String value;")
	assert.Contains(t, clientContent, "static OrderStatus? fromValue(String value)")

	// Int enum
	assert.Contains(t, clientContent, "enum Priority {")
	assert.Contains(t, clientContent, "Low(1)")
	assert.Contains(t, clientContent, "Medium(2)")
	assert.Contains(t, clientContent, "final int value;")
}

func TestGenerator_Generate_WithConstants(t *testing.T) {
	g := New(&config.DartConfig{
		CommonConfig: config.CommonConfig{
			Output: "output",
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

	var clientContent string
	for _, f := range files {
		if f.RelativePath == "client.dart" {
			clientContent = string(f.Content)
			break
		}
	}

	assert.Contains(t, clientContent, "const int MAX_PAGE_SIZE = 100;")
	assert.Contains(t, clientContent, "const String API_VERSION = '1.0.0';")
	assert.Contains(t, clientContent, "const double DEFAULT_RATE = 0.21;")
	assert.Contains(t, clientContent, "const bool ENABLED = true;")
}

func TestGenerator_Generate_WithPatterns(t *testing.T) {
	g := New(&config.DartConfig{
		CommonConfig: config.CommonConfig{
			Output: "output",
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

	var clientContent string
	for _, f := range files {
		if f.RelativePath == "client.dart" {
			clientContent = string(f.Content)
			break
		}
	}

	assert.Contains(t, clientContent, "String UserEventSubject(String userId, String eventType)")
	assert.Contains(t, clientContent, "return 'events.users.$userId.$eventType';")
	assert.Contains(t, clientContent, "String CacheKey(String key)")
	assert.Contains(t, clientContent, "return 'cache:$key';")
}

func TestGenerator_Generate_WithProcedures(t *testing.T) {
	g := New(&config.DartConfig{
		CommonConfig: config.CommonConfig{
			Output: "output",
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

	var clientContent string
	for _, f := range files {
		if f.RelativePath == "client.dart" {
			clientContent = string(f.Content)
			break
		}
	}

	// Check procedure types (with RPC prefix)
	assert.Contains(t, clientContent, "class UsersGetUserInput {")
	assert.Contains(t, clientContent, "class UsersGetUserOutput {")
	assert.Contains(t, clientContent, "typedef UsersGetUserResponse = Response<UsersGetUserOutput>;")

	// Check procedure path in metadata
	assert.Contains(t, clientContent, "'users/getUser'")

	// Check client implementation is NOT present
	assert.NotContains(t, clientContent, "class _BuilderUsersGetUser")
	assert.NotContains(t, clientContent, "Future<UsersGetUserOutput> execute(UsersGetUserInput input)")
}

func TestGenerator_Generate_WithStreams(t *testing.T) {
	g := New(&config.DartConfig{
		CommonConfig: config.CommonConfig{
			Output: "output",
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

	var clientContent string
	for _, f := range files {
		if f.RelativePath == "client.dart" {
			clientContent = string(f.Content)
			break
		}
	}

	// Check stream types (with RPC prefix)
	assert.Contains(t, clientContent, "class ChatMessagesInput {")
	assert.Contains(t, clientContent, "class ChatMessagesOutput {")
	assert.Contains(t, clientContent, "typedef ChatMessagesResponse = Response<ChatMessagesOutput>;")

	// Check stream path in metadata
	assert.Contains(t, clientContent, "'chat/messages'")

	// Check client implementation is NOT present
	assert.NotContains(t, clientContent, "class _BuilderChatMessagesStream")
	assert.NotContains(t, clientContent, "_StreamHandle<ChatMessagesOutput> execute(ChatMessagesInput input)")
}

func TestGenerator_Generate_WithComplexTypes(t *testing.T) {
	g := New(&config.DartConfig{
		CommonConfig: config.CommonConfig{
			Output: "output",
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

	var clientContent string
	for _, f := range files {
		if f.RelativePath == "client.dart" {
			clientContent = string(f.Content)
			break
		}
	}

	// Arrays
	assert.Contains(t, clientContent, "List<String> tags")

	// Multi-dimensional arrays
	assert.Contains(t, clientContent, "List<List<int>> matrix")

	// Maps
	assert.Contains(t, clientContent, "Map<String, String> metadata")

	// Custom type
	assert.Contains(t, clientContent, "User owner")

	// Inline object - should generate a separate class
	assert.Contains(t, clientContent, "class ProductAddress {")
	assert.Contains(t, clientContent, "ProductAddress address")
}

func TestTypeRefToDart(t *testing.T) {
	tests := []struct {
		name   string
		tr     ir.TypeRef
		parent string
		want   string
	}{
		{
			name: "string primitive",
			tr:   ir.TypeRef{Kind: ir.TypeKindPrimitive, Primitive: ir.PrimitiveString},
			want: "String",
		},
		{
			name: "int primitive",
			tr:   ir.TypeRef{Kind: ir.TypeKindPrimitive, Primitive: ir.PrimitiveInt},
			want: "int",
		},
		{
			name: "float primitive",
			tr:   ir.TypeRef{Kind: ir.TypeKindPrimitive, Primitive: ir.PrimitiveFloat},
			want: "double",
		},
		{
			name: "bool primitive",
			tr:   ir.TypeRef{Kind: ir.TypeKindPrimitive, Primitive: ir.PrimitiveBool},
			want: "bool",
		},
		{
			name: "datetime primitive",
			tr:   ir.TypeRef{Kind: ir.TypeKindPrimitive, Primitive: ir.PrimitiveDatetime},
			want: "DateTime",
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
			want: "List<String>",
		},
		{
			name: "2D array of ints",
			tr: ir.TypeRef{
				Kind:            ir.TypeKindArray,
				ArrayDimensions: 2,
				ArrayItem:       &ir.TypeRef{Kind: ir.TypeKindPrimitive, Primitive: ir.PrimitiveInt},
			},
			want: "List<List<int>>",
		},
		{
			name: "map of strings",
			tr: ir.TypeRef{
				Kind:     ir.TypeKindMap,
				MapValue: &ir.TypeRef{Kind: ir.TypeKindPrimitive, Primitive: ir.PrimitiveString},
			},
			want: "Map<String, String>",
		},
		{
			name: "map of custom types",
			tr: ir.TypeRef{
				Kind:     ir.TypeKindMap,
				MapValue: &ir.TypeRef{Kind: ir.TypeKindType, Type: "User"},
			},
			want: "Map<String, User>",
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
			got := typeRefToDart(tt.parent, tt.tr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConvertPatternToDartInterpolation(t *testing.T) {
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
			want:         "'events.$userId'",
		},
		{
			name:         "multiple placeholders",
			template:     "events.users.{userId}.{eventType}",
			placeholders: []string{"userId", "eventType"},
			want:         "'events.users.$userId.$eventType'",
		},
		{
			name:         "placeholder at start",
			template:     "{prefix}.suffix",
			placeholders: []string{"prefix"},
			want:         "'$prefix.suffix'",
		},
		{
			name:         "placeholder at end",
			template:     "prefix.{suffix}",
			placeholders: []string{"suffix"},
			want:         "'prefix.$suffix'",
		},
		{
			name:         "no placeholders",
			template:     "static.path",
			placeholders: []string{},
			want:         "'static.path'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertPatternToDartInterpolation(tt.template, tt.placeholders)
			assert.Equal(t, tt.want, got)
		})
	}
}
