// Package k3 provides convenience functions for supporting k3 toolchain. k3 is
// the predecessor of the ntt project.
package k3

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/nokia/ntt/internal/cache"
	"github.com/nokia/ntt/internal/env"
	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/proc"
	"github.com/nokia/ntt/internal/session"
)

// DefaultEnv is the default environment for k3-based test suites.
var DefaultEnv = map[string]string{
	"CXX":        "g++",
	"CC":         "gcc",
	"ASN1C":      "asn1",
	"ASN1CFLAGS": "-reservedWords ffs -c -charIntegers -listingFile -messageFormat emacs -noDefines -valuerefs -debug -root -soed",
	"ASN2TTCN":   "asn1tottcn3",
	"OSSINFO":    filepath.Join(DataDir(), "asn1"),
	"K3C":        Compiler(),
	"K3R":        Runtime(),
}

// InstallDir returns the directory where k3 is probably installed.
func InstallDir() string {
	if s := os.Getenv("NTTROOT"); s != "" {
		return s
	}
	if s := os.Getenv("K3ROOT"); s != "" {
		return s
	}
	if k3r, err := exec.LookPath("k3r"); err == nil {
		if path, _ := filepath.Abs(filepath.Join(filepath.Dir(k3r), "..")); path != "" {
			return path
		}
	}
	return ""
}

// DataDir returns the directory where additional files are installed (/usr/share/k3).
func DataDir() string {
	if s := env.Getenv("NTT_DATADIR"); s != "" {
		return s
	}
	if dir := filepath.Join(InstallDir(), "share/k3"); fs.IsDir(dir) {
		return dir
	}
	return ""
}

// Compiler returns the path to the TTCN-3 compiler. Compiler will return "mtc"
// if no compiler is found.
func Compiler() string {
	if k3c := os.Getenv("K3C"); k3c != "" {
		return k3c
	}
	if mtc := os.Getenv("MTC"); mtc != "" {
		return mtc
	}
	if mtc := filepath.Join(cmake("mtc_BINARY_DIR"), "source/mtc/mtc"); fs.IsRegular(mtc) {
		return mtc
	}
	return findK3Tool("mtc", "k3c", "k3c.exe")
}

// Runtime returns the path to the TTCN-3 runtime. Runtime will return "k3r" if
// no runtime is found.
func Runtime() string {
	if k3r := os.Getenv("K3R"); k3r != "" {
		return k3r
	}
	return findK3Tool("k3r", "k3r.exe")
}

// Plugins returns a list of k3 plugins.
func Plugins() []string {
	if dirs := k3DevelDirs(); len(dirs) > 0 {
		return dirs
	}
	hints := []string{
		"lib/k3/plugins",
		"lib64/k3/plugins",
		"lib/x86_64/k3/plugins",
	}
	for _, hint := range hints {
		if dir := filepath.Join(InstallDir(), hint); fs.IsDir(dir) {
			return []string{dir}
		}
	}
	return nil
}

// Includes returns a list of TTCN-3 include directories required by the k3 compiler.
func Includes() []string {
	auxDirs := Plugins()
	if dir := DataDir(); dir != "" {
		auxDirs = append(auxDirs, dir)
	}

	var ret []string
	for _, dir := range auxDirs {
		if len(fs.FindTTCN3Files(dir)) > 0 {
			ret = append(ret, dir)
		}
		if dir := filepath.Join(dir, "ttcn3"); len(fs.FindTTCN3Files(dir)) > 0 {
			ret = append(ret, dir)
		}
	}
	return clean(ret...)
}

// NewASN1Codec returns the commands required to compile ASN.1 files.
func NewASN1Codec(vars map[string]string, name string, encoding string, srcs ...string) []*proc.Cmd {
	if vars == nil {
		vars = make(map[string]string)
		for k, v := range DefaultEnv {
			vars[k] = v
		}
	}
	vars["name"] = name
	vars["encoding"] = encoding

	asn1 := proc.Task("$ASN1C -${encoding} $ASN1CFLAGS -output ${name}.enc -prefix ${name} $OSSINFO/asn1dflt.linux-x86_64 ${srcs}")
	asn1.Env = vars
	asn1.Sources = srcs
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
	p := proc.Task("$CXX $CPPFLAGS $CXXFLAGS -shared -fPIC -o ${tgts} ${srcs} $LDFLAGS $EXTRA_LDFLAGS -lk3-plugin")
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

	// We need to remove k3 stdlib files from the source list, (if accidentally
	// inserted by the user) because of a missing module (PCMDmod).
	vars["_sources"] = strings.Join(removeStdlib(srcs), " ")

	for _, dir := range Includes() {
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

func k3DevelDirs() []string {
	var files []string
	if binary_dir := cmake("k3_BINARY_DIR"); binary_dir != "" {
		files = append(files, glob(filepath.Join(binary_dir, "src/k3r-*-plugin"))...)
	}
	if source_dir := cmake("k3_SOURCE_DIR"); source_dir != "" {
		files = append(files, glob(
			filepath.Join(source_dir, "k3r-*-plugin"),
			filepath.Join(source_dir, "src/k3r-*-plugin"),
			filepath.Join(source_dir, "src/ttcn3"),
			filepath.Join(source_dir, "src/libzmq"),
		)...)
	}
	return clean(files...)
}

// clean resolves symlinks and removes duplicates and non-folders.
func clean(paths ...string) []string {
	m := make(map[string]bool)
	for _, path := range paths {
		path, err := filepath.EvalSymlinks(path)
		if err != nil {
			log.Verboseln("k3: clean:", err.Error())
			return nil
		}
		info, err := os.Stat(path)
		if err != nil {
			log.Verboseln("k3: clean:", err.Error())
			return nil
		}
		if info.IsDir() {
			m[path] = true
		}
	}

	ret := make([]string, 0, len(m))
	for k := range m {
		ret = append(ret, k)
	}
	return ret
}

func glob(globs ...string) []string {
	var ret []string
	for _, g := range globs {
		if matches, err := filepath.Glob(g); err == nil {
			ret = append(ret, matches...)
		}
	}
	return ret
}

// cmake returns a variable with given name from CMakeCache.txt.
func cmake(name string) string {
	cache := findCMakeCache()
	if cache == "" {
		return ""
	}

	f, err := os.Open(cache)
	if err != nil {
		log.Verboseln("k3: cmake:", err.Error())
		return ""
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		if v := strings.SplitN(s.Text(), ":", 2); v[0] == name {
			w := strings.SplitN(v[1], "=", 2)
			log.Debugln("k3: cmake:", name, "=", w[1])
			return w[1]
		}
	}
	return ""
}

// findCMakeCache returns the path to the CMakeCache.txt file, by walking up
// the current working directory and the file hierarchy specified by
// environment variable K3R
func findCMakeCache() string {
	find := func(path string) string {
		var res string
		fs.WalkUp(path, func(path string) bool {
			if file := filepath.Join(path, "CMakeCache.txt"); fs.IsRegular(file) {
				res = file
			}
			return true
		})
		return res
	}
	cwd, err := os.Getwd()
	if err != nil {
		log.Verboseln("k3: cmake:", err.Error())
		return ""
	}
	if f := find(cwd); f != "" {
		return f
	}
	if k3r := os.Getenv("K3R"); strings.HasSuffix(k3r, "src/k3r/k3r") {
		return find(filepath.Dir(k3r))
	}

	return ""
}

func findK3Tool(names ...string) string {
	if len(names) == 0 {
		return ""
	}
	for _, name := range names {
		if env := os.Getenv(strings.ToUpper(name)); env != "" {
			return env
		}
		if root := InstallDir(); root != "" {
			if exe, err := exec.LookPath(filepath.Join(root, "bin", name)); err == nil {
				return exe
			}
		}
		if exe, err := exec.LookPath(name); err == nil {
			return exe
		}
	}
	return names[0]
}

// pathf is like fmt.Sprintf but searches the NTT_CACHE environment variable first.
func pathf(f string, v ...interface{}) string {
	return cache.Lookup(fmt.Sprintf(f, v...))
}

func init() {
	session.SharedDir = "/tmp/k3"
}
