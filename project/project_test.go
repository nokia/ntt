package project_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/project"
	"github.com/stretchr/testify/assert"
)

func TestAutomaticEnv(t *testing.T) {
	t.Run("NTT_NAME", func(t *testing.T) {
		os.Unsetenv("NTT_NAME")
		c := &project.Config{}
		c.Name = "test"
		assert.Nil(t, project.AutomaticEnv()(c))
		assert.Equal(t, "test", c.Name)
	})

	t.Run("NTT_NAME", func(t *testing.T) {
		os.Setenv("NTT_NAME", "ntt")
		defer os.Unsetenv("NTT_NAME")
		c := &project.Config{}
		c.Name = "test"
		assert.Nil(t, project.AutomaticEnv()(c))
		assert.Equal(t, "ntt", c.Name)
	})

	t.Run("K3_NAME", func(t *testing.T) {
		os.Setenv("K3_NAME", "k3")
		defer os.Unsetenv("K3_NAME")
		c := &project.Config{}
		assert.Nil(t, project.AutomaticEnv()(c))
		assert.Equal(t, "k3", c.Name)
	})

	t.Run("NTT_NAME/K3_NAME", func(t *testing.T) {
		os.Setenv("K3_NAME", "k3")
		os.Setenv("NTT_NAME", "ntt")
		defer os.Unsetenv("K3_NAME")
		defer os.Unsetenv("NTT_NAME")
		c := &project.Config{}
		assert.Nil(t, project.AutomaticEnv()(c))
		assert.Equal(t, "ntt", c.Name)
	})

	t.Run("NTT_TIMEOUT", func(t *testing.T) {
		os.Setenv("NTT_TIMEOUT", "2.3")
		defer os.Unsetenv("NTT_TIMEOUT")
		c := &project.Config{}
		assert.Nil(t, project.AutomaticEnv()(c))
		assert.Equal(t, time.Duration(2.3*float64(time.Second)), c.Timeout.Duration)
	})

	t.Run("NTT_TIMEOUT", func(t *testing.T) {
		os.Setenv("NTT_TIMEOUT", "2.3s")
		defer os.Unsetenv("NTT_TIMEOUT")
		c := &project.Config{}
		assert.NotNil(t, project.AutomaticEnv()(c))
		assert.Equal(t, time.Duration(0), c.Timeout.Duration)
	})

	t.Run("PathListSeparator", func(t *testing.T) {
		srcs := []string{"a", "b", "c"}
		os.Setenv("NTT_SOURCES", strings.Join(srcs, string(os.PathListSeparator)))
		defer os.Unsetenv("NTT_SOURCES")
		c := &project.Config{}
		assert.Nil(t, project.AutomaticEnv()(c))
		assert.Equal(t, srcs, c.Sources)
	})

	t.Run("PathListSeparator", func(t *testing.T) {
		s := "a b c"
		os.Setenv("NTT_IMPORTS", s)
		defer os.Unsetenv("NTT_IMPORTS")
		c := &project.Config{}
		assert.Nil(t, project.AutomaticEnv()(c))
		assert.Equal(t, []string{s}, c.Imports)
	})

	t.Run("File handling", func(t *testing.T) {
		path := "test://foobar/test.yaml"
		os.Setenv("NTT_PARAMETERS_FILE", path)
		defer os.Unsetenv("NTT_PARAMETERS_FILE")
		c := &project.Config{}
		assert.Nil(t, project.AutomaticEnv()(c))
		assert.Equal(t, path, c.ParametersFile)
	})

	t.Run("File handling", func(t *testing.T) {
		path := "project_test.go"
		os.Setenv("NTT_HOOKS_FILE", path)
		defer os.Unsetenv("NTT_HOOKS_FILE")
		c := &project.Config{}
		assert.Nil(t, project.AutomaticEnv()(c))
		assert.Equal(t, path, c.HooksFile)
	})
}

func TestWithDefaults(t *testing.T) {
	t.Run("cwd", func(t *testing.T) {
		c := &project.Config{}
		assert.Nil(t, project.WithDefaults()(c))
		assert.Equal(t, "project", filepath.Base(c.Root))
		assert.Equal(t, "project", filepath.Base(c.SourceDir))
	})

	t.Run("source_dir", func(t *testing.T) {
		c := &project.Config{}
		c.SourceDir = "source_dir"
		assert.Nil(t, project.WithDefaults()(c))
		assert.Equal(t, "source_dir", c.Root)
		assert.Equal(t, "source_dir", c.SourceDir)
	})

	t.Run("root", func(t *testing.T) {
		c := &project.Config{}
		c.Root = "root"
		assert.Nil(t, project.WithDefaults()(c))
		assert.Equal(t, "root", c.Root)
		assert.Equal(t, "root", c.SourceDir)
	})

	t.Run("root", func(t *testing.T) {
		c := &project.Config{}
		c.Root = "root"
		c.SourceDir = "source_dir"
		assert.Nil(t, project.WithDefaults()(c))
		assert.Equal(t, "root", c.Root)
		assert.Equal(t, "source_dir", c.SourceDir)
	})

	t.Run("sources", func(t *testing.T) {
		c := &project.Config{}
		c.Sources = []string{"a/a", "b/b", "c/c"}
		assert.Nil(t, project.WithDefaults()(c))
		assert.Equal(t, "a", c.Root)
		assert.Equal(t, "a", c.SourceDir)
	})

	t.Run("sources", func(t *testing.T) {
		c := &project.Config{}
		c.Sources = []string{".", "b/b", "c/c"}
		assert.Nil(t, project.WithDefaults()(c))
		assert.Equal(t, ".", c.Root)
		assert.Equal(t, ".", c.SourceDir)
	})

}

func TestWithManifest(t *testing.T) {
	manifest := func(path string, s string) (*project.Config, error) {
		fs.SetContent(path, []byte(strings.ReplaceAll(s, "\t", "        ")))
		c := &project.Config{}
		return c, project.WithManifest(path)(c)
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
