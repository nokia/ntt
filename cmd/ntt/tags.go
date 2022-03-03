package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/nokia/ntt/internal/ntt"
	"github.com/nokia/ntt/project"
	"github.com/nokia/ntt/ttcn3"
	"github.com/nokia/ntt/ttcn3/ast"
	"github.com/spf13/cobra"
)

var (
	TagsCommand = &cobra.Command{
		Use:   "tags",
		Short: "Write a cTags database to stdout",
		Long: `tags  generates an index (or "tag") file for TTCN-3 language objects found in file(s). 

This tag file allows these items to be quickly and easily located by a text
editor or other utility. A "tag" signifies a language object for which an index
entry is available (or, alternatively, the index entry created for that
object).`,

		RunE: tags,
	}
)

func tags(cmd *cobra.Command, args []string) error {
	suite, err := ntt.NewFromArgs(args...)
	if err != nil {
		return err
	}
	files, err := project.Files(suite)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(len(files))

	tags := make([][]string, len(files))

	for i := range files {
		go func(i int) {
			defer wg.Done()
			tree := ttcn3.ParseFile(files[i])
			for _, n := range tree.Tags() {
				pos := tree.Position(n.Pos())
				tags[i] = append(tags[i], NewTag(ast.Name(n), pos.Filename, pos.Line, Kind(n)))
			}
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
