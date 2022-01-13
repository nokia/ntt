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
	"github.com/nokia/ntt/ttcn3/ast"
	"github.com/nokia/ntt/ttcn3/scanner"
	"github.com/nokia/ntt/ttcn3/token"
)

type SemanticTokenType uint32
type SemanticTokenModifiers uint8

type TokenGen struct {
	PrevLine   uint32
	PrevColumn uint32
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
)

const (
	declaration SemanticTokenModifiers = iota
	definition
	readonly
	static
	deprecated
	abstract
	async
	modification
	documentation
	defaultLibrary
)

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

func modifierCalc(args ...SemanticTokenModifiers) uint32 {
	var ret uint32 = 0
	for _, arg := range args {
		ret |= 1 << uint32(arg)
	}
	return ret
}

func NewSyntaxTokensFromCurrentModule(file string) []uint32 {
	scn := &scanner.Scanner{}
	f := fs.Open(file)
	b, _ := f.Bytes()
	fs := loc.NewFileSet()

	scn.Init(fs.AddFile(f.Path(), -1, len(b)), b, nil)
	d := make([]uint32, 0, 20)
	var tg TokenGen
	for {
		pos, tok, lit := scn.Scan()
		if tok == token.EOF {
			break
		}
		if tok.IsKeyword() {
			line := uint32(fs.Position(pos).Line - 1)
			column := uint32(fs.Position(pos).Column - 1)
			d = append(d, tg.NewTuple(line, column, uint32(len(lit)), Keyword, modifierCalc())...)
		}

	}
	return d
}

func NewSemanticTokensFromCurrentModule(syntax *ntt.ParseInfo, fileName string) *protocol.SemanticTokens {
	var tg TokenGen
	d := make([]uint32, 0, 20)
	nodeStack := make([]ast.Node, 0, 10)

	ast.Inspect(syntax.Module, func(n ast.Node) bool {
		if n == nil {
			nodeStack = nodeStack[:len(nodeStack)-1]
			return false
		}
		nodeStack = append(nodeStack, n)
		begin := syntax.Position(n.Pos())
		end := syntax.Position(n.LastTok().End())
		switch node := n.(type) {
		case *ast.StructTypeDecl:
			if node.Name != nil {
				begin = syntax.Position(node.Name.Pos())
				end = syntax.Position(node.Name.End())
				d = append(d, tg.NewTuple(uint32(begin.Line-1), uint32(begin.Column-1), uint32(end.Offset-begin.Offset), Struct, modifierCalc(definition))...)
				return false
			}
		case *ast.Field:
			isSubType := false
			for i := range nodeStack {
				if _, ok := nodeStack[len(nodeStack)-i-1].(*ast.SubTypeDecl); ok {
					isSubType = true
				}
			}
			if isSubType {
				if node.Name != nil {
					begin = syntax.Position(node.Name.Pos())
					end = syntax.Position(node.Name.End())
					d = append(d, tg.NewTuple(uint32(begin.Line-1), uint32(begin.Column-1), uint32(end.Offset-begin.Offset), Struct, modifierCalc(definition))...)
				}
				return false
			}
		case *ast.FuncDecl:
			if node.Name != nil {
				begin = syntax.Position(node.Name.Pos())
				end = syntax.Position(node.Name.End())
				d = append(d, tg.NewTuple(uint32(begin.Line-1), uint32(begin.Column-1), uint32(end.Offset-begin.Offset), Function, modifierCalc(definition))...)
			}
		}
		return true
	})
	d = mergeSortTokenarrays(NewSyntaxTokensFromCurrentModule(fileName), d)
	ret := &protocol.SemanticTokens{Data: d}
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
	// NOTE: having the current file owned by more then one suite should not
	// import from modules originating from both suites. This would
	// in most ways end up with cyclic imports.
	// Thus 'completion' shall collect items only from one suite.
	// Decision: first suite
	syntax := suites[0].ParseWithAllErrors(params.TextDocument.URI.SpanURI().Filename())
	if syntax.Module == nil {
		return nil, syntax.Err
	}

	if syntax.Module.Name == nil {
		return nil, nil
	}
	ret = NewSemanticTokensFromCurrentModule(syntax, params.TextDocument.URI.SpanURI().Filename())
	return ret, nil
}
