package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/template"

	"github.com/nokia/ntt/internal/ntt"
	"github.com/spf13/cobra"
)

var (
	showCmd = &cobra.Command{
		Use:   "show [ <file>...] [-- var...]",
		Short: "show configuration variables used by ntt",
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
	case JSON:
		return printJSON(report, keys)
	case ShSetup:
		return printShellScript(report, keys)
	case len(keys) != 0:
		return printUserKeys(suite, keys)
	default:
		return printDefaultKeys(suite)
	}
}

func printJSON(report *Report, keys []string) error {
	if len(keys) != 0 {
		return fmt.Errorf("command line option --json does not accept additional command line arguments")
	}

	b, err := json.MarshalIndent(report, "", "  ")
	fmt.Println(string(b))
	return err
}

func printShellScript(report *Report, keys []string) error {
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
