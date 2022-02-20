package build

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/target"
	"github.com/nokia/ntt/internal/cache"
	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/ntt"
	"github.com/nokia/ntt/k3"
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
			name, err := suite.Name()
			if err != nil {
				return err
			}
			return build(name, suite)
		},
	}

	ErrNoSources = fmt.Errorf("no sources available")

	DefaultEnv = map[string]string{
		"CXX":        "g++",
		"CC":         "gcc",
		"K3C":        k3.Compiler(),
		"K3CFLAGS":   "--no-watermark --format=tasm",
		"ASN1C":      "asn1",
		"ASN1CFLAGS": "-reservedWords ffs -c -charIntegers -listingFile -messageFormat emacs -noDefines -valuerefs -debug -root -soed",
		"ASN2TTCN":   "asn1tottcn3",
	}
)

func build(name string, p project.Interface) error {
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

	return buildTTCN3(name, srcs...)
}

func buildTTCN3(name string, srcs ...string) error {
	out := Outf("%s.t3xf", name)
	b, err := target.Path(out, srcs...)
	if !b || err != nil {
		return err
	}

	args := []string{"-o", out}
	if env := os.Getenv("K3CFLAGS"); env != "" {
		args = append(args, strings.Fields(env))
	}
	visited := make(map[string]bool)
	for _, src := range srcs {
		dir := filepath.Dir(src)
		if !visited[dir] {
			args = append(args, "-I", dir)
			visited[dir] = true
		}
		args = append(args, src)
	}
	for _, dir := range k3.FindAuxiliaryDirectories() {
		args = append(args, "-I", dir)
	}

	return Exec("$K3C", args...)
}

func buildImport(dir string) ([]string, error) {
	name := fs.Slugify(fs.Stem(dir))

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var (
		asn1Files, ttcn3Files, cFiles []string
		processed                     int
	)

	for _, f := range files {
		switch path := filepath.Join(dir, f.Name()); filepath.Ext(path) {
		case ".asn1", ".asn":
			asn1Files = append(asn1Files, path)
			processed++
		case ".ttcn3", ".ttcn":
			ttcn3Files = append(ttcn3Files, path)
			processed++
		case ".c", ".cxx", ".cpp", ".cc":
			cFiles = append(cFiles, path)
			processed++
		}
	}

	if processed == 0 {
		return nil, fmt.Errorf("%s: %w", dir, ErrNoSources)
	}

	if len(asn1Files) > 0 {
		out, err := genASN1(name, "", asn1Files...)
		if err != nil {
			return nil, err
		}

		if err := buildASN1(name, out...); err != nil {
			return nil, err
		}

		//cmd = "$CC -fPIC -shared $^ -D_OSSGETHEADER -DOSSPRINT $CFLAGS $LDFLAGS $EXTRA_LDFLAGS -l:libasn1code.a -Wl,-Bdynamic -o $@"
		//if err := run(cmd, ""); err != nil {
		//	return nil, err
		//}
		//cmd = fmt.Sprintf("$ASN2TTCN -o $@ $< %smod $ENCODING", name)
		//if err := run(cmd, ""); err != nil {
		//	return nil, err
		//}
	}

	if len(cFiles) > 0 {
		out, err := buildAdapter(name, cFiles...)
		if err != nil {
			return nil, err
		}
		for _, f := range out {
			if fs.HasTTCN3Extension(f) {
				ttcn3Files = append(ttcn3Files, f)
			}

		}
	}
	return ttcn3Files, nil
}

func genASN1(name string, encoding string, srcs ...string) ([]string, error) {

	c := Outf("%senc.c", name)
	h := Outf("%senc.h", name)
	if ok, err := target.Dir(c, srcs...); !ok || err != nil {
		return []string{c, h}, err
	}
	if ok, err := target.Dir(h, srcs...); !ok || err != nil {
		return []string{c, h}, err
	}

	//$ASN1FLAGS
	cmd := "$ASN1C $OSSINFO/asn1dflt.linux-$(uname -m) $^ -$ENCODING $ALL_ASN1FLAGS -output $*.enc -prefix $*"
	return []string{c, h}, Exec("$ASN1C", args...)

}

func buildASN1(name string, srcs ...string) ([]string, error) {
	return nil, nil
}

func buildAdapter(name string, srcs ...string) ([]string, error) {
	out := Outf("k3r-%s-plugin.so", name)
	b, err := target.Path(out, srcs...)
	if !b || err != nil {
		return nil, err
	}

	var args []string

	if env := os.Getenv("CXXFLAGS"); env != "" {
		args = append(args, env)
	}
	if env := os.Getenv("LDFLAGS"); env != "" {
		args = append(args, env)
	}
	if env := os.Getenv("EXTRA_LDFLAGS"); env != "" {
		args = append(args, env)
	}
	args = append(args, "-lk3-plugin", "-shared", "-fPIC", "-o", out)
	return []string{out}, Exec("$CXX", args...)
}

func Exec(name string, args ...string) error {
	expand := func(key string) string {
		if v, ok := os.LookupEnv(key); ok {
			return v
		}
		return DefaultEnv[key]
	}

	name = os.Expand(name, expand)
	for i := range args {
		args[i] = os.Expand(args[i], expand)
	}

	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	log.Debugln("+", cmd.String())
	return cmd.Run()

}

func run(cmd string, files ...string) error {
	return nil
}

func Outf(f string, v ...interface{}) string {
	return cache.Lookup(fmt.Sprintf(f, v...))
}
