# golicense - Go Binary OSS License Scanner

golicense is a tool that scans [compiled Go binaries](https://golang.org/)
and can output all the dependencies, their versions, and their respective
licenses (if known). golicense only works with Go binaries compiled using
Go modules for dependency management.

golicense determines the dependency list quickly and with exact accuracy
since it uses metadata from the Go compiler to determine the _exact_ set of
dependencies embedded in a compiled Go binary. This excludes dependencies that
are not used in the final binary. For example, if a library depends on "foo"
in function "F" but "F" is never called, then the dependency "foo" will not
be present in the final binary.

golicense is not meant to be a complete replacement for open source compliance
companies such as [FOSSA](https://fossa.io/) or
[BlackDuck](https://www.blackducksoftware.com/black-duck-home), both of
which provide hundreds of additional features related to open source
compliance.

**Warning:** The binary itself must be trusted and untampered with to provide
accurate results. It is trivial to modify the dependency information of a
compiled binary. This is the opposite side of the same coin with source-based
dependency analysis where the source must not be tampered.

## Features

  * List dependencies and their associated licenses
  * Cross-reference dependency licenses against an allow/deny list
  * Output reports in the terminal and Excel (XLSX) format
  * Manually specify overrides for specific dependencies if the detection
    is incorrect.

## Example

The example below runs `golicense` against itself from a recent build.

![golicense Example](https://user-images.githubusercontent.com/1299/48667166-468d1080-ea85-11e8-8005-5a44c6a0d10a.gif)

## Installation

To install `golicense`, download the appropriate release for your platform
from the [releases page](https://github.com/cloudentity/golicense/releases).

You can also compile from source using Go 1.11 or later using standard
`go build`. Please ensure that Go modules are enabled (GOPATH not set or
`GO111MODULE` set to "on").

## Usage

`golicense` is used with one or two required arguments. In the one-argument
form, the dependencies and their licenses are listed. In the two-argument
form, a configuration file can be given to specify an allow/deny list of
licenses and more.

```
$ golicense [flags] [BINARY]
$ golicense [flags] [CONFIG] [BINARY]
```

You may also pass mutliple binaries (but only if you are providing a CONFIG).

### Configuration File

The configuration file can specify allow/deny lists of licenses for reports,
license overrides for specific dependencies, and more. The configuration file
format is [HCL](https://github.com/hashicorp/hcl2) or JSON.

Example:

```hcl
allow = ["MIT", "Apache-2.0"]
deny  = ["GNU General Public License v2.0"]
```

```json
{
  "allow": ["MIT", "Apache-2.0"],
  "deny": ["GNU General Public License v2.0"]
}
```

Supported configurations:

  * `allow` (`array<string>`) - A list of names or SPDX IDs of allowed licenses.
  * `deny` (`array<string>`) - A list of names or SPDX IDs of denied licenses.
  * `override` (`map<string, string>`) - A mapping of Go import identifiers
    to translate into a specific license by SPDX ID. This can be used to
	set the license of imports that `golicense` cannot detect so that reports
	pass. It's also possible to provide a name if the license has no SPDX ID.
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

### Excel (XLSX) Reporting Output

If the `-out-xlsx` flag is specified, then an Excel report is generated
and written to the path specified in addition to the terminal output.

```
$ golicense -out-xlsx=report.xlsx ./my-program
```

The Excel report contains the list of dependencies, their versions, the
detected license, and whether the license is allowed or not. The dependencies
are listed in alphabetical order. The row of the dependency will have a
green background if everything is okay, a yellow background if a
license is unknown, or a red background is a license is denied. An example
screenshot is shown below:

![Excel Report](https://user-images.githubusercontent.com/1299/48667086-84893500-ea83-11e8-925c-7929ed441b1b.png)


### SBOM Reporting Output

It's possible to generate report as [SBOM - software bill-of-material](https://cyclonedx.org/)

```
$ golicense -out-sbom=report.xml ./my-program

```

Sample output:

```xml
<bom xmlns="http://cyclonedx.org/schema/bom/1.1" version="1" serialNumber="urn:uuid:16d113cb-029e-4ad0-bd68-c4407c6ce285">
  <components>
    <component type="library">
      <name>github.com/Jeffail/gabs</name>
      <version>v2.5.0</version>
      <purl>pkg:golang/github.com/Jeffail/gabs@2.5.0</purl>
      <licenses>
        <license>
          <id>MIT</id>
        </license>
      </licenses>
    </component>
  </components>
</bom>
```

Additionally if a licence cannot be resolved, it's possible to set custom licence name and optionally provide a url.

Example:

```json
{
  "override": {
    "github.com/krolaw/zipstream": "Custom"
  },
  "sbomLicenseURLs": {
    "github.com/krolaw/zipstream": "https://github.com/krolaw/zipstream/blob/master/LICENSE"
  }
}
```

Output:

```xml
<bom xmlns="http://cyclonedx.org/schema/bom/1.1" version="1" serialNumber="urn:uuid:16d113cb-029e-4ad0-bd68-c4407c6ce285">
  <component type="library">
    <name>github.com/krolaw/zipstream</name>
    <version>v0.0.0-20180621105154-0a2661891f94</version>
    <purl>pkg:golang/github.com/krolaw/zipstream@0.0.0-20180621105154-0a2661891f94</purl>
    <licenses>
      <license>
        <name>Custom</name>
        <url>https://github.com/krolaw/zipstream/blob/master/LICENSE</url>
      </license>
    </licenses>
  </component> 
</bom>
```


## Limitations

There are a number of limitations to `golicense` currently. These are fixable
but work hasn't been done to address these yet. If you feel like taking a stab
at any of these, please do and contribute!

**GitHub API:** The license detected by `golicense` may be incorrect if
a GitHub project changes licenses. `golicense` uses the GitHub API which only
returns the license currently detected; we can't lookup licenses for specific
commit hashes.
