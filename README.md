[![Go Report Card](https://goreportcard.com/badge/github.com/nokia/ntt?style=flat-square)](https://goreportcard.com/report/github.com/nokia/ntt)
[![Build Status](https://travis-ci.com/nokia/ntt.svg?branch=master)](https://travis-ci.com/nokia/ntt)

<p align="center"><img src="https://nokia.github.io/ntt/static/ntt.png"/></p>

# About

This project provides open tools and libraries for testing with
[TTCN-3](http://www.ttcn-3.org). It builds upon 15 years of experience of
running production workloads at Nokia. This repository contains:

* A [modern CLI](#command-line-interface) for test suite configuration and
  execution, including a TTCN-3 language server for a better IDE experience.
* An error tolerant [TTCN-3 parser
  library](https://pkg.go.dev/github.com/nokia/ntt/internal/ttcn3/parser).
* A lazy [TTCN-3 compiler
  library](https://pkg.go.dev/github.com/nokia/ntt/internal/ntt) with focus on
minimizing latency.
* And more to come ...

Note, _the libraries are still internal, to give us some more time for
stabilizing the API_. The first releases will be for Go. C and Python will
follow later.

**What's TTCN-3?**

[TTCN-3](http://www.ttcn-3.org/) is a standardized language for scripting tests.
It supports various types of testing and is a good addition to unit tests.
Despite being clearly from the last millennium, TTCN-3 offers some excellent
features for testing protocols and services:

* Native support for concurrency and alternate code flows (similar to Go).
* A powerful matching mechanism. Think supercharged regular expressions for
  arbitrary data types.
* And it's language agnostic: Adapters and codecs are implemented in any
  language you see fit. So you can leverage any ecosystem you need.

For a nice introduction have a look at this video: https://media.ccc.de/v/XQKBZG


**Why not just use Titan**?

[Titan](https://github.com/eclipse/titan.core) is a mature tool with an
extensive set of adapters and modules. If you just need a TTCN-3 compiler or
runtime, use Titan!

If you need a language server, extra tooling or if you want to develop new
TTCN-3 tools you have come to the right place.


# Contact us

If you have questions, you are welcome to contact us at
[ntt@groups.io](mailto:ntt@groups.io).

You want to contribute? That's great! Kindly read our [contribution
guide](https://github.com/nokia/ntt/blob/master/CONTRIBUTING.md) for more
details.


# Install

You can choose between installing the pre-built binaries or compiling NTT from
source. Using the binaries is usually easier. Compiling from source means you
have more control.

## Install pre-built binaries

NTT provides pre-built binaries for Linux x64 and i386. As soon as we are
confident that the Linux binaries work nicely, we'll explore other deployments,
like for Windows, MacOS or in containers.

If you are running ntt already in Windows or any other platform we don't support
yet, it would be great if you could share your work and create a quick PR.


**Debian Packages**

Download the .deb package from the
[releases](https://github.com/nokia/ntt/releases) page and install with `dpkg -i`.


**RPM Packages**

Download the .rpm package from the
[releases](https://github.com/nokia/ntt/releases) page and install with package
manager of your choice. You can also install directly from the internet:

    sudo dnf install https://github.com/nokia/ntt/releases/download/v0.1/ntt.x86_64.rpm


## Compiling from Source

NTT requires a [Go compiler](https://golang.org/dl/) >= 1.10, git and make to
build. To build and install simply call:

	make
	sudo make install

You may control installation by specifying PREFIX and DESTDIR variables. For example:

	make PREFIX=/
	make install DESTDIR=$HOME/.local


# Getting Started

## Command Line Interface

NTT tools provide a uniform user interface, where possible:

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


**Available Commands**

| Command         | Details
| --------------- | -------
| `ntt version`    | Displays version, git revision and build time if available.
| `ntt list`       | Lists tests, imports and modules of a test suite.
| `ntt show`       | Shows manifest information and variables.
| `ntt tags`       | Creates a tags file with exuberant ctags format.
| `ntt lint`       | Shows warnings and errors. This is a work in progress.
| `ntt mcov`       | Reads a NTT runtime log and reports message coverage.
| `ntt langserver` | Starts a language server waiting on stdin.


**Custom Commands**

You can extend and customize NTT through custom commands. Place any executable
with a name like `ntt-jaegerschnitzel` in your `PATH` and ntt will automatically
make it available as a subcommand. You can then call it just like any other ntt
command:

    $ ntt jaegerschnitzel +6000



**Environment variables**

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


## The Test Suite Manifest

To execute a test suite you usually need more than just a bunch of TTCN-3 source
files: You need generators, adapters, codecs, a lot of scripting, compile time
configuration, runtime configuration, post processing tools, caching of
build-artifacts and more. A manifest file provides a stable frame for
tools to work together nicely.

Every NTT test suite should provide a manifest file `package.yml` at
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


## TTCN-3 Language Server

We are currently implementing the [Language Server Protocol](https://microsoft.github.io/language-server-protocol):

**Features**

| Name                        | Method                            |                    |
| --------------------------- | --------------------------------- | ------------------ |
| Workspace Symbols           | `workspace/symbol`                | :x:                |
| Execute Command             | `workspace/executeCommand`        | :x:                |
| Diagnostics                 | `textDocument/publishDiagnostics` | :x:                |
| Completion                  | `textDocument/completion`         | :x:                |
| Hover                       | `textDocument/hover`              | :x:                |
| Signature Help              | `textDocument/signatureHelp`      | :x:                |
| Goto Definition             | `textDocument/definition`         | :heavy_check_mark: |
| Goto Type Definition        | `textDocument/typeDefinition`     | :x:                |
| Goto Implementation         | `textDocument/implementation`     | :x:                |
| Find References             | `textDocument/references`         | :x:                |
| Document Highlights         | `textDocument/documentHighlight`  | :x:                |
| Document Symbols            | `textDocument/documentSymbol`     | :x:                |
| Code Action                 | `textDocument/codeAction`         | :x:                |
| Code Lens                   | `textDocument/codeLens`           | :x:                |
| Document Formatting         | `textDocument/formatting`         | :x:                |
| Document Range Formatting   | `textDocument/rangeFormatting`    | :x:                |
| Document on Type Formatting | `textDocument/onTypeFormatting`   | :x:                |
| Rename                      | `textDocument/rename`             | :x:                |


**The set of workspace folders**

This is very important. When you open multiple folders, the first one is
considered root folder. If you do not open the right folders, very little will
work. This is the most common issue of ntt language server that we see.

Unfortunately there isn't much you can do, yet. we are currently working on
diagnostics and additional debug messages, so we can figure out what relevant
use cases for editing test suites are.


**Visual Studio Code**

For TTCN-3 support in Visual Studio Code have a look at our co-project
[vscode-ttcn3](https://marketplace.visualstudio.com/items?itemName=Nokia.ttcn3)

**Vim**

Use [prabirshrestha/vim-lsp](https://github.com/prabirshrestha/vim-lsp/), with the following configuration:

```vim
if executable('ntt')
    au User lsp_setup call lsp#register_server({
        \ 'name': 'ntt',
        \ 'cmd': {server_info->['ntt', 'langserver']},
        \ 'whitelist': ['ttcn3'],
        \ })
endif

function! s:on_lsp_buffer_enabled() abort
    setlocal omnifunc=lsp#complete
    setlocal signcolumn=yes
    nmap <buffer> gd <plug>(lsp-definition)
    nmap <buffer> <f2> <plug>(lsp-rename)
endfunction

augroup lsp_install
    au!
    autocmd User lsp_buffer_enabled call s:on_lsp_buffer_enabled()
augroup END
```


## Future Development

Our goal is to improve the TTCN-3 user experience by providing modern tools. We
will focus on adding features to the language server.  
Of course you are very welcome to contribute. Check the
[feature requests](https://github.com/nokia/ntt/labels/enhancement) if something suits
you or create your own feature request.


## License

This project is licensed under the BSD-3-Clause license - see the [LICENSE](https://github.com/nokia/ntt/blob/master/LICENSE).
