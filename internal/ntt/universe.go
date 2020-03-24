package ntt

import "github.com/nokia/ntt/internal/loc"

var Universe scope

var Typ = []*BasicType{
	Invalid:             {Invalid, "invalid"},
	Bitstring:           {Bitstring, "bitstring"},
	Boolean:             {Boolean, "boolean"},
	Charstring:          {Charstring, "charstring"},
	Component:           {Component, "component"},
	Float:               {Float, "float"},
	Hexstring:           {Hexstring, "hexstring"},
	Integer:             {Integer, "integer"},
	Octetstring:         {Octetstring, "octetstring"},
	Omit:                {Omit, "omit"},
	Template:            {Template, "template"},
	Timer:               {Timer, "timer"},
	UniversalCharstring: {UniversalCharstring, "universal charstring"},
	Verdict:             {Verdict, "verdict"},
	String:              {String, "string"},
	Numerical:           {Numerical, "numerical"},
}

type noPos struct{}

func (n *noPos) Pos() loc.Pos { return loc.NoPos }
func (n *noPos) End() loc.Pos { return loc.NoPos }

func init() {
	for _, t := range Typ {
		if Universe.Insert(NewTypeName(&noPos{}, t.String(), t)) != nil {
			panic("internal error: double declaration")
		}
	}
}
