package printer_test

import (
	"os"

	"github.com/nokia/ntt/ttcn3/v2/printer"
)

func ExampleFprint() {
	input := "{\n x:=1, // Comment\ny := 234, // Comment 2\n}"
	if err := printer.Fprint(os.Stdout, input); err != nil {
		panic(err)
	}
	// Output:
	// {
	// 	x := 1,   // Comment
	// 	y := 234, // Comment 2
	// }
}

func ExampleCanonicalPrinter() {
	input := "{\n x:=1, // Comment\ny := 234, // Comment 2\n}"
	p := printer.NewCanonicalPrinter(os.Stdout)
	p.Indent = "  "
	p.Fprint(input)
	// Output:
	// {
	//   x := 1,   // Comment
	//   y := 234, // Comment 2
	// }
}
