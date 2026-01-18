package transform

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/varavelio/vdl/toolchain/internal/urpc/ast"
	"github.com/varavelio/vdl/toolchain/internal/urpc/formatter"
	"github.com/varavelio/vdl/toolchain/internal/urpc/parser"
)

func TestExpandTypes(t *testing.T) {
	require := require.New(t)

	input := `
		version 1

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
	`

	expected := `
		version 1

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
	`

	schema, err := parser.ParserInstance.ParseString("test.urpc", input)
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
	schema, err := parser.ParserInstance.ParseString("test.urpc", `
		version 1
		type User { id: string name: string }
		proc GetUser { input { userId: string } output { user: User } }
	`)
	if err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}
	expanded := ExpandTypes(schema)
	procs := expanded.GetProcs()
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
		version 1

		type User {
			id: string
			name: string
		}

		type Post {
			title: string
			author: User
		}

		proc GetPost {
			input {
				postId: string
			}

			output {
				post: Post
			}
		}
	`

	result, err := ExpandTypesStr("test.urpc", input)
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

	result, err := ExpandTypesStr("test.urpc", "")
	require.NoError(err, "empty input should not error")
	require.Empty(result, "result should be empty for empty input")
}

func TestExpandTypesStr_InvalidInput(t *testing.T) {
	require := require.New(t)

	input := `invalid urpc syntax`

	_, err := ExpandTypesStr("test.urpc", input)
	require.Error(err, "should error on invalid input")
	require.Contains(err.Error(), "parsing", "error should mention parsing")
}

func TestExpandTypesStr_PreservesDocstrings(t *testing.T) {
	require := require.New(t)

	input := `
		version 1

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

	result, err := ExpandTypesStr("test.urpc", input)
	require.NoError(err, "failed to expand types")

	// Verify docstrings are preserved
	require.Contains(result, `""" User type """`, "should preserve User type docstring")
	require.Contains(result, `""" Post type """`, "should preserve Post type docstring")
	require.Contains(result, `""" User ID """`, "should preserve field docstring")
	require.Contains(result, `""" Post author """`, "should preserve author field docstring")
}
