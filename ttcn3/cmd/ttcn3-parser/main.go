package main

import (
	"flag"
	"fmt"
	"github.com/nokia/ntt/ttcn3/syntax"
	"os"
)

var (
	trace = flag.Bool("t", false, "Trace parser")
)

func main() {
	flag.Parse()

	ret := 0

	for _, v := range flag.Args() {
		err := parse(v)
		if err != nil {
			ret = 1
		}
	}

	os.Exit(ret)
}

func parse(file string) error {
	mode := syntax.Mode(syntax.Trace)
	if !*trace {
		mode = 0
	}

	_, err := syntax.ParseModule(syntax.NewFileSet(), file, nil, mode, func(pos syntax.Position, msg string) {
		fmt.Fprintf(os.Stderr, "%s: error: %s\n", pos, msg)
	})

	return err
}
