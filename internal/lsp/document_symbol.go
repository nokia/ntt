package lsp

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/internal/ntt"
	"github.com/nokia/ntt/internal/ttcn3/ast"
	"github.com/nokia/ntt/internal/ttcn3/token"
)

var kindToStringMap = map[token.Kind]string{token.ALTSTEP: "altstep", token.FUNCTION: "function", token.TESTCASE: "testcase"}

func setProtocolRange(begin loc.Position, end loc.Position) protocol.Range {
	return protocol.Range{
		Start: protocol.Position{Line: float64(begin.Line - 1), Character: float64(begin.Column - 1)},
		End:   protocol.Position{Line: float64(end.Line - 1), Character: float64(end.Column - 1)}}
}

func getIdName(e ast.Expr) string {
	name := ""
	ast.Inspect(e, func(n ast.Node) bool {

		if n == nil {
			return false
		}

		switch node := n.(type) {
		case *ast.Ident:
			name = node.Tok.String()
			if node.Tok2.IsValid() {
				name += "." + node.Tok2.String()
			}
			return false
		default:
			return true
		}
	})
	return name
}

func newAllDefinitionSymbolsFromCurrentModule(syntax *ntt.ParseInfo) []interface{} {
	list := make([]interface{}, 0, 20)

	ast.Inspect(syntax.Module, func(n ast.Node) bool {

		if n == nil {
			return false
		}

		switch node := n.(type) {
		case *ast.FuncDecl:
			if node.Name == nil {
				// looks like a syntax error
				return false
			}
			begin := syntax.Position(node.Pos())
			end := syntax.Position(node.LastTok().End())
			kind := protocol.Function
			children := make([]protocol.DocumentSymbol, 0, 5)
			if node.RunsOn != nil && node.RunsOn.Comp != nil {
				kind = protocol.Method
				idBegin := syntax.Position(node.RunsOn.Comp.Pos())
				idEnd := syntax.Position(node.RunsOn.Comp.LastTok().End())
				children = append(children, protocol.DocumentSymbol{Name: "runs on", Detail: getIdName(node.RunsOn.Comp),
					Kind:           protocol.Class,
					Range:          setProtocolRange(idBegin, idEnd),
					SelectionRange: setProtocolRange(idBegin, idEnd)})
			}
			if node.System != nil && node.System.Comp != nil {
				kind = protocol.Method
				idBegin := syntax.Position(node.System.Comp.Pos())
				idEnd := syntax.Position(node.System.Comp.LastTok().End())
				children = append(children, protocol.DocumentSymbol{Name: "system", Detail: getIdName(node.System.Comp),
					Kind:           protocol.Class,
					Range:          setProtocolRange(idBegin, idEnd),
					SelectionRange: setProtocolRange(idBegin, idEnd)})
			}
			if node.Return != nil && node.Return.Type != nil {
				idBegin := syntax.Position(node.Return.Type.Pos())
				idEnd := syntax.Position(node.Return.Type.LastTok().End())
				children = append(children, protocol.DocumentSymbol{Name: "return", Detail: getIdName(node.Return.Type),
					Kind:           protocol.Struct,
					Range:          setProtocolRange(idBegin, idEnd),
					SelectionRange: setProtocolRange(idBegin, idEnd)})
			}
			list = append(list, protocol.DocumentSymbol{Name: node.Name.String(), Detail: kindToStringMap[node.Kind.Kind] + " definition", Kind: kind,
				Range:          setProtocolRange(begin, end),
				SelectionRange: setProtocolRange(begin, end),
				Children:       children})

			return false
		default:
			return true
		}

	})
	return list
}
func (s *Server) documentSymbol(ctx context.Context, params *protocol.DocumentSymbolParams) ([]interface{}, error) {
	start := time.Now()
	defer func() {
		elapsed := time.Since(start)
		log.Debug(fmt.Sprintf("DocumentSymbol took %s.", elapsed))
	}()

	fileName := filepath.Base(params.TextDocument.URI.SpanURI().Filename())
	defaultModuleId := fileName[:len(fileName)-len(filepath.Ext(fileName))]

	suites := s.Owners(params.TextDocument.URI)
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
	ret := make([]interface{}, 0, 1)
	ret = append(ret, protocol.DocumentSymbol{Name: defaultModuleId, Detail: "record type", Kind: protocol.Struct,
		Range:          protocol.Range{Start: protocol.Position{Line: 1, Character: 1}, End: protocol.Position{Line: 20, Character: 1}},
		SelectionRange: protocol.Range{Start: protocol.Position{Line: 1, Character: 1}, End: protocol.Position{Line: 20, Character: 1}}})
	ret = newAllDefinitionSymbolsFromCurrentModule(syntax)
	return ret, nil
}
