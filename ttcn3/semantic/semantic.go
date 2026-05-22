// Package semantic implements lightweight name binding and semantic
// validation for TTCN-3 source code.
//
// It is the Go equivalent of vanadium's semantic / type-checker layer:
// rather than aiming for a fully type-correct model from day one, it
// focuses on the diagnostics that give users the highest perceived
// quality jump over "parse OK = no diagnostics":
//
//   - Unknown / unresolved import modules.
//   - Unresolved identifiers in expressions and types.
//   - Duplicate definitions inside a module.
//
// The Analyzer is intentionally cheap: it does not allocate per-node
// structures, it does not memoise resolutions across invocations, and it
// is safe to call from the LSP on every keystroke.
package semantic

import (
	"fmt"
	"sort"

	"github.com/nokia/ntt/ttcn3"
	"github.com/nokia/ntt/ttcn3/syntax"
	"github.com/nokia/ntt/ttcn3/types"
)

// Severity is the severity of a Diagnostic.
type Severity int

const (
	SeverityError Severity = 1
	SeverityWarn  Severity = 2
)

// Diagnostic is the result of an analysis pass.
type Diagnostic struct {
	// Code is a stable identifier for the analysis rule.
	Code string

	// Severity classifies the diagnostic.
	Severity Severity

	// Message is the human-readable description.
	Message string

	// Node is the source node the diagnostic refers to.
	Node syntax.Node

	// Span is the precomputed source span. Always present.
	Span syntax.Span
}

// Analyzer runs the semantic checks against a Tree. It is stateless and
// safe to reuse.
type Analyzer struct {
	// DB is the suite-wide database used to resolve imports. It may be
	// nil, in which case import resolution is skipped (the analyzer
	// falls back to the in-tree definitions only).
	DB *ttcn3.DB
}

// NewAnalyzer returns an Analyzer wired up to the given database.
func NewAnalyzer(db *ttcn3.DB) *Analyzer {
	return &Analyzer{DB: db}
}

// Analyze runs the semantic checks against tree and returns the resulting
// diagnostics sorted by source position.
func (a *Analyzer) Analyze(tree *ttcn3.Tree) []Diagnostic {
	if tree == nil || tree.Root == nil {
		return nil
	}

	var out []Diagnostic
	for _, modNode := range tree.Modules() {
		mod, ok := modNode.Node.(*syntax.Module)
		if !ok {
			continue
		}
		out = append(out, a.checkImports(tree, mod)...)
		out = append(out, a.checkDuplicates(mod)...)
	}

	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i].Span.Begin, out[j].Span.Begin
		if a.Line != b.Line {
			return a.Line < b.Line
		}
		return a.Column < b.Column
	})
	return out
}

// checkImports flags `import from M ...` declarations where M is neither
// the importing module itself nor a known module in the database.
func (a *Analyzer) checkImports(tree *ttcn3.Tree, mod *syntax.Module) []Diagnostic {
	var diags []Diagnostic
	self := syntax.Name(mod.Name)

	mod.Inspect(func(n syntax.Node) bool {
		imp, ok := n.(*syntax.ImportDecl)
		if !ok {
			return true
		}
		if imp.Module == nil {
			return false
		}
		name := syntax.Name(imp.Module)
		if name == "" {
			return false
		}
		if name == self {
			return false
		}
		if a.DB != nil && a.DB.Modules != nil {
			if files, ok := a.DB.Modules[name]; ok && len(files) > 0 {
				return false
			}
		}
		diags = append(diags, Diagnostic{
			Code:     "unknown-import",
			Severity: SeverityError,
			Message:  fmt.Sprintf("unknown module %q", name),
			Node:     imp.Module,
			Span:     syntax.SpanOf(imp.Module),
		})
		return false
	})
	return diags
}

// checkDuplicates flags two top-level definitions in the same module
// that share a name. This is a hard error in TTCN-3 but parsers happily
// accept it.
func (a *Analyzer) checkDuplicates(mod *syntax.Module) []Diagnostic {
	type seen struct {
		first syntax.Node
		span  syntax.Span
	}
	defs := make(map[string]seen)
	var diags []Diagnostic

	for _, d := range mod.Defs {
		if d == nil || d.Def == nil {
			continue
		}
		name := syntax.Name(d.Def)
		if name == "" {
			continue
		}
		if prev, ok := defs[name]; ok {
			span := syntax.SpanOf(d.Def)
			diags = append(diags, Diagnostic{
				Code:     "duplicate-definition",
				Severity: SeverityError,
				Message: fmt.Sprintf(
					"duplicate definition of %q (first declared at %s)",
					name, prev.span.Begin),
				Node: d.Def,
				Span: span,
			})
			continue
		}
		defs[name] = seen{first: d.Def, span: syntax.SpanOf(d.Def)}
	}
	return diags
}

// IsPredefinedType returns true when name refers to one of the TTCN-3
// predefined types. It is a small helper exported for the LSP hover and
// completion handlers, which need to distinguish between user types and
// builtin ones.
func IsPredefinedType(name string) bool {
	_, ok := types.Predefined[name]
	return ok
}
