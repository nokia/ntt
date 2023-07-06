// Package types implements the type system for TTCN-3.
//
// # TTCN-3 Types
//
// A type is the set of values an expression can have. For example, the
// `integer` type is the set of all whole numbers, the type `integer
// (0..infinity)` is the set of all non-negative whole numbers.
//
// Two types are equal if their sets of possible values are equal. For example:
//
//	integer == integer
//	integer == integer (-infinity..infinity)
//	integer != integer (0..255)
//	record { integer a } != record { integer b }
//	record { integer a, integer b } != record { integer b, integer a }
//
// A type is called a *subtype* of another type if its set of possible values
// is a subset of the other type's set of possible values. For example:
//
//	integer (0..255) < integer
//	boolean (false)  < boolean
//
// Subtypes with an equal set of possible values are also called *synonym* or
// *alias*:
//
//	type integer int // `int` is an alias for `integer`
//
// In TTCN-3 subtypes are created by restricting the set of possible values
// using *constraints*. Examples:
//
//	type integer holeyInt (0..10, 20, 30, 40, 100..200);
//      type charstring ("a".."z", "0..9", pattern "A.*") length(2..10);
//
// A type that has a name is called a *named type*. Types without a name are
// called *anonymous types* or *nested types*. For example:
//
//	var record of record { integer a, integer b } MyVar;
//			       ^~~~~~~
//				named type
//		      ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
//		       nested type
//	    ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
//	     anonymous type
//
// The basic types supported by TTCN-3 are:
//   - numerical types `integer` and `float`
//   - `boolean` type with possible values `false`, `true`
//   - `verdicttype` with possible values `none`, `pass`, `inconc`, `fail` and
//     `error`
//   - strings types `charstring`, `universal charstring`, `hexstring`,
//     `octetstring` and `bitstring`
//
// A *list type* is a collection of elements of the same type. TTCN-3
// distinguishes five kinds of list types:
//   - `record of`: an ordered list
//   - `set of`: an unordered list
//   - *array*: an ordered list with a fixed number of elements
//   - `map`: an unordered list of key-value pairs
//   - *string types*: an ordered list of string type elements. For example, the
//     type of a bitstring elements is itself a bitstring.
//
// A *structured type* is a finite collection of named elements of possibly
// different types:
//   - `record`: a structured type with ordered fields.
//   - `set`: a structured type with unordered fields.
//
// A *union type* is the union of two or more types. For example `union {
// integer, boolean }` is the set of all possible integer and boolean values.
// The individual types of a union are called *union alternatives*.
// Union types are also known as `variant types`, `sum types` or `algebraic
// types`.
//
// TTCN-3 differentiates between *data types* and *reference types*. Data type
// values are fully self-describing and are indistinguishable from each other. In
// contrast to values of reference types, which have an identity.
// For example:
//
//	timer a := 1;
//	timer b := 1;
//	a == b; // false, because each timer has its own identity
//
//	integer a := 1;
//	integer b := 1;
//	a == b; // true, because integers are indistinguishable from each other.
//
// Available reference types in TTCN-3 are:
//   - `timer`
//   - `port`
//   - `default`
//   - `trait`
//   - `configuration`
//   - objects
//   - components
//   - behavior types
//
// The value of a *behavior type* refers to a function, testcase, altstep or an anonymous behavior:
//
//	type function Logger(any args...);
//	var Handler h;
//
//	h := log; // h refers to the log function
//	h := function(any args...) { log(args...); }; // anonymous function
//
// Anonymous behaviors are also known as *lambda expressions*, *closures* or *behaviour literals*.
//
// # Type Compatibility
//
// Type compatibility is required for assignment, parameter passing and comparison.
//
// Type A is compatible with type B, if A equals B:
//
//      type integer A;
//      type integer B;
//      var B b := 1; // OK, because integer == integer

// Type A is compatible with type B, if A is a subtype of B:
//
//	type integer A (0..255);
//	type integer B (0..512);
//	var A a := 1;
//	var B b := a; // OK, because integer(0..255) < integer(0..512)
//
// Types with different base type are not compatible:
//
//	// NOT OK
//	type integer A;
//	type float B;
//
// List types are compatible if their element types are also compatible.
//
// Structured types are compatible if their fields are also compatible. The
// name of the individual fields is not relevant, but the order is:
//
//	// OK
//	type record A { integer(0..255) fa1, float fa2 };
//	type record B { integer(0..512) fb1, float fa2 };
//
//	// NOT OK
//	type record A { integer i, float f };
//	type record B { float f, integer i };
//
// `inout` parameters and receive operations require nominal equality:
//
// Type A is compatible with type B if and only if A has the same type name as
// B.
//
// TODO(5nord) Describe compatibility rules for remaining types (enums, unions, ...)
//
// # Type Operations
//
// Each type has a set of operations that can be applied to its values. For example:
//   - list types have an index operation, concatenation, `lengthof`
//   - numerical types have arithmetic operations like `+`, `-`, `*`
//   - behavior types must callable and support the `apply` operation
//   - testcases must be executable and support the `stop` operation
//   - and many more..
//
// # Type Parameterization
//
// TTCN-3 supports *type parameterization*, also known as *type-polymorphism*,
// *generics* or *(C++) templates*.
//
// TODO: Describe type parameterization
//
// # Type Inference
//
// TODO(5nord) Describe type inference
// TODO(5nord) Describe templates
package types

import "github.com/nokia/ntt/ttcn3/syntax"

// Type represents a TTCN-3 type.
type Type interface {

	// String returns the type name. If no name is available, an string
	// representation of the type is returned.
	String()

	// Underlying returns the underlying type or nil if the type is a root type.
	Underlying() Type

	// Compatible return true if the type is compatible with the given
	// type.
	//
	// For example to check if a value of type A can be assigned to a
	// variable of type B, you would write:
	//
	//      if A.Compatible(B) { ... }
	//
	Compatible(Type) bool
}

// TypeOf infers the type of the given expression. If the type cannot be
// inferred, nil is returned.
func TypeOf(expr syntax.Expr) Type {
	return nil
}
