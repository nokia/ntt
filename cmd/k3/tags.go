package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/nokia/ntt/internal/ntt"
	"github.com/nokia/ntt/internal/ttcn3/ast"
	"github.com/spf13/cobra"
)

var (
	tagsCmd = &cobra.Command{
		Use:   "tags",
		Short: "tags generates an index file for TTCN-3 language objects",
		Long: `tags  generates an index (or "tag") file for TTCN-3 language objects found in file(s). 

This tag file allows these items to be quickly and easily located by a text
editor or other utility. A "tag" signifies a language object for which an index
entry is available (or, alternatively, the index entry created for that
object).`,

		Run: tags,
	}
)

func init() {
	rootCmd.AddCommand(tagsCmd)
}

func tags(cmd *cobra.Command, args []string) {
	suite, err := ntt.NewFromArgs(args...)
	if err != nil {
		fatal(err)
	}
	files, err := suite.Files()
	if err != nil {
		fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(len(files))

	mods := make([]*ntt.ParseInfo, 0, len(files))

	for i := range files {
		go func(i int) {
			mods[i] = suite.Parse(files[i])
			wg.Done()
		}(i)
	}

	wg.Wait()

	for i := range mods {
		if mods[i] == nil || mods[i].Module == nil {
			continue
		}
		ast.Inspect(mods[i].Module, func(n ast.Node) bool {
			pos := mods[i].Position(n.Pos())
			file := pos.Filename
			line := pos.Line

			switch n := n.(type) {
			case *ast.ImportDecl:
				name := n.Module.String()
				if file, _ := suite.FindModule(name); file != "" {
					fmt.Fprintf(os.Stdout, "%s\t%s\t%d\tn\n", name, file, line)
				}
				return false

			case *ast.FriendDecl:
				name := n.Module.String()
				if file, _ := suite.FindModule(name); file != "" {
					fmt.Fprintf(os.Stdout, "%s\t%s\t%d\tn\n", name, file, line)
				}
				return false

			case *ast.SubTypeDecl:
				fmt.Fprintf(os.Stdout, "%s\t%s\t%d\tt\n", n.Field.Name, file, line)
				return false

			case *ast.PortTypeDecl:
				fmt.Fprintf(os.Stdout, "%s\t%s\t%d\tt\n", n.Name, file, line)
				return false

			case *ast.ComponentTypeDecl:
				fmt.Fprintf(os.Stdout, "%s\t%s\t%d\tc\n", n.Name.String(), file, line)
				return false

			case *ast.StructTypeDecl:
				fmt.Fprintf(os.Stdout, "%s\t%s\t%d\tm\n", n.Name.String(), file, line)
				return false

			case *ast.EnumTypeDecl:
				fmt.Fprintf(os.Stdout, "%s\t%s\t%d\te\n", n.Name.String(), file, line)
				return false

			case *ast.BehaviourTypeDecl:
				fmt.Fprintf(os.Stdout, "%s\t%s\t%d\tt\n", n.Name.String(), file, line)
				return false

			case *ast.ValueDecl:
				ast.Declarators(n.Decls, mods[i].FileSet, func(decl ast.Expr, id ast.Node, arrays []ast.Expr, value ast.Expr) {
					name := identName(id)
					pos := mods[i].Position(decl.Pos())
					file := pos.Filename
					line := pos.Line
					fmt.Fprintf(os.Stdout, "%s\t%s\t%d\tv\n", name, file, line)
				})
				return false

			case *ast.FormalPar:
				ast.Declarators([]ast.Expr{n.Name}, mods[i].FileSet, func(decl ast.Expr, id ast.Node, arrays []ast.Expr, value ast.Expr) {
					name := identName(id)
					pos := mods[i].Position(decl.Pos())
					file := pos.Filename
					line := pos.Line
					fmt.Fprintf(os.Stdout, "%s\t%s\t%d\tv\n", name, file, line)
				})
				return false

			case *ast.TemplateDecl:
				fmt.Fprintf(os.Stdout, "%s\t%s\t%d\td\n", n.Name.String(), file, line)
				return false

			case *ast.FuncDecl:
				fmt.Fprintf(os.Stdout, "%s\t%s\t%d\tf\n", n.Name.String(), file, line)
				return true

			case *ast.SignatureDecl:
				fmt.Fprintf(os.Stdout, "%s\t%s\t%d\tf\n", n.Name.String(), file, line)
				return false

			default:
				return true
			}
		})

	}
}

func identName(n ast.Node) string {
	switch n := n.(type) {
	case *ast.Ident:
		return n.String()
	case *ast.ParametrizedIdent:
		return n.Ident.String()
	}
	return "_"
}
