package main

import (
	"strconv"

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

	suite *ntt.Suite
}

func NewReport(args []string) *Report {
	r := Report{Args: args}
	r.suite, r.Err = ntt.NewFromArgs(args...)

	if r.Err == nil {
		r.Name, r.Err = r.suite.Name()
	}

	if r.Err == nil {
		r.Timeout, r.Err = r.suite.Timeout()
	}

	if r.Err == nil {
		r.ParametersFile, r.Err = path(r.suite.ParametersFile())
	}

	if r.Err == nil {
		r.TestHook, r.Err = path(r.suite.TestHook())
	}

	if r.Err == nil {
		r.DataDir, r.Err = r.suite.Getenv("NTT_DATADIR")
	}

	if r.Err == nil {
		if env, err := r.suite.Getenv("NTT_SESSION_ID"); err != nil {
			r.SessionID, r.Err = strconv.Atoi(env)
		}
	}

	if r.Err == nil {
		r.Environ, r.Err = r.suite.Environ()
	}

	if r.Err == nil {
		paths, err := r.suite.Sources()
		r.Sources, r.Err = ntt.PathSlice(paths...), err
	}

	if r.Err == nil {
		paths, err := r.suite.Imports()
		r.Imports, r.Err = ntt.PathSlice(paths...), err
	}

	if r.Err == nil {
		r.Files, r.Err = r.suite.Files()
	}

	if root := r.suite.Root(); root != nil {
		r.SourceDir = root.Path()
	}

	r.AuxFiles = ntt.FindAuxiliaryTTCN3Files()

	return &r
}

func path(f *ntt.File, err error) (string, error) {
	return f.Path(), err
}
