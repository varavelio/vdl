package transform

import (
	"github.com/varavelio/vdl/toolchain/internal/core/analysis"
	"github.com/varavelio/vdl/toolchain/internal/core/ast"
	"github.com/varavelio/vdl/toolchain/internal/formatter"
)

// MergeAndFormat merges all files from a Program into a single formatted schema.
// It combines all AST children from all included files, removing include statements,
// and returns a single unified formatted schema string.
//
// Files are processed in topological order starting from the entry point,
// with included files' content appearing before the file that includes them.
func MergeAndFormat(program *analysis.Program) string {
	if program == nil || len(program.Files) == 0 {
		return ""
	}

	merged := MergeSchemas(program)
	return formatter.FormatSchema(merged)
}

// MergeSchemas merges all files from a Program into a single AST Schema.
// It combines all AST children from all included files, removing include statements.
//
// Files are processed in topological order starting from the entry point,
// with included files' content appearing before the file that includes them.
func MergeSchemas(program *analysis.Program) *ast.Schema {
	if program == nil || len(program.Files) == 0 {
		return &ast.Schema{}
	}

	return mergeFiles(program, program.EntryPoint, make(map[string]bool))
}

// mergeFiles recursively merges files starting from the given file path.
// It processes includes first (depth-first) so that dependencies appear before
// the code that uses them.
func mergeFiles(program *analysis.Program, filePath string, visited map[string]bool) *ast.Schema {
	if visited[filePath] {
		return &ast.Schema{}
	}
	visited[filePath] = true

	file, ok := program.Files[filePath]
	if !ok || file.AST == nil {
		return &ast.Schema{}
	}

	merged := &ast.Schema{
		Children: []*ast.SchemaChild{},
	}

	// First, recursively merge all included files
	for _, includePath := range file.Includes {
		includedSchema := mergeFiles(program, includePath, visited)
		merged.Children = append(merged.Children, includedSchema.Children...)
	}

	// Then, add this file's children (excluding include statements)
	for _, child := range file.AST.Children {
		if child.Kind() != ast.SchemaChildKindInclude {
			merged.Children = append(merged.Children, child)
		}
	}

	return merged
}
