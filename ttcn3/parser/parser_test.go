package parser

import (
	"github.com/nokia/ntt/ttcn3/token"
	"testing"
)

type Expect int

const (
	PASS Expect = iota
	FAIL
)

type Test struct {
	expect Expect
	input  string
}

func TestModules(t *testing.T) {
	modules := []Test{
		{PASS, `module m {}`},
		{PASS, `module m language "str1", "str2" {}`},
	}

	testParse(t, modules, func(p *parser) { p.parseModule() })
}

func TestWithStmts(t *testing.T) {
	withStmts := []Test{
		{PASS, `encode    "str";`},
		{PASS, `variant   "str";`},
		{PASS, `display   "str";`},
		{PASS, `extension "str";`},
		{PASS, `optional  "str";`},
		{PASS, `stepsize  "str";`},
		{PASS, `encode override        "str";`},
		{PASS, `encode @local          "str";`},
		{PASS, `encode @local          "str"."ruleA";`},
		{PASS, `encode ([-])           "str";`},
		{PASS, `encode (a[-])          "str";`},
		{PASS, `encode (group all)     "str";`},
		{PASS, `encode (type all)      "str";`},
		{PASS, `encode (template all)  "str";`},
		{PASS, `encode (const all)     "str";`},
		{PASS, `encode (altstep all)   "str";`},
		{PASS, `encode (testcase all)  "str";`},
		{PASS, `encode (function all)  "str";`},
		{PASS, `encode (signature all) "str";`},
		{PASS, `encode (modulepar all) "str";`},
		{PASS, `encode (type all except {a,b}) "str";`}, // conflict with std
	}

	testParse(t, withStmts, func(p *parser) { p.parseWithStmt() })
}

func TestExprs(t *testing.T) {
	exprs := []Test{
		{PASS, `-`},
		{PASS, `a[-]`},
		{PASS, `-1 * x`},
		{PASS, `-x * y`},
		{PASS, `x := (1+2)*3, y:=a.f()`},
	}

	testParse(t, exprs, func(p *parser) { p.parseExprList() })
}

func TestFuncDecls(t *testing.T) {
	funcDecls := []Test{
		{PASS, `testcase f() {}`},
		{PASS, `testcase f() runs on A[-] {}`},
		{PASS, `testcase f() runs on C system C {}`},
		{PASS, `function f() {}`},
		{PASS, `function f() return int {}`},
		{PASS, `function f() return template int {}`},
		{PASS, `function f() return template(value) int {}`},
		{PASS, `function f() return value int {}`},
		{PASS, `function f @deterministic () {}`},
		{PASS, `function f() runs on A[-] {}`},
		{PASS, `function f() mtc C {}`},
		{PASS, `function f() runs on C mtc C system C {}`},
	}

	testParse(t, funcDecls, func(p *parser) { p.parseFuncDecl() })
}

func TestModuleDefs(t *testing.T) {
	moduleDefs := []Test{
		{PASS, `import from m all;`},
		{PASS, `import from m language "str1", "str2" all;`},
		{PASS, `import from m all except {}`},
		{PASS, `import from m all except {
                        template  all;
                        const     all;
                        altstep   all;
                        testcase  all;
                        function  all;
                        signature all;
                        modulepar all;
                        import    all;
                        type      all }`},
		{PASS, `import from m all except { group all }`},
		{PASS, `import from m all except { group x,y }`},
		{PASS, `import from m {
                        template  all;
                        const     all;
                        altstep   all;
                        testcase  all;
                        function  all;
                        signature all;
                        modulepar all;
                        import    all;
                        type      all }`},
		{PASS, `import from m {
                        group x except { group all }, y }`},
	}
	testParse(t, moduleDefs, func(p *parser) { p.parseModuleDef() })
}

func TestValueDecls(t *testing.T) {
	valueDecls := []Test{
		{PASS, `const integer x;`},
		{PASS, `const int x := 1;`},
		{PASS, `const int x := 1, yi := 2;`},
		{PASS, `const int x[len] := 1, y := 2;`},
		{PASS, `const a[-] x := 1;`},
		{PASS, `const a[1] x[2][3] := x[4];`},
		{PASS, `var int x, y := 2, z;`},
		{PASS, `var template          int x;`},
		{PASS, `var template(omit)    int x;`},
		{PASS, `var template(value)   int x;`},
		{PASS, `var template(present) int x;`},
		{PASS, `var omit    int x;`},
		{PASS, `var value   int x;`},
		{PASS, `var present int x;`},
		{PASS, `var value @lazy int x;`},
		{PASS, `var value @lazy int x, y := ?;`},
		{PASS, `template int x := ?;`},
		{PASS, `template int x modifies y := ?;`},
		{PASS, `template int x(int i) := i;`},
		{PASS, `template @lazy int x := ?;`},
		{PASS, `template @lazy int x(int i) := i;`},
		{PASS, `template @lazy int  x(int i) modifies y := ?;`},
		{PASS, `template @lazy a[-] x(int i) modifies y := ?;`},
		{PASS, `template(omit)    int x := ?;`},
		{PASS, `template(value)   int x := ?;`},
		{PASS, `template(present) int x := ?;`},
		{FAIL, `timer x, y := 1.0, y;`}, // problem with expr list
		{FAIL, `port P x[len], y := 1, z := 2 ;`},
	}

	testParse(t, valueDecls, func(p *parser) { p.parseDecl() })
}

func TestFormalPars(t *testing.T) {
	formalPars := []Test{
		{PASS, `()`},
		{PASS, `(int y)`},
		{PASS, `(int x, int y)`},
		{PASS, `(in int x, out int y, inout int z)`},
		{PASS, `(in template(value) @fuzzy timer x := 1, out timer y)`},
		{PASS, `(out timer y, in template(value) @fuzzy timer x := 1)`},
		{PASS, `(out timer y := -, in value @fuzzy timer x := 1)`},
		{PASS, `(out timer y := -, in value timer x := (1,2,3))`},
	}
	testParse(t, formalPars, func(p *parser) { p.parseParameters() })
}

func TestStmts(t *testing.T) {
	stmts := []Test{
		{PASS, `repeat;`},
		{PASS, `break;`},
		{PASS, `continue;`},
		{PASS, `return;`},
		{PASS, `return x() * 1;`},
		{PASS, `label L1;`},
		{PASS, `goto L2;`},
		{PASS, `for (var int i := 0; i<23; i := i+1) {}`},
		{PASS, `for (i:=x; i<23; i:=i+1) {}`},
		{PASS, `while (23) {}`},
		{PASS, `do {} while (23);`},
		{PASS, `if (1) {}`},
		{PASS, `if (1) {} else {}`},
		{PASS, `if (1) {} else if (2) {} else {}`},
		{PASS, `select union (p.x()) { case(1) {} case else {}}`},
		{PASS, `select  (23) {case(1) {} case else {}}`},
		{PASS, `alt {}`},
		{PASS, `interleave {}`},
	}
	testParse(t, stmts, func(p *parser) { p.parseStmt() })
}

func testParse(t *testing.T, tests []Test, f func(p *parser)) {
	for _, tt := range tests {
		err := anyParse(tt.input, f)
		if tt.expect == PASS && err != nil {
			t.Errorf("Parse(%#q):\n\t%v\n\n", tt.input, err)
		}
		if tt.expect == FAIL && err == nil {
			t.Errorf("breakage vanished: Parse(%#q)", tt.input)
		}
	}
}

func anyParse(input string, f func(p *parser)) error {
	var p parser
	p.init(token.NewFileSet(), "", []byte(input), 0)
	f(&p)
	p.errors.Sort()
	return p.errors.Err()
}
