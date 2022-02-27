package list

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"

	ntt2 "github.com/nokia/ntt"
	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/ntt"
	"github.com/nokia/ntt/project"
	"github.com/nokia/ntt/ttcn3"
	"github.com/nokia/ntt/ttcn3/ast"
	"github.com/nokia/ntt/ttcn3/doc"
	"github.com/nokia/ntt/ttcn3/token"
	"github.com/spf13/cobra"
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

	Basket = ntt2.DefaultBasket
)

func init() {
	Command.PersistentFlags().BoolVarP(&showTags, "tags", "t", false, "enable output of testcase documentation tags")
	Command.AddCommand(listTestsCmd, listModulesCmd, listImportsCmd, listControlsCmd, listModuleParsCmd)
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

		sb, err := ntt2.NewBasket(name, args...)
		if err != nil {
			return err
		}
		Basket.Baskets = append(Basket.Baskets, sb)
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
				if Basket.Match(name, tags) {
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
			name := mod.Ident.String()
			tags := doc.FindAllTags(mod.Tok.Comments())
			if Basket.Match(name, tags) {
				printItem(tree.FileSet, mod.Ident.Pos(), tags, name)
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
				if Basket.Match(name, tags) {
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
				if Basket.Match(name, tags) {
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
				if !Basket.Match(name, tags) {
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
