package config

import (
	"testing"
)

func TestFromArgs(t *testing.T) {
	tt := []struct {
		name string
		args []string
		conf *Config
		ok   bool
	}{
		{
			name: "dir_autodetect",
			ok:   true,
			args: []string{"testdata/a"},
			conf: &Config{
				Name: "a",
				Sources: []string{
					"testdata/a/a.ttcn3",
					"testdata/a/b.ttcn3",
					"testdata/a/c.ttcn3",
				},
			},
		},
		{
			name: "yaml_overwrite_name",
			ok:   true,
			args: []string{"testdata/b"},
			conf: &Config{
				Name: "a",
				Sources: []string{
					"testdata/b/a.ttcn3",
					"testdata/b/b.ttcn3",
					"testdata/b/c.ttcn3",
				},
			},
		},
		{
			name: "yaml",
			ok:   true,
			args: []string{"testdata/c"},
			conf: &Config{
				Name:    "c",
				Sources: []string{"testdata/c/a.ttcn3"},
				Imports: []string{"testdata/b"},
			},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			conf, err := FromArgs(tc.args)
			if tc.ok {
				if err != nil {
					t.Fatalf("unexpected error: %s", err.Error())
				}
				if tc.conf.Name != conf.Name {
					t.Errorf("expected name=%s, got name=%s", tc.conf.Name, conf.Name)
				}
				if !isEqual(tc.conf.Sources, conf.Sources) {
					t.Errorf("expected sources=%v, got sources=%v", tc.conf.Sources, conf.Sources)
				}
				if !isEqual(tc.conf.Imports, conf.Imports) {
					t.Errorf("expected imports=%v, got imports=%v", tc.conf.Imports, conf.Imports)
				}
				if tc.conf.ParametersFile != conf.ParametersFile {
					t.Errorf("expected parameters=%s, got parameters=%s", tc.conf.ParametersFile, conf.ParametersFile)
				}
				if tc.conf.TestHook != conf.TestHook {
					t.Errorf("expected hook=%s, got hook=%s", tc.conf.TestHook, conf.TestHook)
				}
				if tc.conf.Timeout != conf.Timeout {
					t.Errorf("expected timeout=%f, got timeout=%f", tc.conf.Timeout, conf.Timeout)
				}
			} else {
				if err == nil {
					t.Fatalf("unexpected pass")
				}
			}
		})
	}
}

func isEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
