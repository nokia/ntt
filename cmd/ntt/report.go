package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/k3"
	"github.com/nokia/ntt/internal/ntt"
)

type Report struct {
	Args           []string `json:"args"`
	Err            error    `json:"error"`
	Name           string   `json:"name"`
	Timeout        float64  `json:"timeout"`
	ParametersFile string   `json:"parameters_file"`
	TestHook       string   `json:"test_hook"`
	SourceDir      string   `json:"source_dir"`
	DataDir        string   `json:"datadir"`
	SessionID      int      `json:"session_id"`
	Environ        []string `json:"env"`
	Sources        []string `json:"sources"`
	Imports        []string `json:"imports"`
	Files          []string `json:"files"`
	AuxFiles       []string `json:"aux_files"`
	OssInfo        string   `json:"ossinfo"`
	K3             struct {
		Compiler string   `json:"compiler"`
		Runtime  string   `json:"runtime"`
		Builtins []string `json:"builtins"`
	} `json:"k3"`

	suite *ntt.Suite
}

func NewReport(args []string) *Report {
	var err error = nil
	r := Report{Args: args}
	r.suite, r.Err = ntt.NewFromArgs(args...)

	if r.Err == nil {
		r.Name, r.Err = r.suite.Name()
	}

	if r.Err == nil {
		r.Timeout, r.Err = r.suite.Timeout()
	}

	r.ParametersFile, err = path(r.suite.ParametersFile())

	if (r.Err == nil) && (err != nil) {
		r.Err = err
	}
	r.TestHook, err = path(r.suite.TestHook())
	if (r.Err == nil) && (err != nil) {
		r.Err = err
	}

	r.DataDir, err = r.suite.Getenv("NTT_DATADIR")
	if (r.Err == nil) && (err != nil) {
		r.Err = err
	}

	if env, err := r.suite.Getenv("NTT_SESSION_ID"); err == nil {
		r.SessionID, err = strconv.Atoi(env)
		if (r.Err == nil) && (err != nil) {
			r.Err = err
		}
	} else {
		if r.Err == nil {
			r.Err = err
		}
	}

	r.Environ, err = r.suite.Environ()
	if err == nil {
		sort.Strings(r.Environ)
	}
	if (r.Err == nil) && (err != nil) {
		r.Err = err
	}

	{
		paths, err := r.suite.Sources()
		r.Sources = fs.PathSlice(paths...)
		if (r.Err == nil) && (err != nil) {
			r.Err = err
		}
	}

	{
		paths, err := r.suite.Imports()
		r.Imports = fs.PathSlice(paths...)
		if (r.Err == nil) && (err != nil) {
			r.Err = err
		}
	}

	r.Files, err = r.suite.Files()
	if (r.Err == nil) && (err != nil) {
		r.Err = err
	}

	if root := r.suite.Root(); root != nil {
		r.SourceDir = root.Path()
		if path, err := filepath.Abs(r.SourceDir); err == nil {
			r.SourceDir = path
		} else if r.Err == nil {
			r.Err = err
		}
	}

	for _, dir := range k3.FindAuxiliaryDirectories() {
		r.AuxFiles = append(r.AuxFiles, fs.FindTTCN3Files(dir)...)
	}

	r.K3.Compiler = findK3Tool(r.suite, "mtc", "k3c")
	r.K3.Runtime = findK3Tool(r.suite, "k3r")

	r.OssInfo, _ = r.suite.Getenv("OSSINFO")
	hint := filepath.Dir(r.K3.Runtime)
	switch {
	// Probably a regular K3 installation. We assume datadir and libdir are
	// in a sibling folder.
	case strings.HasSuffix(hint, "/bin"):
		r.K3.Builtins = collectFolders(
			hint+"/../lib*/k3/plugins",
			hint+"/../lib*/k3/plugins/ttcn3",
			hint+"/../lib/*/k3/plugins",
			hint+"/../lib/*/k3/plugins/ttcn3",
			hint+"/../share/k3/ttcn3",
		)
		if r.OssInfo == "" {
			r.OssInfo = filepath.Clean(hint + "/../share/k3/asn1")
		}

	// If the runtime seems to be a buildtree of our source repository, we
	// assume the builtins are there as well.
	case strings.HasSuffix(hint, "/src/k3r"):
		// TODO(5nord) the last glob fails if CMAKE_BUILD_DIR is not
		// beneath CMAKE_SOURCE_DIR. Find a way to locate the source
		// dir correctly.
		srcdir := hint + "/../../.."

		r.K3.Builtins = collectFolders(
			hint+"/../k3r-*-plugin",
			srcdir+"/src/k3r-*-plugin",
			srcdir+"/src/ttcn3",
			srcdir+"/src/libzmq",
		)
	}

	return &r
}

func findK3Tool(suite *ntt.Suite, names ...string) string {
	for _, name := range names {
		if env, _ := suite.Getenv(strings.ToUpper(name)); env != "" {
			name = env
		}
		if path, err := exec.LookPath(name); err == nil {
			return path
		}
	}
	return ""
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

func path(f *fs.File, err error) (string, error) {
	if f == nil {
		return "", err
	}

	return f.Path(), err
}
