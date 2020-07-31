package main

import (
	"sync"

	"github.com/nokia/ntt/internal/ntt"
	"github.com/nokia/ntt/internal/ttcn3/ast"
	"github.com/nokia/ntt/plugin"
)

type request struct {
	plugin.GeneratorRequest
}

func build() {
	suite, err := ntt.NewFromArgs(args...)
	if err != nil {
		return err
	}
	files, err := suite.Files()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(len(files))

	tags := make([][]string, len(files))

	for i := range files {
		go func(i int) {
			defer wg.Done()
			mod := suite.Parse(files[i])
			if mod == nil || mod.Module == nil {
				return
			}

			t := make([]string, 0, len(mod.Module.Defs)*2)
			ast.Inspect(mod.Module, func(n ast.Node) bool {
				if n == nil {
					return false
				}

				pos := mod.Position(n.Pos())
				file := pos.Filename
				line := pos.Line

				switch n := n.(type) {
				case *ast.Module:
					t = append(t, NewTag(identName(n.Name), file, line, "n"))
					return true

				case *ast.ImportDecl:
					return false

				case *ast.FriendDecl:
					return false

				case *ast.Field:
					t = append(t, NewTag(identName(n.Name), file, line, "t"))
					return true

				case *ast.PortTypeDecl:
					t = append(t, NewTag(identName(n.Name), file, line, "t"))
					return false

				case *ast.ComponentTypeDecl:
					t = append(t, NewTag(n.Name.String(), file, line, "c"))
					return true

				case *ast.StructTypeDecl:
					t = append(t, NewTag(n.Name.String(), file, line, "m"))
					return true

				case *ast.EnumTypeDecl:
					t = append(t, NewTag(n.Name.String(), file, line, "e"))
					for _, e := range n.Enums {
						line := mod.Position(e.Pos()).Line
						name := identName(e)
						t = append(t, NewTag(name, file, line, "e"))
					}
					return false

				case *ast.EnumSpec:
					for _, e := range n.Enums {
						line := mod.Position(e.Pos()).Line
						name := identName(e)
						t = append(t, NewTag(name, file, line, "e"))
					}
					return false

				case *ast.BehaviourTypeDecl:
					t = append(t, NewTag(n.Name.String(), file, line, "t"))
					return false

				case *ast.ValueDecl:
					ast.Declarators(n.Decls, mod.FileSet, func(decl ast.Expr, id ast.Node, arrays []ast.Expr, value ast.Expr) {
						name := identName(id)
						pos := mod.Position(decl.Pos())
						file := pos.Filename
						line := pos.Line
						t = append(t, NewTag(name, file, line, "v"))
					})
					return false

				case *ast.FormalPar:
					ast.Declarators([]ast.Expr{n.Name}, mod.FileSet, func(decl ast.Expr, id ast.Node, arrays []ast.Expr, value ast.Expr) {
						name := identName(id)
						pos := mod.Position(decl.Pos())
						file := pos.Filename
						line := pos.Line
						t = append(t, NewTag(name, file, line, "v"))
					})
					return false

				case *ast.TemplateDecl:
					t = append(t, NewTag(n.Name.String(), file, line, "d"))
					return true

				case *ast.FuncDecl:
					t = append(t, NewTag(n.Name.String(), file, line, "f"))
					return true

				case *ast.SignatureDecl:
					t = append(t, NewTag(n.Name.String(), file, line, "f"))
					return false

				default:
					return true
				}
			})
			tags[i] = t

		}(i)
	}

	wg.Wait()
}

func identName(n ast.Node) string {
	switch n := n.(type) {
	case *ast.CallExpr:
		return identName(n.Fun)
	case *ast.LengthExpr:
		return identName(n.X)
	case *ast.Ident:
		return n.String()
	case *ast.ParametrizedIdent:
		return n.Ident.String()
	}
	return "_"
}
