# golicense - Go Binary OSS License Scanner

golicense is a tool that scans [compiled Go binaries](https://golang.org/)
and can output all the dependencies, their versions, and their respective
licenses (if known).

golicense is fast and extremely accurate since it uses the metadata from
the Go compiler to determine the _exact_ set of dependencies embedded in
a compiled Go binary. This requires the use of Go modules in Go 1.11 and later.
Binaries compiled without Go modules will not work.

**Warning:** The binary itself must be trusted and untampered with to provide
accurate results. It is trivial to modify the dependency information of a
compiled binary. This is the opposite side of the same coin with source-based
dependency analysis where the source must not be tampered.

## Example

```
$ golicense ./vault
TODO
```

## Features

TODO
