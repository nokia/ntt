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

// ExpandVar expands variables references inside a string using environment
// variables first and then a provided variables map second. If a reference
// could not expanded, the variable reference will stay inside the string and
// an an error will be returned.
func ExpandVar(s string, vars map[string]string) (string, error) {
	return expandVar(s, vars, make(map[string]string))
}

func expandVar(s string, vars map[string]string, visited map[string]string) (string, error) {
	var errs error
	mapper := func(name string) string {
		s, err := getVar(name, vars, visited)
		if err != nil {
			errs = multierror.Append(errs, &NoSuchVariableError{Name: name})
		}
		return s
	}
	return os.Expand(s, mapper), errs
}

// GetVar returns a variable. Variable references in vars are expanded.
// Environment variables are not.
func GetVar(name string, vars map[string]string) (string, error) {
	return getVar(name, vars, make(map[string]string))
}

func getVar(name string, vars map[string]string, visited map[string]string) (string, error) {
	if v, ok := visited[name]; ok {
		return v, nil
	}
	visited[name] = ""

	if v, ok := env.LookupEnv(name); ok {
		visited[name] = v
		return v, nil
	}

	// We must not look for NTT_CACHE in variables sections of package.yml,
	// because this would create an endless loop.
	if name != "NTT_CACHE" && name != "K3_CACHE" {
		if v, ok := vars[name]; ok {
			v, err := expandVar(v, vars, visited)
			visited[name] = v
			return v, err
		}
	}

	if knownVars[name] {
		return "", nil
	}

	return "", &NoSuchVariableError{Name: name}
}

// Variables return all declared variables, in the form "key=value".
func Variables(vars map[string]string) ([]string, error) {
	var errs error

	allKeys := make(map[string]bool)

	for k := range vars {
		allKeys[k] = true
	}

	for k := range env.Parse() {
		allKeys[k] = true
	}

	ret := make([]string, 0, len(allKeys))
	for k := range allKeys {
		v, err := GetVar(k, vars)
		if err != nil {
			errs = multierror.Append(errs, err)
		}
		ret = append(ret, fmt.Sprintf("%s=%s", k, v))
	}
	return ret, nil
}

func Open(path string) (*Project, error) {
	p := Project{root: path}

	// Try reading the manifest
	file := filepath.Join(p.root, ManifestFile)
	if b, err := fs.Content(file); err == nil {
		log.Debugf("%s: update configuration using manifest %q\n", p.String(), file)
		return &p, yaml.UnmarshalStrict(b, &p.Manifest)
	}

	// Fall back to recursive scanning
	log.Debugf("%s: update configuration using available folders\n", p.String())
	return &p, p.findFilesRecursive()
}

// A Project provides meta information about a TTCN-3 test suite. Meta
// information like: Location of configuration and source files, dependency
// list, default values, ...
type Project struct {
	root string

	// Module handling (maps module names to paths)
	modulesMu sync.Mutex
	modules   map[string]string

	Manifest Manifest
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
		src, err := p.evalPath(src)
		if err != nil {
			errs = multierror.Append(errs, err)
		}

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
		dir, err := p.evalPath(dir)
		if err != nil {
			errs = multierror.Append(errs, err)
		}
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

func (p *Project) findFilesRecursive() error {
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
		if err == nil && fs.IsDir(path) {
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
		"../common",
		"../Common",
		"../library",
		"../../../../Common",
	}
	for _, dir := range commonDirs {
		path := filepath.Join(p.root, dir)

		if eval, err := filepath.EvalSymlinks(path); err == nil {
			path = eval
		}

		filepath.Walk(path, addImports)
	}
	return nil
}

// Environ returns a copy of strings representing the environment, in the form "key=value".
func (p *Project) Environ() ([]string, error) {
	return Variables(p.Manifest.Variables)
}

// Expand expands string v using Project.Getenv
func (p *Project) Expand(v string) (string, error) {
	return expandVar(v, p.Manifest.Variables, make(map[string]string))
}

func (p *Project) Getenv(v string) (string, error) {
	return getVar(v, p.Manifest.Variables, make(map[string]string))
}

func (p *Project) evalPath(path string) (string, error) {
	subst, err := p.Expand(path)
	if err == nil {
		path = subst
	}
	return fs.Real(p.Root(), path), err
}

type NoSuchVariableError struct {
	Name string
}

func (e *NoSuchVariableError) Error() string {
	return e.Name + ": variable not defined"
}

var knownVars = map[string]bool{
	"CXXFLAGS":            true,
	"K3CFLAGS":            true,
	"K3RFLAGS":            true,
	"LDFLAGS":             true,
	"LD_LIBRARY_PATH":     true,
	"PATH":                true,
	"NTT_DATADIR":         true,
	"NTT_IMPORTS":         true,
	"NTT_NAME":            true,
	"NTT_PARAMETERS_DIR":  true,
	"NTT_PARAMETERS_FILE": true,
	"NTT_SOURCES":         true,
	"NTT_SOURCE_DIR":      true,
	"NTT_TEST_HOOK":       true,
	"NTT_TIMEOUT":         true,
	"NTT_VARIABLES":       true,
}
