// Package project collects information about test suite organisation by
// implementing various heuristics.
package project

import (
	"crypto/sha1"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/hashicorp/go-multierror"
	"github.com/nokia/ntt/internal/env"
	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/k3"
	"github.com/nokia/ntt/project/manifest"
	"gopkg.in/yaml.v2"
)

// Interface describes a TTCN-3 project.
type Interface interface {
	// Root is the test suite root folder. It is usually the folder where the manifest is.
	Root() string

	// Sources returns a slice of files and directories containing TTCN-3 source files.
	Sources() ([]string, error)

	// Imports returns a slice of additional directories required to build a test executable.
	// Codecs, adapters and libraries are specified by Imports, typically.
	Imports() ([]string, error)
}

// Files returns all .ttcn3 available. It will not return generated .ttcn3 files.
// On error Files will return an error.
func Files(p Interface) ([]string, error) {
	var errs *multierror.Error
	files, err := p.Sources()
	if err != nil {
		errs = multierror.Append(errs, err)
	}

	dirs, err := p.Imports()
	if err != nil {
		errs = multierror.Append(errs, err)
	}

	for _, dir := range dirs {
		f := fs.FindTTCN3Files(dir)
		files = append(files, f...)
	}
	if errs.ErrorOrNil() == nil {
		return files, nil
	}
	return files, multierror.Flatten(errs)
}

// FindAllFiles returns all .ttcn3 files including auxiliary files from
// k3 installation
func FindAllFiles(p Interface) []string {
	files, _ := Files(p)
	// Use auxilliaryFiles from K3 to locate file
	for _, dir := range k3.FindAuxiliaryDirectories() {
		for _, file := range fs.FindTTCN3Files(dir) {
			files = append(files, file)
		}
	}
	return files
}

// ContainsFile returns true, when path is managed by Interface.
func ContainsFile(p Interface, path string) bool {

	// The same file may be referenced by URI or by path. To normalize it
	// we convert everything into URIs.
	uri := fs.URI(path)

	files, _ := Files(p)
	for _, file := range files {
		if fs.URI(file) == uri {
			return true
		}
	}
	return false
}

// Fingerprint calculates a sum to identify a test suite based on its modules.
func Fingerprint(p Interface) string {
	var inputs []string
	files, _ := Files(p)
	for _, file := range files {
		inputs = append(inputs, fs.Stem(file))
	}
	return fmt.Sprintf("project_%x", sha1.Sum([]byte(fmt.Sprint(inputs))))
}

func Open(path string) (*Project, error) {
	p := Project{root: path}

	if file := filepath.Join(p.root, manifest.Name); fs.IsRegular(file) {
		log.Debugf("%s: update configuration using manifest %q\n", p.String(), file)
		return &p, p.readManifest(file)
	}

	log.Debugf("%s: update configuration using available folders\n", p.String())
	return &p, p.readFilesystem()
}

// A Project implements the behaviour expected from ntt ttcn3 suites.
type Project struct {
	root string

	// Module handling (maps module names to paths)
	modulesMu sync.Mutex
	modules   map[string]string

	Manifest manifest.Config
}

// String returns a simple string representation
func (p *Project) String() string {
	return p.root
}

// Root directory is ttcn3 suite.
func (p *Project) Root() string {
	return p.root
}

// Sources returns all TTCN-3 source files.
func (p *Project) Sources() ([]string, error) {
	var errs error

	if env := env.Getenv("NTT_SOURCES"); env != "" {
		return strings.Fields(env), nil
	}

	var srcs []string
	for _, src := range p.Manifest.Sources {
		src := fs.Real(p.Root(), p.expand(src))

		info, err := os.Stat(src)
		switch {
		case err != nil:
			errs = multierror.Append(errs, err)
			srcs = append(srcs, src)

		case info.IsDir():
			files := fs.FindTTCN3Files(src)
			if len(files) == 0 {
				errs = multierror.Append(errs, fmt.Errorf("Could not find any ttcn3 source files in directory %q", src))
			}
			srcs = append(srcs, files...)

		case info.Mode().IsRegular() && fs.HasTTCN3Extension(src):
			srcs = append(srcs, src)

		default:
			errs = multierror.Append(errs, fmt.Errorf("Cannot handle %q. Expecting directory or ttcn3 source file", src))
			srcs = append(srcs, src)
		}

	}
	return srcs, errs
}

func (p *Project) Imports() ([]string, error) {
	var errs error

	if env := env.Getenv("NTT_IMPORTS"); env != "" {
		return strings.Fields(env), nil
	}

	var imports []string
	for _, dir := range p.Manifest.Imports {
		dir := fs.Real(p.Root(), p.expand(dir))
		if info, err := os.Stat(dir); err != nil || !info.IsDir() {
			errs = multierror.Append(errs, fmt.Errorf("%q must be a directory", dir))
		}
		imports = append(imports, dir)
	}
	return imports, errs
}

// FindModule tries to find a .ttcn3 based on its module name.
func (p *Project) FindModule(name string) (string, error) {

	p.modulesMu.Lock()
	defer p.modulesMu.Unlock()

	if p.modules == nil {
		p.modules = make(map[string]string)
		for _, file := range FindAllFiles(p) {
			name := fs.Stem(file)
			p.modules[name] = file
		}
	}
	if file, ok := p.modules[name]; ok {
		return file, nil
	}

	// Use NTT_CACHE to locate file
	if f := fs.Open(name + ".ttcn3").Path(); fs.IsRegular(f) {
		p.modules[name] = f
		return f, nil
	}

	return "", fmt.Errorf("No such module %q", name)
}

func (p *Project) expand(s string) string {
	mapper := func(name string) string {
		if s, ok := env.LookupEnv(name); ok {
			return s
		}
		if s, ok := p.Manifest.Variables[name]; ok {
			return s
		}
		return fmt.Sprintf("${UNKNOWN:%s}", name)
	}

	return os.Expand(s, mapper)
}

func (p *Project) readManifest(file string) error {
	b, err := fs.Open(file).Bytes()
	if err != nil {
		return err
	}
	return yaml.UnmarshalStrict(b, &p.Manifest)
}

func (p *Project) readFilesystem() error {
	addSources := func(path string, info os.FileInfo, err error) error {
		if err == nil && info.IsDir() {
			files, _ := filepath.Glob(filepath.Join(path, "*.ttcn*"))
			if len(files) > 0 {
				log.Debugf("adding sources: %q\n", path)
				p.Manifest.Sources = append(p.Manifest.Sources, fs.Rel(p.Root(), files...)...)
			}
		}
		return nil
	}
	filepath.Walk(p.root, addSources)

	addImports := func(path string, info os.FileInfo, err error) error {
		if err == nil && info.IsDir() {
			files, _ := filepath.Glob(filepath.Join(path, "*.ttcn*"))
			if len(files) > 0 {
				log.Debugf("adding import: %q\n", path)
				p.Manifest.Imports = append(p.Manifest.Imports, path)
			}
		}
		return nil
	}
	commonDirs := []string{
		"../../../sct",
		"../../../../sct",
		"../../../../Common",
	}
	for _, dir := range commonDirs {
		if path := filepath.Join(p.root, dir); fs.IsDir(path) {
			filepath.Walk(path, addImports)
		}
	}
	return nil
}
