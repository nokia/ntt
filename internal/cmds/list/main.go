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
	"github.com/nokia/ntt/project"
	"github.com/nokia/ntt/ttcn3"
	"github.com/nokia/ntt/ttcn3/ast"
	"github.com/nokia/ntt/ttcn3/doc"
	"github.com/nokia/ntt/ttcn3/token"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	Command = &cobra.Command{
		Use:   "list",
		Short: "List various types of objects",
		Long: `List various types of objects.

List control parts, modules, imports or tests. The list command without any explicit
sub-commands will output tests.

List will ignore imported directories when printing tests. If you need to list all
tests from a testsuite you'll have to pass .ttcn3 files as arguments.
Example:

    ntt list $(ntt show -- sources) $(find $(ntt show -- imports) -name \*.ttcn3)



Filtering
---------

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


Baskets
-------

Baskets are filters defined by environment variables of the form:

        NTT_LIST_BASKETS_<name> = <filters>

For example, to define a basket "stable" which excludes all objects with @wip
or @flaky tags:

	export NTT_LIST_BASKETS_stable="-X @wip|@flaky"

Baskets become active when they are listed in colon separated environment
variable NTT_LIST_BASKETS. If you specify multiple baskets, at least of them
must match (OR).

Rule of thumb: all baskets are ORed, all explicit filter options are ANDed.
Example:

	$ export NTT_LIST_BASKETS_stable="--tags-exclude @wip|@flaky"
	$ export NTT_LIST_BASKETS_ipv6="--tags-regex @ipv6"
	$ NTT_LIST_BASKETS=stable:ipv6 ntt list -R @flaky


Above example will output all tests with a @flaky tag and either @wip or @ipv6 tag.


If a basket is not defined by an environment variable, it's equivalent to a
"--tags-regex" filter. For example, to lists all tests, which have either a
@flaky or a @wip tag:

	# Note, flaky and wip baskets are not specified explicitly.
	$ NTT_LIST_BASKETS=flaky:wip ntt list

	# This does the same:
	$ ntt list --tags-regex="@wip|@flaky"

`,

		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			verbose, _ = cmd.Flags().GetBool("verbose")

			suite, err := ntt.NewFromArgs(args...)
			if err != nil {
				return err
			}

			if err := parseFiles(cmd, suite); err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
			}

			return loadBaskets(suite)

		},

		PersistentPostRun: func(cmd *cobra.Command, args []string) { w.Flush() },

		// Listing tests is the default command
		RunE: listTests,
	}

	listTestsCmd      = &cobra.Command{Use: `tests`, RunE: listTests}
	listModulesCmd    = &cobra.Command{Use: `modules`, RunE: listModules}
	listImportsCmd    = &cobra.Command{Use: `imports`, RunE: listImports}
	listControlsCmd   = &cobra.Command{Use: `controls`, RunE: listControls}
	listModuleParsCmd = &cobra.Command{Use: `modulepars`, RunE: listModulePars}

	w = bufio.NewWriter(os.Stdout)

	showTags = false
	verbose  = false
	trees    []*ttcn3.Tree

	baskets = []basket{{name: "default"}}
)

func init() {
	Command.PersistentFlags().BoolVarP(&showTags, "tags", "t", false, "enable output of testcase documentation tags")
	Command.PersistentFlags().StringSliceVarP(&baskets[0].nameRegex, "regex", "r", []string{}, "list objects matching regular * expression.")
	Command.PersistentFlags().StringSliceVarP(&baskets[0].nameExclude, "exclude", "x", []string{}, "exclude objects matching regular * expresion.")
	Command.PersistentFlags().StringSliceVarP(&baskets[0].tagsRegex, "tags-regex", "R", []string{}, "list objects with tags matching regular * expression")
	Command.PersistentFlags().StringSliceVarP(&baskets[0].tagsExclude, "tags-exclude", "X", []string{}, "exclude objects with tags matching * regular expression")
	Command.AddCommand(listTestsCmd, listModulesCmd, listImportsCmd, listControlsCmd, listModuleParsCmd)
}

type basket struct {
	name        string
	nameRegex   []string
	nameExclude []string
	tagsRegex   []string
	tagsExclude []string
}

func newBasket(name string, args []string) (basket, error) {
	b := basket{name: name}

	fs := pflag.NewFlagSet(name, pflag.ContinueOnError)
	fs.StringSliceVarP(&b.nameRegex, "regex", "r", []string{}, "list objects matching regular * expression.")
	fs.StringSliceVarP(&b.nameExclude, "exclude", "x", []string{}, "exclude objects matching regular * expresion.")
	fs.StringSliceVarP(&b.tagsRegex, "tags-regex", "R", []string{}, "list objects with tags matching regular * expression")
	fs.StringSliceVarP(&b.tagsExclude, "tags-exclude", "X", []string{}, "exclude objects with tags matching * regular expression")

	if err := fs.Parse(args); err != nil {
		return b, err
	}

	return b, nil
}

func loadBaskets(suite *ntt.Suite) error {
	env, err := suite.Getenv("NTT_LIST_BASKETS")
	if err != nil || env == "" {
		if _, ok := err.(*project.NoSuchVariableError); ok {
			return nil
		}
		return err
	}

	for _, name := range strings.Split(env, ":") {
		if name == "" {
			continue
		}

		flags, err := suite.Getenv("NTT_LIST_BASKETS_" + name)
		args := strings.Fields(flags)
		if err != nil {
			if _, ok := err.(*project.NoSuchVariableError); !ok {
				return err
			}
			args = []string{"-R", "@" + name}
		}

		basket, err := newBasket(name, args)
		if err != nil {
			return err
		}
		baskets = append(baskets, basket)
	}
	return nil
}

func parseFiles(cmd *cobra.Command, suite *ntt.Suite) error {
	var (
		srcs []string
		err  error
	)

	switch cmd.Name() {
	case "tests", "list":
		srcs, err = suite.Sources()
	default:
		srcs, err = project.Files(suite)
	}

	var wg sync.WaitGroup
	wg.Add(len(srcs))
	trees = make([]*ttcn3.Tree, len(srcs))
	for i, src := range srcs {
		go func(i int, src string) {
			defer wg.Done()
			trees[i] = ttcn3.ParseFile(src)
		}(i, src)
	}
	wg.Wait()
	return err
}

func listTests(cmd *cobra.Command, args []string) error {
	for _, tree := range trees {
		if tree.Err != nil {
			return tree.Err
		}

		var module string
		ast.Inspect(tree.Root, func(n ast.Node) bool {
			if n == nil {
				return false
			}
			switch n := n.(type) {
			case *ast.FuncDecl:
				if !n.IsTest() {
					break
				}

				name := module + "." + n.Name.String()
				tags := doc.FindAllTags(n.Kind.Comments())
				if match(name, tags) {
					printItem(tree.FileSet, n.Pos(), tags, name)
				}

			case *ast.Module:
				module = n.Name.String()
				return true
			case *ast.ModuleDef, *ast.GroupDecl:
				return true
			case *ast.NodeList:
				return true
			}
			return false
		})
	}
	return nil
}

func listModules(cmd *cobra.Command, args []string) error {
	for _, tree := range trees {
		for _, mod := range tree.Modules() {
			name := mod.Name.String()
			tags := doc.FindAllTags(mod.Tok.Comments())
			if match(name, tags) {
				printItem(tree.FileSet, mod.Pos(), tags, name)
			}
		}
	}
	return nil
}

func listImports(cmd *cobra.Command, args []string) error {
	for _, tree := range trees {
		if tree.Err != nil {
			return tree.Err
		}
		var module string
		ast.Inspect(tree.Root, func(n ast.Node) bool {
			if n == nil {
				return false
			}
			switch n := n.(type) {
			case *ast.ImportDecl:
				name := n.Module.String()
				tags := doc.FindAllTags(n.ImportTok.Comments())
				if match(name, tags) {
					printItem(tree.FileSet, n.Pos(), tags, module, name)
				}

			case *ast.Module:
				module = n.Name.String()
				return true
			case *ast.ModuleDef, *ast.GroupDecl:
				return true
			case *ast.NodeList:
				return true
			}
			return false
		})
	}
	return nil
}

func listControls(cmd *cobra.Command, args []string) error {
	for _, tree := range trees {
		if tree.Err != nil {
			return tree.Err
		}
		var module string
		ast.Inspect(tree.Root, func(n ast.Node) bool {
			if n == nil {
				return false
			}
			switch n := n.(type) {
			case *ast.ControlPart:
				name := module + ".control"
				tags := doc.FindAllTags(ast.FirstToken(n).Comments())
				if match(name, tags) {
					printItem(tree.FileSet, n.Pos(), tags, name)
				}

			case *ast.Module:
				module = n.Name.String()
				return true

			case *ast.ModuleDef, *ast.GroupDecl:
				return true
			}
			return false
		})
	}
	return nil
}

func listModulePars(cmd *cobra.Command, args []string) error {
	for _, tree := range trees {
		if tree.Err != nil {
			return tree.Err
		}
		var module string
		ast.Inspect(tree.Root, func(n ast.Node) bool {
			if n == nil {
				return false
			}
			switch n := n.(type) {
			case *ast.ValueDecl:
				if n.Kind.Kind == token.MODULEPAR || n.Kind.Kind == token.ILLEGAL {
					return true
				}

				return false

			case *ast.Declarator:
				name := module + "." + n.Name.String()
				tags := doc.FindAllTags(ast.FirstToken(n).Comments())
				if !match(name, tags) {
					return false
				}
				printItem(tree.FileSet, n.Pos(), tags, name)

			case *ast.Module:
				module = n.Name.String()
				return true

			case *ast.ModuleDef, *ast.GroupDecl, *ast.ModuleParameterGroup:
				return true
			}
			return false
		})
	}
	return nil
}

func printItem(fset *loc.FileSet, pos loc.Pos, tags [][]string, fields ...string) {

	s := strings.Join(fields, "\t")

	if showTags && len(tags) != 0 {
		for _, tag := range tags {
			if verbose {
				p := fset.Position(pos)
				fmt.Fprintf(w, "%s:%d\t", p.Filename, p.Line)
			}
			fmt.Fprintf(w, "%s\t%s\t%s\n", s, tag[0], tag[1])
		}
		return
	}

	if verbose {
		p := fset.Position(pos)
		fmt.Fprintf(w, "%s:%d\t", p.Filename, p.Line)
	}
	fmt.Fprintf(w, "%s\n", s)
}

func match(name string, tags [][]string) bool {
	ok := baskets[0].match(name, tags)
	if len(baskets) == 1 {
		return ok
	}

	for _, basket := range baskets[1:] {
		if basket.match(name, tags) && ok {
			return true
		}
	}
	return false
}

func (b *basket) match(name string, tags [][]string) bool {
	if !b.matchAll(b.nameRegex, name) {
		return false
	}
	if len(b.nameExclude) > 0 && b.matchAll(b.nameExclude, name) {
		return false
	}

	if len(b.tagsRegex) > 0 {
		if len(tags) == 0 {
			return false
		}
		if !b.matchAllTags(b.tagsRegex, tags) {
			return false
		}
	}

	if len(b.tagsExclude) > 0 && b.matchAllTags(b.tagsExclude, tags) {
		return false
	}

	return true
}

func (b *basket) matchAll(regexes []string, s string) bool {
	for _, r := range regexes {
		if ok, _ := regexp.Match(r, []byte(s)); !ok {
			return false
		}
	}
	return true
}

func (b *basket) matchAllTags(regexes []string, tags [][]string) bool {
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
