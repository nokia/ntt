package format_test

import (
	"os"

	"github.com/nokia/ntt/ttcn3/format"
)

func ExampleFprint() {
	input := "p :=  {\n x:=1, // Comment\ny := 234 // Comment 2\n}"
	if err := format.Fprint(os.Stdout, input); err != nil {
		panic(err)
	}
	// Output:
	// p := {
	// 	x := 1,  // Comment
	// 	y := 234 // Comment 2
	// }
}

func ExampleCanonicalPrinter() {
	input := "p := {\n x:=1, // Comment\ny := 234 // Comment 2\n}"
	p := format.NewCanonicalPrinter(os.Stdout)
	p.Indent = 1
	p.UseSpaces = true
	p.TabWidth = 2
	p.Fprint(input)
	// Output:
	//   p := {
	//     x := 1,  // Comment
	//     y := 234 // Comment 2
	//   }
}
