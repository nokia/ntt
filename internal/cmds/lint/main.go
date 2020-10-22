package lint

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"sync"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/ntt"
	"github.com/nokia/ntt/internal/ttcn3/ast"
	"github.com/nokia/ntt/internal/ttcn3/token"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var (
	Command = &cobra.Command{
		Use:   "lint",
		Short: "lint examines TTCN-3 source files and reports suspicious code",
		Long: `lint examines TTCN-3 source files and reports suspicious code.
It may find problems not caught by the compiler, but also constructs
considered "bad style".

Lint's exit code is non-zero for erroneous invocation of the tool or if a
problem was reported.

To list the available checks, run "ntt lint help":

    <none>

For details and flags of a particular check, run "ntt lint help <check>".

For information on writing a new check, see <TBD>.
`,

		RunE: lint,
	}

	config  string
	regexes = make(map[string]*regexp.Regexp)

	style = struct {
		MaxLines      int  `yaml:"max_lines"`
		AlignedBraces bool `yaml:"aligned_braces"`
		Complexity    struct {
			Max          int
			IgnoreGuards bool `yaml:"ignore_guards"`
		}
		Naming struct {
			Modules         map[string]string
			Tests           map[string]string
			Functions       map[string]string
			Altsteps        map[string]string
			Parameters      map[string]string
			ComponentVars   map[string]string `yaml:"component_vars"`
			PortTypes       map[string]string `yaml:"port_types"`
			Ports           map[string]string
			GlobalConsts    map[string]string `yaml:"global_consts"`
			ComponentConsts map[string]string `yaml:"component_consts"`
			Templates       map[string]string
			Locals          map[string]string
		}
		Tags struct {
			Tests map[string]string
		}
		Ignore struct {
			Modules []string
			Files   []string
		}
	}{}
)

func init() {
	Command.PersistentFlags().StringVarP(&config, "config", "c", "", "path to YAML formatted file containing linter configuration")
}

func lint(cmd *cobra.Command, args []string) error {
	if config != "" {
		b, err := ioutil.ReadFile(config)
		if err != nil {
			return err
		}

		if err := yaml.UnmarshalStrict(b, &style); err != nil {
			return err
		}
	}

	if err := buildRegexCache(); err != nil {
		return err
	}

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

	for i := range files {
		go func(i int) {
			defer wg.Done()

			if isWhiteListed(style.Ignore.Files, files[i]) {
				return
			}

			mod := suite.Parse(files[i])
			if mod == nil || mod.Module == nil {
				return
			}

			if isWhiteListed(style.Ignore.Modules, identName(mod.Module.Name)) {
				return
			}

			stack := make([]ast.Node, 1, 64)
			cc := make(map[ast.Node]int)
			ccID := ast.Node(mod.Module)

			ast.Inspect(mod.Module, func(n ast.Node) bool {
				if n == nil {
					stack = stack[:len(stack)-1]
					return false
				}

				stack = append(stack, n)
				fset := mod.FileSet

				switch n := n.(type) {
				case *ast.Module:
					checkNaming(fset, n, style.Naming.Modules)
					checkBraces(fset, n.LBrace, n.RBrace)

				case *ast.FuncDecl:
					ccID = n
					cc[ccID] = 1 // Intial McCabe value

					switch n.Kind.Kind {
					case token.TESTCASE:
						checkNaming(fset, n, style.Naming.Tests)
					case token.FUNCTION:
						checkNaming(fset, n, style.Naming.Functions)
					case token.ALTSTEP:
						checkNaming(fset, n, style.Naming.Altsteps)
					}

					checkLines(fset, n)

				case *ast.FormalPar:
					checkNaming(fset, n, style.Naming.Parameters)

					// We do not descent any further,
					// because we do not want to count
					// cyclomatic complexity for default
					// values.
					return false

				case *ast.PortTypeDecl:
					checkNaming(fset, n, style.Naming.PortTypes)
					checkBraces(fset, n.LBrace, n.RBrace)

				case *ast.Declarator:
					if len(stack) <= 2 {
						return true
					}

					// The parent of a declarator should
					// always be a ValueDecl.  If not, we
					// have some internal issues, it's okay
					// to panic then.
					parent := stack[len(stack)-2].(*ast.ValueDecl)
					scope := stack[:len(stack)-2]

					switch {
					case isPort(parent):
						checkNaming(fset, n, style.Naming.Ports)
					case isConst(parent):
						switch {
						case inGlobalScope(scope):
							checkNaming(fset, n, style.Naming.GlobalConsts)
						case inComponentScope(scope):
							checkNaming(fset, n, style.Naming.ComponentConsts)
						}
					case isVarTemplate(parent):
						checkNaming(fset, n, style.Naming.Templates)
					default:
						switch {
						case inComponentScope(scope):
							checkNaming(fset, n, style.Naming.ComponentVars)
						default:
							checkNaming(fset, n, style.Naming.Locals)
						}
					}

					return true

				case *ast.TemplateDecl:
					checkNaming(fset, n, style.Naming.Templates)

				case *ast.BlockStmt:
					checkBraces(fset, n.LBrace, n.RBrace)
				case *ast.CompositeLiteral:
					checkBraces(fset, n.LBrace, n.RBrace)
				case *ast.ExceptExpr:
					checkBraces(fset, n.LBrace, n.RBrace)
				case *ast.SelectStmt:
					checkBraces(fset, n.LBrace, n.RBrace)
				case *ast.StructSpec:
					checkBraces(fset, n.LBrace, n.RBrace)
				case *ast.EnumSpec:
					checkBraces(fset, n.LBrace, n.RBrace)
				case *ast.ModuleParameterGroup:
					checkBraces(fset, n.LBrace, n.RBrace)
				case *ast.StructTypeDecl:
					checkBraces(fset, n.LBrace, n.RBrace)
				case *ast.EnumTypeDecl:
					checkBraces(fset, n.LBrace, n.RBrace)
				case *ast.ImportDecl:
					checkBraces(fset, n.LBrace, n.RBrace)
				case *ast.GroupDecl:
					checkBraces(fset, n.LBrace, n.RBrace)
				case *ast.WithSpec:
					checkBraces(fset, n.LBrace, n.RBrace)
				case *ast.ParenExpr:
					if n.LParen.Kind == token.LBRACE {
						checkBraces(fset, n.LParen, n.RParen)
					}

				case *ast.ModuleDef:
					// Reset ID for counting cyclomatic complexity.
					ccID = mod.Module

				case *ast.BinaryExpr:
					if n.Op.Kind == token.AND || n.Op.Kind == token.OR {
						cc[ccID]++
					}

				case *ast.IfStmt:
					cc[ccID]++

				case *ast.CaseClause:
					// Do not count case else
					if n.Case != nil {
						cc[ccID]++
					}

				case *ast.CommClause:
					if style.Complexity.IgnoreGuards {
						return true
					}

					// Do not count else-guards
					if n.Else.IsValid() {
						return true
					}
					// Every AltGuard increases cyclomatic complexity.
					cc[ccID]++

					// Every AltGuard expressions also increases complexity.
					if n.X != nil {
						cc[ccID]++
					}

				}
				return true
			})

			checkComplexity(mod.FileSet, cc)
		}(i)
	}

	wg.Wait()
	return nil
}

func checkNaming(fset *loc.FileSet, n ast.Node, patterns map[string]string) {
	if len(patterns) == 0 {
		return
	}

	s := identName(n)
	for p, msg := range patterns {
		expect := true
		if strings.HasPrefix(p, "!") {
			expect = false
			p = p[1:]
		}

		if regexes[p].MatchString(s) != expect {
			report(&errNaming{fset: fset, node: n, msg: msg})
		}

	}
	return
}

func checkLines(fset *loc.FileSet, n ast.Node) {
	if style.MaxLines == 0 {
		return
	}

	begin := fset.Position(n.Pos())
	end := fset.Position(n.End())
	lines := end.Line - begin.Line
	if lines > style.MaxLines {
		report(&errLines{fset: fset, node: n, lines: lines})
	}

}

func checkBraces(fset *loc.FileSet, left ast.Node, right ast.Node) {
	if !style.AlignedBraces {
		return
	}

	p1 := fset.Position(left.Pos())
	p2 := fset.Position(right.Pos())
	if p1.Line != p2.Line && p1.Column != p2.Column {
		report(&errBraces{fset: fset, left: left, right: right})
	}
}

func checkComplexity(fset *loc.FileSet, cc map[ast.Node]int) {
	if style.Complexity.Max == 0 {
		return
	}

	for n, v := range cc {
		if v > style.Complexity.Max {
			report(&errComplexity{fset: fset, node: n, complexity: v})
		}

	}
}

func matchAny(patterns []string, s string) bool {
	for _, p := range patterns {

		expect := true
		if strings.HasPrefix(p, "!") {
			expect = false
			p = p[1:]
		}

		if regexes[p].MatchString(s) == expect {
			return true
		}
	}
	return false
}

func report(e error) {

	// Check if this error is silenced (with a NOLINT-directive for example).
	type silencer interface {
		IsSilent() bool
	}
	if e, ok := e.(silencer); ok && e.IsSilent() {
		return
	}

	fmt.Println(e.Error())
}

func isWhiteListed(list []string, s string) bool {
	if len(list) == 0 {
		return false
	}
	return matchAny(list, s)
}

func inComponentScope(stack []ast.Node) bool {
	for _, n := range stack {
		if _, ok := n.(*ast.ComponentTypeDecl); ok {
			return true
		}
	}
	return false
}

func inGlobalScope(stack []ast.Node) bool {
	for _, n := range stack {
		switch n.(type) {
		case *ast.Module, *ast.ModuleDef, *ast.GroupDecl, *ast.ModuleParameterGroup:
		default:
			return false
		}
	}

	return true
}

func isPort(d *ast.ValueDecl) bool {
	return d.Kind.Kind == token.PORT
}

func isConst(d *ast.ValueDecl) bool {
	return d.Kind.Kind == token.CONST
}

func isVarTemplate(d *ast.ValueDecl) bool {
	return d.TemplateRestriction != nil
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
	case *ast.FuncDecl:
		return n.Name.String()
	case *ast.Module:
		return n.Name.String()
	case *ast.FormalPar:
		return n.Name.String()
	case *ast.PortTypeDecl:
		return identName(n.Name)
	case *ast.Declarator:
		return n.Name.String()
	case *ast.TemplateDecl:
		return n.Name.String()
	}
	return "_"
}

func buildRegexCache() error {

	for p := range style.Naming.Modules {
		if err := cacheRegex(p); err != nil {
			return err
		}
	}
	for p := range style.Naming.Tests {
		if err := cacheRegex(p); err != nil {
			return err
		}
	}
	for p := range style.Naming.Functions {
		if err := cacheRegex(p); err != nil {
			return err
		}
	}
	for p := range style.Naming.Altsteps {
		if err := cacheRegex(p); err != nil {
			return err
		}
	}
	for p := range style.Naming.Parameters {
		if err := cacheRegex(p); err != nil {
			return err
		}
	}
	for p := range style.Naming.ComponentVars {
		if err := cacheRegex(p); err != nil {
			return err
		}
	}
	for p := range style.Naming.PortTypes {
		if err := cacheRegex(p); err != nil {
			return err
		}
	}
	for p := range style.Naming.Ports {
		if err := cacheRegex(p); err != nil {
			return err
		}
	}
	for p := range style.Naming.GlobalConsts {
		if err := cacheRegex(p); err != nil {
			return err
		}
	}
	for p := range style.Naming.ComponentConsts {
		if err := cacheRegex(p); err != nil {
			return err
		}
	}
	for p := range style.Naming.Templates {
		if err := cacheRegex(p); err != nil {
			return err
		}
	}
	for p := range style.Naming.Locals {
		if err := cacheRegex(p); err != nil {
			return err
		}
	}
	for p := range style.Tags.Tests {
		if err := cacheRegex(p); err != nil {
			return err
		}
	}
	for _, p := range style.Ignore.Modules {
		if err := cacheRegex(p); err != nil {
			return err
		}
	}
	for _, p := range style.Ignore.Files {
		if err := cacheRegex(p); err != nil {
			return err
		}
	}
	return nil
}

func cacheRegex(p string) error {
	if strings.HasPrefix(p, "!") {
		p = p[1:]
	}

	if _, ok := regexes[p]; !ok {
		r, err := regexp.Compile(p)
		if err != nil {
			return err
		}
		regexes[p] = r
	}
	return nil
}
