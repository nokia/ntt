package loader

import (
	"fmt"
	"github.com/nokia/ntt/ttcn3/syntax"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	ParseOnly bool
	Files     []string
}

func (conf *Config) FromArgs(args []string) error {
	if len(args) == 0 {
		args = []string{"."}
	}

	if isTTCN3File(args[0]) {
		for _, arg := range args {
			if !isTTCN3File(arg) {
				return fmt.Errorf("named files be .ttcn3 files: %s", arg)
			}
		}
		conf.FromFiles(args)
	} else {
		for _, arg := range args {
			if err := conf.FromDir(arg); err != nil {
				return err
			}
		}
	}

	return nil
}

func (conf *Config) FromFiles(files []string) {
	conf.Files = append(conf.Files, files...)
}

func (conf *Config) FromDir(path string) error {
	hasFiles := false
	root := filepath.Clean(path)
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		switch {
		case info.IsDir():
		case isTTCN3File(path):
			hasFiles = true
			conf.Files = append(conf.Files, path)
		}

		return nil
	})

	if err == nil && !hasFiles {
		return fmt.Errorf("no ttcn3 files in: %s", root)
	}

	return err
}

func isTTCN3File(name string) bool {
	return strings.HasSuffix(name, ".ttcn3") || strings.HasSuffix(name, ".ttcn")
}

func (conf *Config) Load() (*Package, error) {
	pkg := new(Package)
	pkg.Fset = syntax.NewFileSet()
	m, err := parseFiles(pkg.Fset, conf.Files)
	if err != nil {
		return nil, err
	}

	pkg.Modules = m
	return pkg, nil
}

type Package struct {
	Fset    *syntax.FileSet
	Modules []*syntax.Module
}
