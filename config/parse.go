package config

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hcl/hclsyntax"
	"github.com/hashicorp/hcl2/hcl/json"
)

// ParseFile parses the given file for a configuration. The syntax of the
// file is determined based on the filename extension: "hcl" for HCL,
// "json" for JSON, other is an error.
func ParseFile(filename string) (*Config, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	ext := filepath.Ext(filename)
	if len(ext) > 0 {
		ext = ext[1:]
	}

	return Parse(f, filename, ext)
}

// Parse parses the configuration from the given reader. The reader will be
// read to completion (EOF) before returning so ensure that the reader
// does not block forever.
//
// format is either "hcl" or "json"
func Parse(r io.Reader, filename, format string) (*Config, error) {
	switch format {
	case "hcl":
		return parseHCL(r, filename)

	case "json":
		return parseJSON(r, filename)

	default:
		return nil, fmt.Errorf("format must be either 'hcl' or 'json' (%s / %s)", filename, format)
	}
}

func parseHCL(r io.Reader, filename string) (*Config, error) {
	src, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	f, diag := hclsyntax.ParseConfig(src, filename, hcl.Pos{})
	if diag.HasErrors() {
		return nil, diag
	}

	var config Config
	diag = gohcl.DecodeBody(f.Body, nil, &config)
	if diag.HasErrors() {
		return nil, diag
	}

	return &config, nil
}

func parseJSON(r io.Reader, filename string) (*Config, error) {
	src, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	f, diag := json.Parse(src, filename)
	if diag.HasErrors() {
		return nil, diag
	}

	var config Config
	diag = gohcl.DecodeBody(f.Body, nil, &config)
	if diag.HasErrors() {
		return nil, diag
	}

	return &config, nil
}
