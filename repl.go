package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/nokia/ntt/interpreter"
	"github.com/nokia/ntt/runtime"
	"github.com/nokia/ntt/ttcn3/syntax"
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

		root, _, _ := syntax.Parse([]byte(s.Text()))
		if err := root.Err(); err != nil {
			fmt.Println(err.Error())
			continue
		}
		if result := interpreter.Eval(root, env); result != nil {
			fmt.Println(result.Inspect())
		}
	}
	return nil
}
