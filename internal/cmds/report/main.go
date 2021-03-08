package report

import (
	"bufio"
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"text/template"

	"github.com/nokia/ntt/internal/ntt"
	"github.com/spf13/cobra"
)

var (
	Command = &cobra.Command{
		Use:   "report",
		Short: "show test suite information",
		Long: `show test suite information
`,
		RunE: report,
	}

	w = bufio.NewWriter(os.Stdout)

	useJSON  = false
	useJUnit = false

	templateText = ""
)

const (
	summaryTemplate = `{{bold}}==================================  Summary  =================================={{off}}
{{range .Tests.NotPassed}}{{ printf "%-10s %s" .Verdict .Name  | colorize }}
{{else}}{{green}}all tests have passed{{off}}
{{end}}
{{len .Tests}} test cases took {{bold}}{{.Tests.Duration}}{{off}} to execute (total runs: {{len .Runs}}
{{- with .Tests.Failed}}, {{red}}not passed: {{len .}}{{off}}{{end}}
{{- with .Tests.Unstable}}, {{orange}}unstable: {{len .}}{{off}}{{end}})
{{bold}}==============================================================================={{off}}

{{ printf "%s (±%s)" .Tests.Average .Tests.Deviation | printf "Average  : %-30s CPU cores      : " }}{{printf "%d" .Cores}}
{{ printf "Shortest : %-30s Parallel tests : %d" .Tests.Shortest.Duration .MaxJobs }}
{{ printf "Longest  : %-30s Load limit     : %d" .Tests.Longest.Duration .MaxLoad}}
{{ printf "Total    : %-30s Load           : %.2f" .Tests.Total .Loads}}

{{bold}}==============================================================================={{off}}
{{bold}}Final Result: {{.Tests.Result | colorize}}{{off}}
{{bold}}==============================================================================={{off}}
`

	junitTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<testsuites>{{range .Modules}}

<testsuite name="{{.Name}}" tests="{{len .Tests}}" failures="{{len .Tests.Failed}}" errors="" time="{{.Tests.Total.Seconds}}">
{{range .Tests}}<testcase name="{{.Testcase}}" time="{{.Duration.Seconds}}">
  {{if ne .Verdict "pass"}}<failure>Verdict: {{.Verdict}} {{with .Reason}}({{. | html }}){{end}}
{{range .ReasonFiles}}{{.Name}}: {{.Content}}{{end}}
  </failure>
{{end}}</testcase>

{{end}}</testsuite>
{{end}}</testsuites>
`

	jsonTemplate = `{
  "name"          : "{{.Name}}",
  "timestamp"     : {{.Runs.First.Begin.Unix}},
  "cores"         : {{.Cores}},
  "parallel_jobs" : {{.MaxJobs}},
  "max_load"      : {{.MaxLoad}},
  "load_average"  : {{ .Loads | json }},
  "suite": {
    "linecount": {{.LineCount}}
  },
  "tests": {
    "result"   : "{{ .Tests.Result }}",
    "tests"    : {{len .Tests }},
    "failed"   : {{len .Tests.Failed}},
    "unstable" : {{.Tests.Unstable | json}},
    "duration" : {
      "real"  : {{.Tests.Duration.Milliseconds}},
      "total" : {{.Tests.Total.Milliseconds}},
      "min"   : {{.Tests.Shortest.Duration.Milliseconds}},
      "max"   : {{.Tests.Longest.Duration.Milliseconds}},
      "avg"   : {{.Tests.Average.Milliseconds}},
      "dev"   : {{.Tests.Deviation.Milliseconds}}
    }
  },
  "env": {{ .Environ | json }}
}
`
)

func report(cmd *cobra.Command, args []string) error {

	suite, err := ntt.NewFromArgs(args...)
	if err != nil {
		return err
	}

	report, err := NewReport(suite)
	if err != nil {
		return err
	}

	switch {
	case useJSON:
		return reportTemplate(report, jsonTemplate)
	case useJUnit:
		return reportTemplate(report, junitTemplate)
	default:
		return reportTemplate(report, templateText)
	}
}

func reportTemplate(report *Report, text string) error {
	tmpl, err := template.New("ntt-report-template").Funcs(funcMap).Parse(text)
	if err != nil {
		return err
	}

	return tmpl.Execute(os.Stdout, report)
}

var funcMap = template.FuncMap{
	"green": func() string {
		return "[32;1m"
	},
	"red": func() string {
		return "[31;1m"
	},
	"orange": func() string {
		return "[38;5;208;1m"
	},
	"bold": func() string {
		return "[1m"
	},
	"off": func() string {
		return "[0m"
	},

	"colorize": func(s string) string {
		pass := regexp.MustCompile(`(?i)pass`)
		fail := regexp.MustCompile(`(?i)fail|error`)
		unstable := regexp.MustCompile(`(?i)unstable`)
		none := regexp.MustCompile(`(?i)\bnone`)
		inconc := regexp.MustCompile(`(?i)\binconc`)

		re := regexp.MustCompile(`(?i)\b(pass\w*|fail\w*|\w*error\w*|unstable|none|inconc\w*)\b`)
		return re.ReplaceAllStringFunc(s, func(s string) string {
			switch {
			case pass.MatchString(s):
				return "[32;1m" + s + "[0m"
			case fail.MatchString(s):
				return "[31;1m" + s + "[0m"
			case unstable.MatchString(s):
				return "[38;5;208;1m" + s + "[0m"
			case inconc.MatchString(s) || none.MatchString(s):
				return "[38;5;201;1m" + s + "[0m"
			default:
				return s
			}
		})
	},
	"join": func(sep string, v interface{}) string {
		return strings.Join(v.([]string), sep)
	},
	"json": func(v interface{}) (string, error) {
		b, err := json.Marshal(v)
		return string(b), err
	},
}

func init() {
	Command.PersistentFlags().BoolVarP(&useJSON, "json", "", false, "output report in JSON format")
	Command.PersistentFlags().BoolVarP(&useJUnit, "junit", "", false, "output report in Junit format")
	Command.PersistentFlags().StringVarP(&templateText, "template", "t", summaryTemplate, "output report with custom template")
}
