package transform

import (
	"fmt"
	"strings"

	"github.com/varavelio/vdl/toolchain/internal/core/ast"
	"github.com/varavelio/vdl/toolchain/internal/core/parser"
	"github.com/varavelio/vdl/toolchain/internal/formatter"
)

// ExtractTypeStr extracts a specific type declaration from the VDL schema by name.
// It takes a VDL schema as a string and returns only the extracted type as a formatted string.
func ExtractTypeStr(filename, content, typeName string) (string, error) {
	if strings.TrimSpace(content) == "" {
		return "", fmt.Errorf("empty schema content")
	}

	schema, err := parser.ParserInstance.ParseString(filename, content)
	if err != nil {
		return "", fmt.Errorf("error parsing VDL: %w", err)
	}

	typeDecl, err := ExtractType(schema, typeName)
	if err != nil {
		return "", err
	}

	// Create a minimal schema with just the extracted type
	extractedSchema := &ast.Schema{
		Children: []*ast.SchemaChild{
			{Type: typeDecl},
		},
	}

	return formatter.FormatSchema(extractedSchema), nil
}

// ExtractProcStr extracts a specific proc declaration from the VDL schema by RPC and proc name.
// It takes a VDL schema as a string and returns only the extracted proc (wrapped in its RPC) as a formatted string.
func ExtractProcStr(filename, content, rpcName, procName string) (string, error) {
	if strings.TrimSpace(content) == "" {
		return "", fmt.Errorf("empty schema content")
	}

	schema, err := parser.ParserInstance.ParseString(filename, content)
	if err != nil {
		return "", fmt.Errorf("error parsing VDL: %w", err)
	}

	procDecl, err := ExtractProc(schema, rpcName, procName)
	if err != nil {
		return "", err
	}

	// Create a minimal schema with the proc wrapped in its RPC
	extractedSchema := &ast.Schema{
		Children: []*ast.SchemaChild{
			{
				RPC: &ast.RPCDecl{
					Name: rpcName,
					Children: []*ast.RPCChild{
						{Proc: procDecl},
					},
				},
			},
		},
	}

	return formatter.FormatSchema(extractedSchema), nil
}

// ExtractStreamStr extracts a specific stream declaration from the VDL schema by RPC and stream name.
// It takes a VDL schema as a string and returns only the extracted stream (wrapped in its RPC) as a formatted string.
func ExtractStreamStr(filename, content, rpcName, streamName string) (string, error) {
	if strings.TrimSpace(content) == "" {
		return "", fmt.Errorf("empty schema content")
	}

	schema, err := parser.ParserInstance.ParseString(filename, content)
	if err != nil {
		return "", fmt.Errorf("error parsing VDL: %w", err)
	}

	streamDecl, err := ExtractStream(schema, rpcName, streamName)
	if err != nil {
		return "", err
	}

	// Create a minimal schema with the stream wrapped in its RPC
	extractedSchema := &ast.Schema{
		Children: []*ast.SchemaChild{
			{
				RPC: &ast.RPCDecl{
					Name: rpcName,
					Children: []*ast.RPCChild{
						{Stream: streamDecl},
					},
				},
			},
		},
	}

	return formatter.FormatSchema(extractedSchema), nil
}

// ExtractRPCStr extracts a specific RPC declaration from the VDL schema by name.
// It takes a VDL schema as a string and returns only the extracted RPC as a formatted string.
func ExtractRPCStr(filename, content, rpcName string) (string, error) {
	if strings.TrimSpace(content) == "" {
		return "", fmt.Errorf("empty schema content")
	}

	schema, err := parser.ParserInstance.ParseString(filename, content)
	if err != nil {
		return "", fmt.Errorf("error parsing VDL: %w", err)
	}

	rpcDecl, err := ExtractRPC(schema, rpcName)
	if err != nil {
		return "", err
	}

	// Create a minimal schema with just the extracted RPC
	extractedSchema := &ast.Schema{
		Children: []*ast.SchemaChild{
			{RPC: rpcDecl},
		},
	}

	return formatter.FormatSchema(extractedSchema), nil
}

// ExtractConstStr extracts a specific const declaration from the VDL schema by name.
// It takes a VDL schema as a string and returns only the extracted const as a formatted string.
func ExtractConstStr(filename, content, constName string) (string, error) {
	if strings.TrimSpace(content) == "" {
		return "", fmt.Errorf("empty schema content")
	}

	schema, err := parser.ParserInstance.ParseString(filename, content)
	if err != nil {
		return "", fmt.Errorf("error parsing VDL: %w", err)
	}

	constDecl, err := ExtractConst(schema, constName)
	if err != nil {
		return "", err
	}

	// Create a minimal schema with just the extracted const
	extractedSchema := &ast.Schema{
		Children: []*ast.SchemaChild{
			{Const: constDecl},
		},
	}

	return formatter.FormatSchema(extractedSchema), nil
}

// ExtractEnumStr extracts a specific enum declaration from the VDL schema by name.
// It takes a VDL schema as a string and returns only the extracted enum as a formatted string.
func ExtractEnumStr(filename, content, enumName string) (string, error) {
	if strings.TrimSpace(content) == "" {
		return "", fmt.Errorf("empty schema content")
	}

	schema, err := parser.ParserInstance.ParseString(filename, content)
	if err != nil {
		return "", fmt.Errorf("error parsing VDL: %w", err)
	}

	enumDecl, err := ExtractEnum(schema, enumName)
	if err != nil {
		return "", err
	}

	// Create a minimal schema with just the extracted enum
	extractedSchema := &ast.Schema{
		Children: []*ast.SchemaChild{
			{Enum: enumDecl},
		},
	}

	return formatter.FormatSchema(extractedSchema), nil
}

// ExtractPatternStr extracts a specific pattern declaration from the VDL schema by name.
// It takes a VDL schema as a string and returns only the extracted pattern as a formatted string.
func ExtractPatternStr(filename, content, patternName string) (string, error) {
	if strings.TrimSpace(content) == "" {
		return "", fmt.Errorf("empty schema content")
	}

	schema, err := parser.ParserInstance.ParseString(filename, content)
	if err != nil {
		return "", fmt.Errorf("error parsing VDL: %w", err)
	}

	patternDecl, err := ExtractPattern(schema, patternName)
	if err != nil {
		return "", err
	}

	// Create a minimal schema with just the extracted pattern
	extractedSchema := &ast.Schema{
		Children: []*ast.SchemaChild{
			{Pattern: patternDecl},
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

// ExtractProc extracts a specific proc declaration from the schema by RPC name and proc name.
// Returns only the proc declaration, without dependencies or additional nodes.
func ExtractProc(schema *ast.Schema, rpcName, procName string) (*ast.ProcDecl, error) {
	rpc, err := ExtractRPC(schema, rpcName)
	if err != nil {
		return nil, err
	}

	for _, proc := range rpc.GetProcs() {
		if proc.Name == procName {
			return proc, nil
		}
	}
	return nil, fmt.Errorf("proc '%s' not found in RPC '%s'", procName, rpcName)
}

// ExtractStream extracts a specific stream declaration from the schema by RPC name and stream name.
// Returns only the stream declaration, without dependencies or additional nodes.
func ExtractStream(schema *ast.Schema, rpcName, streamName string) (*ast.StreamDecl, error) {
	rpc, err := ExtractRPC(schema, rpcName)
	if err != nil {
		return nil, err
	}

	for _, stream := range rpc.GetStreams() {
		if stream.Name == streamName {
			return stream, nil
		}
	}
	return nil, fmt.Errorf("stream '%s' not found in RPC '%s'", streamName, rpcName)
}

// ExtractRPC extracts a specific RPC declaration from the schema by name.
// Returns only the RPC declaration, without dependencies or additional nodes.
func ExtractRPC(schema *ast.Schema, rpcName string) (*ast.RPCDecl, error) {
	for _, rpc := range schema.GetRPCs() {
		if rpc.Name == rpcName {
			return rpc, nil
		}
	}
	return nil, fmt.Errorf("RPC '%s' not found in schema", rpcName)
}

// ExtractConst extracts a specific const declaration from the schema by name.
// Returns only the const declaration, without dependencies or additional nodes.
func ExtractConst(schema *ast.Schema, constName string) (*ast.ConstDecl, error) {
	for _, constDecl := range schema.GetConsts() {
		if constDecl.Name == constName {
			return constDecl, nil
		}
	}
	return nil, fmt.Errorf("const '%s' not found in schema", constName)
}

// ExtractEnum extracts a specific enum declaration from the schema by name.
// Returns only the enum declaration, without dependencies or additional nodes.
func ExtractEnum(schema *ast.Schema, enumName string) (*ast.EnumDecl, error) {
	for _, enumDecl := range schema.GetEnums() {
		if enumDecl.Name == enumName {
			return enumDecl, nil
		}
	}
	return nil, fmt.Errorf("enum '%s' not found in schema", enumName)
}

// ExtractPattern extracts a specific pattern declaration from the schema by name.
// Returns only the pattern declaration, without dependencies or additional nodes.
func ExtractPattern(schema *ast.Schema, patternName string) (*ast.PatternDecl, error) {
	for _, patternDecl := range schema.GetPatterns() {
		if patternDecl.Name == patternName {
			return patternDecl, nil
		}
	}
	return nil, fmt.Errorf("pattern '%s' not found in schema", patternName)
}
