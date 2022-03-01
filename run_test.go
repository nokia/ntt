package ntt_test

import (
	"testing"

	"github.com/nokia/ntt"
	nttold "github.com/nokia/ntt/internal/ntt"
	"github.com/stretchr/testify/assert"
)

func TestGenerators(t *testing.T) {
	srcs := testSources(t, "testdata/vanilla")

	t.Run("tests", func(t *testing.T) {
		expected := []string{
			"test.A",
			"test.B",
			"test2.C",
		}
		c := ntt.GenerateTests(srcs...)
		actual := consumeStringChannel(c)
		assert.Equal(t, expected, actual)
	})

	t.Run("controls", func(t *testing.T) {
		expected := []string{
			"test.control",
		}
		c := ntt.GenerateControls(srcs...)
		actual := consumeStringChannel(c)
		assert.Equal(t, expected, actual)
	})
}

func testSources(t *testing.T, dir string) []string {
	suite, err := nttold.NewFromArgs(dir)
	if err != nil {
		t.Fatal(err)
	}
	srcs, err := suite.Sources()
	if err != nil {
		t.Fatal(err)
	}
	return srcs
}

func consumeStringChannel(c <-chan string) []string {
	var result []string
	for s := range c {
		result = append(result, s)
	}
	return result
}
