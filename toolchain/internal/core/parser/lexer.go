package parser

import (
	"github.com/alecthomas/participle/v2/lexer"
)

// VDLLexer is the participle lexer definition for the VDL language.
// It uses regex-based lexing provided by participle.
//
// Token order matters - more specific patterns must come before general ones.
// Keywords must come before Ident to have higher priority.
var VDLLexer = lexer.MustSimple([]lexer.SimpleRule{
	// Docstrings (triple quoted strings) - must come before StringLiteral
	// Match """ followed by characters, allowing single/double quotes but not triple, until closing """
	{Name: "Docstring", Pattern: `"""(?:[^"]+|"[^"]|""[^"])*"""`},

	// Comments
	{Name: "CommentBlock", Pattern: `(?s)/\*.*?\*/`},
	{Name: "Comment", Pattern: `//[^\n]*`},

	// Keywords (must come before Ident to have higher priority)
	{Name: "Include", Pattern: `\binclude\b`},
	{Name: "Const", Pattern: `\bconst\b`},
	{Name: "Enum", Pattern: `\benum\b`},
	{Name: "Type", Pattern: `\btype\b`},
	{Name: "String", Pattern: `\bstring\b`},
	{Name: "Int", Pattern: `\bint\b`},
	{Name: "Float", Pattern: `\bfloat\b`},
	{Name: "Bool", Pattern: `\bbool\b`},
	{Name: "Datetime", Pattern: `\bdatetime\b`},
	{Name: "Map", Pattern: `\bmap\b`},

	// Boolean literals
	{Name: "True", Pattern: `\btrue\b`},
	{Name: "False", Pattern: `\bfalse\b`},

	// Literals
	{Name: "FloatLiteral", Pattern: `[0-9]+\.[0-9]+`},
	{Name: "IntLiteral", Pattern: `[0-9]+`},
	{Name: "StringLiteral", Pattern: `"(?:\\"|\\\\|[^"])*"`},

	// Identifiers
	{Name: "Ident", Pattern: `[a-zA-Z_][a-zA-Z0-9_]*`},

	// Spread operator (must come before Illegal to catch it)
	{Name: "Spread", Pattern: `\.\.\.`},

	// Delimiters and operators
	{Name: "Newline", Pattern: `\n`},
	{Name: "At", Pattern: `@`},
	{Name: "LParen", Pattern: `\(`},
	{Name: "RParen", Pattern: `\)`},
	{Name: "LBrace", Pattern: `\{`},
	{Name: "RBrace", Pattern: `\}`},
	{Name: "LBracket", Pattern: `\[`},
	{Name: "RBracket", Pattern: `\]`},
	{Name: "Question", Pattern: `\?`},
	{Name: "Equals", Pattern: `=`},
	{Name: "LessThan", Pattern: `<`},
	{Name: "GreaterThan", Pattern: `>`},

	// Whitespace (excluding newlines)
	{Name: "Whitespace", Pattern: `[ \t\r]+`},
})
