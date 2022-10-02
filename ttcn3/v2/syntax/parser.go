package syntax

// Parse parses the TTCN-3 source code in src and returns a syntax tree.
func Parse(src []byte) Node {
	p := &parser{Scanner: NewScanner(src)}
	p.content = src
	root := p.Push(Root)
	p.peek(1)
	p.next = p.tokens[p.pos]
	for p.next.Kind() != EOF {
		p.parse()
	}
	p.Pop()
	root.tree.lines = p.Lines()
	return root
}

var (
	literals = map[Kind]bool{
		Any:           true,
		Bitstring:     true,
		ErrorLiteral:  true,
		FailLiteral:   true,
		FalseLiteral:  true,
		Float:         true,
		InconcLiteral: true,
		Integer:       true,
		Mul:           true,
		NotANumber:    true,
		NoneLiteral:   true,
		NullKeyword:   true,
		OmitKeyword:   true,
		PassLiteral:   true,
		String:        true,
		Sub:           true,
		TrueLiteral:   true,
	}

	// refs is a list of keywords that are allowed to be used as references.
	refs = map[Kind]bool{
		AddressKeyword:    true,
		AllKeyword:        true,
		AnyKeyword:        true,
		CharstringKeyword: true,
		MapKeyword:        true,
		MtcKeyword:        true,
		SystemKeyword:     true,
		TestcaseKeyword:   true,
		TimerKeyword:      true,
		UniversalKeyword:  true,
		UnmapKeyword:      true,
	}
)

type parser struct {
	Builder
	*Scanner

	next   treeEvent
	tokens []treeEvent
	pos    int
}

func (p *parser) scan() treeEvent {
	for {
		kind, begin, end := p.Scan()
		switch kind {
		case Comment, Preproc:
		case Identifier:
			// The scanner does not distinguish between identifiers and keywords.
			// Therefore the parser must check if the identifier is a keyword.
			if kw, ok := keywords[string(p.content[begin:end])]; ok {
				kind = kw
			}
			fallthrough
		default:
			return newAddToken(kind, begin, end)
		}
	}
}

func (p *parser) peek(i int) treeEvent {
	idx := p.pos + i - 1
	last := len(p.tokens) - 1
	if idx > last {
		n := idx - last
		for i := 0; i < n; i++ {
			p.tokens = append(p.tokens, p.scan())
		}
	}
	return p.tokens[idx]
}

func (p *parser) error(err error) {
	// TODO(5nord) Use NodeErrors with ranges instead of error interface. This makes it easier
	// to assign errors to subtrees.
	p.tree.errs = append(p.tree.errs, err)
}

// consume returns the next token, skipping past any comments and preprocessor directives.
func (p *parser) consume() treeEvent {
	tok := p.tokens[p.pos]
	p.pos++
	if p.pos == len(p.tokens) {
		p.pos = 0
		p.tokens = p.tokens[:0]
	}
	p.peek(1)
	p.next = p.tokens[p.pos]
	return tok
}

// parse parses anything which is TTCN-3.
func (p *parser) parse() {
	switch p.next.Kind() {
	case AltKeyword, InterleaveKeyword:
		p.parseAltStmt()
	case AltstepKeyword:
		p.parseAltstep()
	case ConfigurationKeyword:
		p.parseConfiguration()
	case ConstKeyword:
		p.parseVarDecl()
	case DoKeyword:
		p.parseDoStmt()
	case ExternalKeyword:
		p.parseFunction()
	case ForKeyword:
		p.parseForStmt()
	case FriendKeyword:
		p.parseFriend()
	case GotoKeyword:
		p.parseGotoStmt()
	case GroupKeyword:
		p.parseGroup()
	case IfKeyword:
		p.parseIfStmt()
	case ImportKeyword:
		p.parseImport()
	case LabelKeyword:
		p.parseLabelStmt()
	case ModuleKeyword:
		p.parseModule()
	case ModuleparKeyword:
		p.parseVarDecl()
	case ReturnKeyword:
		p.parseReturn()
	case SelectKeyword:
		p.parseSelectStmt()
	case SignatureKeyword:
		p.parseSignature()
	case TemplateKeyword:
		p.parseTemplate()
	case TestcaseKeyword:
		p.parseTestcase()
	case VarKeyword:
		p.parseVarDecl()
	case WhileKeyword:
		p.parseWhileStmt()
	case LeftBrace:
		p.parseBlock()
	case LeftBracket:
		p.parseGuardStmt()
	case Identifier:
		p.parseStmt()
	default:
		if refs[p.next.Kind()] {
			p.parseStmt()
		} else {
			p.parseExpr()
		}
	}
}

func (p *parser) parseExprList() {
	p.parseExpr()
	for p.next.Kind() == Comma {
		p.consume()
		p.parseExpr()
	}
}

func (p *parser) parseExpr() {
}

func (p *parser) parseBinaryExpr() { panic("binaryExpr: not implemented") }
func (p *parser) parseUnaryExpr()  { panic("binaryExpr: not implemented") }
func (p *parser) parseCallExpr()   { panic("binaryExpr: not implemented") }
func (p *parser) parseIndexExpr()  { panic("binaryExpr: not implemented") }
func (p *parser) parseDotExpr()    { panic("binaryExpr: not implemented") }
