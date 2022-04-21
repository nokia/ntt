// Package k3 provides convenience functions for supporting k3 toolchain. k3 is
// the predecessor of the ntt project.
package k3

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/log"
)

// InstallDir returns the directory where k3 is probably installed.
func InstallDir() string {
	if env := os.Getenv("NTTROOT"); env != "" {
		return env
	}
	if env := os.Getenv("K3ROOT"); env != "" {
		return env
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
	if env := os.Getenv("NTT_DATADIR"); env != "" {
		return env
	}
	if env := os.Getenv("K3_DATADIR"); env != "" {
		return env
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
	cache, err := findCMakeCache()
	if err != nil {
		log.Verboseln("k3: cmake:", err.Error())
		return ""
	}

	f, err := os.Open(cache)
	if err != nil {
		if os.IsNotExist(err) {
			// CMakeCache.txt is optional
			return ""
		}
		log.Verboseln("k3: cmake:", err.Error())
		return ""
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		if v := strings.SplitN(s.Text(), ":", 2); v[0] == name {
			w := strings.SplitN(v[1], "=", 2)
			return w[1]
		}
	}
	return ""
}

// findCMakeCache returns the path to the CMakeCache.txt file, by walking up
// the current working directory and the file hierarchy specified by
// environment variable K3R
func findCMakeCache() (string, error) {
	find := func(path string) string {
		var dir string
		fs.WalkUp(path, func(path string) bool {
			if fs.IsRegular(filepath.Join(path, "CMakeCache.txt")) {
				dir = path
			}
			return true
		})
		return dir
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if dir := find(cwd); dir != "" {
		return dir, nil
	}
	if k3r := os.Getenv("K3R"); strings.HasSuffix(k3r, "src/k3r/k3r") {
		return find(filepath.Dir(k3r)), nil
	}

	return "", nil
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
