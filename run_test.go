package ntt

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
			os.Setenv("NTT_PARAMETERS_FILE", tt.file)
			os.Setenv("NTT_PARAMETERS_DIR", "testdata/parameters")
			defer os.Unsetenv("NTT_PARAMETERS_FILE")
			defer os.Unsetenv("NTT_PARAMETERS_DIR")
			got, _ := testModulePars(t, tt.name)
			assert.Equal(t, tt.want, got)
		})
	}
}

func testModulePars(t *testing.T, name string) ([]string, error) {
	suite, err := NewSuite("testdata/parameters")
	if err != nil {
		t.Fatalf("NewSuite() failed: %v", err)
	}

	m, _, err := suite.TestParameters(name)
	var s []string
	for k, v := range m {
		s = append(s, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(s)
	return s, err
}
