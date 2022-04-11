package ntt

import (
	"fmt"
	"path/filepath"

	"github.com/jinzhu/copier"
	"github.com/nokia/ntt/internal/yaml"
	"github.com/nokia/ntt/k3"
	"github.com/nokia/ntt/project"
)

func (s *Suite) Config() (*project.Config, error) {
	var (
		c    project.Config
		gErr error
	)

	check := func(err error) {
		if err != nil && gErr == nil {
			gErr = err
		}
	}

	m, err := s.parseManifest()
	check(err)
	copier.Copy(&c, m)

	c.Name, err = s.Name()
	check(err)

	c.Sources, err = s.Sources()
	check(err)

	c.Imports, err = s.Imports()
	check(err)

	c.Variables, err = s.Variables()
	check(err)

	f, err := s.TestHook()
	check(err)
	if f != nil {
		c.HooksFile = f.Path()
	}

	f, err = s.ParametersFile()
	check(err)
	if f != nil {
		c.ParametersFile = f.Path()
		b, err := f.Bytes()
		check(err)
		if err == nil {
			if err2 := yaml.Unmarshal(b, &c.Manifest); err2 != nil {
				check(fmt.Errorf("Syntax error in file %s: %w", f.Path(), err2))
			}
		}
	}

	c.ParametersDir, err = s.ParametersDir()
	check(err)

	c.Timeout, err = s.Timeout()
	check(err)

	if path, err := filepath.Abs(c.SourceDir); err == nil {
		c.SourceDir = path
	}
	check(err)

	i, err := s.Id()
	check(err)
	if err == nil {
		c.K3.SessionID = fmt.Sprint(i)
	}

	c.Variables, err = s.Variables()

	c.K3.Compiler = k3.Compiler()
	c.K3.Runtime = k3.Runtime()
	c.K3.Builtins = k3.FindAuxiliaryDirectories()
	c.K3.OssInfo, _ = s.Getenv("OSSINFO")

	return &c, gErr
}
