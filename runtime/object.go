package runtime

import (
	"bytes"
	"errors"
	"fmt"
	"hash/fnv"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/nokia/ntt/ttcn3/ast"
)

type Object interface {
	Inspect() string
	Type() ObjectType
	Equal(Object) bool
}

type ObjectType string

const (
	UNKNOWN      ObjectType = "unknown object"
	UNDEFINED    ObjectType = "undefined value"
	ERROR        ObjectType = "runtime error"
	BREAK        ObjectType = "break event"
	CONTINUE     ObjectType = "continue event"
	RETURN_VALUE ObjectType = "return value"
	INTEGER      ObjectType = "integer"
	FLOAT        ObjectType = "float"
	BOOL         ObjectType = "boolean"
	STRING       ObjectType = "string"
	BITSTRING    ObjectType = "bitstring"
	UNISTRING    ObjectType = "universal string"
	FUNCTION     ObjectType = "function"
	LIST         ObjectType = "list"
	RECORD       ObjectType = "record"
	MAP          ObjectType = "map"
	BUILTIN_OBJ  ObjectType = "builtin function"
	VERDICT      ObjectType = "verdict"
	ENUMERATED   ObjectType = "enumerated"
	ANY          ObjectType = "?"
	ANY_OR_NONE  ObjectType = "*"
)

type Unit int

const (
	Bit    Unit = 1
	Hex    Unit = 4
	Octett Unit = 8
)

func (u Unit) Base() int {
	switch u {
	case Bit:
		return 2
	case Hex, Octett:
		return 16
	default:
		return -1
	}
}

var (
	ErrSyntax = errors.New("invalid syntax")
	Undefined = &singelton{typ: UNDEFINED}
	Break     = &singelton{typ: BREAK}
	Continue  = &singelton{typ: CONTINUE}
	Any       = &singelton{typ: ANY}
	AnyOrNone = &singelton{typ: ANY_OR_NONE}
)

type singelton struct {
	typ ObjectType
}

func (s *singelton) Inspect() string  { return string(s.typ) }
func (s *singelton) Type() ObjectType { return s.typ }

func (s *singelton) Equal(obj Object) bool {
	if other, ok := obj.(*singelton); ok {
		return s.typ == other.typ
	}
	return false
}

type Error struct {
	Err error
}

func (e *Error) Error() string    { return e.Err.Error() }
func (e *Error) Type() ObjectType { return ERROR }
func (e *Error) Inspect() string  { return fmt.Sprintf("Error: %s", e.Error()) }

func (e *Error) Equal(obj Object) bool {
	if other, ok := obj.(*Error); ok {
		return errors.Is(e, other)
	}
	return false
}

func Errorf(format string, a ...interface{}) *Error {
	return &Error{Err: fmt.Errorf(format, a...)}
}

func IsError(v interface{}) bool {
	_, ok := v.(*Error)
	return ok
}

type Bool bool

func (b Bool) Type() ObjectType { return BOOL }
func (b Bool) Inspect() string  { return fmt.Sprintf("%t", b) }
func (b Bool) Bool() bool       { return bool(b) }

func (b Bool) Equal(obj Object) bool {
	if other, ok := obj.(Bool); ok {
		return b == other
	}
	return false
}

func (b Bool) hashKey() hashKey {
	var value uint64
	if b {
		value = 1
	} else {
		value = 0
	}
	return hashKey{Type: b.Type(), Value: value}
}

func NewBool(b bool) Bool {
	return Bool(b)
}

type Float float64

func (f Float) Type() ObjectType { return FLOAT }
func (f Float) Inspect() string  { return fmt.Sprint(float64(f)) }

func (f Float) Equal(obj Object) bool {
	if other, ok := obj.(Float); ok {
		return f == other
	}
	return false
}

func NewFloat(s string) Float {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic(err.Error())
	}
	return Float(f)
}

type Int struct{ *big.Int }

func (i Int) Type() ObjectType { return INTEGER }
func (i Int) Inspect() string  { return i.String() }
func (i Int) Value() *big.Int  { return i.Int }

func (i Int) Equal(obj Object) bool {
	if other, ok := obj.(Int); ok {
		return i.Cmp(other.Int) == 0
	}
	return false
}

func (i Int) hashKey() hashKey {
	h := fnv.New64a()
	h.Write(i.Bytes())
	return hashKey{Type: i.Type(), Value: h.Sum64()}
}

func NewInt(s string) Int {
	i := &big.Int{}
	i.SetString(s, 10)
	return Int{i}
}

type Enum struct {
	Elements map[string]int
}

func (e *Enum) Type() ObjectType { return ENUMERATED }
func (e *Enum) Inspect() string {
	var ret []string
	for name, val := range e.Elements {
		ret = append(ret, fmt.Sprintf("%s(%d)", name, val))
	}
	return "{" + strings.Join(ret, ", ") + "}"
}

func (e *Enum) Equal(obj Object) bool {
	other, ok := obj.(*Enum)
	if !ok {
		return false
	}
	return reflect.DeepEqual(e, other)
}

func NewEnum(nums ...string) *Enum {
	ret := Enum{}
	ret.Elements = make(map[string]int)
	for id, name := range nums {
		ret.Elements[name] = id
	}
	return &ret
}

type String struct {
	Value []rune
}

func (s *String) Type() ObjectType { return STRING }
func (s *String) Inspect() string  { return string(s.Value) }

func (s *String) Equal(obj Object) bool {
	b, ok := obj.(*String)
	if !ok || len(s.Value) != len(b.Value) {
		return false
	}

	for i, v := range s.Value {
		if v != b.Value[i] {
			return false
		}
	}

	return true
}

func (s *String) hashKey() hashKey {
	h := fnv.New64a()
	h.Write([]byte(string(s.Value)))
	return hashKey{Type: s.Type(), Value: h.Sum64()}
}

func (s *String) Len() int {
	return len(s.Value)
}

func (s *String) Get(index int) Object {
	return NewString(string(s.Value[index]))
}

func NewString(s string) *String {
	return &String{Value: []rune(s)}
}

type UniversalString struct {
	String string
}

func (us *UniversalString) Type() ObjectType { return UNISTRING }
func (us *UniversalString) Inspect() string  { return us.String }
func (us *UniversalString) Equal(obj Object) bool {
	if other, ok := obj.(*UniversalString); ok {
		return us.String == other.String
	}
	return false
}
func (us *UniversalString) hashKey() hashKey {
	h := fnv.New64a()
	h.Write([]byte(us.String))
	return hashKey{Type: us.Type(), Value: h.Sum64()}
}
func (us *UniversalString) Len() int {
	return len(us.String)
}

func (us *UniversalString) Get(index int) Object {
	if index >= len(us.String) {
		return &UniversalString{String: ""}
	}
	return &UniversalString{String: string(us.String[index])}
}
func NewUniversalString(s string) *UniversalString {
	return &UniversalString{String: s}
}

type Bitstring struct {
	String string
	Value  *big.Int
	Unit   Unit
	Length int
}

func (b *Bitstring) Type() ObjectType { return BITSTRING }
func (b *Bitstring) Inspect() string {
	switch b.Unit {
	case Bit:
		return fmt.Sprintf("'%0*b'B", b.Length, b.Value)
	case Octett:
		return fmt.Sprintf("'%0*h'O", b.Length, b.Value)
	default:
		return fmt.Sprintf("'%0*h'H", b.Length, b.Value)
	}
}

func (b *Bitstring) Equal(obj Object) bool {
	if other, ok := obj.(*Bitstring); ok {
		return b.Value.Cmp(other.Value) == 0
	}
	return false
}

func (b *Bitstring) hashKey() hashKey {
	h := fnv.New64a()
	h.Write(b.Value.Bytes())
	return hashKey{Type: b.Type(), Value: h.Sum64()}
}

func NewBitstring(s string) (*Bitstring, error) {

	if len(s) < 3 || s[0] != '\'' || s[len(s)-2] != '\'' {
		return nil, ErrSyntax
	}

	var unit Unit
	s = s[:len(s)-1] + strings.ToUpper(string(s[len(s)-1])) // Capitalize unit
	switch s[len(s)-1] {
	case 'B':
		unit = Bit
	case 'H':
		unit = Hex
	case 'O':
		unit = Octett
	default:
		return nil, ErrSyntax
	}
	n := removeWhitespaces(s[1 : len(s)-2])

	if i, ok := new(big.Int).SetString(n, unit.Base()); ok {
		return &Bitstring{String: s, Value: i, Unit: unit, Length: len(n)}, nil
	}

	return NewBitstringWithWildcards(s, unit)
}

func NewBitstringWithWildcards(s string, unit Unit) (*Bitstring, error) {
	n := removeWhitespaces(s[1 : len(s)-2])

	switch unit {
	case Bit, Hex:
	case Octett:
		if !strings.Contains(n, "*") && len(n)%2 != 0 { //number of runes between '' has to be even, unless it contains *
			return nil, ErrSyntax
		}
	default:
		return nil, ErrSyntax
	}

	for _, r := range n {
		switch r {
		case '*', '?', '0', '1':
		case 'A', 'B', 'C', 'D', 'E', 'F', 'a', 'b', 'c', 'd', 'e', 'f', '2', '3', '4', '5', '6', '7', '8', '9':
			if unit == Bit {
				return nil, ErrSyntax
			}
		default:
			return nil, ErrSyntax
		}
	}
	return &Bitstring{String: s, Value: new(big.Int).SetInt64(-1), Unit: unit, Length: len(n)}, nil
}

func (b *Bitstring) Len() int { return b.Length }

func (b *Bitstring) Get(index int) Object {
	width := int(b.Unit)/8 + 1
	// If b.Unit is Octett, each "digit" is two bytes wide
	s := removeWhitespaces(b.String)
	s = "'" + s[1+index*width:1+(index+1)*width] + s[len(s)-2:]
	n, _ := new(big.Int).SetString(s[1:len(s)-2], b.Unit.Base())
	return &Bitstring{String: s, Value: n, Unit: b.Unit, Length: 1} //Length one, even for Octett
}

func BigIntToBitstring(b *big.Int, unit Unit) string {
	return "'" + b.Text(unit.Base()) + "'" + unit.String()
}

func (u Unit) String() string {
	switch u {
	case 1:
		return "B"
	case 4:
		return "H"
	case 8:
		return "O"
	default: // Will never happen
		return "-"
	}
}

func removeWhitespaces(s string) string {
	removeWhitespaces := func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}
	return strings.Map(removeWhitespaces, s)
}

type ListType string

const (
	RECORD_OF   ListType = "" //default
	SET_OF      ListType = "set of"
	COMPLEMENT  ListType = "complement"
	SUBSET      ListType = "subset"
	SUPERSET    ListType = "superset"
	PERMUTATION ListType = "permutation"
)

type List struct {
	ListType
	Elements []Object
}

func (l *List) IsOrdered() bool {
	return l.ListType == ""
}

func (l *List) Type() ObjectType { return LIST }
func (l *List) Inspect() string {
	var ss []string
	for _, obj := range l.Elements {
		if obj != nil {
			ss = append(ss, obj.Inspect())
		} else {
			ss = append(ss, "null")
		}
	}
	return "{" + strings.Join(ss, ", ") + "}"
}

func (l *List) Equal(obj Object) bool {
	other, ok := obj.(*List)
	if !ok {
		return false
	}

	// The order of elements is ignored, when at least one list is
	// unordered.
	//
	// The standard explicitly forbids this. Relaxing this restriction
	// makes untyped assignment lists easier to handle.
	//
	// We assume the proper semeantic checks are done before the runtime.
	if l.IsOrdered() && other.IsOrdered() {
		return EqualObjects(l.Elements, other.Elements)
	}

	return l.ListType == other.ListType && EqualObjectSet(l.Elements, other.Elements)

}

func (l *List) Get(index int) Object {
	return l.Elements[index]
}

func (l *List) Len() int {
	return len(l.Elements)
}

// NewList creates a new ordered list.
func NewList(objs ...Object) *List        { return &List{Elements: objs} }
func NewRecordOf(objs ...Object) *List    { return &List{Elements: objs} }
func NewSetOf(objs ...Object) *List       { return &List{Elements: objs, ListType: SET_OF} }
func NewSuperset(objs ...Object) *List    { return &List{Elements: objs, ListType: SUPERSET} }
func NewSubset(objs ...Object) *List      { return &List{Elements: objs, ListType: SUBSET} }
func NewPermutation(objs ...Object) *List { return &List{Elements: objs, ListType: PERMUTATION} }
func NewComplement(objs ...Object) *List  { return &List{Elements: objs, ListType: COMPLEMENT} }

type Function struct {
	Params *ast.FormalPars
	Body   *ast.BlockStmt
	Env    Scope
}

func (f *Function) Type() ObjectType { return FUNCTION }
func (f *Function) Inspect() string {
	var buf bytes.Buffer
	buf.WriteString("function(\"")
	for i, p := range f.Params.List {
		if i != 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(p.Name.String())
	}
	buf.WriteString(")")
	return buf.String()
}

func (f *Function) Equal(obj Object) bool {
	if other, ok := obj.(*Function); ok {
		// TODO(5nord) When are to functions equal?
		return *f == *other
	}
	return false
}

type ReturnValue struct {
	Value Object
}

func (r *ReturnValue) Type() ObjectType { return RETURN_VALUE }
func (r *ReturnValue) Inspect() string  { return r.Value.Inspect() }

func (r *ReturnValue) Equal(obj Object) bool {
	if other, ok := obj.(*ReturnValue); ok {
		return r.Value.Equal(other.Value)
	}
	return false
}

type Verdict string

const (
	NoneVerdict   Verdict = "none"
	PassVerdict   Verdict = "pass"
	InconcVerdict Verdict = "inconc"
	FailVerdict   Verdict = "fail"
	ErrorVerdict  Verdict = "error"
)

func (v Verdict) Type() ObjectType { return VERDICT }
func (v Verdict) Inspect() string  { return string(v) }
func (v Verdict) Equal(obj Object) bool {
	if other, ok := obj.(Verdict); ok {
		return v == other
	}
	return false
}

func (v Verdict) hashKey() hashKey {
	var value uint64
	switch v {
	case NoneVerdict:
		value = 0
	case PassVerdict:
		value = 1
	case InconcVerdict:
		value = 2
	case FailVerdict:
		value = 3
	case ErrorVerdict:
		value = 4
	default:
		panic(Errorf("unknown verdict"))
	}
	return hashKey{Type: v.Type(), Value: value}

}

type Builtin struct {
	Fn func(args ...Object) Object
}

func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }
func (b *Builtin) Inspect() string  { return "builtin function" }
func (b *Builtin) Equal(obj Object) bool {
	if other, ok := obj.(*Builtin); ok {
		return b == other
	}
	return false
}

type hashable interface {
	hashKey() hashKey
}

type hashKey struct {
	Type  ObjectType
	Value uint64
}

type pair struct {
	Key   Object
	Value Object
}

// Map is a map of objects.
type Map struct {
	pairs map[hashKey][]pair
}

// Get returns the value for the given key.
func (m *Map) Get(key Object) (Object, bool) {
	k, ok := key.(hashable)
	if !ok {
		return Errorf("%s is not hashable", key.Type()), false
	}

	for _, p := range m.pairs[k.hashKey()] {
		if p.Key.Equal(key) {
			return p.Value, true
		}
	}
	return nil, false
}

func (m *Map) Set(key Object, val Object) Object {
	k, ok := key.(hashable)
	if !ok {
		return Errorf("%s is not hashable", key.Type())
	}
	h := k.hashKey()
	m.pairs[h] = append(m.pairs[h], pair{Key: key, Value: val})
	return val
}

func (m *Map) Type() ObjectType { return MAP }
func (m *Map) Inspect() string {
	var buf bytes.Buffer
	pairs := []string{}
	for _, bucket := range m.pairs {
		for _, pair := range bucket {
			pairs = append(pairs, fmt.Sprintf("[%s] := %s", pair.Key.Inspect(), pair.Value.Inspect()))
		}
	}
	buf.WriteString("{")
	buf.WriteString(strings.Join(pairs, ", "))
	buf.WriteString("}")

	return buf.String()
}

func (m *Map) Equal(obj Object) bool {
	other, ok := obj.(*Map)
	if !ok {
		return false
	}
	if len(m.pairs) != len(other.pairs) {
		return false
	}

	for k, a := range m.pairs {
		b, ok := other.pairs[k]
		if !ok {
			return false
		}
		if len(a) != len(b) {
			return false
		}

		for i, v := range a {
			if !v.Value.Equal(b[i].Value) {
				return false
			}
		}
	}

	return true
}

func NewMap() *Map {
	return &Map{pairs: make(map[hashKey][]pair)}
}

// TODO(5nord) For simplicity we reuse the Map implementation. We should implement proper record semantics later.
type Record struct {
	Fields map[string]Object
}

func (r *Record) Get(name string) (Object, bool) {
	val, ok := r.Fields[name]
	return val, ok
}

func (r *Record) Set(name string, val Object) Object {
	r.Fields[name] = val
	return nil
}

func (r *Record) Type() ObjectType { return RECORD }
func (r *Record) Inspect() string {
	var buf bytes.Buffer
	fields := []string{}
	for key, val := range r.Fields {
		fields = append(fields, fmt.Sprintf("%s := %s", key, val.Inspect()))
	}
	buf.WriteString("{")
	buf.WriteString(strings.Join(fields, ", "))
	buf.WriteString("}")

	return buf.String()
}

func (r *Record) Equal(obj Object) bool {
	other, ok := obj.(*Record)
	if !ok {
		return false
	}
	if len(r.Fields) != len(other.Fields) {
		return false
	}

	for k, a := range r.Fields {
		b, ok := other.Fields[k]
		if !ok {
			return false
		}
		if !a.Equal(b) {
			return false
		}
	}

	return true
}

func NewRecord() *Record {
	return &Record{Fields: make(map[string]Object)}
}

// EqualObjects compares two Object slices for equality.
func EqualObjects(a, b []Object) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if !v.Equal(b[i]) {
			return false
		}
	}

	return true
}

// EqualObjectSet compares two Object slices for equality ignoring the order of
// the elements.
//
// Current implementation is O(n^2).
func EqualObjectSet(a, b []Object) bool {
	if len(a) != len(b) {
		return false
	}

	for _, v := range a {
		for _, v2 := range b {
			if !v.Equal(v2) {
				return false
			}
		}
	}
	return true
}
