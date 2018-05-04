package main

import (
	"flag"
	"fmt"
	"github.com/nokia/ntt/ttcn3/parser"
	"github.com/nokia/ntt/ttcn3/token"
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
	mode := parser.Mode(parser.Trace)
	if !*trace {
		mode = 0
	}

	_, err := parser.ParseModule(token.NewFileSet(), file, nil, mode, func(pos token.Position, msg string) {
		fmt.Fprintf(os.Stderr, "%s: error: %s\n", pos, msg)
	})

	return err
}
