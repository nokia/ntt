// Package k3 provides convenience functions for supporting k3 toolchain. k3 is
// the predecessor of the ntt project.
package k3

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Root returns the directory where K3 toolset is installed, by first looking
// at environment variable K3ROOT and then by probing for a installed K3
// runtime engine.
func Root() string {
	if env := os.Getenv("K3ROOT"); env != "" {
		return env
	}

	if env := os.Getenv("NTTROOT"); env != "" {
		return env
	}

	// When k3r is installed, we assume its plugins are installed as well.
	if k3r, err := exec.LookPath("k3r"); err == nil {
		if path := parentDir(k3r); path != "" {
			return path
		}
	}

	return ""
}

// PluginDir returns the directory where k3 plugins are installed.
func PluginDir() string {
	root := Root()
	if root == "" {
		return ""
	}

	hints := []string{
		"lib/k3/plugins",
		"lib64/k3/plugins",
		"lib/x86_64/k3/plugins",
	}
	for _, hint := range hints {
		if dir := filepath.Join(root, hint); isDir(dir) {
			return dir
		}
	}
	return ""
}

func DataDir() string {
	if env := os.Getenv("K3_DATADIR"); env != "" {
		return env
	}
	if env := os.Getenv("NTT_DATADIR"); env != "" {
		return env
	}

	root := Root()
	if root == "" {
		return ""
	}

	if dir := filepath.Join(root, "share/k3"); isDir(dir) {
		return dir
	}
	return ""
}

// FindAuxiliaryDirectories returns a list of auxiliary k3 directories containing TTCN-3 files.
func FindAuxiliaryDirectories() []string {

	// If are using a runtime from out developer source tree, we should
	// also use files and adapter from that tree as well.
	if hint := filepath.Dir(Runtime()); strings.HasSuffix(hint, "/src/k3r") {
		dir := collectFolders(
			hint+"/../k3r-*-plugin",
			hint+"/../../../src/k3r-*-plugin",
			hint+"/../../../src/ttcn3",
			hint+"/../../../src/libzmq",
		)
		if len(dir) > 0 {
			return dir
		}

	}

	auxDirs := []string{
		PluginDir(),
		DataDir(),
	}

	var ret []string
	for _, dir := range auxDirs {
		if ttcn3Dir := filepath.Join(dir, "ttcn3"); dir != "" && isDir(ttcn3Dir) {
			ret = append(ret, ttcn3Dir)
		}
	}
	return ret
}

func collectFolders(globs ...string) []string {
	return removeDuplicates(filterFolders(evalSymlinks(resolveGlobs(globs))))
}

func resolveGlobs(globs []string) []string {
	var ret []string

	for _, g := range globs {
		if matches, err := filepath.Glob(g); err == nil {
			ret = append(ret, matches...)
		}
	}
	return ret
}

func evalSymlinks(links []string) []string {
	var ret []string
	for _, l := range links {
		if path, err := filepath.EvalSymlinks(l); err == nil {
			ret = append(ret, path)
		}
	}
	return ret
}

func filterFolders(paths []string) []string {
	var ret []string
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}

		if info.IsDir() {
			ret = append(ret, path)
		}
	}
	return ret
}

func removeDuplicates(slice []string) []string {
	var ret []string
	h := make(map[string]bool)
	for _, s := range slice {
		if _, v := h[s]; !v {
			h[s] = true
			ret = append(ret, s)
		}
	}
	return ret
}

func Compiler() string {
	return findK3Tool("mtc", "k3c", "k3c.exe")
}

func Runtime() string {
	return findK3Tool("k3r", "k3r.exe")
}

func findK3Tool(names ...string) string {
	if len(names) == 0 {
		return ""
	}
	for _, name := range names {
		if env := os.Getenv(strings.ToUpper(name)); env != "" {
			return env
		}
		if root := Root(); root != "" {
			if exe, err := exec.LookPath(filepath.Join(root, name)); err == nil {
				return exe
			}
		}
		if exe, err := exec.LookPath(name); err == nil {
			return exe
		}
	}
	return names[0]
}
func parentDir(path string) string {
	dir, _ := filepath.Abs(filepath.Join(filepath.Dir(path), ".."))
	return dir
}

func isDir(path string) bool {
	if info, err := os.Stat(path); err == nil {
		return info.IsDir()
	}
	return false
}
