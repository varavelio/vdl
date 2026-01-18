package analyzer

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/uforg/uforpc/urpc/internal/urpc/ast"
	"github.com/uforg/uforpc/urpc/internal/urpc/parser"
)

// parseSchema parses the given input string into an ast.Schema.
// this is a test helper function.
func parseSchema(input string) (*ast.Schema, error) {
	schema, err := parser.ParserInstance.ParseString("test.urpc", input)
	if err != nil {
		return nil, err
	}

	return schema, nil
}

func TestSemanalyzer_ValidVersion(t *testing.T) {
	input := `
		version 1
	`
	combinedSchema, err := parseSchema(input)
	require.NoError(t, err)

	analyzer := newSemanalyzer(combinedSchema)
	errors, err := analyzer.analyze()

	require.NoError(t, err)
	require.Empty(t, errors)
}

func TestSemanalyzer_ValidCustomType(t *testing.T) {
	input := `
		version 1

		type User {
		  id: string
		  age: int
		}
	`
	combinedSchema, err := parseSchema(input)
	require.NoError(t, err)

	analyzer := newSemanalyzer(combinedSchema)
	errors, err := analyzer.analyze()

	require.NoError(t, err)
	require.Empty(t, errors)
}

func TestSemanalyzer_DuplicateCustomType(t *testing.T) {
	input := `
		version 1

		type User {
		  id: string
		  age: int
		}

		// Duplicate type name
		type User {
		  name: string
		}
	`
	combinedSchema, err := parseSchema(input)
	require.NoError(t, err)

	analyzer := newSemanalyzer(combinedSchema)
	errors, err := analyzer.analyze()

	require.Error(t, err)
	require.Len(t, errors, 1)
	require.Contains(t, errors[0].Message, "is already declared")
}

func TestSemanalyzer_InvalidCustomTypeName(t *testing.T) {
	input := `
		version 1

		// camelCase, should be PascalCase
		type invalidTypeName {
		  id: string
		}
	`
	combinedSchema, err := parseSchema(input)
	require.NoError(t, err)

	analyzer := newSemanalyzer(combinedSchema)
	errors, err := analyzer.analyze()

	require.Error(t, err)
	require.Len(t, errors, 1)
	require.Contains(t, errors[0].Message, "must be in PascalCase")
}

func TestSemanalyzer_ValidProcedure(t *testing.T) {
	input := `
		version 1

		type User {
		  id: string
		  age: int
		}

		proc GetUser {
		  input {
		    userId: string
		  }

		  output {
		    user: User
		  }
		}
	`
	combinedSchema, err := parseSchema(input)
	require.NoError(t, err)

	analyzer := newSemanalyzer(combinedSchema)
	errors, err := analyzer.analyze()

	require.NoError(t, err)
	require.Empty(t, errors)
}

func TestSemanalyzer_DuplicateProcedure(t *testing.T) {
	input := `
		version 1

		proc GetUser {
		  input {
		    userId: string
		  }
		}

		// Duplicate procedure name
		proc GetUser {
		  input {
		    id: string
		  }
		}
	`
	combinedSchema, err := parseSchema(input)
	require.NoError(t, err)

	analyzer := newSemanalyzer(combinedSchema)
	errors, err := analyzer.analyze()

	require.Error(t, err)
	require.Len(t, errors, 1)
	require.Contains(t, errors[0].Message, "is already declared")
}

func TestSemanalyzer_DuplicateStream(t *testing.T) {
	input := `
		version 1

		stream GetUser {
		  input {
		    userId: string
		  }
		}

		stream GetUser {
		  input {
		    id: string
		  }
		}
	`
	combinedSchema, err := parseSchema(input)
	require.NoError(t, err)

	analyzer := newSemanalyzer(combinedSchema)
	errors, err := analyzer.analyze()

	require.Error(t, err)
	require.Len(t, errors, 1)
	require.Contains(t, errors[0].Message, "is already declared")
}

func TestSemanalyzer_InvalidProcedureName(t *testing.T) {
	input := `
		version 1

		// camelCase, should be PascalCase
		proc getUser {
		  input {
		    userId: string
		  }
		}
	`
	combinedSchema, err := parseSchema(input)
	require.NoError(t, err)

	analyzer := newSemanalyzer(combinedSchema)
	errors, err := analyzer.analyze()

	require.Error(t, err)
	require.Len(t, errors, 1)
	require.Contains(t, errors[0].Message, "must be in PascalCase")
}

func TestSemanalyzer_NonExistentTypeReference(t *testing.T) {
	input := `
		version 1

		// This type doesn't exist
		proc GetPost {
		  output {
		    post: Post
		  }
		}
	`
	combinedSchema, err := parseSchema(input)
	require.NoError(t, err)

	analyzer := newSemanalyzer(combinedSchema)
	errors, err := analyzer.analyze()

	require.Error(t, err)
	require.Len(t, errors, 1)
	require.Contains(t, errors[0].Message, "is not declared")
}

func TestSemanalyzer_ValidInlineObject(t *testing.T) {
	input := `
		version 1

		type User {
		  id: string
		  address: {
		    street: string
		    city: string
		    zipCode: string
		  }
		}
	`
	combinedSchema, err := parseSchema(input)
	require.NoError(t, err)

	analyzer := newSemanalyzer(combinedSchema)
	errors, err := analyzer.analyze()

	require.NoError(t, err)
	require.Empty(t, errors)
}

func TestSemanalyzer_ValidNestedInlineObject(t *testing.T) {
	input := `
		version 1

		type User {
		  id: string
		  contact: {
		    email: string
		    address: {
		      street: string
		      city: string
		      country: string
		    }
		  }
		}
	`
	combinedSchema, err := parseSchema(input)
	require.NoError(t, err)

	analyzer := newSemanalyzer(combinedSchema)
	errors, err := analyzer.analyze()

	require.NoError(t, err)
	require.Empty(t, errors)
}

func TestSemanalyzer_ValidArray(t *testing.T) {
	input := `
		version 1

		type Matrix {
		  data: int[]
		}
	`
	combinedSchema, err := parseSchema(input)
	require.NoError(t, err)

	analyzer := newSemanalyzer(combinedSchema)
	errors, err := analyzer.analyze()

	require.NoError(t, err)
	require.Empty(t, errors)
}

func TestSemanalyzer_ValidArrayOfObjects(t *testing.T) {
	input := `
		version 1

		type User {
		  id: string
		  addresses: {
		    street: string
		    city: string
		  }[]
		}
	`
	combinedSchema, err := parseSchema(input)
	require.NoError(t, err)

	analyzer := newSemanalyzer(combinedSchema)
	errors, err := analyzer.analyze()

	require.NoError(t, err)
	require.Empty(t, errors)
}

func TestSemanalyzer_EnumValidation(t *testing.T) {
	input := `
		version 1

		type Product {
		  status: string
		  priority: int
		}
	`
	combinedSchema, err := parseSchema(input)
	require.NoError(t, err)

	analyzer := newSemanalyzer(combinedSchema)
	errors, err := analyzer.analyze()

	require.NoError(t, err)
	require.Empty(t, errors)
}

func TestSemanalyzer_DatetimeValidation(t *testing.T) {
	input := `
		version 1

		type Event {
		  startDate: datetime
		  endDate: datetime
		}
	`
	combinedSchema, err := parseSchema(input)
	require.NoError(t, err)

	analyzer := newSemanalyzer(combinedSchema)
	errors, err := analyzer.analyze()

	require.NoError(t, err)
	require.Empty(t, errors)
}

func TestSemanalyzer_OptionalFields(t *testing.T) {
	input := `
		version 1

		type User {
		  id: string
		  email: string
		  phone?: string
		  address?: {
		    street: string
		    city: string
		  }
		}
	`
	combinedSchema, err := parseSchema(input)
	require.NoError(t, err)

	analyzer := newSemanalyzer(combinedSchema)
	errors, err := analyzer.analyze()

	require.NoError(t, err)
	require.Empty(t, errors)
}

func TestSemanalyzer_CircularTypeDependency(t *testing.T) {
	input := `
			version 1

			// Circular dependency: User -> Post -> User
			type User {
			  id: string
			  posts: Post[]
			}

			type Post {
			  id: string
			  author: User  // This creates a circular dependency
			}
		`
	combinedSchema, err := parseSchema(input)
	require.NoError(t, err)

	analyzer := newSemanalyzer(combinedSchema)
	errors, err := analyzer.analyze()

	require.Error(t, err)
	require.NotEmpty(t, errors)

	// Check that at least one error contains the expected message
	found := false
	for _, diag := range errors {
		if strings.Contains(diag.Message, "circular dependency detected between types") {
			found = true
			break
		}
	}
	require.True(t, found, "Expected to find an error about circular dependency")
}

func TestSemanalyzer_CircularTypeDependencyWithOptionalField(t *testing.T) {
	input := `
			version 1

			// Circular dependency with optional field
			type User {
			  id: string
			  posts: Post[]
			}

			type Post {
			  id: string
			  author?: User  // This creates a circular dependency
			}
		`
	combinedSchema, err := parseSchema(input)
	require.NoError(t, err)

	analyzer := newSemanalyzer(combinedSchema)
	errors, err := analyzer.analyze()

	require.Error(t, err)
	require.NotEmpty(t, errors)

	// Check that at least one error contains the expected message
	found := false
	for _, diag := range errors {
		if strings.Contains(diag.Message, "circular dependency detected between types") {
			found = true
			break
		}
	}
	require.True(t, found, "Expected to find an error about circular dependency")
}

func TestSemanalyzer_ProcWithMultipleInputSections(t *testing.T) {
	input := `
			version 1

			// Procedure with multiple 'input' sections
			proc InvalidProc {
			  input {
			    id: string
			  }

			  input {  // Duplicate 'input' section
			    name: string
			  }
			}
		`
	combinedSchema, err := parseSchema(input)
	require.NoError(t, err)

	analyzer := newSemanalyzer(combinedSchema)
	errors, err := analyzer.analyze()

	require.Error(t, err)
	require.Len(t, errors, 1)
	require.Contains(t, errors[0].Message, "cannot have more than one 'input' section")
}

func TestSemanalyzer_ProcWithMultipleOutputSections(t *testing.T) {
	input := `
			version 1

			// Procedure with multiple 'output' sections
			proc InvalidProc {
			  output {
			    success: bool
			  }

			  output {  // Duplicate 'output' section
			    message: string
			  }
			}
		`
	combinedSchema, err := parseSchema(input)
	require.NoError(t, err)

	analyzer := newSemanalyzer(combinedSchema)
	errors, err := analyzer.analyze()

	require.Error(t, err)
	require.Len(t, errors, 1)
	require.Contains(t, errors[0].Message, "cannot have more than one 'output' section")
}

func TestSemanalyzer_StreamWithMultipleInputSections(t *testing.T) {
	input := `
			version 1

			// Stream with multiple 'input' sections
			stream InvalidStream {
			  input {
			    id: string
			  }

			  input {  // Duplicate 'input' section
			    name: string
			  }
			}
		`
	combinedSchema, err := parseSchema(input)
	require.NoError(t, err)

	analyzer := newSemanalyzer(combinedSchema)
	errors, err := analyzer.analyze()

	require.Error(t, err)
	require.Len(t, errors, 1)
	require.Contains(t, errors[0].Message, "cannot have more than one 'input' section")
}

func TestSemanalyzer_StreamWithMultipleOutputSections(t *testing.T) {
	input := `
			version 1

			// Stream with multiple 'output' sections
			stream InvalidStream {
			  output {
			    success: bool
			  }

			  output {  // Duplicate 'output' section
			    message: string
			  }
			}
		`
	combinedSchema, err := parseSchema(input)
	require.NoError(t, err)

	analyzer := newSemanalyzer(combinedSchema)
	errors, err := analyzer.analyze()

	require.Error(t, err)
	require.Len(t, errors, 1)
	require.Contains(t, errors[0].Message, "cannot have more than one 'output' section")
}

func TestSemanalyzer_CompleteSchema(t *testing.T) {
	input := `
		version 1

		type Address {
		  street: string
		  city: string
		  zipCode: string
		}

		type BaseUser {
		  id: string
		  username: string
		}

		type User {
			user: BaseUser
		  email: string
		  password: string
		  age: int
		  isActive: bool
		  address: Address
		  tags: string[]
		  other: {
		    lastLogin: datetime
		    preferences: {
		      theme: string
		      notifications: bool
		    }
		  }
		}

		proc CreateUser {
		  input {
		    user: User
		  }

		  output {
		    success: bool
		    userId: string
		    errors: string[]
		  }
		}

		proc GetUser {
		  input {
		    userId: string
		  }

		  output {
		    user: User
		  }
		}

		stream MyStream {
		  input {
		    userId: string
		  }

		  output {
		    user: User
		  }
		}
	`
	combinedSchema, err := parseSchema(input)
	require.NoError(t, err)

	analyzer := newSemanalyzer(combinedSchema)
	errors, err := analyzer.analyze()

	require.NoError(t, err)
	require.Empty(t, errors)
}
