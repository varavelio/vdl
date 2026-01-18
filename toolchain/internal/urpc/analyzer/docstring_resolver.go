package analyzer

import (
	"errors"
	"fmt"
	"os"

	"github.com/uforg/uforpc/urpc/internal/urpc/ast"
)

// docstringResolver is in charge of verifying that all external docstrings
// exists, read them and update all docstrings in the schema with the
// file content of the referenced markdown files.
type docstringResolver struct {
	fileProvider FileProvider
}

// newDocstringResolver creates a new resolver. See resolver for more details.
func newDocstringResolver(fileProvider FileProvider) *docstringResolver {
	return &docstringResolver{fileProvider: fileProvider}
}

// resolve finds all external docstrings and resolves them.
//
// It modifies the provided schema pointer in place and also returns it.
//
// Returns:
//   - The modified schema with all external docstrings resolved.
//   - A list of diagnostics that occurred during the analysis.
//   - The first diagnostic converted to Error interface if any.
func (r *docstringResolver) resolve(astSchema *ast.Schema) (*ast.Schema, []Diagnostic, error) {
	diagnostics := []Diagnostic{}

	for _, docstring := range astSchema.GetDocstrings() {
		diagnostics = r.resolveExternalDocstring(docstring, diagnostics)
	}

	for _, typeDecl := range astSchema.GetTypes() {
		if typeDecl.Docstring != nil {
			diagnostics = r.resolveExternalDocstring(typeDecl.Docstring, diagnostics)
		}

		for _, field := range typeDecl.GetFlattenedFields() {
			if field.Docstring != nil {
				diagnostics = r.resolveExternalDocstring(field.Docstring, diagnostics)
			}
		}
	}

	for _, proc := range astSchema.GetProcs() {
		if proc.Docstring != nil {
			diagnostics = r.resolveExternalDocstring(proc.Docstring, diagnostics)
		}

		for _, child := range proc.Children {
			if child.Input != nil {
				for _, field := range child.Input.GetFlattenedFields() {
					if field.Docstring != nil {
						diagnostics = r.resolveExternalDocstring(field.Docstring, diagnostics)
					}
				}
			}

			if child.Output != nil {
				for _, field := range child.Output.GetFlattenedFields() {
					if field.Docstring != nil {
						diagnostics = r.resolveExternalDocstring(field.Docstring, diagnostics)
					}
				}
			}
		}
	}

	for _, stream := range astSchema.GetStreams() {
		if stream.Docstring != nil {
			diagnostics = r.resolveExternalDocstring(stream.Docstring, diagnostics)
		}

		for _, child := range stream.Children {
			if child.Input != nil {
				for _, field := range child.Input.GetFlattenedFields() {
					if field.Docstring != nil {
						diagnostics = r.resolveExternalDocstring(field.Docstring, diagnostics)
					}
				}
			}

			if child.Output != nil {
				for _, field := range child.Output.GetFlattenedFields() {
					if field.Docstring != nil {
						diagnostics = r.resolveExternalDocstring(field.Docstring, diagnostics)
					}
				}
			}
		}
	}

	// Return the first diagnostic as error if any
	if len(diagnostics) > 0 {
		return astSchema, diagnostics, diagnostics[0]
	}
	return astSchema, nil, nil
}

// resolveExternalDocstring is the logic to resolve a single external docstring
func (r *docstringResolver) resolveExternalDocstring(docstring *ast.Docstring, diagnostics []Diagnostic) []Diagnostic {
	externalPath, isExternal := docstring.GetExternal()
	if !isExternal {
		return diagnostics
	}

	content, _, err := r.fileProvider.GetFileAndHash(docstring.Pos.Filename, externalPath)
	if errors.Is(err, os.ErrNotExist) {
		return append(diagnostics, Diagnostic{
			Positions: Positions{
				Pos:    docstring.Pos,
				EndPos: docstring.EndPos,
			},
			Message: fmt.Sprintf("external markdown file not found: %s", externalPath),
		})
	}
	if err != nil {
		return append(diagnostics, Diagnostic{
			Positions: Positions{
				Pos:    docstring.Pos,
				EndPos: docstring.EndPos,
			},
			Message: fmt.Sprintf("error reading external markdown file: %v", err),
		})
	}

	docstring.Value = content
	return diagnostics
}
