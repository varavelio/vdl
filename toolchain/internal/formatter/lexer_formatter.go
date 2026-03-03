package formatter

import (
	"fmt"
	"strings"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/varavelio/vdl/toolchain/internal/core/parser"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

type fmtToken struct {
	Type    string
	Value   string
	Line    int
	Column  int
	Inline  bool
	EndLine int
}

type node interface {
	startLine() int
	endLine() int
	nodeKind() string
}

type baseNode struct {
	start int
	end   int
}

func (n baseNode) startLine() int { return n.start }
func (n baseNode) endLine() int   { return n.end }

type commentNode struct {
	baseNode
	Text   string
	Inline bool
}

func (n *commentNode) nodeKind() string { return "comment" }

type docstringNode struct {
	baseNode
	Raw string
}

func (n *docstringNode) nodeKind() string { return "docstring" }

type annotationNode struct {
	baseNode
	Name string
	Arg  *literalNode
}

type includeNode struct {
	baseNode
	Doc      *docstringNode
	Ann      []*annotationNode
	Path     string
	Trailing *commentNode
}

func (n *includeNode) nodeKind() string { return "include" }

type typeNode struct {
	baseNode
	Doc      *docstringNode
	Ann      []*annotationNode
	Name     string
	Type     fieldTypeNode
	Trailing *commentNode
}

func (n *typeNode) nodeKind() string { return "type" }

type constNode struct {
	baseNode
	Doc      *docstringNode
	Ann      []*annotationNode
	Name     string
	TypeName *string
	Value    literalNode
	Trailing *commentNode
}

func (n *constNode) nodeKind() string { return "const" }

type enumNode struct {
	baseNode
	Doc      *docstringNode
	Ann      []*annotationNode
	Name     string
	Members  []*enumMemberNode
	Trailing *commentNode
}

func (n *enumNode) nodeKind() string { return "enum" }

type enumMemberNode struct {
	baseNode
	Comment  *commentNode
	Spread   *referenceNode
	Doc      *docstringNode
	Ann      []*annotationNode
	Name     string
	Value    *enumValueNode
	Trailing *commentNode
}

type enumValueNode struct {
	Str *string
	Int *string
}

type fieldTypeNode struct {
	Named *string
	Map   *fieldTypeNode
	Obj   *objectTypeNode
	Dims  int
}

type objectTypeNode struct {
	Members []*typeMemberNode
}

type typeMemberNode struct {
	baseNode
	Comment     *commentNode
	Standalone  *docstringNode
	Spread      *referenceNode
	Field       *fieldNode
	Trailing    *commentNode
	HasRendered bool
}

type fieldNode struct {
	baseNode
	Doc      *docstringNode
	Ann      []*annotationNode
	Name     string
	Optional bool
	Type     fieldTypeNode
	Trailing *commentNode
}

type literalNode struct {
	Obj    *objectLiteralNode
	Array  *arrayLiteralNode
	Scalar *scalarLiteralNode
}

type objectLiteralNode struct {
	Entries []*objectEntryNode
}

type objectEntryNode struct {
	baseNode
	Comment  *commentNode
	Spread   *referenceNode
	Key      string
	Value    *literalNode
	Trailing *commentNode
}

type arrayLiteralNode struct {
	Elements []*arrayElementNode
}

type arrayElementNode struct {
	baseNode
	Comment  *commentNode
	Value    *literalNode
	Trailing *commentNode
}

type scalarLiteralNode struct {
	Str   *string
	Float *string
	Int   *string
	True  bool
	False bool
	Ref   *referenceNode
}

type referenceNode struct {
	Name   string
	Member *string
}

type docNode struct {
	Items []node
}

type tokenParser struct {
	tokens []fmtToken
	idx    int
}

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

	w := newFmtWriter()
	printDocument(w, doc)

	out := strutil.LimitConsecutiveNewlines(w.String(), 2)
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

func (p *tokenParser) parseDocument() (*docNode, error) {
	items := []node{}
	var pendingDoc *docstringNode
	var pendingAnn []*annotationNode

	for !p.is("EOF") {
		tok := p.peek()

		if pendingDoc != nil && tok.Line-pendingDoc.endLine() > 1 {
			items = append(items, pendingDoc)
			pendingDoc = nil
		}

		switch tok.Type {
		case "Comment", "CommentBlock":
			if pendingDoc != nil {
				items = append(items, pendingDoc)
				pendingDoc = nil
			}
			c := p.parseComment()
			if c.Inline && len(items) > 0 {
				attachTrailing(items[len(items)-1], c)
			} else {
				items = append(items, c)
			}
		case "Docstring":
			d := p.parseDocstring()
			if pendingDoc != nil {
				items = append(items, pendingDoc)
			}
			pendingDoc = d
		case "At":
			ann, err := p.parseAnnotation()
			if err != nil {
				return nil, err
			}
			pendingAnn = append(pendingAnn, ann)
		case "Include", "Type", "Const", "Enum":
			decl, err := p.parseTopDecl(pendingDoc, pendingAnn)
			if err != nil {
				return nil, err
			}
			pendingDoc = nil
			pendingAnn = nil
			items = append(items, decl)
		default:
			return nil, p.unexpected(tok, "top-level declaration")
		}
	}

	if pendingDoc != nil {
		items = append(items, pendingDoc)
	}

	return &docNode{Items: items}, nil
}

func attachTrailing(prev node, c *commentNode) {
	switch n := prev.(type) {
	case *includeNode:
		n.Trailing = c
	case *typeNode:
		n.Trailing = c
	case *constNode:
		n.Trailing = c
	case *enumNode:
		n.Trailing = c
	}
}

func (p *tokenParser) parseTopDecl(doc *docstringNode, anns []*annotationNode) (node, error) {
	start := p.peek().Line
	switch p.peek().Type {
	case "Include":
		p.next()
		strTok, err := p.expect("StringLiteral")
		if err != nil {
			return nil, err
		}
		return &includeNode{
			baseNode: baseNode{start: start, end: strTok.EndLine},
			Doc:      doc,
			Ann:      anns,
			Path:     unquote(strTok.Value),
		}, nil
	case "Type":
		p.next()
		nameTok, err := p.expect("Ident")
		if err != nil {
			return nil, err
		}
		ft, endLine, err := p.parseFieldType()
		if err != nil {
			return nil, err
		}
		return &typeNode{
			baseNode: baseNode{start: start, end: endLine},
			Doc:      doc,
			Ann:      anns,
			Name:     nameTok.Value,
			Type:     ft,
		}, nil
	case "Const":
		p.next()
		nameTok, err := p.expect("Ident")
		if err != nil {
			return nil, err
		}
		var typeName *string
		if p.isNameToken(p.peek()) && p.peek().Type != "Equals" {
			t := p.next()
			tn := t.Value
			typeName = &tn
		}
		if _, err := p.expect("Equals"); err != nil {
			return nil, err
		}
		lit, endLine, err := p.parseLiteral()
		if err != nil {
			return nil, err
		}
		return &constNode{
			baseNode: baseNode{start: start, end: endLine},
			Doc:      doc,
			Ann:      anns,
			Name:     nameTok.Value,
			TypeName: typeName,
			Value:    lit,
		}, nil
	case "Enum":
		p.next()
		nameTok, err := p.expect("Ident")
		if err != nil {
			return nil, err
		}
		if _, err := p.expect("LBrace"); err != nil {
			return nil, err
		}
		members, endLine, err := p.parseEnumMembers()
		if err != nil {
			return nil, err
		}
		return &enumNode{
			baseNode: baseNode{start: start, end: endLine},
			Doc:      doc,
			Ann:      anns,
			Name:     nameTok.Value,
			Members:  members,
		}, nil
	default:
		return nil, p.unexpected(p.peek(), "declaration")
	}
}

func (p *tokenParser) parseEnumMembers() ([]*enumMemberNode, int, error) {
	members := []*enumMemberNode{}
	var pendingDoc *docstringNode
	var pendingAnn []*annotationNode
	for !p.is("RBrace") && !p.is("EOF") {
		tok := p.peek()
		if pendingDoc != nil && tok.Line-pendingDoc.endLine() > 1 {
			members = append(members, &enumMemberNode{baseNode: baseNode{start: pendingDoc.startLine(), end: pendingDoc.endLine()}, Comment: nil, Doc: pendingDoc})
			pendingDoc = nil
		}

		switch tok.Type {
		case "Comment", "CommentBlock":
			c := p.parseComment()
			if c.Inline && len(members) > 0 {
				members[len(members)-1].Trailing = c
			} else {
				members = append(members, &enumMemberNode{baseNode: baseNode{start: c.startLine(), end: c.endLine()}, Comment: c})
			}
		case "Docstring":
			pendingDoc = p.parseDocstring()
		case "At":
			ann, err := p.parseAnnotation()
			if err != nil {
				return nil, 0, err
			}
			pendingAnn = append(pendingAnn, ann)
		case "Spread":
			spStart := tok.Line
			p.next()
			ref, endLine, err := p.parseReference()
			if err != nil {
				return nil, 0, err
			}
			members = append(members, &enumMemberNode{
				baseNode: baseNode{start: spStart, end: endLine},
				Spread:   ref,
			})
			pendingDoc = nil
			pendingAnn = nil
		default:
			if !p.isNameToken(tok) {
				return nil, 0, p.unexpected(tok, "enum member")
			}
			nameTok := p.next()
			member := &enumMemberNode{
				baseNode: baseNode{start: nameTok.Line, end: nameTok.EndLine},
				Doc:      pendingDoc,
				Ann:      pendingAnn,
				Name:     nameTok.Value,
			}
			if p.is("Equals") {
				p.next()
				if p.is("StringLiteral") {
					t := p.next()
					val := unquote(t.Value)
					member.Value = &enumValueNode{Str: &val}
					member.end = t.EndLine
				} else if p.is("IntLiteral") {
					t := p.next()
					v := t.Value
					member.Value = &enumValueNode{Int: &v}
					member.end = t.EndLine
				} else {
					return nil, 0, p.unexpected(p.peek(), "enum literal")
				}
			}
			members = append(members, member)
			pendingDoc = nil
			pendingAnn = nil
		}
	}

	r, err := p.expect("RBrace")
	if err != nil {
		return nil, 0, err
	}
	return members, r.EndLine, nil
}

func (p *tokenParser) parseFieldType() (fieldTypeNode, int, error) {
	start := p.peek().Line
	var base fieldTypeNode
	switch p.peek().Type {
	case "Map":
		p.next()
		if _, err := p.expect("LBracket"); err != nil {
			return fieldTypeNode{}, 0, err
		}
		inner, _, err := p.parseFieldType()
		if err != nil {
			return fieldTypeNode{}, 0, err
		}
		if _, err := p.expect("RBracket"); err != nil {
			return fieldTypeNode{}, 0, err
		}
		base.Map = &inner
	case "LBrace":
		p.next()
		members, endLine, err := p.parseTypeMembers()
		if err != nil {
			return fieldTypeNode{}, 0, err
		}
		base.Obj = &objectTypeNode{Members: members}
		start = endLine
	default:
		if !p.isTypeNameToken(p.peek()) {
			return fieldTypeNode{}, 0, p.unexpected(p.peek(), "type")
		}
		t := p.next()
		name := t.Value
		base.Named = &name
		start = t.EndLine
	}

	end := start
	for p.is("LBracket") {
		lb := p.next()
		if _, err := p.expect("RBracket"); err != nil {
			return fieldTypeNode{}, 0, err
		}
		base.Dims++
		end = lb.EndLine
	}

	return base, end, nil
}

func (p *tokenParser) parseTypeMembers() ([]*typeMemberNode, int, error) {
	members := []*typeMemberNode{}
	var pendingDoc *docstringNode
	var pendingAnn []*annotationNode
	for !p.is("RBrace") && !p.is("EOF") {
		tok := p.peek()
		if pendingDoc != nil && tok.Line-pendingDoc.endLine() > 1 {
			members = append(members, &typeMemberNode{
				baseNode:   baseNode{start: pendingDoc.startLine(), end: pendingDoc.endLine()},
				Standalone: pendingDoc,
			})
			pendingDoc = nil
		}

		switch tok.Type {
		case "Comment", "CommentBlock":
			c := p.parseComment()
			if c.Inline && len(members) > 0 {
				members[len(members)-1].Trailing = c
			} else {
				members = append(members, &typeMemberNode{baseNode: baseNode{start: c.startLine(), end: c.endLine()}, Comment: c})
			}
		case "Docstring":
			pendingDoc = p.parseDocstring()
		case "At":
			ann, err := p.parseAnnotation()
			if err != nil {
				return nil, 0, err
			}
			pendingAnn = append(pendingAnn, ann)
		case "Spread":
			spTok := p.next()
			ref, endLine, err := p.parseReference()
			if err != nil {
				return nil, 0, err
			}
			members = append(members, &typeMemberNode{
				baseNode: baseNode{start: spTok.Line, end: endLine},
				Spread:   ref,
			})
			pendingDoc = nil
			pendingAnn = nil
		default:
			if !p.isNameToken(tok) {
				return nil, 0, p.unexpected(tok, "field")
			}
			field, err := p.parseField(pendingDoc, pendingAnn)
			if err != nil {
				return nil, 0, err
			}
			members = append(members, &typeMemberNode{
				baseNode: baseNode{start: field.startLine(), end: field.endLine()},
				Field:    field,
			})
			pendingDoc = nil
			pendingAnn = nil
		}
	}

	r, err := p.expect("RBrace")
	if err != nil {
		return nil, 0, err
	}
	return members, r.EndLine, nil
}

func (p *tokenParser) parseField(doc *docstringNode, anns []*annotationNode) (*fieldNode, error) {
	nameTok := p.next()
	f := &fieldNode{
		baseNode: baseNode{start: nameTok.Line, end: nameTok.EndLine},
		Doc:      doc,
		Ann:      anns,
		Name:     nameTok.Value,
	}
	if p.is("Question") {
		p.next()
		f.Optional = true
	}
	typeNode, endLine, err := p.parseFieldType()
	if err != nil {
		return nil, err
	}
	f.Type = typeNode
	f.end = max(f.end, endLine)
	return f, nil
}

func (p *tokenParser) parseLiteral() (literalNode, int, error) {
	tok := p.peek()
	switch tok.Type {
	case "LBrace":
		p.next()
		entries := []*objectEntryNode{}
		for !p.is("RBrace") && !p.is("EOF") {
			if p.is("Comment") || p.is("CommentBlock") {
				c := p.parseComment()
				entries = append(entries, &objectEntryNode{baseNode: baseNode{start: c.startLine(), end: c.endLine()}, Comment: c})
				continue
			}

			if p.is("Spread") {
				sp := p.next()
				ref, end, err := p.parseReference()
				if err != nil {
					return literalNode{}, 0, err
				}
				entries = append(entries, &objectEntryNode{baseNode: baseNode{start: sp.Line, end: end}, Spread: ref})
				continue
			}

			if !p.isNameToken(p.peek()) {
				return literalNode{}, 0, p.unexpected(p.peek(), "object entry")
			}
			k := p.next()
			lit, end, err := p.parseLiteral()
			if err != nil {
				return literalNode{}, 0, err
			}
			entries = append(entries, &objectEntryNode{baseNode: baseNode{start: k.Line, end: end}, Key: k.Value, Value: &lit})
		}
		r, err := p.expect("RBrace")
		if err != nil {
			return literalNode{}, 0, err
		}
		return literalNode{Obj: &objectLiteralNode{Entries: entries}}, r.EndLine, nil
	case "LBracket":
		p.next()
		elements := []*arrayElementNode{}
		for !p.is("RBracket") && !p.is("EOF") {
			if p.is("Comment") || p.is("CommentBlock") {
				c := p.parseComment()
				elements = append(elements, &arrayElementNode{baseNode: baseNode{start: c.startLine(), end: c.endLine()}, Comment: c})
				continue
			}
			lit, end, err := p.parseLiteral()
			if err != nil {
				return literalNode{}, 0, err
			}
			elements = append(elements, &arrayElementNode{baseNode: baseNode{start: tok.Line, end: end}, Value: &lit})
		}
		r, err := p.expect("RBracket")
		if err != nil {
			return literalNode{}, 0, err
		}
		return literalNode{Array: &arrayLiteralNode{Elements: elements}}, r.EndLine, nil
	default:
		s, endLine, err := p.parseScalarLiteral()
		if err != nil {
			return literalNode{}, 0, err
		}
		return literalNode{Scalar: &s}, endLine, nil
	}
}

func (p *tokenParser) parseScalarLiteral() (scalarLiteralNode, int, error) {
	tok := p.peek()
	s := scalarLiteralNode{}
	switch tok.Type {
	case "StringLiteral":
		t := p.next()
		v := unquote(t.Value)
		s.Str = &v
		return s, t.EndLine, nil
	case "FloatLiteral":
		t := p.next()
		v := t.Value
		s.Float = &v
		return s, t.EndLine, nil
	case "IntLiteral":
		t := p.next()
		v := t.Value
		s.Int = &v
		return s, t.EndLine, nil
	case "True":
		p.next()
		s.True = true
		return s, tok.EndLine, nil
	case "False":
		p.next()
		s.False = true
		return s, tok.EndLine, nil
	default:
		if !p.isNameToken(tok) {
			return scalarLiteralNode{}, 0, p.unexpected(tok, "scalar literal")
		}
		ref, endLine, err := p.parseReference()
		if err != nil {
			return scalarLiteralNode{}, 0, err
		}
		s.Ref = ref
		return s, endLine, nil
	}
}

func (p *tokenParser) parseReference() (*referenceNode, int, error) {
	if !p.isNameToken(p.peek()) {
		return nil, 0, p.unexpected(p.peek(), "reference")
	}
	tok := p.next()
	ref := &referenceNode{Name: tok.Value}
	end := tok.EndLine
	if p.is("Dot") {
		p.next()
		if !p.isNameToken(p.peek()) {
			return nil, 0, p.unexpected(p.peek(), "reference member")
		}
		m := p.next().Value
		ref.Member = &m
		end = p.prev().EndLine
	}
	return ref, end, nil
}

func (p *tokenParser) parseAnnotation() (*annotationNode, error) {
	at, err := p.expect("At")
	if err != nil {
		return nil, err
	}
	nameTok, err := p.expect("Ident")
	if err != nil {
		return nil, err
	}
	ann := &annotationNode{
		baseNode: baseNode{start: at.Line, end: nameTok.EndLine},
		Name:     nameTok.Value,
	}
	if p.is("LParen") {
		p.next()
		lit, endLine, err := p.parseLiteral()
		if err != nil {
			return nil, err
		}
		ann.Arg = &lit
		if _, err := p.expect("RParen"); err != nil {
			return nil, err
		}
		ann.end = endLine
	}
	return ann, nil
}

func (p *tokenParser) parseComment() *commentNode {
	tok := p.next()
	return &commentNode{
		baseNode: baseNode{start: tok.Line, end: tok.EndLine},
		Text:     tok.Value,
		Inline:   tok.Inline,
	}
}

func (p *tokenParser) parseDocstring() *docstringNode {
	tok := p.next()
	return &docstringNode{baseNode: baseNode{start: tok.Line, end: tok.EndLine}, Raw: tok.Value}
}

func (p *tokenParser) unexpected(tok fmtToken, expected string) error {
	return fmt.Errorf("format parse error at line %d:%d, expected %s, got %s (%q)", tok.Line, tok.Column, expected, tok.Type, tok.Value)
}

func (p *tokenParser) expect(tt string) (fmtToken, error) {
	if !p.is(tt) {
		return fmtToken{}, p.unexpected(p.peek(), tt)
	}
	return p.next(), nil
}

func (p *tokenParser) is(tt string) bool {
	return p.peek().Type == tt
}

func (p *tokenParser) next() fmtToken {
	t := p.tokens[p.idx]
	p.idx++
	return t
}

func (p *tokenParser) prev() fmtToken {
	if p.idx < 1 {
		return fmtToken{}
	}
	return p.tokens[p.idx-1]
}

func (p *tokenParser) peek() fmtToken {
	if p.idx >= len(p.tokens) {
		return fmtToken{Type: "EOF"}
	}
	return p.tokens[p.idx]
}

func (p *tokenParser) isNameToken(tok fmtToken) bool {
	switch tok.Type {
	case "Ident", "Include", "Const", "Enum", "Map", "Type", "String", "Int", "Float", "Bool", "Datetime", "True", "False":
		return true
	default:
		return false
	}
}

func (p *tokenParser) isTypeNameToken(tok fmtToken) bool {
	switch tok.Type {
	case "Ident", "String", "Int", "Float", "Bool", "Datetime":
		return true
	default:
		return false
	}
}

type fmtWriter struct {
	b           strings.Builder
	indent      int
	lineStarted bool
}

func newFmtWriter() *fmtWriter { return &fmtWriter{} }

func (w *fmtWriter) String() string { return w.b.String() }

func (w *fmtWriter) writeIndent() {
	if w.lineStarted {
		return
	}
	w.b.WriteString(strings.Repeat("  ", w.indent))
	w.lineStarted = true
}

func (w *fmtWriter) line(s string) {
	w.writeIndent()
	w.b.WriteString(s)
	w.b.WriteByte('\n')
	w.lineStarted = false
}

func (w *fmtWriter) lineWithTrailing(s string, trailing *commentNode) {
	if trailing == nil {
		w.line(s)
		return
	}
	w.line(s + " " + trailing.Text)
}

func (w *fmtWriter) blank() {
	if w.b.Len() == 0 {
		return
	}
	if !strings.HasSuffix(w.b.String(), "\n") {
		w.b.WriteByte('\n')
	}
	w.b.WriteByte('\n')
	w.lineStarted = false
}

func printDocument(w *fmtWriter, d *docNode) {
	for i, item := range d.Items {
		if i > 0 && shouldBreakBetweenTop(d.Items[i-1], item) {
			w.blank()
		}
		printTopNode(w, item)
	}
}

func shouldBreakBetweenTop(prev, curr node) bool {
	if prev == nil || curr == nil {
		return false
	}
	if prev.nodeKind() == "include" && curr.nodeKind() == "include" {
		return false
	}
	if prev.nodeKind() == "comment" || curr.nodeKind() == "comment" {
		return curr.startLine()-prev.endLine() > 1
	}
	if prev.nodeKind() == "docstring" || curr.nodeKind() == "docstring" {
		return true
	}
	prevConst, prevIsConst := prev.(*constNode)
	currConst, currIsConst := curr.(*constNode)
	if prevIsConst && currIsConst {
		if isSingleLineConst(prevConst) && isSingleLineConst(currConst) {
			return false
		}
		return true
	}
	return true
}

func printTopNode(w *fmtWriter, n node) {
	switch t := n.(type) {
	case *commentNode:
		w.line(t.Text)
	case *docstringNode:
		printDocstring(w, t.Raw)
	case *includeNode:
		if t.Doc != nil {
			printDocstring(w, t.Doc.Raw)
		}
		for _, a := range t.Ann {
			printAnnotation(w, a)
		}
		w.lineWithTrailing(`include "`+strutil.EscapeQuotes(t.Path)+`"`, t.Trailing)
	case *typeNode:
		if t.Doc != nil {
			printDocstring(w, t.Doc.Raw)
		}
		for _, a := range t.Ann {
			printAnnotation(w, a)
		}
		name := strutil.ToPascalCase(t.Name)
		if t.Type.Obj != nil {
			if len(t.Type.Obj.Members) == 0 {
				w.lineWithTrailing("type "+name+" {}"+strings.Repeat("[]", t.Type.Dims), t.Trailing)
				break
			}
			w.line("type " + name + " {")
			w.indent++
			for i, m := range t.Type.Obj.Members {
				if i > 0 && shouldBreakTypeMembers(t.Type.Obj.Members[i-1], m) {
					w.blank()
				}
				printTypeMember(w, m)
			}
			w.indent--
			w.lineWithTrailing("}"+strings.Repeat("[]", t.Type.Dims), t.Trailing)
		} else {
			typeText := renderFieldType(t.Type, typeRenderCtx{namedRef: refTypeDecl, spreadRef: refTypeDecl, scalarRef: refConstDecl, enumMemberRef: refEnumMember})
			w.lineWithTrailing("type "+name+" "+typeText, t.Trailing)
		}
	case *constNode:
		if t.Doc != nil {
			printDocstring(w, t.Doc.Raw)
		}
		for _, a := range t.Ann {
			printAnnotation(w, a)
		}
		tn := ""
		if t.TypeName != nil {
			tn = " " + formatTypeName(*t.TypeName)
		}
		lhs := "const " + strutil.ToCamelCase(t.Name) + tn + " = "
		rhs := renderLiteral(t.Value, literalRenderCtx{spreadRef: refConstDecl, scalarRef: refConstDecl, enumMemberRef: refEnumMember})
		printMultilineStatement(w, lhs, rhs, t.Trailing)
	case *enumNode:
		if t.Doc != nil {
			printDocstring(w, t.Doc.Raw)
		}
		for _, a := range t.Ann {
			printAnnotation(w, a)
		}
		name := strutil.ToPascalCase(t.Name)
		if len(t.Members) == 0 {
			w.lineWithTrailing("enum "+name+" {}", t.Trailing)
			return
		}
		w.line("enum " + name + " {")
		w.indent++
		for i, m := range t.Members {
			if i > 0 && shouldBreakEnumMembers(t.Members[i-1], m) {
				w.blank()
			}
			printEnumMember(w, m)
		}
		w.indent--
		w.line("}")
	}
}

func shouldBreakEnumMembers(prev, curr *enumMemberNode) bool {
	if prev == nil || curr == nil {
		return false
	}
	if curr.Comment != nil || prev.Comment != nil {
		return curr.startLine()-prev.endLine() > 1
	}
	if curr.Doc != nil {
		return true
	}
	return false
}

func printEnumMember(w *fmtWriter, m *enumMemberNode) {
	if m.Comment != nil {
		w.line(m.Comment.Text)
		return
	}
	if m.Doc != nil {
		printDocstring(w, m.Doc.Raw)
	}
	for _, a := range m.Ann {
		printAnnotation(w, a)
	}
	if m.Spread != nil {
		w.lineWithTrailing("..."+renderReference(*m.Spread, refEnumDecl), m.Trailing)
		return
	}
	if m.Name == "" {
		return
	}
	line := strutil.ToPascalCase(m.Name)
	if m.Value != nil {
		if m.Value.Str != nil {
			line += ` = "` + strutil.EscapeQuotes(*m.Value.Str) + `"`
		} else if m.Value.Int != nil {
			line += " = " + *m.Value.Int
		}
	}
	w.lineWithTrailing(line, m.Trailing)
}

func printAnnotation(w *fmtWriter, a *annotationNode) {
	name := strutil.ToCamelCase(a.Name)
	if a.Arg == nil {
		w.line("@" + name)
		return
	}
	renderedArg := renderLiteral(*a.Arg, literalRenderCtx{spreadRef: refConstDecl, scalarRef: refConstDecl, enumMemberRef: refEnumMember})
	if !strings.Contains(renderedArg, "\n") {
		w.line("@" + name + "(" + renderedArg + ")")
		return
	}
	lines := strings.Split(renderedArg, "\n")
	w.line("@" + name + "(" + lines[0])
	for i := 1; i < len(lines)-1; i++ {
		w.line(lines[i])
	}
	w.line(lines[len(lines)-1] + ")")
}

func printDocstring(w *fmtWriter, raw string) {
	content := strings.TrimPrefix(raw, `"""`)
	content = strings.TrimSuffix(content, `"""`)
	normalized := content
	if strings.Contains(normalized, "\n") {
		normalized = strutil.NormalizeIndent(normalized)
		lines := strings.Split(normalized, "\n")
		for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
			lines = lines[1:]
		}
		for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
			lines = lines[:len(lines)-1]
		}
		if len(lines) <= 1 {
			if len(lines) == 0 {
				w.line(`""" """`)
				return
			}
			w.line(`""" ` + strings.TrimSpace(lines[0]) + ` """`)
			return
		}
		w.line(`"""`)
		for _, l := range lines {
			w.line(l)
		}
		w.line(`"""`)
		return
	}
	trimmed := strings.TrimSpace(normalized)
	w.line(`""" ` + trimmed + ` """`)
}

func formatTypeName(name string) string {
	if isPrimitiveName(name) {
		return name
	}
	return strutil.ToPascalCase(name)
}

func isPrimitiveName(name string) bool {
	switch name {
	case "string", "int", "float", "bool", "datetime":
		return true
	default:
		return false
	}
}

type refCase int

const (
	refTypeDecl refCase = iota
	refEnumDecl
	refConstDecl
	refEnumMember
)

type typeRenderCtx struct {
	namedRef      refCase
	spreadRef     refCase
	scalarRef     refCase
	enumMemberRef refCase
}

type literalRenderCtx struct {
	spreadRef     refCase
	scalarRef     refCase
	enumMemberRef refCase
}

func renderFieldType(ft fieldTypeNode, ctx typeRenderCtx) string {
	base := ""
	switch {
	case ft.Named != nil:
		base = formatTypeName(*ft.Named)
	case ft.Map != nil:
		base = "map[" + renderFieldType(*ft.Map, ctx) + "]"
	case ft.Obj != nil:
		base = renderObjectType(*ft.Obj)
	}
	for i := 0; i < ft.Dims; i++ {
		base += "[]"
	}
	return base
}

func renderObjectType(obj objectTypeNode) string {
	if len(obj.Members) == 0 {
		return "{}"
	}
	w := newFmtWriter()
	w.line("{")
	w.indent++
	for i, m := range obj.Members {
		if i > 0 && shouldBreakTypeMembers(obj.Members[i-1], m) {
			w.blank()
		}
		printTypeMember(w, m)
	}
	w.indent--
	w.line("}")
	return strings.TrimSuffix(strings.TrimPrefix(w.String(), ""), "\n")
}

func shouldBreakTypeMembers(prev, curr *typeMemberNode) bool {
	if prev == nil || curr == nil {
		return false
	}
	if curr.Comment != nil || prev.Comment != nil {
		return curr.startLine()-prev.endLine() > 1
	}
	if curr.Standalone != nil {
		return true
	}
	if curr.Field != nil && curr.Field.Doc != nil {
		return true
	}
	if prev.Field != nil && curr.Field != nil {
		if !isSingleLineField(prev.Field) || !isSingleLineField(curr.Field) {
			return true
		}
		return false
	}
	if (prev.Field != nil && !isSingleLineField(prev.Field)) || (curr.Field != nil && !isSingleLineField(curr.Field)) {
		return true
	}
	return false
}

func printTypeMember(w *fmtWriter, m *typeMemberNode) {
	if m.Comment != nil {
		w.line(m.Comment.Text)
		return
	}
	if m.Standalone != nil {
		printDocstring(w, m.Standalone.Raw)
		return
	}
	if m.Spread != nil {
		w.lineWithTrailing("..."+renderReference(*m.Spread, refTypeDecl), m.Trailing)
		return
	}
	if m.Field != nil {
		if m.Trailing != nil && m.Field.Trailing == nil {
			m.Field.Trailing = m.Trailing
		}
		printField(w, m.Field)
	}
}

func printField(w *fmtWriter, f *fieldNode) {
	if f.Doc != nil {
		printDocstring(w, f.Doc.Raw)
	}
	for _, a := range f.Ann {
		printAnnotation(w, a)
	}
	name := strutil.ToCamelCase(f.Name)
	if f.Optional {
		name += "?"
	}
	if f.Type.Obj != nil {
		if len(f.Type.Obj.Members) == 0 {
			w.lineWithTrailing(name+" {}"+strings.Repeat("[]", f.Type.Dims), f.Trailing)
			return
		}
		w.line(name + " {")
		w.indent++
		for i, m := range f.Type.Obj.Members {
			if i > 0 && shouldBreakTypeMembers(f.Type.Obj.Members[i-1], m) {
				w.blank()
			}
			printTypeMember(w, m)
		}
		w.indent--
		w.lineWithTrailing("}"+strings.Repeat("[]", f.Type.Dims), f.Trailing)
		return
	}
	w.lineWithTrailing(name+" "+renderFieldType(f.Type, typeRenderCtx{namedRef: refTypeDecl}), f.Trailing)
}

func renderLiteral(l literalNode, ctx literalRenderCtx) string {
	if l.Obj != nil {
		return renderObjectLiteral(*l.Obj, ctx)
	}
	if l.Array != nil {
		return renderArrayLiteral(*l.Array, ctx)
	}
	if l.Scalar != nil {
		return renderScalar(*l.Scalar, ctx)
	}
	return ""
}

func renderObjectLiteral(o objectLiteralNode, ctx literalRenderCtx) string {
	if len(o.Entries) == 0 {
		return "{}"
	}
	simple := true
	for _, e := range o.Entries {
		if e.Comment != nil {
			simple = false
			break
		}
	}
	if simple && len(o.Entries) == 1 && o.Entries[0].Spread == nil {
		e := o.Entries[0]
		if e.Value != nil && e.Value.Scalar != nil {
			return "{ " + strutil.ToCamelCase(e.Key) + " " + renderLiteral(*e.Value, ctx) + " }"
		}
	}
	w := newFmtWriter()
	w.line("{")
	w.indent++
	for _, e := range o.Entries {
		if e.Comment != nil {
			w.line(e.Comment.Text)
			continue
		}
		if e.Spread != nil {
			w.line("..." + renderReference(*e.Spread, ctx.spreadRef))
			continue
		}
		key := strutil.ToCamelCase(e.Key)
		val := renderLiteral(*e.Value, ctx)
		if !strings.Contains(val, "\n") {
			w.line(key + " " + val)
			continue
		}
		lines := strings.Split(val, "\n")
		w.line(key + " " + lines[0])
		for i := 1; i < len(lines); i++ {
			w.line(lines[i])
		}
	}
	w.indent--
	w.line("}")
	return strings.TrimSuffix(w.String(), "\n")
}

func renderArrayLiteral(a arrayLiteralNode, ctx literalRenderCtx) string {
	if len(a.Elements) == 0 {
		return "[]"
	}
	parts := make([]string, 0, len(a.Elements))
	hasMultiline := false
	for _, e := range a.Elements {
		if e.Comment != nil {
			continue
		}
		rendered := renderLiteral(*e.Value, ctx)
		if strings.Contains(rendered, "\n") {
			hasMultiline = true
		}
		parts = append(parts, rendered)
	}
	if !hasMultiline {
		return "[" + strings.Join(parts, " ") + "]"
	}
	w := newFmtWriter()
	w.line("[")
	w.indent++
	for _, part := range parts {
		if !strings.Contains(part, "\n") {
			w.line(part)
			continue
		}
		for _, line := range strings.Split(part, "\n") {
			w.line(line)
		}
	}
	w.indent--
	w.line("]")
	return strings.TrimSuffix(w.String(), "\n")
}

func renderScalar(s scalarLiteralNode, ctx literalRenderCtx) string {
	if s.Str != nil {
		return `"` + strutil.EscapeQuotes(*s.Str) + `"`
	}
	if s.Float != nil {
		return *s.Float
	}
	if s.Int != nil {
		return *s.Int
	}
	if s.True {
		return "true"
	}
	if s.False {
		return "false"
	}
	if s.Ref != nil {
		if s.Ref.Member != nil {
			return renderReference(*s.Ref, ctx.enumMemberRef)
		}
		return renderReference(*s.Ref, ctx.scalarRef)
	}
	return ""
}

func renderReference(r referenceNode, c refCase) string {
	name := r.Name
	switch c {
	case refTypeDecl, refEnumDecl:
		name = strutil.ToPascalCase(name)
	case refConstDecl:
		name = strutil.ToCamelCase(name)
	}
	if r.Member != nil {
		return strutil.ToPascalCase(name) + "." + strutil.ToPascalCase(*r.Member)
	}
	return name
}

func unquote(s string) string {
	if len(s) >= 2 && strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`) {
		return s[1 : len(s)-1]
	}
	return s
}

func printMultilineStatement(w *fmtWriter, lhs, rhs string, trailing *commentNode) {
	if !strings.Contains(rhs, "\n") {
		w.lineWithTrailing(lhs+rhs, trailing)
		return
	}
	lines := strings.Split(rhs, "\n")
	w.line(lhs + lines[0])
	for i := 1; i < len(lines)-1; i++ {
		w.line(lines[i])
	}
	last := lines[len(lines)-1]
	if trailing != nil {
		w.line(last + " " + trailing.Text)
		return
	}
	w.line(last)
}

func isSingleLineConst(c *constNode) bool {
	if c == nil {
		return false
	}
	if c.Doc != nil || len(c.Ann) > 0 {
		return false
	}
	rhs := renderLiteral(c.Value, literalRenderCtx{spreadRef: refConstDecl, scalarRef: refConstDecl, enumMemberRef: refEnumMember})
	return !strings.Contains(rhs, "\n")
}

func isSingleLineField(f *fieldNode) bool {
	if f == nil {
		return false
	}
	if f.Doc != nil || len(f.Ann) > 0 {
		return false
	}
	if f.Type.Obj != nil && len(f.Type.Obj.Members) > 0 {
		return false
	}
	return true
}
