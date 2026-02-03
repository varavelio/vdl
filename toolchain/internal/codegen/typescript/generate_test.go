package typescript

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config/configtypes"
	"github.com/varavelio/vdl/toolchain/internal/core/analysis"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
	"github.com/varavelio/vdl/toolchain/internal/core/vfs"
)

func TestGenerator_Name(t *testing.T) {
	g := New(&configtypes.TypeScriptTargetConfig{})
	assert.Equal(t, "typescript", g.Name())
}

func parseAndBuildIR(t *testing.T, content string) *irtypes.IrSchema {
	fs := vfs.New()
	path := "/test.vdl"
	fs.WriteFileCache(path, []byte(content))

	program, diags := analysis.Analyze(fs, path)
	require.Empty(t, diags, "analysis failed")

	return ir.FromProgram(program)
}

// findFile returns the content of a file with the given name from the generated files.
func findFile(files []File, name string) string {
	for _, f := range files {
		if f.RelativePath == name {
			return string(f.Content)
		}
	}
	return ""
}

func TestGenerator_Generate_Empty(t *testing.T) {
	g := New(&configtypes.TypeScriptTargetConfig{
		Output:    "out",
		GenClient: configtypes.Ptr(true),
	})

	schema := parseAndBuildIR(t, "")

	files, err := g.Generate(context.Background(), schema)
	require.NoError(t, err)
	// Expect: core.ts, types.ts, index.ts (no catalog.ts, no client.ts as no RPCs)
	require.Len(t, files, 3)

	coreContent := findFile(files, "core.ts")
	assert.Contains(t, coreContent, "export type Response<T>")

	typesContent := findFile(files, "types.ts")
	assert.NotContains(t, typesContent, "import { Response }")
	assert.NotContains(t, typesContent, "import type { Response }")
}

func TestGenerator_Generate_WithEnums(t *testing.T) {
	g := New(&configtypes.TypeScriptTargetConfig{
		Output: "out",
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
	require.Len(t, files, 3) // core.ts, types.ts, index.ts (no catalog)

	content := findFile(files, "types.ts")

	// String enum
	assert.Contains(t, content, `export type OrderStatus = "pending" | "shipped" | "delivered";`)
	assert.Contains(t, content, `OrderStatusList`)
	assert.Contains(t, content, `function isOrderStatus(value: unknown): value is OrderStatus`)

	// Int enum
	assert.Contains(t, content, `export type Priority = 1 | 2 | 3;`)
	assert.Contains(t, content, `PriorityList`)
}

func TestGenerator_Generate_WithConstants(t *testing.T) {
	g := New(&configtypes.TypeScriptTargetConfig{
		Output: "out",
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
	require.Len(t, files, 4) // core.ts, types.ts, constants.ts, index.ts (no catalog)

	content := findFile(files, "constants.ts")
	assert.Contains(t, content, "export const MAX_PAGE_SIZE: number = 100;")
	assert.Contains(t, content, `export const API_VERSION: string = "1.0.0";`)
	assert.Contains(t, content, "export const DEFAULT_RATE: number = 0.21;")
	assert.Contains(t, content, "export const ENABLED: boolean = true;")

	// Verify constants are exported in index.ts
	indexContent := findFile(files, "index.ts")
	assert.Contains(t, indexContent, `export * from "./constants"`)
}

func TestGenerator_Generate_WithPatterns(t *testing.T) {
	g := New(&configtypes.TypeScriptTargetConfig{
		Output: "out",
	})

	vdl := `
		pattern UserEventSubject = "events.users.{userId}.{eventType}"
		pattern CacheKey = "cache:{key}"
	`
	schema := parseAndBuildIR(t, vdl)

	files, err := g.Generate(context.Background(), schema)
	require.NoError(t, err)
	require.Len(t, files, 4) // core.ts, types.ts, patterns.ts, index.ts (no catalog)

	content := findFile(files, "patterns.ts")
	assert.Contains(t, content, "export function UserEventSubject(userId: string, eventType: string): string")
	assert.Contains(t, content, "return `events.users.${userId}.${eventType}`")
	assert.Contains(t, content, "export function CacheKey(key: string): string")
	assert.Contains(t, content, "return `cache:${key}`")

	// Verify patterns are exported in index.ts
	indexContent := findFile(files, "index.ts")
	assert.Contains(t, indexContent, `export * from "./patterns"`)
}

func TestGenerator_Generate_WithProcedures(t *testing.T) {
	g := New(&configtypes.TypeScriptTargetConfig{
		Output:    "out",
		GenClient: configtypes.Ptr(true),
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
	require.Len(t, files, 5) // coreTypes.ts, types.ts, catalog.ts, client.ts, index.ts

	typesContent := findFile(files, "types.ts")
	catalogContent := findFile(files, "catalog.ts")
	clientContent := findFile(files, "client.ts")

	// Check procedure types (with RPC prefix)
	assert.Contains(t, typesContent, "export type UsersGetUserInput = {")
	assert.Contains(t, typesContent, "export type UsersGetUserOutput = {")
	assert.Contains(t, typesContent, "export type UsersGetUserResponse = Response<UsersGetUserOutput>")

	// Check imports in client.ts
	assert.Contains(t, clientContent, `import type { Response, OperationType, OperationDefinition } from "./core";`)
	assert.Contains(t, clientContent, `import { VdlError, asError, sleep } from "./core";`)

	// Check procedure names list
	assert.Contains(t, catalogContent, `"/Users/GetUser"`)

	// Check client implementation
	assert.Contains(t, clientContent, "class builderUsersGetUser")
	assert.Contains(t, clientContent, "async execute(input: vdlTypes.UsersGetUserInput): Promise<vdlTypes.UsersGetUserOutput>")
}

func TestGenerator_Generate_WithStreams(t *testing.T) {
	g := New(&configtypes.TypeScriptTargetConfig{
		Output:    "out",
		GenClient: configtypes.Ptr(true),
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
	require.Len(t, files, 5) // coreTypes.ts, types.ts, catalog.ts, client.ts, index.ts

	typesContent := findFile(files, "types.ts")
	catalogContent := findFile(files, "catalog.ts")
	clientContent := findFile(files, "client.ts")

	// Check stream types (with RPC prefix)
	assert.Contains(t, typesContent, "export type ChatMessagesInput = {")
	assert.Contains(t, typesContent, "export type ChatMessagesOutput = {")
	assert.Contains(t, typesContent, "export type ChatMessagesResponse = Response<ChatMessagesOutput>")

	// Check stream names list
	assert.Contains(t, catalogContent, `"/Chat/Messages"`)

	// Check client implementation
	assert.Contains(t, clientContent, "class builderChatMessagesStream")
	assert.Contains(t, clientContent, "execute(input: vdlTypes.ChatMessagesInput)")
}

func TestGenerator_Generate_WithComplexTypes(t *testing.T) {
	g := New(&configtypes.TypeScriptTargetConfig{
		Output: "out",
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
	require.Len(t, files, 3) // core.ts, types.ts, index.ts (no catalog)

	content := findFile(files, "types.ts")

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
			want: "number",
		},
		{
			name: "float primitive",
			tr:   irtypes.TypeRef{Kind: irtypes.TypeKindPrimitive, PrimitiveName: irtypes.Ptr(irtypes.PrimitiveTypeFloat)},
			want: "number",
		},
		{
			name: "bool primitive",
			tr:   irtypes.TypeRef{Kind: irtypes.TypeKindPrimitive, PrimitiveName: irtypes.Ptr(irtypes.PrimitiveTypeBool)},
			want: "boolean",
		},
		{
			name: "datetime primitive",
			tr:   irtypes.TypeRef{Kind: irtypes.TypeKindPrimitive, PrimitiveName: irtypes.Ptr(irtypes.PrimitiveTypeDatetime)},
			want: "Date",
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
			want: "string[]",
		},
		{
			name: "2D array of ints",
			tr: irtypes.TypeRef{
				Kind:      irtypes.TypeKindArray,
				ArrayDims: irtypes.Ptr(int64(2)),
				ArrayType: &irtypes.TypeRef{Kind: irtypes.TypeKindPrimitive, PrimitiveName: irtypes.Ptr(irtypes.PrimitiveTypeInt)},
			},
			want: "number[][]",
		},
		{
			name: "map of strings",
			tr: irtypes.TypeRef{
				Kind:    irtypes.TypeKindMap,
				MapType: &irtypes.TypeRef{Kind: irtypes.TypeKindPrimitive, PrimitiveName: irtypes.Ptr(irtypes.PrimitiveTypeString)},
			},
			want: "Record<string, string>",
		},
		{
			name: "map of custom types",
			tr: irtypes.TypeRef{
				Kind:    irtypes.TypeKindMap,
				MapType: &irtypes.TypeRef{Kind: irtypes.TypeKindType, TypeName: irtypes.Ptr("User")},
			},
			want: "Record<string, User>",
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
	g := New(&configtypes.TypeScriptTargetConfig{
		Output:    "out",
		GenClient: configtypes.Ptr(false),
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
	require.Len(t, files, 4) // core.ts, types.ts, catalog.ts, index.ts (no client.ts)

	typesContent := findFile(files, "types.ts")
	clientContent := findFile(files, "client.ts")

	// Procedure types should still be generated
	assert.Contains(t, typesContent, "export type UsersGetUserInput")
	assert.Contains(t, typesContent, "export type UsersGetUserOutput")

	// But client file should NOT exist
	assert.Empty(t, clientContent)
}

func TestGenerator_Generate_WithDeprecation(t *testing.T) {
	g := New(&configtypes.TypeScriptTargetConfig{
		Output: "out",
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

	content := findFile(files, "types.ts")
	assert.Contains(t, content, "@deprecated Use NewUser instead")
	assert.Contains(t, content, "@deprecated")
}

func TestGenerator_Generate_ImportExtension(t *testing.T) {
	tests := []struct {
		name      string
		extension *configtypes.TypescriptImportExtension
		expected  string
	}{
		{
			name:      "none (default)",
			extension: nil,
			expected:  `from "./core";`,
		},
		{
			name:      "explicit none",
			extension: configtypes.Ptr(configtypes.TypescriptImportExtensionNone),
			expected:  `from "./core";`,
		},
		{
			name:      ".js extension",
			extension: configtypes.Ptr(configtypes.TypescriptImportExtensionJs),
			expected:  `from "./core.js";`,
		},
		{
			name:      ".ts extension",
			extension: configtypes.Ptr(configtypes.TypescriptImportExtensionTs),
			expected:  `from "./core.ts";`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := New(&configtypes.TypeScriptTargetConfig{
				Output:          "out",
				ImportExtension: tt.extension,
			})

			schema := parseAndBuildIR(t, `
				type User { id: string }
				rpc S { proc P { input { u: User } output { u: User } } }
			`)

			files, err := g.Generate(context.Background(), schema)
			require.NoError(t, err)

			typesContent := findFile(files, "types.ts")
			assert.Contains(t, typesContent, tt.expected)
			assert.Contains(t, typesContent, "import type { Response }")

			indexContent := findFile(files, "index.ts")
			ext := config.GetImportExtension(tt.extension)
			if ext == "" || ext == configtypes.TypescriptImportExtensionNone {
				assert.Contains(t, indexContent, `export * from "./core";`)
			} else {
				assert.Contains(t, indexContent, `export * from "./core`+string(ext)+`";`)
			}
		})
	}
}
