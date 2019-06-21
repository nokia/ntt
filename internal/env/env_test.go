package env

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type variable struct {
	key string
	val string
}

func initEnv(vars ...variable) func() {
	for _, v := range vars {
		os.Setenv(v.key, v.val)
	}
	return func() {
		for _, v := range vars {
			os.Unsetenv(v.key)
		}
	}
}

func TestGet(t *testing.T) {
	vars := []variable{
		{"ENVTEST_A", "b"},
		{"ENVTEST_SPECIAL", `p1.x=23 p1.y=5 A={"p1.x=\"fnord\"", ""}`},
		{"ENVTESTER_1", "2"},
		{"envtest_x", "y"},
		{"ENVTEST_x", "z"},
	}

	cleanup := initEnv(vars...)
	defer cleanup()

	e := New("envtest")
	e.Set("a", "over 6000")
	e.Set("ER_1", "hello")

	// a is overwritten by environment variables
	assert.Equal(t, "b", e.Get("a"))

	// case is fixed automatically
	assert.Equal(t, "b", e.Get("A"))

	// spaces special characters are handled correctly
	assert.Equal(t, `p1.x=23 p1.y=5 A={"p1.x=\"fnord\"", ""}`, e.Get("special"))

	// wrong prefix (missing underscore)
	assert.Equal(t, "hello", e.Get("ER_1"))

	// environment variables use upper case
	assert.Equal(t, "", e.Get("x"))
	assert.Equal(t, "", e.Get("x"))
}

func TestReset(t *testing.T) {
	vars := []variable{
		{"ENVTEST_A", "b"},
	}
	cleanup := initEnv(vars...)
	defer cleanup()

	e := New("envtest")
	e.Set("b", "c")
	e.Reset()

	// a is overwritten by environment variables
	assert.Equal(t, "b", e.Get("a"))

	// b is cleared
	assert.Equal(t, "", e.Get("b"))
}

func TestKeys(t *testing.T) {
	vars := []variable{
		{"ENVTEST_X", "yes"},
		{"ENVTEST_A", "2"},
		{"ENVTEST_B", "1"},
		{"ENVTEST_1", "true"},
		{"ENVTEST_D", "true"},
		{"envtest_e", "true"},
	}
	cleanup := initEnv(vars...)
	defer cleanup()

	e := New("envtest")
	e.Set("c", "true")
	e.Set("X", "true")
	e.Set("x", "true")

	keys := []string{
		"1",
		"a",
		"b",
		"c",
		"d",
		"x",
	}
	assert.Equal(t, keys, e.Keys())
}

func TestExpansion(t *testing.T) {
	vars := []variable{
		{"CFLAGS", "-I${ENVTEST_SOURCE_DIR}/"},
		{"SOMEVAR", "hello"},
		{"ENVTEST_A", "${SOMEVAR} ${ENVTEST_SOMEVAR}!"},
	}
	cleanup := initEnv(vars...)
	defer cleanup()

	e := New("envtest")
	e.Set("source_dir", "..")
	e.Set("somevar", "world")
	e.Set("b", "${ENVTEST_B}")
	e.Set("c", "${CFLAGS}")

	// Test expansion from environment and from internal map
	assert.Equal(t, "hello world!", e.Get("a"))

	// Test recursive expansion
	assert.Equal(t, "${ENVTEST_B}", e.Get("b"))

	// Test indirect expansion
	assert.Equal(t, "-I${ENVTEST_SOURCE_DIR}/", e.Get("c"))
}
