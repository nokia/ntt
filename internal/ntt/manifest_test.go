package ntt_test

import (
	"os"
	"testing"

	"github.com/nokia/ntt/internal/fs"
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
		f := fs.Open("./package.yml")
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
		v, _ := suite.Sources()
		for i := 0; i < 10; i++ {
			v, _ = suite.Sources()
		}
		assert.Equal(t, []string{"a.ttcn3", "b.ttcn3"}, v)

		// Identical files may be added twice.
		suite.AddSources("a.ttcn3", "b.ttcn3")
		v, _ = suite.Sources()
		assert.Equal(t, []string{"a.ttcn3", "b.ttcn3", "a.ttcn3", "b.ttcn3"}, v)

		// Setting root removes all previous configured sources.
		//
		// This root contains just some .ttcn3 files without manifest.
		suite.SetRoot("./testdata/suite1")
		suite.AddSources("a.ttcn3", "b.ttcn3")
		v, _ = suite.Sources()
		assert.Equal(t, []string{"testdata/suite1/a.ttcn3", "testdata/suite1/x.ttcn3", "testdata/suite1/a.ttcn3", "testdata/suite1/b.ttcn3"}, v)

		// This root contains a manifest.
		suite.SetRoot("./testdata/suite2")
		suite.AddSources("a.ttcn3", "b.ttcn3")
		v, _ = suite.Sources()
		assert.Equal(t, []string{
			"testdata/suite2/a1.ttcn3",
			"testdata/suite2/a2.ttcn3",
			"testdata/suite2/dir1/a3.ttcn3",
			"testdata/suite2/a.ttcn3",
			"testdata/suite2/b.ttcn3",
		}, v)
	})

}

func TestImports(t *testing.T) {
	suite := &ntt.Suite{}

	// This handle is used to overwrite package.yml with custom import testing
	// stuff.
	conf := fs.Open("./testdata/suite2/package.yml")
	conf.SetBytes([]byte(`imports: [ "dir1" ]`))
	suite.SetRoot("./testdata/suite2")
	v, err := suite.Imports()
	assert.Nil(t, err)
	assert.Equal(t, []string{"testdata/suite2/dir1"}, v)

	conf.SetBytes([]byte(`imports: [ "${SOMETHING_UNKNOWN}/dir1" ]`))
	suite.SetRoot("./testdata/suite2")
	v, err = suite.Imports()
	assert.Nil(t, err)
	assert.Equal(t, []string{"${SOMETHING_UNKNOWN}/dir1"}, v)
}

func TestName(t *testing.T) {
	suite := &ntt.Suite{}

	// Initial call to name shall return an error.
	n, _ := suite.Name()
	assert.Equal(t, "_", n)

	suite.AddSources("${SOMETHING}/dir1.ttcn3/foo.ttcn3", "bar", "fnord.ttcn3")
	n, _ = suite.Name()
	assert.Equal(t, "foo", n)

	suite.SetRoot("testdata/suite2")
	n, _ = suite.Name()
	assert.Equal(t, "suite2", n)

	suite.AddSources("${SOMETHING}/dir1.ttcn3/foo.ttcn3", "bar", "fnord.ttcn3")
	n, _ = suite.Name()
	assert.Equal(t, "suite2", n)

	conf := fs.Open("testdata/suite2/package.yml")
	conf.SetBytes([]byte("name: fnord"))
	suite.SetRoot("./testdata/suite2")
	n, _ = suite.Name()
	assert.Equal(t, "fnord", n)

	// This tests tells what happens when the package.yml has an syntax error
	// Underlying implementation does not exit on error, hence the name
	// wont be "" but something useable.
	conf.SetBytes([]byte(`name: [ 23.5, "See fnords, now!"]`))
	suite.SetRoot("testdata/suite2")
	n, _ = suite.Name()
	assert.Equal(t, "suite2", n)

	suite.SetName("haaraxwd")
	n, _ = suite.Name()
	assert.Equal(t, "haaraxwd", n)
}

func TestTestHook(t *testing.T) {
	suite := &ntt.Suite{}
	h, err := suite.TestHook()
	assert.Nil(t, err)
	assert.Nil(t, h)
}

func TestParametersFile(t *testing.T) {
	suite := &ntt.Suite{}
	h, err := suite.ParametersFile()
	assert.Nil(t, err)
	assert.Nil(t, h)

}
func strs(files []*fs.File) []string {
	ret := make([]string, len(files))
	for i := range files {
		ret[i] = files[i].String()
	}
	return ret
}
