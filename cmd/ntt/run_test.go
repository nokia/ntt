package main

import (
	"fmt"
	"os"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModulePars(t *testing.T) {
	// There are to paths how module parameters reach the runtime:
	// 1. From the default parameters file (K3_PARAMETERS_FILE)
	// 2. From the parameters directory (K3_PARAMETERS_DIR/$MODULE/$TEST.parameters)

	s, err := NewSuite("../../testdata/parameters")
	if err != nil {
		t.Fatalf("NewSuite() failed: %v", err)
	}

	tests := []struct {
		file string
		name string
		want []string
	}{
		{name: "", want: nil},
		{name: "xxx", want: nil},
		{name: "xxx.xxx", want: nil},
		{file: "xxx", want: nil},
		{file: "good.parameters", want: []string{"X=X from good", "Y=Y from good"}},
		{file: "good.parameters", name: "A", want: []string{"X=X from good", "Y=Y from good"}},
		{file: "good.parameters", name: "test.A", want: []string{"X=X from good", "Y=Y from good"}},
		{file: "good.parameters", name: "test.B", want: []string{"X=X from good", "Y=Y from B", "Z=Z from B"}},
		{file: "good.parameters", name: "test.C", want: nil},
		{file: "bad.parameters", want: nil},
		{file: "bad.parameters", name: "test.B", want: nil},
	}
	for _, tt := range tests {
		t.Run(t.Name(), func(t *testing.T) {
			os.Setenv("K3_PARAMETERS_FILE", tt.file)
			os.Setenv("K3_PARAMETERS_DIR", "../../testdata/parameters")
			defer os.Unsetenv("K3_PARAMETERS_FILE")
			defer os.Unsetenv("K3_PARAMETERS_DIR")
			got, _ := testModulePars(tt.name, s)
			assert.Equal(t, tt.want, got)
		})
	}
}

func testModulePars(name string, suite *Suite) ([]string, error) {
	m, err := ModulePars(name, suite)
	var s []string
	for k, v := range m {
		s = append(s, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(s)
	return s, err
}
