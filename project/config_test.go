package project_test

import (
	"errors"
	"os"
	"strconv"
	"testing"

	"github.com/nokia/ntt/project"
	"github.com/stretchr/testify/assert"
)

func TestConfigGet(t *testing.T) {
	t.Run("empty name", func(t *testing.T) {
		c := project.Config{}
		v, err := c.Get("")
		assert.Nil(t, v)
		assert.True(t, errors.Is(err, project.ErrNoSuchName))
	})

	t.Run("invalid name", func(t *testing.T) {
		c := project.Config{}
		v, err := c.Get("!!!")
		assert.Nil(t, v)
		assert.True(t, errors.Is(err, project.ErrNoSuchName))
	})

	t.Run("prefix name", func(t *testing.T) {
		c := project.Config{}
		v, err := c.Get("NTT_NAME")
		assert.Nil(t, v)
		assert.True(t, errors.Is(err, project.ErrNoSuchName))
	})

	t.Run("simple name", func(t *testing.T) {
		c := project.Config{}
		v, err := c.Get("name")
		assert.Equal(t, "", v)
		assert.Nil(t, err)

		c.Name = "a name"
		v, err = c.Get("name")
		assert.Equal(t, "a name", v)
		assert.Nil(t, err)
	})

	t.Run("name alias", func(t *testing.T) {
		c := project.Config{}
		c.ParametersFile = "foo"
		v, err := c.Get("parameters_file")
		assert.Equal(t, "foo", v)
		assert.Nil(t, err)
	})

	t.Run("environment", func(t *testing.T) {
		os.Setenv("NTT_PARAMETERS_FILE", "fromEnv")
		defer os.Unsetenv("NTT_PARAMETERS_FILE")

		c := project.Config{}
		c.ParametersFile = "foo"
		v, err := c.Get("parameters_file")
		assert.Equal(t, "fromEnv", v)
		assert.Nil(t, err)
	})

	t.Run("environment types", func(t *testing.T) {
		os.Setenv("NTT_SOURCES", " a b  c")
		defer os.Unsetenv("NTT_SOURCES")

		c := project.Config{}
		v, err := c.Get("sources")
		assert.Equal(t, []string{"a", "b", "c"}, v)
		assert.Nil(t, err)
	})

	t.Run("invalid environment types", func(t *testing.T) {
		os.Setenv("NTT_TIMEOUT", " a b  c")
		defer os.Unsetenv("NTT_TIMEOUT")

		c := project.Config{}
		v, err := c.Get("timeout")
		assert.Nil(t, v)
		assert.True(t, errors.Is(err, strconv.ErrSyntax), err)
	})
}
