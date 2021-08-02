// Package project collects information about test suite organisation by
// implementing various heuristics.
package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/hashicorp/go-multierror"
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

// Files returns all available .ttcn3 files. It will not return intermediate .ttcn3 files.
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

func Open(path string) (*Project, error) {
	p := Project{
		root: path,
	}

	return &p, p.Update()
}

type Project struct {
	root string

	// Module handling (maps module names to paths)
	modulesMu sync.Mutex
	modules   map[string]string

	Manifest manifest.Config
}

func (p *Project) Update() error {

	p.modules = nil // Clear module cache

	if file := filepath.Join(p.root, manifest.Name); isRegular(file) {
		log.Debugf("%s: update configuration using manifest %q\n", p.String(), file)
		return p.readManifest(file)
	}
	log.Debugf("%s: update configuration using available folders\n", p.String())
	p.readFilesystem()
	return nil
}

func (p *Project) String() string {
	return p.root
}

// Root directory is ttcn3 suite.
func (p *Project) Root() string {
	return p.root
}

func (p *Project) Name() string {
	if env := getenv("NTT_NAME"); env != "" {
		return env
	}
	if p.Manifest.Name != "" {
		return p.expand(p.Manifest.Name)
	}
	if p.root != "" {
		return slugify(filepath.Base(p.root))
	}
	if files, _ := Files(p); len(files) > 0 {
		file := files[0]
		if abs, err := filepath.Abs(file); err == nil {
			return slugify(strings.TrimSuffix(filepath.Base(abs), filepath.Ext(abs)))
		}
		return slugify(filepath.Base(file))
	}
	return "_"
}

func (p *Project) TestHook() string {
	if env := getenv("NTT_TEST_HOOK"); env != "" {
		return env
	}
	if p.Manifest.TestHook != "" {
		return fix(p.Root(), p.expand(p.Manifest.TestHook))
	}
	if hook := fix(p.Root(), p.Name()+".control"); isRegular(hook) {
		return hook
	}
	if files, _ := Files(p); len(files) > 0 {
		dir := filepath.Dir(files[0])
		if hook := filepath.Join(dir, p.Name()+".control"); isRegular(hook) {
			return hook
		}
	}
	return ""
}

func (p *Project) ParametersFile() string {
	if env := getenv("NTT_PARAMETERS_FILE"); env != "" {
		return fix(p.ParametersDir(), env)
	}
	if p.Manifest.ParametersFile != "" {
		return fix(p.ParametersDir(), p.expand(p.Manifest.ParametersFile))
	}
	if file := fix(p.ParametersDir(), p.Name()+".parameters"); isRegular(file) {
		return file
	}
	return ""
}

func (p *Project) ParametersDir() string {
	if env := getenv("NTT_PARAMETERS_DIR"); env != "" {
		return env
	}
	if p.Manifest.ParametersDir != "" {
		return fix(p.Root(), p.expand(p.Manifest.ParametersDir))
	}
	if p.Root() != "" {
		return p.Root()
	}
	if files, _ := Files(p); len(files) > 0 {
		if file, err := filepath.Abs(files[0]); err != nil {
			return filepath.Dir(file)
		}
	}
	return ""
}

// Sources returns all TTCN-3 source files.
func (p *Project) Sources() ([]string, error) {
	var errs error

	if env := getenv("NTT_SOURCES"); env != "" {
		return strings.Fields(env), nil
	}

	var srcs []string
	for _, src := range p.Manifest.Sources {
		src := fix(p.Root(), p.expand(src))

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

	if env := getenv("NTT_IMPORTS"); env != "" {
		return strings.Fields(env), nil
	}

	var imports []string
	for _, dir := range p.Manifest.Imports {
		imports = append(imports, fix(p.Root(), p.expand(dir)))
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
			name := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
			p.modules[name] = file
		}
	}
	if file, ok := p.modules[name]; ok {
		return file, nil
	}

	// Use NTT_CACHE to locate file
	if f := fs.Open(name + ".ttcn3").Path(); isRegular(f) {
		p.modules[name] = f
		return f, nil
	}

	return "", fmt.Errorf("No such module %q", name)
}
func (p *Project) expand(s string) string {
	// TODO(5nord) implement expansion with variables section
	return s
}

func (p *Project) readManifest(file string) error {
	b, err := fs.Open(file).Bytes()
	if err != nil {
		return err
	}
	return yaml.UnmarshalStrict(b, &p.Manifest)
}

func (p *Project) readFilesystem() {
	addSources := func(path string, info os.FileInfo, err error) error {
		if err == nil && info.IsDir() {
			files, _ := filepath.Glob(filepath.Join(path, "*.ttcn*"))
			if len(files) > 0 {
				log.Debugf("adding sources: %q\n", path)
				p.Manifest.Sources = append(p.Manifest.Sources, rel(p.Root(), files...)...)
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
		if path := filepath.Join(p.root, dir); isDir(path) {
			filepath.Walk(path, addImports)
		}
	}
}

func rel(base string, paths ...string) []string {
	if len(paths) == 0 {
		return nil
	}
	ret := make([]string, len(paths))
	for i, path := range paths {
		if r, err := filepath.Rel(base, path); err == nil {
			ret[i] = r
		} else {
			ret[i] = path
		}
	}
	return ret
}
