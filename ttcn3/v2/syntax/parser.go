package syntax

import (
	"fmt"

	"github.com/nokia/ntt/internal/log"
)

// Parse parses the TTCN-3 source code in src and returns a syntax tree.
func Parse(src []byte) Node {
	p := &parser{Scanner: NewScanner(src)}
	p.content = src
	root := p.Push(Root)
	p.peek(1)
	p.next = p.tokens[p.pos]
	for p.next.Kind() != EOF {
		p.parseTTCN3()
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

func (p *parser) accept(kk ...Kind) bool {
	if len(kk) > 3 {
		log.Trace("warning: accept called with more than 3 arguments")
	}
	for _, k := range kk {
		if p.next.Kind() == k {
			return true
		}
	}
	return false
}

func (p *parser) expect(k Kind) bool {
	if p.next.Kind() == k {
		p.consume()
		return true
	} else {
		p.error(fmt.Errorf("expected %s, got %s", k, p.next.Kind()))
		return false
	}
}

func (p *parser) parseDecl() bool { return p.parseTTCN3() }
func (p *parser) parseStmt() bool { return p.parseTTCN3() }

// parseTTCN3 parses any main TTCN-3 construct: modules, declarations,
// statements and expressions.
func (p *parser) parseTTCN3() bool {
	switch tok := p.next.Kind(); tok {

	case AltKeyword, InterleaveKeyword:
		return p.parseAltStmt()
	case LeftBrace:
		return p.parseBlock()
	case DoKeyword:
		return p.parseDoStmt()
	case ForKeyword:
		return p.parseForStmt()
	case GotoKeyword:
		return p.parseGotoStmt()
	case LeftBracket:
		return p.parseGuardStmt()
	case IfKeyword:
		return p.parseIfStmt()
	case LabelKeyword:
		return p.parseLabelStmt()
	case PortKeyword:
		return p.parsePortDecl()
	case ReturnKeyword:
		return p.parseReturnStmt()
	case SelectKeyword:
		return p.parseSelectStmt()
	case WhileKeyword:
		return p.parseWhileStmt()
	case AltstepKeyword:
		return p.parseAltstep()
	case ConfigurationKeyword:
		return p.parseConfiguration()
	case CreateKeyword:
		return p.parseConstructor()
	case ControlKeyword:
		return p.parseControl()
	case FinallyKeyword:
		return p.parseDestructor()
	case FriendKeyword:
		return p.parseFriend()
	case ExternalKeyword:
		return p.parseFunction()
	case GroupKeyword:
		return p.parseGroup()
	case ImportKeyword:
		return p.parseImport()
	case SignatureKeyword:
		return p.parseSignature()
	case TemplateKeyword:
		return p.parseTemplate()
	case TestcaseKeyword:
		// Resolve conflict between `testcase.stop` and a testcase definition.
		if p.peek(2).Kind() == Dot {
			return p.parseExpr()
		}
		return p.parseTestcase()
	case TypeKeyword:
		switch p.peek(2).Kind() {
		case AltstepKeyword:
			return p.parseAltstepType()
		case ClassKeyword:
			return p.parseClass()
		case ComponentKeyword:
			return p.parseComponent()
		case EnumeratedKeyword:
			return p.parseEnum()
		case FunctionKeyword:
			return p.parseFunctionType()
		case MapKeyword:
			return p.parseMap()
		case PortKeyword:
			return p.parsePort()
		case TestcaseKeyword:
			return p.parseTestcaseType()
		case RecordKeyword, SetKeyword, UnionKeyword:
			switch p.peek(3).Kind() {
			case LengthKeyword, OfKeyword:
				return p.parseList()
			default:
				return p.parseStruct()
			}
		default:
			return p.parseSubType()
		}
	case ConstKeyword, ModuleparKeyword, VarKeyword:
		return p.parseVarDecl()

	case ModuleKeyword:
		return p.parseModule()

	case TimerKeyword:
		// There's a conflict between `timer.stop` (expression) and
		// `timer t` (declaration). We resolve it by looking ahead, if
		// the next token is an identifier, we assume it's a
		// declaration. Everything else is parsed as expression.
		if p.peek(2).Kind() == Identifier {
			return p.parseVarDecl()
		}
		return p.parseExpr()

	default:
		return p.parseExpr()
	}
}

func (p *parser) parseImportStmt() bool {
	// Conflict with Refs and "all except Refs"
	return false
}

// parseRef has conflicts with various `all` and `any` references.
func (p *parser) parseRef() bool {

	switch p.next.Kind() {
	case AddressKeyword:
		p.expect(AddressKeyword)
	case AllKeyword:
		p.expect(AllKeyword)
		switch p.next.Kind() {
		case ComponentKeyword:
			return p.expect(ComponentKeyword)
		case PortKeyword:
			return p.expect(PortKeyword)
		case TimerKeyword:
			return p.expect(TimerKeyword)
		}
		return true
	case AnyKeyword:
		p.expect(AnyKeyword)
		switch p.next.Kind() {
		case ComponentKeyword:
			return p.expect(ComponentKeyword)
		case PortKeyword:
			return p.expect(PortKeyword)
		case TimerKeyword:
			return p.expect(TimerKeyword)
		}
	case MapKeyword:
		p.expect(MapKeyword)
	case MtcKeyword:
		p.expect(MtcKeyword)
	case SelfKeyword:
		p.expect(SelfKeyword)
	case SystemKeyword:
		p.expect(SystemKeyword)
	case ThisKeyword:
		p.expect(ThisKeyword)
	case TimerKeyword:
		p.expect(TimerKeyword)
	case UniversalKeyword:
		p.expect(UniversalKeyword)
		p.expect(CharstringKeyword)
	case UnmapKeyword:
		p.expect(UnmapKeyword)
	case Identifier:
		// TODO(5nord): Evaluate if its faster to handle identifiers before this switch.
		p.expect(Identifier)
		if p.accept(Less) {
			return p.parseTypePars()
		}
		return true
	}
	p.error(fmt.Errorf("unexpected token %v", p.next))
	return false
}

func (p *parser) parseNestedType() bool {
	switch p.next.Kind() {
	case AddressKeyword, AllKeyword, AnyKeyword, Identifier, MapKeyword, MtcKeyword, SelfKeyword, SystemKeyword, ThisKeyword, TimerKeyword, UniversalKeyword, UnmapKeyword:
		return p.parseRef()
	case RecordKeyword, SetKeyword, UnionKeyword:
		if p.peek(2).Kind() == LeftBrace {
			return p.parseNestedStruct()
		} else {
			return p.parseNestedList()
		}
	case EnumeratedKeyword:
		return p.parseNestedEnum()
	}
	p.error(fmt.Errorf("unexpected token %v", p.next))
	return false
}

func (p *parser) parseExprList() bool {
	p.parseExpr()
	for p.next.Kind() == Comma {
		p.consume()
		p.parseExpr()
	}
	return false
}

func (p *parser) parseExpr() bool {
	return false
}

func (p *parser) parseBinaryExpr() bool { panic("binaryExpr: not implemented") }
func (p *parser) parseUnaryExpr() bool  { panic("binaryExpr: not implemented") }
func (p *parser) parseCallExpr() bool   { panic("binaryExpr: not implemented") }
func (p *parser) parseIndexExpr() bool  { panic("binaryExpr: not implemented") }
func (p *parser) parseDotExpr() bool    { panic("binaryExpr: not implemented") }
