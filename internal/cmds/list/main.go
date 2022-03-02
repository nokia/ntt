package list

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

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

		// Listing tests is the default command
		RunE: list,
	}

	w = bufio.NewWriter(os.Stdout)

	showFiles   = false
	showTags    = false
	formatJSON  = false
	formatPlain = true
	first       = true

	Basket ntt2.Basket
)

func init() {
	flags := Command.PersistentFlags()
	flags.BoolVar(&showFiles, "with-filename", false, "Print the filename for each match.")
	flags.BoolVar(&showTags, "with-tags", false, "Print documentation tags for each match.")
	flags.BoolVarP(&showTags, "tags", "t", false, "Print documentation tags for each match.")
	flags.MarkDeprecated("tags", "please use --with-tags instead")
	Command.PersistentFlags().AddFlagSet(ntt2.BasketFlags())
	Command.AddCommand(
		&cobra.Command{Use: `tests`, RunE: list},
		&cobra.Command{Use: `modules`, RunE: list},
		&cobra.Command{Use: `imports`, RunE: list},
		&cobra.Command{Use: `controls`, RunE: list},
		&cobra.Command{Use: `modulepars`, RunE: list},
	)

}

func list(cmd *cobra.Command, args []string) error {

	var err error
	Basket, err = ntt2.NewBasketWithFlags("list", cmd.Flags())
	Basket.LoadFromEnv("NTT_LIST_BASKETS")
	if err != nil {
		return err
	}

	suite, err := ntt.NewFromArgs(args...)
	if err != nil {
		return err
	}

	formatJSON, _ = cmd.Flags().GetBool("json")
	formatPlain, _ = cmd.Flags().GetBool("plain")

	if formatJSON {
		fmt.Fprintln(w, "[")
	}

	files, err := filesOfInterest(cmd.Use, suite)
	for _, f := range files {
		tree := ttcn3.ParseFile(f)
		if tree.Err != nil {
			return tree.Err
		}

		var module string
		ast.Inspect(tree.Root, func(n ast.Node) bool {
			if n != nil {
				switch n := n.(type) {
				case *ast.Module:
					module = ast.Name(n.Name)
					if cmd.Use == "modules" {
						Print(tree, n.Pos(), module, n.Tok.Comments())
						return false
					}
					return true
				case *ast.FuncDecl:
					if n.IsTest() && (cmd.Use == "list" || cmd.Use == "tests") {
						Print(tree, n.Pos(), module+"."+n.Name.String(), n.Kind.Comments())
					}
				case *ast.ImportDecl:
					if cmd.Use == "imports" {
						Print(tree, n.Pos(), fmt.Sprintf("%s\t%s", module, n.Module.String()), n.ImportTok.Comments())
					}
				case *ast.ControlPart:
					if cmd.Use == "controls" {
						Print(tree, n.Pos(), module+".control", ast.FirstToken(n).Comments())
					}
				case *ast.Declarator:
					if cmd.Use == "modulepars" {
						Print(tree, n.Pos(), module+"."+n.Name.String(), ast.FirstToken(n).Comments())
					}
				case *ast.ValueDecl:
					if n.Kind.Kind == token.MODULEPAR || n.Kind.Kind == token.ILLEGAL {
						return true
					}
				case *ast.NodeList, *ast.ModuleDef, *ast.GroupDecl, *ast.ModuleParameterGroup:
					return true

				}
			}
			return false
		})
	}

	if formatJSON {
		fmt.Fprintln(w, "]")
	}
	w.Flush()
	return err
}

type Match struct {
	Filename string `json:"filename,omitempty"`
	Line     int    `json:"line,omitempty"`
	Column   int    `json:"column,omitempty"`
	ID       string `json:"id,omitempty"`
	Tags     []Tag  `json:"tags,omitempty"`
}

type Tag struct {
	Key   string
	Value string
}

func (t Tag) String() string {
	if t.Value != "" {
		return fmt.Sprintf("%s:%s", t.Key, t.Value)
	}
	return t.Key
}
func (t Tag) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

func Print(tree *ttcn3.Tree, pos loc.Pos, id string, comments string) {
	tags := doc.FindAllTags(comments)
	if !Basket.Match(id, tags) {
		return
	}

	p := tree.Position(pos)

	var prettyTags []Tag
	for _, tag := range tags {
		t := Tag{Key: tag[0]}
		if len(tag) > 1 {
			t.Value = tag[1]
		}
		prettyTags = append(prettyTags, t)
	}
	switch {
	case formatJSON:
		if !first {
			w.Write([]byte(",\n"))
		}
		first = false
		b, err := json.Marshal(Match{
			Filename: p.Filename,
			Line:     p.Line,
			Column:   p.Column,
			ID:       id,
			Tags:     prettyTags,
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		w.Write(b)
	default:
		if showTags && len(tags) != 0 {
			for _, tag := range prettyTags {
				PrintMatch(p, id, tag.String())
			}
			return
		}

		PrintMatch(p, id)
	}
}

func PrintMatch(pos loc.Position, id string, tags ...string) {
	if showFiles {
		fmt.Fprintf(w, "%s:%d\t", pos.Filename, pos.Line)
	}
	fmt.Fprintf(w, id)
	for _, t := range tags {
		fmt.Fprintf(w, "\t%s", t)
	}
	fmt.Fprintln(w)
}

func filesOfInterest(cmd string, p project.Interface) ([]string, error) {
	switch cmd {
	case "tests", "controls", "list":
		return p.Sources()
	default:
		return project.Files(p)
	}

}
