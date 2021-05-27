# Reference
{: .no_toc }

* TOC
{:toc}


# Commands

## k3objdump

k3objdump displays information about files in [TTCN-3 executable format](https://pkg.go.dev/github.com/nokia/ntt/k3/t3xf) files.
T3XF is a binary representation of input TTCN-3 source text.

You may use command line flags `--all` or `--legacy` disassemble T3XF files:

```
$ k3objdump --all example.t3xf
       0: 03 00 00 00 nop
       4: 43 01 00 00 natlong
          02 00 00 00 2
      12: 33 00 00 00 version
      16: d3 00 00 00 scan
      20: 06 00 00 00   =1
      24: d3 00 00 00   scan
      28: 0e 00 00 00     =3
      32: d3 00 00 00     scan
      36: 53 01 00 00       ieee754dp
          00 00 00 00
          00 00 f0 3f       1.000000
      48: 83 00 00 00     block
      52: 43 0e 00 00     float
      56: 83 01 00 00     name
          0a 00 00 00
          67 75 61 72
          64 54 69 6d
          65 72 00 00     'guardTimer'
      76: 13 0a 00 00     mpard
          ...
```

## ntt

`ntt` is the main command line front end for working with TTCN-3.
It provides a uniform user interface, where possible:

    ntt <command> [<sources>...] [--] [<args>...]

* `<command>`: The command you want to execute, sub-commands are possible.
* `<sources>...`: The test suite sources. This might be a list of .ttcn3 files
  or the test suite root directory. If your test suite requires additional
  adapters, the test suite root directory must contain a manifest file.
* `--`: This marker is required to separate the sources list from the
  remaining arguments.
* `<args>...`: Remaining arguments.

Example:

    ntt show foo.ttcn3 bar.ttcn3 -- name sources


**Custom Commands**

You can extend and customize ntt through custom commands. Place any executable
with a name like `ntt-banana` in your `PATH` and ntt will automatically make it
available as a subcommand. You can then call it just like any other ntt
command:

    $ ntt banana +6000


**Environment variables**

You may define environment variable `NTT_SOURCE_DIR` to specify a default test
suite root directory:

    $ ntt list                      # Lists tests in current working directory
    $ export NTT_SOURCE_DIR=~/foo
    $ ntt list                      # Now, ntt lists tests in ~/foo


Environment variable `NTT_CACHE` is a colon-separated list of directories and has
similar purpose and behaviour like GNU Make's VPATH. It is use to find files
like `ntt.env`:

    $ echo "FOO=23" > ntt.env
    $ mkdir -p build && cd build
    $ NTT_CACHE=.. ntt show -- FOO
    23


**Debugging**

For debugging purposes you may increase ntt's loglevel with the `--verbose`
command line option or `K3_DEBUG` environment variable.

For performance analysis you may enable [profiling](https://blog.golang.org/pprof)
using the `--cpuprofile` option.


## ntt build
The build command provides package manager functionality and builds a test
executable. Note, we build command is only available for Nokia internal
environments.

If no arguments are specified ntt builds a test executable from current
directory.

Use environment variable `PARALLEL_BUILD_JOBS` to specify how many build
steps shall be executed in parallel (default is number of CPU cores).

Imported ASN1 files will be pass to OSS ASN.1 toolchain, every C and C++ file
will be compiled into a K3 runtime plugin.


## ntt completion
ntt provide bash completion. To load completion run

       . <(ntt completion)

To configure your bash shell to load completions for each session add to your bashrc

        # ~/.bashrc or ~/.profile
        . <(ntt completion)

Note, if bash-completion is not installed on Linux, please install the
'bash-completion' package via your distribution's package manager.

Note, the generated completion file is not compatible with zsh. Please open an
issue (or even pull request) should require support for zsh.


## ntt langserver

Start TTCN-3 language server and wait for input on stdin. This command is
usually used by IDE or editors only.

## ntt lint

The lint command examines TTCN-3 source files and reports suspicious code.
It may find problems not caught by the compiler, but also constructs
considered "bad style".

Lint's exit code is non-zero for erroneous invocation of the tool or if a
problem was reported.


**Formatting Checks**

* `max_lines`: Number of lines a behaviour body must not exceed.
* `aligned_braces`: Braces must be in the same column or same line.
* `require_case_else`: Every select-statement must have one case-else.


**Cyclomatic Complexity Checks**

* `complexity.max`: Cyclomatic complexity muss not exceed.
* `complexity.ignore_guards`: Ignore complexity of alt- and interleave guards


**Naming Convention Checks**

* `naming.modules`: Checks for module identifiers.
* `naming.tests`: Checks for test-case identifier.
* `naming.functions`: Checks for function identifiers.
* `naming.altsteps`: Checks for altstep identifiers.
* `naming.parameters`: Checks for parameter identifiers.
* `naming.component_vars`: Checks for component variable identifiers.
* `naming.var_templates`: Checks for variable template identifiers.
* `naming.port_types`: Checks for port type identifiers.
* `naming.ports`: Checks for port instance identifiers.
* `naming.global_consts`: Checks for global constant identifiers.
* `naming.component_consts`: Checks for component scoped constant identifiers.
* `naming.templates`: Checks for constant template identifiers.
* `naming.locals`: Checks for local variable identifiers.

* `tags.tests`: Checks for test-case tags.


**White-Listing**

* `ignore.modules`: Ignore modules
* `ignore.files`: Ignore files


**Refactoring**

When TTCN-3 code is refactored incrementally, it happens that references to
legacy code are faster added than one can remove them. This check helps with a
warning, as soon as the usage of a symbol exceed a defined limit.


**Unused Symbols**

* `unused.modules`: Checks for unused modules


**Example**

```yaml
aligned_braces: true
require_case_else: true
max_lines: 40

usage:
  "foo":
    limit: 12
    text: Use "bar" instead.

unused:
  modules: true

complexity:
  max: 15
  ignore_guards: true

tags:
  tests:
    "@author": "testcases must have a @author tag"

naming:
  tests:
    # An exlamation mark inverts the match.
    "!.{130,}": "testcase identifiers must not be longer than 130 characters"

  functions:
    "^[a-z]"      : "function identifiers must begin with a lower case letter"
    "!^(f|func)_" : "function identifiers must not begin with f_ or func_"

  global_consts:
    "^[A-Z0-9_]+$": "global constants must be UPPER_CASE"

ignore:
  modules:
    # Ignore generated modules
    - "^Protobuf_.+$"

  files:
    # Ignore all files from generated folders
    - "generated/"
```

## ntt list
List various types of objects.

List control parts, modules, imports or tests. The list command without any explicit
sub-commands will output tests.

List will ignore imported directories when printing tests. If you need to list all
tests from a testsuite you'll have to pass .ttcn3 files as arguments.
Example:

    ntt list $(ntt show -- sources) $(find $(ntt show -- imports) -name \*.ttcn3)


**Filtering**

You can use regular expressions to filter objects. If you pass multiple regular
expressions, all of them must match (AND). Example:

	$ cat example.ttcn3
	testcase foo() ...
	testcase bar() ...
	testcase foobar() ...
	...

	$ ntt list --regex=foo --regex=bar
	example.foobar

	$ ntt list --regex='foo|bar'
	example.foo
	example.bar
	example.foobar


Similarly, you can also specify regular expressions for documentation tags.
Example:

	$ cat example.ttcn3
	// @one
	// @two some-value
	testcase foo() ...

	// @two: some-other-value
	testcase bar() ...
	...

	$ ntt list --tags-regex=@one --tags-regex=@two
	example.foo

	$ ntt list --tags-regex='@two: some'
	example.foo
	example.bar


**Baskets**

Baskets are filters defined by environment variables of the form:

        NTT_LIST_BASKETS_<name> = <filters>

For example, to define a basket "stable" which excludes all objects with @wip
or @flaky tags:

	export NTT_LIST_BASKETS_stable="-X @wip|@flaky"

Baskets become active when they are listed in colon separated environment
variable `NTT_LIST_BASKETS`. If you specify multiple baskets, at least of them
must match (OR).

Rule of thumb: all baskets are ORed, all explicit filter options are ANDed.
Example:

	$ export NTT_LIST_BASKETS_stable="--tags-exclude @wip|@flaky"
	$ export NTT_LIST_BASKETS_ipv6="--tags-regex @ipv6"
	$ NTT_LIST_BASKETS=stable:ipv6 ntt list -R @flaky


Above example will output all tests with a @flaky tag and either @wip or @ipv6 tag.


If a basket is not defined by an environment variable, it's equivalent to a
"--tags-regex" filter. For example, to lists all tests, which have either a
@flaky or a @wip tag:

	# Note, flaky and wip baskets are not specified explicitly.
	$ NTT_LIST_BASKETS=flaky:wip ntt list

	# This does the same:
	$ ntt list --tags-regex="@wip|@flaky"

## ntt mcov

TODO: Explain the message coverage tool.

## ntt report
The report command shows a summary of the latest test run. The summary includes
information such as a list of tests which did not pass, average run times, CPU
load, etc.
Command line options `--json` and `--junit` show similar output, but with JSON
or JUNIT formatting.


**Templating**

ntt uses the Go templates format which you can use to specify custom output templates.
Example:

{% raw %}
	ntt report --template "{{.Name}} took {{.Tests.Duration}}"
{% endraw %}

Available Objects:

 * `.Report` is a collection of test runs
 * `.Report.Cores`: number of CPU cores
 * `.Report.Environ`: list of environment variable
 * `.Report.Getenv`: value of an environment variable
 * `.Report.LineCount`: number of TTCN-3 source code lines
 * `.Report.MaxJobs`: maximum number of parallel test jobs
 * `.Report.MaxLoad`: maximum allowed CPU load
 * `.Report.Modules`: a list of collection sorted by module
 * `.Report.Name`: name of the collection
 * `.Report.Runs`: list of test runs
 * `.Report.Tests`: list of tests (with final verdict)

 * `.RunSlice` is a list of test runs
 * `.RunSlice.Load`: Returnurfsystemload slice for every run
 * `.RunSlice.Average`: Average duration of runs (median)
 * `.RunSlice.Deviation`: Standard deviation
 * `.RunSlice.Duration`: Timespan of first and last test run
 * `.RunSlice.Failed`: A slice of failed test runs (inconc, none, error, fail, ...)
 * `.RunSlice.First`: First test run
 * `.RunSlice.Last`: Last test run
 * `.RunSlice.Longest`: Longest test run
 * `.RunSlice.NotPassed`: A slice of tests without 'pass' verdict
 * `.RunSlice.Result`: Final result (PASSED, FAILED, UNSTABLE, NOEXEC)
 * `.RunSlice.Shortest`: Shortest test run
 * `.RunSlice.Total`: Sum of all test run durations
 * `.RunSlice.Unstable`: List of unstable test runs

 * `.Run` is a individual test run
 * `.Run.ID`: test run ID (e.g. test.Stable_A-2)
 * `.Run.Name`: full qualified test name (test.Stable_A)
 * `.Run.Instance`: test instance (e.g. 2)
 * `.Run.Module`: module name (test)
 * `.Run.Testcase`: testcase name (e.g. Stable_A)
 * `.Run.Verdict`: the test verdict (pass, fail, none, ...)
 * `.Run.Begin`: when the test was started (time.Time Go object)
 * `.Run.End`: when the test ended (time.Time Go object)
 * `.Run.Duration`: a time.Duration Go object
 * `.Run.Load`: the system load when the test was started
 * `.Run.MaxMem`: the maximum memory used when the test ended
 * `.Run.Reason`: optional reason for verdicts
 * `.Run.ReasonFiles`: content of \*.reason files
 * `.Run.RunnerID`: the ID of the runner exeuting the run
 * `.Run.WorkingDir`: working Directory of the test

 * `.File` is a (reason) file
 * `.File.Name`: path to file
 * `.File.Content`: content of file


Additional filters:

 * `green`: output ANSI sequences for color green
 * `red`: output ANSI sequences for color red
 * `orange`: output ANSI sequences for color orange
 * `bold`: output ANSI sequences for bold text
 * `off`: output ANSI sequences to reset attributes
 * `colorize`: colorize output
 * `join`: join input with a separator
 * `json`: encode input using JSON format
 * `min`: returns the minimum of a float slice
 * `max`: returns the maximum of a float slice
 * `median`: returns the median of a float slice



**Examples**

Summary template:
```
{% raw %}
{{bold}}==================================  Summary  =================================={{off}}
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
{% endraw %}
```


JUnit template:
```xml
{% raw %}
<?xml version="1.0" encoding="UTF-8"?>
<testsuites>{{range .Modules}}

<testsuite name="{{.Name}}" tests="{{len .Tests}}" failures="{{len .Tests.Failed}}" errors="" time="{{.Tests.Total.Seconds}}">
{{range .Tests}}<testcase name="{{.Testcase}}" time="{{.Duration.Seconds}}">
  {{if and (ne .Verdict "unstable") (ne .Verdict "pass")}}<failure>Verdict: {{.Verdict}} {{with .Reason}}({{. | html }}){{end}}
{{range .ReasonFiles}}{{.Name}}: {{.Content}}{{end}}
  </failure>
{{end}}</testcase>

{{end}}</testsuite>
{{end}}</testsuites>
{% endraw %}
```

JSON template:
```json
{% raw %}
{
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
{% endraw %}
```

## ntt run
Run tests from a TTCN-3 test suite. Note, this command is only available in Nokia internal environments.

**Module Parameters**

Module parameters can be passed by file and will be read by ntt automatically.
Default name is `$NTT_NAME.parameters`. This value can be overwritten by
configuration key `parameters_file` in the manifest file.

Test specific module parameter are load from file `$SCT_SOURCE_DIR/modulePar/$MODULE/$TEST.parameters`.

## ntt show
Show test suite information like name, sources, environment variables, ...

The show command provide additional output formats:

* `--json`: output in JSON format
* `--sh`: output test suite data for shell consumption


## ntt tags
The tags command generates an index (or "tag") file for TTCN-3 language objects found in file(s).

This tag file allows these items to be quickly and easily located by a text
editor or other utility. A "tag" signifies a language object for which an index
entry is available (or, alternatively, the index entry created for that
object).

The tags command will also generate tags for fields, members, ... .


## ntt version
Display ntt version if available.

## ttcn3c
ttcn3c parses TTCN-3 files and generates output based on the options given. The `--generator` argument specifies what generator plugin shall be used by ttcn3c. Default is a t3xf output.

# API

The Go API is described here:
* [k3](https://pkg.go.dev/github.com/nokia/ntt/k3): convenience functions for supporting k3 toolchain
* [k3/log](https://pkg.go.dev/github.com/nokia/ntt/k3/log): parsing k3 runtime log files
* [k3/t3xf](https://pkg.go.dev/github.com/nokia/ntt/k3/t3xf): decoding TTCN-3 Executable Format

# gRPC

TODO: explain gRPC interfaces.


# Manifest file package.yml

To execute a test suite you usually need more than just a bunch of TTCN-3 source
files: You need generators, adapters, codecs, a lot of scripting, compile time
configuration, runtime configuration, post processing tools, caching of
build-artifacts and more. A manifest file provides a stable frame for
tools to work together nicely.

Every ntt test suite should provide a manifest file `package.yml` at
the root of the test suite directory structure. Supported fields:


| Name               | Type              | Details
| ------------------ | ----------------- | --------
| `name`             | string            | Name of the test suite.
| `sources`          | string[]          | TTCN-3 Source files containing tests.
| `imports`          | string[]          | Packages the suite depends on. This could be adapters, codecs, generators, ...
| `timeout`          | number            | Default timeout for tests in seconds.
| `test_hook`        | string            | Path to test hook script.
| `parameters_file`  | string            | Path to module parameters file.
| `variables`        | map[string]string | A key value list of custom variables.


**Environment Variables**

Manifest values can be overwritten by environment variables. Environment
variables will always take precedence over regular variables. Regular variables
have to be declared in a TOML formatted file `ntt.env` or in `variables` section in
the manifest:

    $ echo '{"variables": {"NTT_NAME": "OrignalName" }, "name": "$NTT_NAME" }' > package.yml

    $ ntt show -- name
    OriginalName

    $ NTT_NAME=NewName ntt show -- name
    NewName


You also can overwrite arrays like `sources` or `imports` with environment
variables (`NTT_SOURCES="foo.ttcn3 bar.ttcn3" ...`), but note that spaces might
become problematic.

# CMake

ntt provides a CMake module with functions to help use the NTT/K3 Test infrastructure. It
provides function `add_ttcn3_suite` for generating a test suite
manifest and function `protobuf_generate_ttcn3` for calling a protoc generator plugin.

TODO: explain CMake interfaces in greater details, possibly with examples.


# VSCode Extension Settings

TODO: Explain vscode settings
