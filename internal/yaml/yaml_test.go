package yaml_test

import (
	"strings"
	"testing"
	"time"

	"github.com/nokia/ntt/internal/yaml"
	"github.com/stretchr/testify/assert"
)

type Struct struct {
	Name      string        `json:",omitempty"`
	CamelCase string        `json:"camel_case,omitempty"`
	Timeout   yaml.Duration `json:",omitempty"`
}

type MyConfig struct {
	Timeout    float64           `yaml:"timeout,omitempty"`
	Parameters map[string]string `yaml:"parameters,omitempty"`
}
type EmbeddingStruct struct {
	Struct `json:",inline"`
}

// TestUnmarshal verfies that the YAML package behaves as expected.
func TestUnmarshal(t *testing.T) {
	t.Run("support embedded fields", func(t *testing.T) {
		var e EmbeddingStruct
		err := yaml.Unmarshal([]byte(`name: Foo`), &e)
		assert.Nil(t, err)
		assert.Equal(t, "Foo", e.Name)
	})

	t.Run("map fields with underscore", func(t *testing.T) {
		var s Struct
		err := yaml.Unmarshal([]byte(`camel_case: Foo`), &s)
		assert.Nil(t, err)
		assert.Equal(t, "Foo", s.CamelCase)
	})

	t.Run("parse float as time.Duration", func(t *testing.T) {
		var s Struct
		err := yaml.Unmarshal([]byte(`timeout: 1.5`), &s)
		assert.Nil(t, err)
		assert.Equal(t, 1.5, s.Timeout.Seconds())
	})

	t.Run("report unknown fields", func(t *testing.T) {
		var s Struct
		err := yaml.Unmarshal([]byte(`foobar: 1.5`), &s)
		assert.NotNil(t, err)
	})

	t.Run("report type errror", func(t *testing.T) {
		var s Struct
		err := yaml.Unmarshal([]byte(`timeout: Foo`), &s)
		assert.NotNil(t, err)
	})

	t.Run("test multi-line quotes", func(t *testing.T) {
		var s Struct
		err := yaml.Unmarshal([]byte(`name: |  Foo\n  Bar`), &s)
		assert.Nil(t, err)
		assert.Equal(t, `Foo\n  Bar`, s.Name)
	})

	t.Run("test params without quotes", func(t *testing.T) {
		var c MyConfig
		err := yaml.Unmarshal([]byte(`timeout: 3.42e-5
parameters:
  mpar.anInt: -123

  mpar.aBool: true
`), &c)
		assert.Nil(t, err)
		assert.Equal(t, 3.42e-5, c.Timeout)
		assert.Equal(t, "-123", c.Parameters["mpar.anInt"])
		assert.Equal(t, "true", c.Parameters["mpar.aBool"])
	})

	t.Run("test params w/o quotes w/o empty lines", func(t *testing.T) {
		var c MyConfig
		err := yaml.Unmarshal([]byte(`timeout: 3.42e-5
parameters:
  mpar.anInt: -123
  mpar.aBool: true
`), &c)
		assert.Nil(t, err)
		assert.Equal(t, 3.42e-5, c.Timeout)
		assert.Equal(t, "-123", c.Parameters["mpar.anInt"])
		assert.Equal(t, "true", c.Parameters["mpar.aBool"])
	})

	t.Run("test params w multi line strings", func(t *testing.T) {
		var c MyConfig
		err := yaml.Unmarshal([]byte(`timeout: 3.42e-5
parameters:
  mpar.aCompound1: |
      { hex := '0a0a'H, str := "a string"}
  mpar.aCharstring: |
      "Hello World"
`), &c)
		assert.Nil(t, err)
		assert.Equal(t, 3.42e-5, c.Timeout)
		assert.Equal(t, "{ hex := '0a0a'H, str := \"a string\"}\n", c.Parameters["mpar.aCompound1"])
		assert.Equal(t, "\"Hello World\"\n", c.Parameters["mpar.aCharstring"])
	})
}

func TestMarshalJSON(t *testing.T) {
	marshal := func(v interface{}) (string, error) {
		b, err := yaml.MarshalJSON(v)
		return strings.TrimSpace(string(b)), err
	}

	t.Run("support embedded fields", func(t *testing.T) {
		b, err := marshal(EmbeddingStruct{
			Struct: Struct{
				Name: "Foo",
			},
		})
		assert.Nil(t, err)
		assert.Equal(t, `{"name": "Foo"}`, b)
	})

	t.Run("use fields with underscore", func(t *testing.T) {
		b, err := marshal(Struct{CamelCase: "Foo"})
		assert.Nil(t, err)
		assert.Equal(t, `{"camel_case": "Foo"}`, b)
	})

	t.Run("encode time.Duration as float", func(t *testing.T) {
		b, err := marshal(Struct{Timeout: yaml.Duration{time.Millisecond * 100}})
		assert.Nil(t, err)
		assert.Equal(t, `{"timeout": 0.1}`, b)
	})
}
