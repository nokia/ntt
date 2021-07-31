// Package project collects information about test suite organisation by
// implementing various heuristics.
package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/project/manifest"
	"gopkg.in/yaml.v2"
)

// Project describes a TTCN-3 project.
type Project interface {
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
func Files(p Project) ([]string, error) {
	files, err := p.Sources()
	if err != nil {
		return nil, err
	}

	dirs, err := p.Imports()
	if err != nil {
		return nil, err
	}

	for _, dir := range dirs {
		f := fs.FindTTCN3Files(dir)
		files = append(files, f...)
	}

	return files, nil
}

// ContainsFile returns true, when path is managed by Project.
func ContainsFile(p Project, path string) bool {
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

func Open(path string) (Project, error) {
	p := project{
		root: path,
	}

	return &p, p.Update()
}

type project struct {
	root     string
	Manifest manifest.Config
}

func (p *project) Update() error {
	if file := filepath.Join(p.root, manifest.Name); isRegular(file) {
		log.Debugf("%s: update configuration using manifest %q", p.String(), file)
		return p.readManifest(file)
	}
	log.Debugf("%s: update configuration using available folders", p.String())
	p.readFilesystem()
	return nil
}

func (p *project) String() string {
	return p.root
}

// Root directory is ttcn3 suite.
func (p *project) Root() string {
	return p.root
}

func (p *project) Name() string {
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

func (p *project) TestHook() string {
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

func (p *project) ParametersFile() string {
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

func (p *project) ParametersDir() string {
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
func (p *project) Sources() ([]string, error) {
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

func (p *project) Imports() ([]string, error) {
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

func (p *project) expand(s string) string {
	// TODO(5nord) implement expansion with variables section
	return s
}

func (p *project) readManifest(file string) error {
	b, err := fs.Open(file).Bytes()
	if err != nil {
		return err
	}
	return yaml.UnmarshalStrict(b, &p.Manifest)
}

func (p *project) readFilesystem() {
	addSources := func(path string, info os.FileInfo, err error) error {
		if err == nil && info.IsDir() {
			files, _ := filepath.Glob(filepath.Join(path, "*.ttcn*"))
			if len(files) > 0 {
				log.Debugf("adding sources: %q", path)
				p.Manifest.Sources = append(p.Manifest.Sources, files...)
			}
		}
		return nil
	}
	filepath.Walk(p.root, addSources)

	addImports := func(path string, info os.FileInfo, err error) error {
		if err == nil && info.IsDir() {
			files, _ := filepath.Glob(filepath.Join(path, "*.ttcn*"))
			if len(files) > 0 {
				log.Debugf("adding import: %q", path)
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
			filepath.Walk(filepath.Join(p.root, "../../../../"), addImports)
		}
	}
}
