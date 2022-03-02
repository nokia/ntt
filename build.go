package ntt

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nokia/ntt/build"
	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/k3"
	"github.com/nokia/ntt/project"
)

// ErrNoSources is returned when no source files are found.
var ErrNoSources = fmt.Errorf("no sources")

// BuildProject builds a project with the given name. It the build of the
// project or any of its dependencies fails, the error is returned.
func BuildProject(name string, p project.Interface) error {
	start := time.Now()
	defer func() {
		log.Debugf("build %s took %s\n", name, time.Since(start))
	}()
	builders, err := PlanProject(name, p)
	if err != nil {
		return err
	}
	for _, b := range builders {
		if err := b.Build(); err != nil {
			return err
		}
	}
	return err
}

// PlanProject returns a list of builders required to build the given project.
//
// Using a different backend than k3 is currently not supported.
func PlanProject(name string, p project.Interface) ([]build.Builder, error) {
	var ret []build.Builder

	srcs, err := p.Sources()
	if err != nil {
		return nil, err
	}

	imports, err := p.Imports()
	if err != nil {
		return nil, err
	}

	for _, dir := range imports {
		builders, err := PlanImport(dir)
		if err != nil {
			return nil, err
		}
		for _, b := range builders {
			for _, o := range b.Targets() {
				if fs.HasTTCN3Extension(o) {
					srcs = append(srcs, o)
				}
			}

			// Skip T3XF Builders, because we'll build t3xf speparately.
			if t, ok := b.(*k3.T3XF); ok {
				srcs = append(srcs, t.Sources()...)
			} else {
				ret = append(ret, b)
			}
		}
	}

	if len(srcs) == 0 {
		return nil, fmt.Errorf("test root folder %s: %w", p.Root(), ErrNoSources)
	}

	ret = append(ret, k3.NewT3XFBuilder(name, srcs...))
	return ret, nil
}

// PlanImport returns a list of builders required to build the given import
// directory. An import directory can be a TTCN-3 library, a ASN.1 codec or a
// k3 plugin.
//
// Using a different backend than k3 is currently not supported.
func PlanImport(dir string) ([]build.Builder, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var (
		builders                      []build.Builder
		asn1Files, ttcn3Files, cFiles []string
		processed                     int
	)

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		switch path := filepath.Join(dir, f.Name()); filepath.Ext(path) {
		case ".asn1", ".asn":
			asn1Files = append(asn1Files, path)
			processed++
		case ".ttcn3", ".ttcn":
			ttcn3Files = append(ttcn3Files, path)
			processed++
		case ".c", ".cxx", ".cpp", ".cc":
			// Skip ASN1 codecs
			if strings.HasSuffix(path, ".enc.c") {
				continue
			}
			cFiles = append(cFiles, path)
			processed++
		}
	}
	if processed == 0 {
		return nil, fmt.Errorf("%s: %w", dir, ErrNoSources)
	}
	name := fs.Slugify(fs.Stem(dir))
	if len(asn1Files) > 0 {
		builders = append(builders, k3.NewASN1Codec(name, encoding(name), asn1Files...))
	}
	if len(cFiles) > 0 {
		builders = append(builders, k3.NewPluginBuilder(name, cFiles...))
	}
	if len(ttcn3Files) > 0 {
		builders = append(builders, k3.NewT3XFBuilder(name, ttcn3Files...))
	}
	return builders, nil
}

func encoding(name string) string {
	if strings.Contains(strings.ToLower(name), "rrc") {
		return "uper"
	}
	return "per"
}
