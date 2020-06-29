package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
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

	mods := make([]*ntt.ParseInfo, len(files))

	for i := range files {
		go func(i int) {
			mods[i] = suite.Parse(files[i])
			wg.Done()
		}(i)
	}

	wg.Wait()

	tags := make(map[string]struct{})

	for i := range mods {
		if mods[i] == nil || mods[i].Module == nil {
			continue
		}
		ast.Inspect(mods[i].Module, func(n ast.Node) bool {
			if n == nil {
				return false
			}

			pos := mods[i].Position(n.Pos())
			file := pos.Filename
			line := pos.Line

			switch n := n.(type) {
			case *ast.ImportDecl:
				name := n.Module.String()
				if file, _ := suite.FindModule(name); file != "" {
					tags[fmt.Sprintf("%s\t%s\t%d;\"\tn", name, file, 1)] = struct{}{}
				}
				return false

			case *ast.FriendDecl:
				name := n.Module.String()
				if file, _ := suite.FindModule(name); file != "" {
					tags[fmt.Sprintf("%s\t%s\t%d;\"\tn", name, file, 1)] = struct{}{}
				}
				return false

			case *ast.SubTypeDecl:
				tags[fmt.Sprintf("%s\t%s\t%d;\"\tt", n.Field.Name, file, line)] = struct{}{}
				return false

			case *ast.PortTypeDecl:
				tags[fmt.Sprintf("%s\t%s\t%d;\"\tt", n.Name, file, line)] = struct{}{}
				return false

			case *ast.ComponentTypeDecl:
				tags[fmt.Sprintf("%s\t%s\t%d;\"\tc", n.Name.String(), file, line)] = struct{}{}
				return false

			case *ast.StructTypeDecl:
				tags[fmt.Sprintf("%s\t%s\t%d;\"\tm", n.Name.String(), file, line)] = struct{}{}
				return false

			case *ast.EnumTypeDecl:
				tags[fmt.Sprintf("%s\t%s\t%d;\"\te", n.Name.String(), file, line)] = struct{}{}
				return false

			case *ast.BehaviourTypeDecl:
				tags[fmt.Sprintf("%s\t%s\t%d;\"\tt", n.Name.String(), file, line)] = struct{}{}
				return false

			case *ast.ValueDecl:
				ast.Declarators(n.Decls, mods[i].FileSet, func(decl ast.Expr, id ast.Node, arrays []ast.Expr, value ast.Expr) {
					name := identName(id)
					pos := mods[i].Position(decl.Pos())
					file := pos.Filename
					line := pos.Line
					tags[fmt.Sprintf("%s\t%s\t%d;\"\tv", name, file, line)] = struct{}{}
				})
				return false

			case *ast.FormalPar:
				ast.Declarators([]ast.Expr{n.Name}, mods[i].FileSet, func(decl ast.Expr, id ast.Node, arrays []ast.Expr, value ast.Expr) {
					name := identName(id)
					pos := mods[i].Position(decl.Pos())
					file := pos.Filename
					line := pos.Line
					tags[fmt.Sprintf("%s\t%s\t%d;\"\tv", name, file, line)] = struct{}{}
				})
				return false

			case *ast.TemplateDecl:
				tags[fmt.Sprintf("%s\t%s\t%d;\"\td", n.Name.String(), file, line)] = struct{}{}
				return false

			case *ast.FuncDecl:
				tags[fmt.Sprintf("%s\t%s\t%d;\"\tf", n.Name.String(), file, line)] = struct{}{}
				return true

			case *ast.SignatureDecl:
				tags[fmt.Sprintf("%s\t%s\t%d;\"\tf", n.Name.String(), file, line)] = struct{}{}
				return false

			default:
				return true
			}
		})
	}

	lines := make([]string, 0, len(tags))
	for k := range tags {
		lines = append(lines, k)
	}
	sort.Strings(lines)

	w := bufio.NewWriter(os.Stdout)
	fmt.Fprintln(w, "!_TAG_FILE_FORMAT	2	//")
	fmt.Fprintln(w, "!_TAG_FILE_SORTED	1	/0=unsorted, 1=sorted/")
	fmt.Fprintln(w, "!_TAG_PROGRAM_NAME	ttcn3_ctags	//")
	fmt.Fprintln(w, "!_TAG_PROGRAM_VERSION	1.0	//")

	for i := range lines {
		fmt.Fprintln(w, lines[i])
	}
	w.Flush()
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
