package builtins

import "github.com/nokia/ntt/runtime"

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
		default:
			return matchRecordOf(a.(*runtime.List), b)
		}
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

func matchRecordOf(a, b *runtime.List) (bool, error) {
	/*
	 * Backtrack to previous * on mismatch and retry starting one
	 * object later in the list.  Because * matches all objects
	 * (no exception for /), it can be easily proven that there's
	 * never a need to backtrack multiple levels.
	 */
	var back_pat, back_str = -1, -1
	str := a.Elements
	pat := b.Elements

	/*
	 * Loop over each object in pat, matching
	 * it against the remaining unmatched tail of str.  Return false
	 * on mismatch, or true after matching the trailing nul bytes.
	 */
	for i, j := 0, 0; ; {
		var (
			aContinues                = len(str) > i
			bContinues                = len(pat) > j
			c, d       runtime.Object = runtime.Undefined, runtime.Undefined
		)
		if aContinues {
			c = str[i]
		}
		if bContinues {
			d = pat[j]
		}
		i++
		j++

		switch d {
		case runtime.Any: /* Wildcard: anything but nul */
			if !aContinues {
				return false, runtime.Errorf("End of first RecordOf reached, second still continues")
			}
		case runtime.AnyOrNone: /* Any-length wildcard */
			if moreAfter := len(pat) > j; !moreAfter { /* Optimize trailing * case */
				return true, nil
			}
			back_pat = j
			i-- /* Allow zero-length match */
			back_str = i
		default: /* Literal character */
			if !aContinues && !bContinues {
				return true, nil
			}
			if ok, _ := match(c, d); ok {
				break
			}
			if !aContinues {
				return false, runtime.Errorf("End of first RecordOf reached, second still continues")
			}
			if back_pat < 0 {
				return false, runtime.Errorf("Pattern doesn't match, Element number %d mismatch", i-1) /* No Backtracking possible */
			}

			/* Try again from last *, one character later in str. */
			j = back_pat
			back_str++
			i = back_str
		}
	}
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
		isMissing = false
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
		isMissing = false
		isAny     = -1
	)
	for _, valueA := range a.Elements {
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
		err := runtime.Errorf("At least one %s or '?' missing in second List", valueA)
		return false, err
	}
	return true, nil
}
