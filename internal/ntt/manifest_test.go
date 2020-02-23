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
