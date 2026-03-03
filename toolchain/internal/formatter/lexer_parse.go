package formatter

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
		return &includeNode{baseNode: baseNode{start: start, end: strTok.EndLine}, Doc: doc, Ann: anns, Path: unquote(strTok.Value)}, nil
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
		return &typeNode{baseNode: baseNode{start: start, end: endLine}, Doc: doc, Ann: anns, Name: nameTok.Value, Type: ft}, nil
	case "Const":
		p.next()
		nameTok, err := p.expect("Ident")
		if err != nil {
			return nil, err
		}
		var typeName *string
		if p.isNameToken(p.peek()) && p.peek().Type != "Equals" {
			t := p.next().Value
			typeName = &t
		}
		if _, err := p.expect("Equals"); err != nil {
			return nil, err
		}
		lit, endLine, err := p.parseLiteral()
		if err != nil {
			return nil, err
		}
		return &constNode{baseNode: baseNode{start: start, end: endLine}, Doc: doc, Ann: anns, Name: nameTok.Value, TypeName: typeName, Value: lit}, nil
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
		return &enumNode{baseNode: baseNode{start: start, end: endLine}, Doc: doc, Ann: anns, Name: nameTok.Value, Members: members}, nil
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
			members = append(members, &enumMemberNode{baseNode: baseNode{start: pendingDoc.startLine(), end: pendingDoc.endLine()}, Doc: pendingDoc})
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
			members = append(members, &enumMemberNode{baseNode: baseNode{start: spStart, end: endLine}, Spread: ref})
			pendingDoc = nil
			pendingAnn = nil
		default:
			if !p.isNameToken(tok) {
				return nil, 0, p.unexpected(tok, "enum member")
			}
			nameTok := p.next()
			member := &enumMemberNode{baseNode: baseNode{start: nameTok.Line, end: nameTok.EndLine}, Doc: pendingDoc, Ann: pendingAnn, Name: nameTok.Value}
			if p.is("Equals") {
				p.next()
				switch {
				case p.is("StringLiteral"):
					t := p.next()
					v := unquote(t.Value)
					member.Value = &enumValueNode{Str: &v}
					member.end = t.EndLine
				case p.is("IntLiteral"):
					t := p.next()
					v := t.Value
					member.Value = &enumValueNode{Int: &v}
					member.end = t.EndLine
				default:
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
