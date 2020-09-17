package tags

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/nokia/ntt/internal/ntt"
	"github.com/nokia/ntt/internal/ttcn3/ast"
	"github.com/spf13/cobra"
)

var (
	Command = &cobra.Command{
		Use:   "tags",
		Short: "tags generates an index file for TTCN-3 language objects",
		Long: `tags  generates an index (or "tag") file for TTCN-3 language objects found in file(s). 

This tag file allows these items to be quickly and easily located by a text
editor or other utility. A "tag" signifies a language object for which an index
entry is available (or, alternatively, the index entry created for that
object).`,

		RunE: tags,
	}

	w = bufio.NewWriter(os.Stdout)
)

func tags(cmd *cobra.Command, args []string) error {
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

				case *ast.Declarator:
					t = append(t, NewTag(n.Name.String(), file, line, "v"))
					return false

				case *ast.FormalPar:
					t = append(t, NewTag(n.Name.String(), file, line, "v"))
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

	var lines []string
	for i := range tags {
		lines = append(lines, tags[i]...)
	}

	sort.Strings(lines)

	w := bufio.NewWriter(os.Stdout)
	fmt.Fprintln(w, "!_TAG_FILE_FORMAT	2	//")
	fmt.Fprintln(w, "!_TAG_FILE_SORTED	1	/0=unsorted, 1=sorted/")
	fmt.Fprintln(w, "!_TAG_PROGRAM_NAME	ntt	//")
	fmt.Fprintln(w, strings.Join(lines, "\n"))
	w.Flush()
	return nil
}

func NewTag(name string, file string, line int, kind string) string {
	return fmt.Sprintf("%s\t%s\t%d;\"\t%s", name, file, line, kind)
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
