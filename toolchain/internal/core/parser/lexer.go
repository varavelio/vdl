package parser

import (
	"github.com/alecthomas/participle/v2/lexer"
)

// URPCLexer is the participle lexer definition for the UFO language.
// It uses regex-based lexing provided by participle.
//
// Token order matters - more specific patterns must come before general ones.
// Keywords must come before Ident to have higher priority.
var URPCLexer = lexer.MustSimple([]lexer.SimpleRule{
	// Docstrings (triple quoted strings) - must come before StringLiteral
	// Match """ followed by characters, allowing single/double quotes but not triple, until closing """
	{Name: "Docstring", Pattern: `"""(?:[^"]+|"[^"]|""[^"])*"""`},

	// Comments
	{Name: "CommentBlock", Pattern: `/\*([^*]|\*[^/])*\*/`},
	{Name: "Comment", Pattern: `//[^\n]*`},

	// Keywords (must come before Ident to have higher priority)
	{Name: "Include", Pattern: `\binclude\b`},
	{Name: "Rpc", Pattern: `\brpc\b`},
	{Name: "Const", Pattern: `\bconst\b`},
	{Name: "Enum", Pattern: `\benum\b`},
	{Name: "Pattern", Pattern: `\bpattern\b`},
	{Name: "Map", Pattern: `\bmap\b`},
	{Name: "Deprecated", Pattern: `\bdeprecated\b`},
	{Name: "Type", Pattern: `\btype\b`},
	{Name: "Proc", Pattern: `\bproc\b`},
	{Name: "Stream", Pattern: `\bstream\b`},
	{Name: "Input", Pattern: `\binput\b`},
	{Name: "Output", Pattern: `\boutput\b`},
	{Name: "String", Pattern: `\bstring\b`},
	{Name: "Int", Pattern: `\bint\b`},
	{Name: "Float", Pattern: `\bfloat\b`},
	{Name: "Bool", Pattern: `\bbool\b`},
	{Name: "Datetime", Pattern: `\bdatetime\b`},

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
	{Name: "Colon", Pattern: `:`},
	{Name: "Comma", Pattern: `,`},
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
