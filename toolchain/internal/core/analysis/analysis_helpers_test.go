package analysis_test

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/varavelio/vdl/urpc/internal/core/analysis"
	"github.com/varavelio/vdl/urpc/internal/core/parser"
	"github.com/varavelio/vdl/urpc/internal/core/vfs"
)

// testdataDir is the path to the testdata directory relative to this test file.
const testdataDir = "testdata"

// expectPattern matches lines like: // @expect: E201
var expectPattern = regexp.MustCompile(`^//\s*@expect:\s*([A-Z]\d{3})`)

// expectedCodes represents the expected diagnostic codes for a test file.
type expectedCodes struct {
	codes    []string // Expected error codes
	noErrors bool     // If true, expect zero diagnostics
}

// parseExpectedCodes reads the first lines of a .ufo file to extract @expect comments.
// Files in the "valid" directory are expected to have no errors.
// Files in other directories must have at least one @expect comment.
func parseExpectedCodes(t *testing.T, filePath string) expectedCodes {
	t.Helper()

	// Files in valid/ directory should have no errors
	if strings.Contains(filePath, "/valid/") || strings.Contains(filePath, "\\valid\\") {
		return expectedCodes{noErrors: true}
	}

	file, err := os.Open(filePath)
	require.NoError(t, err, "failed to open test file: %s", filePath)
	defer file.Close()

	var codes []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Stop at first non-comment, non-empty line
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if !strings.HasPrefix(trimmed, "//") {
			break
		}

		matches := expectPattern.FindStringSubmatch(line)
		if len(matches) == 2 {
			codes = append(codes, matches[1])
		}
	}
	require.NoError(t, scanner.Err())

	if len(codes) == 0 {
		t.Fatalf("test file %s has no @expect comments; add // @expect: CODE at the top", filePath)
	}

	return expectedCodes{codes: codes}
}

// analyzeTestFile parses and analyzes a single .ufo file.
func analyzeTestFile(t *testing.T, filePath string) (*analysis.Program, []analysis.Diagnostic) {
	t.Helper()

	content, err := os.ReadFile(filePath)
	require.NoError(t, err, "failed to read test file: %s", filePath)

	schema, err := parser.ParserInstance.ParseString(filePath, string(content))
	require.NoError(t, err, "failed to parse test file: %s", filePath)

	return analysis.AnalyzeSchema(schema, filePath)
}

// analyzeMultiFileTest sets up a VFS with multiple files and analyzes from an entry point.
// It loads all .ufo and .md files to support external docstring resolution.
func analyzeMultiFileTest(t *testing.T, dir string, entryFile string) (*analysis.Program, []analysis.Diagnostic) {
	t.Helper()

	fs := vfs.New()

	// Walk the directory and add all .ufo and .md files to VFS
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// Include .ufo and .md files (for external docstrings)
		if !strings.HasSuffix(path, ".ufo") && !strings.HasSuffix(path, ".md") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Create virtual path preserving relative structure from dir
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		virtualPath := "/" + filepath.ToSlash(relPath)
		fs.WriteFileCache(virtualPath, content)
		return nil
	})
	require.NoError(t, err, "failed to walk directory: %s", dir)

	return analysis.Analyze(fs, "/"+entryFile)
}

// extractCodes extracts all diagnostic codes from a slice of diagnostics.
func extractCodes(diags []analysis.Diagnostic) []string {
	codes := make([]string, len(diags))
	for i, d := range diags {
		codes[i] = d.Code
	}
	return codes
}

// containsCode checks if a code is present in a slice of codes.
func containsCode(codes []string, target string) bool {
	for _, c := range codes {
		if c == target {
			return true
		}
	}
	return false
}

// globTestFiles finds all .ufo files matching a pattern within testdata.
func globTestFiles(t *testing.T, pattern string) []string {
	t.Helper()

	fullPattern := filepath.Join(testdataDir, pattern)
	matches, err := filepath.Glob(fullPattern)
	require.NoError(t, err, "failed to glob pattern: %s", fullPattern)

	return matches
}

// globTestFilesRecursive finds all .ufo files in a directory recursively.
func globTestFilesRecursive(t *testing.T, dir string) []string {
	t.Helper()

	fullDir := filepath.Join(testdataDir, dir)
	var files []string

	err := filepath.Walk(fullDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".ufo") {
			files = append(files, path)
		}
		return nil
	})
	require.NoError(t, err, "failed to walk directory: %s", fullDir)

	return files
}

// relativePath returns the path relative to testdata for cleaner test names.
func relativePath(t *testing.T, path string) string {
	t.Helper()

	rel, err := filepath.Rel(testdataDir, path)
	if err != nil {
		return path
	}
	return rel
}
