package lsp

import (
	"context"
	"time"

	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/ttcn3"
	"github.com/nokia/ntt/ttcn3/syntax"
)

var kindToStringMap = map[syntax.Kind]string{
	syntax.ALTSTEP: "altstep", syntax.FUNCTION: "function", syntax.TESTCASE: "testcase",
	syntax.UNION: "union", syntax.RECORD: "record", syntax.SET: "set"}

func getComponentTypeDecl(tree *ttcn3.Tree, node *syntax.ComponentTypeDecl) protocol.DocumentSymbol {
	var children []protocol.DocumentSymbol = nil

	if len(node.Extends) > 0 {
		children = getExtendsComponents(tree, node.Extends)
	}
	if node.Body != nil && node.Body.Stmts != nil {
		if l := len(node.Body.Stmts); l > 0 && children == nil {
			children = make([]protocol.DocumentSymbol, 0, l)
		}
		children = append(children, getComponentVars(tree, node.Body.Stmts)...)

	}
	spn := syntax.SpanOf(node)
	rng := setProtocolRange(spn.Begin, spn.End)
	return protocol.DocumentSymbol{Name: node.Name.String(), Detail: "component type", Kind: protocol.Class,
		Range:          rng,
		SelectionRange: rng,
		Children:       children}
}
func getExtendsComponents(tree *ttcn3.Tree, expr []syntax.Expr) []protocol.DocumentSymbol {
	l := len(expr)
	list := make([]protocol.DocumentSymbol, 0, l)

	for _, v := range expr {
		spn := syntax.SpanOf(v)
		rng := setProtocolRange(spn.Begin, spn.End)
		list = append(list, protocol.DocumentSymbol{Name: syntax.Name(v), Kind: protocol.Class,
			Range:          rng,
			SelectionRange: rng,
		})
	}
	begin := syntax.Begin(expr[0])
	end := syntax.End(expr[l-1])
	extends := make([]protocol.DocumentSymbol, 0, 1)
	extends = append(extends, protocol.DocumentSymbol{Name: "extends", Kind: protocol.Array,
		Range:          setProtocolRange(begin, end),
		SelectionRange: setProtocolRange(begin, end), Children: list})
	return extends
}

func getFunctionDecl(tree *ttcn3.Tree, node *syntax.FuncDecl) protocol.DocumentSymbol {
	begin := syntax.Begin(node)
	end := syntax.End(node)
	kind := protocol.Function
	children := make([]protocol.DocumentSymbol, 0, 5)
	if node.RunsOn != nil && node.RunsOn.Comp != nil {
		kind = protocol.Method
		idBegin := syntax.Begin(node.RunsOn.Comp)
		idEnd := syntax.End(node.RunsOn.Comp)
		children = append(children, protocol.DocumentSymbol{Name: "runs on", Detail: syntax.Name(node.RunsOn.Comp),
			Kind:           protocol.Class,
			Range:          setProtocolRange(idBegin, idEnd),
			SelectionRange: setProtocolRange(idBegin, idEnd)})
	}
	if node.System != nil && node.System.Comp != nil {
		kind = protocol.Method
		idBegin := syntax.Begin(node.System.Comp)
		idEnd := syntax.End(node.System.Comp)
		children = append(children, protocol.DocumentSymbol{Name: "system", Detail: syntax.Name(node.System.Comp),
			Kind:           protocol.Class,
			Range:          setProtocolRange(idBegin, idEnd),
			SelectionRange: setProtocolRange(idBegin, idEnd)})
	}
	if node.Return != nil && node.Return.Type != nil {
		idBegin := syntax.Begin(node.Return.Type)
		idEnd := syntax.End(node.Return.Type)
		children = append(children, protocol.DocumentSymbol{Name: "return", Detail: syntax.Name(node.Return.Type),
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

func getTypeList(tree *ttcn3.Tree, types []syntax.Expr) []protocol.DocumentSymbol {
	retv := make([]protocol.DocumentSymbol, 0, len(types))
	for _, t := range types {
		begin := syntax.Begin(t)
		end := syntax.End(t)
		if name := syntax.Name(t); len(name) > 0 {
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

func getPortTypeDecl(tree *ttcn3.Tree, node *syntax.PortTypeDecl) protocol.DocumentSymbol {
	begin := syntax.Begin(node)
	end := syntax.End(node)
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
		begin := syntax.Begin(attr)
		end := syntax.End(attr)
		switch node := attr.(type) {
		case *syntax.PortAttribute:
			switch node.Kind.Kind() {
			case syntax.ADDRESS:
				portChildren = append(portChildren, protocol.DocumentSymbol{
					Name:           "address",
					Detail:         syntax.Name(node.Types[0]) + " type",
					Kind:           protocol.Struct,
					Range:          setProtocolRange(begin, end),
					SelectionRange: setProtocolRange(begin, end)})
			case syntax.IN, syntax.OUT, syntax.INOUT:
				portChildren = append(portChildren, protocol.DocumentSymbol{
					Name:           node.Kind.String(),
					Kind:           protocol.Array,
					Range:          setProtocolRange(begin, end),
					SelectionRange: setProtocolRange(begin, end),
					Children:       getTypeList(tree, node.Types)})
			}

		case *syntax.PortMapAttribute:
		}
	}
	retv.Children = portChildren
	return retv
}

func getSignatureDecl(tree *ttcn3.Tree, sig *syntax.SignatureDecl) protocol.DocumentSymbol {
	begin := syntax.Begin(sig)
	end := syntax.End(sig)
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
			begin := syntax.Begin(sig.Return.Type)
			end := syntax.End(sig.Return.Type)
			retv.Children = append(retv.Children, protocol.DocumentSymbol{
				Name:           syntax.Name(sig.Return.Type),
				Detail:         "return type",
				Kind:           protocol.Struct,
				Range:          setProtocolRange(begin, end),
				SelectionRange: setProtocolRange(begin, end)})
		}
		if sig.Exception != nil {
			begin := syntax.Begin(sig.ExceptionTok)
			end := syntax.End(sig.Exception)
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

func getSubTypeDecl(tree *ttcn3.Tree, node *syntax.SubTypeDecl) protocol.DocumentSymbol {
	var children []protocol.DocumentSymbol = nil
	begin := syntax.Begin(node)
	end := syntax.End(node)
	detail := "subtype"
	kind := protocol.Struct
	if listNode, ok := node.Field.Type.(*syntax.ListSpec); ok {
		detail = kindToStringMap[listNode.Kind.Kind()] + " of type"
		kind = protocol.Array
		children = getElemTypeInfo(tree, listNode.ElemType)
	}

	return protocol.DocumentSymbol{Name: node.Field.Name.String(), Detail: detail, Kind: kind,
		Range:          setProtocolRange(begin, end),
		SelectionRange: setProtocolRange(begin, end),
		Children:       children}
}

func getTemplateDecl(tree *ttcn3.Tree, node *syntax.TemplateDecl) protocol.DocumentSymbol {
	var modifies []protocol.DocumentSymbol = nil
	begin := syntax.Begin(node)
	end := syntax.End(node)
	if node.ModifiesTok != nil {
		modifName := syntax.Name(node.Base)
		if len(modifName) > 0 {
			begin := syntax.Begin(node.Base)
			end := syntax.End(node.Base)
			modifies = make([]protocol.DocumentSymbol, 0, 1)
			modifies = append(modifies, protocol.DocumentSymbol{Name: modifName, Detail: "template", Kind: protocol.Constant,
				Range:          setProtocolRange(begin, end),
				SelectionRange: setProtocolRange(begin, end),
				Children:       nil})
		}
	}
	typeName := syntax.Name(node.Type)

	return protocol.DocumentSymbol{Name: node.Name.String(), Detail: "template " + typeName, Kind: protocol.Constant,
		Range:          setProtocolRange(begin, end),
		SelectionRange: setProtocolRange(begin, end),
		Children:       modifies}
}

func getValueDecls(tree *ttcn3.Tree, val *syntax.ValueDecl) []protocol.DocumentSymbol {
	vdecls := make([]protocol.DocumentSymbol, 0, 2)
	begin := syntax.Begin(val)
	end := syntax.End(val)
	kind := protocol.Variable
	typeName := syntax.Name(val.Type)
	if len(typeName) > 0 {
		typeName = " " + typeName
	}
	switch val.Kind.Kind() {
	case syntax.PORT:
		kind = protocol.Interface
	case syntax.TIMER:
		kind = protocol.Event
	case syntax.CONST, syntax.TEMPLATE, syntax.MODULEPAR:
		kind = protocol.Constant
	}

	for _, name := range val.Decls {
		vdecls = append(vdecls, protocol.DocumentSymbol{Name: syntax.Name(name), Detail: val.Kind.String() + typeName, Kind: kind,
			Range:          setProtocolRange(begin, end),
			SelectionRange: setProtocolRange(begin, end),
			Children:       nil})
	}
	return vdecls
}

func getComponentVars(tree *ttcn3.Tree, stmt []syntax.Stmt) []protocol.DocumentSymbol {
	vdecls := make([]protocol.DocumentSymbol, 0, len(stmt))
	for _, v := range stmt {
		if n, ok := v.(*syntax.DeclStmt); ok {
			switch node := n.Decl.(type) {
			case *syntax.ValueDecl:
				vdecls = append(vdecls, getValueDecls(tree, node)...)
			default:
			}
		}
	}
	return vdecls
}

func getElemTypeInfo(tree *ttcn3.Tree, n syntax.TypeSpec) []protocol.DocumentSymbol {
	typeSymb := make([]protocol.DocumentSymbol, 0, 1)
	begin := syntax.Begin(n)
	end := syntax.End(n)
	switch node := n.(type) {
	case *syntax.RefSpec:
		typeSymb = append(typeSymb, protocol.DocumentSymbol{Name: syntax.Name(node), Detail: "element type", Kind: protocol.Struct,
			Range:          setProtocolRange(begin, end),
			SelectionRange: setProtocolRange(begin, end),
			Children:       nil})
	}
	return typeSymb
}

func NewAllDefinitionSymbolsFromCurrentModule(tree *ttcn3.Tree) []interface{} {
	list := make([]interface{}, 0, 20)

	tree.Inspect(func(n syntax.Node) bool {

		if n == nil {
			return false
		}
		begin := syntax.Begin(n)
		end := syntax.End(n)
		switch node := n.(type) {
		case *syntax.FuncDecl:
			if node.Name == nil {
				// looks like a tree error
				return false
			}
			list = append(list, getFunctionDecl(tree, node))
			return false
		case *syntax.ComponentTypeDecl:
			if node.Name == nil {
				return false
			}
			list = append(list, getComponentTypeDecl(tree, node))
			return false
		case *syntax.PortTypeDecl:
			if node.Name == nil {
				return false
			}
			list = append(list, getPortTypeDecl(tree, node))
			return false
		case *syntax.EnumTypeDecl:
			if node.Name == nil {
				return false
			}
			list = append(list, protocol.DocumentSymbol{Name: node.Name.String(), Detail: "enum type", Kind: protocol.Enum,
				Range:          setProtocolRange(begin, end),
				SelectionRange: setProtocolRange(begin, end),
				Children:       nil})
			return false
		case *syntax.SignatureDecl:
			if node.Name == nil {
				return false
			}
			list = append(list, getSignatureDecl(tree, node))
			return false
		case *syntax.SubTypeDecl:
			if node.Field == nil || node.Field.Name == nil {
				return false
			}
			list = append(list, getSubTypeDecl(tree, node))
			return false
		case *syntax.StructTypeDecl:
			if node.Name == nil {
				return false
			}
			detail := kindToStringMap[node.Kind.Kind()] + " type"
			list = append(list, protocol.DocumentSymbol{Name: node.Name.String(), Detail: detail, Kind: protocol.Struct,
				Range:          setProtocolRange(begin, end),
				SelectionRange: setProtocolRange(begin, end),
				Children:       nil})
			return false
		case *syntax.BehaviourTypeDecl:
			if node.Name == nil {
				return false
			}
			list = append(list, protocol.DocumentSymbol{Name: node.Name.String(), Detail: " subtype", Kind: protocol.Operator,
				Range:          setProtocolRange(begin, end),
				SelectionRange: setProtocolRange(begin, end),
				Children:       nil})
			return false
		case *syntax.ValueDecl:
			for _, vdecl := range getValueDecls(tree, node) {
				list = append(list, vdecl)
			}
			return false
		case *syntax.TemplateDecl:
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
