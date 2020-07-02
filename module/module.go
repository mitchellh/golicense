package module

import (
	"fmt"
	"regexp"
	"strings"
)

// Module represents a single Go module.
//
// Depending on the source that this is parsed from, fields may be empty.
// All helper functions on Module work with zero values. See their associated
// documentation for more information on exact behavior.
type Module struct {
	Path    string // Import path, such as "github.com/cloudentity/golicense"
	Version string // Version like "v1.2.3"
	Hash    string // Hash such as "h1:abcd1234"
}

// String returns a human readable string format.
func (m *Module) String() string {
	return fmt.Sprintf("%s (%s)", m.Path, m.Version)
}

// ParseExeData parses the raw dependency information from a compiled Go
// binary's readonly data section. Any unexpected values will return errors.
func ParseExeData(raw string) ([]Module, error) {
	var result []Module
	for _, line := range strings.Split(strings.TrimSpace(raw), "\n") {
		row := strings.Split(line, "\t")

		// Ignore non-dependency information, such as path/mod. The
		// "=>" syntax means it is a replacement.
		if row[0] != "dep" && row[0] != "=>" {
			continue
		}

		if len(row) == 3 {
			// A row with 3 can occur if there is no hash data for the
			// dependency.
			row = append(row, "")
		}

		if len(row) != 4 {
			return nil, fmt.Errorf(
				"Unexpected raw dependency format: %s", line)
		}

		// If the path ends in an import version, strip it since we have
		// an exact version available in Version.
		if loc := importVersionRe.FindStringIndex(row[1]); loc != nil {
			row[1] = row[1][:loc[0]]
		}

		next := Module{
			Path:    row[1],
			Version: row[2],
			Hash:    row[3],
		}

		// If this is a replacement, then replace the last result
		if row[0] == "=>" {
			result[len(result)-1] = next
			continue
		}

		result = append(result, next)
	}

	return result, nil
}

// importVersionRe is a regular expression that matches the trailing
// import version specifiers like `/v12` on an import that is Go modules
// compatible.
var importVersionRe = regexp.MustCompile(`/v\d+$`)
