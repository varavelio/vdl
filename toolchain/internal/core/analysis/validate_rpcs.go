package analysis

import (
	"fmt"

	"github.com/varavelio/vdl/toolchain/internal/core/ast"
)

// validateRPCs validates RPC declarations:
// - Procedure names must be unique within an RPC
// - Stream names must be unique within an RPC
// - Procedure and stream cannot share the same name
func validateRPCs(symbols *symbolTable) []Diagnostic {
	var diagnostics []Diagnostic

	for _, rpc := range symbols.vdls {
		diagnostics = append(diagnostics, validateRPC(rpc)...)
	}

	return diagnostics
}

// validateRPC validates a single RPC declaration.
func validateRPC(rpc *RPCSymbol) []Diagnostic {
	var diagnostics []Diagnostic

	// Track all endpoint names (procs + streams)
	endpointNames := make(map[string]endpointInfo)

	// Check procedure uniqueness
	for name, proc := range rpc.Procs {
		if existing, ok := endpointNames[name]; ok {
			var code string
			var msg string
			if existing.kind == "procedure" {
				code = CodeDuplicateProc
				msg = fmt.Sprintf("procedure %q is already declared in RPC %q at %s:%d:%d",
					name, rpc.Name, existing.file, existing.pos.Line, existing.pos.Column)
			} else {
				code = CodeProcStreamSameName
				msg = fmt.Sprintf("procedure %q conflicts with stream of the same name in RPC %q at %s:%d:%d",
					name, rpc.Name, existing.file, existing.pos.Line, existing.pos.Column)
			}
			diagnostics = append(diagnostics, newDiagnostic(
				proc.File,
				proc.Pos,
				proc.EndPos,
				code,
				msg,
			))
		} else {
			endpointNames[name] = endpointInfo{
				kind: "procedure",
				file: proc.File,
				pos:  proc.Pos,
			}
		}
	}

	// Check stream uniqueness
	for name, stream := range rpc.Streams {
		if existing, ok := endpointNames[name]; ok {
			var code string
			var msg string
			if existing.kind == "stream" {
				code = CodeDuplicateStream
				msg = fmt.Sprintf("stream %q is already declared in RPC %q at %s:%d:%d",
					name, rpc.Name, existing.file, existing.pos.Line, existing.pos.Column)
			} else {
				code = CodeProcStreamSameName
				msg = fmt.Sprintf("stream %q conflicts with procedure of the same name in RPC %q at %s:%d:%d",
					name, rpc.Name, existing.file, existing.pos.Line, existing.pos.Column)
			}
			diagnostics = append(diagnostics, newDiagnostic(
				stream.File,
				stream.Pos,
				stream.EndPos,
				code,
				msg,
			))
		} else {
			endpointNames[name] = endpointInfo{
				kind: "stream",
				file: stream.File,
				pos:  stream.Pos,
			}
		}
	}

	return diagnostics
}

// endpointInfo tracks information about a procedure or stream for duplicate detection.
type endpointInfo struct {
	kind string // "procedure" or "stream"
	file string
	pos  ast.Position
}

// buildRPCSymbol creates an RPCSymbol from an AST RPCDecl.
// It also returns diagnostics for duplicate procs/streams within the same RPC.
func buildRPCSymbol(decl *ast.RPCDecl, file string) (*RPCSymbol, []Diagnostic) {
	var diagnostics []Diagnostic

	var docstring *string
	if decl.Docstring != nil {
		s := string(decl.Docstring.Value)
		docstring = &s
	}

	var deprecated *DeprecationInfo
	if decl.Deprecated != nil {
		msg := ""
		if decl.Deprecated.Message != nil {
			msg = string(*decl.Deprecated.Message)
		}
		deprecated = &DeprecationInfo{Message: msg}
	}

	rpc := &RPCSymbol{
		Symbol: Symbol{
			Name:       decl.Name,
			File:       file,
			Pos:        decl.Pos,
			EndPos:     decl.EndPos,
			Docstring:  docstring,
			Deprecated: deprecated,
		},
		Procs:          make(map[string]*ProcSymbol),
		Streams:        make(map[string]*StreamSymbol),
		DeclaredIn:     []string{file},
		StandaloneDocs: []*DocSymbol{},
	}

	// Track endpoint names for duplicate detection
	endpointNames := make(map[string]endpointInfo)

	// Process children
	for _, child := range decl.Children {
		if child.Proc != nil {
			proc := buildProcSymbol(child.Proc, file)
			if existing, ok := endpointNames[proc.Name]; ok {
				var code string
				var msg string
				if existing.kind == "procedure" {
					code = CodeDuplicateProc
					msg = fmt.Sprintf("procedure %q is already declared in RPC %q at %s:%d:%d",
						proc.Name, rpc.Name, existing.file, existing.pos.Line, existing.pos.Column)
				} else {
					code = CodeProcStreamSameName
					msg = fmt.Sprintf("procedure %q conflicts with stream of the same name in RPC %q at %s:%d:%d",
						proc.Name, rpc.Name, existing.file, existing.pos.Line, existing.pos.Column)
				}
				diagnostics = append(diagnostics, newDiagnostic(
					proc.File,
					proc.Pos,
					proc.EndPos,
					code,
					msg,
				))
			} else {
				endpointNames[proc.Name] = endpointInfo{
					kind: "procedure",
					file: proc.File,
					pos:  proc.Pos,
				}
			}
			rpc.Procs[proc.Name] = proc
		}
		if child.Stream != nil {
			stream := buildStreamSymbol(child.Stream, file)
			if existing, ok := endpointNames[stream.Name]; ok {
				var code string
				var msg string
				if existing.kind == "stream" {
					code = CodeDuplicateStream
					msg = fmt.Sprintf("stream %q is already declared in RPC %q at %s:%d:%d",
						stream.Name, rpc.Name, existing.file, existing.pos.Line, existing.pos.Column)
				} else {
					code = CodeProcStreamSameName
					msg = fmt.Sprintf("stream %q conflicts with procedure of the same name in RPC %q at %s:%d:%d",
						stream.Name, rpc.Name, existing.file, existing.pos.Line, existing.pos.Column)
				}
				diagnostics = append(diagnostics, newDiagnostic(
					stream.File,
					stream.Pos,
					stream.EndPos,
					code,
					msg,
				))
			} else {
				endpointNames[stream.Name] = endpointInfo{
					kind: "stream",
					file: stream.File,
					pos:  stream.Pos,
				}
			}
			rpc.Streams[stream.Name] = stream
		}
		if child.Docstring != nil {
			rpc.StandaloneDocs = append(rpc.StandaloneDocs, &DocSymbol{
				Content: string(child.Docstring.Value),
				Pos:     child.Docstring.Pos,
				EndPos:  child.Docstring.EndPos,
				File:    file,
			})
		}
	}

	return rpc, diagnostics
}

// buildProcSymbol creates a ProcSymbol from an AST ProcDecl.
func buildProcSymbol(decl *ast.ProcDecl, file string) *ProcSymbol {
	var docstring *string
	if decl.Docstring != nil {
		s := string(decl.Docstring.Value)
		docstring = &s
	}

	var deprecated *DeprecationInfo
	if decl.Deprecated != nil {
		msg := ""
		if decl.Deprecated.Message != nil {
			msg = string(*decl.Deprecated.Message)
		}
		deprecated = &DeprecationInfo{Message: msg}
	}

	proc := &ProcSymbol{
		Symbol: Symbol{
			Name:       decl.Name,
			File:       file,
			Pos:        decl.Pos,
			EndPos:     decl.EndPos,
			Docstring:  docstring,
			Deprecated: deprecated,
		},
		AST: decl,
	}

	// Process input/output blocks
	for _, child := range decl.Children {
		if child.Input != nil {
			proc.Input = buildInputBlockSymbol(child.Input, file)
		}
		if child.Output != nil {
			proc.Output = buildOutputBlockSymbol(child.Output, file)
		}
	}

	return proc
}

// buildStreamSymbol creates a StreamSymbol from an AST StreamDecl.
func buildStreamSymbol(decl *ast.StreamDecl, file string) *StreamSymbol {
	var docstring *string
	if decl.Docstring != nil {
		s := string(decl.Docstring.Value)
		docstring = &s
	}

	var deprecated *DeprecationInfo
	if decl.Deprecated != nil {
		msg := ""
		if decl.Deprecated.Message != nil {
			msg = string(*decl.Deprecated.Message)
		}
		deprecated = &DeprecationInfo{Message: msg}
	}

	stream := &StreamSymbol{
		Symbol: Symbol{
			Name:       decl.Name,
			File:       file,
			Pos:        decl.Pos,
			EndPos:     decl.EndPos,
			Docstring:  docstring,
			Deprecated: deprecated,
		},
		AST: decl,
	}

	// Process input/output blocks
	for _, child := range decl.Children {
		if child.Input != nil {
			stream.Input = buildInputBlockSymbol(child.Input, file)
		}
		if child.Output != nil {
			stream.Output = buildOutputBlockSymbol(child.Output, file)
		}
	}

	return stream
}

// buildInputBlockSymbol creates a BlockSymbol from an AST input/output block.
func buildInputBlockSymbol(input *ast.ProcOrStreamDeclChildInput, file string) *BlockSymbol {
	block := &BlockSymbol{
		Pos:     input.Pos,
		EndPos:  input.EndPos,
		Fields:  []*FieldSymbol{},
		Spreads: []*SpreadRef{},
	}

	for _, child := range input.Children {
		if child.Field != nil {
			block.Fields = append(block.Fields, buildFieldSymbol(child.Field, file))
		}
		if child.Spread != nil {
			block.Spreads = append(block.Spreads, &SpreadRef{
				TypeName: child.Spread.TypeName,
				Pos:      child.Spread.Pos,
				EndPos:   child.Spread.EndPos,
			})
		}
	}

	return block
}

// buildOutputBlockSymbol creates a BlockSymbol from an AST output block.
func buildOutputBlockSymbol(output *ast.ProcOrStreamDeclChildOutput, file string) *BlockSymbol {
	block := &BlockSymbol{
		Pos:     output.Pos,
		EndPos:  output.EndPos,
		Fields:  []*FieldSymbol{},
		Spreads: []*SpreadRef{},
	}

	for _, child := range output.Children {
		if child.Field != nil {
			block.Fields = append(block.Fields, buildFieldSymbol(child.Field, file))
		}
		if child.Spread != nil {
			block.Spreads = append(block.Spreads, &SpreadRef{
				TypeName: child.Spread.TypeName,
				Pos:      child.Spread.Pos,
				EndPos:   child.Spread.EndPos,
			})
		}
	}

	return block
}
