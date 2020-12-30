---
nav_order: 3
---

# Getting started
{: .fs-9 }

ntt is a free and open application stack for TTCN-3, which lets you focus on writing tests
{: .fs-6 .fw-300 }

<details open markdown="block">
  <summary>
    Table of contents
  </summary>
  {: .text-delta }
1. TOC
{:toc}
</details>


_**We are currently updating this page, most content should be back until Friday.  
Sorry for the inconvenience.**_

<!--



# Command Line Interface

provide a uniform user interface, where possible:

    ntt <command> [<sources>...] [--] [<args>...]

* `<command>`: The command you want to execute, sub-commands are possible.
* `<sources>...`: The test suite sources. This might be a list of .ttcn3 files
  or the test suite root directory. If your test suite requires additional
  adapters, the test suite root directory must contain a manifest file (see below).
* `--`: This marker is required to separate the sources list from the
  remaining arguments.
* `<args>...`: Remaining arguments.

Example:

    ntt show foo.ttcn3 bar.ttcn3 -- name sources


## Available Commands

| Command         | Details
| --------------- | -------
| `ntt version`    | Displays version, git revision and build time if available.
| `ntt list`       | Lists tests, imports and modules of a test suite.
| `ntt show`       | Shows manifest information and variables.
| `ntt tags`       | Creates a tags file with exuberant ctags format.
| `ntt lint`       | Shows warnings and errors. This is a work in progress.
| `ntt mcov`       | Reads a ntt runtime log and reports message coverage.
| `ntt langserver` | Starts a language server waiting on stdin.


## Custom Commands

You can extend and customize ntt through custom commands. Place any executable
with a name like `ntt-jaegerschnitzel` in your `PATH` and ntt will automatically
make it available as a subcommand. You can then call it just like any other ntt
command:

    $ ntt jaegerschnitzel +6000



## Environment variables

You may define environment variable `NTT_SOURCE_DIR` to specify a test suite root directory:

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

## Related Tools and Scripts

This section will be completed in a few day. Sorry for the inconvenience.

-->

<!--

## ttcn3c

## k3objdump

## ntt-mcov

## Formatting

Much energy spent in arguing about coding style. Which is a waste especially because
formatting rules play only a minor role in code quality.  
Future versions ntt will include a formatter for TTCN-3 code.

## CMake

_Description of our CMake scripts will be published in a few days._

## gRPC

_Description of our gRPC interfaces will also be published in a few days._
-->

## Test Suite Manifest

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


