package transform

import (
	"fmt"
	"strings"

	"github.com/varavelio/vdl/toolchain/internal/core/ast"
	"github.com/varavelio/vdl/toolchain/internal/core/formatter"
	"github.com/varavelio/vdl/toolchain/internal/core/parser"
)

// ExtractTypeStr extracts a specific type declaration from the URPC schema by name.
// It takes a URPC schema as a string and returns only the extracted type as a formatted string.
func ExtractTypeStr(filename, content, typeName string) (string, error) {
	if strings.TrimSpace(content) == "" {
		return "", fmt.Errorf("empty schema content")
	}

	schema, err := parser.ParserInstance.ParseString(filename, content)
	if err != nil {
		return "", fmt.Errorf("error parsing URPC: %w", err)
	}

	typeDecl, err := ExtractType(schema, typeName)
	if err != nil {
		return "", err
	}

	// Create a minimal schema with just the version and extracted type
	extractedSchema := &ast.Schema{
		Children: []*ast.SchemaChild{
			{Type: typeDecl},
		},
	}

	return formatter.FormatSchema(extractedSchema), nil
}

// ExtractProcStr extracts a specific proc declaration from the URPC schema by name.
// It takes a URPC schema as a string and returns only the extracted proc as a formatted string.
func ExtractProcStr(filename, content, procName string) (string, error) {
	if strings.TrimSpace(content) == "" {
		return "", fmt.Errorf("empty schema content")
	}

	schema, err := parser.ParserInstance.ParseString(filename, content)
	if err != nil {
		return "", fmt.Errorf("error parsing URPC: %w", err)
	}

	procDecl, err := ExtractProc(schema, procName)
	if err != nil {
		return "", err
	}

	// Create a minimal schema with just the version and extracted proc
	extractedSchema := &ast.Schema{
		Children: []*ast.SchemaChild{
			{Proc: procDecl},
		},
	}

	return formatter.FormatSchema(extractedSchema), nil
}

// ExtractStreamStr extracts a specific stream declaration from the URPC schema by name.
// It takes a URPC schema as a string and returns only the extracted stream as a formatted string.
func ExtractStreamStr(filename, content, streamName string) (string, error) {
	if strings.TrimSpace(content) == "" {
		return "", fmt.Errorf("empty schema content")
	}

	schema, err := parser.ParserInstance.ParseString(filename, content)
	if err != nil {
		return "", fmt.Errorf("error parsing URPC: %w", err)
	}

	streamDecl, err := ExtractStream(schema, streamName)
	if err != nil {
		return "", err
	}

	// Create a minimal schema with just the version and extracted stream
	extractedSchema := &ast.Schema{
		Children: []*ast.SchemaChild{
			{Stream: streamDecl},
		},
	}

	return formatter.FormatSchema(extractedSchema), nil
}

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
