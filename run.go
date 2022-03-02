package ntt

import (
	"context"

	"github.com/nokia/ntt/ttcn3"
	"github.com/nokia/ntt/ttcn3/ast"
	"github.com/nokia/ntt/ttcn3/doc"
)

// GenerateIDs emits given IDs to a channel.
func GenerateIDs(ids ...string) <-chan string {
	return GenerateIDsWithContext(context.Background(), ids...)
}

// GenerateIDs emits given IDs to a channel.
func GenerateIDsWithContext(ctx context.Context, ids ...string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for _, id := range ids {
			out <- id
		}
	}()
	return out
}

// GenerateTests emits all test ids from given TTCN-3 files to a channel.
func GenerateTests(files ...string) <-chan string {
	return GenerateTestsWithContext(context.Background(), Basket{}, files...)
}

// GenerateTestsWithContext emits all test ids from given TTCN-3 files to a channel.
func GenerateTestsWithContext(ctx context.Context, b Basket, files ...string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for _, src := range files {
			tree := ttcn3.ParseFile(src)
			for _, def := range tree.Funcs() {
				if n := def.Node.(*ast.FuncDecl); n.IsTest() {
					if !generate(ctx, def, b, out) {
						return
					}
				}
			}
		}
	}()
	return out
}

// GenerateControls emits all control function ids from given TTCN-3 files to a channel.
func GenerateControls(files ...string) <-chan string {
	return GenerateControlsWithContext(context.Background(), Basket{}, files...)
}

// GenerateControls emits all control function ids from given TTCN-3 files to a channel.
func GenerateControlsWithContext(ctx context.Context, b Basket, files ...string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for _, src := range files {
			tree := ttcn3.ParseFile(src)
			for _, n := range tree.Controls() {
				if !generate(ctx, n, b, out) {
					return
				}
			}

		}
	}()
	return out
}

func generate(ctx context.Context, def *ttcn3.Definition, b Basket, c chan<- string) bool {
	id := def.Tree.QualifiedName(def.Ident)
	tags := doc.FindAllTags(ast.FirstToken(def.Node).Comments())
	if b.Match(id, tags) {
		select {
		case c <- id:
		case <-ctx.Done():
			return false
		}
	}
	return true
}
