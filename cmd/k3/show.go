package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/nokia/ntt/internal/config"
	"github.com/spf13/cobra"
)

var (
	showCmd = &cobra.Command{
		Use:   "show [var...] [-- <file>...]",
		Short: "show configuration variables used by k3",
		RunE:  show,
	}

	vars    = map[string]string{}
	allKeys = []string{
		"name",
		"source_dir",
		"sources",
		"imports",
		"ttcn3_files",
		"parameters_file",
		"test_hook",
		"timeout",
	}
)

func show(cmd *cobra.Command, args []string) error {

	keys, files := splitArgs(args, cmd.ArgsLenAtDash())

	conf, err := config.FromArgs(files)
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error(), '\n')
		return nil
	}

	if conf != nil {
		vars["name"] = conf.Name
		vars["source_dir"] = conf.Dir
		vars["sources"] = strings.Join(conf.Sources, " ")
		vars["imports"] = strings.Join(conf.Imports, " ")
		vars["parameters_file"] = conf.ParametersFile
		vars["test_hook"] = conf.TestHook
		vars["timeout"] = strconv.FormatFloat(conf.Timeout, 'f', -1, 64)
		ttcn3_files := findTTCN3Files(conf.Sources...)
		if conf.Imports != nil {
			ttcn3_files = append(ttcn3_files, findTTCN3Files(conf.Imports...)...)
		}
		vars["ttcn3_files"] = strings.Join(ttcn3_files, " ")
	}

	// TODO(5nord) find better default for keys if conf is nil
	if len(keys) == 0 {
		keys = allKeys
	}

	for _, k := range keys {
		if v, ok := vars[k]; ok {
			// TODO(5nord) only print variable name is default keys are used
			fmt.Printf("K3_%s=\"%s\"\n", strings.ToUpper(k), v)
		}
	}
	return nil
}

func splitArgs(args []string, pos int) ([]string, []string) {
	if pos < 0 {
		return args, []string{}
	}
	return args[:pos], args[pos:]
}

func findTTCN3Files(dirs ...string) []string {
	ret := make([]string, len(dirs))
	for _, d := range dirs {
		// TODO(5nord) make paths valid (relativ to config.Dir)
		if files, err := config.FindTTCN3Files(d); err == nil {
			ret = append(ret, files...)
		}
	}
	return ret
}
