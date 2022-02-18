package build

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nokia/ntt/internal/cache"
	"github.com/nokia/ntt/internal/ntt"
	"github.com/nokia/ntt/project"
	"github.com/spf13/cobra"
)

var (
	Command = &cobra.Command{
		Use:   "build",
		Short: "Builds compiles TTCN-3 source and imports specified by the import paths.",
		RunE: func(cmd *cobra.Command, args []string) error {
			suite, err := ntt.NewFromArgs(args...)
			if err != nil {
				return err
			}
			return build(suite)
		},
	}

	ErrNoSources = fmt.Errorf("no sources available")
)

func build(p project.Interface) error {
	srcs, err := p.Sources()
	if err != nil {
		return err
	}

	imports, err := p.Imports()
	if err != nil {
		return err
	}
	for _, dir := range imports {
		files, err := buildImport(dir)
		if err != nil {
			return err
		}
		srcs = append(srcs, files...)
	}

	return nil
}

func buildImport(dir string) ([]string, error) {
	//name := fs.Slugify(fs.Stem(dir))

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var asn1Files, ttcn3Files, cFiles []string
	for _, f := range files {
		switch filepath.Ext(f.Name()) {
		case ".asn1":
			asn1Files = append(asn1Files, filepath.Join(dir, f.Name()))
		case ".ttcn3":
			ttcn3Files = append(ttcn3Files, filepath.Join(dir, f.Name()))
		case ".c":
			cFiles = append(cFiles, filepath.Join(dir, f.Name()))
		}
	}

	if len(asn1Files) == 0 && len(ttcn3Files) == 0 && len(cFiles) == 0 {
		return nil, fmt.Errorf("%s: %w", dir, ErrNoSources)
	}

	return nil, nil
}

func Filef(f string, v ...interface{}) string {
	return cache.Lookup(fmt.Sprintf(f, v...))
}
