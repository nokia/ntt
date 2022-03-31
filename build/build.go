// Package build provides a simple interface and helper functions to build packages.
package build

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/nokia/ntt/internal/cache"
	"github.com/nokia/ntt/internal/env"
)

// Builder is the interface for building.
type Builder interface {
	// Build builds a target.
	Build() error

	// It returns the path to the output files.
	Targets() []string
}

// Pathf is like fmt.Sprintf but searches the NTT_CACHE environment variable first.
func Pathf(f string, v ...interface{}) string {
	return cache.Lookup(fmt.Sprintf(f, v...))
}

// FieldsExpand first expands environment variables and then splits the string.
// Empty fields are removed.
func FieldsExpand(s string) []string {
	return FieldsExpandWithDefault(s, nil)
}

// FieldsExpandWithDefault works like FieldsExpand, but uses default values
// defined in the map, if environment variable is not set.
func FieldsExpandWithDefault(s string, defaults map[string]string) []string {
	expand := func(s string) string {
		if s, ok := env.LookupEnv(s); ok {
			return s
		}
		return defaults[s]
	}
	var fields []string
	for _, f := range strings.Fields(os.Expand(s, expand)) {
		if f != "" {
			fields = append(fields, f)
		}
	}
	return fields
}

// Command expands given arguments and returns a command. Stdout and Stderr, as
// well as the environment, are inherited from the parent process.
func Command(args ...string) *exec.Cmd {
	return CommandWithEnv(nil, args...)
}

// CommandWithEnv expands given arguments and returns a command. Stdout and Stderr, as
// well as the environment, are inherited from the parent process.
func CommandWithEnv(env map[string]string, args ...string) *exec.Cmd {
	var cmdArgs []string
	for _, arg := range args {
		cmdArgs = append(cmdArgs, FieldsExpandWithDefault(arg, env)...)
	}
	if len(cmdArgs) == 0 {
		cmdArgs = []string{""}
	}
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Env = append(cmd.Env, os.Environ()...)
	return cmd
}

// NeedsRebuild first expands environment variables like $FOO or ${FOO}, and
// then report is any of the source files is newer than the target.
//
// If the target does not exist, it is considered to need to be rebuilt. It's
// an error if any of the sources does not exist.
func NeedsRebuild(target string, sources ...string) (bool, error) {
	stat, err := os.Stat(os.ExpandEnv(target))
	if os.IsNotExist(err) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	for _, src := range sources {
		src = os.ExpandEnv(src)
		sstat, err := os.Stat(src)
		if err != nil {
			return false, err
		}
		if sstat.ModTime().After(stat.ModTime()) {
			return true, nil
		}
	}
	return false, nil
}
