package env_test

import (
	"bufio"
	"os"
	"strings"
	"testing"

	"github.com/nokia/ntt/internal/env"
	"github.com/stretchr/testify/assert"
)

var formats = []struct {
	in     string
	out    env.Env
	preset bool
}{
	// parses unquoted values
	{`FOO=bar`, env.Env{"FOO": "bar"}, false},

	// parses values with spaces around equal sign
	{`FOO =bar`, env.Env{"FOO": "bar"}, false},
	{`FOO= bar`, env.Env{"FOO": "bar"}, false},

	// parses values with leading spaces
	{`  FOO=bar`, env.Env{"FOO": "bar"}, false},

	// parses values with following spaces
	{`FOO=bar  `, env.Env{"FOO": "bar"}, false},

	// parses double quoted values
	{`FOO="bar"`, env.Env{"FOO": "bar"}, false},

	// parses double quoted values with following spaces
	{`FOO="bar"  `, env.Env{"FOO": "bar"}, false},

	// parses single quoted values
	{`FOO='bar'`, env.Env{"FOO": "bar"}, false},

	// parses single quoted values with following spaces
	{`FOO='bar'  `, env.Env{"FOO": "bar"}, false},

	// parses escaped double quotes
	{`FOO="escaped\"bar"`, env.Env{"FOO": `escaped"bar`}, false},

	// parses empty values
	{`FOO=`, env.Env{"FOO": ""}, false},

	// expands variables found in values
	{"FOO=test\nBAR=$FOO", env.Env{"FOO": "test", "BAR": "test"}, false},

	// parses variables wrapped in brackets
	{"FOO=test\nBAR=${FOO}bar", env.Env{"FOO": "test", "BAR": "testbar"}, false},

	// reads variables from ENV when expanding if not found in local env
	{`BAR=$FOO`, env.Env{"BAR": "test"}, true},

	// expands undefined variables to an empty string
	{`BAR=$FOO`, env.Env{"BAR": ""}, false},

	// expands variables in quoted strings
	{"FOO=test\nBAR=\"quote $FOO\"", env.Env{"FOO": "test", "BAR": "quote test"}, false},

	// does not expand variables in single quoted strings
	{"BAR='quote $FOO'", env.Env{"BAR": "quote $FOO"}, false},

	// does not expand escaped variables
	{`FOO="foo\$BAR"`, env.Env{"FOO": "foo$BAR"}, false},
	{`FOO="foo\${BAR}"`, env.Env{"FOO": "foo${BAR}"}, false},
	{"FOO=test\nBAR=\"foo\\${FOO} ${FOO}\"", env.Env{"FOO": "test", "BAR": "foo${FOO} test"}, false},

	// parses yaml style options
	{"OPTION_A: 1", env.Env{"OPTION_A": "1"}, false},

	// parses export keyword
	{"export OPTION_A=2", env.Env{"OPTION_A": "2"}, false},

	// allows export line if you want to do it that way
	{"OPTION_A=2\nexport OPTION_A", env.Env{"OPTION_A": "2"}, false},

	// expands newlines in quoted strings
	{`FOO="bar\nbaz"`, env.Env{"FOO": "bar\nbaz"}, false},

	// parses variables with "." in the name
	{`FOO.BAR=foobar`, env.Env{"FOO.BAR": "foobar"}, false},

	// strips unquoted values
	{`foo=bar `, env.Env{"foo": "bar"}, false}, // not 'bar '

	// ignores empty lines
	{"\n \t  \nfoo=bar\n \nfizz=buzz", env.Env{"foo": "bar", "fizz": "buzz"}, false},

	// ignores inline comments
	{"foo=bar # this is foo", env.Env{"foo": "bar"}, false},

	// allows # in quoted value
	{`foo="bar#baz" # comment`, env.Env{"foo": "bar#baz"}, false},

	// ignores comment lines
	{"\n\n\n # HERE GOES FOO \nfoo=bar", env.Env{"foo": "bar"}, false},

	// parses # in quoted values
	{`foo="ba#r"`, env.Env{"foo": "ba#r"}, false},
	{"foo='ba#r'", env.Env{"foo": "ba#r"}, false},

	// parses # in quoted values with following spaces
	{`foo="ba#r"  `, env.Env{"foo": "ba#r"}, false},
	{`foo='ba#r'  `, env.Env{"foo": "ba#r"}, false},

	// supports carriage return
	{"FOO=bar\rbaz=fbb", env.Env{"FOO": "bar", "baz": "fbb"}, false},

	// supports carriage return combine with new line
	{"FOO=bar\r\nbaz=fbb", env.Env{"FOO": "bar", "baz": "fbb"}, false},

	// expands carriage return in quoted strings
	{`FOO="bar\rbaz"`, env.Env{"FOO": "bar\rbaz"}, false},

	// escape $ properly when no alphabets/numbers/_  are followed by it
	{`FOO="bar\\$ \\$\\$"`, env.Env{"FOO": "bar$ $$"}, false},

	// ignore $ when it is not escaped and no variable is followed by it
	{`FOO="bar $ "`, env.Env{"FOO": "bar $ "}, false},

	// parses unquoted values with spaces after separator
	{`FOO= bar`, env.Env{"FOO": "bar"}, false},

	// allows # in quoted value with spaces after separator
	{`foo= "bar#baz" # comment`, env.Env{"foo": "bar#baz"}, false},

	// allows = in double quoted values with newlines (typically base64 padding)
	{`foo="---\na==\n---"`, env.Env{"foo": "---\na==\n---"}, false},
}

var errorFormats = []struct {
	in  string
	out env.Env
	err string
}{
	// allows export line if you want to do it that way and checks for unset variables
	{"OPTION_A=2\nexport OH_NO_NOT_SET", env.Env{"OPTION_A": "2"}, "line `export OH_NO_NOT_SET` has an unset variable"},

	// throws an error if line format is incorrect
	{`lol$wut`, env.Env{}, "line `lol$wut` doesn't match format"},
}

var fixtures = []struct {
	filename string
	results  env.Env
}{
	{
		"fixtures/exported.env",
		env.Env{
			"OPTION_A": "2",
			"OPTION_B": `\n`,
		},
	},
	{
		"fixtures/plain.env",
		env.Env{
			"OPTION_A": "1",
			"OPTION_B": "2",
			"OPTION_C": "3",
			"OPTION_D": "4",
			"OPTION_E": "5",
		},
	},
	{
		"fixtures/quoted.env",
		env.Env{
			"OPTION_A": "1",
			"OPTION_B": "2",
			"OPTION_C": "",
			"OPTION_D": `\n`,
			"OPTION_E": "1",
			"OPTION_F": "2",
			"OPTION_G": "",
			"OPTION_H": "\n",
		},
	},
	{
		"fixtures/yaml.env",
		env.Env{
			"OPTION_A": "1",
			"OPTION_B": "2",
			"OPTION_C": "",
			"OPTION_D": `\n`,
		},
	},
}

func TestParse(t *testing.T) {
	for _, tt := range formats {
		if tt.preset {
			os.Setenv("FOO", "test")
		}

		exp := env.Parse(strings.NewReader(tt.in))
		assert.Equal(t, tt.out, exp)
		os.Clearenv()
	}
}

func TestStrictParse(t *testing.T) {
	for _, tt := range errorFormats {
		env, err := env.StrictParse(strings.NewReader(tt.in))
		assert.Equal(t, tt.err, err.Error())
		assert.Equal(t, tt.out, env)
	}
}

func TestLoad(t *testing.T) {
	for _, tt := range fixtures {
		err := env.Load(tt.filename)
		assert.Nil(t, err)

		for key, val := range tt.results {
			assert.Equal(t, val, os.Getenv(key))
		}

		os.Clearenv()
	}
}

func TestLoad_default(t *testing.T) {
	k := "HELLO"
	v := "world"

	err := env.Load()
	assert.Nil(t, err)
	assert.Equal(t, v, os.Getenv(k))
	os.Clearenv()
}

func TestLoad_overriding(t *testing.T) {
	k := "HELLO"
	v := "universe"

	os.Setenv(k, v)
	err := env.Load()
	assert.Nil(t, err)
	assert.Equal(t, v, os.Getenv(k))
	os.Clearenv()
}

func TestLoad_overrideVars(t *testing.T) {
	os.Setenv("A", "fromEnv")
	err := env.Load("fixtures/vars.env")
	assert.Nil(t, err)
	assert.Equal(t, "fromEnv", os.Getenv("B"))
	os.Clearenv()
}

func TestLoad_overrideVars2(t *testing.T) {
	os.Setenv("C", "fromEnv")
	err := env.Load("fixtures/vars.env")
	assert.Nil(t, err)
	assert.Equal(t, "fromEnv", os.Getenv("D"))
	os.Clearenv()
}

func TestLoad_Env(t *testing.T) {
	err := env.Load(".env.invalid")
	assert.NotNil(t, err)
}

func TestLoad_nonExist(t *testing.T) {
	file := ".env.not.exist"

	err := env.Load(file)
	if err == nil {
		t.Errorf("env.Load(`%s`) => error: `no such file or directory` != nil", file)
	}
}

func TestLoad_unicodeBOMFixture(t *testing.T) {
	file := "fixtures/bom.env"

	f, err := os.Open(file)
	assert.Nil(t, err)

	scanner := bufio.NewScanner(f)

	i := 1
	bom := string([]byte{239, 187, 191})

	for scanner.Scan() {
		if i == 1 {
			line := scanner.Text()
			assert.True(t, strings.HasPrefix(line, bom))
		}
	}
}

func TestLoad_unicodeBOM(t *testing.T) {
	file := "fixtures/bom.env"

	err := env.Load(file)
	assert.Nil(t, err)
	assert.Equal(t, "UTF-8", os.Getenv("BOM"))
	os.Clearenv()
}

func TestMust_Load(t *testing.T) {
	for _, tt := range fixtures {
		assert.NotPanics(t, func() {
			env.Must(env.Load, tt.filename)

			for key, val := range tt.results {
				assert.Equal(t, val, os.Getenv(key))
			}

			os.Clearenv()
		}, "Caling env.Must with env.Load should NOT panic")
	}
}

func TestMust_Load_default(t *testing.T) {
	assert.NotPanics(t, func() {
		env.Must(env.Load)

		tkey := "HELLO"
		val := "world"

		assert.Equal(t, val, os.Getenv(tkey))
		os.Clearenv()
	}, "Caling env.Must with env.Load without arguments should NOT panic")
}

func TestMust_Load_nonExist(t *testing.T) {
	assert.Panics(t, func() { env.Must(env.Load, ".env.not.exist") }, "Caling env.Must with env.Load and non exist file SHOULD panic")
}

func TestOverLoad_overriding(t *testing.T) {
	k := "HELLO"
	v := "universe"

	os.Setenv(k, v)
	err := env.OverLoad()
	assert.Nil(t, err)
	assert.Equal(t, "world", os.Getenv(k))
	os.Clearenv()
}

func TestMustOverLoad_nonExist(t *testing.T) {
	assert.Panics(t, func() { env.Must(env.OverLoad, ".env.not.exist") }, "Caling env.Must with Overenv.Load and non exist file SHOULD panic")
}

func TestApply(t *testing.T) {
	os.Setenv("HELLO", "world")
	r := strings.NewReader("HELLO=universe")
	err := env.Apply(r)
	assert.Nil(t, err)
	assert.Equal(t, "world", os.Getenv("HELLO"))
	os.Clearenv()
}

func TestOverApply(t *testing.T) {
	os.Setenv("HELLO", "world")
	r := strings.NewReader("HELLO=universe")
	err := env.OverApply(r)
	assert.Nil(t, err)
	assert.Equal(t, "universe", os.Getenv("HELLO"))
	os.Clearenv()
}
