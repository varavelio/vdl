package formatter

import (
	"strings"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/varavelio/vdl/toolchain/internal/core/parser"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

func formatLexerBased(filename, content string) (string, error) {
	tokens, err := lexTokens(filename, content)
	if err != nil {
		return "", err
	}

	p := &tokenParser{tokens: tokens}
	doc, err := p.parseDocument()
	if err != nil {
		return "", err
	}

	output := newFormatterOutput()
	printDocument(output, doc)

	out := strutil.LimitConsecutiveNewlines(output.String(), 2)
	out = strings.TrimSpace(out)
	if out == "" {
		return "", nil
	}
	return out + "\n", nil
}

func lexTokens(filename, content string) ([]fmtToken, error) {
	lex, err := parser.VDLLexer.LexString(filename, content)
	if err != nil {
		return nil, err
	}

	symbolToName := map[lexer.TokenType]string{}
	for name, sym := range parser.VDLLexer.Symbols() {
		symbolToName[sym] = name
	}

	lines := strings.Split(content, "\n")
	result := make([]fmtToken, 0, len(content)/4)
	for {
		tok, err := lex.Next()
		if err != nil {
			return nil, err
		}

		if tok.Type == lexer.EOF {
			result = append(result, fmtToken{Type: "EOF", Line: tok.Pos.Line, Column: tok.Pos.Column, EndLine: tok.Pos.Line})
			break
		}

		typeName := symbolToName[tok.Type]
		if typeName == "Whitespace" || typeName == "Newline" {
			continue
		}

		line := tok.Pos.Line
		col := tok.Pos.Column
		inline := false
		if (typeName == "Comment" || typeName == "CommentBlock") && line > 0 && line <= len(lines) {
			prefix := ""
			if col > 1 {
				row := lines[line-1]
				if col-1 <= len(row) {
					prefix = row[:col-1]
				} else {
					prefix = row
				}
			}
			inline = strings.TrimSpace(prefix) != ""
		}

		endLine := line + strings.Count(tok.Value, "\n")
		result = append(result, fmtToken{
			Type:    typeName,
			Value:   tok.Value,
			Line:    line,
			Column:  col,
			Inline:  inline,
			EndLine: endLine,
		})
	}

	return result, nil
}

func unquote(s string) string {
	if len(s) >= 2 && strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`) {
		return s[1 : len(s)-1]
	}
	return s
}
