// Package k3 provides convenience functions for supporting k3 toolchain. k3 is
// the predecessor of the ntt project.
package k3

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar"
	"github.com/nokia/ntt/internal/cache"
	"github.com/nokia/ntt/internal/cmake"
	"github.com/nokia/ntt/internal/env"
	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/proc"
)

// ErrNotFound is returned when k3 or part of k3 is not found.
var ErrNotFound = errors.New("not found")

type Instance struct {
	compiler  string
	runtime   string
	plugins   []string
	includes  []string
	libK3     string
	cLibDirs  []string
	cIncludes []string

	ossInfo     string
	ossDefaults string
}

func New(prefix string) (*Instance, error) {
	k := Instance{}

	if k3c := env.Getenv("K3C"); k3c != "" {
		k.compiler = k3c
	} else {
		k3c, err := glob(prefix + "/{bin,libexec}/{mtc,k3c}")
		if err != nil {
			return nil, fmt.Errorf("%s: mtc: %w", prefix, err)
		}
		k.compiler = k3c[0]
	}

	if k3r := env.Getenv("K3R"); k3r != "" {
		k.runtime = k3r
	} else {
		k3r, err := glob(prefix + "/{bin,libexec}/k3r")
		if err != nil {
			return nil, fmt.Errorf("k3r: %w", err)
		}
		k.runtime = k3r[0]
	}

	pluginLib, err := glob(prefix + "/{lib64,lib}/libk3-plugin.so")
	if err != nil {
		return nil, fmt.Errorf("libk3-plugin.so: %w", err)
	}
	k.cLibDirs = append(k.cLibDirs, filepath.Dir(pluginLib[0]))

	cIncludes, err := glob(prefix + "/include/k3")
	if err != nil {
		return nil, fmt.Errorf("k3 includes: %w", err)
	}
	k.cIncludes = append(k.cIncludes, filepath.Dir(cIncludes[0]))

	plugins, err := glob(prefix + "/{lib64,lib}/k3/plugins/")
	if err != nil {
		return nil, fmt.Errorf("k3 plugins: %w", err)
	}
	k.plugins = append(k.plugins, plugins[0])

	includes, err := glob(
		k.plugins[0]+"/ttcn3",
		prefix+"/share/k3/ttcn3")
	if err != nil {
		return nil, fmt.Errorf("k3 includes: %w", err)
	}
	k.includes = includes

	if ossInfo, err := glob(prefix + "/share/k3/asn1/ossinfo"); err == nil {
		k.ossInfo = filepath.Dir(ossInfo[0])
	}
	if ossDefaults, err := glob(prefix + "/share/k3/asn1/asn1dflt.*"); err == nil {
		k.ossDefaults = ossDefaults[0]
	}

	return &k, nil
}

var k3 = &Instance{
	runtime:  "k3r",
	compiler: "mtc",
}

func init() {
	for _, ev := range []string{"NTTROOT", "K3ROOT"} {
		if root := env.Getenv(ev); root != "" {
			k, err := New(root)
			if err != nil {
				log.Printf("k3: %s: %s\n", ev, err)
				return
			}
			if k != nil {
				k3 = k
			}
		}
	}

	k, err := findDevel()
	if err != nil {
		log.Printf("k3: devel: %s\n", err)
		return
	}
	if k != nil {
		k3 = k
		return
	}

	if prefix := findPrefix(); prefix != "" {
		k, err := New(prefix)
		if err != nil {
			log.Printf("k3: %s\n", err)
		}
		if k != nil {
			k3 = k
		}
	}
}

func findDevel() (*Instance, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	cache := cmake.FindCache(cwd)
	if cache == nil {
		return nil, nil
	}

	k3BinaryDir, err := cache.Get("K3_BINARY_DIR")
	if err != nil {
		if errors.Is(err, cmake.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	k3SourceDir, err := cache.Get("K3_SOURCE_DIR")
	if err != nil {
		return nil, err
	}
	mtcBinaryDir, err := cache.Get("mtc_BINARY_DIR")
	if err != nil {
		return nil, err
	}

	k := Instance{}

	if k3c := env.Getenv("K3C"); k3c != "" {
		k.compiler = k3c
	} else {
		k3c, err := glob(mtcBinaryDir + "/source/mtc/mtc")
		if err != nil {
			return nil, fmt.Errorf("mtc: %w", err)
		}
		k.compiler = k3c[0]
	}

	if k3r := env.Getenv("K3R"); k3r != "" {
		k.runtime = k3r
	} else {
		k3r, err := glob(k3BinaryDir + "/src/k3r/k3r")
		if err != nil {
			return nil, fmt.Errorf("k3r: %w", err)
		}
		k.runtime = k3r[0]
	}

	pluginLib, err := glob(k3BinaryDir + "/src/libk3-plugin/libk3-plugin.so")
	if err != nil {
		return nil, fmt.Errorf("libk3-plugin.so: %w", err)
	}
	k.cLibDirs = append(k.cLibDirs, filepath.Dir(pluginLib[0]))

	ossLibs, err := glob(k3SourceDir + "/lib/ossasn1/lib64")
	if err != nil {
		return nil, fmt.Errorf("oss libs: %w", err)
	}
	k.cLibDirs = append(k.cLibDirs, ossLibs...)

	cIncludes, err := glob(k3SourceDir + "/include/k3")
	if err != nil {
		return nil, fmt.Errorf("k3 includes: %w", err)
	}
	k.cIncludes = append(k.cIncludes, filepath.Dir(cIncludes[0]))

	ossIncludes, err := glob(k3SourceDir + "/lib/ossasn1/include")
	if err != nil {
		return nil, fmt.Errorf("oss includes: %w", err)
	}
	k.cIncludes = append(k.cIncludes, ossIncludes...)

	plugins, err := glob(k3BinaryDir + "/src/k3r-*-plugin")
	if err != nil {
		return nil, fmt.Errorf("k3 plugins: %w", err)
	}
	k.plugins = append(k.plugins, plugins[0])

	includes, err := glob(
		k3SourceDir+"/src/k3r-*-plugin",
		k3SourceDir+"/src/libzmq",
		k3SourceDir+"/src/ttcn3",
	)
	if err != nil {
		return nil, fmt.Errorf("k3 includes: %w", err)
	}
	k.includes = includes

	if ossInfo, err := glob(k3BinaryDir + "/ossinfo"); err == nil {
		k.ossInfo = filepath.Dir(ossInfo[0])
	}
	if ossDefaults, err := glob(k3SourceDir + "/lib/ossasn1/asn1dflt.*"); err == nil {
		k.ossDefaults = ossDefaults[0]
	}

	return &k, nil

}

func findPrefix() string {
	for _, name := range []string{"k3r", "k3s"} {
		if exe, err := exec.LookPath(name); err == nil {
			prefix, err := filepath.EvalSymlinks(filepath.Join(filepath.Dir(exe), ".."))
			if err != nil {
				log.Printf("k3: %s: %s\n", exe, err)
				return ""
			}
			return prefix
		}
	}
	return ""
}

func glob(patterns ...string) ([]string, error) {
	var ret []string
	for _, p := range patterns {
		matches, err := doublestar.Glob(p)
		if err != nil {
			return nil, err
		}
		ret = append(ret, matches...)
	}
	if len(ret) == 0 {
		return nil, ErrNotFound
	}
	return ret, nil
}

// DefaultEnv is the default environment for k3-based test suites.
var DefaultEnv = map[string]string{
	"CXX":        "g++",
	"CC":         "gcc",
	"ASN1C":      "asn1",
	"ASN1CFLAGS": "-reservedWords ffs -c -charIntegers -listingFile -messageFormat emacs -noDefines -valuerefs -debug -root -soed",
	"ASN2TTCN":   "asn1tottcn3",
}

// OssInfo returns the path to the ossinfo file.
func OssInfo() string {
	return k3.ossInfo
}

// Compiler returns the path to the TTCN-3 compiler. Compiler will return "mtc"
// if no compiler is found.
func Compiler() string {
	return k3.compiler
}

// Runtime returns the path to the TTCN-3 runtime. Runtime will return "k3r" if
// no runtime is found.
func Runtime() string {
	return k3.runtime
}

func CLibDirs() []string {
	return k3.cLibDirs
}

// Plugins returns a list of k3 plugins.
func Plugins() []string {
	return k3.plugins
}

// Includes returns a list of TTCN-3 include directories required by the k3 compiler.
func Includes() []string {
	return k3.includes
}

// NewASN1Codec returns the commands required to compile ASN.1 files.
func NewASN1Codec(vars map[string]string, name string, encoding string, srcs ...string) []*proc.Cmd {
	if vars == nil {
		vars = make(map[string]string)
		for k, v := range DefaultEnv {
			vars[k] = v
		}
	}

	if _, ok := vars["OSSINFO"]; !ok && k3.ossInfo != "" {
		vars["OSSINFO"] = k3.ossInfo
	}

	var cFlags []string
	if f, ok := vars["CFLAGS"]; ok {
		cFlags = append(cFlags, f)
	}
	for _, inc := range k3.cIncludes {
		cFlags = append(cFlags, "-I"+inc)
	}
	vars["CFLAGS"] = strings.Join(cFlags, " ")

	var ldFlags []string
	if f, ok := vars["LDFLAGS"]; ok {
		ldFlags = append(ldFlags, f)
	}
	for _, dir := range k3.cLibDirs {
		ldFlags = append(ldFlags, "-L"+dir)
	}
	vars["LDFLAGS"] = strings.Join(ldFlags, " ")

	vars["name"] = name
	vars["encoding"] = encoding

	asn1 := proc.Task("$ASN1C -${encoding} $ASN1CFLAGS -output ${name}.enc -prefix ${name} ${srcs}")
	asn1.Env = vars
	asn1.Sources = []string{k3.ossDefaults}
	asn1.Sources = append(asn1.Sources, srcs...)
	asn1.Targets = []string{
		pathf("%s.enc.c", name),
		pathf("%s.enc.h", name),
	}

	lib := proc.Task("$CC -fPIC -shared -D_OSSGETHEADER -DOSSPRINT $CPPFLAGS $CFLAGS $LDFLAGS $EXTRA_LDFLAGS ${srcs} -l:libasn1code.a -Wl,-Bdynamic -o ${tgts}")
	lib.Env = vars
	lib.Sources = asn1.Targets
	lib.Targets = []string{
		pathf("%slib.so", name),
	}

	mod := proc.Task("$ASN2TTCN -o ${tgts} ${srcs} ${name}mod  ${encoding}")
	mod.Env = vars
	mod.Sources = lib.Targets
	mod.Targets = []string{
		pathf("%smod.ttcn3", name),
	}

	return []*proc.Cmd{asn1, lib, mod}
}

// NewPlugin returns the commands for building a k3 plugin.
func NewPlugin(vars map[string]string, name string, srcs ...string) []*proc.Cmd {
	if vars == nil {
		vars = make(map[string]string)
		for k, v := range DefaultEnv {
			vars[k] = v
		}
	}

	var cFlags []string
	if f, ok := vars["CFLAGS"]; ok {
		cFlags = append(cFlags, f)
	}
	for _, inc := range k3.cIncludes {
		cFlags = append(cFlags, "-I"+inc)
	}
	vars["CFLAGS"] = strings.Join(cFlags, " ")

	var ldFlags []string
	if f, ok := vars["LDFLAGS"]; ok {
		ldFlags = append(ldFlags, f)
	}
	for _, dir := range k3.cLibDirs {
		ldFlags = append(ldFlags, "-L"+dir)
	}
	vars["LDFLAGS"] = strings.Join(ldFlags, " ")

	p := proc.Task("$CXX $CPPFLAGS $CFLAGS $CXXFLAGS -shared -fPIC -o ${tgts} ${srcs} $LDFLAGS $EXTRA_LDFLAGS -lk3-plugin")
	p.Env = vars
	p.Sources = srcs
	p.Targets = []string{pathf("k3r-%s-plugin.so", name)}
	return []*proc.Cmd{p}
}

// NewT3XF returns the commands for building a T3XF.
func NewT3XF(vars map[string]string, t3xf string, srcs []string, imports ...string) []*proc.Cmd {
	if vars == nil {
		vars = make(map[string]string)
		for k, v := range DefaultEnv {
			vars[k] = v
		}
	}
	if _, ok := vars["K3C"]; !ok {
		vars["K3C"] = k3.compiler
	}

	// We need to remove k3 stdlib files from the source list, (if accidentally
	// inserted by the user) because of a missing module (PCMDmod).
	vars["_sources"] = strings.Join(removeStdlib(srcs), " ")

	for _, dir := range k3.includes {
		vars["_includes"] += fmt.Sprintf(" -I%s", dir)
	}

	// We must not use imported TTCN-3 files directly, but their include directory
	// instead. Because of some missing protobuf modules.
	visited := make(map[string]bool)
	for _, file := range imports {
		if dir := filepath.Dir(file); !visited[dir] {
			visited[dir] = true
			vars["_includes"] += fmt.Sprintf(" -I%s", dir)
		}
	}

	t := proc.Task("$K3C $K3CFLAGS $_includes -o ${tgts} ${_sources}")
	t.Sources = append(srcs, imports...)
	t.Targets = []string{t3xf}
	t.Env = vars
	t.Before = func(t *proc.Cmd) error {
		// We must not modify the t3xf file while it's beeing executed by k3r.
		// Removing it ensures a new inode and all is fine.
		if err := os.Remove(t3xf); !errors.Is(err, os.ErrNotExist) {
			return err
		}
		return nil
	}
	return []*proc.Cmd{t}
}

func removeStdlib(srcs []string) []string {

	// There are multiple installations of the stdlib, so we cannot compare
	// for identity but use a hash instead.
	sum := func(file string) ([]byte, error) {
		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		h := sha256.New()
		if _, err := io.Copy(h, f); err != nil {
			return nil, err
		}
		return h.Sum(nil), nil
	}

	// We build a map of hashes of the stdlib files. The key is the base
	// name of the file.
	stdlib := make(map[string][]byte)
	for _, dir := range Includes() {
		for _, file := range fs.FindTTCN3Files(dir) {
			if s, err := sum(file); err == nil {
				stdlib[fs.Stem(file)] = s
			}
		}
	}

	isStdlib := func(file string) bool {
		if s1 := stdlib[fs.Stem(file)]; s1 != nil {
			if s2, err := sum(file); err == nil {
				return bytes.Equal(s1, s2)
			}
		}
		return false
	}

	var ret []string
	for _, src := range srcs {
		if !isStdlib(src) {
			ret = append(ret, src)
		}
	}
	return ret
}

type TTCN3Library struct{ srcs []string }

func (l *TTCN3Library) String() string    { return fmt.Sprintf("%s", l.srcs) }
func (l *TTCN3Library) Outputs() []string { return l.srcs }
func (l *TTCN3Library) Inputs() []string  { return l.srcs }
func (l *TTCN3Library) Run() error        { return nil }

// NewTTCN3Library returns the commands for building a TTCN-3 library.
func NewTTCN3Library(vars map[string]string, name string, srcs ...string) []*TTCN3Library {
	return []*TTCN3Library{{srcs}}
}

// pathf is like fmt.Sprintf but searches the NTT_CACHE environment variable first.
func pathf(f string, v ...interface{}) string {
	return cache.Lookup(fmt.Sprintf(f, v...))
}
