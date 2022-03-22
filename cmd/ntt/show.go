package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"text/template"

	"github.com/nokia/ntt/internal/env"
	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/ntt"
	"github.com/nokia/ntt/project"
	"github.com/spf13/cobra"
)

var (
	ShowCommand = &cobra.Command{
		Use:   "show [ <file>...] [-- var...]",
		Short: "Show test suite configuration.",
		RunE:  show,
	}
)

func show(cmd *cobra.Command, args []string) error {
	sources, keys := splitArgs(args, cmd.ArgsLenAtDash())
	suite, err := ntt.NewFromArgs(sources...)
	if err != nil {
		return err
	}

	c, err := suite.Config()

	r := ConfigReport{
		Args:   args,
		Config: c,
		err:    err,
	}

	checkErr := func(err error) {
		if err != nil && r.err == nil {
			r.err = err
		}
	}

	r.DataDir = env.Getenv("NTT_DATADIR")

	r.Environ, err = suite.Environ()
	sort.Strings(r.Environ)
	checkErr(err)

	r.Files, err = project.Files(suite)
	checkErr(err)

	for _, dir := range c.K3.Builtins {
		r.AuxFiles = append(r.AuxFiles, fs.FindTTCN3Files(dir)...)
	}

	switch {
	case outputJSON:
		return printJSON(&r, keys)
	case ShSetup:
		return printShellScript(&r, keys)
	case len(keys) != 0:
		return printValues(c, keys)
	default:
		keys := []string{
			"name",
			"root",
			"source_dir",
			"sources",
			"imports",
			"timeout",
			"parameters_dir",
			"parameters_file",
			"test_hook",
			"datadir",
			"session_id",
		}
		return printKeyValues(c, keys)
	}
}

func printJSON(report *ConfigReport, keys []string) error {
	if len(keys) != 0 {
		return fmt.Errorf("command line option --json does not accept additional command line arguments")
	}

	b, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}
	fmt.Println(string(b))
	return report.err
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
{{ if .HooksFile      -}} export K3_TEST_HOOK='{{ .HooksFile }}'            {{- end }}
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

func get(c *project.Config, key string) ([]string, error) {
	var (
		v   interface{}
		err error
	)

	if key == "env" {
		v = c.Variables
	} else {
		v, err = c.Get(key)
		if err != nil {
			return nil, err
		}
	}

	switch reflect.TypeOf(v).Kind() {
	case reflect.Slice:
		return v.([]string), nil
	case reflect.Map:
		s := make([]string, 0, len(v.(map[string]string)))
		for k, v := range v.(map[string]string) {
			s = append(s, fmt.Sprintf("'%s=%s'", k, v))
		}
		sort.Strings(s)
		return s, nil
	default:
		return []string{fmt.Sprint(v)}, nil
	}
}

func printValues(c *project.Config, keys []string) error {
	for _, key := range keys {
		s, err := get(c, key)
		if err != nil {
			return err
		}
		for _, v := range s {
			fmt.Println(v)
		}
	}
	return nil
}

func printKeyValues(c *project.Config, keys []string) error {
	for _, key := range keys {
		s, err := get(c, key)
		if err != nil {
			return err
		}
		fmt.Printf("NTT_%s=\"%s\"\n", strings.ToUpper(key), strings.Join(s, " "))
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
	Args []string `json:"args"`
	*project.Config
	SourceDir string   `json:"source_dir"`
	DataDir   string   `json:"datadir"`
	Environ   []string `json:"env"`
	Files     []string `json:"files"`
	AuxFiles  []string `json:"aux_files"`
	err       error
}

func (r *ConfigReport) Err() string {
	if r.err != nil {
		return r.err.Error()
	}
	return ""
}
