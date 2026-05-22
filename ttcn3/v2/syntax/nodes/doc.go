// Package nodes contains the schema-generated AST node types for the
// next-generation TTCN-3 syntax tree.
//
// The package is the Go equivalent of Vanadium's src/ast: every node
// is defined once in nodes.yaml and the generator under gen/ turns
// the schema into struct types, kind constants, a Visitor interface
// and a children accessor. This keeps the tree definition in one
// place instead of spread across nodes.go, nodes_gen.go and a maze
// of switch statements.
//
// The schema-driven design and the layout of nodes.yaml are adopted
// from Vanadium (https://github.com/makekryl/vanadium) by Mikhail
// Krylov, BSD-3. See THIRD_PARTY_NOTICES.md at the repository root.
//
// To regenerate the file after editing nodes.yaml:
//
//	cd ttcn3/v2/syntax/nodes && go generate
package nodes

//go:generate go run ./gen
