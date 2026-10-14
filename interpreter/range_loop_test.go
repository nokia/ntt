package interpreter_test

import (
	"testing"

	"github.com/nokia/ntt/runtime"
)

// TestRangeLoop runs the range-based loop (ETSI 19.4.2): the range is
// evaluated once, its elements assigned in order, an uninitialised one
// leaving the variable unbound; a variable declared outside keeps the
// last element. It was parsed and never run.
func TestRangeLoop(t *testing.T) {
	assertSameVerdictBothClocks(t, `module M {
		type component C {}
		type record of integer RI;
		type record of charstring RC;
		type integer A3[1..3];
		type record R { integer f }
		type record of R RR;
		type map from charstring to integer MT;
		function f(inout integer calls) return RI { calls := calls + 1; return {7, 8} }
		function firstBig(RI a) return integer {
			for (var integer x in a) { if (x > 1) { return x } }
			return -1
		}
		testcase tc() runs on C {
			var charstring s := "";
			for (var hexstring e in '12D'H) { s := s & hex2str(e) }
			for (var octetstring e in '12DE'O) { s := s & oct2str(e) }
			for (var bitstring e in '10'B) { s := s & bit2str(e) }
			for (var charstring c in "xy") { s := s & c }
			if (s != "12D12DE10xy") { setverdict(fail, "strings: ", s) }

			var integer i := 0, n := 0;
			var charstring e;
			for (e in RC:{"a", -, "c"}) { if (i == 1 and isbound(e)) { setverdict(fail, "bound") } i := i + 1 }
			if (i != 3 or e != "c") { setverdict(fail, "outer variable: ", e) }

			var integer calls := 0;
			for (var integer c in f(calls)) { n := n + c }
			if (calls != 1 or n != 15) { setverdict(fail, "range evaluated ", calls, " times") }

			i := 0;
			var RI a := {1, 2, 3};
			for (var integer x in a) { a[i+1] := 10*x; i := i + 1 }
			if (a != {1, 10, 20, 30}) { setverdict(fail, "range not evaluated once: ", a) }

			var A3 arr := {4, 5, 6};
			n := 0;
			for (var integer x in arr) { n := n * 10 + x }
			if (n != 456) { setverdict(fail, "array: ", n) }

			n := 0;
			for (var integer x in RI:{1, 2, 3, 4, 5}) { if (x == 2) { continue } if (x == 4) { break } n := n + x }
			if (n != 4) { setverdict(fail, "break/continue: ", n) }
			if (firstBig({0, 1, 5, 9}) != 5) { setverdict(fail, "return") }

			var RR rr := {{f := 1}, {f := 2}};
			for (var R r in rr) { r.f := 99 }
			if (rr != {{f := 1}, {f := 2}}) { setverdict(fail, "element aliased: ", rr) }

			var MT m := { ["b"] := 2, ["a"] := 1 };
			n := 0;
			for (var charstring k in m.from) { n := n + m[k] }
			for (var integer v in m.to) { n := n + 10 * v }
			if (n != 33 or m.from != {"a", "b"}) { setverdict(fail, "map: ", n) }
			setverdict(pass)
		}
	}`, runtime.PassVerdict)
}

// TestRangeLoopRepeat: a repeat in a range loop of an alt branch repeats
// the alt (ETSI 20.2).
func TestRangeLoopRepeat(t *testing.T) {
	assertSameVerdictBothClocks(t, `module M {
		type component C { timer t }
		type record of integer RI;
		testcase tc() runs on C {
			var integer k := 0;
			t.start(0.1);
			alt {
				[] t.timeout {
					k := k + 1;
					for (var integer x in RI:{1, 2}) { if (k < 3) { t.start(0.1); repeat } }
					if (k == 3) { setverdict(pass) } else { setverdict(fail, k) }
				}
			}
		}
	}`, runtime.PassVerdict)
}

// TestAltstepAlternativeBeforeAGuard: `[] as1()` takes no block, and the
// guarded alternative after it is one of the alt's (it was read as an
// index of the altstep's result, which panicked).
func TestAltstepAlternativeBeforeAGuard(t *testing.T) {
	assertSameVerdictBothClocks(t, `module M {
		type port P message { inout integer }
		type component C { port P p; timer t }
		altstep as1() runs on C { [] p.receive { setverdict(fail, "received") } }
		testcase tc() runs on C {
			var boolean g := true;
			t.start(0.1);
			alt {
				[] as1()
				[g] t.timeout { setverdict(pass) }
			}
		}
	}`, runtime.PassVerdict)
}
