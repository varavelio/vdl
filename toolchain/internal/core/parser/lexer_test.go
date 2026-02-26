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

func TestLexerDelimitersAndOperators(t *testing.T) {
	tokens, err := lexString("@(){}[]?=")
	require.NoError(t, err)

	require.Equal(t, []tokenInfo{
		{Type: "At", Value: "@"},
		{Type: "LParen", Value: "("},
		{Type: "RParen", Value: ")"},
		{Type: "LBrace", Value: "{"},
		{Type: "RBrace", Value: "}"},
		{Type: "LBracket", Value: "["},
		{Type: "RBracket", Value: "]"},
		{Type: "Question", Value: "?"},
		{Type: "Equals", Value: "="},
		{Type: "EOF", Value: ""},
	}, tokens)
}

func TestLexerKeywords(t *testing.T) {
	tokens, err := lexString("include const enum map type string int float bool datetime")
	require.NoError(t, err)

	require.Equal(t, []tokenInfo{
		{Type: "Include", Value: "include"},
		{Type: "Const", Value: "const"},
		{Type: "Enum", Value: "enum"},
		{Type: "Map", Value: "map"},
		{Type: "Type", Value: "type"},
		{Type: "String", Value: "string"},
		{Type: "Int", Value: "int"},
		{Type: "Float", Value: "float"},
		{Type: "Bool", Value: "bool"},
		{Type: "Datetime", Value: "datetime"},
		{Type: "EOF", Value: ""},
	}, filterTokens(tokens))
}

func TestLexerRemovedKeywordsBecomeIdentifiers(t *testing.T) {
	tokens, err := lexString("rpc proc stream input output pattern deprecated")
	require.NoError(t, err)

	require.Equal(t, []tokenInfo{
		{Type: "Ident", Value: "rpc"},
		{Type: "Ident", Value: "proc"},
		{Type: "Ident", Value: "stream"},
		{Type: "Ident", Value: "input"},
		{Type: "Ident", Value: "output"},
		{Type: "Ident", Value: "pattern"},
		{Type: "Ident", Value: "deprecated"},
		{Type: "EOF", Value: ""},
	}, filterTokens(tokens))
}

func TestLexerTypeAndConstSamples(t *testing.T) {
	input := `
		@rpc type Chat {
			@proc SendMessage {
				input {
					chatId string
				}
			}
		}

		const appConfig AppConfigType = {
			...baseConfig
			port 8080
			targets [
			  {
			    go {
					  output "./gen/go"
					}
				}
			]
		}
	`

	tokens, err := lexString(input)
	require.NoError(t, err)
	filtered := filterTokens(tokens)

	var hasAt bool
	var hasSpread bool
	for _, tok := range filtered {
		if tok.Type == "At" {
			hasAt = true
		}
		if tok.Type == "Spread" {
			hasSpread = true
		}

	}

	require.True(t, hasAt)
	require.True(t, hasSpread)
}

func TestLexerLiteralsAndComments(t *testing.T) {
	t.Run("boolean and numeric literals", func(t *testing.T) {
		tokens, err := lexString(`true false 1 2 3 1.5 "hello"`)
		require.NoError(t, err)

		require.Equal(t, []tokenInfo{
			{Type: "True", Value: "true"},
			{Type: "False", Value: "false"},
			{Type: "IntLiteral", Value: "1"},
			{Type: "IntLiteral", Value: "2"},
			{Type: "IntLiteral", Value: "3"},
			{Type: "FloatLiteral", Value: "1.5"},
			{Type: "StringLiteral", Value: `"hello"`},
			{Type: "EOF", Value: ""},
		}, filterTokens(tokens))
	})

	t.Run("docstring comments and spread", func(t *testing.T) {
		input := `
			/* block */
			// line
			""" docs """
			...Base
		`
		tokens, err := lexString(input)
		require.NoError(t, err)

		types := []string{}
		for _, tok := range filterTokens(tokens) {
			types = append(types, tok.Type)
		}

		require.Equal(t, []string{"CommentBlock", "Comment", "Docstring", "Spread", "Ident", "EOF"}, types)
	})
}

func TestLexerFullSampleTypes(t *testing.T) {
	input := `
		include "./common.vdl"

		@rpc type Chat {
			@proc SendMessage {
				input {
					chatId string[]
					message map[string]
				}
				output {
					messageId string
				}
			}
		}

		@pattern const UserTopic = "events.users.{userId}"

		enum AllRoles {
			SuperAdmin = "super"
			...StandardRoles
		}
	`

	tokens, err := lexString(input)
	require.NoError(t, err)

	filtered := filterTokens(tokens)
	actualTypes := []string{}
	for _, tok := range filtered {
		actualTypes = append(actualTypes, tok.Type)
	}

	require.Equal(t, []string{
		"Include", "StringLiteral",
		"At", "Ident", "Type", "Ident", "LBrace",
		"At", "Ident", "Ident", "LBrace",
		"Ident", "LBrace", "Ident", "String", "LBracket", "RBracket", "Ident", "LBracket", "String", "RBracket", "RBrace",
		"Ident", "LBrace", "Ident", "String", "RBrace",
		"RBrace", "RBrace",
		"At", "Ident", "Const", "Ident", "Equals", "StringLiteral",
		"Enum", "Ident", "LBrace", "Ident", "Equals", "StringLiteral", "Spread", "Ident", "RBrace",
		"EOF",
	}, actualTypes)
}
