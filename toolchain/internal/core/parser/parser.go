package parser

import (
	"github.com/alecthomas/participle/v2"
	"github.com/varavelio/vdl/toolchain/internal/core/ast"
)

// Error is an alias for participle.Error
type Error = participle.Error

// Parser is an alias for participle.Parser with ast.Schema as the root node
type Parser = participle.Parser[ast.Schema]

// ParserInstance is a pre-built parser instance for VDL schemas.
var ParserInstance = participle.MustBuild[ast.Schema](
	participle.Lexer(VDLLexer),
	participle.Elide("Newline", "Whitespace", "Comment", "CommentBlock"),
	// We need lookahead to distinguish associated docstrings/annotations and declaration starts.
	participle.UseLookahead(8),
)
