package transpile

import (
	"embed"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/uforg/uforpc/urpc/internal/schema"
	"github.com/uforg/uforpc/urpc/internal/urpc/parser"
	"github.com/uforg/uforpc/urpc/internal/util/testutil"
)

// Run all transpiler tests
func TestTranspile(t *testing.T) {
	testFiles := collectTestFiles("")

	for _, file := range testFiles {
		assertTranspile(t, file)
	}
}

// This test is used to debug the transpiler using a single file.
// Feel free to modify the prefix to test other files.
func TestTranspileOnlyOne(t *testing.T) {
	testFiles := collectTestFiles("001")

	for _, file := range testFiles {
		assertTranspile(t, file)
	}
}

//////////////////
// TEST HELPERS //
//////////////////

//go:embed tests/*
var urpcTestFiles embed.FS

// testFile represents a URPC schema and its expected JSON representation.
type testFile struct {
	name string
	urpc string
	json string
}

// collectTestFiles collects all test files from the embedded filesystem.
func collectTestFiles(withPrefix string) []testFile {
	files, err := urpcTestFiles.ReadDir("tests")
	if err != nil {
		panic(err)
	}

	testFiles := []testFile{}
	for _, file := range files {
		if !strings.HasPrefix(file.Name(), withPrefix) {
			continue
		}

		if strings.HasSuffix(file.Name(), ".urpc") {
			content, err := urpcTestFiles.ReadFile("tests/" + file.Name())
			if err != nil {
				panic(err)
			}

			testFiles = append(testFiles, testFile{
				name: strings.TrimSuffix(file.Name(), ".urpc"),
				urpc: string(content),
			})
		}
	}

	// Populate the JSON content for each test file
	for i, file := range testFiles {
		jsonContent, err := urpcTestFiles.ReadFile("tests/" + file.name + ".json")
		if err != nil {
			panic(err)
		}
		testFiles[i].json = string(jsonContent)
	}

	return testFiles
}

// assertTranspile asserts that the transpilation of a URPC Schema AST <> JSON
// is correct in back and forth.
func assertTranspile(t *testing.T, file testFile) {
	t.Helper()

	astSchema, err := parser.ParserInstance.ParseString("", file.urpc)
	require.NoError(t, err, file.name+": error parsing URPC schema")

	jsonSchema, err := schema.ParseSchema(file.json)
	require.NoError(t, err, file.name+": error parsing JSON schema")

	// 1. Test AST > JSON > AST
	ast2json, err := ToJSON(*astSchema)
	require.NoError(t, err, file.name+": AJA -> error transpiling AST to JSON")
	require.Equal(t, jsonSchema, ast2json, file.name+": AJA -> incorrect JSON schema")

	ast2json2ast, err := ToURPC(ast2json)
	require.NoError(t, err, file.name+": AJAJ -> error transpiling JSON to AST")
	testutil.ASTEqualNoPos(t, astSchema, &ast2json2ast, file.name+": AJAJ -> incorrect AST schema")

	// 2. Test JSON > AST > JSON
	json2ast, err := ToURPC(jsonSchema)
	require.NoError(t, err, file.name+": JAJ -> error transpiling JSON to AST")
	testutil.ASTEqualNoPos(t, astSchema, &json2ast, file.name+": JAJ -> incorrect AST schema")

	json2ast2json, err := ToJSON(json2ast)
	require.NoError(t, err, file.name+": JAJA -> error transpiling AST to JSON")
	require.Equal(t, jsonSchema, json2ast2json, file.name+": JAJA -> incorrect JSON schema")
}
