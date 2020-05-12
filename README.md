[![Go Report Card](https://goreportcard.com/badge/github.com/nokia/ntt?style=flat-square)](https://goreportcard.com/report/github.com/nokia/ntt)
[![Build Status](https://travis-ci.com/nokia/ntt.svg?branch=master)](https://travis-ci.com/nokia/ntt)

<p align="center"><img src="https://user-images.githubusercontent.com/12729086/81290451-63961480-9068-11ea-82b0-572c9d763fa7.png"/></p>

# About

This project provides open tools and libraries for testing with
[TTCN-3](http://www.ttcn-3.org). It builds upon 15 years of experience of
running production workloads at Nokia. This repository contains:

* A TTCN-3 language server for a better IDE experience.
* A modern CLI for test suite configuration and execution.
* An error tolerant [TTCN-3 parser library](https://pkg.go.dev/github.com/nokia/ntt/internal/ttcn3/parser).
* A lazy [TTCN-3 compiler library](https://pkg.go.dev/github.com/nokia/ntt/internal/ntt) with focus on minimizing latency.
* And more to come ...

Note, the libraries are still internal, to give us some more time for
fine-tuning the API. The first releases will be for Go. But C and Python will
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

Because this is a Go project you may use the usual Go methods to build and
install NTT. Note, you usually won't have proper version info.

NTT requires only a [Go compiler](https://golang.org/dl/) >= 1.10 to build. The
simplest method to build and install NTT is by `go get`. 


**Go Get**

`go get` will download, build and install NTT into your GOPATH:

        go get -u github.com/nokia/ntt/cmd/...

You'll find the commands usually in `$HOME/go/bin`.


**Git Clone && Go Build**

For the manual method you also need a Git client. You can decide where to
clone the repository or where to install NTT:

    $ git clone https://github.com/nokia/ntt.git
    $ cd ntt
    $ go build ./cmd/k3

Go will download all dependencies, if you have Go modules support available. If
you don't want Go to connect to the Internet, you may disable Go proxies and
enable vendoring:

    $ GOPROXY=off go build -mod=vendor ./cmd/k3


If you want to provide your own version information, you may also pass linker flags:

    $ go build -ldflags="-X main.version=1.2.3 -X main.commit=deadbeef123 -X main.date=today" ./cmd/k3
    $ ./k3 version
    k3 1.2.3, commit deadbeef123, built at today


# Getting Started

To execute a test suite you usually need more than just a bunch of TTCN-3 source
files: You need generators, adapters, codecs, a lot of scripting, compile time
configuration, runtime configuration, post processing tools, caching of
build-artifacts and much, much more. A manifest file provides a stable frame for
tools to work together nicely.

**The Test Suite Manifest**

Every NTT test suite should provide a manifest file `package.yml` at
the root of the test suite directory structure. Supported fields:


| Name               | Type     | Details
| ------------------ | -------- | --------
| `name`             | string   | Name of the test suite.
| `sources`          | string[] | TTCN-3 Source files containing tests.
| `imports`          | string[] | Packages the suite depends on. This could be adapters, codecs, generators, ...
| `timeout`          | number   | Default timeout for tests in seconds.
| `test_hook`        | string   | Path to test hook script.
| `parameters_file`  | string   | Path to module parameters file.


You can use variables and environment variables in the manifest. Variables have
to be declared in a TOML formatted file `k3.env`. Environment variables always
take precedence:

    $ echo "name: OriginalName" > package.yml

    $ k3 show -- name
    OriginalName

    $ K3_NAME=NewName k3 show -- name
    NewName


You also can overwrite arrays like `sources` or `imports` with environment
variables (`K3_SOURCES="foo.ttcn3 bar.ttcn3" ...`), but note that spaces might
become problematic.


## Command Line Interface

NTT tools provide a uniform user interface, where possible:

    k3 <command> [<sources>...] [--] [<args>...]

* `<command>`: The command you want to execute, sub-commands are possible.
* `<sources>...`: The test suite sources. This might be a list of .ttcn3 files
  or the test suite root directory. If your test suite requires additional
  adapters, the test suite root directory must contain a manifest file.
* `<-->`: This marker is required to separate the sources list from the
  remaining arguments.
* `<args>...`: Remaining arguments.

Example:

    k3 show foo.ttcn3 bar.ttcn3 -- name sources


**Available Commands**

| Command         | Details
| --------------- | -------
| `k3 version`    | Displays version, git revision and build time if available.
| `k3 list`       | Lists tests, imports and modules of a test suite.
| `k3 show`       | Shows manifest information and variables.
| `k3 tags`       | Creates a tags file with exuberant ctags format.
| `k3 lint`       | Shows warnings and errors. This is a work in progress.
| `k3 mcov`       | Reads a K3 runtime log and reports message coverage.
| `k3 langserver` | Starts a language server waiting on stdin.


**Custom Commands**

You can extend and customize NTT through custom commands. Place any executable
with a name like `k3-jaegerschnitzel` in your `PATH` and k3 will automatically
make it available as a subcommand. You can then call it just like any other k3
command:

    $ k3 jaegerschnitzel +6000


**Environment Variables**

Manifest values can be overwritten by environment variables. Environment
variables will always take precedence over regular variables. Regular variables
have to be declared in a variables file `k3.env`.

Environment variable `K3_CACHE` is a colon-separated list of directories and has
similar purpose and behaviour like GNU Make's VPATH. It is use to find files
like `k3.env`.


## TTCN-3 Language Server

We are currently implementing the [Language Server Protocol](https://microsoft.github.io/language-server-protocol):

**Features**

| Name                        | Method                            |                    |
| --------------------------- | --------------------------------- | ------------------ |
| Workspace Symbols           | `workspace/symbol`                | :x:                |
| Execute Command             | `workspace/executeCommand`        | :x:                |
| Diagnostics                 | `textDocument/publishDiagnostics` | :x:                |
| Completion                  | `textDocument/completion`         | :heavy_check_mark: |
| Hover                       | `textDocument/hover`              | :x:                |
| Signature Help              | `textDocument/signatureHelp`      | :x:                |
| Goto Definition             | `textDocument/definition`         | :x:                |
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
work. This is the most common issue of k3 language server that we see.

Unfortunately there isn't much you can do, yet. we are currently working on
diagnostics and additional debug messages, so we can figure out what relevant
use cases for editing test suites are.


**Visual Studio Code**

For TTCN-3 support in Visual Studio Code have a look at our co-project
[vscode-ttcn3](https://marketplace.visualstudio.com/items?itemName=Nokia.ttcn3)

**Vim**

Use [prabirshrestha/vim-lsp], with the following configuration:

```vim
augroup LspTTCN3
  au!
  autocmd User lsp_setup call lsp#register_server({
      \ 'name': 'k3',
      \ 'cmd': {server_info->['k3', 'langserver']},
      \ 'whitelist': ['ttcn3'],
      \ })
  autocmd FileType ttcn3 setlocal omnifunc=lsp#complete
  autocmd FileType ttcn3 nmap <buffer> gd <plug>(lsp-definition)
  "autocmd FileType ttcn3 nmap <buffer> ,n <plug>(lsp-next-error)
  "autocmd FileType ttcn3 nmap <buffer> ,p <plug>(lsp-previous-error)
augroup END
```


## Future Development

Our goal is to improve the TTCN-3 user experience by providing modern tools. We
will focus on adding features to the language server.  
Of course you are very welcome to contribute. Check the
[feature requests](https://github.com/nokia/ntt/labels/enhancement) if something suits
you or create your own feature request.
