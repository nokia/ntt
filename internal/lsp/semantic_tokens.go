package lsp

import (
	"context"
	"fmt"
	"time"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/internal/ntt"
	"github.com/nokia/ntt/ttcn3"
	"github.com/nokia/ntt/ttcn3/ast"
	"github.com/nokia/ntt/ttcn3/scanner"
	"github.com/nokia/ntt/ttcn3/token"
)

type SemanticTokenType uint32
type SemanticTokenModifiers uint32

type TokenGen struct {
	PrevLine   uint32
	PrevColumn uint32
}

type SemTokVisitor struct {
	tree        *ttcn3.Tree
	db          *ttcn3.DB
	actualToken SemanticTokenType
	actualModif SemanticTokenModifiers
	Data        []uint32
	tg          TokenGen
	startOffs   loc.Pos
	endOffs     loc.Pos
	nodeStack   []ast.Node
}

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

const (
	Declaration SemanticTokenModifiers = 1 << iota
	Definition
	Readonly
	static
	deprecated
	abstract
	async
	modification
	documentation
	DefaultLibrary
	Undefined = 0
)

// decide on emitting keywords into the semantic token list
var EmitKeywords bool = false

var predefTypeMap map[string]bool = map[string]bool{
	"anytype": true, "bitstring": true, "boolean": true, "charstring": true, "default": true, "float": true, "hexstring": true,
	"integer": true, "octetstring": true, "universal charstring": true, "verdicttype": true}

var libraryFuncMap map[string]bool = map[string]bool{
	"int2char": true, "int2unichar": true, "int2bit": true, "int2enum": true, "int2hex": true, "int2oct": true, "int2str": true, "int2float": true,
	"float2int": true, "char2int": true, "char2oct": true, "unichar2int": true, "unichar2oct": true, "bit2int": true, "bit2hex": true,
	"bit2oct": true, "bit2str": true, "hex2int": true, "hex2bit": true, "hex2oct": true, "hex2str": true, "oct2int": true, "oct2bit": true,
	"oct2hex": true, "oct2str": true, "oct2char": true, "oct2unichar": true, "str2int": true, "str2hex": true, "str2oct": true, "str2float": true,
	"enum2int": true, "any2unistr": true, "lengthof": true, "sizeof": true, "ispresent": true, "ischosen": true, "isvalue": true, "isbound": true,
	"istemplatekind": true, "regexp": true, "substr": true, "replace": true, "encvalue": true, "decvalue": true, "encvalue_unichar": true,
	"decvalue_unichar": true, "encvalue_o": true, "decvalue_o": true, "get_stringencoding": true, "remove_bom": true, "rnd": true,
	"testcasename": true, "hostid": true, "match": true, "setverdict": true, "log": true}

var tokenTypes = []string{
	"namespace", "type", "class", "enum", "interface", "struct", "typeParameter", "parameter", "variable", "property", "enumMember",
	"event", "function", "method", "macro", "keyword", "modifier", "comment", "string", "number", "regexp", "operator"}

var tokenModifiers = []string{
	"declaration", "definition", "readonly", "static", "deprecated", "abstract", "async", "modification", "documentation", "defaultLibrary"}

func (tg *TokenGen) NewTuple(line uint32, column uint32, length uint32, tokenType SemanticTokenType, modifier uint32) []uint32 {
	var res_line, res_column uint32
	if tg.PrevLine == line {
		res_column = column - tg.PrevColumn
		res_line = 0

	} else {
		res_line = line - tg.PrevLine
		tg.PrevLine = line
		res_column = column
	}
	tg.PrevColumn = column
	return []uint32{res_line, res_column, uint32(length), uint32(tokenType), modifier}
}

func isPosGt(line1 uint32, col1 uint32, line2 uint32, col2 uint32) bool {
	if line1 > line2 {
		return true
	}
	if line1 == line2 {
		return col1 > col2
	}
	return false
}

func mergeSortTokenarrays(toka1 []uint32, toka2 []uint32) []uint32 {
	const recordLen int = 5
	if len(toka1) == 0 {
		return toka2
	}
	if len(toka2) == 0 {
		return toka1
	}
	res := make([]uint32, 0, len(toka1)+len(toka2))
	linet1 := toka1[0]
	colt1 := toka1[1]
	tokGen := TokenGen{}
	linet2 := toka2[0]
	colt2 := toka2[1]
	for i, j := 0, 0; ; {
		if isPosGt(linet1, colt1, linet2, colt2) {
			res = append(res, tokGen.NewTuple(linet2, colt2, toka2[j+2], SemanticTokenType(toka2[j+3]), toka2[j+4])...)
			j += recordLen
			if j >= len(toka2) {
				if i < len(toka1) {
					res = append(res, tokGen.NewTuple(linet1, colt1, toka1[i+2], SemanticTokenType(toka1[i+3]), toka1[i+4])...)
					res = append(res, toka1[i+recordLen:]...)
				}
				return res
			}
			linet2 += toka2[j]
			if toka2[j] == 0 {
				colt2 += toka2[j+1]
			} else {
				colt2 = toka2[j+1]
			}
		} else {
			res = append(res, tokGen.NewTuple(linet1, colt1, toka1[i+2], SemanticTokenType(toka1[i+3]), toka1[i+4])...)
			i += recordLen
			if i >= len(toka1) {
				if j < len(toka2) {
					res = append(res, tokGen.NewTuple(linet2, colt2, toka2[j+2], SemanticTokenType(toka2[j+3]), toka2[j+4])...)
					res = append(res, toka2[j+recordLen:]...)
				}
				return res
			}
			linet1 += toka1[i]
			if toka1[i] == 0 {
				colt1 += toka1[i+1]
			} else {
				colt1 = toka1[i+1]
			}
		}
	}
}

func NewSyntaxTokensFromCurrentModule(file string, txtRange protocol.Range) []uint32 {
	scn := &scanner.Scanner{}
	f := fs.Open(file)
	b, _ := f.Bytes()
	fs := loc.NewFileSet()
	locF := fs.AddFile(f.Path(), -1, len(b))
	scn.Init(locF, b, nil)

	d := make([]uint32, 0, 20)
	var tg TokenGen
	prevTok := token.ILLEGAL
	for {
		pos, tok, lit := scn.Scan()
		if (uint32(fs.Position(pos).Line) < txtRange.Start.Line+1) ||
			(uint32(fs.Position(pos).Line) == txtRange.Start.Line+1) &&
				(uint32(fs.Position(pos).Column) < txtRange.Start.Character+1) {
			continue
		}
		if (uint32(fs.Position(pos).Line) > txtRange.End.Line+1) ||
			(uint32(fs.Position(pos).Line) == txtRange.End.Line+1) &&
				(uint32(fs.Position(pos).Column) > txtRange.End.Character+1) {
			break
		}

		if tok == token.EOF {
			break
		}
		if tok.IsKeyword() && (tok != token.UNIVERSAL && !(prevTok == token.UNIVERSAL && tok == token.CHARSTRING)) {
			if tok != token.CHARSTRING {
				line := uint32(fs.Position(pos).Line - 1)
				column := uint32(fs.Position(pos).Column - 1)
				d = append(d, tg.NewTuple(line, column, uint32(len(lit)), Keyword, 0)...)
			}
		}
		prevTok = tok
	}
	return d
}

func (tokv *SemTokVisitor) pushNodeStack(n ast.Node) {
	tokv.nodeStack = append(tokv.nodeStack, n)
}

func (tokv *SemTokVisitor) popNodeStack() {
	tokv.nodeStack = tokv.nodeStack[:len(tokv.nodeStack)-1]
}

func (tokv *SemTokVisitor) VisitModuleDefs(n ast.Node) bool {
	if n == nil {
		tokv.popNodeStack()
		return false
	}
	if tokv.endOffs < n.Pos() {
		return false // definitely stop recursing
	}
	tokv.pushNodeStack(n)
	if tokv.startOffs > n.Pos() {
		return true
	}

	begin := tokv.tree.Position(n.Pos())
	end := tokv.tree.Position(n.LastTok().End())
	switch node := n.(type) {
	case *ast.Module:
		if node.Name != nil {
			begin = tokv.tree.Position(node.Name.Pos())
			end = tokv.tree.Position(node.Name.End())
			tokv.Data = append(tokv.Data, tokv.tg.NewTuple(uint32(begin.Line-1), uint32(begin.Column-1), uint32(end.Offset-begin.Offset), Namespace, uint32(Definition))...)
		}
		for _, def := range node.Defs {
			ast.Inspect(def, tokv.VisitModuleDefs)
		}
		tokv.popNodeStack()
		return false
	case *ast.FormalPar:
		if node.Type != nil {
			ast.Inspect(node.Type, tokv.VisitModuleDefs)
		}
		if node.Name != nil {
			begin := tokv.tree.Position(node.Name.Pos())
			end := tokv.tree.Position(node.Name.End())
			tokv.Data = append(tokv.Data, tokv.tg.NewTuple(uint32(begin.Line-1), uint32(begin.Column-1), uint32(end.Offset-begin.Offset), Parameter, uint32(Declaration))...)
		}
		tokv.popNodeStack()
		return false
	case *ast.StructTypeDecl:
		if node.Name != nil {
			begin = tokv.tree.Position(node.Name.Pos())
			end = tokv.tree.Position(node.Name.End())
			tokv.Data = append(tokv.Data, tokv.tg.NewTuple(uint32(begin.Line-1), uint32(begin.Column-1), uint32(end.Offset-begin.Offset), Struct, uint32(Definition))...)
			for _, field := range node.Fields {
				ast.Inspect(field, tokv.VisitModuleDefs)
			}
			tokv.popNodeStack()
			return false
		}
	case *ast.Field:

		isSubType := false
		for i := range tokv.nodeStack {
			if _, ok := tokv.nodeStack[len(tokv.nodeStack)-i-1].(*ast.SubTypeDecl); ok {
				isSubType = true
				break
			}
		}
		if node.Type != nil {
			//tokv.actualModif = Undefined
			ast.Inspect(node.Type, tokv.VisitModuleDefs)
		}
		if isSubType {
			if node.Name != nil {
				begin = tokv.tree.Position(node.Name.Pos())
				end = tokv.tree.Position(node.Name.End())
				tokv.Data = append(tokv.Data, tokv.tg.NewTuple(uint32(begin.Line-1), uint32(begin.Column-1), uint32(end.Offset-begin.Offset), Type, uint32(Definition))...)
			}
		}

		tokv.popNodeStack()
		return false
	case *ast.PortTypeDecl:
		if node.Name != nil {
			begin = tokv.tree.Position(node.Name.Pos())
			end = tokv.tree.Position(node.Name.End())
			tokv.Data = append(tokv.Data, tokv.tg.NewTuple(uint32(begin.Line-1), uint32(begin.Column-1), uint32(end.Offset-begin.Offset), Interface, uint32(Definition))...)
			for _, attr := range node.Attrs {
				ast.Inspect(attr, tokv.VisitModuleDefs)
			}
		}
		tokv.popNodeStack()
		return false

	case *ast.ComponentTypeDecl:
		if node.Name != nil {
			begin = tokv.tree.Position(node.Name.Pos())
			end = tokv.tree.Position(node.Name.End())
			tokv.Data = append(tokv.Data, tokv.tg.NewTuple(uint32(begin.Line-1), uint32(begin.Column-1), uint32(end.Offset-begin.Offset), Class, uint32(Definition))...)
			if node.Body != nil {
				ast.Inspect(node.Body, tokv.VisitModuleDefs)
			}
		}
		return false
	case *ast.FuncDecl:
		if node.Name != nil {
			begin = tokv.tree.Position(node.Name.Pos())
			end = tokv.tree.Position(node.Name.End())
			tokv.Data = append(tokv.Data, tokv.tg.NewTuple(uint32(begin.Line-1), uint32(begin.Column-1), uint32(end.Offset-begin.Offset), Function, uint32(Definition))...)
		}
		if node.Params != nil {
			ast.Inspect(node.Params, tokv.VisitModuleDefs)
		}
		if node.Return != nil {
			ast.Inspect(node.Return, tokv.VisitModuleDefs)
		}
		if node.Body != nil {
			ast.Inspect(node.Body, tokv.VisitModuleDefs)
		}
		tokv.popNodeStack()
		return false
	case *ast.TemplateDecl:
		if node.Type != nil {
			tokv.actualToken = Type
			ast.Inspect(node.Type, tokv.VisitModuleDefs)
			tokv.actualToken = None
		}
		if node.Name != nil {
			begin = tokv.tree.Position(node.Name.Pos())
			end = tokv.tree.Position(node.Name.End())
			tokv.Data = append(tokv.Data, tokv.tg.NewTuple(uint32(begin.Line-1), uint32(begin.Column-1), uint32(end.Offset-begin.Offset), Variable, uint32(Declaration|Readonly))...)
		}
		if node.Params != nil {
			ast.Inspect(node.Params, tokv.VisitModuleDefs)
		}
		if node.Value != nil {
			ast.Inspect(node.Value, tokv.VisitModuleDefs)
		}
		tokv.popNodeStack()
		return false
	case *ast.ValueDecl:
		ast.Inspect(node.Type, tokv.VisitModuleDefs)
		tokv.actualToken = Variable
		tokv.actualModif = Declaration
		if node.Kind.Kind == token.CONST {
			tokv.actualModif |= Readonly
		}
		for _, decl := range node.Decls {
			ast.Inspect(decl, tokv.VisitModuleDefs)
		}
		tokv.actualToken = None
		tokv.actualModif = Undefined
		tokv.popNodeStack()
		return false
	case *ast.Declarator:
		if node.Name != nil {
			begin = tokv.tree.Position(node.Name.Pos())
			end = tokv.tree.Position(node.Name.End())
			tokv.Data = append(tokv.Data, tokv.tg.NewTuple(uint32(begin.Line-1), uint32(begin.Column-1), uint32(end.Offset-begin.Offset), tokv.actualToken, uint32(tokv.actualModif))...)
		}
		tokv.actualToken = None
		tokv.actualModif = Undefined
		if node.Value != nil {
			ast.Inspect(node.Value, tokv.VisitModuleDefs)
		}
		tokv.popNodeStack()
		return false
	case *ast.Ident:
		begin := tokv.tree.Position(node.Pos())
		end := tokv.tree.Position(node.End())
		switch node.Tok.Kind {
		case token.IDENT:
			if _, ok := predefTypeMap[node.String()]; ok {
				tokv.Data = append(tokv.Data, tokv.tg.NewTuple(uint32(begin.Line-1), uint32(begin.Column-1), uint32(end.Offset-begin.Offset), Type, uint32(DefaultLibrary))...)
			} else if _, ok := libraryFuncMap[node.String()]; ok {
				tokv.Data = append(tokv.Data, tokv.tg.NewTuple(uint32(begin.Line-1), uint32(begin.Column-1), uint32(end.Offset-begin.Offset), Function, uint32(DefaultLibrary))...)
				/*} else if _, ok := (*tokv.modNameSet)[node.String()]; ok {
				tokv.Data = append(tokv.Data, tokv.tg.NewTuple(uint32(begin.Line-1), uint32(begin.Column-1), uint32(end.Offset-begin.Offset), Namespace, 0)...)
				*/
			} else if def := tokv.tree.LookupWithDB(node, tokv.db); len(def) > 0 {
				switch defNode := def[0].Node.(type) {
				case *ast.StructTypeDecl, *ast.Field:
					tokv.Data = append(tokv.Data, tokv.tg.NewTuple(uint32(begin.Line-1), uint32(begin.Column-1), uint32(end.Offset-begin.Offset), Type, 0)...)
				case *ast.ImportDecl:
					tokv.Data = append(tokv.Data, tokv.tg.NewTuple(uint32(begin.Line-1), uint32(begin.Column-1), uint32(end.Offset-begin.Offset), Namespace, 0)...)
				case *ast.FormalPar:
					tokv.Data = append(tokv.Data, tokv.tg.NewTuple(uint32(begin.Line-1), uint32(begin.Column-1), uint32(end.Offset-begin.Offset), Parameter, 0)...)
				case *ast.FuncDecl:
					tokv.Data = append(tokv.Data, tokv.tg.NewTuple(uint32(begin.Line-1), uint32(begin.Column-1), uint32(end.Offset-begin.Offset), Function, 0)...)
				case *ast.ValueDecl:
					var modif SemanticTokenModifiers = Undefined
					if (defNode.Kind.Kind == token.CONST) || (defNode.Kind.Kind == token.TEMPLATE) {
						modif = Readonly
					}
					tokv.Data = append(tokv.Data, tokv.tg.NewTuple(uint32(begin.Line-1), uint32(begin.Column-1), uint32(end.Offset-begin.Offset), Variable, uint32(modif))...)
				}

			} else if tokv.actualToken != None {
				tokv.Data = append(tokv.Data, tokv.tg.NewTuple(uint32(begin.Line-1), uint32(begin.Column-1), uint32(end.Offset-begin.Offset), tokv.actualToken, uint32(tokv.actualModif))...)
			}
		case token.UNIVERSAL:
			if node.Tok2.Kind == token.CHARSTRING {
				tokv.Data = append(tokv.Data, tokv.tg.NewTuple(uint32(begin.Line-1), uint32(begin.Column-1), uint32(end.Offset-begin.Offset), Type, uint32(DefaultLibrary))...)
			}
		case token.CHARSTRING:
			tokv.Data = append(tokv.Data, tokv.tg.NewTuple(uint32(begin.Line-1), uint32(begin.Column-1), uint32(end.Offset-begin.Offset), Type, uint32(DefaultLibrary))...)
		}
		tokv.popNodeStack()
		return false
	}
	return true
}

func NewSemTokVisitor(tree *ttcn3.Tree, db *ttcn3.DB, txtRange protocol.Range) *SemTokVisitor {
	return &SemTokVisitor{
		tree:        tree,
		db:          db,
		actualToken: None,
		actualModif: Undefined,
		Data:        make([]uint32, 0, 20),
		startOffs:   tree.Pos(int(txtRange.Start.Line+1), int(txtRange.Start.Character+1)),
		endOffs:     tree.Pos(int(txtRange.End.Line+1), int(txtRange.End.Character+1)),
		tg:          TokenGen{},
		nodeStack:   make([]ast.Node, 0, 10)}
}

func NewSemanticTokensFromCurrentModule(tree *ttcn3.Tree, db *ttcn3.DB, suite *ntt.Suite, fileName string, txtRange protocol.Range) *protocol.SemanticTokens {
	stVisitor := NewSemTokVisitor(tree, db, txtRange)

	for _, mod := range tree.Modules() {
		ast.Inspect(mod, stVisitor.VisitModuleDefs)
		if EmitKeywords {
			stVisitor.Data = mergeSortTokenarrays(NewSyntaxTokensFromCurrentModule(fileName, txtRange), stVisitor.Data)
		}
	}
	stVisitor.Data[0] -= txtRange.Start.Line
	stVisitor.Data[1] -= txtRange.Start.Character
	ret := &protocol.SemanticTokens{Data: stVisitor.Data}
	return ret
}

func (s *Server) semanticTokens(ctx context.Context, params *protocol.SemanticTokensParams) (*protocol.SemanticTokens, error) {
	var ret *protocol.SemanticTokens = nil
	start := time.Now()
	defer func() {
		if err := recover(); err != nil {
			// in case of a panic, just continue as this might be a common situation during typing
			ret = nil
			log.Debug(fmt.Sprintf("Info: %s.", err))
		}
	}()
	defer func() {
		elapsed := time.Since(start)
		log.Debug(fmt.Sprintf("SemanticTokens took %s.", elapsed))
	}()

	suites := s.Owners(params.TextDocument.URI)
	// a completely new and empty file belongs to no suites at all
	if len(suites) == 0 {
		return nil, nil
	}

	tree := ttcn3.ParseFile(params.TextDocument.URI.SpanURI().Filename())
	mods := tree.Modules()

	modEnd := tree.Position(mods[len(mods)-1].End())
	txtRange := protocol.Range{Start: protocol.Position{Line: 0, Character: 0}, End: protocol.Position{Line: uint32(modEnd.Line - 1), Character: uint32(modEnd.Column - 1)}}
	ret = NewSemanticTokensFromCurrentModule(tree, &s.db, suites[0], params.TextDocument.URI.SpanURI().Filename(), txtRange)
	return ret, nil
}

func (s *Server) semanticTokensRange(ctx context.Context, params *protocol.SemanticTokensRangeParams) (*protocol.SemanticTokens, error) {
	var ret *protocol.SemanticTokens = nil
	start := time.Now()
	defer func() {
		if err := recover(); err != nil {
			// in case of a panic, just continue as this might be a common situation during typing
			ret = nil
			log.Debug(fmt.Sprintf("Info: %s.", err))
		}
	}()
	defer func() {
		elapsed := time.Since(start)
		log.Debug(fmt.Sprintf("SemanticTokensRange took %s.", elapsed))
	}()

	suites := s.Owners(params.TextDocument.URI)
	// a completely new and empty file belongs to no suites at all
	if len(suites) == 0 {
		return nil, nil
	}

	tree := ttcn3.ParseFile(params.TextDocument.URI.SpanURI().Filename())
	ret = NewSemanticTokensFromCurrentModule(tree, &s.db, suites[0], params.TextDocument.URI.SpanURI().Filename(), params.Range)
	return ret, nil
}
