package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/nokia/ntt/project"
	"github.com/nokia/ntt/ttcn3"
	"github.com/nokia/ntt/ttcn3/syntax"
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
	files, err := project.Files(Project)
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
				pos := syntax.Begin(n)
				tags[i] = append(tags[i], NewTag(syntax.Name(n), pos.Filename, pos.Line, Kind(n)))
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

func Kind(n syntax.Node) string {
	switch n.(type) {
	case *syntax.Module:
		return "n"
	case *syntax.Field:
		return "t"
	case *syntax.PortTypeDecl:
		return "t"
	case *syntax.ComponentTypeDecl:
		return "c"
	case *syntax.StructTypeDecl:
		return "m"
	case *syntax.EnumTypeDecl:
		return "e"
	case *syntax.BehaviourTypeDecl:
		return "t"
	case *syntax.Declarator:
		return "v"
	case *syntax.FormalPar:
		return "v"
	case *syntax.TemplateDecl:
		return "d"
	case *syntax.FuncDecl:
		return "f"
	case *syntax.SignatureDecl:
		return "f"
	default:
		return "e"
	}
}
