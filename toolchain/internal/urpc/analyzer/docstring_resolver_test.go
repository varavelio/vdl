package analyzer

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/uforg/uforpc/urpc/internal/urpc/parser"
)

// mockFileProvider is a mock implementation of FileProvider for testing
type mockFileProvider struct {
	files map[string]string
}

func (m *mockFileProvider) GetFileAndHash(relativeTo string, path string) (string, string, error) {
	// Try with the path as is
	if content, ok := m.files[path]; ok {
		return content, "mock-hash", nil
	}

	// If relativeTo is provided, try with the path relative to it
	if relativeTo != "" {
		relativePath := relativeTo + "/" + path
		if content, ok := m.files[relativePath]; ok {
			return content, "mock-hash", nil
		}
	}

	return "", "", os.ErrNotExist
}

func TestResolver(t *testing.T) {
	t.Run("Schema with external markdown", func(t *testing.T) {
		// Create a mock file provider with a schema that references an external markdown
		provider := &mockFileProvider{
			files: map[string]string{
				"/main.urpc": `
						version 1

						""" docs/user.md """

						type User {
							id: string
							name: string
						}

						proc GetUser {
							input {
								id: string
							}
							output {
								user: User
							}
						}
					`,
				"/main.urpc/docs/user.md": "# User Documentation\n\nThis is the documentation for the User type.\n",
			},
		}

		// Parse the schema entry point
		astSchema, err := parser.ParserInstance.ParseString("/main.urpc", provider.files["/main.urpc"])
		require.NoError(t, err)

		// Resolve the schema docstrings
		resolver := newDocstringResolver(provider)
		astSchema, diagnostics, err := resolver.resolve(astSchema)
		require.NoError(t, err)
		require.Empty(t, diagnostics)

		// Verify that the docstring was resolved
		require.NotNil(t, astSchema)
		require.Len(t, astSchema.GetDocstrings(), 1)
		require.Equal(t, "# User Documentation\n\nThis is the documentation for the User type.\n", astSchema.GetDocstrings()[0].Value)
	})

	t.Run("Schema with missing external markdown", func(t *testing.T) {
		// Create a mock file provider with a schema that references a non-existent external markdown
		provider := &mockFileProvider{
			files: map[string]string{
				"/main.urpc": `
						version 1

						""" docs/missing.md """

						type User {
							id: string
							name: string
						}

						proc GetUser {
							input {
								id: string
							}
							output {
								user: User
							}
						}
					`,
			},
		}

		// Parse the schema entry point
		astSchema, err := parser.ParserInstance.ParseString("/main.urpc", provider.files["/main.urpc"])
		require.NoError(t, err)

		// Resolve the schema docstrings
		resolver := newDocstringResolver(provider)
		astSchema, diagnostics, err := resolver.resolve(astSchema)
		// Expect an error for missing external markdown
		require.Error(t, err)
		require.Contains(t, err.Error(), "external markdown file not found")

		// Verify that error diagnostics were generated
		require.NotEmpty(t, diagnostics)
		require.Contains(t, diagnostics[0].Message, "external markdown file not found")

		// Verify that the schema was still processed
		require.NotNil(t, astSchema)
		require.Len(t, astSchema.GetTypes(), 1)
		require.Len(t, astSchema.GetProcs(), 1)
	})

	t.Run("Schema with external markdowns in different nodes", func(t *testing.T) {
		// Create a mock file provider with a schema that has external markdowns in different types of nodes
		provider := &mockFileProvider{
			files: map[string]string{
				"/main.urpc": `
						version 1

						""" docs/overview.md """

						""" docs/type.md """
						type User {
							""" docs/field.md """
							id: string
							email: string
							foo?: {
								bar?: {
									""" docs/field.md """
									baz?: string
								}[]
							}[]
						}

						""" docs/proc.md """
						proc GetUser {
							input {
								""" docs/field.md """
								id: string
								foo?: {
									bar?: {
										""" docs/field.md """
										baz?: string
									}[]
								}[]
							}
							output {
								user: User
							}
						}
						
						""" docs/stream.md """
						stream MyStream {
							input {
								id: string
							}
							output {
								""" docs/field.md """
								user: User
								foo?: {
									bar?: {
										""" docs/field.md """
										baz?: string
									}[]
								}[]
							}
						}
					`,
				"/main.urpc/docs/overview.md": "# API Overview\n\nThis is the main API documentation.\n",
				"/main.urpc/docs/type.md":     "# User Type\n\nRepresents a user in the system.\n",
				"/main.urpc/docs/proc.md":     "# GetUser Procedure\n\nRetrieves a user by ID.\n",
				"/main.urpc/docs/stream.md":   "# MyStream Stream\n\nStream for user events.\n",
				"/main.urpc/docs/field.md":    "Field docstring",
			},
		}

		// Parse the schema entry point
		astSchema, err := parser.ParserInstance.ParseString("/main.urpc", provider.files["/main.urpc"])
		require.NoError(t, err)

		// Resolve the schema docstrings
		resolver := newDocstringResolver(provider)
		astSchema, diagnostics, err := resolver.resolve(astSchema)
		require.NoError(t, err)
		require.Empty(t, diagnostics)

		// Verify that all docstrings were resolved
		require.NotNil(t, astSchema)

		// Check standalone docstring
		require.Len(t, astSchema.GetDocstrings(), 1)
		require.Equal(t, "# API Overview\n\nThis is the main API documentation.\n", astSchema.GetDocstrings()[0].Value)

		// Check type docstring
		require.Len(t, astSchema.GetTypes(), 1)
		require.Equal(t, "# User Type\n\nRepresents a user in the system.\n", astSchema.GetTypes()[0].Docstring.Value)

		// Check type fields docstrings
		require.Equal(t, "Field docstring", astSchema.GetTypes()[0].Children[0].Field.Docstring.Value)
		require.Equal(t, "Field docstring", astSchema.GetTypes()[0].Children[2].Field.Type.Base.Object.Children[0].Field.Type.Base.Object.Children[0].Field.Docstring.Value)

		// Check proc docstring
		require.Len(t, astSchema.GetProcs(), 1)
		require.Equal(t, "# GetUser Procedure\n\nRetrieves a user by ID.\n", astSchema.GetProcs()[0].Docstring.Value)

		// Check proc input fields docstrings
		require.Equal(t, "Field docstring", astSchema.GetProcs()[0].Children[0].Input.Children[0].Field.Docstring.Value)
		require.Equal(t, "Field docstring", astSchema.GetProcs()[0].Children[0].Input.Children[1].Field.Type.Base.Object.Children[0].Field.Type.Base.Object.Children[0].Field.Docstring.Value)

		// Check stream docstring
		require.Len(t, astSchema.GetStreams(), 1)
		require.Equal(t, "# MyStream Stream\n\nStream for user events.\n", astSchema.GetStreams()[0].Docstring.Value)

		// Check stream output fields docstrings
		require.Equal(t, "Field docstring", astSchema.GetStreams()[0].Children[1].Output.Children[0].Field.Docstring.Value)
		require.Equal(t, "Field docstring", astSchema.GetStreams()[0].Children[1].Output.Children[1].Field.Type.Base.Object.Children[0].Field.Type.Base.Object.Children[0].Field.Docstring.Value)
	})

	t.Run("Basic schema with no imports", func(t *testing.T) {
		// Create a mock file provider with a single file
		provider := &mockFileProvider{
			files: map[string]string{
				"/main.urpc": `
					version 1

					type User {
						id: string
						name: string
					}

					proc GetUser {
						input {
							id: string
						}
						output {
							user: User
						}
					}
				`,
			},
		}

		// Parse the schema entry point
		astSchema, err := parser.ParserInstance.ParseString("/main.urpc", provider.files["/main.urpc"])
		require.NoError(t, err)

		// Resolve the schema docstrings
		resolver := newDocstringResolver(provider)
		astSchema, diagnostics, err := resolver.resolve(astSchema)
		require.NoError(t, err)
		require.Empty(t, diagnostics)

		// Verify the combined schema
		require.NotNil(t, astSchema)
		require.Len(t, astSchema.GetTypes(), 1)
		require.Len(t, astSchema.GetProcs(), 1)
		require.Equal(t, "User", astSchema.GetTypes()[0].Name)
		require.Equal(t, "GetUser", astSchema.GetProcs()[0].Name)
	})
}
