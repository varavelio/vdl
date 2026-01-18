package analysis

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/varavelio/vdl/urpc/internal/core/ast"
	"github.com/varavelio/vdl/urpc/internal/core/parser"
	"github.com/varavelio/vdl/urpc/internal/core/vfs"
)

// resolver handles the resolution of includes and external docstrings.
// It traverses the include graph, parses files, and resolves external markdown docstrings.
type resolver struct {
	fs          *vfs.FileSystem
	files       map[string]*File // Resolved files by absolute path
	visited     map[string]bool  // Tracks files currently in the resolution stack (cycle detection)
	diagnostics []Diagnostic
}

// newResolver creates a new resolver instance.
func newResolver(fs *vfs.FileSystem) *resolver {
	return &resolver{
		fs:          fs,
		files:       make(map[string]*File),
		visited:     make(map[string]bool),
		diagnostics: []Diagnostic{},
	}
}

// resolve performs complete resolution starting from an entry point.
// It parses all files, resolves includes recursively, and resolves external docstrings.
// Returns the map of resolved files and any diagnostics encountered.
func (r *resolver) resolve(entryPoint string) (map[string]*File, []Diagnostic) {
	absPath := r.fs.Resolve("", entryPoint)
	r.resolveFile(absPath, nil)
	return r.files, r.diagnostics
}

// resolveFile recursively resolves a single file and its includes.
// includeStack tracks the include chain for cycle detection error messages.
func (r *resolver) resolveFile(absPath string, includeStack []string) {
	// Already fully resolved?
	if _, ok := r.files[absPath]; ok {
		return
	}

	// Currently being resolved? (cycle detection)
	if r.visited[absPath] {
		cycle := append(includeStack, absPath)
		r.diagnostics = append(r.diagnostics, newDiagnostic(
			absPath,
			ast.Position{Filename: absPath, Line: 1, Column: 1},
			ast.Position{Filename: absPath, Line: 1, Column: 1},
			CodeCircularInclude,
			fmt.Sprintf("circular include detected: %s", strings.Join(cycle, " -> ")),
		))
		return
	}

	// Mark as being resolved
	r.visited[absPath] = true

	// Read the file
	content, err := r.fs.ReadFile(absPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			r.diagnostics = append(r.diagnostics, newDiagnostic(
				absPath,
				ast.Position{Filename: absPath, Line: 1, Column: 1},
				ast.Position{Filename: absPath, Line: 1, Column: 1},
				CodeFileNotFound,
				fmt.Sprintf("file not found: %s", absPath),
			))
		} else {
			r.diagnostics = append(r.diagnostics, newDiagnostic(
				absPath,
				ast.Position{Filename: absPath, Line: 1, Column: 1},
				ast.Position{Filename: absPath, Line: 1, Column: 1},
				CodeFileReadError,
				fmt.Sprintf("failed to read file: %v", err),
			))
		}
		delete(r.visited, absPath)
		return
	}

	// Parse the file
	schema, err := parser.ParserInstance.ParseString(absPath, string(content))
	if err != nil {
		pos := ast.Position{Filename: absPath, Line: 1, Column: 1}
		msg := fmt.Sprintf("parse error: %v", err)

		// Try to extract position from parser error
		if pErr, ok := err.(parser.Error); ok {
			pos = pErr.Position()
			msg = fmt.Sprintf("parse error: %s", pErr.Message())
		}

		r.diagnostics = append(r.diagnostics, newDiagnostic(
			absPath,
			pos,
			pos,
			CodeParseError,
			msg,
		))
		delete(r.visited, absPath)
		return
	}

	// Create the file entry
	file := &File{
		Path:     absPath,
		AST:      schema,
		Includes: []string{},
	}

	// Resolve includes
	newStack := append(includeStack, absPath)
	for _, include := range schema.GetIncludes() {
		includePath := r.fs.Resolve(absPath, string(include.Path))
		file.Includes = append(file.Includes, includePath)
		r.resolveFile(includePath, newStack)
	}

	// Resolve external docstrings
	r.resolveDocstrings(schema, absPath)

	// Mark as fully resolved
	r.files[absPath] = file
	delete(r.visited, absPath)
}

// resolveDocstrings finds and resolves all external docstrings in a schema.
// External docstrings reference markdown files (e.g., """ ./docs/intro.md """).
func (r *resolver) resolveDocstrings(schema *ast.Schema, filePath string) {
	// Schema-level docstrings
	for _, doc := range schema.GetDocstrings() {
		r.resolveDocstring(doc, filePath)
	}

	// Type docstrings and field docstrings
	for _, typeDecl := range schema.GetTypes() {
		if typeDecl.Docstring != nil {
			r.resolveDocstring(typeDecl.Docstring, filePath)
		}
		r.resolveFieldDocstrings(typeDecl.Children, filePath)
	}

	// Const docstrings
	for _, constDecl := range schema.GetConsts() {
		if constDecl.Docstring != nil {
			r.resolveDocstring(constDecl.Docstring, filePath)
		}
	}

	// Enum docstrings
	for _, enumDecl := range schema.GetEnums() {
		if enumDecl.Docstring != nil {
			r.resolveDocstring(enumDecl.Docstring, filePath)
		}
	}

	// Pattern docstrings
	for _, patternDecl := range schema.GetPatterns() {
		if patternDecl.Docstring != nil {
			r.resolveDocstring(patternDecl.Docstring, filePath)
		}
	}

	// RPC docstrings
	for _, rpc := range schema.GetRPCs() {
		if rpc.Docstring != nil {
			r.resolveDocstring(rpc.Docstring, filePath)
		}
		r.resolveRPCDocstrings(rpc, filePath)
	}
}

// resolveFieldDocstrings resolves docstrings in type children (fields, comments, spreads).
func (r *resolver) resolveFieldDocstrings(children []*ast.TypeDeclChild, filePath string) {
	for _, child := range children {
		if child.Field != nil && child.Field.Docstring != nil {
			r.resolveDocstring(child.Field.Docstring, filePath)
		}
		// Recursively handle inline objects
		if child.Field != nil && child.Field.Type.Base != nil && child.Field.Type.Base.Object != nil {
			r.resolveFieldDocstrings(child.Field.Type.Base.Object.Children, filePath)
		}
	}
}

// resolveRPCDocstrings resolves docstrings within an RPC declaration.
func (r *resolver) resolveRPCDocstrings(rpc *ast.RPCDecl, filePath string) {
	for _, child := range rpc.Children {
		// Standalone docstrings in RPC
		if child.Docstring != nil {
			r.resolveDocstring(child.Docstring, filePath)
		}

		// Proc docstrings
		if child.Proc != nil {
			if child.Proc.Docstring != nil {
				r.resolveDocstring(child.Proc.Docstring, filePath)
			}
			r.resolveProcOrStreamDocstrings(child.Proc.Children, filePath)
		}

		// Stream docstrings
		if child.Stream != nil {
			if child.Stream.Docstring != nil {
				r.resolveDocstring(child.Stream.Docstring, filePath)
			}
			r.resolveProcOrStreamDocstrings(child.Stream.Children, filePath)
		}
	}
}

// resolveProcOrStreamDocstrings resolves docstrings in proc/stream children.
func (r *resolver) resolveProcOrStreamDocstrings(children []*ast.ProcOrStreamDeclChild, filePath string) {
	for _, child := range children {
		if child.Input != nil {
			r.resolveInputOutputDocstrings(child.Input.Children, filePath)
		}
		if child.Output != nil {
			r.resolveInputOutputDocstrings(child.Output.Children, filePath)
		}
	}
}

// resolveInputOutputDocstrings resolves docstrings in input/output block children.
func (r *resolver) resolveInputOutputDocstrings(children []*ast.InputOutputChild, filePath string) {
	for _, child := range children {
		if child.Field != nil && child.Field.Docstring != nil {
			r.resolveDocstring(child.Field.Docstring, filePath)
		}
		// Recursively handle inline objects in fields
		if child.Field != nil && child.Field.Type.Base != nil && child.Field.Type.Base.Object != nil {
			r.resolveInlineObjectDocstrings(child.Field.Type.Base.Object.Children, filePath)
		}
	}
}

// resolveInlineObjectDocstrings resolves docstrings in inline object children.
func (r *resolver) resolveInlineObjectDocstrings(children []*ast.TypeDeclChild, filePath string) {
	for _, child := range children {
		if child.Field != nil && child.Field.Docstring != nil {
			r.resolveDocstring(child.Field.Docstring, filePath)
		}
		// Recursively handle nested inline objects
		if child.Field != nil && child.Field.Type.Base != nil && child.Field.Type.Base.Object != nil {
			r.resolveInlineObjectDocstrings(child.Field.Type.Base.Object.Children, filePath)
		}
	}
}

// resolveDocstring resolves a single docstring if it references an external file.
func (r *resolver) resolveDocstring(doc *ast.Docstring, filePath string) {
	externalPath, isExternal := doc.GetExternal()
	if !isExternal {
		return
	}

	// Resolve the path relative to the current file
	absPath := r.fs.Resolve(filePath, externalPath)

	// Read the external file
	content, err := r.fs.ReadFile(absPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			r.diagnostics = append(r.diagnostics, newDiagnosticFromPositions(
				doc.Positions,
				CodeDocstringFileNotFound,
				fmt.Sprintf("external docstring file not found: %s", externalPath),
			))
		} else {
			r.diagnostics = append(r.diagnostics, newDiagnosticFromPositions(
				doc.Positions,
				CodeFileReadError,
				fmt.Sprintf("failed to read external docstring file: %v", err),
			))
		}
		return
	}

	// Replace the docstring value with the file content
	doc.Value = ast.DocstringValue(content)
}
