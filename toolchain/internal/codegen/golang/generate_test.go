package golang

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
)

func TestGenerator_Name(t *testing.T) {
	g := New(&config.GoConfig{})
	assert.Equal(t, "golang", g.Name())
}

func TestGenerator_Generate_Empty(t *testing.T) {
	g := New(&config.GoConfig{
		CommonConfig: config.CommonConfig{
			Output: "api.go",
		},
		Package: "api",
		ServerConfig: config.ServerConfig{
			GenServer: true,
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
	// Expect types.go, optional.go, rpc_server.go, rpc_client.go
	require.Len(t, files, 4)

	fileMap := make(map[string]string)
	for _, f := range files {
		fileMap[f.RelativePath] = string(f.Content)
	}

	require.Contains(t, fileMap, "types.go")
	assert.Contains(t, fileMap["types.go"], "package api")
	require.Contains(t, fileMap, "optional.go")
	require.Contains(t, fileMap, "rpc_server.go")
	require.Contains(t, fileMap, "rpc_client.go")
}

func TestGenerator_Generate_WithTypes(t *testing.T) {

	g := New(&config.GoConfig{
		CommonConfig: config.CommonConfig{
			Output: "api.go",
		},
		Package: "api",
		ServerConfig: config.ServerConfig{
			GenServer: true,
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
	// Expect types.go, optional.go, rpc_server.go, rpc_client.go
	require.Len(t, files, 4)

	fileMap := make(map[string]string)
	for _, f := range files {
		fileMap[f.RelativePath] = string(f.Content)
	}

	require.Contains(t, fileMap, "types.go")
	content := fileMap["types.go"]
	assert.Contains(t, content, "type User struct")

	assert.Contains(t, content, "Id")
	assert.Contains(t, content, "Email")
	assert.Contains(t, content, "Optional[int64]")
}

func TestGenerator_Generate_WithEnums(t *testing.T) {
	g := New(&config.GoConfig{
		CommonConfig: config.CommonConfig{
			Output: "api.go",
		},
		Package: "api",
	})

	schema := &ir.Schema{
		Enums: []ir.Enum{
			{
				Name:      "OrderStatus",
				ValueType: ir.EnumValueTypeString,
				Members: []ir.EnumMember{
					{Name: "Pending", Value: "Pending"},
					{Name: "Shipped", Value: "Shipped"},
					{Name: "Delivered", Value: "Delivered"},
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
	// Expect types.go and optional.go
	require.Len(t, files, 2)

	fileMap := make(map[string]string)
	for _, f := range files {
		fileMap[f.RelativePath] = string(f.Content)
	}

	require.Contains(t, fileMap, "types.go")
	content := fileMap["types.go"]

	// String enum
	assert.Contains(t, content, "type OrderStatus string")
	assert.Contains(t, content, `OrderStatusPending`)
	assert.Contains(t, content, `OrderStatusShipped`)
	assert.Contains(t, content, `OrderStatusDelivered`)

	// Int enum
	assert.Contains(t, content, "type Priority int")
	assert.Contains(t, content, "PriorityLow")
	assert.Contains(t, content, "PriorityMedium")
	assert.Contains(t, content, "PriorityHigh")
}

func TestGenerator_Generate_WithConstants(t *testing.T) {
	g := New(&config.GoConfig{
		CommonConfig: config.CommonConfig{
			Output: "api.go",
		},
		Package: "api",
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
	// Expect types.go, optional.go, and consts.go
	require.Len(t, files, 3)

	// Find consts.go
	var constsFile File
	for _, f := range files {
		if f.RelativePath == "consts.go" {
			constsFile = f
			break
		}
	}
	require.NotEmpty(t, constsFile.RelativePath, "consts.go not found")

	content := string(constsFile.Content)
	assert.Contains(t, content, "const MAX_PAGE_SIZE = 100")
	assert.Contains(t, content, `const API_VERSION = "1.0.0"`)
	assert.Contains(t, content, "const DEFAULT_RATE = 0.21")
	assert.Contains(t, content, "const ENABLED = true")
}

func TestGenerator_Generate_WithPatterns(t *testing.T) {
	g := New(&config.GoConfig{
		CommonConfig: config.CommonConfig{
			Output: "api.go",
		},
		Package: "api",
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
	// Expect types.go, optional.go, and patterns.go
	require.Len(t, files, 3)

	var patternsFile File
	for _, f := range files {
		if f.RelativePath == "patterns.go" {
			patternsFile = f
			break
		}
	}
	require.NotEmpty(t, patternsFile.RelativePath, "patterns.go not found")

	content := string(patternsFile.Content)
	assert.Contains(t, content, "func UserEventSubject(userId string, eventType string) string")
	assert.Contains(t, content, `"events.users." + userId + "." + eventType`)
	assert.Contains(t, content, "func CacheKey(key string) string")
}

func TestGenerator_Generate_WithProcedures(t *testing.T) {
	g := New(&config.GoConfig{
		CommonConfig: config.CommonConfig{
			Output: "api.go",
		},
		Package: "api",
		ServerConfig: config.ServerConfig{
			GenServer: true,
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

	// files map for easy lookup
	fileMap := make(map[string]string)
	for _, f := range files {
		fileMap[f.RelativePath] = string(f.Content)
	}

	// Check types.go
	require.Contains(t, fileMap, "types.go")
	typesContent := fileMap["types.go"]
	assert.Contains(t, typesContent, "type UsersGetUserInput struct")
	assert.Contains(t, typesContent, "type UsersGetUserOutput struct")

	// Check rpc_server.go (core)
	require.Contains(t, fileMap, "rpc_server.go")
	serverCore := fileMap["rpc_server.go"]
	assert.Contains(t, serverCore, "type Server[T any] struct")

	// Check rpc_client.go (core)
	require.Contains(t, fileMap, "rpc_client.go")
	clientCore := fileMap["rpc_client.go"]
	assert.Contains(t, clientCore, "type Client struct")

	// Check rpc_users_server.go
	require.Contains(t, fileMap, "rpc_users_server.go")
	usersServer := fileMap["rpc_users_server.go"]
	assert.Contains(t, usersServer, "type procUsersGetUserEntry[T any] struct")
	assert.Contains(t, usersServer, "func (e procUsersGetUserEntry[T]) Handle")

	// Check rpc_users_client.go
	require.Contains(t, fileMap, "rpc_users_client.go")
	usersClient := fileMap["rpc_users_client.go"]
	assert.Contains(t, usersClient, "type clientBuilderUsersGetUser struct")
	assert.Contains(t, usersClient, "func (b *clientBuilderUsersGetUser) Execute")
}

func TestGenerator_Generate_WithStreams(t *testing.T) {
	g := New(&config.GoConfig{
		CommonConfig: config.CommonConfig{
			Output: "api.go",
		},
		Package: "api",
		ServerConfig: config.ServerConfig{
			GenServer: true,
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

	fileMap := make(map[string]string)
	for _, f := range files {
		fileMap[f.RelativePath] = string(f.Content)
	}

	// Check types.go
	require.Contains(t, fileMap, "types.go")
	typesContent := fileMap["types.go"]
	assert.Contains(t, typesContent, "type ChatMessagesInput struct")
	assert.Contains(t, typesContent, "type ChatMessagesOutput struct")

	// Check rpc_chat_server.go
	require.Contains(t, fileMap, "rpc_chat_server.go")
	chatServer := fileMap["rpc_chat_server.go"]
	assert.Contains(t, chatServer, "type streamChatMessagesEntry[T any] struct")
	assert.Contains(t, chatServer, "func (e streamChatMessagesEntry[T]) Handle")
	assert.Contains(t, chatServer, "func (e streamChatMessagesEntry[T]) UseEmit")

	// Check rpc_chat_client.go
	require.Contains(t, fileMap, "rpc_chat_client.go")
	chatClient := fileMap["rpc_chat_client.go"]
	assert.Contains(t, chatClient, "type clientBuilderChatMessagesStream struct")
}

func TestGenerator_Generate_WithComplexTypes(t *testing.T) {
	g := New(&config.GoConfig{
		CommonConfig: config.CommonConfig{
			Output: "api.go",
		},
		Package: "api",
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
	// Expect types.go and optional.go
	require.Len(t, files, 2)

	fileMap := make(map[string]string)
	for _, f := range files {
		fileMap[f.RelativePath] = string(f.Content)
	}

	require.Contains(t, fileMap, "types.go")
	content := fileMap["types.go"]

	// Arrays
	assert.Contains(t, content, "[]string")

	// Multi-dimensional arrays
	assert.Contains(t, content, "[][]int64")

	// Maps
	assert.Contains(t, content, "map[string]string")

	// Custom type
	assert.Contains(t, content, "Owner")
	assert.Contains(t, content, "User")

	// Inline object - should generate a separate type
	assert.Contains(t, content, "type ProductAddress struct")
	assert.Contains(t, content, "City")
}

func TestTypeRefToGo(t *testing.T) {
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
			want: "int64",
		},
		{
			name: "float primitive",
			tr:   ir.TypeRef{Kind: ir.TypeKindPrimitive, Primitive: ir.PrimitiveFloat},
			want: "float64",
		},
		{
			name: "bool primitive",
			tr:   ir.TypeRef{Kind: ir.TypeKindPrimitive, Primitive: ir.PrimitiveBool},
			want: "bool",
		},
		{
			name: "datetime primitive",
			tr:   ir.TypeRef{Kind: ir.TypeKindPrimitive, Primitive: ir.PrimitiveDatetime},
			want: "time.Time",
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
			want: "[]string",
		},
		{
			name: "2D array of ints",
			tr: ir.TypeRef{
				Kind:            ir.TypeKindArray,
				ArrayDimensions: 2,
				ArrayItem:       &ir.TypeRef{Kind: ir.TypeKindPrimitive, Primitive: ir.PrimitiveInt},
			},
			want: "[][]int64",
		},
		{
			name: "map of strings",
			tr: ir.TypeRef{
				Kind:     ir.TypeKindMap,
				MapValue: &ir.TypeRef{Kind: ir.TypeKindPrimitive, Primitive: ir.PrimitiveString},
			},
			want: "map[string]string",
		},
		{
			name: "map of custom types",
			tr: ir.TypeRef{
				Kind:     ir.TypeKindMap,
				MapValue: &ir.TypeRef{Kind: ir.TypeKindType, Type: "User"},
			},
			want: "map[string]User",
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
			got := typeRefToGo(tt.parent, tt.tr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParsePatternTemplate(t *testing.T) {
	tests := []struct {
		name         string
		template     string
		placeholders []string
		wantParts    []string
	}{
		{
			name:         "simple pattern",
			template:     "events.{userId}",
			placeholders: []string{"userId"},
			wantParts:    []string{`"events."`, "userId"},
		},
		{
			name:         "multiple placeholders",
			template:     "events.users.{userId}.{eventType}",
			placeholders: []string{"userId", "eventType"},
			wantParts:    []string{`"events.users."`, "userId", `"."`, "eventType"},
		},
		{
			name:         "placeholder at start",
			template:     "{prefix}.suffix",
			placeholders: []string{"prefix"},
			wantParts:    []string{"prefix", `".suffix"`},
		},
		{
			name:         "placeholder at end",
			template:     "prefix.{suffix}",
			placeholders: []string{"suffix"},
			wantParts:    []string{`"prefix."`, "suffix"},
		},
		{
			name:         "no placeholders",
			template:     "static.path",
			placeholders: []string{},
			wantParts:    []string{`"static.path"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parsePatternTemplate(tt.template, tt.placeholders)
			assert.Equal(t, tt.wantParts, got)

			// Verify that joined parts would produce valid Go code
			joined := strings.Join(got, " + ")
			t.Logf("Result: return %s", joined)
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
