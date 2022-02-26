package main

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func addHelp(topics []string, short string, long string) {
	// Note, we currently describe only internal help topics. Therefore we
	// don't display them for our external users, yet.
	path, err := os.Executable()
	if err != nil {
		return
	}
	if filepath.Base(path) != "k3" {
		return
	}

	h := &cobra.Command{
		Use:     topics[0],
		Aliases: topics[1:],
		Short:   short,
		Long:    long,
	}

	Command.AddCommand(h)
}

func init() {
	addHelp([]string{"plugins", "codecs", "adapters", "plugin"}, "How to write plugins", `Plugins.

Plugins connect the TTCN-3 runtime with your test environment and application
under test. There three flavors you can choose from:

* 'External Functions', also called platform adapters, allow executing code
   from other languages with C-bindings, such as C++ or Go.

* 'Connectors', also called system adapters, allow communication with other
   processes.

* 'Codecs', also called codecs ;-), allow conversion of TTCN-3 data to and from
   aribtrary formats.

Plugins are a challenging topic, because you need to know about some very
specific TTCN-3 details, K3-specific behaviour as well as details about dynamic
linking of C++ libraries.
Therefore, we try to provide all possible plugins, so you don't need to and may
focus on your domain. You'll find a currated list at:

  https://gitlabe1.ext.net.nokia.com/k3/k3/-/wikis/home#plugins


Should you still want to write a custom plugin, all you need to do is to put C
or C++ files into their own directory. K3 will take of their compilation:
All C++ or C files in a directory are automatically compiled as k3-plugin.
One directory is one plugin. The name of the plugin will be the basename of
the directory.

You'll find some (aged) documentation on how to write a plugin at:

    https://gitlabe1.ext.net.nokia.com/k3/k3/wikis/plugins
    https://gitlabe1.ext.net.nokia.com/k3/k3/wikis/plugins/k3r-classic-interface


If you need additional libraries you may use environment variables to modify
compilation. This is how a plugin is currently compiled:

	$CXX $CXXFLAGS <sources> $LDFLAGS $EXTRA_LDFLAGS -lk3-plugin -shared -fPIC -o k3r-<name>-plugin.so

Environment variable may be set-up via 'k3.env' or in the variables section of
the manifest-file.

If you have questions or if you want to use your own build-system you are
welcome to contact us.

`)

	addHelp([]string{"package.yml", "manifest"}, "What is the manifest file for?", `The Test Suite Manifest (packacke.yml).

To execute a test suite you usually need more than just a bunch of TTCN-3 source
files: You need generators, adapters, codecs, a lot of scripting, compile time
configuration, runtime configuration, post processing tools, caching of
build-artifacts and more. A manifest file provides a stable frame for
tools to work together nicely.

Therefore, every test suite should provide a manifest file 'package.yml' at the
root of the test suite directory structure. Supported fields:


| Name             | Type              | Details
| ---------------- | ----------------- | --------
| name             | string            | Name of the test suite.
| sources          | string[]          | TTCN-3 Source files containing tests.
| imports          | string[]          | Packages the suite depends on. This could be adapters, codecs, generators, ...
| timeout          | number            | Default timeout for tests in seconds.
| test_hook        | string            | Path to test hook script.
| parameters_file  | string            | Path to module parameters file.
| variables        | map[string]string | A key value list of custom variables.


**Environment Variables**

Manifest values can be overwritten by environment variables. Environment
variables will always take precedence over regular variables. Regular variables
have to be declared in a TOML formatted file 'k3.env' or in 'variables' section in
the manifest:

    $ echo '{"variables": {"K3_NAME": "OrignalName" }, "name": "$K3_NAME" }' > package.yml

    $ k3 show -- name
    OriginalName

    $ K3_NAME=NewName k3 show -- name
    NewName


You also can overwrite arrays like 'sources' or 'imports' with environment
variables ('K3_SOURCES="foo.ttcn3 bar.ttcn3" ...'), but note that spaces might
become problematic.


**CMake**

Commiting a 'package.yml' to your test suite source directory might be
problematic if the manifest depends on code-generators, which need to be build
by CMake first.

Therefore we support a CMake module to generate the manifest file during CMake
generation/compilation. In that scenario the manifest file won't be in
$CMAKE_CURRENT_SOURCE_DIR, but in $CMAKE_CURRENT_BINARY_DIR.

`)

	addHelp([]string{"hooks", "hook", "test_hook"}, "How to setup a test and control the SUT", `The Test Hook.

The test hook is used to configure, start and stop applications or additional
tools like tcpdump. It's default name is ${K3_NAME}.control. Alternatively it
may be pecified environment variable K3_TEST_HOOK or by the manifest key
test_hook.

The same hook is called for different events. The event is passed as first
argument. Following events are currently supported:

  before-build  Called before build.

  after-build   Called after build.

  before-run    Called before run.

  setup         Called by SutControl.setup

  run           Called by SutControl.run. Usually called to execute SUT. Is
                expected to block.

  teardown      Called by SutControl.teardown. Usually at the end of a test.
                Please note, if the TTCN-3 runtime is forced to quit, for
		example by signal, this action might not be called.

  after-run     Called after run. Can used by post-processing tools.


Several environment variables are provided. Test suite specific environment
variables:

  K3_NAME           Name of the test suite.
  K3_TIMEOUT        The test execution timeout in seconds, if specified.
  K3_SESSION_ID     A system-wide unique integer. Can be used for deriving IP
                    addresses.
  K3_SOURCE_DIR     Directory of package.yml. If test suite has no root folder,
                    this variable will be empty.

And all variables from k3.env and variables-sections (try 'k3 show -- env').

Event specific environment variables:

  K3_TEST_NAME      Full qualified test-name
  K3_TEST_VERDICT   Verdict of executed test case (only available for after-run)
  K3_TEST_LOG       Path to log-file (only available for after-run)
  K3_SUT_ID         String to identify multiple SUT instances. Only availble
                    in setup, run and teardown.

Following IO-redirection is used when test hook script is called for events
setup, run, teardown:

  stdout    $K3_TEST_NAME.$K3_SUT_ID.out
  stderr    $K3_TEST_NAME.$K3_SUT_ID.err


**Tipps and Tricks**

Test hooks usually dispatch events to their final destination: 'after-build'
triggers linters, 'run' starts a SUT, 'after-run' triggers post-processing
tools, ...

Test hooks may be written in any language. Should you use bash, be careful with
back-processes: They don't terminate when k3 dies. You may use the monitor tool
'k3-terminator' to handle this.

Example script:

  #!/bin/bash -e

  function setup()
  {
      echo "Preparing SUT for test $K3_TEST_NAME"
  }

  function run()
  {
      exec $K3_NAME
  }

  function teardown()
  {
      echo "Remove some temporary files"
  }

  case "$1" in
     setup|run|teardown) $1;;
     *) echo >&2 "$0: unused action: '$1'";;
  esac
`)

	addHelp([]string{"asn1", "ASN.1"}, "", `ASN.1 codecs.

K3 approach to handle ASN.1 files is explicit mapping:

All ASN.1 files in a directory are automatically compiled into a k3-asn1 codec.
The name of the codec will be the basename of the directory. If you want to
decode not only PDU, but also embedded payloads, you may use pragmas. For
example:

	FNORD-directives

	DEFINITIONS AUTOMATIC TAGS ::=

	BEGIN

	--<ASN1.PDU FNORD-Definitions.PayloadMsg1>--
	--<ASN1.PDU FNORD-Definitions.PayloadMsg2>--

	-- a module cannot be empty
	Null ::= NULL

	END


An example setup could look like this:

	$ ls
	package.yml FNORD/ tests/

	$ cat package.yml
	sources:
	  - tests/
	imports:
	  - FNORD/

	$ ls FNORD/
	FNORD-Definitions.asn FNORD-directives.asn

	$ mkdir -p build/ && cd build/
	$ k3 build ..
	$ ls
	FNORD.enc.c  FNORD.enc.h  FNORD.enc.lst  FNORDlib.so  FNORDmod.ttcn3

The file FNORDlib.so will be used by k3 to decode and encode ASN.1 coded
messages. The file FNORDmod.ttcn3 contains the TTCN-3-mapped ASN.1 types.

Note, K3 also supports incremental compilation.
`)

}
