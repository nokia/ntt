---
nav_order: 4
nav_exclude: true
---

# Best Practices

* ntt is a buildsystem (test executable, adapters, ...): introduce convention

Because TTCN-3 standard doesn't dictate how to configure or build a test suite,
developers need to know a great deal about internals: How to start the system
under test? Where to store common environment variables? How to build adapters?
This is were ntt comes in. It provides a stable frame so developers can to
focus on writing tests.


* pass extra information: setverdict(pass, “UE bearer setup completed”); ?
* use 'none' verdict
* use 'inconc' for timeouts
* use extra information for 'fail'

