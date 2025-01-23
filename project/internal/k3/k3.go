// Package k3 provides convenience functions for supporting k3 toolchain. k3 is
// the predecessor of the ntt project.
package k3

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/trace"
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
	Root string

	Compiler  string
	Runtime   string
	Plugins   []string
	Includes  []string
	CLibDirs  []string
	CIncludes []string

	OssInfo     string
	ossDefaults string
}

var k3 = &Instance{
	Runtime:  "k3r",
	Compiler: "mtc",
}

type searchFunc func(ctx context.Context) (*Instance, error)

func Find() Instance {
	return *k3
}

func init() {

	ctx, task := trace.NewTask(context.Background(), "k3.init")
	defer task.End()

	searchers := []searchFunc{
		searchEnvionment("NTTROOT"),
		searchEnvionment("K3ROOT"),
		searchRepo(),
		searchPath("k3r"),
		searchPath("k3s"),
		searchPath("ntt"),
	}

	for _, s := range searchers {
		k, err := s(ctx)
		if err != nil {
			log.Tracef(ctx, "k3", "error: %s\n", err)
			return
		}
		if k != nil {
			log.Tracef(ctx, "k3", "using: %+v\n", k)
			k3 = k
			return
		}
	}
	log.Tracef(ctx, "k3", "not found\n")
}

func searchEnvionment(v string) searchFunc {
	return func(ctx context.Context) (*Instance, error) {
		region := trace.StartRegion(ctx, "search environment")
		defer region.End()
		if root := env.Getenv(v); root != "" {
			return getInstall(ctx, root)
		}
		return nil, nil
	}
}

func searchRepo() searchFunc {
	return func(ctx context.Context) (*Instance, error) {
		region := trace.StartRegion(ctx, "search repo")
		defer region.End()

		cwd, err := os.Getwd()
		if err != nil {
			return nil, err

		}

		cache := cmake.FindCache(cwd)
		if cache == nil {
			return nil, nil
		}
		log.Tracef(ctx, "k3", "found CMakeCache.txt: %+v\n", cache)

		k3BinaryDir, err := cache.Get("k3_BINARY_DIR")
		if err != nil {
			if errors.Is(err, cmake.ErrNotFound) {
				log.Tracef(ctx, "k3", "k3_BINARY_DIR not found in CMakeCache.txt\n")
				return nil, nil
			}
			return nil, err
		}
		k3SourceDir, err := cache.Get("k3_SOURCE_DIR")
		if err != nil {
			return nil, err
		}
		mtcBinaryDir, err := cache.Get("mtc_BINARY_DIR")
		if err != nil && !errors.Is(err, cmake.ErrNotFound) {
			return nil, err
		}
		return getRepo(ctx, k3SourceDir, k3BinaryDir, mtcBinaryDir)
	}
}

func searchPath(exe string) searchFunc {
	return func(ctx context.Context) (*Instance, error) {
		region := trace.StartRegion(ctx, "search path")
		defer region.End()

		path, err := exec.LookPath(exe)
		if err != nil {
			if errors.Is(err, exec.ErrNotFound) {
				// Clear error to allow other searchers to run.
				err = nil
			}
			return nil, err
		}

		prefix := filepath.Clean(filepath.Join(filepath.Dir(path), ".."))

		// Symlinks make checking for directories harder. We resolve them, when possible.
		if target, err := filepath.EvalSymlinks(prefix); err == nil {
			log.Tracef(ctx, "k3", "resolved symlink: %s -> %s\n", prefix, target)
			prefix = target
		}

		k, err := getInstall(ctx, prefix)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				// Clear error to allow other searchers to run.
				err = nil
			}
			return nil, err
		}
		return k, nil
	}
}

func getInstall(ctx context.Context, prefix string) (*Instance, error) {

	region := trace.StartRegion(ctx, "get install")
	defer region.End()
	log.Tracef(ctx, "k3", "get install: %s\n", prefix)

	k := Instance{Root: prefix}

	if k3c := env.Getenv("K3C"); k3c != "" {
		k.Compiler = k3c
	} else {
		k3c, err := glob(ctx, prefix+"/{libexec,bin}/{mtc,k3c}")
		if err != nil {
			return nil, fmt.Errorf("%s: mtc: %w", prefix, err)
		}
		// glob unfortunately sorts the results, so we need to take the last one.
		k.Compiler = k3c[len(k3c)-1]
	}

	if k3r := env.Getenv("K3R"); k3r != "" {
		k.Runtime = k3r
	} else {
		k3r, err := glob(ctx, prefix+"/{libexec,bin}/k3r")
		if err != nil {
			return nil, fmt.Errorf("k3r: %w", err)
		}
		// glob unfortunately sorts the results, so we need to take the last one.
		k.Runtime = k3r[len(k3r)-1]
	}

	pluginLib, err := glob(ctx, prefix+"/{lib64,lib}/libk3-plugin.so")
	if err != nil {
		return nil, fmt.Errorf("libk3-plugin.so: %w", err)
	}
	k.CLibDirs = append(k.CLibDirs, filepath.Dir(pluginLib[0]))

	cIncludes, err := glob(ctx, prefix+"/include/k3")
	if err != nil {
		return nil, fmt.Errorf("k3 includes: %w", err)
	}
	k.CIncludes = append(k.CIncludes, filepath.Dir(cIncludes[0]))

	plugins, err := glob(ctx, prefix+"/{lib64,lib}/k3/plugins/")
	if err != nil {
		return nil, fmt.Errorf("k3 plugins: %w", err)
	}
	k.Plugins = append(k.Plugins, plugins[0])

	includes, err := glob(ctx,
		k.Plugins[0]+"/ttcn3",
		prefix+"/share/k3/ttcn3")
	if err != nil {
		return nil, fmt.Errorf("k3 includes: %w", err)
	}
	k.Includes = includes

	if ossInfo, err := glob(ctx, prefix+"/share/k3/asn1/ossinfo"); err == nil {
		k.OssInfo = filepath.Dir(ossInfo[0])
	}
	if ossDefaults, err := glob(ctx, prefix+"/share/k3/asn1/asn1dflt.*"); err == nil {
		k.ossDefaults = ossDefaults[0]
	}

	return &k, nil
}

func getRepo(ctx context.Context, k3SourceDir, k3BinaryDir, mtcBinaryDir string) (*Instance, error) {
	region := trace.StartRegion(ctx, "get repo")
	defer region.End()
	log.Tracef(ctx, "k3", "get repo: %s %s %s\n", k3SourceDir, k3BinaryDir, mtcBinaryDir)

	k := Instance{Root: k3BinaryDir}

	if k3c := env.Getenv("K3C"); k3c != "" {
		k.Compiler = k3c
	} else if mtcBinaryDir != "" {
		k3c, err := glob(ctx, mtcBinaryDir+"/source/mtc/mtc")
		if err != nil {
			return nil, fmt.Errorf("mtc: %w", err)
		}
		k.Compiler = k3c[0]
	} else {
		k.Compiler = "mtc"
	}

	if k3r := env.Getenv("K3R"); k3r != "" {
		k.Runtime = k3r
	} else {
		k3r, err := glob(ctx, k3BinaryDir+"/src/k3r/k3r")
		if err != nil {
			return nil, fmt.Errorf("k3r: %w", err)
		}
		k.Runtime = k3r[0]
	}

	pluginLib, err := glob(ctx, k3BinaryDir+"/src/libk3-plugin/libk3-plugin.so")
	if err != nil {
		return nil, fmt.Errorf("libk3-plugin.so: %w", err)
	}
	k.CLibDirs = append(k.CLibDirs, filepath.Dir(pluginLib[0]))

	ossLibs, err := glob(ctx, k3SourceDir+"/lib/ossasn1/lib64")
	if err != nil {
		return nil, fmt.Errorf("oss libs: %w", err)
	}
	k.CLibDirs = append(k.CLibDirs, ossLibs...)

	cIncludes, err := glob(ctx, k3SourceDir+"/include/k3")
	if err != nil {
		return nil, fmt.Errorf("k3 includes: %w", err)
	}
	k.CIncludes = append(k.CIncludes, filepath.Dir(cIncludes[0]))

	ossIncludes, err := glob(ctx, k3SourceDir+"/lib/ossasn1/include")
	if err != nil {
		return nil, fmt.Errorf("oss includes: %w", err)
	}
	k.CIncludes = append(k.CIncludes, ossIncludes...)

	plugins, err := glob(ctx, k3BinaryDir+"/src/k3r-*-plugin")
	if err != nil {
		return nil, fmt.Errorf("k3 plugins: %w", err)
	}
	k.Plugins = append(k.Plugins, plugins...)

	includes, err := glob(ctx,
		k3SourceDir+"/src/k3r-*-plugin",
		k3SourceDir+"/src/libzmq",
		k3SourceDir+"/src/ttcn3",
	)
	if err != nil {
		return nil, fmt.Errorf("k3 includes: %w", err)
	}
	k.Includes = includes

	if ossInfo, err := glob(ctx, k3BinaryDir+"/ossinfo"); err == nil {
		k.OssInfo = filepath.Dir(ossInfo[0])
	}
	if ossDefaults, err := glob(ctx, k3SourceDir+"/lib/ossasn1/asn1dflt.*"); err == nil {
		k.ossDefaults = ossDefaults[0]
	}

	return &k, nil

}

func glob(ctx context.Context, patterns ...string) ([]string, error) {
	region := trace.StartRegion(ctx, "glob")
	defer region.End()

	var ret []string
	for _, p := range patterns {
		matches, err := doublestar.Glob(p)
		log.Tracef(ctx, "k3", "glob: %s -> %v\n", p, matches)
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
	"ASN1CFLAGS": "-reservedWords ffs -c -charIntegers -listingFile -messageFormat emacs -noDefines -valuerefs -debug -root -soed -compat decoderUpdatesInputAddress",
	"ASN2TTCN":   "asn1tottcn3",
}

// OssInfo returns the path to the ossinfo file.
func OssInfo() string {
	return k3.OssInfo
}

// Compiler returns the path to the TTCN-3 compiler. Compiler will return "mtc"
// if no compiler is found.
func Compiler() string {
	return k3.Compiler
}

// Runtime returns the path to the TTCN-3 runtime. Runtime will return "k3r" if
// no runtime is found.
func Runtime() string {
	return k3.Runtime
}

func CLibDirs() []string {
	return k3.CLibDirs
}

// Plugins returns a list of k3 plugins.
func Plugins() []string {
	return k3.Plugins
}

// Includes returns a list of TTCN-3 include directories required by the k3 compiler.
func Includes() []string {
	return k3.Includes
}

// NewASN1Codec returns the commands required to compile ASN.1 files.
func NewASN1Codec(vars map[string]string, name string, encoding string, srcs ...string) []*proc.Cmd {
	if vars == nil {
		vars = make(map[string]string)
		for k, v := range DefaultEnv {
			vars[k] = v
		}
	}

	if _, ok := vars["OSSINFO"]; !ok && k3.OssInfo != "" {
		vars["OSSINFO"] = k3.OssInfo
	}

	var cFlags []string
	if f, ok := vars["CFLAGS"]; ok {
		cFlags = append(cFlags, f)
	}
	for _, inc := range k3.CIncludes {
		cFlags = append(cFlags, "-I"+inc)
	}
	vars["CFLAGS"] = strings.Join(cFlags, " ")

	var ldFlags []string
	if f, ok := vars["LDFLAGS"]; ok {
		ldFlags = append(ldFlags, f)
	}
	for _, dir := range k3.CLibDirs {
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
	for _, inc := range k3.CIncludes {
		cFlags = append(cFlags, "-I"+inc)
	}
	vars["CFLAGS"] = strings.Join(cFlags, " ")

	var ldFlags []string
	if f, ok := vars["LDFLAGS"]; ok {
		ldFlags = append(ldFlags, f)
	}
	for _, dir := range k3.CLibDirs {
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
		vars["K3C"] = k3.Compiler
	}

	// We need to remove k3 stdlib files from the source list, (if accidentally
	// inserted by the user) because of a missing module (PCMDmod).
	vars["_sources"] = strings.Join(removeStdlib(srcs), " ")

	for _, dir := range k3.Includes {
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
