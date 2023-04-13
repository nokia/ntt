package lsp

import (
	"context"
	"time"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/ttcn3"
	"github.com/nokia/ntt/ttcn3/ast"
	"github.com/nokia/ntt/ttcn3/token"
)

var kindToStringMap = map[token.Kind]string{
	token.ALTSTEP: "altstep", token.FUNCTION: "function", token.TESTCASE: "testcase",
	token.UNION: "union", token.RECORD: "record", token.SET: "set"}

func setProtocolRange(begin loc.Position, end loc.Position) protocol.Range {
	return protocol.Range{
		Start: protocol.Position{Line: uint32(begin.Line - 1), Character: uint32(begin.Column - 1)},
		End:   protocol.Position{Line: uint32(end.Line - 1), Character: uint32(end.Column - 1)}}
}

func getComponentTypeDecl(tree *ttcn3.Tree, node *ast.ComponentTypeDecl) protocol.DocumentSymbol {
	var children []protocol.DocumentSymbol = nil
	begin := tree.Position(node.Pos())
	end := tree.Position(node.LastTok().End())

	if len(node.Extends) > 0 {
		children = getExtendsComponents(tree, node.Extends)
	}
	if node.Body != nil && node.Body.Stmts != nil {
		if l := len(node.Body.Stmts); l > 0 && children == nil {
			children = make([]protocol.DocumentSymbol, 0, l)
		}
		children = append(children, getComponentVars(tree, node.Body.Stmts)...)

	}
	return protocol.DocumentSymbol{Name: node.Name.String(), Detail: "component type", Kind: protocol.Class,
		Range:          setProtocolRange(begin, end),
		SelectionRange: setProtocolRange(begin, end),
		Children:       children}
}
func getExtendsComponents(tree *ttcn3.Tree, expr []ast.Expr) []protocol.DocumentSymbol {
	l := len(expr)
	list := make([]protocol.DocumentSymbol, 0, l)

	for _, v := range expr {
		begin := tree.Position(v.Pos())
		end := tree.Position(v.LastTok().End())
		list = append(list, protocol.DocumentSymbol{Name: ast.Name(v), Kind: protocol.Class,
			Range:          setProtocolRange(begin, end),
			SelectionRange: setProtocolRange(begin, end)})
	}
	begin := tree.Position(expr[0].Pos())
	end := tree.Position(expr[l-1].LastTok().End())
	extends := make([]protocol.DocumentSymbol, 0, 1)
	extends = append(extends, protocol.DocumentSymbol{Name: "extends", Kind: protocol.Array,
		Range:          setProtocolRange(begin, end),
		SelectionRange: setProtocolRange(begin, end), Children: list})
	return extends
}

func getFunctionDecl(tree *ttcn3.Tree, node *ast.FuncDecl) protocol.DocumentSymbol {
	begin := tree.Position(node.Pos())
	end := tree.Position(node.LastTok().End())
	kind := protocol.Function
	children := make([]protocol.DocumentSymbol, 0, 5)
	if node.RunsOn != nil && node.RunsOn.Comp != nil {
		kind = protocol.Method
		idBegin := tree.Position(node.RunsOn.Comp.Pos())
		idEnd := tree.Position(node.RunsOn.Comp.LastTok().End())
		children = append(children, protocol.DocumentSymbol{Name: "runs on", Detail: ast.Name(node.RunsOn.Comp),
			Kind:           protocol.Class,
			Range:          setProtocolRange(idBegin, idEnd),
			SelectionRange: setProtocolRange(idBegin, idEnd)})
	}
	if node.System != nil && node.System.Comp != nil {
		kind = protocol.Method
		idBegin := tree.Position(node.System.Comp.Pos())
		idEnd := tree.Position(node.System.Comp.LastTok().End())
		children = append(children, protocol.DocumentSymbol{Name: "system", Detail: ast.Name(node.System.Comp),
			Kind:           protocol.Class,
			Range:          setProtocolRange(idBegin, idEnd),
			SelectionRange: setProtocolRange(idBegin, idEnd)})
	}
	if node.Return != nil && node.Return.Type != nil {
		idBegin := tree.Position(node.Return.Type.Pos())
		idEnd := tree.Position(node.Return.Type.LastTok().End())
		children = append(children, protocol.DocumentSymbol{Name: "return", Detail: ast.Name(node.Return.Type),
			Kind:           protocol.Struct,
			Range:          setProtocolRange(idBegin, idEnd),
			SelectionRange: setProtocolRange(idBegin, idEnd)})
	}
	detail := node.Kind.String() + " definition"
	if node.External != nil {
		detail = "external " + detail
	}
	return protocol.DocumentSymbol{Name: node.Name.String(), Detail: detail, Kind: kind,
		Range:          setProtocolRange(begin, end),
		SelectionRange: setProtocolRange(begin, end),
		Children:       children}
}

func getTypeList(tree *ttcn3.Tree, types []ast.Expr) []protocol.DocumentSymbol {
	retv := make([]protocol.DocumentSymbol, 0, len(types))
	for _, t := range types {
		begin := tree.Position(t.Pos())
		end := tree.Position(t.LastTok().End())
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

func getPortTypeDecl(tree *ttcn3.Tree, node *ast.PortTypeDecl) protocol.DocumentSymbol {
	begin := tree.Position(node.Pos())
	end := tree.Position(node.LastTok().End())
	kindstr := ""
	if node.Kind != nil {
		kindstr = node.Kind.String()
	}
	retv := protocol.DocumentSymbol{
		Name:           node.Name.String(),
		Detail:         kindstr + " port type",
		Kind:           protocol.Interface,
		Range:          setProtocolRange(begin, end),
		SelectionRange: setProtocolRange(begin, end)}
	portChildren := make([]protocol.DocumentSymbol, 0, 6)
	for _, attr := range node.Attrs {
		begin := tree.Position(attr.Pos())
		end := tree.Position(attr.LastTok().End())
		switch node := attr.(type) {
		case *ast.PortAttribute:
			switch node.Kind.Kind() {
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
					Children:       getTypeList(tree, node.Types)})
			}

		case *ast.PortMapAttribute:
		}
	}
	retv.Children = portChildren
	return retv
}

func getSignatureDecl(tree *ttcn3.Tree, sig *ast.SignatureDecl) protocol.DocumentSymbol {
	begin := tree.Position(sig.Pos())
	end := tree.Position(sig.LastTok().End())
	kindstr := "blocking"
	if sig.NoBlock != nil {
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
			begin := tree.Position(sig.Return.Type.Pos())
			end := tree.Position(sig.Return.Type.LastTok().End())
			retv.Children = append(retv.Children, protocol.DocumentSymbol{
				Name:           ast.Name(sig.Return.Type),
				Detail:         "return type",
				Kind:           protocol.Struct,
				Range:          setProtocolRange(begin, end),
				SelectionRange: setProtocolRange(begin, end)})
		}
		if sig.Exception != nil {
			begin := tree.Position(sig.ExceptionTok.Pos())
			end := tree.Position(sig.Exception.LastTok().End())
			retv.Children = append(retv.Children, protocol.DocumentSymbol{
				Name:           "Exceptions",
				Kind:           protocol.Array,
				Range:          setProtocolRange(begin, end),
				SelectionRange: setProtocolRange(begin, end),
				Children:       getTypeList(tree, sig.Exception.List)})
		}
	}
	return retv
}

func getSubTypeDecl(tree *ttcn3.Tree, node *ast.SubTypeDecl) protocol.DocumentSymbol {
	var children []protocol.DocumentSymbol = nil
	begin := tree.Position(node.Pos())
	end := tree.Position(node.LastTok().End())
	detail := "subtype"
	kind := protocol.Struct
	if listNode, ok := node.Field.Type.(*ast.ListSpec); ok {
		detail = kindToStringMap[listNode.Kind.Kind()] + " of type"
		kind = protocol.Array
		children = getElemTypeInfo(tree, listNode.ElemType)
	}

	return protocol.DocumentSymbol{Name: node.Field.Name.String(), Detail: detail, Kind: kind,
		Range:          setProtocolRange(begin, end),
		SelectionRange: setProtocolRange(begin, end),
		Children:       children}
}

func getTemplateDecl(tree *ttcn3.Tree, node *ast.TemplateDecl) protocol.DocumentSymbol {
	var modifies []protocol.DocumentSymbol = nil
	begin := tree.Position(node.Pos())
	end := tree.Position(node.LastTok().End())
	if node.ModifiesTok != nil {
		modifName := ast.Name(node.Base)
		if len(modifName) > 0 {
			begin := tree.Position(node.Base.Pos())
			end := tree.Position(node.Base.End())
			modifies = make([]protocol.DocumentSymbol, 0, 1)
			modifies = append(modifies, protocol.DocumentSymbol{Name: modifName, Detail: "template", Kind: protocol.Constant,
				Range:          setProtocolRange(begin, end),
				SelectionRange: setProtocolRange(begin, end),
				Children:       nil})
		}
	}
	typeName := ast.Name(node.Type)

	return protocol.DocumentSymbol{Name: node.Name.String(), Detail: "template " + typeName, Kind: protocol.Constant,
		Range:          setProtocolRange(begin, end),
		SelectionRange: setProtocolRange(begin, end),
		Children:       modifies}
}

func getValueDecls(tree *ttcn3.Tree, val *ast.ValueDecl) []protocol.DocumentSymbol {
	vdecls := make([]protocol.DocumentSymbol, 0, 2)
	begin := tree.Position(val.Pos())
	end := tree.Position(val.LastTok().End())
	kind := protocol.Variable
	typeName := ast.Name(val.Type)
	if len(typeName) > 0 {
		typeName = " " + typeName
	}
	switch val.Kind.Kind() {
	case token.PORT:
		kind = protocol.Interface
	case token.TIMER:
		kind = protocol.Event
	case token.CONST, token.TEMPLATE, token.MODULEPAR:
		kind = protocol.Constant
	}

	for _, name := range val.Decls {
		vdecls = append(vdecls, protocol.DocumentSymbol{Name: ast.Name(name), Detail: val.Kind.String() + typeName, Kind: kind,
			Range:          setProtocolRange(begin, end),
			SelectionRange: setProtocolRange(begin, end),
			Children:       nil})
	}
	return vdecls
}

func getComponentVars(tree *ttcn3.Tree, stmt []ast.Stmt) []protocol.DocumentSymbol {
	vdecls := make([]protocol.DocumentSymbol, 0, len(stmt))
	for _, v := range stmt {
		if n, ok := v.(*ast.DeclStmt); ok {
			switch node := n.Decl.(type) {
			case *ast.ValueDecl:
				vdecls = append(vdecls, getValueDecls(tree, node)...)
			default:
			}
		}
	}
	return vdecls
}

func getElemTypeInfo(tree *ttcn3.Tree, n ast.TypeSpec) []protocol.DocumentSymbol {
	typeSymb := make([]protocol.DocumentSymbol, 0, 1)
	begin := tree.Position(n.Pos())
	end := tree.Position(n.LastTok().End())
	switch node := n.(type) {
	case *ast.RefSpec:
		typeSymb = append(typeSymb, protocol.DocumentSymbol{Name: ast.Name(node), Detail: "element type", Kind: protocol.Struct,
			Range:          setProtocolRange(begin, end),
			SelectionRange: setProtocolRange(begin, end),
			Children:       nil})
	}
	return typeSymb
}

func NewAllDefinitionSymbolsFromCurrentModule(tree *ttcn3.Tree) []interface{} {
	list := make([]interface{}, 0, 20)

	tree.Root.Inspect(func(n ast.Node) bool {

		if n == nil {
			return false
		}
		begin := tree.Position(n.Pos())
		end := tree.Position(n.LastTok().End())
		switch node := n.(type) {
		case *ast.FuncDecl:
			if node.Name == nil {
				// looks like a tree error
				return false
			}
			list = append(list, getFunctionDecl(tree, node))
			return false
		case *ast.ComponentTypeDecl:
			if node.Name == nil {
				return false
			}
			list = append(list, getComponentTypeDecl(tree, node))
			return false
		case *ast.PortTypeDecl:
			if node.Name == nil {
				return false
			}
			list = append(list, getPortTypeDecl(tree, node))
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
			list = append(list, getSignatureDecl(tree, node))
			return false
		case *ast.SubTypeDecl:
			if node.Field == nil || node.Field.Name == nil {
				return false
			}
			list = append(list, getSubTypeDecl(tree, node))
			return false
		case *ast.StructTypeDecl:
			if node.Name == nil {
				return false
			}
			detail := kindToStringMap[node.Kind.Kind()] + " type"
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
			for _, vdecl := range getValueDecls(tree, node) {
				list = append(list, vdecl)
			}
			return false
		case *ast.TemplateDecl:
			if node.Name == nil {
				return false
			}
			list = append(list, getTemplateDecl(tree, node))
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
		if err := recover(); err != nil {
			// in case of a panic, just continue as this might be a common situation during typing
			log.Debugf("Panic: %#v\n", err)
		}
		elapsed := time.Since(start)
		log.Debugf("DocumentSymbol took %s.\n", elapsed)
	}()

	file := params.TextDocument.URI.SpanURI().Filename()
	tree := ttcn3.ParseFile(file)
	return NewAllDefinitionSymbolsFromCurrentModule(tree), nil
}
