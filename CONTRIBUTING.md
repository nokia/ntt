# Contributing to NTT

:+1::tada: First off, thanks for taking the time to contribute! :tada::+1:

We welcome contributions to this project of any kind including documentation, plugins,
organization, tutorials, bug reports, issues, feature requests, feature
implementations, pull requests, answering questions on the mailing list, helping
to manage issues, etc.

The first contribution can be scary, but there's no need to worry. Everyone has
to start somewhere and we are all nice people who find it great when someone is
interested in our project.


**Reporting Issues**

If you believe you have found an issue in ntt, please use the GitHub issue tracker
to report the problem. If you're not sure if it's a bug or not, start by asking.


**Asking Questions**

Our mailing list [ntt@groups.io](mailto:ntt@groups.io) and [GitHub
Discussions](https://github.com/nokia/ntt/discussions) are great places for
asking questions and having discussions.

## Contributing Code

Use [GitHub pull requests](https://docs.github.com/en/get-started/quickstart/github-flow) to contribute code.

If you are doing this the first time you'll probably have to setup your
environment. There are many options, you can use [GitHub
Codespaces](https://docs.github.com/en/codespaces) and develop in your browser
remotely. Or you install everything on your machine and develop locally:

**Installing Go**

Install Go using your package manager or a manual installation as [described
here](https://go.dev/doc/install).  
You don't always need the latest version of Go, but you should not use versions
older than two years either.


**Installing Visual Studio Code**

We suggest you install [Visual Studio Code](https://code.visualstudio.com/docs).  
This is not a hard requirement, you can develop in any editor you like. But we
find Visual Studio Code is the easiest to start with, especially when you are
developing with a Microsoft Windows System.

**Installing Git**

We use Git for version control and collaborative development. Make sure you
[Git installed](https://github.com/git-guides/install-git).

**What you should know**

This project's main programming language is Go. If you are new to Go, you'll
find an introduction [here](https://go.dev/learn/).
If you are not sure about the coding style have a look at [Effective Go](https://go.dev/doc/effective_go)
and [Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

Some basic understanding how to use Git will of great use.

0. [Fork][fork] and clone the repository
0. Create a new branch: `git checkout -b my-branch-name`
0. Make your change and remember to add tests
0. Build the project locally and run local tests.
0. Push to your fork and [submit a pull request][pr]
0. Pat your self on the back and wait for your pull request to be reviewed and merged.

Here are a few things you can do that will increase the likelihood of your pull
request being accepted:

* Write tests.
* Format your code.
* Keep your change as focused as possible. If there are multiple changes you
  would like to make that are not dependent upon each other, submit them as
  separate pull requests.
* Write [good commit messages](https://chris.beams.io/posts/git-commit/).


## Issue and Pull Request Labels

Labels help us track and manage issues and pull requests.

| Label Name         |                                    | Description
| ------------------ | ---------------------------------- | -----------
| `enhancement`      | [search][search-label-enhancement] | Feature Requests
| `bug`              | [search][search-label-bug]         | Something isn't working
| `duplicate`        | [search][search-label-duplicate]   | This issue or pull request already exists
| `good first issue` | [search][search-label-first]       | Good for newcomers
| `help wanted`      | [search][search-label-help]        | Extra attention is needed
| `invalid`          | [search][search-label-invalid]     | This doesn't seem right
| `question`         | [search][search-label-question]    | Further information is requested
| `wontfix`          | [search][search-label-wontfix]     | This will not be worked on

## Releasing

Besides source we also provide pre-built binaries. Those binary releases are
built using [GoReleaser](https://goreleaser.com/). 

Everything is automated, if you want to release ntt, just push a git-tag to
this repository. Have a look at existing tags to see how we name things.

When your git-tag went through CI successfully, you'll find a new draft in the
[releases section](https://github.com/nokia/ntt/releases). Edit this draft,
select your tag and write some nice release notes.  

Release notes should be relevant to our users:

* Describe what the new feature is used for and what's problems is solves.
* Add screenshots or screencasts to clarify.
* When breaking compatibility explain why and how the user can fix resulting
  issues.
* Give shoutouts to our contributors, because we are a community!

Again, have a look at previous releases to get some inspiration what to write.

**Dry run**
If you want to release ntt _manually_, install
[goreleaser](https://goreleaser.com/) and try a dry-run first:

	$ goreleaser --snapshot --skip-publish --rm-dist


## Resources

* [How to Contribute to Open Source](https://opensource.guide/how-to-contribute/)
* [Using Pull Requests](https://help.github.com/articles/about-pull-requests/)
* [GitHub Help](https://help.github.com)


[fork]: https://github.com/nokia/ntt/fork
[pr]: https://github.com/nokia/ntt/compare
[code-of-conduct]: CODE_OF_CONDUCT.md
[search-label-enhancement]: https://github.com/nokia/ntt/labels/enhancement
[search-label-duplicate]: https://github.com/nokia/ntt/labels/duplicate
[search-label-first]: https://github.com/nokia/ntt/labels/good%20first%20issue
[search-label-help]: https://github.com/nokia/ntt/labels/help%20wanted
[search-label-invalid]: https://github.com/nokia/ntt/labels/invalid
[search-label-question]: https://github.com/nokia/ntt/labels/question
[search-label-wontfix]: https://github.com/nokia/ntt/labels/wontfix
