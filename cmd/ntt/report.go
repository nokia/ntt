package main

import (
	"path/filepath"
	"sort"
	"strconv"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/ntt"
	"github.com/nokia/ntt/k3"
	"github.com/nokia/ntt/project"
)

type Report struct {
	Args           []string `json:"args"`
	Name           string   `json:"name"`
	Timeout        float64  `json:"timeout"`
	ParametersDir  string   `json:"parameters_dir"`
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
	err   error
}

func (r *Report) Err() string {
	if r.err != nil {
		return r.err.Error()
	}
	return ""
}

func NewReport(args []string) *Report {
	var err error = nil
	r := Report{Args: args}
	r.suite, r.err = ntt.NewFromArgs(args...)

	if r.err == nil {
		r.Name, r.err = r.suite.Name()
	}

	if r.err == nil {
		r.Timeout, r.err = r.suite.Timeout()
	}

	r.ParametersDir, err = r.suite.ParametersDir()

	if (r.err == nil) && (err != nil) {
		r.err = err
	}

	r.ParametersFile, err = path(r.suite.ParametersFile())

	if (r.err == nil) && (err != nil) {
		r.err = err
	}
	r.TestHook, err = path(r.suite.TestHook())
	if (r.err == nil) && (err != nil) {
		r.err = err
	}

	r.DataDir, err = r.suite.Getenv("NTT_DATADIR")
	if (r.err == nil) && (err != nil) {
		r.err = err
	}

	if env, err := r.suite.Getenv("NTT_SESSION_ID"); err == nil {
		r.SessionID, err = strconv.Atoi(env)
		if (r.err == nil) && (err != nil) {
			r.err = err
		}
	} else {
		if r.err == nil {
			r.err = err
		}
	}

	r.Environ, err = r.suite.Environ()
	if err == nil {
		sort.Strings(r.Environ)
	}
	if (r.err == nil) && (err != nil) {
		r.err = err
	}

	{
		paths, err := r.suite.Sources()
		r.Sources = paths
		if (r.err == nil) && (err != nil) {
			r.err = err
		}
	}

	{
		paths, err := r.suite.Imports()
		r.Imports = paths
		if (r.err == nil) && (err != nil) {
			r.err = err
		}
	}

	r.Files, err = project.Files(r.suite)
	if (r.err == nil) && (err != nil) {
		r.err = err
	}

	if root := r.suite.Root(); root != "" {
		r.SourceDir = root
		if path, err := filepath.Abs(r.SourceDir); err == nil {
			r.SourceDir = path
		} else if r.err == nil {
			r.err = err
		}
	}

	for _, dir := range k3.FindAuxiliaryDirectories() {
		r.AuxFiles = append(r.AuxFiles, fs.FindTTCN3Files(dir)...)
	}

	r.K3.Compiler = k3.Compiler()
	r.K3.Runtime = k3.Runtime()
	r.K3.Builtins = k3.BuiltinDirectories()

	r.OssInfo, _ = r.suite.Getenv("OSSINFO")
	return &r
}

func path(f *fs.File, err error) (string, error) {
	if f == nil {
		return "", err
	}

	return f.Path(), err
}
