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
	descriptionString() string
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
	if _, ok := Predefined[t.Name]; ok {
		return t.Name
	}
	if t.Name == "" {
		return t.descriptionString()
	}
	return t.Name + " [" + t.descriptionString() + "]"
}

func (t *PrimitiveType) descriptionString() string {
	if _, ok := Predefined[t.Name]; ok {
		return t.Name
	}
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
//   - a RecordOf is a ordered collection, while SetOf is an unordered
//     collection or pairs.
//   - an Array has a different string representation than a RecordOf:
//     `integer[5]` vs. `record length(5) of integer`
//
// The behaviour for other kinds than Array, RecordOf, Hexstring,
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
	if _, ok := Predefined[t.Name]; ok {
		return t.Name
	}
	if t.Name == "" {
		return t.descriptionString()
	}
	return t.Name + " [" + t.descriptionString() + "]"
}

func (t *ListType) descriptionString() string {
	if _, ok := Predefined[t.Name]; ok {
		return t.Name
	}
	elem := "any"
	if t.ElementType != nil && !isString(t.Kind) {
		elem = t.ElementType.descriptionString()
	}
	switch t.Kind {
	case RecordOf, Any:
		var lengthConstraint string = ""
		if t.LengthConstraint.Expr != nil {
			lengthConstraint = "length(" + t.LengthConstraint.String() + ") "
		}
		return "record " + lengthConstraint + "of " + elem
	case SetOf:
		var lengthConstraint string = ""
		if t.LengthConstraint.Expr != nil {
			lengthConstraint = "length(" + t.LengthConstraint.String() + ") "
		}
		return "set " + lengthConstraint + "of " + elem
	case Charstring, Octetstring, Hexstring, Bitstring, UniversalCharstring:
		var lengthConstraint string = ""
		if t.LengthConstraint.Expr != nil {
			lengthConstraint = " length(" + t.LengthConstraint.String() + ")"
		}
		return t.Kind.String() + lengthConstraint
	case Array:
		if eleType, ok := t.ElementType.(*ListType); ok && (eleType.Kind == RecordOf || eleType.Kind == SetOf) {
			elem = "(" + elem + ")"
		} else if _, ok := t.ElementType.(*MapType); ok {
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
	if _, ok := Predefined[t.Name]; ok {
		return t.Name
	}
	if t.Name == "" {
		return t.descriptionString()
	}
	if t.Kind == Component || t.Kind == Port || t.Kind == Trait || t.Kind == Object {
		return t.Name
	}
	return t.Name + " [" + t.descriptionString() + "]"
}

func (t *StructuredType) descriptionString() string {
	if _, ok := Predefined[t.Name]; ok {
		return t.Name
	}
	var typeKind string
	switch t.Kind {
	case Record, Set, Any:
		if t.Kind == Set {
			typeKind = "set"
		} else {
			typeKind = "record"
		}
		var fields []string
		for _, f := range t.Fields {
			fields = append(fields, f.String())
		}
		if t.ValueConstraints == nil {
			return typeKind + " {" + strings.Join(fields, ", ") + "}"
		}
		var constraints []string
		for _, v := range t.ValueConstraints {
			constraints = append(constraints, v.String())
		}
		return typeKind + " {" + strings.Join(fields, ", ") + "} (" + strings.Join(constraints, ", ") + ")"
	case Component, Trait, Port:
		typeKind = t.Kind.String()
	case Object:
		typeKind = "class"
	}
	if t.Name != "" {
		return t.Name
	}
	return typeKind
}

// A Field represents a fields in structures types.
type Field struct {
	Type
	Name     string
	Optional bool
	Default  bool
	Abstract bool
}

func (f *Field) String() string {
	if f.Optional {
		return f.Type.descriptionString() + " optional"
	}
	return f.Type.descriptionString()
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

func (t *BehaviourType) descriptionString() string {
	return t.String()
}

// A MapType represents a map type.
type MapType struct {
	Name     string
	From, To Type
}

func (t *MapType) String() string {
	if t.Name == "" {
		return t.descriptionString()
	}
	return t.Name + " [" + t.descriptionString() + "]"
}

func (t *MapType) descriptionString() string {
	res := []string{"any", "any"}
	if t.From != nil {
		res[0] = t.From.descriptionString()
	}
	if t.To != nil {
		res[1] = t.To.descriptionString()
	}

	return fmt.Sprintf("map from %s to %s", res[0], res[1])
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

func TypeOf(n syntax.Expr) Type {
	//types to check:
	//ValueLiteral, CompositeLiteral, BinaryExpr, UnaryExpr, Ident
	switch n := n.(type) {
	case *syntax.ValueLiteral:
		switch n.Tok.Kind() {
		case syntax.INT:
			return Predefined["integer"]
		case syntax.FLOAT, syntax.NAN:
			return Predefined["float"]
		case syntax.TRUE, syntax.FALSE:
			return Predefined["boolean"]
		case syntax.PASS, syntax.FAIL, syntax.INCONC, syntax.NONE, syntax.ERROR:
			return Predefined["verdicttype"]
		case syntax.STRING:
			for _, r := range n.Tok.String() {
				if r < 32 || 126 < r {
					return Predefined["universal charstring"]
				}
			}
			return Predefined["charstring"]

		case syntax.BSTRING:
			s := n.Tok.String()
			if len(s) == 0 {
				return nil
			}
			switch s[len(s)-1] {
			case 'H', 'h':
				return Predefined["hexstring"]
			case 'O', 'o':
				return Predefined["octetstring"]
			case 'B', 'b':
				return Predefined["bitstring"]
			}
			return nil
		}
	case *syntax.Ident:
		// names?
		if n, ok := Predefined[n.String()]; ok {
			return n
		}
		if n.String() == "infinity" {
			return Predefined["float"]
		}
	case *syntax.BinaryExpr:
		switch n.Op.Kind() {
		case syntax.LT, syntax.GT, syntax.LE, syntax.GE, syntax.EQ, syntax.NE, syntax.AND, syntax.OR, syntax.XOR:
			return Predefined["boolean"]
		case syntax.ADD, syntax.SUB, syntax.MUL, syntax.DIV, syntax.AND4B, syntax.OR4B, syntax.XOR4B, syntax.CONCAT:
			if operandType := TypeOf(n.X); operandType == TypeOf(n.Y) {
				return operandType
			}
			return nil
		case syntax.MOD, syntax.REM:
			return Predefined["integer"]
		case syntax.SHL, syntax.SHR, syntax.ROL, syntax.ROR:
			return TypeOf(n.X)
		case syntax.ASSIGN:
			// Investigate: Label on the left, omit on the right, Maps
			if n.Y.FirstTok().Kind() == syntax.OMIT {
				if t := TypeOf(n.X); t != nil {
					return t
				} else {
					return Predefined["boolean"]
				}
			}
			if X, ok := n.X.(*syntax.IndexExpr); ok && X.X == nil && TypeOf(X.Index) != Predefined["integer"] {
				return &MapType{From: TypeOf(X.Index), To: TypeOf(n.Y)}
			}
			return TypeOf(n.Y)
		}
	case *syntax.UnaryExpr:
		switch n.Op.Kind() {
		case syntax.NOT:
			return Predefined["boolean"]
		case syntax.NOT4B, syntax.ADD, syntax.SUB:
			return TypeOf(n.X)
		case syntax.INC, syntax.DEC:
			return Predefined["integer"]
		}
	}
	return nil
}

func isSuper(type1 Type, type2 Type) Type {
	if type1 == Predefined["any"] || type2 == Predefined["any"] {
		return Predefined["any"]
	}
	switch type1 := type1.(type) {
	case *PrimitiveType:
		if type2, ok := type2.(*PrimitiveType); !ok {
			return nil
		} else {
			if type1.Kind == type2.Kind {
				if type1.ValueConstraints == nil {
					return type1
				}
				if type2.ValueConstraints == nil {
					return type2
				}
				// ToDo: Add more ConstraintChecks
			}
			return nil
		}
	case *ListType:
		if type2, ok := type2.(*ListType); !ok {
			return nil
		} else {
			if type1.Kind == type2.Kind {
				if type1.ElementType == type2.ElementType {
					return type1
				}
				if isString(type1.Kind) {
					// Check Constraints
					return nil
				}
				// check ElementType (Array RecordOf SetOf)
				superElement := isSuper(type1.ElementType, type2.ElementType)
				if superElement == type1.ElementType {
					return type1
				}
				if superElement == type2.ElementType {
					return type2
				}
				return nil
			}
			switch type1.Kind {
			case Hexstring, Octetstring, Bitstring, SetOf:
				return nil
			case Charstring:
				if type2.Kind == UniversalCharstring {
					// check Constraints
					return type2
				}
				return nil
			case UniversalCharstring:
				if type2.Kind == Charstring {
					// check Constraints
					return type1
				}
				return nil
			case RecordOf:
				if type2.Kind == Array {
					// check ElementType and Constraints
					return type1
				}
				return nil
			case Array:
				if type2.Kind == RecordOf {
					// check ElementType and Constraints
					return type2
				}
				return nil
			}
		}
	case *StructuredType:
		if type2, ok := type2.(*StructuredType); !ok {
			return nil
		} else {
			if type1.Kind == type2.Kind && len(type1.Fields) == len(type2.Fields) {
				for i, f := range type1.Fields {
					if f.Optional == type2.Fields[i].Optional && isSuper(f.Type, type2.Fields[i].Type) == f.Type {
						continue
					} else {
						return nil
					}
				}
				return type1
			}
		}
	case *MapType:
		if type2, ok := type2.(*MapType); !ok || type1.From != type2.From {
			return nil
		} else {
			if superTo := isSuper(type1.To, type2.To); superTo == type1.To {
				return type1
			} else if superTo == type2.To {
				return type2
			}
		}
	}
	return nil
}
