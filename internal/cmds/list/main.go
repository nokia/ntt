package list

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/ntt"
	"github.com/nokia/ntt/internal/ttcn3/ast"
	"github.com/nokia/ntt/internal/ttcn3/doc"
	"github.com/spf13/cobra"
)

var (
	Command = &cobra.Command{
		Use:   "list",
		Short: "List various types of objects",
		Long: `List various types of objects.

List modules, imports or tests. The list command without any explicit
sub-commands will output tests.

List will not output objects from imported directories. If you need to list all
objects from a testsuite you currently have to pass .ttcn3 files as arguments.
Example:

    ntt list $(ntt show -- sources) $(find $(ntt show -- imports) -name \*.ttcn3)


You can use regular expressions to filter objects. If you pass multiple regular
expressions, all of them must match (AND). Example:

	$ cat example.ttcn3
	testcase foo() ...
	testcase bar() ...
	testcase foobar() ...
	...

	$ ntt list --regex=foo --regex=bar
	example.foobar

	$ ntt list --regex='foo|bar'
	example.foo
	example.bar
	example.foobar


Similarly, you can also specify regular expressions for documentation tags.
Example:

	$ cat example.ttcn3
	// @one
	// @two some-value
	testcase foo() ...

	// @two: some-other-value
	testcase bar() ...
	...

	$ ntt list --tags-regex=@one --tags-regex=@two
	example.foo

	$ ntt list --tags-regex='@two: some'
	example.foo
	example.bar

`,

		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			verbose, _ = cmd.Flags().GetBool("verbose")

			suite, err := ntt.NewFromArgs(args...)
			if err != nil {
				return err
			}

			var srcs []string

			switch cmd.Name() {
			case "tests", "list":
				var x []*ntt.File
				x, err = suite.Sources()
				srcs = ntt.PathSlice(x...)
			default:
				srcs, err = suite.Files()
			}

			if err != nil {
				return err
			}

			var wg sync.WaitGroup
			wg.Add(len(srcs))
			infos = make([]*ntt.ParseInfo, len(srcs))
			for i, src := range srcs {
				go func(i int, src string) {
					defer wg.Done()
					infos[i] = suite.Parse(src)
				}(i, src)
			}
			wg.Wait()
			return nil
		},

		PersistentPostRun: func(cmd *cobra.Command, args []string) { w.Flush() },

		// Listing tests is the default command
		RunE: listTests,
	}

	listTestsCmd   = &cobra.Command{Use: `tests`, RunE: listTests}
	listModulesCmd = &cobra.Command{Use: `modules`, RunE: listModules}
	listImportsCmd = &cobra.Command{Use: `imports`, RunE: listImports}

	w = bufio.NewWriter(os.Stdout)

	showTags = false
	verbose  = false
	infos    []*ntt.ParseInfo

	ItemsRegex   = []string{}
	ItemsExclude = []string{}
	TagsRegex    = []string{}
	TagsExclude  = []string{}
)

func init() {
	Command.PersistentFlags().BoolVarP(&showTags, "tags", "t", false, "enable output of testcase documentation tags")
	Command.PersistentFlags().StringSliceVarP(&ItemsRegex, "regex", "r", []string{}, "list objects matching regular * expression.")
	Command.PersistentFlags().StringSliceVarP(&ItemsExclude, "exclude", "x", []string{}, "exclude objects matching regular * expresion.")
	Command.PersistentFlags().StringSliceVarP(&TagsRegex, "tags-regex", "R", []string{}, "list objects with tags matching regular * expression")
	Command.PersistentFlags().StringSliceVarP(&TagsExclude, "tags-exclude", "X", []string{}, "exclude objects with tags matching * regular expression")
	Command.AddCommand(listTestsCmd, listModulesCmd, listImportsCmd)
}

func listTests(cmd *cobra.Command, args []string) error {
	for _, info := range infos {
		if info.Err != nil {
			return info.Err
		}
		ast.Inspect(info.Module, func(n ast.Node) bool {
			if n == nil {
				return false
			}
			switch n := n.(type) {
			case *ast.FuncDecl:
				if !n.IsTest() {
					break
				}

				name := info.Module.Name.String() + "." + n.Name.String()
				tags := doc.FindAllTags(n.Kind.Comments())
				if match(name, tags) {
					printItem(info.FileSet, n.Pos(), tags, name)
				}

			case *ast.Module, *ast.ModuleDef, *ast.GroupDecl:
				return true
			}
			return false
		})
	}
	return nil
}

func listModules(cmd *cobra.Command, args []string) error {
	for _, info := range infos {
		if info.Err != nil {
			return info.Err
		}
		name := info.Module.Name.String()
		tags := doc.FindAllTags(info.Module.Tok.Comments())
		if match(name, tags) {
			printItem(info.FileSet, info.Module.Pos(), tags, name)
		}
	}
	return nil
}

func listImports(cmd *cobra.Command, args []string) error {
	for _, info := range infos {
		if info.Err != nil {
			return info.Err
		}
		ast.Inspect(info.Module, func(n ast.Node) bool {
			if n == nil {
				return false
			}
			switch n := n.(type) {
			case *ast.ImportDecl:
				name := n.Module.String()
				tags := doc.FindAllTags(n.ImportTok.Comments())
				if match(name, tags) {
					printItem(info.FileSet, n.Pos(), tags, info.Module.Name.String(), name)
				}

			case *ast.Module, *ast.ModuleDef, *ast.GroupDecl:
				return true
			}
			return false
		})
	}
	return nil
}

func printItem(fset *loc.FileSet, pos loc.Pos, tags [][]string, fields ...string) {

	if verbose {
		p := fset.Position(pos)
		fmt.Fprintf(w, "%s:%d\t", p.Filename, p.Line)
	}

	s := strings.Join(fields, "\t")

	if showTags && len(tags) != 0 {
		for _, tag := range tags {
			fmt.Fprintf(w, "%s\t%s\t%s\n", s, tag[0], tag[1])
		}
		return
	}

	fmt.Fprintf(w, "%s\n", s)
}

func match(name string, tags [][]string) bool {
	if !matchAll(ItemsRegex, name) {
		return false
	}
	if len(ItemsExclude) > 0 && matchAll(ItemsExclude, name) {
		return false
	}

	if len(TagsRegex) > 0 {
		if len(tags) == 0 {
			return false
		}
		if !matchAllTags(TagsRegex, tags) {
			return false
		}
	}

	if len(TagsExclude) > 0 && matchAllTags(TagsExclude, tags) {
		return false
	}

	return true
}

func matchAll(regexes []string, s string) bool {
	for _, r := range regexes {
		if ok, _ := regexp.Match(r, []byte(s)); !ok {
			return false
		}
	}
	return true
}

func matchAllTags(regexes []string, tags [][]string) bool {
next:
	for _, r := range regexes {
		f := strings.SplitN(r, ":", 2)
		for i := range f {
			f[i] = strings.TrimSpace(f[i])
		}
		for _, tag := range tags {
			if ok, _ := regexp.Match(f[0], []byte(tag[0])); !ok {
				continue
			}
			if len(f) > 1 {
				if ok, _ := regexp.Match(f[1], []byte(tag[1])); !ok {
					continue
				}
			}
			continue next
		}
		return false
	}
	return true
}
