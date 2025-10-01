package transform

import (
	"fmt"

	"github.com/uforg/uforpc/urpc/internal/urpc/ast"
)

// ExtractType extracts a specific type declaration from the schema by name.
// Returns only the type declaration, without dependencies or additional nodes.
func ExtractType(schema *ast.Schema, typeName string) (*ast.TypeDecl, error) {
	for _, typeDecl := range schema.GetTypes() {
		if typeDecl.Name == typeName {
			return typeDecl, nil
		}
	}
	return nil, fmt.Errorf("type '%s' not found in schema", typeName)
}

// ExtractProc extracts a specific proc declaration from the schema by name.
// Returns only the proc declaration, without dependencies or additional nodes.
func ExtractProc(schema *ast.Schema, procName string) (*ast.ProcDecl, error) {
	for _, procDecl := range schema.GetProcs() {
		if procDecl.Name == procName {
			return procDecl, nil
		}
	}
	return nil, fmt.Errorf("proc '%s' not found in schema", procName)
}

// ExtractStream extracts a specific stream declaration from the schema by name.
// Returns only the stream declaration, without dependencies or additional nodes.
func ExtractStream(schema *ast.Schema, streamName string) (*ast.StreamDecl, error) {
	for _, streamDecl := range schema.GetStreams() {
		if streamDecl.Name == streamName {
			return streamDecl, nil
		}
	}
	return nil, fmt.Errorf("stream '%s' not found in schema", streamName)
}
