# golicense - Go Binary OSS License Scanner

golicense is a tool that scans [compiled Go binaries](https://golang.org/)
and can output all the dependencies, their versions, and their respective
licenses (if known).

golicense determines the dependency list quickly and with exact accuracy
since it uses metadata from the Go compiler to determine the _exact_ set of
dependencies embedded in a compiled Go binary. This excludes dependencies that
are not used in the final binary. For example, if a library depends on "foo"
in function "F" but "F" is never called, then the dependency "foo" will not
be present in the final binary.

**Warning:** The binary itself must be trusted and untampered with to provide
accurate results. It is trivial to modify the dependency information of a
compiled binary. This is the opposite side of the same coin with source-based
dependency analysis where the source must not be tampered.

## Features

  * List dependencies, their versions, and their checksum
  * Find and list the license associated with a dependency
  * Cross-reference dependency licenses against an allow/deny list to
    generate compliance reports.

## Example

The example below runs `golicense` against itself from a recent build.

```
$ golicense ./golicense
github.com/agext/levenshtein         Apache License 2.0
github.com/apparentlymart/go-textseg Other
github.com/davecgh/go-spew           ISC License
github.com/fatih/color               MIT License
github.com/google/go-cmp             BSD 3-Clause "New" or "Revised" License
github.com/google/go-github          BSD 3-Clause "New" or "Revised" License
github.com/google/go-querystring     BSD 3-Clause "New" or "Revised" License
github.com/gosuri/uilive             MIT License
github.com/hashicorp/errwrap         Mozilla Public License 2.0
github.com/hashicorp/go-multierror   Mozilla Public License 2.0
github.com/hashicorp/hcl2            Mozilla Public License 2.0
github.com/mattn/go-colorable        MIT License
github.com/mattn/go-isatty           MIT License
github.com/mitchellh/go-wordwrap     MIT License
github.com/pmezard/go-difflib        Other
github.com/rsc/goversion             BSD 3-Clause "New" or "Revised" License
github.com/stretchr/objx             MIT License
github.com/stretchr/testify          Other
github.com/zclconf/go-cty            MIT License
golang.org/x/crypto                  Other
golang.org/x/net                     Other
golang.org/x/oauth2                  BSD 3-Clause "New" or "Revised" License
golang.org/x/sys                     Other
golang.org/x/text                    Other
```

## Usage

`golicense` is used with one or two required arguments. In the one-argument
form, the dependencies and their licenses are listed. In the two-argument
form, a configuration file can be given to specify an allow/deny list of
licenses and more.

```
$ golicense [flags] [BINARY]
$ golicense [flags] [CONFIG] [BINARY]
```

### Configuration File

The configuration file can specify allow/deny lists of licenses for reports,
license overrides for specific dependencies, and more. The configuration file
format is [HCL](https://github.com/hashicorp/hcl2) or JSON.

Example:

```hcl
allow = ["MIT", "Apache-2.0"]
deny  = ["GNU General Public License v2.0"]
```

Supported configurations:

  * `allow` (`array<string>`) - A list of names or SPDX IDs of allowed licenses.
  * `deny` (`array<string>`) - A list of names or SPDX IDs of denied licenses.
  * `override` (`map<string, string>`) - A mapping of Go import identifiers
    to translate into a specific license by SPDX ID. This can be used to
	set the license of imports that `golicense` cannot detect so that reports
	pass.
  * `translate` (`map<string, string>`) - A mapping of Go import identifiers
    to translate into alternate import identifiers. Example:
	"gopkg.in/foo/bar.v2" to "github.com/foo/bar". If the map key starts and
	ends with `/` then it is treated as a regular expression. In this case,
	the map value can use `\1`, `\2`, etc. to reference capture groups.

### GitHub Authentication

`golicense` uses the GitHub API to look up licenses. This doesn't require
any authentication out of the box but will be severely rate limited.
It is recommended that you generate a [personal access token](https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/) to increase the rate limit. The personal access token doesn't require any
special access unless it needs to look at private repositories you have
access to, in which case it should be granted the `repo` permission.
Specify your token using the `GITHUB_TOKEN` environment variable.

```
$ export GITHUB_TOKEN=abcd1234
$ golicense ./binary
```

## Limitations

There are a number of limitations to `golicense` currently. These are fixable
but work hasn't been done to address these yet. If you feel like taking a stab
at any of these, please do and contribute!

**GitHub API:** The license detected by `golicense` may be incorrect if
a GitHub project changes licenses. `golicense` uses the GitHub API which only
returns the license currently detected; we can't lookup licenses for specific
commit hashes.

**Import Redirects:** Import paths that redirect to GitHub projects
(such as `gonum.org/v1/gonum`) aren't properly translated currently. To fix
this we should use the `go get` HTTP protocol to detect these and do the
proper translation. For now, you can work around this using explicit overrides
via a configuration file.
