package runtime

import (
	"fmt"
	"math/big"
	"math/rand"
	"strings"
)

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

	"rnd": {Fn: func(args ...Object) Object {
		if len(args) != 0 {
			return Errorf("wrong number of arguments. got=%d, want=0", len(args))
		}

		return Float{Float: big.NewFloat(rand.Float64())}
	}},

	"int2float": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return Errorf("wrong number of arguments. got=%d, want=1", len(args))
		}

		i, ok := args[0].(Int)
		if !ok {
			return Errorf("%s arguments not supported", args[0].Type())
		}
		return Float{Float: new(big.Float).SetInt(i.Int)}
	}},

	"float2int": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return Errorf("wrong number of arguments. got=%d, want=1", len(args))
		}

		f, ok := args[0].(Float)
		if !ok {
			return Errorf("%s arguments not supported", args[0].Type())
		}
		i, _ := f.Int(nil)
		return Int{Int: i}
	}},

	"log": {Fn: func(args ...Object) Object {
		var ss []string
		for _, arg := range args {
			ss = append(ss, arg.Inspect())
		}
		fmt.Println(strings.Join(ss, " "))
		return nil
	}},
}

type Builtin struct {
	Fn func(args ...Object) Object
}

func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }
func (b *Builtin) Inspect() string  { return "builtin function" }
