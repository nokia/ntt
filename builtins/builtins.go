package builtins

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/nokia/ntt/runtime"
)

func Lengthof(args ...runtime.Object) runtime.Object {
	type lengther interface {
		Len() int
	}
	if l, ok := args[0].(lengther); ok {
		return runtime.NewInt(l.Len())
	}
	return runtime.Errorf("%s types have no length", args[0].Type())
}

func Rnd(...runtime.Object) runtime.Object {
	return runtime.Float(rand.Float64())
}

func Int2Str(args ...runtime.Object) runtime.Object {
	return runtime.NewCharstring(args[0].(runtime.Int).String())
}

func Int2Char(args ...runtime.Object) runtime.Object {
	n := args[0].(runtime.Int)
	if i := n.Uint64(); n.IsUint64() && i <= 127 {
		return runtime.NewCharstring(string(rune(i)))
	}
	return runtime.Errorf("Argument is out of range. Range is from 0 to 127. Int = %s", n.String())
}

func Str2Int(args ...runtime.Object) runtime.Object {
	s := args[0].(*runtime.String)
	i, err := strconv.Atoi(s.String())
	if err != nil {
		return runtime.Errorf("invalid syntax: %s", s.String())
	}
	return runtime.NewInt(i)
}

func Str2Float(args ...runtime.Object) runtime.Object {
	s := args[0].(*runtime.String)

	_, err := strconv.ParseFloat(s.String(), 64)
	if err != nil {
		return runtime.Errorf("invalid syntax: %s", s.String())
	}
	return runtime.NewFloat(s.String())
}

func Int2Unichar(args ...runtime.Object) runtime.Object {
	n := args[0].(runtime.Int)
	if i := n.Uint64(); n.IsUint64() && i <= 2147483647 {
		return runtime.NewUniversalString(string(rune(i)))
	}
	return runtime.Errorf("Argument is out of range. Range is from 0 to 2147483647. Int = %s", n.String())
}

func Unichar2Int(args ...runtime.Object) runtime.Object {
	s := args[0].(*runtime.String)
	if s.Len() != 1 {
		return runtime.Errorf("argument must be of length=1")
	}
	return runtime.NewInt(int(s.Value[0]))
}

func Int2Bit(args ...runtime.Object) runtime.Object {
	i := args[0].(runtime.Int)
	if i.Sign() < 0 {
		return runtime.Errorf("%s invalue is less than zero", args[0].Type())
	}

	l := args[1].(runtime.Int)
	if l.Sign() < 0 {
		return runtime.Errorf("length must be greater or equal than zero")
	}

	if i.BitLen() > int(l.Int64()) {
		return runtime.Errorf("%s value requires more than %d bits", i.String(), l.BitLen())
	}

	if !l.IsInt64() {
		return runtime.Errorf("length argument out of range (int64)")
	}

	return &runtime.Binarystring{
		String: fmt.Sprintf("'%0*s'B", l.Int64(), i.Text(2)),
		Value:  i.Value(),
		Unit:   runtime.Bit,
		Length: int(l.Int64()),
	}
}

func Int2Float(args ...runtime.Object) runtime.Object {
	return runtime.NewFloat(args[0].(runtime.Int).String())
}

func Float2Int(args ...runtime.Object) runtime.Object {
	return runtime.NewInt(int(args[0].(runtime.Float)))
}

func Int2Enum(args ...runtime.Object) runtime.Object {
	i := args[0].(runtime.Int)
	if !i.IsInt64() {
		return runtime.Errorf("integer value out of range (int64)")
	}
	e, ok := args[1].(*runtime.EnumValue)
	if !ok {
		return runtime.Errorf("second argument must be an enum value")
	}
	if err := e.SetValueById(int(i.Int64())); err != nil {
		return err
	}
	return nil
}

func Enum2Int(args ...runtime.Object) runtime.Object {
	e, ok := args[0].(*runtime.EnumValue)
	if !ok {
		return runtime.Errorf("Argument must be an enum value")
	}

	ranges, err := e.ReturnIdByValue()
	if err != nil {
		return runtime.Errorf("Can't find ID of provided value")
	}
	smallestId := 0
	for i, enumRange := range ranges {
		if enumRange.First < smallestId || i == 0 {
			smallestId = enumRange.First
		}
	}
	return runtime.NewInt(smallestId)
}

func Log(args ...runtime.Object) runtime.Object {
	var ss []string
	for _, arg := range args {
		ss = append(ss, arg.Inspect())
	}
	fmt.Println(strings.Join(ss, " "))
	return nil
}

func Match(args ...runtime.Object) runtime.Object {
	if len(args) != 2 {
		return runtime.Errorf("wrong number of arguments. got=%d, want=2", len(args))
	}

	b, err := match(args[0], args[1])
	if err != nil {
		return runtime.Errorf("match: %w", err)
	}

	return runtime.Bool(b)
}

func makeSet(lt runtime.ListType, args ...runtime.Object) func(...runtime.Object) runtime.Object {
	return func(args ...runtime.Object) runtime.Object {
		return &runtime.List{ListType: lt, Elements: args}
	}
}

func init() {
	builtins := map[string]func(...runtime.Object) runtime.Object{
		"float2int(in float f) return integer":                  Float2Int,
		"int2bit(in integer i, in integer l) return bitstring":  Int2Bit,
		"int2char(in integer i) return charstring":              Int2Char,
		"int2enum(in integer i, out any e)":                     Int2Enum,
		"enum2int(in any e) return integer":                     Enum2Int,
		"int2float(in integer i) return float":                  Int2Float,
		"int2str(in integer i) return charstring":               Int2Str,
		"str2int(in charstring s) return integer":               Str2Int,
		"str2float(in charstring s) return float":               Str2Float,
		"int2unichar(in integer i) return universal charstring": Int2Unichar,
		"lengthof(in any a) return integer":                     Lengthof,
		"rnd() return float":                                    Rnd,
		"unichar2int(in universal charstring s) return integer": Unichar2Int,

		"log":         Log,
		"match":       Match,
		"superset":    makeSet(runtime.SUPERSET),
		"subset":      makeSet(runtime.SUBSET),
		"permutation": makeSet(runtime.PERMUTATION),
		"complement":  makeSet(runtime.COMPLEMENT),
	}

	for name, builtin := range builtins {
		if err := runtime.AddBuiltin(name, builtin); err != nil {
			panic(err)
		}
	}
}
