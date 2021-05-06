package report

import (
	"bufio"
	"encoding/json"
	"os"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"github.com/nokia/ntt/internal/ntt"
	"github.com/spf13/cobra"
)

var (
	Command = &cobra.Command{
		Use:   "report",
		Short: "show information about latest test run",
		Long: `show information about latest test run",

The report command shows a summary of the latest test run. The summary includes
information such as a list of tests which did not pass, average run times, CPU
load, etc.
Command line options '--json' and '--junit' show similar output, but with JSON
or JUNIT formatting.


Templating
----------

ntt uses the Go templates format which you can use to specify custom output templates.
Example:

	ntt report --template "{{.Name}} took {{.Tests.Duration}}"


Available Objects

  .Report is a collection of test runs
  .Report.Cores:     number of CPU cores
  .Report.Environ:   list of environment variable
  .Report.Getenv:    value of an environment variable
  .Report.LineCount: number of TTCN-3 source code lines
  .Report.MaxJobs:   maximum number of parallel test jobs
  .Report.MaxLoad:   maximum allowed CPU load
  .Report.Modules:   a list of collection sorted by module
  .Report.Name:      name of the collection
  .Report.Runs:      list of test runs
  .Report.Tests:     list of tests (with final verdict)

  .RunSlice is a list of test runs
  .RunSlice.Load:      Return systemload slice for every run
  .RunSlice.Average:   Average duration of runs (median)
  .RunSlice.Deviation: Standard deviation
  .RunSlice.Duration:  Timespan of first and last test run
  .RunSlice.Failed:    A slice of failed test runs (inconc, none, error, fail, ...)
  .RunSlice.First:     First test run
  .RunSlice.Last:      Last test run
  .RunSlice.Longest:   Longest test run
  .RunSlice.NotPassed: A slice of tests without 'pass' verdict
  .RunSlice.Result:    Final result (PASSED, FAILED, UNSTABLE, NOEXEC)
  .RunSlice.Shortest:  Shortest test run
  .RunSlice.Total:     Sum of all test run durations
  .RunSlice.Unstable:  List of unstable test runs

  .Run is a individual test run
  .Run.ID:          test run ID (e.g. test.Stable_A-2)
  .Run.Name:        full qualified test name (test.Stable_A)
  .Run.Instance:    test instance (e.g. 2)
  .Run.Module:      module name (test)
  .Run.Testcase:    testcase name (e.g. Stable_A)
  .Run.Verdict:     the test verdict (pass, fail, none, ...)
  .Run.Begin:       when the test was started (time.Time Go object)
  .Run.End:         when the test ended (time.Time Go object)
  .Run.Duration:    a time.Duration Go object
  .Run.Load:        the system load when the test was started
  .Run.MaxMem:      the maximum memory used when the test ended
  .Run.Reason:      optional reason for verdicts
  .Run.ReasonFiles: content of *.reason files
  .Run.RunnerID:    the ID of the runner exeuting the run
  .Run.WorkingDir:  working Directory of the test

  .File is a (reason) file
  .File.Name:    path to file
  .File.Content: content of file


Additional commands

  green:    output ANSI sequences for color green
  red:      output ANSI sequences for color red
  orange:   output ANSI sequences for color orange
  bold:     output ANSI sequences for bold text
  off:      output ANSI sequences to reset attributes
  colorize: colorize output
  join:     join input with a separator
  json:     encode input using JSON format
  min:      returns the minimum of a float slice
  max:      returns the maximum of a float slice
  median:   returns the median of a float slice



Summary template:
` + summaryTemplate + `


JUnit template:
` + junitTemplate + `


JSON template:
` + jsonTemplate + `


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
{{else}}{{if eq (len .Tests) 0}}{{orange}}{{bold}}WARNING: No matching test cases found!{{off}}
{{else}}{{green}}all tests have passed{{off}}
{{end}}{{end}}
{{len .Tests}} test cases took {{bold}}{{.Tests.Duration}}{{off}} to execute (total runs: {{len .Runs}}
{{- with .Tests.Failed}}, {{red}}not passed: {{len .}}{{off}}{{end}}
{{- with .Tests.Unstable}}, {{orange}}unstable: {{len .}}{{off}}{{end}})
{{bold}}==============================================================================={{off}}

{{ printf "%s (±%s)" .Tests.Average .Tests.Deviation | printf "Average  : %-30s CPU cores      : " }}{{printf "%d" .Cores}}
{{ printf "Shortest : %-30s Parallel tests : %d" .Tests.Shortest.Duration .MaxJobs }}
{{ printf "Longest  : %-30s Load limit     : %d" .Tests.Longest.Duration .MaxLoad}}
{{ printf "Total    : %-30s Load average   : %.2f" .Tests.Total (median .Tests.Load)}}

{{bold}}==============================================================================={{off}}
{{bold}}Final Result: {{.Tests.Result | colorize}}{{off}}
{{bold}}==============================================================================={{off}}
`

	junitTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<testsuites>{{range .Modules}}

<testsuite name="{{.Name}}" tests="{{len .Tests}}" failures="{{len .Tests.Failed}}" errors="" time="{{.Tests.Total.Seconds}}">
{{range .Tests}}<testcase name="{{.Testcase}}" time="{{.Duration.Seconds}}">
  {{if and (ne .Verdict "unstable") (ne .Verdict "pass")}}<failure>Verdict: {{.Verdict}} {{with .Reason}}({{. | html }}){{end}}
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
  "suite": {
    "linecount": {{.LineCount}}
  },
  "load": {
    "min" : {{min .Tests.Load}},
    "max" : {{max .Tests.Load}},
    "avg" : {{median .Tests.Load}}
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

	if templateText == "" {
		templateText = summaryTemplate
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

		re := regexp.MustCompile(`(?i)\b(NORUN|pass\w*|fail\w*|\w*error\w*|unstable|none|inconc\w*)\b`)
		return re.ReplaceAllStringFunc(s, func(s string) string {
			switch {
			case s == "NORUN":
				return "[38;5;208;1m" + s + "[0m"
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
	"min": func(v []float64) float64 {
		if len(v) == 0 {
			return 0
		}

		x := v[0]
		for _, r := range v {
			if r < x {
				x = r
			}
		}
		return x
	},
	"max": func(v []float64) float64 {
		if len(v) == 0 {
			return 0
		}

		x := v[0]
		for _, r := range v {
			if r > x {
				x = r
			}
		}
		return x
	},
	"median": func(v []float64) float64 {
		if len(v) == 0 {
			return 0
		}

		sort.Float64s(v)

		n := len(v) / 2

		// odd v length
		if len(v)&1 == 1 {
			return v[n]
		}

		// even v length
		return (v[n-1] + v[n]) / 2
	},
}

func init() {
	Command.PersistentFlags().BoolVarP(&useJSON, "json", "", false, "output report in JSON format")
	Command.PersistentFlags().BoolVarP(&useJUnit, "junit", "", false, "output report in Junit format")
	Command.PersistentFlags().StringVarP(&templateText, "template", "t", "", "output report with custom template")
}
