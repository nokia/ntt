# Getting started with ntt

This walkthrough gets you from "I just installed ntt" to running a
TTCN-3 test and using the LSP features in your editor in under five
minutes. It assumes you already have ntt on your PATH; see the
[install instructions](../README.md#install) if you don't.

## 1. Create a test project

Pick a working directory and drop in a single test file. We'll use
`/tmp/hello-ntt` for this example - swap in any path you like, but make
sure the path you pick has no spaces in it (a Windows-friendly
restriction).

```sh
mkdir -p /tmp/hello-ntt
cd /tmp/hello-ntt
```

Create a `package.yml` so ntt recognises this folder as a project root:

```yaml
# /tmp/hello-ntt/package.yml
name: hello
source_dir: .
```

Create the test file itself:

```ttcn3
// /tmp/hello-ntt/Hello.ttcn3
module Hello {
    function add(integer a, integer b) return integer {
        return a + b;
    }

    testcase TC_Add() runs on system {
        if (add(2, 3) == 5) {
            setverdict(pass);
        } else {
            setverdict(fail);
        }
    }

    control {
        execute(TC_Add());
    }
}
```

## 2. List discovered tests

```sh
ntt list
```

You should see:

```
Hello.TC_Add
```

If you don't, double-check that you ran `ntt` from inside the project
directory and that `package.yml` is in the same folder as the `.ttcn3`
file.

## 3. Run the test

```sh
ntt run
```

ntt executes the `control` block, which in turn calls our testcase, and
reports a pass verdict:

```
=== RUN   Hello.TC_Add
=== PASS  Hello.TC_Add  0.001s
```

If you change `add(2, 3) == 5` to `add(2, 3) == 6` and re-run, you'll
get a failing verdict and a non-zero exit status - exactly what a CI
pipeline needs.

## 4. Format the code

```sh
ntt format Hello.ttcn3
```

This normalises whitespace and reflows long lines while preserving
behaviour. Use `--diff` to preview the changes without writing them
back to disk.

## 5. Lint the project

```sh
ntt lint
```

ntt's built-in linter catches common mistakes (missing default cases,
unused imports, suspicious type coercions). Configure it via the
`lint:` section of `package.yml`; the defaults are sensible for a quick
start.

## 6. Hook up your editor

Both [VS Code](https://marketplace.visualstudio.com/items?itemName=Nokia.ttcn3)
and [vim-lsp-settings](https://github.com/mattn/vim-lsp-settings)
auto-detect ntt as the TTCN-3 language server. Open the project folder
and the editor lights up with:

- semantic highlighting and inlay hints
- diagnostics on save
- jump to definition (including jumps from TTCN-3 into ASN.1 files)
- format-on-save via `ntt format`
- code actions (`source.organizeImports`, lint quick-fixes)

You can also point any LSP-aware editor at the binary by configuring
it to launch:

```sh
ntt langserver
```

## 7. Working with Titan projects

If you already have a `*.tpd` file from
[Eclipse Titan](https://projects.eclipse.org/projects/tools.titan),
ntt can read it directly. Run any ntt command from a directory that
contains a `.tpd` (and no `package.yml`) and ntt will discover the
descriptor automatically. You can also point at it explicitly:

```sh
ntt list path/to/myproject.tpd
```

Referenced sub-projects (`<ReferencedProject>` in the XML) are loaded
transitively, so the same command works for multi-module Titan
projects.


## 8. Where to go next

- Browse the [ARCHITECTURE.md](../ARCHITECTURE.md) doc to understand
  how ntt is laid out internally.
- Run `ntt help` for the full command list - `run`, `list`, `tags`,
  `show`, `report`, `cover`, and others.
- Look at `examples/cmake/` for a CMake-driven integration that builds
  test adapters together with the TTCN-3 code.
- File feature requests and bug reports on
  [GitHub issues](https://github.com/nokia/ntt/issues).
