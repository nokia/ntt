package ntt_test

import (
	"os"
	"testing"

	"github.com/nokia/ntt/internal/ntt"
	"github.com/stretchr/testify/assert"
)

func TestTimeout(t *testing.T) {
	t.Run("Env", func(t *testing.T) {
		defer os.Unsetenv("NTT_TIMEOUT")

		suite := &ntt.Suite{}
		v, err := suite.Timeout()
		assert.Nil(t, err)
		assert.Zero(t, v)

		os.Setenv("NTT_TIMEOUT", "0")
		v, err = suite.Timeout()
		assert.Nil(t, err)
		assert.Zero(t, v)

		os.Setenv("NTT_TIMEOUT", "0.0")
		v, err = suite.Timeout()
		assert.Nil(t, err)
		assert.Zero(t, v)

		os.Setenv("NTT_TIMEOUT", "23.5")
		v, err = suite.Timeout()
		assert.Nil(t, err)
		assert.Equal(t, float64(23.5), v)

		os.Setenv("NTT_TIMEOUT", "some-string")
		v, err = suite.Timeout()
		assert.NotNil(t, err)
		assert.Zero(t, v)
	})

	t.Run("Root", func(t *testing.T) {
		os.Unsetenv("NTT_TIMEOUT")
		defer os.Unsetenv("NTT_TIMEOUT")

		suite := &ntt.Suite{}

		suite.SetRoot("./not_existent/")
		v, err := suite.Timeout()
		assert.Nil(t, err)
		assert.Zero(t, v)

		suite.SetRoot(".")
		f := suite.File("./package.yml")
		f.SetBytes([]byte(`timeout: 23.5`))
		v, err = suite.Timeout()
		assert.Nil(t, err)
		assert.Equal(t, float64(23.5), v)

		f.SetBytes([]byte(`timeout: hello master`))
		v, err = suite.Timeout()
		assert.NotNil(t, err)

		os.Setenv("NTT_TIMEOUT", "5.72")
		v, err = suite.Timeout()
		assert.Nil(t, err)
		assert.Equal(t, float64(5.72), v)
	})
}

func TestSources(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		suite := &ntt.Suite{}
		v, err := suite.Sources()
		assert.Nil(t, err)
		assert.Equal(t, 0, len(v))
	})

	t.Run("Files", func(t *testing.T) {
		suite := &ntt.Suite{}

		// Now we are adding two source manually without having a root folder.
		// Multiple calls to Sources should not change the number of Sources.
		suite.AddSources("a.ttcn3", "b.ttcn3")
		v, err := suite.Sources()
		for i := 0; i < 10; i++ {
			v, err = suite.Sources()
		}
		assert.Nil(t, err)
		assert.Equal(t, []string{"a.ttcn3", "b.ttcn3"}, strs(v))

		// Identical files may be added twice.
		suite.AddSources("a.ttcn3", "b.ttcn3")
		v, err = suite.Sources()
		assert.Nil(t, err)
		assert.Equal(t, []string{"a.ttcn3", "b.ttcn3", "a.ttcn3", "b.ttcn3"}, strs(v))

		// Environment shall overwrite configured sources.
		os.Setenv("NTT_SOURCES", "x.ttcn3	y.ttcn3")
		v, err = suite.Sources()
		assert.Nil(t, err)
		assert.Equal(t, []string{"x.ttcn3", "y.ttcn3"}, strs(v))
		os.Unsetenv("NTT_SOURCES")

		// Setting root removes all previous configured sources.
		//
		// This root contains just some .ttcn3 files without manifest.
		suite.SetRoot("./testdata/suite1")
		suite.AddSources("a.ttcn3", "b.ttcn3")
		v, err = suite.Sources()
		assert.Nil(t, err)
		assert.Equal(t, []string{"testdata/suite1/a.ttcn3", "testdata/suite1/x.ttcn3", "a.ttcn3", "b.ttcn3"}, strs(v))

		// This root contains a manifest.
		suite.SetRoot("./testdata/suite2")
		suite.AddSources("a.ttcn3", "b.ttcn3")
		v, err = suite.Sources()
		assert.Nil(t, err)
		assert.Equal(t, []string{
			"testdata/suite2/a1.ttcn3",
			"testdata/suite2/a2.ttcn3",
			"testdata/suite2/dir1/a3.ttcn3",
			"a.ttcn3",
			"b.ttcn3",
		}, strs(v))
	})

}

func strs(files []*ntt.File) []string {
	ret := make([]string, len(files))
	for i := range files {
		ret[i] = files[i].Path()
	}
	return ret
}
