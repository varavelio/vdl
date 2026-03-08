package analysis

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/varavelio/vdl/toolchain/internal/core/ast"
)

const configSchemaFileName = "vdl.config.vdl"

func (v *validator) validateFileNames() []Diagnostic {
	var diagnostics []Diagnostic

	for _, file := range v.files {
		if v.ctx.Err() != nil {
			return diagnostics
		}
		if file == nil {
			continue
		}

		if diag := validateSelfFileName(file.Path); diag != nil {
			diagnostics = append(diagnostics, *diag)
		}

		if file.AST == nil {
			continue
		}

		includes := file.AST.GetIncludes()
		for idx, include := range includes {
			if v.ctx.Err() != nil {
				return diagnostics
			}

			includeTarget := string(include.Path)
			if idx < len(file.Includes) && file.Includes[idx] != "" {
				includeTarget = file.Includes[idx]
			}

			if diag := validateIncludeFileName(file.Path, include, includeTarget); diag != nil {
				diagnostics = append(diagnostics, *diag)
			}
		}
	}

	return diagnostics
}

func validateSelfFileName(filePath string) *Diagnostic {
	base := filepath.Base(filePath)
	if base == configSchemaFileName {
		return nil
	}
	if isRegularSchemaFileName(base) {
		return nil
	}

	pos := ast.Position{Filename: filePath, Line: 1, Column: 1}
	diag := newDiagnostic(
		filePath,
		pos,
		pos,
		CodeInvalidFileName,
		fmt.Sprintf("file name %q is invalid: expected [a-z0-9_]+.vdl (or %q for self file validation)", base, configSchemaFileName),
	)
	return &diag
}

func validateIncludeFileName(currentFile string, include *ast.Include, includeTarget string) *Diagnostic {
	base := filepath.Base(includeTarget)

	if base == configSchemaFileName {
		diag := newDiagnostic(
			currentFile,
			include.Pos,
			include.EndPos,
			CodeInvalidIncludeFile,
			fmt.Sprintf("include %q is invalid: %q cannot be imported", string(include.Path), configSchemaFileName),
		)
		return &diag
	}

	if isRegularSchemaFileName(base) {
		return nil
	}

	diag := newDiagnostic(
		currentFile,
		include.Pos,
		include.EndPos,
		CodeInvalidIncludeFile,
		fmt.Sprintf("include %q is invalid: target file name %q must match [a-z0-9_]+.vdl", string(include.Path), base),
	)
	return &diag
}

func isRegularSchemaFileName(fileName string) bool {
	if !strings.HasSuffix(fileName, ".vdl") {
		return false
	}

	name := strings.TrimSuffix(fileName, ".vdl")
	if name == "" {
		return false
	}

	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			continue
		}
		return false
	}

	return true
}
