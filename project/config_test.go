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

func TestGetVariables(t *testing.T) {
	c := project.Config{
		Sources: []string{"1", "2", "3"},
		Variables: map[string]string{
			"foo":     "fromConf",
			"name":    "fromConf",
			"sources": "a b c",
		},
	}
	t.Run("invalid", func(t *testing.T) {
		v, err := c.Get("fo")
		assert.Nil(t, v)
		assert.True(t, errors.Is(err, project.ErrNoSuchName))
	})
	t.Run("simple", func(t *testing.T) {
		v, err := c.Get("foo")
		assert.Equal(t, "fromConf", v)
		assert.Nil(t, err)
	})
	t.Run("simple env", func(t *testing.T) {
		os.Setenv("NTT_FOO", "fromEnv")
		defer os.Unsetenv("NTT_FOO")
		v, err := c.Get("foo")
		assert.Equal(t, "fromEnv", v)
		assert.Nil(t, err)
	})
	t.Run("conflict", func(t *testing.T) {
		v, err := c.Get("name")
		assert.Equal(t, "", v)
		assert.Nil(t, err)
	})
	t.Run("conflict", func(t *testing.T) {
		v, err := c.Get("sources")
		assert.Equal(t, []string{"1", "2", "3"}, v)
		assert.Nil(t, err)
	})
	t.Run("conflict env", func(t *testing.T) {
		os.Setenv("NTT_SOURCES", "x y z")
		defer os.Unsetenv("NTT_SOURCES")
		v, err := c.Get("sources")
		assert.Equal(t, []string{"x", "y", "z"}, v)
		assert.Nil(t, err)
	})
}
