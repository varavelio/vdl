package parser

import (
	"testing"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/stretchr/testify/require"
)

// tokenInfo is a helper struct for testing
type tokenInfo struct {
	Type  string
	Value string
}

// lexString is a helper function that lexes a string and returns tokens
func lexString(input string) ([]tokenInfo, error) {
	lex, err := VDLLexer.LexString("test.vdl", input)
	if err != nil {
		return nil, err
	}

	var tokens []tokenInfo
	for {
		tok, err := lex.Next()
		if err != nil {
			return nil, err
		}

		// Get token name from symbols
		tokenName := ""
		for name, typ := range VDLLexer.Symbols() {
			if typ == tok.Type {
				tokenName = name
				break
			}
		}
		if tok.Type == lexer.EOF {
			tokenName = "EOF"
		}

		tokens = append(tokens, tokenInfo{Type: tokenName, Value: tok.Value})

		if tok.Type == lexer.EOF {
			break
		}
	}
	return tokens, nil
}

// filterTokens removes whitespace and newline tokens for cleaner test assertions
func filterTokens(tokens []tokenInfo) []tokenInfo {
	var filtered []tokenInfo
	for _, tok := range tokens {
		if tok.Type != "Whitespace" && tok.Type != "Newline" {
			filtered = append(filtered, tok)
		}
	}
	return filtered
}

func TestLexer(t *testing.T) {
	t.Run("TestLexerBasicDelimiters", func(t *testing.T) {
		input := ",:(){}[]?=<>"

		tokens, err := lexString(input)
		require.NoError(t, err)

		expected := []tokenInfo{
			{Type: "Comma", Value: ","},
			{Type: "Colon", Value: ":"},
			{Type: "LParen", Value: "("},
			{Type: "RParen", Value: ")"},
			{Type: "LBrace", Value: "{"},
			{Type: "RBrace", Value: "}"},
			{Type: "LBracket", Value: "["},
			{Type: "RBracket", Value: "]"},
			{Type: "Question", Value: "?"},
			{Type: "Equals", Value: "="},
			{Type: "LessThan", Value: "<"},
			{Type: "GreaterThan", Value: ">"},
			{Type: "EOF", Value: ""},
		}

		require.Equal(t, expected, tokens)
	})

	t.Run("TestLexerNewLines", func(t *testing.T) {
		input := ",:\n(){}"

		tokens, err := lexString(input)
		require.NoError(t, err)

		expected := []tokenInfo{
			{Type: "Comma", Value: ","},
			{Type: "Colon", Value: ":"},
			{Type: "Newline", Value: "\n"},
			{Type: "LParen", Value: "("},
			{Type: "RParen", Value: ")"},
			{Type: "LBrace", Value: "{"},
			{Type: "RBrace", Value: "}"},
			{Type: "EOF", Value: ""},
		}

		require.Equal(t, expected, tokens)
	})

	t.Run("TestLexerKeywords", func(t *testing.T) {
		input := "include rpc const enum pattern map type proc input output string int float bool datetime deprecated stream"

		tokens, err := lexString(input)
		require.NoError(t, err)

		filtered := filterTokens(tokens)
		expected := []tokenInfo{
			{Type: "Include", Value: "include"},
			{Type: "Rpc", Value: "rpc"},
			{Type: "Const", Value: "const"},
			{Type: "Enum", Value: "enum"},
			{Type: "Pattern", Value: "pattern"},
			{Type: "Map", Value: "map"},
			{Type: "Type", Value: "type"},
			{Type: "Proc", Value: "proc"},
			{Type: "Input", Value: "input"},
			{Type: "Output", Value: "output"},
			{Type: "String", Value: "string"},
			{Type: "Int", Value: "int"},
			{Type: "Float", Value: "float"},
			{Type: "Bool", Value: "bool"},
			{Type: "Datetime", Value: "datetime"},
			{Type: "Deprecated", Value: "deprecated"},
			{Type: "Stream", Value: "stream"},
			{Type: "EOF", Value: ""},
		}

		require.Equal(t, expected, filtered)
	})

	t.Run("TestLexerBooleanLiterals", func(t *testing.T) {
		input := "true false"

		tokens, err := lexString(input)
		require.NoError(t, err)

		filtered := filterTokens(tokens)
		expected := []tokenInfo{
			{Type: "True", Value: "true"},
			{Type: "False", Value: "false"},
			{Type: "EOF", Value: ""},
		}

		require.Equal(t, expected, filtered)
	})

	t.Run("TestLexerIdentifiers", func(t *testing.T) {
		input := "hello world someIdentifier hello123 _underscore"

		tokens, err := lexString(input)
		require.NoError(t, err)

		filtered := filterTokens(tokens)
		expected := []tokenInfo{
			{Type: "Ident", Value: "hello"},
			{Type: "Ident", Value: "world"},
			{Type: "Ident", Value: "someIdentifier"},
			{Type: "Ident", Value: "hello123"},
			{Type: "Ident", Value: "_underscore"},
			{Type: "EOF", Value: ""},
		}

		require.Equal(t, expected, filtered)
	})

	t.Run("TestLexerNumbers", func(t *testing.T) {
		input := "1 2 3 456 789"

		tokens, err := lexString(input)
		require.NoError(t, err)

		filtered := filterTokens(tokens)
		expected := []tokenInfo{
			{Type: "IntLiteral", Value: "1"},
			{Type: "IntLiteral", Value: "2"},
			{Type: "IntLiteral", Value: "3"},
			{Type: "IntLiteral", Value: "456"},
			{Type: "IntLiteral", Value: "789"},
			{Type: "EOF", Value: ""},
		}

		require.Equal(t, expected, filtered)
	})

	t.Run("TestLexerFloats", func(t *testing.T) {
		input := "1.2 3.45 67.89"

		tokens, err := lexString(input)
		require.NoError(t, err)

		filtered := filterTokens(tokens)
		expected := []tokenInfo{
			{Type: "FloatLiteral", Value: "1.2"},
			{Type: "FloatLiteral", Value: "3.45"},
			{Type: "FloatLiteral", Value: "67.89"},
			{Type: "EOF", Value: ""},
		}

		require.Equal(t, expected, filtered)
	})

	t.Run("TestLexerStrings", func(t *testing.T) {
		input := `"hello" "world" "hello world!" "hello \"quotes\" \\"`

		tokens, err := lexString(input)
		require.NoError(t, err)

		filtered := filterTokens(tokens)
		expected := []tokenInfo{
			{Type: "StringLiteral", Value: `"hello"`},
			{Type: "StringLiteral", Value: `"world"`},
			{Type: "StringLiteral", Value: `"hello world!"`},
			{Type: "StringLiteral", Value: `"hello \"quotes\" \\"`},
			{Type: "EOF", Value: ""},
		}

		require.Equal(t, expected, filtered)
	})

	t.Run("TestLexerComments", func(t *testing.T) {
		input := "// This is a comment\ninclude"

		tokens, err := lexString(input)
		require.NoError(t, err)

		filtered := filterTokens(tokens)
		expected := []tokenInfo{
			{Type: "Comment", Value: "// This is a comment"},
			{Type: "Include", Value: "include"},
			{Type: "EOF", Value: ""},
		}

		require.Equal(t, expected, filtered)
	})

	t.Run("TestLexerCommentBlockWithManyStars", func(t *testing.T) {
		input := "/**** Hello ****/"

		tokens, err := lexString(input)
		require.NoError(t, err)

		expected := []tokenInfo{
			{Type: "CommentBlock", Value: "/**** Hello ****/"},
			{Type: "EOF", Value: ""},
		}

		require.Equal(t, expected, tokens)
	})

	t.Run("TestLexerCommentBlockMixedWithDocstring", func(t *testing.T) {
		input := `
			/** Hello **/

			""" Docstring """
			type Foo {}
		`
		tokens, err := lexString(input)
		require.NoError(t, err)

		filtered := filterTokens(tokens)

		var types []string
		for _, tok := range filtered {
			types = append(types, tok.Type)
		}

		expectedTypes := []string{"CommentBlock", "Docstring", "Type", "Ident", "LBrace", "RBrace", "EOF"}
		require.Equal(t, expectedTypes, types)
	})

	t.Run("TestLexerDocstrings", func(t *testing.T) {
		input := `""" This is a docstring """`

		tokens, err := lexString(input)
		require.NoError(t, err)

		expected := []tokenInfo{
			{Type: "Docstring", Value: `""" This is a docstring """`},
			{Type: "EOF", Value: ""},
		}

		require.Equal(t, expected, tokens)
	})

	t.Run("TestLexerDocstringsWithQuotes", func(t *testing.T) {
		input := `""" This is a docstring with ""quotes"" inside """`

		tokens, err := lexString(input)
		require.NoError(t, err)

		expected := []tokenInfo{
			{Type: "Docstring", Value: `""" This is a docstring with ""quotes"" inside """`},
			{Type: "EOF", Value: ""},
		}

		require.Equal(t, expected, tokens)
	})

	t.Run("TestLexerSpread", func(t *testing.T) {
		input := "...AuditMetadata"

		tokens, err := lexString(input)
		require.NoError(t, err)

		expected := []tokenInfo{
			{Type: "Spread", Value: "..."},
			{Type: "Ident", Value: "AuditMetadata"},
			{Type: "EOF", Value: ""},
		}

		require.Equal(t, expected, tokens)
	})

	t.Run("TestLexerMapType", func(t *testing.T) {
		input := "map<int>"

		tokens, err := lexString(input)
		require.NoError(t, err)

		expected := []tokenInfo{
			{Type: "Map", Value: "map"},
			{Type: "LessThan", Value: "<"},
			{Type: "Int", Value: "int"},
			{Type: "GreaterThan", Value: ">"},
			{Type: "EOF", Value: ""},
		}

		require.Equal(t, expected, tokens)
	})

	t.Run("TestLexerEnumWithValues", func(t *testing.T) {
		input := `
			enum Priority {
				Low = 1
				High = 3
			}
		`

		tokens, err := lexString(input)
		require.NoError(t, err)

		filtered := filterTokens(tokens)
		expected := []tokenInfo{
			{Type: "Enum", Value: "enum"},
			{Type: "Ident", Value: "Priority"},
			{Type: "LBrace", Value: "{"},
			{Type: "Ident", Value: "Low"},
			{Type: "Equals", Value: "="},
			{Type: "IntLiteral", Value: "1"},
			{Type: "Ident", Value: "High"},
			{Type: "Equals", Value: "="},
			{Type: "IntLiteral", Value: "3"},
			{Type: "RBrace", Value: "}"},
			{Type: "EOF", Value: ""},
		}

		require.Equal(t, expected, filtered)
	})

	t.Run("TestLexerConstDeclaration", func(t *testing.T) {
		input := `const MAX_PAGE_SIZE = 100`

		tokens, err := lexString(input)
		require.NoError(t, err)

		filtered := filterTokens(tokens)
		expected := []tokenInfo{
			{Type: "Const", Value: "const"},
			{Type: "Ident", Value: "MAX_PAGE_SIZE"},
			{Type: "Equals", Value: "="},
			{Type: "IntLiteral", Value: "100"},
			{Type: "EOF", Value: ""},
		}

		require.Equal(t, expected, filtered)
	})

	t.Run("TestLexerPatternDeclaration", func(t *testing.T) {
		input := `pattern UserEventSubject = "events.users.{userId}.{eventType}"`

		tokens, err := lexString(input)
		require.NoError(t, err)

		filtered := filterTokens(tokens)
		expected := []tokenInfo{
			{Type: "Pattern", Value: "pattern"},
			{Type: "Ident", Value: "UserEventSubject"},
			{Type: "Equals", Value: "="},
			{Type: "StringLiteral", Value: `"events.users.{userId}.{eventType}"`},
			{Type: "EOF", Value: ""},
		}

		require.Equal(t, expected, filtered)
	})

	t.Run("TestLexerRpcBlock", func(t *testing.T) {
		input := `
			rpc Catalog {
				proc CreateProduct {
					input {
						product: Product
					}
					output {
						success: bool
					}
				}
			}
		`

		tokens, err := lexString(input)
		require.NoError(t, err)

		filtered := filterTokens(tokens)
		expected := []tokenInfo{
			{Type: "Rpc", Value: "rpc"},
			{Type: "Ident", Value: "Catalog"},
			{Type: "LBrace", Value: "{"},
			{Type: "Proc", Value: "proc"},
			{Type: "Ident", Value: "CreateProduct"},
			{Type: "LBrace", Value: "{"},
			{Type: "Input", Value: "input"},
			{Type: "LBrace", Value: "{"},
			{Type: "Ident", Value: "product"},
			{Type: "Colon", Value: ":"},
			{Type: "Ident", Value: "Product"},
			{Type: "RBrace", Value: "}"},
			{Type: "Output", Value: "output"},
			{Type: "LBrace", Value: "{"},
			{Type: "Ident", Value: "success"},
			{Type: "Colon", Value: ":"},
			{Type: "Bool", Value: "bool"},
			{Type: "RBrace", Value: "}"},
			{Type: "RBrace", Value: "}"},
			{Type: "RBrace", Value: "}"},
			{Type: "EOF", Value: ""},
		}

		require.Equal(t, expected, filtered)
	})

	t.Run("TestLexerDeprecated", func(t *testing.T) {
		input := `
			deprecated("Use NewType instead")
			type OldType {}
		`

		tokens, err := lexString(input)
		require.NoError(t, err)

		filtered := filterTokens(tokens)
		expected := []tokenInfo{
			{Type: "Deprecated", Value: "deprecated"},
			{Type: "LParen", Value: "("},
			{Type: "StringLiteral", Value: `"Use NewType instead"`},
			{Type: "RParen", Value: ")"},
			{Type: "Type", Value: "type"},
			{Type: "Ident", Value: "OldType"},
			{Type: "LBrace", Value: "{"},
			{Type: "RBrace", Value: "}"},
			{Type: "EOF", Value: ""},
		}

		require.Equal(t, expected, filtered)
	})

	t.Run("TestLexerOptionalField", func(t *testing.T) {
		input := `tags?: string[]`

		tokens, err := lexString(input)
		require.NoError(t, err)

		filtered := filterTokens(tokens)
		expected := []tokenInfo{
			{Type: "Ident", Value: "tags"},
			{Type: "Question", Value: "?"},
			{Type: "Colon", Value: ":"},
			{Type: "String", Value: "string"},
			{Type: "LBracket", Value: "["},
			{Type: "RBracket", Value: "]"},
			{Type: "EOF", Value: ""},
		}

		require.Equal(t, expected, filtered)
	})

	t.Run("TestLexerInclude", func(t *testing.T) {
		input := `include "./foo.vdl"`

		tokens, err := lexString(input)
		require.NoError(t, err)

		filtered := filterTokens(tokens)
		expected := []tokenInfo{
			{Type: "Include", Value: "include"},
			{Type: "StringLiteral", Value: `"./foo.vdl"`},
			{Type: "EOF", Value: ""},
		}

		require.Equal(t, expected, filtered)
	})

	t.Run("TestLexerTypeWithSpread", func(t *testing.T) {
		input := `
			type Article {
				...AuditMetadata
				title: string
			}
		`

		tokens, err := lexString(input)
		require.NoError(t, err)

		filtered := filterTokens(tokens)
		expected := []tokenInfo{
			{Type: "Type", Value: "type"},
			{Type: "Ident", Value: "Article"},
			{Type: "LBrace", Value: "{"},
			{Type: "Spread", Value: "..."},
			{Type: "Ident", Value: "AuditMetadata"},
			{Type: "Ident", Value: "title"},
			{Type: "Colon", Value: ":"},
			{Type: "String", Value: "string"},
			{Type: "RBrace", Value: "}"},
			{Type: "EOF", Value: ""},
		}

		require.Equal(t, expected, filtered)
	})

	t.Run("TestLexerStreamBlock", func(t *testing.T) {
		input := `
			stream NewMessage {
				input {
					chatId: string
				}
				output {
					id: string
				}
			}
		`

		tokens, err := lexString(input)
		require.NoError(t, err)

		filtered := filterTokens(tokens)
		expected := []tokenInfo{
			{Type: "Stream", Value: "stream"},
			{Type: "Ident", Value: "NewMessage"},
			{Type: "LBrace", Value: "{"},
			{Type: "Input", Value: "input"},
			{Type: "LBrace", Value: "{"},
			{Type: "Ident", Value: "chatId"},
			{Type: "Colon", Value: ":"},
			{Type: "String", Value: "string"},
			{Type: "RBrace", Value: "}"},
			{Type: "Output", Value: "output"},
			{Type: "LBrace", Value: "{"},
			{Type: "Ident", Value: "id"},
			{Type: "Colon", Value: ":"},
			{Type: "String", Value: "string"},
			{Type: "RBrace", Value: "}"},
			{Type: "RBrace", Value: "}"},
			{Type: "EOF", Value: ""},
		}

		require.Equal(t, expected, filtered)
	})

	t.Run("TestLexerCompleteVDLFile", func(t *testing.T) {
		input := `
			include "./common.vdl"

			// Single line comment
			/* Multi-line
				comment */

			""" Standalone docstring """

			const VERSION = "1.0.0"
			const MAX_RETRIES = 3
			const TIMEOUT = 1.5
			const ENABLE_LOGS = true

			deprecated("Use NewEnum instead")
			enum OldEnum {
					Legacy
			}

			enum Status {
					Pending
					Active = "ACTIVE"
					Closed = 10
			}

			pattern CacheKey = "item:{id}"

			""" Type docstring """
			type User {
					id: string
					age: int
					score: float
					isActive: bool
					createdAt: datetime
					
					// Arrays and Maps
					tags: string[]
					matrix: int[][]
					metadata: map<string>
					counts: map<map<int>>
					lookup: map<string>[]
					
					// Optional
					profile?: Profile
					
					// Inline object
					address: {
							street: string
							zip: string
					}
					
					// Composition
					...AuditInfo
			}

			rpc UserService {
					""" Proc docstring """
					deprecated
					proc GetUser {
							input {
									id: string
									...Pagination
							}
							output {
									user: User
									// Nested inline object in output
									extra: {
											flag: bool
											meta: {
													source: string
											}
									}
							}
					}

					stream Subscribe {
							input {
									topic: string
							}
							output {
									event: map<string>
							}
					}
			}
		`

		tokens, err := lexString(input)
		require.NoError(t, err)

		filtered := filterTokens(tokens)

		expectedTypes := []string{
			"Include", "StringLiteral",
			"Comment",
			"CommentBlock",
			"Docstring",
			"Const", "Ident", "Equals", "StringLiteral",
			"Const", "Ident", "Equals", "IntLiteral",
			"Const", "Ident", "Equals", "FloatLiteral",
			"Const", "Ident", "Equals", "True",
			"Deprecated", "LParen", "StringLiteral", "RParen",
			"Enum", "Ident", "LBrace",
			"Ident",
			"RBrace",
			"Enum", "Ident", "LBrace",
			"Ident",
			"Ident", "Equals", "StringLiteral",
			"Ident", "Equals", "IntLiteral",
			"RBrace",
			"Pattern", "Ident", "Equals", "StringLiteral",
			"Docstring",
			"Type", "Ident", "LBrace",
			"Ident", "Colon", "String",
			"Ident", "Colon", "Int",
			"Ident", "Colon", "Float",
			"Ident", "Colon", "Bool",
			"Ident", "Colon", "Datetime",
			"Comment",
			"Ident", "Colon", "String", "LBracket", "RBracket",
			"Ident", "Colon", "Int", "LBracket", "RBracket", "LBracket", "RBracket",
			"Ident", "Colon", "Map", "LessThan", "String", "GreaterThan",
			"Ident", "Colon", "Map", "LessThan", "Map", "LessThan", "Int", "GreaterThan", "GreaterThan",
			"Ident", "Colon", "Map", "LessThan", "String", "GreaterThan", "LBracket", "RBracket",
			"Comment",
			"Ident", "Question", "Colon", "Ident",
			"Comment",
			"Ident", "Colon", "LBrace",
			"Ident", "Colon", "String",
			"Ident", "Colon", "String",
			"RBrace",
			"Comment",
			"Spread", "Ident",
			"RBrace",
			"Rpc", "Ident", "LBrace",
			"Docstring",
			"Deprecated",
			"Proc", "Ident", "LBrace",
			"Input", "LBrace",
			"Ident", "Colon", "String",
			"Spread", "Ident",
			"RBrace",
			"Output", "LBrace",
			"Ident", "Colon", "Ident",
			"Comment",
			"Ident", "Colon", "LBrace",
			"Ident", "Colon", "Bool",
			"Ident", "Colon", "LBrace",
			"Ident", "Colon", "String",
			"RBrace",
			"RBrace",
			"RBrace",
			"RBrace",
			"Stream", "Ident", "LBrace",
			"Input", "LBrace",
			"Ident", "Colon", "String",
			"RBrace",
			"Output", "LBrace",
			"Ident", "Colon", "Map", "LessThan", "String", "GreaterThan",
			"RBrace",
			"RBrace",
			"RBrace",
			"EOF",
		}

		var actualTypes []string
		for _, tok := range filtered {
			actualTypes = append(actualTypes, tok.Type)
		}

		require.Equal(t, expectedTypes, actualTypes)
	})
}
