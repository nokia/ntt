// Package types implements a type system for TTCN-3.
package types

import (
	"fmt"
	"strings"

	"github.com/nokia/ntt/ttcn3/syntax"
)

var Predefined = map[string]Type{
	"any":                  &PrimitiveType{Kind: Any, Name: "any"},
	"integer":              &PrimitiveType{Kind: Integer, Name: "integer"},
	"float":                &PrimitiveType{Kind: Float, Name: "float"},
	"boolean":              &PrimitiveType{Kind: Boolean, Name: "boolean"},
	"verdicttype":          &PrimitiveType{Kind: Verdict, Name: "verdicttype"},
	"bitstring":            &ListType{Kind: Bitstring, Name: "bitstring"},
	"charstring":           &ListType{Kind: Charstring, Name: "charstring"},
	"hexstring":            &ListType{Kind: Hexstring, Name: "hexstring"},
	"octetstring":          &ListType{Kind: Octetstring, Name: "octetstring"},
	"universal charstring": &ListType{Kind: UniversalCharstring, Name: "universal charstring"},
	"default":              &BehaviourType{Kind: Altstep, Name: "default"},
	"timer":                &StructuredType{Kind: Object, Name: "timer"},
	"anytype":              &StructuredType{Kind: Union, Name: "anytype"},
}

func init() {
	// The element type of string types is the string type itself.
	Predefined["bitstring"].(*ListType).ElementType = Predefined["bitstring"]
	Predefined["hexstring"].(*ListType).ElementType = Predefined["hexstring"]
	Predefined["octetstring"].(*ListType).ElementType = Predefined["octetstring"]
	Predefined["charstring"].(*ListType).ElementType = Predefined["charstring"]
	Predefined["universal charstring"].(*ListType).ElementType = Predefined["universal charstring"]
}

// Kind is a bitmask describing the kinds of values a type can have, such as
// "integer" or "record".
type Kind uint64

const (
	Any   Kind = 0
	Error Kind = 1 << iota
	Incomplete

	// Primitive types
	Boolean
	Enumerated
	Float
	Integer
	Verdict

	// List types
	Array
	Bitstring
	Charstring
	Hexstring
	Map
	Octetstring
	RecordOf
	SetOf
	UniversalCharstring

	// Structured types
	Anytype
	Component
	Object
	Port
	Record
	Set
	Timer
	Trait
	Union

	// Behaviour types
	Altstep
	Configuration
	Default
	Function
	Testcase
)

var kindNames = map[Kind]string{
	Error:               "error",
	Incomplete:          "incomplete",
	Boolean:             "boolean",
	Enumerated:          "enumerated",
	Float:               "float",
	Integer:             "integer",
	Verdict:             "verdict",
	Array:               "array",
	Bitstring:           "bitstring",
	Charstring:          "charstring",
	Hexstring:           "hexstring",
	Map:                 "map",
	Octetstring:         "octetstring",
	RecordOf:            "record of",
	SetOf:               "set of",
	UniversalCharstring: "universal charstring",
	Anytype:             "anytype",
	Component:           "component",
	Object:              "object",
	Port:                "port",
	Record:              "record",
	Set:                 "set",
	Timer:               "timer",
	Trait:               "trait",
	Union:               "union",
	Altstep:             "altstep",
	Configuration:       "configuration",
	Default:             "default",
	Function:            "function",
	Testcase:            "testcase",
}

// String returns a string describing the kind, such as "error or integer or float".
func (k Kind) String() string {
	if k == 0 {
		return "any"
	}
	var ret []string
	for i := uint(0); i < 64; i++ {
		if k&(1<<i) != 0 {
			if name, ok := kindNames[1<<i]; ok {
				ret = append(ret, name)
			}
		}
	}
	if len(ret) == 0 {
		return "unknown"
	}
	return strings.Join(ret, " or ")
}

// A Type represents a TTCN-3 type.
type Type interface {
	String() string
}

// A PrimitiveType represents a non-composite type, such as integer, boolean,
// verdicttype, ...
type PrimitiveType struct {
	Kind
	Name             string
	ValueConstraints []Value
	Methods          map[string]Type
}

func (t *PrimitiveType) String() string {
	if t.ValueConstraints == nil {
		return t.Kind.String()
	}
	var constraints []string
	for _, v := range t.ValueConstraints {
		constraints = append(constraints, v.String())
	}
	return fmt.Sprintf("%s (%s)", t.Kind.String(), strings.Join(constraints, ", "))

}

// A ListType represents a collection of values of the same type, such as a
// record, set, map, hexstring, ...
//
// The behaviour of a list type is determined by its kind:
//   - a RecordOf is a ordered collection, while Map is an unordered
//     collection or pairs.
//   - an Array has a different string representation than a RecordOf:
//     `integer[5]` vs. `record length(5) of integer`
//
// The behaviour for other kinds than Array, RecordOf, Map, Hexstring,
// Bitstring, Octetstring, Charstring, UniversalCharstring is undefined and may
// or may not cause a runtime error.
type ListType struct {
	Kind
	Name             string
	ElementType      Type
	ValueConstraints []Value
	LengthConstraint Value
	Methods          map[string]Type
}

func (t *ListType) String() string {
	elem := "any"
	if t.ElementType != nil && !isString(t.Kind) {
		elem = t.ElementType.String()
	}
	switch t.Kind {
	case RecordOf, Any:
		var lengthConstraint string = " "
		if t.LengthConstraint.Expr != nil {
			lengthConstraint = " length(" + t.LengthConstraint.String() + ") "
		}
		return "record" + lengthConstraint + "of " + elem
	case SetOf:
		var lengthConstraint string = " "
		if t.LengthConstraint.Expr != nil {
			lengthConstraint = " length(" + t.LengthConstraint.String() + ") "
		}
		return "set" + lengthConstraint + "of " + elem
	case Map:
		if _, ok := t.ElementType.(*PairType); !ok {
			return "map from " + elem + " to any"
		}
		return "map from " + elem
	case Charstring, Octetstring, Hexstring, Bitstring, UniversalCharstring:
		var lengthConstraint string = ""
		if t.LengthConstraint.Expr != nil {
			lengthConstraint = " length(" + t.LengthConstraint.String() + ")"
		}
		return t.Kind.String() + lengthConstraint
	case Array:
		if eleType, ok := t.ElementType.(*ListType); ok && (eleType.Kind == RecordOf || eleType.Kind == SetOf || eleType.Kind == Map) {
			elem = "(" + elem + ")"
		}
		if t.LengthConstraint.Expr == nil {
			return elem + "[]"
		}
		return elem + "[" + t.LengthConstraint.String() + "]"
	}
	return ""
}

func isString(t Kind) bool {
	if t == Bitstring || t == Charstring || t == Octetstring || t == Hexstring || t == UniversalCharstring {
		return true
	}
	return false
}

// A StructuredType represents structured types, such as record, set, union,
// class, ...
type StructuredType struct {
	Kind
	Name             string
	External         bool
	Fields           []Field
	Extends          []Type
	ValueConstraints []Value
	Methods          map[string]Type
}

func (t *StructuredType) String() string {
	return ""
}

// A Field represents a fields in structures types.
type Field struct {
	Type
	Name     string
	Optional bool
	Default  bool
	Abstract bool
}

// A BehaviourType represents behaviour (testcases, functions, behaviour types, ...)
type BehaviourType struct {
	Kind
	Name       string
	External   bool
	Parameters []Type
	Receiver   Type
	RunsOn     Type
	System     Type
	MTC        Type
	Port       Type
	ExecuteOn  Type
	Return     Type
}

func (t *BehaviourType) String() string {
	return ""
}

// A PairType represents a pair. Pairs are not specified by TTCN-3 standard explicitly. It is for modeling map types as
// a set of key-value-pairs.
type PairType struct {
	First, Second Type
}

func (t *PairType) String() string {
	res := []string{"any", "any"}
	if t.First != nil {
		res[0] = t.First.String()
	}
	if t.Second != nil {
		res[1] = t.Second.String()
	}

	return strings.Join(res, " to ")
}

// A Value represents a single value constraint, such as '1' or '10..20'.
type Value struct {
	syntax.Expr
}

func (v Value) String() string {
	return printExpr(v.Expr)
}

func printExpr(e syntax.Expr) string {
	switch n := e.(type) {
	case *syntax.ValueLiteral:
		return n.Tok.String()
	case *syntax.Ident:
		return n.String()
	case *syntax.BinaryExpr:
		switch n.Op.Kind() {
		case syntax.RANGE:
			return fmt.Sprintf("%s..%s", printExpr(n.X), printExpr(n.Y))
		case syntax.ASSIGN:
			return fmt.Sprintf("%s := %s", printExpr(n.X), printExpr(n.Y))
		}
	}
	panic(fmt.Sprintf("not implemented: %T", e))
}
