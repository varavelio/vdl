// Package analyzer provides semantic analysis for URPC schemas.
// This implementation prioritizes simplicity and maintainability over performance
// by performing a full analysis without caching results between calls.
package analyzer

import (
	"fmt"

	"github.com/uforg/uforpc/urpc/internal/urpc/ast"
	"github.com/uforg/uforpc/urpc/internal/urpc/parser"
)

// Analyzer manages the analysis process for URPC schemas without caching.
type Analyzer struct {
	fileProvider      FileProvider
	docstringResolver *docstringResolver
}

// NewAnalyzer creates a new cache-less Analyzer instance.
func NewAnalyzer(fileProvider FileProvider) (*Analyzer, error) {
	return &Analyzer{
		fileProvider:      fileProvider,
		docstringResolver: newDocstringResolver(fileProvider),
	}, nil
}

// Analyze performs semantic analysis on a URPC schema starting from the given entry point.
// It parses the entry point file, resolves all docstrings and then performs the semantic analysis.
//
// It consists of two phases:
//   - Resolution phase: Parses the entry point file and resolves all external docstrings.
//   - Semantic analysis phase: Performs semantic analysis on the resolved schema.
func (a *Analyzer) Analyze(entryPointFilePath string) (*ast.Schema, []Diagnostic, error) {
	fileContent, _, err := a.fileProvider.GetFileAndHash("", entryPointFilePath)
	if err != nil {
		diag := Diagnostic{
			Positions: Positions{
				Pos:    ast.Position{Filename: entryPointFilePath, Line: 1, Column: 1, Offset: 0},
				EndPos: ast.Position{Filename: entryPointFilePath, Line: 1, Column: 1, Offset: 0},
			},
			Message: fmt.Sprintf("failed to read entry point file: %s", err.Error()),
		}
		return nil, []Diagnostic{diag}, err
	}

	astSchema, err := parser.ParserInstance.ParseString(entryPointFilePath, fileContent)
	if err != nil {
		diag := Diagnostic{
			Positions: Positions{
				Pos:    ast.Position{Filename: entryPointFilePath, Line: 1, Column: 1, Offset: 0},
				EndPos: ast.Position{Filename: entryPointFilePath, Line: 1, Column: 1, Offset: 0},
			},
			Message: fmt.Sprintf("error parsing file: %v", err),
		}

		// Assert parser error if possible
		if parserErr, ok := err.(parser.Error); ok {
			diag = Diagnostic{
				Positions: Positions{
					Pos:    parserErr.Position(),
					EndPos: parserErr.Position(),
				},
				Message: parserErr.Message(),
			}
		}

		return nil, []Diagnostic{diag}, err
	}

	astSchema, dsResolverDiagnostics, _ := a.docstringResolver.resolve(astSchema)
	if len(dsResolverDiagnostics) > 0 {
		return astSchema, dsResolverDiagnostics, dsResolverDiagnostics[0]
	}

	semanalyzer := newSemanalyzer(astSchema)
	semanalyzerDiagnostics, _ := semanalyzer.analyze()
	if len(semanalyzerDiagnostics) > 0 {
		return astSchema, semanalyzerDiagnostics, semanalyzerDiagnostics[0]
	}

	return astSchema, nil, nil
}

func (a *Analyzer) AnalyzeAstSchema(astSchema *ast.Schema) ([]Diagnostic, error) {
	semanalyzer := newSemanalyzer(astSchema)

	semanalyzerDiagnostics, _ := semanalyzer.analyze()
	if len(semanalyzerDiagnostics) > 0 {
		return semanalyzerDiagnostics, semanalyzerDiagnostics[0]
	}

	return nil, nil
}
