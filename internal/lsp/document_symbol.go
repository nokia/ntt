package lsp

import (
	"context"
	"fmt"
	"time"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/internal/ntt"
	"github.com/nokia/ntt/internal/ttcn3/ast"
	"github.com/nokia/ntt/internal/ttcn3/token"
)

var kindToStringMap = map[token.Kind]string{
	token.ALTSTEP: "altstep", token.FUNCTION: "function", token.TESTCASE: "testcase",
	token.UNION: "union", token.RECORD: "record", token.SET: "set"}

func setProtocolRange(begin loc.Position, end loc.Position) protocol.Range {
	return protocol.Range{
		Start: protocol.Position{Line: float64(begin.Line - 1), Character: float64(begin.Column - 1)},
		End:   protocol.Position{Line: float64(end.Line - 1), Character: float64(end.Column - 1)}}
}

func getElemTypeInfo(syntax *ntt.ParseInfo, n ast.TypeSpec) []protocol.DocumentSymbol {
	typeSymb := make([]protocol.DocumentSymbol, 0, 1)
	begin := syntax.Position(n.Pos())
	end := syntax.Position(n.LastTok().End())
	switch node := n.(type) {
	case *ast.RefSpec:
		typeSymb = append(typeSymb, protocol.DocumentSymbol{Name: ast.Name(node), Detail: "element type", Kind: protocol.Struct,
			Range:          setProtocolRange(begin, end),
			SelectionRange: setProtocolRange(begin, end),
			Children:       nil})
	}
	return typeSymb
}
func NewAllDefinitionSymbolsFromCurrentModule(syntax *ntt.ParseInfo) []interface{} {
	list := make([]interface{}, 0, 20)

	ast.Inspect(syntax.Module, func(n ast.Node) bool {

		if n == nil {
			return false
		}
		begin := syntax.Position(n.Pos())
		end := syntax.Position(n.LastTok().End())
		switch node := n.(type) {
		case *ast.FuncDecl:
			if node.Name == nil {
				// looks like a syntax error
				return false
			}

			kind := protocol.Function
			children := make([]protocol.DocumentSymbol, 0, 5)
			if node.RunsOn != nil && node.RunsOn.Comp != nil {
				kind = protocol.Method
				idBegin := syntax.Position(node.RunsOn.Comp.Pos())
				idEnd := syntax.Position(node.RunsOn.Comp.LastTok().End())
				children = append(children, protocol.DocumentSymbol{Name: "runs on", Detail: ast.Name(node.RunsOn.Comp),
					Kind:           protocol.Class,
					Range:          setProtocolRange(idBegin, idEnd),
					SelectionRange: setProtocolRange(idBegin, idEnd)})
			}
			if node.System != nil && node.System.Comp != nil {
				kind = protocol.Method
				idBegin := syntax.Position(node.System.Comp.Pos())
				idEnd := syntax.Position(node.System.Comp.LastTok().End())
				children = append(children, protocol.DocumentSymbol{Name: "system", Detail: ast.Name(node.System.Comp),
					Kind:           protocol.Class,
					Range:          setProtocolRange(idBegin, idEnd),
					SelectionRange: setProtocolRange(idBegin, idEnd)})
			}
			if node.Return != nil && node.Return.Type != nil {
				idBegin := syntax.Position(node.Return.Type.Pos())
				idEnd := syntax.Position(node.Return.Type.LastTok().End())
				children = append(children, protocol.DocumentSymbol{Name: "return", Detail: ast.Name(node.Return.Type),
					Kind:           protocol.Struct,
					Range:          setProtocolRange(idBegin, idEnd),
					SelectionRange: setProtocolRange(idBegin, idEnd)})
			}
			detail := kindToStringMap[node.Kind.Kind] + " definition"
			if node.External.IsValid() {
				detail = "external " + detail
			}
			list = append(list, protocol.DocumentSymbol{Name: node.Name.String(), Detail: detail, Kind: kind,
				Range:          setProtocolRange(begin, end),
				SelectionRange: setProtocolRange(begin, end),
				Children:       children})
			return false
		case *ast.ComponentTypeDecl:
			if node.Name == nil {
				return false
			}
			list = append(list, protocol.DocumentSymbol{Name: node.Name.String(), Detail: "component type", Kind: protocol.Class,
				Range:          setProtocolRange(begin, end),
				SelectionRange: setProtocolRange(begin, end),
				Children:       nil})
			return false
		case *ast.PortTypeDecl:
			if node.Name == nil {
				return false
			}
			list = append(list, protocol.DocumentSymbol{Name: node.Name.String(), Detail: "port type", Kind: protocol.Interface,
				Range:          setProtocolRange(begin, end),
				SelectionRange: setProtocolRange(begin, end),
				Children:       nil})
			return false
		case *ast.EnumTypeDecl:
			if node.Name == nil {
				return false
			}
			list = append(list, protocol.DocumentSymbol{Name: node.Name.String(), Detail: "enum type", Kind: protocol.Enum,
				Range:          setProtocolRange(begin, end),
				SelectionRange: setProtocolRange(begin, end),
				Children:       nil})
			return false
		case *ast.SubTypeDecl:
			var children []protocol.DocumentSymbol = nil
			detail := "subtype"
			kind := protocol.Struct

			if node.Field == nil || node.Field.Name == nil {
				return false
			}

			if listNode, ok := node.Field.Type.(*ast.ListSpec); ok {
				detail = kindToStringMap[listNode.Kind.Kind] + " of type"
				kind = protocol.Array
				children = getElemTypeInfo(syntax, listNode.ElemType)
			}

			list = append(list, protocol.DocumentSymbol{Name: node.Field.Name.String(), Detail: detail, Kind: kind,
				Range:          setProtocolRange(begin, end),
				SelectionRange: setProtocolRange(begin, end),
				Children:       children})
			return false
		case *ast.StructTypeDecl:
			if node.Name == nil {
				return false
			}
			detail := kindToStringMap[node.Kind.Kind] + " type"
			list = append(list, protocol.DocumentSymbol{Name: node.Name.String(), Detail: detail, Kind: protocol.Struct,
				Range:          setProtocolRange(begin, end),
				SelectionRange: setProtocolRange(begin, end),
				Children:       nil})
			return false
		case *ast.BehaviourTypeDecl:
			if node.Name == nil {
				return false
			}
			list = append(list, protocol.DocumentSymbol{Name: node.Name.String(), Detail: " subtype", Kind: protocol.Operator,
				Range:          setProtocolRange(begin, end),
				SelectionRange: setProtocolRange(begin, end),
				Children:       nil})
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
	ret := NewAllDefinitionSymbolsFromCurrentModule(syntax)
	return ret, nil
}
