package suite

import (
	"path/filepath"

	"github.com/nokia/ntt/internal/fs"
)

// Package represents a directory associated with TTCN-3 source files.
type Package struct {
	dir     string
	Name    string
	Sources []string
	Imports []string
}

func NewPackage(dir string) *Package {
	pkg := &Package{}
	pkg.init(dir)
	return pkg
}

func (pkg *Package) init(dir string) {
	pkg.SetDir(dir)
	pkg.SetName(pkg.dir)
}

// Dir returns the directory the package is assiciated with.
func (pkg *Package) Dir() string {
	return pkg.dir
}

// SetDir sets associated directory for the package. The path is cleaned.
func (pkg *Package) SetDir(dir string) {
	pkg.dir = filepath.Clean(dir)
}

// SetName sets name for the package. If name is a path-name, SetName will use
// its base name.
func (pkg *Package) SetName(name string) {
	pkg.Name = fs.Basename(name)
}
