---
layout: default
title: About
nav_order: 1
description: "ntt is a modern toolset for language agnostic testing with TTCN-3. It provides a language server, code generators and much more."
permalink: /
---

# Modern tools for TTCN-3
{: .fs-9 }

ntt gives you the tools you need to write tests efficiently.
{: .fs-6 .fw-300 }

[IDE Support](editors){: .btn .btn-primary .fs-5 .mb-4 .mb-md-0 .mr-2 }
[Get started now](getting-started){: .btn .fs-5 .mb-4 .mb-md-0 }

---

**ntt** is a free and open application stack for testing with
[TTCN-3](#whats-ttcn-3).
<a name="k3"></a>  It's the incarnation of its closed source predecessor k3. It
builds upon 15 years of experience of running production workloads at Nokia.
ntt is designed with low latency and high fidelity appliances in mind.  Current
releases include:

* a [TTCN-3 language server](editors) for a better IDE experience.
* comfortable [tools](getting-started) supporting Nokia [best practices](getting-started) for test
  suite configuration and analysis.
* a CMake module for easier [build system integration](getting-started#cmake).
* and [more to come](#roadmap) ...

**ntt** is still a young project, so we do not provide a TTCN-3 compiler,
runtime or any open ASN.1 tools, yet. If you need such, we recommend
[Titan](https://github.com/eclipse/titan.core)! Titan offers a mature compiler
and runtime with an extensive set of adapters and modules.


## What's TTCN-3?

[TTCN-3](http://www.ttcn-3.org/) is a standardized domain-specific language for
scripting tests. It supports various types of testing and is a good addition
to unit tests. Despite being clearly from the last millennium, TTCN-3 offers
some excellent features for testing protocols and services:

* native support for concurrency and alternate code flows (similar to Go).
* a powerful matching mechanism. Think supercharged regular expressions for
  arbitrary data types.
* and it's language agnostic: Adapters and codecs are implemented in any
  language you see fit. So you can leverage any ecosystem you need.


## What's the project status?

**ntt** is used in production by Nokia in 4G and 5G. But still, consider this
software unstable. Like in every young project features need time to settle and
not all parts have the same test coverage, yet. ntt is no exception. Also,
consider ntt experimental as we are exploring new ways of working with TTCN-3.


**ntt** has full support for TTCN-3 version 4.10.1 and understands most parts
of the standardized extensions for documentation comments, performance and
real-time testing, advanced parameterization, behaviour types, as well as
various extensions from [Titan](https://github.com/eclipse/titan.core) and
[k3](#k3).


## Roadmap

Our major goal is to improve the TTCN-3 user experience by providing modern
tools and by establishing best practices for working with TTCN-3 test suites.
We will also continue updating ntt for the latest TTCN-3 standards and will be
adding one or the other feature.  
But our focus is on advanced [IDE support](editors) first and your feedback can
help us to achieve this goal.

We plan on finishing goto-definition first and then we'll implement initial
diagnostics.


## Call for Contributors

As a new project, we have plenty of room to grow and many plans we want to
execute. If you want to report a bug, or if you are a developer, tester,
designer, DevOps, documentation enthusiast or think you can and want to join
ntt, we want you with us! Read more about becoming a contributor in [our
contribution guide](https://github.com/nokia/ntt/blob/master/CONTRIBUTING.md).


## Contact us

If you have any questions you are very welcome to contact us at
[ntt@groups.io](mailto:ntt@groups.io).

## License

ntt is distributed by an [BSD 3-clause license](https://github.com/nokia/ntt/blob/master/LICENSE).
