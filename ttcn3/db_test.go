package ttcn3_test

import (
	"testing"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/ttcn3"
)

var index_file1 = `
	module M1 { import from M2 all }
	module M2 { type enumerated E { E1 } }
`

var index_file2 = `
	module M1 { type component C { var integer x } }
	module M3 { const integer E1 := 1 }
`

type SliceMap map[string][]string

func TestIndex(t *testing.T) {
	fs.SetContent("file1.ttcn3", []byte(index_file1))
	fs.SetContent("file2.ttcn3", []byte(index_file2))
	db := ttcn3.DB{}
	db.Index("file1.ttcn3", "file2.ttcn3")
	testMapsEqual(t, makeSliceMap(db.Modules), SliceMap{
		"M1": []string{"file1.ttcn3", "file2.ttcn3"},
		"M2": []string{"file1.ttcn3"},
		"M3": []string{"file2.ttcn3"},
	})
	testMapsEqual(t, makeSliceMap(db.Names), SliceMap{
		"M1": []string{"file1.ttcn3", "file2.ttcn3"},
		"M2": []string{"file1.ttcn3"},
		"E":  []string{"file1.ttcn3"},
		"E1": []string{"file1.ttcn3", "file2.ttcn3"},
		"C":  []string{"file2.ttcn3"},
		"M3": []string{"file2.ttcn3"},
	})

}

func testMapsEqual(t *testing.T, a, b SliceMap) {
	if !equalSliceMap(a, b) {
		t.Errorf("Maps not equal:\n\t got = %v\n\twant = %v", a, b)
	}
}

func equalSliceMap(a, b SliceMap) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if !equal(v, b[k]) {
			return false
		}
	}
	return true
}

func makeSliceMap(m map[string]map[string]bool) SliceMap {
	sm := SliceMap{}
	for k, v := range m {
		sm[k] = make([]string, 0, len(v))
		for kk := range v {
			sm[k] = append(sm[k], kk)
		}
	}
	return sm
}
