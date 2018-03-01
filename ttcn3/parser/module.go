package parser

import (
	"github.com/nokia/ntt/ttcn3/ast"
)

// parseModule parses a module definition
//
// Module   = "module" ID [Language] "{" Decls "}"
// Language = "language" STRING
//
func (p *parser) parseModule() *ast.Module {
	if p.trace {
		defer un(trace(p, "Module"))
	}
	return nil
}
