package runtime

import (
	"bytes"
	"errors"
	"fmt"
	"hash/fnv"
	"math/big"
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
	FUNCTION     ObjectType = "function"
	LIST         ObjectType = "list"
	RECORD       ObjectType = "record"
	MAP          ObjectType = "map"
	BUILTIN_OBJ  ObjectType = "builtin function"
	VERDICT      ObjectType = "verdict"
	ANY          ObjectType = "?"
	ANY_OR_NONE  ObjectType = "*"

	Bit    Unit = 1
	Hex    Unit = 4
	Octett Unit = 8

	NoneVerdict   Verdict = "none"
	PassVerdict   Verdict = "pass"
	InconcVerdict Verdict = "inconc"
	FailVerdict   Verdict = "fail"
	ErrorVerdict  Verdict = "error"
)

type Unit int

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

type String struct {
	Value string
}

func (s *String) Type() ObjectType { return STRING }
func (s *String) Inspect() string  { return s.Value }

func (s *String) Equal(obj Object) bool {
	if other, ok := obj.(*String); ok {
		return s.Value == other.Value
	}
	return false
}

func (s *String) hashKey() hashKey {
	h := fnv.New64a()
	h.Write([]byte(s.Value))
	return hashKey{Type: s.Type(), Value: h.Sum64()}
}

func NewString(s string) *String {
	return &String{Value: s}
}

type Bitstring struct {
	Value *big.Int
	Unit  Unit
}

func (b *Bitstring) Type() ObjectType { return BITSTRING }
func (b *Bitstring) Inspect() string {
	switch b.Unit {
	case Bit:
		return fmt.Sprintf("'%b'B", b.Value)
	case Octett:
		return fmt.Sprintf("'%h'O", b.Value)
	default:
		return fmt.Sprintf("'%h'H", b.Value)
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
	switch strings.ToUpper(string(s[len(s)-1])) {
	case "B":
		unit = Bit
	case "H":
		unit = Hex
	case "O":
		unit = Octett
	default:
		return nil, ErrSyntax
	}

	removeWhitespaces := func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}
	s = strings.Map(removeWhitespaces, s[1:len(s)-2])

	if i, ok := new(big.Int).SetString(s, unit.Base()); ok {
		return &Bitstring{Value: i, Unit: unit}, nil
	}

	// TODO(5nord) parse and return Bitstring templates (e.g. '01*1'B)
	return nil, ErrSyntax
}

type List struct {
	Elements  []Object
	Unordered bool
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
	if l.Unordered || other.Unordered {
		return EqualObjectSet(l.Elements, other.Elements)
	}

	return EqualObjects(l.Elements, other.Elements)
}

func NewList(i int) *List {
	l := &List{}
	for ; i > 0; i-- {
		l.Elements = append(l.Elements, Undefined)
	}
	return l
}

func NewRecordOf() *List {
	return NewList(0)
}

func NewSetOf() *List {
	l := NewList(0)
	l.Unordered = true
	return l
}

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
