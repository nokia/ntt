package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/ttcn3/ast/eval"
	"github.com/nokia/ntt/internal/ttcn3/parser"
	"github.com/nokia/ntt/runtime"
)

func repl() error {

	env := runtime.NewEnv(nil)
	s := bufio.NewScanner(os.Stdin)
	fmt.Printf("ntt %s (%s, %s)\n", version, commit, date)
	for {
		fmt.Printf(">> ")
		if !s.Scan() {
			break
		}

		fset := loc.NewFileSet()
		root, err := parser.Parse(fset, "<stdin>", s.Text())
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		if result := eval.Eval(root, env); result != nil {
			fmt.Println(result.Inspect())
		}
	}
	return nil
}
