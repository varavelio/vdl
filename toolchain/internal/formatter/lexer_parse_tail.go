package formatter

import "fmt"

func (p *tokenParser) parseTypeMembers() ([]*typeMemberNode, int, error) {
	members := []*typeMemberNode{}
	var pendingDoc *docstringNode
	var pendingAnn []*annotationNode

	for !p.is("RBrace") && !p.is("EOF") {
		tok := p.peek()
		if pendingDoc != nil && tok.Line-pendingAttachmentEndLine(pendingDoc, pendingAnn) > 1 {
			members = append(members, &typeMemberNode{baseNode: baseNode{start: pendingDoc.startLine(), end: pendingDoc.endLine()}, Standalone: pendingDoc})
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
			members = append(members, &typeMemberNode{baseNode: baseNode{start: spTok.Line, end: endLine}, Spread: ref})
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
			members = append(members, &typeMemberNode{baseNode: baseNode{start: field.startLine(), end: field.endLine()}, Field: field})
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
	f := &fieldNode{baseNode: baseNode{start: nameTok.Line, end: nameTok.EndLine}, Doc: doc, Ann: anns, Name: nameTok.Value}
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
				if c.Inline && len(entries) > 0 {
					entries[len(entries)-1].Trailing = c
				} else {
					entries = append(entries, &objectEntryNode{baseNode: baseNode{start: c.startLine(), end: c.endLine()}, Comment: c})
				}
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
		lb := p.next()
		multilineIntent := !p.is("RBracket") && p.peek().Line > lb.Line
		elements := []*arrayElementNode{}
		for !p.is("RBracket") && !p.is("EOF") {
			if p.is("Comment") || p.is("CommentBlock") {
				c := p.parseComment()
				if c.Inline && len(elements) > 0 {
					elements[len(elements)-1].Trailing = c
				} else {
					elements = append(elements, &arrayElementNode{baseNode: baseNode{start: c.startLine(), end: c.endLine()}, Comment: c})
				}
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
		return literalNode{Array: &arrayLiteralNode{Elements: elements, MultilineIntent: multilineIntent}}, r.EndLine, nil
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
	ann := &annotationNode{baseNode: baseNode{start: at.Line, end: nameTok.EndLine}, Name: nameTok.Value}
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
	return &commentNode{baseNode: baseNode{start: tok.Line, end: tok.EndLine}, Text: tok.Value, Inline: tok.Inline}
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

func (p *tokenParser) is(tt string) bool { return p.peek().Type == tt }

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
