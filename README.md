# NTT - A framework for working with TTCN-3 

[TTCN-3](http://www.ttcn-3.org/) (Testing and Test Control Notation Version 3)
is a standardized test specification language that applies to a variety of
application domains and types of testing. It has been used since 2000 in
standardization as well as in industry, research, international projects and
academia.

Unfortunately tool support for TTCN-3 is quite antiquated, which makes
maintaining big TTCN-3 code bases difficult and developing new TTCN-3 tools
expensive.

This project aims to provide maintainable tools and libraries, helping testers
and developers not to reinvent the TTCN-3 wheel.

Please note, NTT is still much work in progress. First task is to provide a
TTCN-3 parser, which supports core language specification as well as behaviour
types and advanced parametrization. Only after that task is finished, we can
start with fancy tools, like:

  * Wide range IDE support, using [Language Server Protocol](https://microsoft.github.io/language-server-protocol/)
  * Analytics and visualization (cyclomatic complexity, coverage, ...)
  * Formatting and refactoring tools
  * Code generators for other backends (go, erlang, llvm, titan, ...)
  * Support for other IDLs (ASN.1, Protobuf, IDL, ...)
  * Frontends for configuration-management, monitoring, ...

Documentation and functionality will grow over time. If you are interested in
contributing, just create a pull-request. A contribution guide will be provided
when the TTCN-3 parser is released.


# Getting Started

If you want to dive into TTCN-3 right now, have a look at [Titan](https://github.com/eclipse/titan.core/),
which already has a lot of features and a huge set of adapters to choose from.


I am currently developing the TTCN-3 parser library, so there isn't much to see,
yet. Assuming you have Go installed you can try out the parser tool by using the
`go get` method:

        go get -u github.com/nokia/ntt/ttcn3/cmd/ttcn3-parser

Please note, the parser might end in an infinite loop, as I am currently
refactoring expression-handling.
