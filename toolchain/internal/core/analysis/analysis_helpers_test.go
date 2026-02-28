package analysis_test

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/varavelio/vdl/toolchain/internal/core/analysis"
	"github.com/varavelio/vdl/toolchain/internal/core/vfs"
)

const testdataDir = "testdata"

const (
	directiveExpect       = "// @expect:"
	directiveExpectNoErr  = "// @expect-no-errors"
	directiveAssert       = "// @assert:"
	dirErrors             = "/errors/"
	assertHasType         = "has-type"
	assertHasEnum         = "has-enum"
	assertHasConst        = "has-const"
	assertFileCount       = "file-count"
	assertStandaloneDocs  = "standalone-docs"
	assertTypeSpread      = "type-spread"
	assertResolvedType    = "field-resolved-type"
	assertResolvedEnum    = "field-resolved-enum"
	assertTypeDocContains = "type-doc-contains"
	assertFieldDocContain = "field-doc-contains"
	assertStdDocContains  = "standalone-doc-contains"
	assertConstKind       = "const-kind"
	assertDiagContains    = "diagnostic-message-contains"
)

type goldenSpec struct {
	ExpectCodes   []string
	ExpectNoError bool
	Assertions    []goldenAssertion
}

type goldenAssertion struct {
	Key  string
	Args []string
	Raw  string
}

type goldenCase struct {
	Name      string
	RootDir   string
	EntryFile string
	Spec      goldenSpec
}

func discoverGoldenCases(t *testing.T) []goldenCase {
	t.Helper()

	allVDL := collectVDLFiles(t)
	mainFiles := make([]string, 0)
	mainDirs := make(map[string]bool)

	for _, file := range allVDL {
		if filepath.Base(file) == "main.vdl" {
			mainFiles = append(mainFiles, file)
			mainDirs[filepath.Dir(file)] = true
		}
	}
	sort.Strings(mainFiles)

	cases := make([]goldenCase, 0, len(allVDL))

	for _, main := range mainFiles {
		root := filepath.Dir(main)
		cases = append(cases, goldenCase{
			Name:      toTestdataRelative(root),
			RootDir:   root,
			EntryFile: "main.vdl",
			Spec:      parseGoldenSpec(t, main),
		})
	}

	for _, file := range allVDL {
		if filepath.Base(file) == "main.vdl" {
			continue
		}
		if insideAnyMainDir(file, mainDirs) {
			continue
		}
		cases = append(cases, goldenCase{
			Name:      toTestdataRelative(file),
			RootDir:   filepath.Dir(file),
			EntryFile: filepath.Base(file),
			Spec:      parseGoldenSpec(t, file),
		})
	}

	sort.Slice(cases, func(i, j int) bool { return cases[i].Name < cases[j].Name })
	return cases
}

func collectVDLFiles(t *testing.T) []string {
	t.Helper()

	var files []string
	err := filepath.Walk(testdataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".vdl") {
			files = append(files, path)
		}
		return nil
	})
	require.NoError(t, err)
	sort.Strings(files)
	return files
}

func insideAnyMainDir(path string, mainDirs map[string]bool) bool {
	for dir := range mainDirs {
		if isInsideDir(path, dir) {
			return true
		}
	}
	return false
}

func isInsideDir(path, dir string) bool {
	rel, err := filepath.Rel(dir, path)
	if err != nil {
		return false
	}
	if rel == "." {
		return true
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func parseGoldenSpec(t *testing.T, entryFile string) goldenSpec {
	t.Helper()

	f, err := os.Open(entryFile)
	require.NoError(t, err)
	defer f.Close()

	var spec goldenSpec
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "//") {
			break
		}

		switch {
		case strings.HasPrefix(line, directiveExpect):
			code := strings.TrimSpace(strings.TrimPrefix(line, directiveExpect))
			require.Truef(t, len(code) == 4 && strings.HasPrefix(code, "E"), "invalid expect code %q in %s", code, entryFile)
			spec.ExpectCodes = append(spec.ExpectCodes, code)
		case strings.HasPrefix(line, directiveExpectNoErr):
			spec.ExpectNoError = true
		case strings.HasPrefix(line, directiveAssert):
			raw := strings.TrimSpace(strings.TrimPrefix(line, directiveAssert))
			parts := strings.Fields(raw)
			require.NotEmptyf(t, parts, "invalid assert directive in %s: %q", entryFile, line)
			spec.Assertions = append(spec.Assertions, goldenAssertion{Key: parts[0], Args: parts[1:], Raw: raw})
		}
	}
	require.NoError(t, scanner.Err())

	if spec.ExpectNoError && len(spec.ExpectCodes) > 0 {
		t.Fatalf("file %s declares both @expect and @expect-no-errors", entryFile)
	}

	if !spec.ExpectNoError && len(spec.ExpectCodes) == 0 {
		if strings.Contains(filepath.ToSlash(entryFile), dirErrors) {
			t.Fatalf("error test file %s has no @expect directive", entryFile)
		}
		spec.ExpectNoError = true
	}

	return spec
}

func runGoldenCase(t *testing.T, tc goldenCase) (*analysis.Program, []analysis.Diagnostic) {
	t.Helper()

	fs := vfs.New()
	err := filepath.Walk(tc.RootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".vdl") && !strings.HasSuffix(path, ".md") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(tc.RootDir, path)
		if err != nil {
			return err
		}
		fs.WriteFileCache("/"+filepath.ToSlash(rel), content)
		return nil
	})
	require.NoError(t, err)

	return analysis.Analyze(fs, "/"+filepath.ToSlash(tc.EntryFile))
}

func assertGoldenResult(t *testing.T, tc goldenCase, program *analysis.Program, diagnostics []analysis.Diagnostic) {
	t.Helper()
	require.NotNil(t, program)

	if tc.Spec.ExpectNoError {
		require.Emptyf(t, diagnostics, "expected no diagnostics, got %s", formatDiagnosticCodes(diagnostics))
	} else {
		codes := extractCodes(diagnostics)
		for _, expected := range tc.Spec.ExpectCodes {
			require.Containsf(t, codes, expected, "expected code %s not found in %v", expected, codes)
		}
	}

	for _, a := range tc.Spec.Assertions {
		applyAssertion(t, a, program, diagnostics)
	}
}

func applyAssertion(t *testing.T, a goldenAssertion, program *analysis.Program, diagnostics []analysis.Diagnostic) {
	t.Helper()

	requireArgs := func(n int) {
		if len(a.Args) < n {
			t.Fatalf("assertion %q expects at least %d args, got %d", a.Raw, n, len(a.Args))
		}
	}

	switch a.Key {
	case assertHasType:
		requireArgs(1)
		_, ok := program.Types[a.Args[0]]
		require.Truef(t, ok, "type %q not found", a.Args[0])

	case assertHasEnum:
		requireArgs(1)
		_, ok := program.Enums[a.Args[0]]
		require.Truef(t, ok, "enum %q not found", a.Args[0])

	case assertHasConst:
		requireArgs(1)
		_, ok := program.Consts[a.Args[0]]
		require.Truef(t, ok, "const %q not found", a.Args[0])

	case assertFileCount:
		requireArgs(1)
		expected, err := strconv.Atoi(a.Args[0])
		require.NoError(t, err)
		require.Equal(t, expected, len(program.Files))

	case assertStandaloneDocs:
		requireArgs(1)
		expected, err := strconv.Atoi(a.Args[0])
		require.NoError(t, err)
		require.Equal(t, expected, len(program.StandaloneDocs))

	case assertTypeSpread:
		requireArgs(2)
		typ := getType(t, program, a.Args[0])
		for _, s := range typ.Spreads {
			if s.Name == a.Args[1] {
				return
			}
		}
		t.Fatalf("spread %q not found in type %q", a.Args[1], a.Args[0])

	case assertResolvedType:
		requireArgs(3)
		f := getTypeField(t, program, a.Args[0], a.Args[1])
		require.NotNil(t, f.Type)
		require.NotNil(t, f.Type.ResolvedType)
		require.Equal(t, a.Args[2], f.Type.ResolvedType.Name)

	case assertResolvedEnum:
		requireArgs(3)
		f := getTypeField(t, program, a.Args[0], a.Args[1])
		require.NotNil(t, f.Type)
		require.NotNil(t, f.Type.ResolvedEnum)
		require.Equal(t, a.Args[2], f.Type.ResolvedEnum.Name)

	case assertTypeDocContains:
		requireArgs(2)
		typ := getType(t, program, a.Args[0])
		require.NotNil(t, typ.Docstring)
		require.Contains(t, strings.ToLower(*typ.Docstring), strings.ToLower(strings.Join(a.Args[1:], " ")))

	case assertFieldDocContain:
		requireArgs(3)
		f := getTypeField(t, program, a.Args[0], a.Args[1])
		require.NotNil(t, f.Docstring)
		require.Contains(t, strings.ToLower(*f.Docstring), strings.ToLower(strings.Join(a.Args[2:], " ")))

	case assertStdDocContains:
		requireArgs(1)
		needle := strings.ToLower(strings.Join(a.Args, " "))
		for _, ds := range program.StandaloneDocs {
			if strings.Contains(strings.ToLower(ds.Content), needle) {
				return
			}
		}
		t.Fatalf("no standalone doc contains %q", strings.Join(a.Args, " "))

	case assertConstKind:
		requireArgs(2)
		c := getConst(t, program, a.Args[0])
		expected, ok := parseConstKind(a.Args[1])
		require.Truef(t, ok, "invalid const-kind %q", a.Args[1])
		require.Equal(t, expected, c.ValueType)

	case assertDiagContains:
		requireArgs(2)
		code := a.Args[0]
		needle := strings.ToLower(strings.Join(a.Args[1:], " "))
		for _, d := range diagnostics {
			if d.Code == code && strings.Contains(strings.ToLower(d.Message), needle) {
				return
			}
		}
		t.Fatalf("no diagnostic with code=%s containing %q", code, strings.Join(a.Args[1:], " "))

	default:
		t.Fatalf("unknown assertion key %q in %q", a.Key, a.Raw)
	}
}

func getType(t *testing.T, program *analysis.Program, name string) *analysis.TypeSymbol {
	t.Helper()
	typ, ok := program.Types[name]
	require.Truef(t, ok, "type %q not found", name)
	return typ
}

func getConst(t *testing.T, program *analysis.Program, name string) *analysis.ConstSymbol {
	t.Helper()
	cnst, ok := program.Consts[name]
	require.Truef(t, ok, "const %q not found", name)
	return cnst
}

func getTypeField(t *testing.T, program *analysis.Program, typeName, fieldName string) *analysis.FieldSymbol {
	t.Helper()
	typ := getType(t, program, typeName)
	for _, f := range typ.Fields {
		if f.Name == fieldName {
			return f
		}
	}
	t.Fatalf("field %q not found in type %q", fieldName, typeName)
	return nil
}

func parseConstKind(v string) (analysis.ConstValueType, bool) {
	switch strings.ToLower(v) {
	case "string":
		return analysis.ConstValueTypeString, true
	case "int":
		return analysis.ConstValueTypeInt, true
	case "float":
		return analysis.ConstValueTypeFloat, true
	case "bool":
		return analysis.ConstValueTypeBool, true
	case "object":
		return analysis.ConstValueTypeObject, true
	case "array":
		return analysis.ConstValueTypeArray, true
	case "reference":
		return analysis.ConstValueTypeReference, true
	case "unknown":
		return analysis.ConstValueTypeUnknown, true
	default:
		return analysis.ConstValueTypeUnknown, false
	}
}

func extractCodes(diags []analysis.Diagnostic) []string {
	codes := make([]string, len(diags))
	for i, d := range diags {
		codes[i] = d.Code
	}
	return codes
}

func formatDiagnosticCodes(diags []analysis.Diagnostic) string {
	parts := make([]string, 0, len(diags))
	for _, d := range diags {
		parts = append(parts, fmt.Sprintf("%s(%s)", d.Code, d.Message))
	}
	return strings.Join(parts, ", ")
}

func toTestdataRelative(path string) string {
	rel, err := filepath.Rel(testdataDir, path)
	if err != nil {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(rel)
}
