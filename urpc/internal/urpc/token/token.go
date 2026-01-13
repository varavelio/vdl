package token

// TokenType represents the type of a token.
// The string values match the token names used in the lexer rules.
type TokenType string

const (
	// Special tokens (handled by participle)
	EOF = "EOF"

	// Identifiers, comments and docstrings
	Ident        TokenType = "Ident"
	Comment      TokenType = "Comment"      // Single line comment with //
	CommentBlock TokenType = "CommentBlock" // Multiline comment with /* */
	Docstring    TokenType = "Docstring"

	// Literals
	StringLiteral TokenType = "StringLiteral"
	IntLiteral    TokenType = "IntLiteral"
	FloatLiteral  TokenType = "FloatLiteral"
	True          TokenType = "True"
	False         TokenType = "False"

	// Operators and delimiters
	Newline     TokenType = "Newline"
	Whitespace  TokenType = "Whitespace"
	Colon       TokenType = "Colon"
	Comma       TokenType = "Comma"
	LParen      TokenType = "LParen"
	RParen      TokenType = "RParen"
	LBrace      TokenType = "LBrace"
	RBrace      TokenType = "RBrace"
	LBracket    TokenType = "LBracket"
	RBracket    TokenType = "RBracket"
	Question    TokenType = "Question"
	Equals      TokenType = "Equals"
	LessThan    TokenType = "LessThan"
	GreaterThan TokenType = "GreaterThan"
	Spread      TokenType = "Spread" // ...

	// Keywords
	Include    TokenType = "Include"
	Rpc        TokenType = "Rpc"
	Const      TokenType = "Const"
	Enum       TokenType = "Enum"
	Pattern    TokenType = "Pattern"
	Map        TokenType = "Map"
	Deprecated TokenType = "Deprecated"
	Type       TokenType = "Type"
	Proc       TokenType = "Proc"
	Stream     TokenType = "Stream"
	Input      TokenType = "Input"
	Output     TokenType = "Output"
	String     TokenType = "String"
	Int        TokenType = "Int"
	Float      TokenType = "Float"
	Bool       TokenType = "Bool"
	Datetime   TokenType = "Datetime"
)
