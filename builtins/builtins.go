package builtins

import (
	"fmt"
	"math/big"
	"math/rand"
	"strings"

	"github.com/nokia/ntt/runtime"
)

func Lengthof(args ...runtime.Object) runtime.Object {
	if len(args) != 1 {
		return runtime.Errorf("wrong number of arguments. got=%d, want=1", len(args))
	}

	switch arg := args[0].(type) {
	case *runtime.String:
		return runtime.Int{Int: big.NewInt(int64(len(arg.Value)))}
	case *runtime.Bitstring:
		return runtime.Int{Int: big.NewInt(int64(arg.Value.BitLen() / int(arg.Unit)))}
	}
	return runtime.Errorf("%s arguments not supported", args[0].Type())
}

func Rnd(args ...runtime.Object) runtime.Object {
	if len(args) != 0 {
		return runtime.Errorf("wrong number of arguments. got=%d, want=0", len(args))
	}

	return runtime.Float(rand.Float64())
}

func Int2Bit(args ...runtime.Object) runtime.Object {

	if len(args) != 2 {
		return runtime.Errorf("wrong number of arguments. got=%d, want=2", len(args))
	}

	number, ok := args[0].(runtime.Int)
	if !ok {
		return runtime.Errorf("%s arguments not supported", args[0].Type())
	}
	if number.Sign() == -1 {
		return runtime.Errorf("%s invalue is less than zero", args[0].Type())
	}

	lengthArg, lenOk := args[1].(runtime.Int)
	if !lenOk {
		return runtime.Errorf("%s arguments not supported", args[1].Type())
	}
	length := int(lengthArg.Uint64());
	if number.BitLen() > length {
		return runtime.Errorf("%v value requires more than %d bits", number, length)
	}

	return &runtime.Bitstring{Value: number.Value(), Unit: runtime.Bit, Padding: 4}
}

func Int2Float(args ...runtime.Object) runtime.Object {
	if len(args) != 1 {
		return runtime.Errorf("wrong number of arguments. got=%d, want=1", len(args))
	}

	i, ok := args[0].(runtime.Int)
	if !ok {
		return runtime.Errorf("%s arguments not supported", args[0].Type())
	}

	f, _ := new(big.Float).SetInt(i.Int).Float64()
	return runtime.Float(f)
}

func Float2Int(args ...runtime.Object) runtime.Object {
	if len(args) != 1 {
		return runtime.Errorf("wrong number of arguments. got=%d, want=1", len(args))
	}

	f, ok := args[0].(runtime.Float)
	if !ok {
		return runtime.Errorf("%s arguments not supported", args[0].Type())
	}

	i, _ := new(big.Float).SetFloat64(float64(f)).Int(nil)
	return runtime.Int{Int: i}
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
	runtime.AddBuiltin("lengthof", Lengthof)
	runtime.AddBuiltin("rnd", Rnd)
	runtime.AddBuiltin("int2bit", Int2Bit)
	runtime.AddBuiltin("int2float", Int2Float)
	runtime.AddBuiltin("float2int", Float2Int)
	runtime.AddBuiltin("log", Log)
	runtime.AddBuiltin("match", Match)
	runtime.AddBuiltin("superset", makeSet(runtime.SUPERSET))
	runtime.AddBuiltin("subset", makeSet(runtime.SUBSET))
	runtime.AddBuiltin("permutation", makeSet(runtime.PERMUTATION))
	runtime.AddBuiltin("complement", makeSet(runtime.COMPLEMENT))
}
