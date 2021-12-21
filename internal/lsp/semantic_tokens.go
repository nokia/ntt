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

func (tg *TokenGen) NewTuple(line uint32, column uint32, length int, tokenType SemanticTokenType, modifier uint32) []uint32 {
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
			d = append(d, tg.NewTuple(line, column, len(lit), Keyword, modifierCalc())...)
		}

	}
	return d
}

func NewSemanticTokensFromCurrentModule(syntax *ntt.ParseInfo, fileName string) *protocol.SemanticTokens {
	d := make([]uint32, 0, 20)
	ast.Inspect(syntax.Module, func(n ast.Node) bool {
		if n == nil {
			return false
		}
		begin := syntax.Position(n.Pos())
		end := syntax.Position(n.LastTok().End())
		switch node := n.(type) {
		case *ast.Token:
			if node.Kind.IsKeyword() {
				d = append(d, uint32(begin.Line), uint32(begin.Column), uint32(end.Offset-begin.Offset), uint32(Keyword), modifierCalc())
				return false
			}
		}
		return true
	})
	d = append(d, NewSyntaxTokensFromCurrentModule(fileName)...)
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
