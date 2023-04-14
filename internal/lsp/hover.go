package lsp

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/ttcn3"
	"github.com/nokia/ntt/ttcn3/ast"
)

func (md *MarkdownHover) Print(sign string, comment string, posRef string) protocol.MarkupContent {
	// make line breaks conform to markdown spec
	comment = strings.ReplaceAll(comment, "\n", "  \n")
	return protocol.MarkupContent{Kind: "markdown", Value: "```typescript\n" + string(sign) + "\n```\n - - -\n" + comment + "\n - - -\n" + posRef}
}

func (md *MarkdownHover) LinkForNode(def *ttcn3.Definition) string {
	p := def.Position(def.Node.Pos())
	return fmt.Sprintf("[module %s](%s#L%dC%d)", def.ModuleOf(def.Node).Name.String(), def.Filename(), p.Line, p.Column)
}

func (pt *PlainTextHover) Print(sign string, comment string, posRef string) protocol.MarkupContent {
	val := sign
	log.Debugf("@Hover: %q, comment %q, posRef: %q\n", val, comment, posRef)
	if len(val) > 0 && val[len(val)-1] != '\n' {
		val += "\n"
	}
	if len(comment) > 0 {
		val += "__________\n" + comment
		log.Debugf("@Hover: %v\n", val)
		if val[len(val)-1] != '\n' {
			val += "\n"
		}
	}
	if len(posRef) > 0 {
		val += strings.Repeat("_", len(posRef)+2) + "\n" + posRef
	}
	return protocol.MarkupContent{Kind: "plaintext", Value: val}
}

func (pt *PlainTextHover) LinkForNode(def *ttcn3.Definition) string {
	return fmt.Sprintf("[module %s]", def.ModuleOf(def.Node).Name.String())
}

func getSignature(def *ttcn3.Definition) string {
	var sig bytes.Buffer
	var prefix = ""
	fh := fs.Open(def.Filename())
	content, _ := fh.Bytes()
	switch node := def.Node.(type) {
	case *ast.FuncDecl:
		sig.WriteString(node.Kind.String() + " " + node.Name.String())
		sig.Write(content[node.Params.Pos()-1 : node.Params.End()])
		if node.RunsOn != nil {
			sig.WriteString("\n  ")
			sig.Write(content[node.RunsOn.Pos()-1 : node.RunsOn.End()])
		}
		if node.System != nil {
			sig.WriteString("\n  ")
			sig.Write(content[node.System.Pos()-1 : node.System.End()])
		}
		if node.Return != nil {
			sig.WriteString("\n  ")
			sig.Write(content[node.Return.Pos()-1 : node.Return.End()])
		}
	case *ast.ValueDecl, *ast.TemplateDecl, *ast.FormalPar, *ast.StructTypeDecl, *ast.ComponentTypeDecl, *ast.EnumTypeDecl, *ast.PortTypeDecl:
		sig.Write(content[def.Node.Pos()-1 : def.Node.End()])
	case *ast.Field:
		if parent := def.ParentOf(node); parent != nil {
			if _, ok := parent.(*ast.SubTypeDecl); ok {
				prefix = "type "
			}
		}
		sig.Write(content[def.Node.Pos()-1 : def.Node.End()])
	case *ast.Module:
		fmt.Fprintf(&sig, "module %s\n", node.Name)
	default:
		log.Debugf("getSignature: unknown Type:%T\n", node)
	}
	return prefix + sig.String()
}

func (s *Server) hover(ctx context.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	var (
		file      = string(params.TextDocument.URI.SpanURI())
		line      = int(params.Position.Line) + 1
		col       = int(params.Position.Character) + 1
		comment   string
		signature string
		posRef    string
		defFound  = false
	)

	tree := ttcn3.ParseFile(file)
	x := tree.ExprAt(tree.Pos(line, col))
	if x == nil {
		return nil, nil
	}

	for _, def := range tree.LookupWithDB(x, &s.db) {
		defFound = true

		if firstTok := ast.FirstToken(def.Node); firstTok == nil {
			continue
		} else {
			comment = firstTok.Comments()
			signature = getSignature(def)
			if tree.Root != def.Root {
				posRef = s.clientCapability.HoverContent.LinkForNode(def)
			}
		}
	}
	if !defFound {
		// look for predefined functions
		if id := ast.Name(x); len(id) > 0 {
			for _, predef := range PredefinedFunctions {
				if predef.Label == id+"(...)" {
					comment = predef.Documentation
					signature = predef.Signature
				}
			}
		}
	}

	hoverContents := s.clientCapability.HoverContent.Print(signature, comment, posRef)
	hover := &protocol.Hover{Contents: hoverContents}

	return hover, nil
}
