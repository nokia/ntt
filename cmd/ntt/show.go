package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/ntt"
	"github.com/nokia/ntt/k3"
	"github.com/nokia/ntt/project"
	"github.com/spf13/cobra"
)

var (
	showCmd = &cobra.Command{
		Use:   "show [ <file>...] [-- var...]",
		Short: "Show test suite configuration.",
		RunE:  show,
	}

	stringers = map[string]func(suite *ntt.Suite) string{
		"name": func(suite *ntt.Suite) string {
			s, err := suite.Name()
			if err != nil {
				fatal(err)
			}
			return s
		},

		"sources": func(suite *ntt.Suite) string {
			srcs, err := suite.Sources()
			if err != nil {
				fatal(err)
			}
			return strings.Join(srcs, "\n")
		},

		"imports": func(suite *ntt.Suite) string {
			imps, err := suite.Imports()
			if err != nil {
				fatal(err)
			}
			return strings.Join(imps, "\n")
		},
		"timeout": func(suite *ntt.Suite) string {
			t, err := suite.Timeout()
			if err != nil {
				fatal(err)
			}
			if t > 0 {
				return fmt.Sprint(t)
			}
			return ""

		},

		"parameters_dir": func(suite *ntt.Suite) string {
			d, err := suite.ParametersDir()
			if err != nil {
				fatal(err)
			}
			return d
		},

		"parameters_file": func(suite *ntt.Suite) string {
			f, err := suite.ParametersFile()
			if err != nil {
				fatal(err)
			}
			if f != nil {
				return f.Path()
			}
			return ""
		},

		"test_hook": func(suite *ntt.Suite) string {
			f, err := suite.TestHook()
			if err != nil {
				fatal(err)
			}
			if f != nil {
				return f.Path()
			}
			return ""
		},

		"source_dir": func(suite *ntt.Suite) string {
			if root := suite.Root(); root != "" {
				return root
			}
			return ""
		},

		"datadir": func(suite *ntt.Suite) string {
			s, err := suite.Getenv("NTT_DATADIR")
			if err != nil {
				fatal(err)
			}
			return s
		},

		"session_id": func(suite *ntt.Suite) string {
			s, err := suite.Getenv("NTT_SESSION_ID")
			if err != nil {
				fatal(err)
			}
			return s
		},

		"env": func(suite *ntt.Suite) string {
			env, err := suite.Environ()
			if err != nil {
				fatal(err)
			}
			sort.Strings(env)
			res := make([]string, 0, len(env))
			for _, e := range env {
				res = append(res, fmt.Sprintf(`'%s'`, e))
			}
			return strings.Join(res, "\n")
		},
	}
)

func show(cmd *cobra.Command, args []string) error {

	sources, keys := splitArgs(args, cmd.ArgsLenAtDash())

	suite, err := ntt.NewFromArgs(sources...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	report := NewReport(sources)

	switch {
	case outputJSON:
		return printJSON(report, keys)
	case ShSetup:
		return printShellScript(report, keys)
	case len(keys) != 0:
		return printUserKeys(suite, keys)
	default:
		return printDefaultKeys(suite)
	}
}

func printJSON(report *ConfigReport, keys []string) error {
	if len(keys) != 0 {
		return fmt.Errorf("command line option --json does not accept additional command line arguments")
	}

	b, err := json.MarshalIndent(report, "", "  ")
	fmt.Println(string(b))
	return err
}

func printShellScript(report *ConfigReport, keys []string) error {
	const shellTemplate = `# This is a generated output of ntt show. Args: {{ .Args }}

# k3-hook calls the K3 test hook (if defined) with action passed by $1.
function k3-hook()
{
    if [ -n "$K3_TEST_HOOK" ]; then
        K3_SOURCES="${K3_SOURCES[*]}" \
        K3_IMPORTS="${K3_IMPORTS[*]}" \
        K3_TTCN3_FILES="${K3_TTCN3_FILES[*]}" \
            "$K3_TEST_HOOK" "$@" 1>&2
    fi
}

{{ if .Name           -}} export K3_NAME='{{ .Name }}'                      {{- end }}
{{ if gt .Timeout 0.0 -}} export K3_TIMEOUT='{{ .Timeout }}'                {{- end }}
{{ if .ParametersDir  -}} export K3_PARAMETERS_DIR='{{ .ParametersDir }}'   {{- end }}
{{ if .ParametersFile -}} export K3_PARAMETERS_FILE='{{ .ParametersFile }}' {{- end }}
{{ if .TestHook       -}} export K3_TEST_HOOK='{{ .TestHook }}'             {{- end }}
{{ if .SourceDir      -}} export K3_SOURCE_DIR='{{ .SourceDir }}'           {{- end }}
{{ if .DataDir        -}} export K3_DATADIR='{{ .DataDir }}'                {{- end }}
{{ if .SessionID      -}} export K3_SESSION_ID='{{ .SessionID }}'           {{- end }}

{{ if .K3.Compiler    -}} export K3C='{{ .K3.Compiler }}'           {{- end }}
{{ if .K3.Runtime     -}} export K3R='{{ .K3.Runtime  }}'           {{- end }}
{{ if .OssInfo        -}} export OSSINFO='{{ .OssInfo }}'           {{- end }}

{{ range .Environ }}export '{{.}}'
{{end}}

K3_SOURCES=(
{{ range .Sources }}	{{.}}
{{end}})

K3_IMPORTS=(
{{ range .Imports }}	{{.}}
{{end}})

K3_TTCN3_FILES=(
{{ range .Files }}	{{.}}
{{end}}
	# Auxiliary files from K3
{{ range .AuxFiles }}	{{.}}
{{end}})

K3_BUILTINS=(
{{ range .K3.Builtins }}	{{.}}
{{end}})

{{ if .Err }}
# ERROR
#
# Output might not be complete, because some errors have occurred during
# execution. We return "false", to give you the chance to detect this
# situation
read -r -d '' K3_ERROR <<'EOF'
{{.Err}}
EOF
false
{{ end }}
`
	if len(keys) != 0 {
		return fmt.Errorf("command line option --sh does not accept additional command line arguments")
	}

	t := template.Must(template.New("k3-sh-setup").Parse(shellTemplate))
	if err := t.Execute(os.Stdout, report); err != nil {
		fmt.Printf(`
# ERROR: Internal template did not compile: %s
#
# Output might not be complete, because some errors have occurred during
# execution. We return "false", to give you the chance to detect this
# situation
false
`, err.Error())
		return err
	}

	if err := report.Err(); err != "" {
		return fmt.Errorf("%s", err)
	}
	return nil
}

func printUserKeys(suite *ntt.Suite, keys []string) error {

	for _, key := range keys {
		if fun, found := stringers[key]; found {
			if s := fun(suite); s != "" {
				fmt.Println(s)
			}
			continue
		}

		s, err := suite.Getenv(key)
		if err != nil {
			return err
		}

		fmt.Println(s)
	}
	return nil
}

func printDefaultKeys(suite *ntt.Suite) error {
	for _, key := range []string{
		"name",
		"sources",
		"imports",
		"timeout",
		"parameters_dir",
		"parameters_file",
		"test_hook",
		"source_dir",
		"datadir",
		"session_id",
	} {
		if s := stringers[key](suite); s != "" {
			fmt.Printf("K3_%s=\"%s\"\n", strings.ToUpper(key), strings.Replace(s, "\n", " ", -1))
		}
	}

	return nil
}

// splitArgs splits an argument list at pos. Pos is usually the position of '--'
// (see cobra.Command.ArgsLenAtDash).
//
// Is pos < 0, the second list will be empty
func splitArgs(args []string, pos int) ([]string, []string) {
	if pos < 0 {
		return args, []string{}
	}
	return args[:pos], args[pos:]
}

type ConfigReport struct {
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

func (r *ConfigReport) Err() string {
	if r.err != nil {
		return r.err.Error()
	}
	return ""
}

func NewReport(args []string) *ConfigReport {
	var err error = nil
	r := ConfigReport{Args: args}
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
	r.K3.Builtins = k3.FindAuxiliaryDirectories()
	r.OssInfo, _ = r.suite.Getenv("OSSINFO")
	return &r
}

func path(f *fs.File, err error) (string, error) {
	if f == nil {
		return "", err
	}

	return f.Path(), err
}
