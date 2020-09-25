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
			mod, nodes := suite.Tags(files[i])
			if mod == nil || mod.Module == nil {
				return
			}

			t := make([]string, 0, len(mod.Module.Defs)*2)
			for _, n := range nodes {
				pos := mod.Position(n.Pos())
				t = append(t, NewTag(ast.Name(n), pos.Filename, pos.Line, Kind(n)))
			}
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

func Kind(n ast.Node) string {
	switch n.(type) {
	case *ast.Module:
		return "n"
	case *ast.Field:
		return "t"
	case *ast.PortTypeDecl:
		return "t"
	case *ast.ComponentTypeDecl:
		return "c"
	case *ast.StructTypeDecl:
		return "m"
	case *ast.EnumTypeDecl:
		return "e"
	case *ast.BehaviourTypeDecl:
		return "t"
	case *ast.Declarator:
		return "v"
	case *ast.FormalPar:
		return "v"
	case *ast.TemplateDecl:
		return "d"
	case *ast.FuncDecl:
		return "f"
	case *ast.SignatureDecl:
		return "f"
	default:
		return "e"
	}
}
