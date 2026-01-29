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

// findFileContent finds a file by name and returns its content.
func findFileContent(files []File, name string) string {
	for _, f := range files {
		if f.RelativePath == name {
			return string(f.Content)
		}
	}
	return ""
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
	require.Len(t, files, 2) // core_types.dart and index.dart

	// Check core_types.dart exists
	coreContent := findFileContent(files, "core_types.dart")
	require.NotEmpty(t, coreContent)
	assert.Contains(t, coreContent, "class Response<T>")
	assert.Contains(t, coreContent, "class VdlError")

	// Check index.dart exists
	indexContent := findFileContent(files, "index.dart")
	require.NotEmpty(t, indexContent)
	assert.Contains(t, indexContent, "export 'core_types.dart';")
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

	typesContent := findFileContent(files, "types.dart")
	require.NotEmpty(t, typesContent, "types.dart should be generated")

	// Basic class structure
	assert.Contains(t, typesContent, "class User {")
	assert.Contains(t, typesContent, "final String id;")
	assert.Contains(t, typesContent, "final String email;")
	assert.Contains(t, typesContent, "final int? age;")
	assert.Contains(t, typesContent, "factory User.fromJson")
	assert.Contains(t, typesContent, "Map<String, dynamic> toJson()")

	// copyWith method
	assert.Contains(t, typesContent, "User copyWith({")

	// == operator and hashCode
	assert.Contains(t, typesContent, "bool operator ==(Object other)")
	assert.Contains(t, typesContent, "int get hashCode")

	// toString
	assert.Contains(t, typesContent, "String toString()")

	// Check index exports types.dart
	indexContent := findFileContent(files, "index.dart")
	assert.Contains(t, indexContent, "export 'types.dart';")
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

	// Enums are now in types.dart
	typesContent := findFileContent(files, "types.dart")
	require.NotEmpty(t, typesContent, "types.dart should be generated")

	// String enum
	assert.Contains(t, typesContent, "enum OrderStatus {")
	assert.Contains(t, typesContent, "Pending('pending')")
	assert.Contains(t, typesContent, "Shipped('shipped')")
	assert.Contains(t, typesContent, "final String value;")
	assert.Contains(t, typesContent, "static OrderStatus? fromValue(String value)")

	// Int enum
	assert.Contains(t, typesContent, "enum Priority {")
	assert.Contains(t, typesContent, "Low(1)")
	assert.Contains(t, typesContent, "Medium(2)")
	assert.Contains(t, typesContent, "final int value;")

	// Check index exports types.dart (not enums.dart)
	indexContent := findFileContent(files, "index.dart")
	assert.Contains(t, indexContent, "export 'types.dart';")
	assert.NotContains(t, indexContent, "export 'enums.dart';")
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

	constantsContent := findFileContent(files, "constants.dart")
	require.NotEmpty(t, constantsContent, "constants.dart should be generated")

	assert.Contains(t, constantsContent, "const int MAX_PAGE_SIZE = 100;")
	assert.Contains(t, constantsContent, "const String API_VERSION = '1.0.0';")
	assert.Contains(t, constantsContent, "const double DEFAULT_RATE = 0.21;")
	assert.Contains(t, constantsContent, "const bool ENABLED = true;")

	// Check index exports constants.dart
	indexContent := findFileContent(files, "index.dart")
	assert.Contains(t, indexContent, "export 'constants.dart';")
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

	patternsContent := findFileContent(files, "patterns.dart")
	require.NotEmpty(t, patternsContent, "patterns.dart should be generated")

	assert.Contains(t, patternsContent, "String UserEventSubject(String userId, String eventType)")
	assert.Contains(t, patternsContent, "return 'events.users.$userId.$eventType';")
	assert.Contains(t, patternsContent, "String CacheKey(String key)")
	assert.Contains(t, patternsContent, "return 'cache:$key';")

	// Check index exports patterns.dart
	indexContent := findFileContent(files, "index.dart")
	assert.Contains(t, indexContent, "export 'patterns.dart';")
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

	// Procedures are now in types.dart
	typesContent := findFileContent(files, "types.dart")
	require.NotEmpty(t, typesContent, "types.dart should be generated")

	// Check procedure types (with RPC prefix)
	assert.Contains(t, typesContent, "class UsersGetUserInput {")
	assert.Contains(t, typesContent, "class UsersGetUserOutput {")
	assert.Contains(t, typesContent, "typedef UsersGetUserResponse = Response<UsersGetUserOutput>;")

	// Check procedure path in metadata
	assert.Contains(t, typesContent, "'Users/GetUser'")

	// Check client implementation is NOT present
	assert.NotContains(t, typesContent, "class _BuilderUsersGetUser")
	assert.NotContains(t, typesContent, "Future<UsersGetUserOutput> execute(UsersGetUserInput input)")

	// Check index exports types.dart (not procedures.dart)
	indexContent := findFileContent(files, "index.dart")
	assert.Contains(t, indexContent, "export 'types.dart';")
	assert.NotContains(t, indexContent, "export 'procedures.dart';")
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

	// Streams are now in types.dart
	typesContent := findFileContent(files, "types.dart")
	require.NotEmpty(t, typesContent, "types.dart should be generated for streams")

	// Check stream types (with RPC prefix)
	assert.Contains(t, typesContent, "class ChatMessagesInput {")
	assert.Contains(t, typesContent, "class ChatMessagesOutput {")
	assert.Contains(t, typesContent, "typedef ChatMessagesResponse = Response<ChatMessagesOutput>;")

	// Check stream path in metadata
	assert.Contains(t, typesContent, "'Chat/Messages'")

	// Check client implementation is NOT present
	assert.NotContains(t, typesContent, "class _BuilderChatMessagesStream")
	assert.NotContains(t, typesContent, "_StreamHandle<ChatMessagesOutput> execute(ChatMessagesInput input)")
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

	typesContent := findFileContent(files, "types.dart")
	require.NotEmpty(t, typesContent, "types.dart should be generated")

	// Arrays
	assert.Contains(t, typesContent, "List<String> tags")

	// Multi-dimensional arrays
	assert.Contains(t, typesContent, "List<List<int>> matrix")

	// Maps
	assert.Contains(t, typesContent, "Map<String, String> metadata")

	// Custom type
	assert.Contains(t, typesContent, "User owner")

	// Inline object - should generate a separate class
	assert.Contains(t, typesContent, "class ProductAddress {")
	assert.Contains(t, typesContent, "ProductAddress address")
}

func TestGenerator_Generate_WithRPCCatalog(t *testing.T) {
	g := New(&config.DartConfig{
		CommonConfig: config.CommonConfig{
			Output: "output",
		},
	})

	vdl := `
		rpc Users {
			proc GetUser {
				input { userId: string }
				output { id: string name: string }
			}
			stream UserUpdates {
				input { userId: string }
				output { status: string }
			}
		}
		rpc Products {
			proc ListProducts {
				input { page: int }
				output { items: string[] }
			}
		}
	`
	schema := parseAndBuildIR(t, vdl)

	files, err := g.Generate(context.Background(), schema)
	require.NoError(t, err)

	catalogContent := findFileContent(files, "rpc_catalog.dart")
	require.NotEmpty(t, catalogContent, "rpc_catalog.dart should be generated")

	// Check OperationType enum
	assert.Contains(t, catalogContent, "enum OperationType {")
	assert.Contains(t, catalogContent, "proc,")
	assert.Contains(t, catalogContent, "stream;")

	// Check OperationDefinition class
	assert.Contains(t, catalogContent, "class OperationDefinition {")

	// Check VDLProcedures list
	assert.Contains(t, catalogContent, "const List<OperationDefinition> vdlProcedures = [")
	assert.Contains(t, catalogContent, "rpcName: 'Users', name: 'GetUser', type: OperationType.proc")
	assert.Contains(t, catalogContent, "rpcName: 'Products', name: 'ListProducts', type: OperationType.proc")

	// Check VDLStreams list
	assert.Contains(t, catalogContent, "const List<OperationDefinition> vdlStreams = [")
	assert.Contains(t, catalogContent, "rpcName: 'Users', name: 'UserUpdates', type: OperationType.stream")

	// Check VDLPaths
	assert.Contains(t, catalogContent, "abstract class VDLPaths {")
	assert.Contains(t, catalogContent, "static const users = _UsersPaths._();")
	assert.Contains(t, catalogContent, "class _UsersPaths {")
	assert.Contains(t, catalogContent, "String get getUser => '/Users/GetUser';")
	assert.Contains(t, catalogContent, "String get userUpdates => '/Users/UserUpdates';")

	// Check index exports rpc_catalog.dart
	indexContent := findFileContent(files, "index.dart")
	assert.Contains(t, indexContent, "export 'rpc_catalog.dart';")
}

func TestGenerator_Generate_WithEnumInType(t *testing.T) {
	g := New(&config.DartConfig{
		CommonConfig: config.CommonConfig{
			Output: "output",
		},
	})

	vdl := `
		enum Status {
			Active = "active"
			Inactive = "inactive"
		}

		type User {
			name: string
			status: Status
		}
	`
	schema := parseAndBuildIR(t, vdl)

	files, err := g.Generate(context.Background(), schema)
	require.NoError(t, err)

	typesContent := findFileContent(files, "types.dart")
	require.NotEmpty(t, typesContent, "types.dart should be generated")

	// Check that enum JSON conversion is used in fromJson
	assert.Contains(t, typesContent, "StatusJson.fromJson(json['status']")

	// Check that enum toJson is used
	assert.Contains(t, typesContent, "status.toJson()")

	// Enums are now in the same file, so no import needed
	assert.NotContains(t, typesContent, "import 'enums.dart';")
}

func TestGenerator_Generate_MultipleFiles(t *testing.T) {
	g := New(&config.DartConfig{
		CommonConfig: config.CommonConfig{
			Output: "output",
		},
	})

	vdl := `
		enum Status { Active = "active" }
		const VERSION = "1.0"
		pattern Key = "key:{id}"
		type User { id: string status: Status }
		rpc Users {
			proc Get { input { id: string } output { user: User } }
		}
	`
	schema := parseAndBuildIR(t, vdl)

	files, err := g.Generate(context.Background(), schema)
	require.NoError(t, err)

	// Should generate 6 files:
	// core_types.dart, constants.dart, patterns.dart, types.dart, rpc_catalog.dart, index.dart
	// (enums and procedures are now in types.dart)
	require.Len(t, files, 6)

	// Verify each file exists
	assert.NotEmpty(t, findFileContent(files, "core_types.dart"))
	assert.NotEmpty(t, findFileContent(files, "constants.dart"))
	assert.NotEmpty(t, findFileContent(files, "patterns.dart"))
	assert.NotEmpty(t, findFileContent(files, "types.dart"))
	assert.NotEmpty(t, findFileContent(files, "rpc_catalog.dart"))
	assert.NotEmpty(t, findFileContent(files, "index.dart"))

	// Verify enums.dart and procedures.dart do NOT exist
	assert.Empty(t, findFileContent(files, "enums.dart"))
	assert.Empty(t, findFileContent(files, "procedures.dart"))

	// Verify index exports all files (except enums.dart and procedures.dart)
	indexContent := findFileContent(files, "index.dart")
	assert.Contains(t, indexContent, "export 'core_types.dart';")
	assert.Contains(t, indexContent, "export 'constants.dart';")
	assert.Contains(t, indexContent, "export 'patterns.dart';")
	assert.Contains(t, indexContent, "export 'types.dart';")
	assert.Contains(t, indexContent, "export 'rpc_catalog.dart';")
	assert.NotContains(t, indexContent, "export 'enums.dart';")
	assert.NotContains(t, indexContent, "export 'procedures.dart';")
}

func TestGenerator_Generate_FileHeader(t *testing.T) {
	g := New(&config.DartConfig{
		CommonConfig: config.CommonConfig{
			Output: "output",
		},
	})

	schema := parseAndBuildIR(t, "type User { id: string }")

	files, err := g.Generate(context.Background(), schema)
	require.NoError(t, err)

	// Check all files have the correct header
	for _, f := range files {
		content := string(f.Content)
		assert.Contains(t, content, "// Code generated by VDL v", "file %s should have version header", f.RelativePath)
		assert.Contains(t, content, "DO NOT EDIT", "file %s should have DO NOT EDIT warning", f.RelativePath)
		assert.Contains(t, content, "https://vdl.varavel.com", "file %s should have VDL URL", f.RelativePath)
		// Should NOT contain license info
		assert.NotContains(t, content, "MIT License", "file %s should not have license", f.RelativePath)
		assert.NotContains(t, content, "COPYRIGHT", "file %s should not have copyright", f.RelativePath)
	}
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
