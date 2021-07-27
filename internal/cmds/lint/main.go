package lint

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/ntt"
	"github.com/nokia/ntt/internal/ttcn3/ast"
	"github.com/nokia/ntt/internal/ttcn3/doc"
	"github.com/nokia/ntt/internal/ttcn3/token"
	"github.com/nokia/ntt/project"
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


Formatting Checks

    max_lines          Number of lines a behaviour body must not exceed.
    aligned_braces     Braces must be in the same column or same line.
    require_case_else  Every select-statement must have one case-else.


Cyclomatic Complexity Checks

    complexity.max           Cyclomatic complexity muss not exceed.
    complexity.ignore_guards Ignore complexity of alt- and interleave guards


Naming Convention Checks

    naming.modules            Checks for module identifiers.
    naming.tests              Checks for test-case identifier.
    naming.functions          Checks for function identifiers.
    naming.altsteps           Checks for altstep identifiers.
    naming.parameters         Checks for parameter identifiers.
    naming.component_vars     Checks for component variable identifiers.
    naming.var_templates      Checks for variable template identifiers.
    naming.port_types         Checks for port type identifiers.
    naming.ports              Checks for port instance identifiers.
    naming.global_consts      Checks for global constant identifiers.
    naming.component_consts   Checks for component scoped constant identifiers.
    naming.templates          Checks for constant template identifiers.
    naming.locals             Checks for local variable identifiers.

    tags.tests                Checks for test-case tags.


White-Listing

    ignore.modules    Ignore modules
    ignore.files      Ignore files


Refactoring

When TTCN-3 code is refactored incrementally, it happens that references to
legacy code are faster added than one can remove them. This check helps with a
warning, as soon as the usage of a symbol exceed a defined limit.


Unused Symbols

    unused.modules    Checks for unused modules


Example configuration file:

	aligned_braces: true
	require_case_else: true
	max_lines: 40

	usage:
	  "foo":
	    limit: 12
	    text: Use "bar" instead.

	unused:
	  modules: true

	complexity:
	  max: 15
	  ignore_guards: true

	tags:
	  tests:
	    "@author": "testcases must have a @author tag"

	naming:
	  tests:
	    # An exlamation mark inverts the match.
	    "!.{130,}": "testcase identifiers must not be longer than 130 characters"

	  functions:
	    "^[a-z]"      : "function identifiers must begin with a lower case letter"
	    "!^(f|func)_" : "function identifiers must not begin with f_ or func_"

	  global_consts:
	    "^[A-Z0-9_]+$": "global constants must be UPPER_CASE"

	ignore:
	  modules:
	    # Ignore generated modules
	    - "^Protobuf_.+$"

	  files:
	    # Ignore all files from generated folders
	    - "generated/"

For information on writing new checks, see <TBD>.
`,

		RunE: lint,
	}

	config  string
	regexes = make(map[string]*regexp.Regexp)
	issues  = 0

	usedModules  = make(map[string]Import)
	usedModuleMu sync.Mutex

	style = struct {
		MaxLines        int  `yaml:"max_lines"`
		AlignedBraces   bool `yaml:"aligned_braces"`
		RequireCaseElse bool `yaml:"require_case_else"`
		Complexity      struct {
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
			VarTemplates    map[string]string `yaml:"var_templates"`
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
		Usage map[string]*struct {
			Text  string
			Limit int
			count int
		}
		Unused struct {
			Modules bool
		}
	}{}
)

type Import struct {
	Fset     *loc.FileSet
	Node     *ast.ImportDecl
	Path     string
	From     string
	Imported string
}

func init() {
	Command.PersistentFlags().StringVarP(&config, "config", "c", "", "path to YAML formatted file containing linter configuration")
}

func lint(cmd *cobra.Command, args []string) error {
	suite, err := ntt.NewFromArgs(args...)
	if err != nil {
		return err
	}

	c := suite.File(".ntt-lint.yml")
	b, err := c.Bytes()
	if err != nil {
		log.Verbose(err.Error())
		return nil
	}

	if err := yaml.UnmarshalStrict(b, &style); err != nil {
		return err
	}

	if err := buildRegexCache(); err != nil {
		return err
	}

	files, err := project.Files(suite)
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

			if isWhiteListed(style.Ignore.Modules, ast.Name(mod.Module.Name)) {
				return
			}

			stack := make([]ast.Node, 1, 64)
			cc := make(map[ast.Node]int)
			ccID := ast.Node(mod.Module)

			caseElse := make(map[ast.Node]int)
			var selectID *ast.SelectStmt

			ast.Inspect(mod.Module, func(n ast.Node) bool {
				if n == nil {
					stack = stack[:len(stack)-1]
					return false
				}

				stack = append(stack, n)
				fset := mod.FileSet

				switch n := n.(type) {
				case *ast.Ident:
					checkUsage(fset, n)

				case *ast.Module:
					checkNaming(fset, n, style.Naming.Modules)
					checkBraces(fset, n.LBrace, n.RBrace)

				case *ast.FuncDecl:
					ccID = n
					cc[ccID] = 1 // Intial McCabe value

					switch n.Kind.Kind {
					case token.TESTCASE:
						checkNaming(fset, n, style.Naming.Tests)
						checkTags(fset, n, style.Tags.Tests)
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
						checkNaming(fset, n, style.Naming.VarTemplates)
					case isVar(parent):
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
					selectID = n
					caseElse[selectID] = 0
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
					checkImport(fset, n, mod.Module)
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
					if isCaseElse(n) {
						caseElse[selectID]++
					} else {
						// Do not count case else for complexity
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
			checkCaseElse(mod.FileSet, caseElse)
		}(i)
	}

	wg.Wait()

	checkSuite(suite)

	switch issues {
	case 0:
		return nil
	case 1:
		return fmt.Errorf("1 issue found.")
	default:
		return fmt.Errorf("%d issues found.", issues)
	}

}

func checkNaming(fset *loc.FileSet, n ast.Node, patterns map[string]string) {
	checkPatterns(fset, n, patterns, ast.Name(n))
}

func checkTags(fset *loc.FileSet, n ast.Node, patterns map[string]string) {
	if len(patterns) == 0 {
		return
	}

	var tags []string
	for _, t := range doc.FindAllTags(ast.FirstToken(n).Comments()) {
		tags = append(tags, strings.Join(t, ":"))
	}

	checkPatterns(fset, n, patterns, tags...)
}

func checkPatterns(fset *loc.FileSet, n ast.Node, patterns map[string]string, ss ...string) {
next:
	for p, msg := range patterns {
		expect := true
		if strings.HasPrefix(p, "!") {
			expect = false
			p = p[1:]
		}

		// Match any.
		for _, s := range ss {
			if regexes[p].MatchString(s) == expect {
				continue next
			}
		}

		// If we could not match any, we report an error
		report(&errPattern{fset: fset, node: n, msg: msg})
	}
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

func checkCaseElse(fset *loc.FileSet, caseElse map[ast.Node]int) {
	if !style.RequireCaseElse {
		return
	}
	for n, v := range caseElse {
		if v == 0 {
			report(&errMissingCaseElse{fset: fset, node: n})
		}
	}
}

func checkUsage(fset *loc.FileSet, n *ast.Ident) {

	if style.Usage == nil {
		return
	}
	id := n.String()
	u, ok := style.Usage[id]
	if !ok {
		return
	}
	u.count++
	if u.count >= u.Limit {
		report(&errUsageExceedsLimit{
			fset:  fset,
			node:  n,
			usage: u.count,
			limit: u.Limit,
			text:  u.Text})
	}
}

func checkImport(fset *loc.FileSet, n *ast.ImportDecl, mod *ast.Module) {
	if !style.Unused.Modules {
		return
	}

	imported := ast.Name(n.Module)
	importing := ast.Name(mod.Name)

	usedModuleMu.Lock()
	usedModules[imported] = Import{
		Node:     n,
		Fset:     fset,
		Path:     fset.Position(n.Pos()).Filename,
		From:     importing,
		Imported: imported,
	}
	usedModuleMu.Unlock()

}

func checkSuite(suite *ntt.Suite) {

	if !style.Unused.Modules {
		return
	}

	pkgs, _ := suite.Imports()
	for _, pkg := range pkgs {
		files, _ := filepath.Glob(pkg + "/*.ttcn3")
		for _, file := range files {
			if isWhiteListed(style.Ignore.Files, file) {
				continue
			}

			mod := filepath.Base(file)
			mod = strings.TrimSuffix(mod, filepath.Ext(mod))

			if isWhiteListed(style.Ignore.Modules, mod) {
				return
			}

			if _, found := usedModules[mod]; !found {
				report(&errUnusedModule{file: file})
			}
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

	issues++
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

func isVar(d *ast.ValueDecl) bool {
	return !isVarTemplate(d) && d.Kind.Kind == token.VAR
}

func isVarTemplate(d *ast.ValueDecl) bool {
	return d.TemplateRestriction != nil
}

func isCaseElse(n *ast.CaseClause) bool {
	return n.Case == nil
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
	for p := range style.Naming.VarTemplates {
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
