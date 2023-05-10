package project

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/yaml"
	"github.com/stretchr/testify/assert"
)

func TestNameFromURI(t *testing.T) {
	tests := []struct {
		uri  string
		want string
		err  error
	}{
		{uri: "", want: ""},
		{uri: ".", want: "project"},
		{uri: "foo", want: "foo"},
		{uri: "foo bar", want: "foo_bar"},
		{uri: "/foo/bar", want: "bar"},
		{uri: "file:///foo/bar", want: "bar"},
		{uri: "file://foo/bar", want: "bar"},
		//TODO
		//{uri: "file://.", want: "project"},
	}
	for _, tt := range tests {
		got, err := NameFromURI(tt.uri)
		if !errors.Is(err, tt.err) {
			t.Errorf("NameFromURI(%q) error = %v, want %v", tt.uri, err, tt.err)
		}
		if got != tt.want {
			t.Errorf("NameFromURI(%q) = %q, want %q", tt.uri, got, tt.want)
		}
	}
}

func TestEncodingFromURI(t *testing.T) {
	tests := []struct {
		uri  string
		want string
		err  error
	}{
		{uri: "", want: "per"},
		{uri: "s1ap", want: "per"},
		{uri: "bananap", want: "per"},
		{uri: "EUTRA_RRC", want: "uper"},
	}
	for _, tt := range tests {
		got, err := EncodingFromURI(tt.uri)
		if !errors.Is(err, tt.err) {
			t.Errorf("EncodingFromURI(%q) error = %v, want %v", tt.uri, err, tt.err)
		}
		if got != tt.want {
			t.Errorf("EncodingFromURI(%q) = %q, want %q", tt.uri, got, tt.want)
		}
	}

}

func TestImportTasks(t *testing.T) {
	tests := []struct {
		path   string
		result []string
		err    error
	}{
		{path: "./testdata/ImportTasks/invalid/notexist", err: os.ErrNotExist},
		{path: "./testdata/ImportTasks/invalid/file.ttcn3", err: syscall.ENOTDIR},
		{path: "./testdata/ImportTasks/invalid/dirs", err: ErrNoSources},
		{path: "./testdata/ImportTasks/other", err: ErrNoSources},
		{path: "./testdata/ImportTasks/lib", result: []string{"testdata/ImportTasks/lib/a.ttcn3", "testdata/ImportTasks/lib/b.ttcn3"}},
	}

	for _, tt := range tests {
		result, err := ImportTasks(&Config{}, tt.path)
		if !errors.Is(err, tt.err) {
			t.Errorf("%v: %v, want %v", tt.path, err, tt.err)
		}
		if len(tt.result) == 0 {
			continue
		}
		if len(result) != 1 {
			t.Errorf("Unexpected result: %v", result)
			continue
		}

		actual := result[0].Inputs()
		if !equal(actual, tt.result) {
			t.Errorf("%v: %v, want %v", tt.path, actual, tt.result)
		}
	}

}

func TestAutomaticEnv(t *testing.T) {
	t.Run("NTT_NAME", func(t *testing.T) {
		os.Unsetenv("NTT_NAME")
		c := &Config{}
		c.Name = "test"
		assert.Nil(t, AutomaticEnv()(c))
		assert.Equal(t, "test", c.Name)
	})

	t.Run("NTT_NAME", func(t *testing.T) {
		os.Setenv("NTT_NAME", "ntt")
		defer os.Unsetenv("NTT_NAME")
		c := &Config{}
		c.Name = "test"
		assert.Nil(t, AutomaticEnv()(c))
		assert.Equal(t, "ntt", c.Name)
	})

	t.Run("K3_NAME", func(t *testing.T) {
		os.Setenv("K3_NAME", "k3")
		defer os.Unsetenv("K3_NAME")
		c := &Config{}
		assert.Nil(t, AutomaticEnv()(c))
		assert.Equal(t, "k3", c.Name)
	})

	t.Run("NTT_NAME/K3_NAME", func(t *testing.T) {
		os.Setenv("K3_NAME", "k3")
		os.Setenv("NTT_NAME", "ntt")
		defer os.Unsetenv("K3_NAME")
		defer os.Unsetenv("NTT_NAME")
		c := &Config{}
		assert.Nil(t, AutomaticEnv()(c))
		assert.Equal(t, "ntt", c.Name)
	})

	t.Run("NTT_TIMEOUT", func(t *testing.T) {
		os.Setenv("NTT_TIMEOUT", "2.3")
		defer os.Unsetenv("NTT_TIMEOUT")
		c := &Config{}
		assert.Nil(t, AutomaticEnv()(c))
		assert.Equal(t, time.Duration(2.3*float64(time.Second)), c.Timeout.Duration)
	})

	t.Run("NTT_TIMEOUT", func(t *testing.T) {
		os.Setenv("NTT_TIMEOUT", "2.3s")
		defer os.Unsetenv("NTT_TIMEOUT")
		c := &Config{}
		assert.NotNil(t, AutomaticEnv()(c))
		assert.Equal(t, time.Duration(0), c.Timeout.Duration)
	})

	t.Run("PathListSeparator", func(t *testing.T) {
		srcs := []string{"a", "b", "c"}
		os.Setenv("NTT_SOURCES", strings.Join(srcs, string(os.PathListSeparator)))
		defer os.Unsetenv("NTT_SOURCES")
		c := &Config{}
		assert.Nil(t, AutomaticEnv()(c))
		assert.Equal(t, srcs, c.Sources)
	})

	t.Run("PathListSeparator", func(t *testing.T) {
		s := "a b c"
		os.Setenv("NTT_IMPORTS", s)
		defer os.Unsetenv("NTT_IMPORTS")
		c := &Config{}
		assert.Nil(t, AutomaticEnv()(c))
		assert.Equal(t, []string{s}, c.Imports)
	})

	t.Run("File handling", func(t *testing.T) {
		path := "test://foobar/test.yaml"
		os.Setenv("NTT_PARAMETERS_FILE", path)
		defer os.Unsetenv("NTT_PARAMETERS_FILE")
		c := &Config{}
		assert.Nil(t, AutomaticEnv()(c))
		assert.Equal(t, path, c.ParametersFile)
	})

	t.Run("File handling", func(t *testing.T) {
		path := "project_test.go"
		os.Setenv("NTT_HOOKS_FILE", path)
		defer os.Unsetenv("NTT_HOOKS_FILE")
		c := &Config{}
		assert.Nil(t, AutomaticEnv()(c))
		assert.Equal(t, path, c.HooksFile)
	})
}

func TestWithDefaults(t *testing.T) {
	t.Run("cwd", func(t *testing.T) {
		c := &Config{}
		assert.Nil(t, WithDefaults()(c))
		assert.Equal(t, "project", filepath.Base(c.Root))
		assert.Equal(t, "project", filepath.Base(c.SourceDir))
	})

	t.Run("source_dir", func(t *testing.T) {
		c := &Config{}
		c.SourceDir = "source_dir"
		assert.Nil(t, WithDefaults()(c))
		assert.Equal(t, "source_dir", c.Root)
		assert.Equal(t, "source_dir", c.SourceDir)
	})

	t.Run("root", func(t *testing.T) {
		c := &Config{}
		c.Root = "root"
		assert.Nil(t, WithDefaults()(c))
		assert.Equal(t, "root", c.Root)
		assert.Equal(t, "root", c.SourceDir)
	})

	t.Run("root", func(t *testing.T) {
		c := &Config{}
		c.Root = "root"
		c.SourceDir = "source_dir"
		assert.Nil(t, WithDefaults()(c))
		assert.Equal(t, "root", c.Root)
		assert.Equal(t, "source_dir", c.SourceDir)
	})

	t.Run("sources", func(t *testing.T) {
		c := &Config{}
		c.Sources = []string{"a/a", "b/b", "c/c"}
		assert.Nil(t, WithDefaults()(c))
		assert.Equal(t, "a", c.Root)
		assert.Equal(t, "a", c.SourceDir)
	})

	t.Run("sources", func(t *testing.T) {
		c := &Config{}
		c.Sources = []string{".", "b/b", "c/c"}
		assert.Nil(t, WithDefaults()(c))
		assert.Equal(t, ".", c.Root)
		assert.Equal(t, ".", c.SourceDir)

		// Note: In previous implementation the default-name was always
		// the base-name of the root folder.
		// But when we execute many single file tests, this is not what we want.
		// This test ensures that the default-name is the base-name of the first source file.
		assert.Equal(t, "project", c.Name)
	})

}

func TestWithManifest(t *testing.T) {
	manifest := func(path string, s string) (*Config, error) {
		fs.SetContent(path, []byte(strings.ReplaceAll(s, "\t", "        ")))
		c := &Config{}
		return c, WithManifest(path)(c)
	}
	t.Run("empty", func(t *testing.T) {
		c, err := manifest("package.yml", "")
		assert.Nil(t, err)
		assert.Equal(t, "package.yml", c.ManifestFile)
	})
	t.Run("paths", func(t *testing.T) {
		c, err := manifest("package.yml", "hooks_file: file")
		assert.Nil(t, err)
		assert.Equal(t, "file", c.HooksFile)
	})
	t.Run("paths", func(t *testing.T) {
		c, err := manifest("foo/bar/package.yml", "hooks_file: file")
		assert.Nil(t, err)
		assert.Equal(t, "foo/bar/file", c.HooksFile)
	})
	t.Run("paths", func(t *testing.T) {
		c, err := manifest("foo/bar/package.yml", "hooks_file: $VAR")
		assert.Nil(t, err)
		assert.Equal(t, "${VAR}", c.HooksFile)
	})
	t.Run("paths", func(t *testing.T) {
		c, err := manifest("foo/bar/package.yml", "hooks_file: /file")
		assert.Nil(t, err)
		assert.Equal(t, "/file", c.HooksFile)
	})
	t.Run("paths", func(t *testing.T) {
		c, err := manifest("foo/bar/package.yml", "hooks_file: https://file.txt")
		assert.Nil(t, err)
		assert.Equal(t, "https://file.txt", c.HooksFile)
	})
	t.Run("variables", func(t *testing.T) {
		c, err := manifest("foo/bar/package.yml", `
            variables:
              VAR: file
            hooks_file: $VAR`)
		assert.Nil(t, err)
		assert.Equal(t, "foo/bar/file", c.HooksFile)
	})
}

func TestParametersMergeRules(t *testing.T) {
	unmarshal := func(s string) Parameters {
		var p Parameters
		if err := yaml.Unmarshal([]byte(s), &p); err != nil {
			t.Fatal(err)
		}
		return p
	}

	a := unmarshal(`
                timeout: 0
                presets:
                   "A":
                      timeout: 1
                   "B":
                      timeout: 2
                execute:
                  - test: "TC1"
                    timeout: 3`)

	b := unmarshal(`
                timeout: 4
                presets:
                   "A":
                      test: "*"
                   "B":
                      timeout: 5
                   "C":
                      timeout: 6
                execute:
                  - test: "TC1"
                    timeout: 7`)

	actual := MergeParameters(a, b)
	expected := unmarshal(`
                timeout: 4
                presets:
                  "A":
                    test: "*"
                    timeout: 1
                  "B":
                    timeout: 5
                  "C":
                    timeout: 6
                execute:
                  - test: "TC1"
                    timeout: 3
                  - test: "TC1"
                    timeout: 7`)

	assert.Equal(t, expected.TestConfig, actual.TestConfig)
	assert.Equal(t, expected.Presets["A"], actual.Presets["A"])
	assert.Equal(t, expected.Presets["B"], actual.Presets["B"])
	assert.Equal(t, expected.Presets["C"], actual.Presets["C"])
	assert.Equal(t, expected.Execute, actual.Execute)
}

func TestRules(t *testing.T) {
	tests := []struct {
		Presets, Except, Only []string
		Want                  bool
	}{
		{Presets: nil, Except: nil, Only: nil, Want: true},
		{Presets: []string{"A"}, Except: nil, Only: nil, Want: true},
		{Presets: []string{"A", "B"}, Except: nil, Only: nil, Want: true},

		{Presets: nil, Except: []string{"A"}, Only: nil, Want: true},
		{Presets: []string{"A"}, Except: []string{"A"}, Only: nil, Want: false},
		{Presets: []string{"A"}, Except: []string{"B"}, Only: nil, Want: true},
		{Presets: []string{"A"}, Except: []string{"A", "B"}, Only: nil, Want: false},
		{Presets: []string{"A", "B"}, Except: []string{"A"}, Only: nil, Want: false},

		{Presets: nil, Except: nil, Only: []string{"A"}, Want: false},
		{Presets: []string{"A"}, Except: nil, Only: []string{"A"}, Want: true},
		{Presets: []string{"A"}, Except: nil, Only: []string{"A", "B"}, Want: true},
		{Presets: []string{"A"}, Except: nil, Only: []string{"B"}, Want: false},
		{Presets: []string{"A", "B"}, Except: nil, Only: []string{"B"}, Want: true},

		{Presets: nil, Except: []string{"A"}, Only: []string{"A"}, Want: false},
		{Presets: []string{"A"}, Except: []string{"A"}, Only: []string{"A"}, Want: false},
		{Presets: []string{"A"}, Except: []string{"B"}, Only: []string{"A"}, Want: true},
		{Presets: []string{"A", "B"}, Except: []string{"B"}, Only: []string{"A"}, Want: false},
	}
	for _, tt := range tests {
		t.Run(t.Name(), func(t *testing.T) {
			rules := Rules{
				Except: &ExecuteCondition{tt.Except},
				Only:   &ExecuteCondition{tt.Only},
			}
			got := matchRules(rules, tt.Presets...)
			assert.Equal(t, tt.Want, got, fmt.Sprintf("Presets: %v, Except: %v, Only: %v", tt.Presets, tt.Except, tt.Only))
		})
	}
}

// equal returns true if a and b are equal string slices, order is ignored.
func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	m := make(map[string]int, len(a))
	for i := range a {
		m[a[i]]++
		m[b[i]]--
	}
	for _, v := range m {
		if v != 0 {
			return false
		}
	}
	return true
}
