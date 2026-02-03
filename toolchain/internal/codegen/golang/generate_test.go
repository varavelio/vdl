package golang

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config/configtypes"
	"github.com/varavelio/vdl/toolchain/internal/core/analysis"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
	"github.com/varavelio/vdl/toolchain/internal/core/vfs"
)

func boolPtr(b bool) *bool { return &b }

func TestGenerator_Name(t *testing.T) {
	g := New(&configtypes.GoConfig{})
	assert.Equal(t, "golang", g.Name())
}

func parseAndBuildIR(t *testing.T, content string) *irtypes.IrSchema {
	fs := vfs.New()
	path := "/test.vdl"
	fs.WriteFileCache(path, []byte(content))

	program, diags := analysis.Analyze(fs, path)
	require.Empty(t, diags, "analysis failed")

	return ir.FromProgram(program)
}

func TestGenerator_Generate_Empty(t *testing.T) {
	g := New(&configtypes.GoConfig{
		Output:    "api.go",
		Package:   "api",
		GenServer: boolPtr(true),
		GenClient: boolPtr(true),
	})

	schema := parseAndBuildIR(t, "")

	files, err := g.Generate(context.Background(), schema)
	require.NoError(t, err)
	// Expect core.go, pointers.go, server.go, client.go
	// types.go is NOT generated because there are no types/procs/streams.
	require.Len(t, files, 4)

	fileMap := make(map[string]string)
	for _, f := range files {
		fileMap[f.RelativePath] = string(f.Content)
	}

	require.Contains(t, fileMap, "core.go")
	assert.Contains(t, fileMap["core.go"], "package api")
	require.Contains(t, fileMap, "pointers.go")
	require.Contains(t, fileMap, "server.go")
	require.Contains(t, fileMap, "client.go")
	require.NotContains(t, fileMap, "types.go")
}

func TestGenerator_Generate_WithTypes(t *testing.T) {

	g := New(&configtypes.GoConfig{
		Output:    "api.go",
		Package:   "api",
		GenServer: boolPtr(true),
		GenClient: boolPtr(true),
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
	// Expect core.go, types.go, optional.go, server.go, client.go
	require.Len(t, files, 5)

	fileMap := make(map[string]string)
	for _, f := range files {
		fileMap[f.RelativePath] = string(f.Content)
	}

	require.Contains(t, fileMap, "core.go")
	require.Contains(t, fileMap, "types.go")
	content := fileMap["types.go"]
	assert.Contains(t, content, "type User struct")

	assert.Contains(t, content, "Id")
	assert.Contains(t, content, "Email")
	// Optional fields now use pointers instead of Optional[T]
	assert.Contains(t, content, "*int64")
	// Verify getters are generated
	assert.Contains(t, content, "func (x *User) GetAge() int64")
	assert.Contains(t, content, "func (x *User) GetAgeOr(defaultValue int64) int64")
}

func TestGenerator_Generate_WithEnums(t *testing.T) {
	g := New(&configtypes.GoConfig{
		Output:  "api.go",
		Package: "api",
	})

	vdl := `
		enum OrderStatus {
			Pending = "Pending"
			Shipped = "Shipped"
			Delivered = "Delivered"
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
	// Expect core_types.go, types.go and pointers.go
	require.Len(t, files, 3)

	fileMap := make(map[string]string)
	for _, f := range files {
		fileMap[f.RelativePath] = string(f.Content)
	}

	require.Contains(t, fileMap, "core.go")
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
	g := New(&configtypes.GoConfig{
		Output:  "api.go",
		Package: "api",
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
	// Expect core.go, pointers.go, constants.go
	// types.go is NOT generated because there are no types/procs/streams.
	require.Len(t, files, 3)

	// Find constants.go
	var constsFile File
	for _, f := range files {
		if f.RelativePath == "constants.go" {
			constsFile = f
			break
		}
	}
	require.NotEmpty(t, constsFile.RelativePath, "constants.go not found")

	content := string(constsFile.Content)
	assert.Contains(t, content, "const MAX_PAGE_SIZE = 100")
	assert.Contains(t, content, `const API_VERSION = "1.0.0"`)
	assert.Contains(t, content, "const DEFAULT_RATE = 0.21")
	assert.Contains(t, content, "const ENABLED = true")
}

func TestGenerator_Generate_WithPatterns(t *testing.T) {
	g := New(&configtypes.GoConfig{
		Output:  "api.go",
		Package: "api",
	})

	vdl := `
		pattern UserEventSubject = "events.users.{userId}.{eventType}"
		pattern CacheKey = "cache:{key}"
	`
	schema := parseAndBuildIR(t, vdl)

	files, err := g.Generate(context.Background(), schema)
	require.NoError(t, err)
	// Expect core.go, pointers.go, patterns.go
	// types.go is NOT generated because there are no types/procs/streams.
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
	g := New(&configtypes.GoConfig{
		Output:    "api.go",
		Package:   "api",
		GenServer: boolPtr(true),
		GenClient: boolPtr(true),
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

	// files map for easy lookup
	fileMap := make(map[string]string)
	for _, f := range files {
		fileMap[f.RelativePath] = string(f.Content)
	}

	// Check core.go
	require.Contains(t, fileMap, "core.go")

	// Check types.go
	require.Contains(t, fileMap, "types.go")
	typesContent := fileMap["types.go"]
	assert.Contains(t, typesContent, "type UsersGetUserInput struct")
	assert.Contains(t, typesContent, "type UsersGetUserOutput struct")

	// Check catalog.go
	require.Contains(t, fileMap, "catalog.go")
	catalogContent := fileMap["catalog.go"]
	assert.Contains(t, catalogContent, "VDLProcedures")
	assert.Contains(t, catalogContent, "RPCName: \"Users\"")

	// Check server.go (core + RPCs)
	require.Contains(t, fileMap, "server.go")
	serverCore := fileMap["server.go"]
	assert.Contains(t, serverCore, "type Server[T any] struct")
	// RPC specific code should be here too
	assert.Contains(t, serverCore, "type procUsersGetUserEntry[T any] struct")
	assert.Contains(t, serverCore, "func (e procUsersGetUserEntry[T]) Handle")

	// Check client.go (core + RPCs)
	require.Contains(t, fileMap, "client.go")
	clientCore := fileMap["client.go"]
	assert.Contains(t, clientCore, "type Client struct")
	// RPC specific code should be here too
	assert.Contains(t, clientCore, "type clientBuilderUsersGetUser struct")
	assert.Contains(t, clientCore, "func (b *clientBuilderUsersGetUser) Execute")

	// Files rpc_users_server.go and rpc_users_client.go are no longer generated
	require.NotContains(t, fileMap, "rpc_users_server.go")
	require.NotContains(t, fileMap, "rpc_users_client.go")
}

func TestGenerator_Generate_WithStreams(t *testing.T) {
	g := New(&configtypes.GoConfig{
		Output:    "api.go",
		Package:   "api",
		GenServer: boolPtr(true),
		GenClient: boolPtr(true),
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

	fileMap := make(map[string]string)
	for _, f := range files {
		fileMap[f.RelativePath] = string(f.Content)
	}

	// Check core.go
	require.Contains(t, fileMap, "core.go")

	// Check types.go
	require.Contains(t, fileMap, "types.go")
	typesContent := fileMap["types.go"]
	assert.Contains(t, typesContent, "type ChatMessagesInput struct")
	assert.Contains(t, typesContent, "type ChatMessagesOutput struct")

	// Check catalog.go
	require.Contains(t, fileMap, "catalog.go")
	catalogContent := fileMap["catalog.go"]
	assert.Contains(t, catalogContent, "VDLStreams")
	assert.Contains(t, catalogContent, "RPCName: \"Chat\"")

	// Check server.go (core + streams)
	require.Contains(t, fileMap, "server.go")
	serverCore := fileMap["server.go"]
	assert.Contains(t, serverCore, "type streamChatMessagesEntry[T any] struct")
	assert.Contains(t, serverCore, "func (e streamChatMessagesEntry[T]) Handle")
	assert.Contains(t, serverCore, "func (e streamChatMessagesEntry[T]) UseEmit")

	// Check client.go (core + streams)
	require.Contains(t, fileMap, "client.go")
	clientCore := fileMap["client.go"]
	assert.Contains(t, clientCore, "type clientBuilderChatMessagesStream struct")

	// Files rpc_chat_server.go and rpc_chat_client.go are no longer generated
	require.NotContains(t, fileMap, "rpc_chat_server.go")
	require.NotContains(t, fileMap, "rpc_chat_client.go")
}

func TestGenerator_Generate_WithComplexTypes(t *testing.T) {
	g := New(&configtypes.GoConfig{
		Output:  "api.go",
		Package: "api",
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
	// Expect core_types.go, types.go and pointers.go
	require.Len(t, files, 3)

	fileMap := make(map[string]string)
	for _, f := range files {
		fileMap[f.RelativePath] = string(f.Content)
	}

	require.Contains(t, fileMap, "core.go")
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
		tr     irtypes.TypeRef
		parent string
		want   string
	}{
		{
			name: "string primitive",
			tr:   irtypes.TypeRef{Kind: irtypes.TypeKindPrimitive, PrimitiveName: irtypes.Ptr(irtypes.PrimitiveTypeString)},
			want: "string",
		},
		{
			name: "int primitive",
			tr:   irtypes.TypeRef{Kind: irtypes.TypeKindPrimitive, PrimitiveName: irtypes.Ptr(irtypes.PrimitiveTypeInt)},
			want: "int64",
		},
		{
			name: "float primitive",
			tr:   irtypes.TypeRef{Kind: irtypes.TypeKindPrimitive, PrimitiveName: irtypes.Ptr(irtypes.PrimitiveTypeFloat)},
			want: "float64",
		},
		{
			name: "bool primitive",
			tr:   irtypes.TypeRef{Kind: irtypes.TypeKindPrimitive, PrimitiveName: irtypes.Ptr(irtypes.PrimitiveTypeBool)},
			want: "bool",
		},
		{
			name: "datetime primitive",
			tr:   irtypes.TypeRef{Kind: irtypes.TypeKindPrimitive, PrimitiveName: irtypes.Ptr(irtypes.PrimitiveTypeDatetime)},
			want: "time.Time",
		},
		{
			name: "custom type",
			tr:   irtypes.TypeRef{Kind: irtypes.TypeKindType, TypeName: irtypes.Ptr("User")},
			want: "User",
		},
		{
			name: "enum",
			tr:   irtypes.TypeRef{Kind: irtypes.TypeKindEnum, EnumName: irtypes.Ptr("OrderStatus")},
			want: "OrderStatus",
		},
		{
			name: "1D array of strings",
			tr: irtypes.TypeRef{
				Kind:      irtypes.TypeKindArray,
				ArrayDims: irtypes.Ptr(int64(1)),
				ArrayType: &irtypes.TypeRef{Kind: irtypes.TypeKindPrimitive, PrimitiveName: irtypes.Ptr(irtypes.PrimitiveTypeString)},
			},
			want: "[]string",
		},
		{
			name: "2D array of ints",
			tr: irtypes.TypeRef{
				Kind:      irtypes.TypeKindArray,
				ArrayDims: irtypes.Ptr(int64(2)),
				ArrayType: &irtypes.TypeRef{Kind: irtypes.TypeKindPrimitive, PrimitiveName: irtypes.Ptr(irtypes.PrimitiveTypeInt)},
			},
			want: "[][]int64",
		},
		{
			name: "map of strings",
			tr: irtypes.TypeRef{
				Kind:    irtypes.TypeKindMap,
				MapType: &irtypes.TypeRef{Kind: irtypes.TypeKindPrimitive, PrimitiveName: irtypes.Ptr(irtypes.PrimitiveTypeString)},
			},
			want: "map[string]string",
		},
		{
			name: "map of custom types",
			tr: irtypes.TypeRef{
				Kind:    irtypes.TypeKindMap,
				MapType: &irtypes.TypeRef{Kind: irtypes.TypeKindType, TypeName: irtypes.Ptr("User")},
			},
			want: "map[string]User",
		},
		{
			name:   "inline object",
			tr:     irtypes.TypeRef{Kind: irtypes.TypeKindObject},
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
