pkg
===

`pkg` is a collection of Go packages that provide various different functionality. The packages in this project are
generally independent and strive to depend on as few external packages as possible. Many of the packages in the project
are also independent from each other.

An overview of some of the packages is included below. Refer to the package comments for individual packages for more
information.

cli
---
`cli` provides a framework for creating CLI applications. It provides support for commands, subcommands, flags, before
and after hooks, documentation, deprecation, command-line completion and other functionality.

matcher
-------
`matcher` allows files to be matched based on their name or path. Supports composing and combining matchers and provides
data structures that can be used as configuration to specify include and exclude rules.

pkgpath
-------
`pkgpath` provides functions for getting Go package paths. Provides functions for getting the paths to all of the
packages rooted in a directory and for converting between different representations of package paths including relative,
`GOPATH`-relative and absolute. Depends on `matcher`.

specdir
-------
`specdir` provides the ability to define specifications for directory layouts, verify that existing directories match
the specification and create new directory structures based on a specification.

License
-------
This project is made available under the [Apache 2.0 License](http://www.apache.org/licenses/LICENSE-2.0).
