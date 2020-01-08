# Contributing

Thanks for taking the time to join our community and start contributing.
These guidelines will help you get started with the ir2proxy project.
Please note that we require [DCO sign off](#dco-sign-off).  

## Building from source

This section describes how to build ir2proxy from source.

### Prerequisites

1. *Install Go*

    ir2proxy requires [Go 1.13][1] or later.
    We also assume that you're familiar with Go's [`GOPATH` workspace][3] convention, and have the appropriate environment variables set.

### Fetch the source

ir2proxy uses [`go modules`][2] for dependency management. 

In order to make PRs, however, you'll need to make your own fork of the ir2proxy repo. A suggestion on how to do that is below, which will place the code in your `$GOPATH` and set you up so you can `git pull` on master and get the upstream master.

#### Source setup suggestion

Clone the ir2proxy repo into `$GOPATH/github.com/projectcontour/ir2proxy`.

```
go get github.com/projectcontour/ir2proxy
```

Fix `origin` to your fork:

```
git remote rename origin upstream
git remote add origin git@github.com:youngnick/ir2proxy.git
```

This ensures that the source code on disk remains at `$GOPATH/src/github.com/projectcontour/ir2proxy` while the remote repository is configured for your fork.

The remainder of this document assumes your terminal's working directory is the repo root.

### Building

To build ir2proxy, run:

```
make
```

This uses `go install` and produces a `ir2proxy` binary in your `$GOPATH/bin` directory.

### Running the unit tests

Once you have ir2proxy building, you can run all the unit tests for the project:

```
make check
```

This assumes your working directory is set to `$GOPATH/src/github.com/projectcontour/ir2proxy`.

To run the tests for a single package, change to package directory and run:

```
go test .
```

## Contribution workflow

This section describes the process for contributing a bug fix or new feature.
It follows from the previous section, so if you haven't set up your Go workspace and built ir2proxy from source, do that first.

### Before you submit a pull request

This project operates according to the _talk, then code_ rule.
If you plan to submit a pull request for anything more than a typo or obvious bug fix, first you _should_ [raise an issue][5] to discuss your proposal, before submitting any code.

It's likely that most of the contributions required will also require new tests.
`ir2proxy` is currently at 100% test coverage for the internal code, let's keep it like that.
See the the [translator testing document](internal/translator/TESTING.md) and [validate testing document](internal/validate/TESTING.md) for details.

### Commit message and PR guidelines

- Have a short subject on the first line and a body. The body can be empty.
- Use the imperative mood (ie "If applied, this commit will (subject)" should make sense).
- There must be a DCO line ("Signed-off-by: David Cheney <cheneyd@vmware.com>"), see [DCO Sign Off](#dco-sign-off) below
- Put a summary of the main area affected by the commit at the start,
with a colon as delimiter. For example 'docs:', 'internal/(packagename):', 'design:' or something similar.
- Try to keep your number of commits in a PR low. Generally we
tend to squash before opening the PR, then have PR feedback as
extra commits.
- Do not merge commits that don't relate to the affected issue (e.g. "Updating from PR comments", etc). Should
the need to cherrypick a commit or rollback arise, it should be clear what a specific commit's purpose is.
- If master has moved on, you'll need to rebase before we can merge,
so merging upstream master or rebasing from upstream before opening your
PR will probably save you some time.
- PRs *must* include a `Fixes #NNNN` or `Updates #NNNN` comment. Remember that
`Fixes` will close the associated issue, and `Updates` will link the PR to it.

#### Commit message template

```
<packagename>: <imperative mood short description>

Updates #NNNN
Fixes #MMMM

Signed-off-by: Your Name you@youremail.com

<longer change description/justification>

```

#### Sample commit message

```
internal\translator: Add quux functions

Fixes #xxyyz

Signed-off-by: Your Name you@youremail.com

To implement the quux functions from #xxyyz, we need to
florble the greep dots, then ensure that the florble is
warbed.
```

### Pre commit CI

Before a change is submitted it should pass all the pre commit CI jobs. (That is, `make check`.)
If there are unrelated test failures the change can be merged so long as a reference to an issue that tracks the test failures is provided.

## DCO Sign off

All authors to the project retain copyright to their work. However, to ensure
that they are only submitting work that they have rights to, we are requiring
everyone to acknowledge this by signing their work.

Any copyright notices in this repository should specify the authors as "The
project authors".

To sign your work, just add a line like this at the end of your commit message:

```
Signed-off-by: Nick Young <ynick@vmware.com>
```

This can easily be done with the `--signoff` option to `git commit`.

By doing this you state that you can certify the following (from https://developercertificate.org/):

```
Developer Certificate of Origin
Version 1.1

Copyright (C) 2004, 2006 The Linux Foundation and its contributors.
1 Letterman Drive
Suite D4700
San Francisco, CA, 94129

Everyone is permitted to copy and distribute verbatim copies of this
license document, but changing it is not allowed.


Developer's Certificate of Origin 1.1

By making a contribution to this project, I certify that:

(a) The contribution was created in whole or in part by me and I
    have the right to submit it under the open source license
    indicated in the file; or

(b) The contribution is based upon previous work that, to the best
    of my knowledge, is covered under an appropriate open source
    license and I have the right under that license to submit that
    work with modifications, whether created in whole or in part
    by me, under the same open source license (unless I am
    permitted to submit under a different license), as indicated
    in the file; or

(c) The contribution was provided directly to me by some other
    person who certified (a), (b) or (c) and I have not modified
    it.

(d) I understand and agree that this project and the contribution
    are public and that a record of the contribution (including all
    personal information I submit with it, including my sign-off) is
    maintained indefinitely and may be redistributed consistent with
    this project or the open source license(s) involved.
```

[1]: https://golang.org/dl/
[2]: https://github.com/golang/go/wiki/Modules
[3]: https://golang.org/doc/code.html
[4]: https://developercertificate.org/
[5]: https://github.com/projectcontour/ir2proxy/issues
