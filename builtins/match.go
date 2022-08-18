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

	if r, ok := b.(*runtime.Record); ok {
		return matchRecord(a.(*runtime.Record), r)
	}

	if r, ok := b.(*runtime.List); ok {
		return matchSetOf(a.(*runtime.List), r)
	}

	return a.Equal(b), nil
}

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

func matchSetOf(a, b *runtime.List) (bool, error) {
	containsStar := false
	var temp *runtime.List = runtime.NewList(len(b.Elements))
	temp.Elements = temp.Elements[0:0]
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

/*
Checks if List a is a Superset of List b.
*/
func matchIsASupersetB(a, b *runtime.List) (bool, error) {
	var cloneA *runtime.List = a
	isMissing, numOfAny := false, 0
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

/*
Checks if List a is a Subset of List b.
*/
func matchIsASubsetB(a, b *runtime.List) (bool, error) {
	var cloneB *runtime.List = b
	isMissing, isAny := false, -1
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
