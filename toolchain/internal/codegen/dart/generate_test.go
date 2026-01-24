package dart

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
	g := New(&config.DartConfig{})
	assert.Equal(t, "dart", g.Name())
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
	g := New(&config.DartConfig{
		CommonConfig: config.CommonConfig{
			Output: "output",
		},
	})

	schema := parseAndBuildIR(t, "")

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

	vdl := `
		const MAX_PAGE_SIZE = 100
		const API_VERSION = "1.0.0"
		const DEFAULT_RATE = 0.21
		const ENABLED = true
	`
	schema := parseAndBuildIR(t, vdl)

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

	vdl := `
		pattern UserEventSubject = "events.users.{userId}.{eventType}"
		pattern CacheKey = "cache:{key}"
	`
	schema := parseAndBuildIR(t, vdl)

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
	assert.Contains(t, clientContent, "'Users/GetUser'")

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
	assert.Contains(t, clientContent, "'Chat/Messages'")

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
	// ... (rest of the tests using TypeRef directly can remain as they test low-level logic)
	// But since this is a unit test for a private function (if it were private),
	// or specific type conversion logic, we might want to keep it.
	// Since typeRefToDart is not exported, it's testing internal logic.
	// We can keep it as is, or remove it if we rely on full generation tests.
	// For now I'll keep it but fix imports if needed.
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
