package runtime

import "math/big"

var Builtins = map[string]*Builtin{
	"lengthof": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return Errorf("wrong number of arguments. got=%d, want=1", len(args))
		}

		switch arg := args[0].(type) {
		case *String:
			return Int{Int: big.NewInt(int64(len(arg.Value)))}
		case *Bitstring:
			return Int{Int: big.NewInt(int64(arg.Value.BitLen() / int(arg.Unit)))}

		}
		return Errorf("%s arguments not supported", args[0].Type())
	}},
}

type Builtin struct {
	Fn func(args ...Object) Object
}

func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }
func (b *Builtin) Inspect() string  { return "builtin function" }
