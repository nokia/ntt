package ntt

import (
	"github.com/nokia/ntt/ttcn3"
	"github.com/nokia/ntt/ttcn3/ast"
	"github.com/nokia/ntt/ttcn3/doc"
)

// GenerateIDs emits given IDs to a channel.
func GenerateIDs(ids []string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for _, id := range ids {
			out <- id
		}
	}()
	return out
}

// GenerateTestsWithBasket emits all test ids from given TTCN-3 files to a channel.
func GenerateTestsWithBasket(files []string, b Basket) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for _, src := range files {
			tree := ttcn3.ParseFile(src)
			for _, def := range tree.Funcs() {
				if n := def.Node.(*ast.FuncDecl); n.IsTest() {
					id := tree.QualifiedName(n.Name)
					tags := doc.FindAllTags(n.Kind.Comments())
					if b.Match(id, tags) {
						out <- id
					}
				}
			}
		}
	}()
	return out
}

// GenerateTests emits all test ids from given TTCN-3 files to a channel.
func GenerateTests(files []string) <-chan string {
	return GenerateTestsWithBasket(files, Basket{})
}

// GenerateControls emits all control function ids from given TTCN-3 files to a channel.
func GenerateControlsWithBasket(files []string, b Basket) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for _, src := range files {
			tree := ttcn3.ParseFile(src)
			for _, n := range tree.Controls() {
				id := tree.QualifiedName(n.Ident)
				tags := doc.FindAllTags(ast.FirstToken(n.Ident).Comments())
				if b.Match(id, tags) {
					out <- id
				}
			}

		}
	}()
	return out
}

// GenerateControls emits all control function ids from given TTCN-3 files to a channel.
func GenerateControls(files []string) <-chan string {
	return GenerateControlsWithBasket(files, Basket{})
}
