package lsp

import (
	"context"
	"fmt"
	"time"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/ttcn3"
	"github.com/nokia/ntt/ttcn3/ast"
	"github.com/nokia/ntt/ttcn3/token"
)

type SemanticTokenType uint32

const (
	Namespace SemanticTokenType = iota
	/**
	 * Represents a generic type. Acts as a fallback for types which
	 * can't be mapped to a specific type like class or enum.
	 */
	Type
	Class
	Enum
	Interface
	Struct
	TypeParameter
	Parameter
	Variable
	Property
	EnumMember
	Event
	Function
	Method
	Macro
	Keyword
	Modifier
	Comment
	String
	Number
	Regexp
	Operator
	// Not part of the LSP
	None
)

var TokenTypes = []string{
	Namespace:     "namespace",
	Type:          "type",
	Class:         "class",
	Enum:          "enum",
	Interface:     "interface",
	Struct:        "struct",
	TypeParameter: "typeParameter",
	Parameter:     "parameter",
	Variable:      "variable",
	Property:      "property",
	EnumMember:    "enumMember",
	Event:         "event",
	Function:      "function",
	Method:        "method",
	Macro:         "macro",
	Keyword:       "keyword",
	Modifier:      "modifier",
	Comment:       "comment",
	String:        "string",
	Number:        "number",
	Regexp:        "regexp",
	Operator:      "operator",
}

type SemanticTokenModifiers uint32

const (
	Undefined   SemanticTokenModifiers = 0
	Declaration SemanticTokenModifiers = 1 << iota
	Definition
	Readonly
	Static
	Deprecated
	Abstract
	Async
	Modification
	Documentation
	DefaultLibrary
)

var TokenModifiers = []string{
	Undefined:      "undefined",
	Declaration:    "declaration",
	Definition:     "definition",
	Readonly:       "readonly",
	Static:         "static",
	Deprecated:     "deprecated",
	Abstract:       "abstract",
	Async:          "async",
	Modification:   "modification",
	Documentation:  "documentation",
	DefaultLibrary: "defaultLibrary",
}

var builtins = map[string]SemanticTokenType{
	"any2unistr":           Function,
	"anytype":              Type,
	"bit2hex":              Function,
	"bit2int":              Function,
	"bit2oct":              Function,
	"bit2str":              Function,
	"bitstring":            Type,
	"boolean":              Type,
	"char2int":             Function,
	"char2oct":             Function,
	"charstring":           Type,
	"decvalue":             Function,
	"decvalue_o":           Function,
	"decvalue_unichar":     Function,
	"default":              Type,
	"encvalue":             Function,
	"encvalue_o":           Function,
	"encvalue_unichar":     Function,
	"enum2int":             Function,
	"float":                Type,
	"float2int":            Function,
	"get_stringencoding":   Function,
	"hex2bit":              Function,
	"hex2int":              Function,
	"hex2oct":              Function,
	"hex2str":              Function,
	"hexstring":            Type,
	"hostid":               Function,
	"int2bit":              Function,
	"int2char":             Function,
	"int2enum":             Function,
	"int2float":            Function,
	"int2hex":              Function,
	"int2oct":              Function,
	"int2str":              Function,
	"int2unichar":          Function,
	"integer":              Type,
	"isbound":              Function,
	"ischosen":             Function,
	"ispresent":            Function,
	"istemplatekind":       Function,
	"isvalue":              Function,
	"lengthof":             Function,
	"log":                  Function,
	"match":                Function,
	"oct2bit":              Function,
	"oct2char":             Function,
	"oct2hex":              Function,
	"oct2int":              Function,
	"oct2str":              Function,
	"oct2unichar":          Function,
	"octetstring":          Type,
	"regexp":               Function,
	"remove_bom":           Function,
	"replace":              Function,
	"rnd":                  Function,
	"setverdict":           Function,
	"sizeof":               Function,
	"str2float":            Function,
	"str2hex":              Function,
	"str2int":              Function,
	"str2oct":              Function,
	"substr":               Function,
	"testcasename":         Function,
	"unichar2int":          Function,
	"unichar2oct":          Function,
	"universal charstring": Type,
	"verdicttype":          Type,
}

func (s *Server) semanticTokens(ctx context.Context, params *protocol.SemanticTokensParams) (*protocol.SemanticTokens, error) {
	file := string(params.TextDocument.URI)
	tree := ttcn3.ParseFile(file)
	begin := tree.Root.Pos()
	end := tree.Root.End()
	return s.semanticTokensRecover(tree, &s.db, begin, end)
}

func (s *Server) semanticTokensRange(ctx context.Context, params *protocol.SemanticTokensRangeParams) (*protocol.SemanticTokens, error) {
	file := string(params.TextDocument.URI)
	tree := ttcn3.ParseFile(file)
	begin := tree.Pos(int(params.Range.Start.Line)+1, int(params.Range.Start.Character+1))
	end := tree.Pos(int(params.Range.End.Line+1), int(params.Range.End.Character+1))
	return s.semanticTokensRecover(tree, &s.db, begin, end)
}

func (s *Server) semanticTokensRecover(tree *ttcn3.Tree, db *ttcn3.DB, begin loc.Pos, end loc.Pos) (*protocol.SemanticTokens, error) {
	start := time.Now()
	defer func() {
		// In case of a panic, just continue as this might be a common situation during typing
		if err := recover(); err != nil {
			log.Debugf("Recovered from panic: %v\n", err)
		}
		log.Debug(fmt.Sprintf("SemanticTokens for %s took %s.", tree.Filename(), time.Since(start)))
	}()

	return &protocol.SemanticTokens{Data: SemanticTokens(tree, db, begin, end)}, nil
}

func SemanticTokens(tree *ttcn3.Tree, db *ttcn3.DB, begin loc.Pos, end loc.Pos) []uint32 {
	var tokens []uint32
	line := 0
	col := 0
	ast.Inspect(tree.Root, func(n ast.Node) bool {
		if n == nil || n.End() < begin || end < n.Pos() {
			return false
		}

		if id, ok := n.(*ast.Ident); ok {

			// token type and modifiers
			var (
				typ SemanticTokenType
				mod SemanticTokenModifiers
			)
			if id.IsName {
				typ, mod = DefinitionToken(tree, id)
			} else {
				typ, mod = ReferenceToken(tree, db, id)
			}

			if typ != None {
				tokens, line, col = appendToken(tokens, id.Tok, tree, typ, mod, line, col)
				if id.Tok2.IsValid() {
					tokens, line, col = appendToken(tokens, id.Tok2, tree, typ, mod, line, col)
				}
			}
		}
		return true
	})
	log.Debugf("SemanticTokens: %d identifiers, %d bytes\n", len(tokens)/5, end-begin)
	return tokens
}

func appendToken(tokens []uint32, tok ast.Token, tree *ttcn3.Tree, typ SemanticTokenType, mod SemanticTokenModifiers, line int, col int) ([]uint32, int, int) {
	pos := tree.Position(tok.Pos())
	pos.Line -= 1
	pos.Column -= 1

	// relative line
	relLine := pos.Line - line
	if relLine != 0 {
		col = 0
	}
	line = pos.Line

	// relative column
	relCol := pos.Column - col
	col = pos.Column

	// token width
	width := int(tok.End() - tok.Pos())
	tokens = append(tokens, uint32(relLine), uint32(relCol), uint32(width), uint32(typ), uint32(mod))
	return tokens, line, col
}

func DefinitionToken(tree *ttcn3.Tree, id ast.Node) (SemanticTokenType, SemanticTokenModifiers) {
	switch n := tree.ParentOf(id).(type) {
	case *ast.Module:
		return Namespace, Definition
	case *ast.ImportDecl:
		return Namespace, Undefined
	case *ast.FormalPar:
		return Parameter, Declaration
	case *ast.StructTypeDecl:
		return Struct, Definition
	case *ast.EnumTypeDecl:
		if id == n.Name {
			return Enum, Definition
		}
		return EnumMember, Definition
	case *ast.EnumSpec:
		return EnumMember, Definition
	case *ast.Field:
		if _, ok := tree.ParentOf(n).(*ast.SubTypeDecl); ok {
			return Type, Definition
		}
	case *ast.PortTypeDecl:
		return Interface, Definition
	case *ast.ComponentTypeDecl:
		return Class, Definition
	case *ast.FuncDecl:
		return Function, Definition
	case *ast.TemplateDecl:
		return Variable, Declaration | Readonly
	case *ast.ValueDecl:
		typ := Variable
		mod := Declaration
		if _, ok := tree.ParentOf(tree.ParentOf(tree.ParentOf(n))).(*ast.ComponentTypeDecl); ok {
			typ = Property
		}
		switch n.Kind.Kind {
		case token.CONST, token.MODULEPAR:
			mod |= Readonly
		case token.PORT:
			typ = Interface
		}
		return typ, mod
	case *ast.Declarator:
		return DefinitionToken(tree, n)
	}
	return None, Undefined
}

func ReferenceToken(tree *ttcn3.Tree, db *ttcn3.DB, id *ast.Ident) (SemanticTokenType, SemanticTokenModifiers) {
	name := id.String()
	if typ, ok := builtins[name]; ok {
		return typ, DefaultLibrary
	}
	defs := tree.LookupWithDB(id, db)
	if len(defs) == 0 {
		return None, Undefined
	}
	if len(defs) > 1 {
		log.Debugf("ReferenceToken: multiple definitions for %s\n", name)
	}
	typ, mod := DefinitionToken(defs[0].Tree, defs[0].Ident)
	return typ, mod &^ (Definition | Declaration)
}
