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
		return runtime.Int{Int: big.NewInt(int64(arg.Length))}
	case *runtime.UniversalString:
		return runtime.Int{Int: big.NewInt(int64(len(arg.String)))}
	}
	return runtime.Errorf("%s arguments not supported", args[0].Type())
}

func Rnd(args ...runtime.Object) runtime.Object {
	if len(args) != 0 {
		return runtime.Errorf("wrong number of arguments. got=%d, want=0", len(args))
	}

	return runtime.Float(rand.Float64())
}

func Int2str(args ...runtime.Object) runtime.Object {
	if len(args) != 1 {
		return runtime.Errorf("wrong number of arguments. got=%d, want=1", len(args))
	}
	number, ok := args[0].(runtime.Int)
	if !ok {
		return runtime.Errorf("%s arguments not supported", args[0].Type())
	}
	if !number.IsInt64() {
		return runtime.Errorf("Provided argument is not 64bit-integer")
	}
	return &runtime.String{Value: []rune(fmt.Sprintf("%d", number.Int64()))}
}

func Int2char(args ...runtime.Object) runtime.Object {
	if len(args) != 1 {
		return runtime.Errorf("wrong number of arguments. got=%d, want=1", len(args))
	}
	number, ok := args[0].(runtime.Int)
	if !ok {
		return runtime.Errorf("%s arguments not supported", args[0].Type())
	}
	if !number.IsInt64() {
		return runtime.Errorf("Provided argument is not integer.")
	}
	i := number.Int64()
	if i < 0 || i > 127 {
		return runtime.Errorf("Argument is out of range. Range is from 0 to 127. Int = %d", i)
	}

	return &runtime.String{Value: []rune(fmt.Sprintf("%c", i))}
}

func Int2unichar(args ...runtime.Object) runtime.Object {

	if len(args) != 1 {
		return runtime.Errorf("wrong number of arguments. got=%d, want=1", len(args))
	}
	number, ok := args[0].(runtime.Int)
	if !ok {
		return runtime.Errorf("%s arguments not supported", args[0].Type())
	}
	if !number.IsInt64() {
		return runtime.Errorf("argument is not int64")
	}
	numberI64 := number.Int64()
	if numberI64 < 0 {
		return runtime.Errorf("value must be grater or equal to 0")
	}
	if numberI64 > 2147483647 {
		return runtime.Errorf("value must be less than 2147483647")
	}
	return runtime.NewUniversalString(fmt.Sprintf("%c", numberI64))
}

func Unichar2int(args ...runtime.Object) runtime.Object {
	if len(args) != 1 {
		return runtime.Errorf("wrong number of arguments. got=%d, want=1", len(args))
	}

	switch args[0].Type() {
	case runtime.UNISTRING:
		ustring, _ := args[0].(*runtime.UniversalString)
		if ustring.Len() != 1 {
			return runtime.Errorf("argument must be of lenght=1")
		}
		return runtime.NewInt(fmt.Sprintf("%d", ustring.String[0]))
	case runtime.STRING:
		pstring, _ := args[0].(*runtime.String)
		if pstring.Len() != 1 {
			return runtime.Errorf("argument must be of lenght=1")
		}
		return runtime.NewInt(fmt.Sprintf("%d", pstring.Value[0]))
	default:
		return runtime.Errorf("%s arguments not supported", args[0].Type())
	}
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
	if !lengthArg.IsInt64() {
		return runtime.Errorf("length argument out of range (int64)")
	}
	length := int(lengthArg.Int64())
	if number.BitLen() > length {
		return runtime.Errorf("%v value requires more than %d bits", number, length)
	}
	if length < 0 {
		return runtime.Errorf("length must be greater or equal than zero")
	}
	return &runtime.Bitstring{String: fmt.Sprintf("'%0*s'B", length, number.Value().Text(2)), Value: number.Value(), Unit: runtime.Bit, Length: length}
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

func Int2Enum(args ...runtime.Object) runtime.Object {

	if len(args) != 2 {
		return runtime.Errorf("wrong number of arguments. got=%d, want=2", len(args))
	}

	number, ok := args[0].(runtime.Int)
	if !ok {
		return runtime.Errorf("%s arguments not supported", args[0].Type())
	}
	if !number.IsInt64() {
		return runtime.Errorf("Provided argument is not integer of 64 bits.")
	}
	object, ok := args[1].(*runtime.EnumValue)
	if !ok {
		return runtime.Errorf("%s arguments not supported", args[1].Type())
	}

	n := int(number.Int64())

	err := object.SetValueById(n)
	if err != nil {
		return runtime.Errorf("Can't find value with provided int = %d", n)
	}

	return nil
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
	runtime.AddBuiltin("int2str", Int2str)
	runtime.AddBuiltin("int2bit", Int2Bit)
	runtime.AddBuiltin("int2float", Int2Float)
	runtime.AddBuiltin("float2int", Float2Int)
	runtime.AddBuiltin("int2enum", Int2Enum)
	runtime.AddBuiltin("int2char", Int2char)
	runtime.AddBuiltin("int2unichar", Int2unichar)
	runtime.AddBuiltin("unichar2int", Unichar2int)
	runtime.AddBuiltin("log", Log)
	runtime.AddBuiltin("match", Match)
	runtime.AddBuiltin("superset", makeSet(runtime.SUPERSET))
	runtime.AddBuiltin("subset", makeSet(runtime.SUBSET))
	runtime.AddBuiltin("permutation", makeSet(runtime.PERMUTATION))
	runtime.AddBuiltin("complement", makeSet(runtime.COMPLEMENT))
}
