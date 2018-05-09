# NTT - A framework for working with TTCN-3 

[TTCN-3](etsi-homepage) (Testing and Test Control Notation Version 3) is
a standardized test specification language that applies to a variety of
application domains and types of testing. It has been used since 2000 in
standardization as well as in industry, research, international projects and
academia.

[etsi-homepage]: http://www.ttcn-3.org/

Unfortunately tool support for TTCN-3 is quite antiquated, which makes
maintaining big TTCN-3 code bases difficult and developing new TTCN-3 tools
expensive.

This project aims to provide maintainable tools and libraries, helping testers
and developers not to reinvent the TTCN-3 wheel.

Please note, NTT is still much work in progress. First task is to provide a
TTCN-3 parser, which supports core language specification as well as behaviour
types and advanced parametrization. Only after that task is finished, we can
start with fancy tools, like:

  * Wide range IDE support, using [Language Server Protocol](lsp-homepage)
  * Analytics and visualization (cyclomatic complexity, coverage, ...)
  * Formatting and refactoring tools
  * Code generators for other backends (go, erlang, llvm, titan, ...)
  * Support for other IDLs (ASN.1, Protobuf, IDL, ...)
  * Frontends for configuration-management, monitoring, ...

Documentation and functionality will grow over time. If you are interested in
contributing, just create a pull-request. A contribution guide will be provided
when the TTCN-3 parser is released.

[lsp-homepage]: https://microsoft.github.io/language-server-protocol/


# Getting Started

I am currently developing the TTCN-3 parser library, so there isn't much to see,
yet. Assuming you have Go installed you can try out `ttcn3-parser` by using the
`go get` method:

        go get -u github.com/nokia/ntt/ttcn3/cmd/ttcn3-parser

Please note, the parser might end in an infinite loop, as I am currently
refactoring expression-handling.
