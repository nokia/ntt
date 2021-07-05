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

func getExtendsComponents(syntax *ntt.ParseInfo, expr []ast.Expr) []protocol.DocumentSymbol {
	l := len(expr)
	list := make([]protocol.DocumentSymbol, 0, l)

	for _, v := range expr {
		begin := syntax.Position(v.Pos())
		end := syntax.Position(v.LastTok().End())
		list = append(list, protocol.DocumentSymbol{Name: ast.Name(v), Kind: protocol.Class,
			Range:          setProtocolRange(begin, end),
			SelectionRange: setProtocolRange(begin, end)})
	}
	begin := syntax.Position(expr[0].Pos())
	end := syntax.Position(expr[l-1].LastTok().End())
	extends := make([]protocol.DocumentSymbol, 0, 1)
	extends = append(extends, protocol.DocumentSymbol{Name: "extends", Kind: protocol.Array,
		Range:          setProtocolRange(begin, end),
		SelectionRange: setProtocolRange(begin, end), Children: list})
	return extends
}

func getTypeList(syntax *ntt.ParseInfo, types []ast.Expr) []protocol.DocumentSymbol {
	retv := make([]protocol.DocumentSymbol, 0, len(types))
	for _, t := range types {
		begin := syntax.Position(t.Pos())
		end := syntax.Position(t.LastTok().End())
		if name := ast.Name(t); len(name) > 0 {
			retv = append(retv, protocol.DocumentSymbol{
				Name:           name,
				Detail:         "type",
				Kind:           protocol.Struct,
				Range:          setProtocolRange(begin, end),
				SelectionRange: setProtocolRange(begin, end)})
		}
	}
	return retv
}

func getPortTypeDecl(syntax *ntt.ParseInfo, node *ast.PortTypeDecl) protocol.DocumentSymbol {
	begin := syntax.Position(node.Pos())
	end := syntax.Position(node.LastTok().End())
	kindstr := ""
	if node.Kind.IsValid() {
		kindstr = node.Kind.Lit
	}
	retv := protocol.DocumentSymbol{
		Name:           node.Name.String(),
		Detail:         kindstr + " port type",
		Kind:           protocol.Interface,
		Range:          setProtocolRange(begin, end),
		SelectionRange: setProtocolRange(begin, end)}
	portChildren := make([]protocol.DocumentSymbol, 0, 6)
	for _, attr := range node.Attrs {
		begin := syntax.Position(attr.Pos())
		end := syntax.Position(attr.LastTok().End())
		switch node := attr.(type) {
		case *ast.PortAttribute:
			switch node.Kind.Kind {
			case token.ADDRESS:
				portChildren = append(portChildren, protocol.DocumentSymbol{
					Name:           "address",
					Detail:         ast.Name(node.Types[0]) + " type",
					Kind:           protocol.Struct,
					Range:          setProtocolRange(begin, end),
					SelectionRange: setProtocolRange(begin, end)})
			case token.IN, token.OUT, token.INOUT:
				portChildren = append(portChildren, protocol.DocumentSymbol{
					Name:           node.Kind.String(),
					Kind:           protocol.Array,
					Range:          setProtocolRange(begin, end),
					SelectionRange: setProtocolRange(begin, end),
					Children:       getTypeList(syntax, node.Types)})
			}

		case *ast.PortMapAttribute:
		}
	}
	retv.Children = portChildren
	return retv
}

func getSignatureDecl(syntax *ntt.ParseInfo, sig *ast.SignatureDecl) protocol.DocumentSymbol {
	begin := syntax.Position(sig.Pos())
	end := syntax.Position(sig.LastTok().End())
	kindstr := "blocking"
	if sig.NoBlock.IsValid() {
		kindstr = "non-blocking"
	}
	retv := protocol.DocumentSymbol{
		Name:           sig.Name.String(),
		Detail:         kindstr + " signature",
		Kind:           protocol.Function,
		Range:          setProtocolRange(begin, end),
		SelectionRange: setProtocolRange(begin, end)}
	if sig.Return != nil || sig.Exception != nil {
		retv.Children = make([]protocol.DocumentSymbol, 0, 2)
		if sig.Return != nil {
			begin := syntax.Position(sig.Return.Type.Pos())
			end := syntax.Position(sig.Return.Type.LastTok().End())
			retv.Children = append(retv.Children, protocol.DocumentSymbol{
				Name:           ast.Name(sig.Return.Type),
				Detail:         "return type",
				Kind:           protocol.Struct,
				Range:          setProtocolRange(begin, end),
				SelectionRange: setProtocolRange(begin, end)})
		}
		if sig.Exception != nil {
			begin := syntax.Position(sig.ExceptionTok.Pos())
			end := syntax.Position(sig.Exception.LastTok().End())
			retv.Children = append(retv.Children, protocol.DocumentSymbol{
				Name:           "Exceptions",
				Kind:           protocol.Array,
				Range:          setProtocolRange(begin, end),
				SelectionRange: setProtocolRange(begin, end),
				Children:       getTypeList(syntax, sig.Exception.List)})
		}
	}
	return retv
}

func getValueDecls(syntax *ntt.ParseInfo, val *ast.ValueDecl) []protocol.DocumentSymbol {
	vdecls := make([]protocol.DocumentSymbol, 0, 2)
	begin := syntax.Position(val.Pos())
	end := syntax.Position(val.LastTok().End())
	kind := protocol.Variable
	typeName := ast.Name(val.Type)
	if len(typeName) > 0 {
		typeName = " " + typeName
	}
	switch val.Kind.Kind {
	case token.PORT:
		kind = protocol.Interface
	case token.TIMER:
		kind = protocol.Event
	case token.CONST, token.TEMPLATE, token.MODULEPAR:
		kind = protocol.Constant
	}

	for _, name := range val.Decls {
		vdecls = append(vdecls, protocol.DocumentSymbol{Name: ast.Name(name), Detail: val.Kind.Lit + typeName, Kind: kind,
			Range:          setProtocolRange(begin, end),
			SelectionRange: setProtocolRange(begin, end),
			Children:       nil})
	}
	return vdecls
}

func getComponentVars(syntax *ntt.ParseInfo, stmt []ast.Stmt) []protocol.DocumentSymbol {
	vdecls := make([]protocol.DocumentSymbol, 0, len(stmt))
	for _, v := range stmt {
		if n, ok := v.(*ast.DeclStmt); ok {
			switch node := n.Decl.(type) {
			case *ast.ValueDecl:
				vdecls = append(vdecls, getValueDecls(syntax, node)...)
			default:
			}
		}
	}
	return vdecls
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
			var children []protocol.DocumentSymbol = nil
			if len(node.Extends) > 0 {
				children = getExtendsComponents(syntax, node.Extends)
			}
			if node.Body != nil && node.Body.Stmts != nil {
				if l := len(node.Body.Stmts); l > 0 && children == nil {
					children = make([]protocol.DocumentSymbol, 0, l)
				}
				children = append(children, getComponentVars(syntax, node.Body.Stmts)...)

			}
			list = append(list, protocol.DocumentSymbol{Name: node.Name.String(), Detail: "component type", Kind: protocol.Class,
				Range:          setProtocolRange(begin, end),
				SelectionRange: setProtocolRange(begin, end),
				Children:       children})
			return false
		case *ast.PortTypeDecl:
			if node.Name == nil {
				return false
			}
			list = append(list, getPortTypeDecl(syntax, node))
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
		case *ast.SignatureDecl:
			if node.Name == nil {
				return false
			}
			list = append(list, getSignatureDecl(syntax, node))
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
		case *ast.ValueDecl:
			for _, vdecl := range getValueDecls(syntax, node) {
				list = append(list, vdecl)
			}
			return false
		case *ast.TemplateDecl:
			if node.Name == nil {
				return false
			}
			var modifies []protocol.DocumentSymbol = nil
			if node.ModifiesTok.IsValid() {
				modifName := ast.Name(node.Base)
				if len(modifName) > 0 {
					begin := syntax.Position(node.Base.Pos())
					end := syntax.Position(node.Base.End())
					modifies = make([]protocol.DocumentSymbol, 0, 1)
					modifies = append(modifies, protocol.DocumentSymbol{Name: modifName, Detail: "template", Kind: protocol.Constant,
						Range:          setProtocolRange(begin, end),
						SelectionRange: setProtocolRange(begin, end),
						Children:       nil})
				}
			}
			typeName := ast.Name(node.Type)
			list = append(list, protocol.DocumentSymbol{Name: node.Name.String(), Detail: "template " + typeName, Kind: protocol.Constant,
				Range:          setProtocolRange(begin, end),
				SelectionRange: setProtocolRange(begin, end),
				Children:       modifies})
			return false
		default:
			return true
		}

	})
	return list
}
func (s *Server) documentSymbol(ctx context.Context, params *protocol.DocumentSymbolParams) ([]interface{}, error) {
	var ret []interface{} = nil
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
		log.Debug(fmt.Sprintf("DocumentSymbol took %s.", elapsed))
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
	ret = NewAllDefinitionSymbolsFromCurrentModule(syntax)
	return ret, nil
}
