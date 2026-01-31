// Package analysis provides semantic analysis for VDL schemas.
//
// The analysis package is the "Semantic Brain" of the compiler. It receives an
// entry point file, recursively resolves all imports, validates complete semantics
// according to the VDL specification, and produces a unified Program with all
// symbols merged into a global namespace.
//
// # Usage
//
//	fs := vfs.New()
//	program, diagnostics := analysis.Analyze(fs, "main.vdl")
//	if len(diagnostics) > 0 {
//	    // Handle errors - but program is still usable for LSP features
//	    for _, d := range diagnostics {
//	        fmt.Println(d)
//	    }
//	}
//	// Use program (may be partial if there were errors)...
//
// # Analysis Pipeline
//
// The analysis process consists of four phases:
//
//  1. Resolution: Parse entry point, resolve includes recursively, resolve external docstrings
//  2. Symbol Collection: Collect all types, enums, constants, patterns, RPCs into a symbol table
//  3. Validation: Run all semantic validators (naming, types, spreads, enums, patterns, etc.)
//  4. Result: Return Program (best-effort) along with all diagnostics
//
// # Error Handling (Best-Effort Strategy)
//
// Analysis uses a best-effort strategy: it always attempts to return a Program
// that is as complete as possible, even when errors are encountered. This enables
// LSP features (go-to-definition, hover, completions) to remain functional even
// in files with errors.
//
// The returned Program will contain all symbols that could be successfully parsed
// and collected, regardless of validation errors. Check the diagnostics slice to
// determine if the schema is fully valid.
package analysis

import (
	"context"

	"github.com/varavelio/vdl/toolchain/internal/core/ast"
	"github.com/varavelio/vdl/toolchain/internal/core/vfs"
)

// Analyze performs complete semantic analysis starting from an entry point.
//
// It parses the entry point file, resolves all includes recursively (detecting
// circular dependencies), resolves external docstrings from markdown files,
// validates all semantic rules, and builds a unified Program with all symbols
// merged into a global namespace.
//
// This function uses a best-effort strategy: it always returns a Program that
// is as complete as possible, even when errors are found. This enables LSP
// features to work even in files with errors.
//
// Parameters:
//   - fs: Virtual filesystem for reading files (supports caching and dirty buffers)
//   - entryPoint: Path to the main .vdl file (can be relative or absolute)
//
// Returns:
//   - *Program: The program with all successfully collected symbols (never nil)
//   - []Diagnostic: All errors found during analysis (empty slice if successful)
//
// Example:
//
//	fs := vfs.New()
//	program, diagnostics := analysis.Analyze(fs, "schema/main.vdl")
//	if len(diagnostics) > 0 {
//	    for _, d := range diagnostics {
//	        log.Printf("%s", d)
//	    }
//	    // Program is still usable for LSP features
//	}
//	// Use program...
func Analyze(fs *vfs.FileSystem, entryPoint string) (*Program, []Diagnostic) {
	return AnalyzeWithContext(context.Background(), fs, entryPoint)
}

// AnalyzeWithContext performs complete semantic analysis with context support for cancellation.
//
// This variant allows the caller to cancel the analysis early if needed (e.g., when the
// user types again before analysis completes). The context is checked at key points
// during the analysis pipeline: before resolution, after resolution, and before validation.
//
// If the context is cancelled, this function returns nil, nil immediately.
//
// See Analyze for full documentation.
func AnalyzeWithContext(ctx context.Context, fs *vfs.FileSystem, entryPoint string) (*Program, []Diagnostic) {
	// Check for cancellation before starting
	if ctx.Err() != nil {
		return nil, nil
	}

	absPath := fs.Resolve("", entryPoint)
	var allDiags []Diagnostic

	// Phase 1: Resolution
	resolver := newResolverWithContext(ctx, fs)
	files, resolverDiags := resolver.resolve(entryPoint)
	allDiags = append(allDiags, resolverDiags...)

	// Check for cancellation after resolution
	if ctx.Err() != nil {
		return nil, nil
	}

	// Even if resolution had errors, we may have partial files to work with
	// If no files at all, return an empty program
	if len(files) == 0 {
		return newProgram(absPath), allDiags
	}

	// Phase 2: Symbol Collection
	validator := newValidatorWithContext(ctx, files)
	collectionDiags := validator.collect()
	allDiags = append(allDiags, collectionDiags...)

	// Check for cancellation after collection
	if ctx.Err() != nil {
		return nil, nil
	}

	// Phase 3: Validation
	validationDiags := validator.validate()
	allDiags = append(allDiags, validationDiags...)

	// Check for cancellation after validation
	if ctx.Err() != nil {
		return nil, nil
	}

	// Phase 4: Result - Always return the program (best-effort)
	return validator.buildProgram(absPath), allDiags
}

// AnalyzeSchema performs semantic analysis on a pre-parsed schema.
//
// This is useful for:
//   - Single-file scenarios where includes are not used
//   - Testing individual schemas
//   - Scenarios where includes have been pre-resolved
//
// Note: This function does NOT resolve includes. If the schema contains include
// statements, they will be ignored and any types from included files will not
// be available.
//
// This function uses a best-effort strategy: it always returns a Program that
// is as complete as possible, even when errors are found.
//
// Parameters:
//   - schema: Pre-parsed AST schema
//   - absoluteFilePath: The absolute file path to use for error reporting
//
// Returns:
//   - *Program: The program with all successfully collected symbols (never nil)
//   - []Diagnostic: All errors found during analysis (empty slice if successful)
func AnalyzeSchema(schema *ast.Schema, absoluteFilePath string) (*Program, []Diagnostic) {
	// Create a single-file map
	files := map[string]*File{
		absoluteFilePath: {
			Path:     absoluteFilePath,
			AST:      schema,
			Includes: []string{},
		},
	}

	// Phase 2: Symbol Collection
	validator := newValidator(files)
	collectionDiags := validator.collect()

	// Phase 3: Validation
	validationDiags := validator.validate()

	// Combine all diagnostics
	allDiags := append(collectionDiags, validationDiags...)

	// Phase 4: Result - Always return the program (best-effort)
	return validator.buildProgram(absoluteFilePath), allDiags
}
