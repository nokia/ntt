package builtins

import "github.com/nokia/ntt/runtime"

type sliceHolder interface {
	Get(index int) runtime.Object
	Len() int
}

// match returns true given objects match.
func match(a, b runtime.Object) (bool, error) {
	if a == runtime.Any || a == runtime.AnyOrNone || b == runtime.Any || b == runtime.AnyOrNone {
		return true, nil
	}
	if a.Type() != b.Type() {
		return false, runtime.Errorf("type mismatch: %s != %s", a.Type(), b.Type())
	}

	switch b := b.(type) {
	case *runtime.Record:
		return matchRecord(a.(*runtime.Record), b)
	case *runtime.List:
		switch b.ListType {
		case runtime.SET_OF:
			return matchSetOf(a.(*runtime.List), b)
		case runtime.SUPERSET:
			return matchIsASupersetB(a.(*runtime.List), b)
		case runtime.SUBSET:
			return matchIsASubsetB(a.(*runtime.List), b)
		default:
			return matchRecordOf(a.(*runtime.List), b)
		}
	case *runtime.String:
		//Leaving out this Comparision would create an endless loop where the same
		//function is called with the same parameters recursively
		if c := a.(*runtime.String); c.Len() == 1 && b.Len() == 1 {
			if b.Value[0] == rune('?') || b.Value[0] == rune('*') || b.Value[0] == c.Value[0] {
				return true, nil
			}
			return false, runtime.Errorf("tried to match %d with %d", c.Value[0], b.Value[0])

		}
		return matchRecordOf(a.(*runtime.String), b)
	default:
		return a.Equal(b), nil
	}
}

// matchRecord returns true given records match.
func matchRecord(a, b *runtime.Record) (bool, error) {
	if len(b.Fields) != len(a.Fields) {
		return false, runtime.Errorf("Records don't have equal amounts of Fields")
	}
	for k, y := range b.Fields {
		if x, ok := a.Fields[k]; ok {
			if ret, err := match(x, y); !ret {
				return false, err
			}
		} else {
			return false, runtime.Errorf("Value mismatch: field %s in second Record not found in first", k)
		}
	}
	return true, nil
}

// matchRecordOf returns true if recordOfs match
func matchRecordOf(val, pat sliceHolder) (bool, error) {
	i, back_i := 0, -1
	j, back_j := 0, -1
	for i < val.Len() && j < pat.Len() {
		// if pat is *runtime.String, Get(i) returns another *runtime.String
		// whose Value array contains a single rune
		if pat.Get(j) == runtime.AnyOrNone || pat.Get(j).Equal(runtime.NewCharstring("*")) {
			j++
			back_j = j          // Pattern Element after *
			back_i = i          // First Value Element which could be matched with that *
			if j == pat.Len() { // Optimize trailing * case
				return true, nil
			}
		} else if ok, _ := match(val.Get(i), pat.Get(j)); !ok { // Literal character or ?
			if back_j < 0 {
				return false, runtime.Errorf("Pattern doesn't match, Element number %d mismatch", i-1) /* No Backtracking possible */
			}
			// Try again from last *, one character later in str.
			j = back_j
			back_i++
			i = back_i
		} else {
			i++
			j++
		}
		if j == pat.Len() && i != val.Len() {
			if back_j < 0 {
				return false, runtime.Errorf("Second RecordOf is matched entirely, first isn't")
			}
			// Try again from last *, one character later in str.
			j = back_j
			back_i++
			i = back_i
		}
	}
	// reached if i == len(val) || j == len(pat)
	if val.Len() == i {
		for ; j < pat.Len(); j++ {
			if pat.Get(j) != runtime.AnyOrNone && !pat.Get(j).Equal(runtime.NewCharstring("*")) {
				return false, runtime.Errorf("First RecordOf is entirely matched, second isn't")
			}
		}
		return true, nil
	}
	// reached if i != len(val) && j == len(pat) == 0 (non-zero case covered in loop)
	return false, runtime.Errorf("Template empty")
}

// matchSetOf returns true given sets match.
func matchSetOf(a, b *runtime.List) (bool, error) {
	containsStar := false
	temp := runtime.NewSetOf()
	for _, y := range b.Elements {
		if y == runtime.AnyOrNone {
			containsStar = true
		} else {
			temp.Elements = append(temp.Elements, y)
		}
	}
	if !containsStar && len(a.Elements) > len(temp.Elements) {
		return false, runtime.Errorf("First List contains more Elements than second")
	}
	if len(a.Elements) < len(temp.Elements) {
		return false, runtime.Errorf("First List doesn't contain enough elements")
	}
	return matchIsASupersetB(a, temp)
}

// matchIsASupersetB returns true if a is a superset of b
func matchIsASupersetB(a, b *runtime.List) (bool, error) {
	var (
		cloneA    = a
		isMissing = true
		numOfAny  = 0
	)

	for _, valueB := range b.Elements {
		if valueB == runtime.AnyOrNone {
			continue
		}
		if valueB == runtime.Any {
			numOfAny++
			continue
		}
		isMissing = true
		for i, valueA := range cloneA.Elements {
			if ok, _ := match(valueA, valueB); ok {
				cloneA.Elements[i] = cloneA.Elements[len(cloneA.Elements)-1]
				cloneA.Elements = cloneA.Elements[:len(cloneA.Elements)-1]
				isMissing = false
				break
			}
			isMissing = true
		}
		if !isMissing {
			continue
		}
		return false, runtime.Errorf("At least one %s missing in first List", valueB)
	}
	if len(cloneA.Elements) < numOfAny {
		return false, runtime.Errorf("%d element/s missing in first List", numOfAny-len(cloneA.Elements))
	}
	return true, nil
}

// matchIsASubsetB returns true if a is a subset of b
func matchIsASubsetB(a, b *runtime.List) (bool, error) {
	var (
		cloneB    = b
		isMissing = true
		isAny     = -1
	)
	for _, valueA := range a.Elements {
		isMissing = true
		for i, valueB := range cloneB.Elements {
			if valueB == runtime.AnyOrNone {
				return true, nil
			}

			if valueB == runtime.Any {
				isAny = i
			} else if ok, _ := match(valueA, valueB); ok {
				cloneB.Elements[i] = cloneB.Elements[len(cloneB.Elements)-1]
				cloneB.Elements = cloneB.Elements[:len(cloneB.Elements)-1]
				isMissing = false
				isAny = -1
				break
			}
			isMissing = true
		}
		if !isMissing {
			continue
		}
		if isAny >= 0 {
			cloneB.Elements[isAny] = cloneB.Elements[len(cloneB.Elements)-1]
			cloneB.Elements = cloneB.Elements[:len(cloneB.Elements)-1]
			isMissing = false
			isAny = -1
			continue
		}
		return false, runtime.Errorf("At least one %s or '?' missing in second List", valueA)
	}
	return true, nil
}
