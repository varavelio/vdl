package transform

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/varavelio/vdl/toolchain/internal/core/ast"
	"github.com/varavelio/vdl/toolchain/internal/core/parser"
	"github.com/varavelio/vdl/toolchain/internal/formatter"
)

func TestExpandTypes(t *testing.T) {
	require := require.New(t)

	input := `
		""" Type1 Docstring """
		type Type1 {
			field1: string
			""" Field2 Docstring """
			field2: int
			field3: {
				subfield1: datetime
				""" Subfield2 Docstring """
				subfield2: bool[]
			}[]
		}

		""" Type2 Docstring """
		type Type2 {
			field1: {
				""" Subfield1 Docstring """
				subfield1: Type1[]

				subfield2: Type1[]
			}[]
		}

		rpc TestService {
			""" Proc1 Docstring """
			proc Proc1 {
				input {
					field1: {
						""" Subfield1 Docstring """
						subfield1: Type2[]
					}[]
				}

				output {
					field1: {
						subfield1: Type2[]
					}[]
				}
			}

			""" Stream1 Docstring """
			stream Stream1 {
				input {
					field1: {
						""" Subfield1 Docstring """
						subfield1: Type2[]
					}[]
				}

				output {
					field1: {
						subfield1: Type2[]
					}[]
				}
			}
		}
	`

	expected := `
		""" Type1 Docstring """
		type Type1 {
			field1: string
			""" Field2 Docstring """
			field2: int
			field3: {
				subfield1: datetime
				""" Subfield2 Docstring """
				subfield2: bool[]
			}[]
		}

		""" Type2 Docstring """
		type Type2 {
			field1: {
				""" Subfield1 Docstring """
				subfield1: {
					field1: string
					""" Field2 Docstring """
					field2: int
					field3: {
						subfield1: datetime
						""" Subfield2 Docstring """
						subfield2: bool[]
					}[]
				}[]

				subfield2: {
					field1: string
					""" Field2 Docstring """
					field2: int
					field3: {
						subfield1: datetime
						""" Subfield2 Docstring """
						subfield2: bool[]
					}[]
				}[]
			}[]
		}

		rpc TestService {
			""" Proc1 Docstring """
			proc Proc1 {
				input {
					field1: {
						""" Subfield1 Docstring """
						subfield1: {
							field1: {
								""" Subfield1 Docstring """
								subfield1: {
									field1: string
									""" Field2 Docstring """
									field2: int
									field3: {
										subfield1: datetime
										""" Subfield2 Docstring """
										subfield2: bool[]
									}[]
								}[]

								subfield2: {
									field1: string
									""" Field2 Docstring """
									field2: int
									field3: {
										subfield1: datetime
										""" Subfield2 Docstring """
										subfield2: bool[]
									}[]
								}[]
							}[]
						}[]
					}[]
				}

				output {
					field1: {
						subfield1: {
							field1: {
								""" Subfield1 Docstring """
								subfield1: {
									field1: string
									""" Field2 Docstring """
									field2: int
									field3: {
										subfield1: datetime
										""" Subfield2 Docstring """
										subfield2: bool[]
									}[]
								}[]

								subfield2: {
									field1: string
									""" Field2 Docstring """
									field2: int
									field3: {
										subfield1: datetime
										""" Subfield2 Docstring """
										subfield2: bool[]
									}[]
								}[]
							}[]
						}[]
					}[]
				}
			}

			""" Stream1 Docstring """
			stream Stream1 {
				input {
					field1: {
						""" Subfield1 Docstring """
						subfield1: {
							field1: {
								""" Subfield1 Docstring """
								subfield1: {
									field1: string
									""" Field2 Docstring """
									field2: int
									field3: {
										subfield1: datetime
										""" Subfield2 Docstring """
										subfield2: bool[]
									}[]
								}[]

								subfield2: {
									field1: string
									""" Field2 Docstring """
									field2: int
									field3: {
										subfield1: datetime
										""" Subfield2 Docstring """
										subfield2: bool[]
									}[]
								}[]
							}[]
						}[]
					}[]
				}

				output {
					field1: {
						subfield1: {
							field1: {
								""" Subfield1 Docstring """
								subfield1: {
									field1: string
									""" Field2 Docstring """
									field2: int
									field3: {
										subfield1: datetime
										""" Subfield2 Docstring """
										subfield2: bool[]
									}[]
								}[]

								subfield2: {
									field1: string
									""" Field2 Docstring """
									field2: int
									field3: {
										subfield1: datetime
										""" Subfield2 Docstring """
										subfield2: bool[]
									}[]
								}[]
							}[]
						}[]
					}[]
				}
			}
		}
	`

	schema, err := parser.ParserInstance.ParseString("test.vdl", input)
	if err != nil {
		require.NoError(err, "failed to parse input schema")
	}

	expanded := ExpandTypes(schema)

	gotStr := formatter.FormatSchema(expanded)
	expectedStr, err := formatter.Format("", expected)
	if err != nil {
		require.NoError(err, "failed to format expected schema")
	}

	require.Equal(expectedStr, gotStr, "expanded schema does not match expected")
}

func TestExpandTypes_SimpleTypeReference(t *testing.T) {
	schema, err := parser.ParserInstance.ParseString("test.vdl", `
		type User { id: string name: string }
		rpc TestService {
			proc GetUser { input { userId: string } output { user: User } }
		}
	`)
	if err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}
	expanded := ExpandTypes(schema)
	rpcs := expanded.GetRPCs()
	if len(rpcs) != 1 {
		t.Fatalf("expected 1 rpc, got %d", len(rpcs))
	}
	procs := rpcs[0].GetProcs()
	if len(procs) != 1 {
		t.Fatalf("expected 1 proc, got %d", len(procs))
	}
	proc := procs[0]
	var outputField *ast.Field
	for _, child := range proc.Children {
		if child.Output != nil {
			for _, foc := range child.Output.Children {
				if foc.Field != nil && foc.Field.Name == "user" {
					outputField = foc.Field
					break
				}
			}
		}
	}
	if outputField == nil {
		t.Fatal("output field 'user' not found")
	}
	if outputField.Type.Base.Named != nil {
		t.Errorf("expected field to be expanded to inline object, but got named type: %s", *outputField.Type.Base.Named)
	}
	if outputField.Type.Base.Object == nil {
		t.Fatal("expected field to have inline object")
	}
	fields := extractFieldsFromObject(outputField.Type.Base.Object)
	if len(fields) != 2 {
		t.Fatalf("expected 2 fields in expanded object, got %d", len(fields))
	}
	fieldNames := []string{fields[0].Name, fields[1].Name}
	if !contains(fieldNames, "id") || !contains(fieldNames, "name") {
		t.Errorf("expected fields 'id' and 'name', got %v", fieldNames)
	}
}

func extractFieldsFromObject(obj *ast.FieldTypeObject) []*ast.Field {
	var fields []*ast.Field
	for _, foc := range obj.Children {
		if foc.Field != nil {
			fields = append(fields, foc.Field)
		}
	}
	return fields
}

func contains(slice []string, item string) bool {
	return slices.Contains(slice, item)
}

func TestExpandTypesStr(t *testing.T) {
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

	result, err := ExpandTypesStr("test.vdl", input)
	require.NoError(err, "failed to expand types from string")
	require.NotEmpty(result, "result should not be empty")

	// Verify that User type reference was expanded in Post
	require.Contains(result, "type Post", "should contain Post type")
	require.Contains(result, "author:", "should contain author field")

	// Verify that Post type reference was expanded in proc output
	require.Contains(result, "proc GetPost", "should contain GetPost proc")
	require.Contains(result, "title:", "should contain expanded Post fields in proc")
}

func TestExpandTypesStr_EmptyInput(t *testing.T) {
	require := require.New(t)

	result, err := ExpandTypesStr("test.vdl", "")
	require.NoError(err, "empty input should not error")
	require.Empty(result, "result should be empty for empty input")
}

func TestExpandTypesStr_InvalidInput(t *testing.T) {
	require := require.New(t)

	input := `invalid vdl syntax`

	_, err := ExpandTypesStr("test.vdl", input)
	require.Error(err, "should error on invalid input")
	require.Contains(err.Error(), "parsing", "error should mention parsing")
}

func TestExpandTypesStr_PreservesDocstrings(t *testing.T) {
	require := require.New(t)

	input := `
		""" User type """
		type User {
			""" User ID """
			id: string
		}

		""" Post type """
		type Post {
			""" Post author """
			author: User
		}
	`

	result, err := ExpandTypesStr("test.vdl", input)
	require.NoError(err, "failed to expand types")

	// Verify docstrings are preserved
	require.Contains(result, `""" User type """`, "should preserve User type docstring")
	require.Contains(result, `""" Post type """`, "should preserve Post type docstring")
	require.Contains(result, `""" User ID """`, "should preserve field docstring")
	require.Contains(result, `""" Post author """`, "should preserve author field docstring")
}

func TestExpandTypes_WithMapType(t *testing.T) {
	require := require.New(t)

	input := `
		type User {
			id: string
			name: string
		}

		type Config {
			users: map<User>
			settings: map<string>
		}
	`

	schema, err := parser.ParserInstance.ParseString("test.vdl", input)
	require.NoError(err, "failed to parse input schema")

	expanded := ExpandTypes(schema)

	// The User type inside the map should be expanded
	types := expanded.GetTypes()
	require.Len(types, 2, "expected 2 types")

	var configType *ast.TypeDecl
	for _, t := range types {
		if t.Name == "Config" {
			configType = t
			break
		}
	}
	require.NotNil(configType, "Config type not found")

	// Find the 'users' field
	var usersField *ast.Field
	for _, child := range configType.Children {
		if child.Field != nil && child.Field.Name == "users" {
			usersField = child.Field
			break
		}
	}
	require.NotNil(usersField, "users field not found")
	require.NotNil(usersField.Type.Base.Map, "users field should be a map")
	require.NotNil(usersField.Type.Base.Map.ValueType.Base.Object, "map value should be expanded to inline object")
}

// TestExpandTypes_CircularReference tests that circular type references don't cause stack overflow.
// When TypeA references TypeB and TypeB references TypeA, expansion should stop and keep the reference as named.
func TestExpandTypes_CircularReference(t *testing.T) {
	require := require.New(t)

	input := `
		type TypeA {
			name: string
			refB: TypeB
		}

		type TypeB {
			value: int
			refA: TypeA
		}
	`

	schema, err := parser.ParserInstance.ParseString("test.vdl", input)
	require.NoError(err, "failed to parse input schema")

	// This should NOT panic or hang - it should complete successfully
	expanded := ExpandTypes(schema)
	require.NotNil(expanded, "expanded schema should not be nil")

	// Verify both types exist
	types := expanded.GetTypes()
	require.Len(types, 2, "expected 2 types")

	// Verify TypeA was expanded but the circular reference to TypeA inside TypeB was kept as named
	var typeA *ast.TypeDecl
	for _, t := range types {
		if t.Name == "TypeA" {
			typeA = t
			break
		}
	}
	require.NotNil(typeA, "TypeA not found")

	// Find the refB field in TypeA
	var refBField *ast.Field
	for _, child := range typeA.Children {
		if child.Field != nil && child.Field.Name == "refB" {
			refBField = child.Field
			break
		}
	}
	require.NotNil(refBField, "refB field not found")

	// refB should be expanded to inline object (TypeB's structure)
	require.NotNil(refBField.Type.Base.Object, "refB should be expanded to inline object")

	// Inside the expanded TypeB object, find refA
	var refAField *ast.Field
	for _, child := range refBField.Type.Base.Object.Children {
		if child.Field != nil && child.Field.Name == "refA" {
			refAField = child.Field
			break
		}
	}
	require.NotNil(refAField, "refA field inside expanded TypeB not found")

	// refA should remain as a named reference (to break the cycle)
	require.NotNil(refAField.Type.Base.Named, "refA should remain as named reference to break cycle")
	require.Equal("TypeA", *refAField.Type.Base.Named, "refA should reference TypeA")
}

// TestExpandTypes_SelfReference tests that a type referencing itself doesn't cause stack overflow.
func TestExpandTypes_SelfReference(t *testing.T) {
	require := require.New(t)

	input := `
		type Node {
			value: string
			parent: Node
			children: Node[]
		}
	`

	schema, err := parser.ParserInstance.ParseString("test.vdl", input)
	require.NoError(err, "failed to parse input schema")

	// This should NOT panic or hang
	expanded := ExpandTypes(schema)
	require.NotNil(expanded, "expanded schema should not be nil")

	types := expanded.GetTypes()
	require.Len(types, 1, "expected 1 type")

	nodeType := types[0]
	require.Equal("Node", nodeType.Name)

	// Find the parent field
	var parentField *ast.Field
	for _, child := range nodeType.Children {
		if child.Field != nil && child.Field.Name == "parent" {
			parentField = child.Field
			break
		}
	}
	require.NotNil(parentField, "parent field not found")

	// parent should remain as named reference (self-reference cycle)
	require.NotNil(parentField.Type.Base.Named, "parent should remain as named reference")
	require.Equal("Node", *parentField.Type.Base.Named, "parent should reference Node")

	// Find the children field
	var childrenField *ast.Field
	for _, child := range nodeType.Children {
		if child.Field != nil && child.Field.Name == "children" {
			childrenField = child.Field
			break
		}
	}
	require.NotNil(childrenField, "children field not found")

	// children should also remain as named reference
	require.NotNil(childrenField.Type.Base.Named, "children should remain as named reference")
	require.Equal("Node", *childrenField.Type.Base.Named, "children should reference Node")
	require.True(childrenField.Type.IsArray(), "children should be an array")
}

// TestExpandTypes_SpreadFlattening tests that spread operators are properly flattened.
func TestExpandTypes_SpreadFlattening(t *testing.T) {
	require := require.New(t)

	input := `
		type BaseFields {
			id: string
			createdAt: datetime
		}

		type User {
			...BaseFields
			name: string
			email: string
		}
	`

	schema, err := parser.ParserInstance.ParseString("test.vdl", input)
	require.NoError(err, "failed to parse input schema")

	expanded := ExpandTypes(schema)
	require.NotNil(expanded, "expanded schema should not be nil")

	types := expanded.GetTypes()
	require.Len(types, 2, "expected 2 types")

	// Find User type
	var userType *ast.TypeDecl
	for _, t := range types {
		if t.Name == "User" {
			userType = t
			break
		}
	}
	require.NotNil(userType, "User type not found")

	// User should have 4 fields (id, createdAt from spread + name, email)
	// and NO spread nodes
	fieldCount := 0
	spreadCount := 0
	fieldNames := []string{}

	for _, child := range userType.Children {
		if child.Field != nil {
			fieldCount++
			fieldNames = append(fieldNames, child.Field.Name)
		}
		if child.Spread != nil {
			spreadCount++
		}
	}

	require.Equal(0, spreadCount, "spreads should be flattened, not preserved")
	require.Equal(4, fieldCount, "User should have 4 fields after flattening spread")
	require.Contains(fieldNames, "id", "should contain id from BaseFields")
	require.Contains(fieldNames, "createdAt", "should contain createdAt from BaseFields")
	require.Contains(fieldNames, "name", "should contain name")
	require.Contains(fieldNames, "email", "should contain email")
}

// TestExpandTypes_SpreadChain tests that chained spreads are properly flattened.
func TestExpandTypes_SpreadChain(t *testing.T) {
	require := require.New(t)

	input := `
		type Level1 {
			field1: string
		}

		type Level2 {
			...Level1
			field2: int
		}

		type Level3 {
			...Level2
			field3: bool
		}
	`

	schema, err := parser.ParserInstance.ParseString("test.vdl", input)
	require.NoError(err, "failed to parse input schema")

	expanded := ExpandTypes(schema)

	// Find Level3 type
	var level3Type *ast.TypeDecl
	for _, t := range expanded.GetTypes() {
		if t.Name == "Level3" {
			level3Type = t
			break
		}
	}
	require.NotNil(level3Type, "Level3 type not found")

	// Level3 should have 3 fields: field1, field2, field3
	fieldNames := []string{}
	for _, child := range level3Type.Children {
		if child.Field != nil {
			fieldNames = append(fieldNames, child.Field.Name)
		}
		if child.Spread != nil {
			t.Error("spread should not be present after expansion")
		}
	}

	require.Len(fieldNames, 3, "Level3 should have 3 fields")
	require.Contains(fieldNames, "field1", "should contain field1 from Level1")
	require.Contains(fieldNames, "field2", "should contain field2 from Level2")
	require.Contains(fieldNames, "field3", "should contain field3 from Level3")
}

// TestExpandTypes_SpreadInInputOutput tests that spreads in input/output blocks are flattened.
func TestExpandTypes_SpreadInInputOutput(t *testing.T) {
	require := require.New(t)

	input := `
		type PaginationParams {
			page: int
			limit: int
		}

		type PaginatedResponse {
			totalItems: int
			totalPages: int
		}

		rpc TestService {
			proc ListItems {
				input {
					...PaginationParams
					filterByName: string
				}

				output {
					...PaginatedResponse
					items: string[]
				}
			}
		}
	`

	schema, err := parser.ParserInstance.ParseString("test.vdl", input)
	require.NoError(err, "failed to parse input schema")

	expanded := ExpandTypes(schema)

	rpcs := expanded.GetRPCs()
	require.Len(rpcs, 1, "expected 1 RPC")

	procs := rpcs[0].GetProcs()
	require.Len(procs, 1, "expected 1 proc")

	proc := procs[0]

	// Check input block
	var inputFields []string
	var inputSpreads int
	for _, child := range proc.Children {
		if child.Input != nil {
			for _, ioc := range child.Input.Children {
				if ioc.Field != nil {
					inputFields = append(inputFields, ioc.Field.Name)
				}
				if ioc.Spread != nil {
					inputSpreads++
				}
			}
		}
	}

	require.Equal(0, inputSpreads, "input should have no spreads after expansion")
	require.Len(inputFields, 3, "input should have 3 fields")
	require.Contains(inputFields, "page", "input should contain page from PaginationParams")
	require.Contains(inputFields, "limit", "input should contain limit from PaginationParams")
	require.Contains(inputFields, "filterByName", "input should contain filterByName")

	// Check output block
	var outputFields []string
	var outputSpreads int
	for _, child := range proc.Children {
		if child.Output != nil {
			for _, ioc := range child.Output.Children {
				if ioc.Field != nil {
					outputFields = append(outputFields, ioc.Field.Name)
				}
				if ioc.Spread != nil {
					outputSpreads++
				}
			}
		}
	}

	require.Equal(0, outputSpreads, "output should have no spreads after expansion")
	require.Len(outputFields, 3, "output should have 3 fields")
	require.Contains(outputFields, "totalItems", "output should contain totalItems from PaginatedResponse")
	require.Contains(outputFields, "totalPages", "output should contain totalPages from PaginatedResponse")
	require.Contains(outputFields, "items", "output should contain items")
}

// TestExpandTypes_MultipleSpreads tests that multiple spreads in a single type are all flattened.
func TestExpandTypes_MultipleSpreads(t *testing.T) {
	require := require.New(t)

	input := `
		type Timestamps {
			createdAt: datetime
			updatedAt: datetime
		}

		type Identifiable {
			id: string
		}

		type Entity {
			...Identifiable
			...Timestamps
			name: string
		}
	`

	schema, err := parser.ParserInstance.ParseString("test.vdl", input)
	require.NoError(err, "failed to parse input schema")

	expanded := ExpandTypes(schema)

	// Find Entity type
	var entityType *ast.TypeDecl
	for _, t := range expanded.GetTypes() {
		if t.Name == "Entity" {
			entityType = t
			break
		}
	}
	require.NotNil(entityType, "Entity type not found")

	// Entity should have 4 fields: id, createdAt, updatedAt, name
	fieldNames := []string{}
	for _, child := range entityType.Children {
		if child.Field != nil {
			fieldNames = append(fieldNames, child.Field.Name)
		}
		if child.Spread != nil {
			t.Error("spread should not be present after expansion")
		}
	}

	require.Len(fieldNames, 4, "Entity should have 4 fields")
	require.Contains(fieldNames, "id", "should contain id from Identifiable")
	require.Contains(fieldNames, "createdAt", "should contain createdAt from Timestamps")
	require.Contains(fieldNames, "updatedAt", "should contain updatedAt from Timestamps")
	require.Contains(fieldNames, "name", "should contain name")
}
