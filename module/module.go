package module

import (
	"fmt"
	"strings"
)

// Module represents a single Go module.
//
// Depending on the source that this is parsed from, fields may be empty.
// All helper functions on Module work with zero values. See their associated
// documentation for more information on exact behavior.
type Module struct {
	Path    string // Import path, such as "github.com/mitchellh/golicense"
	Version string // Version like "v1.2.3"
	Hash    string // Hash such as "h1:abcd1234"
}

// ParseExeData parses the raw dependency information from a compiled Go
// binary's readonly data section. Any unexpected values will return errors.
func ParseExeData(raw string) ([]Module, error) {
	var result []Module
	for _, line := range strings.Split(strings.TrimSpace(raw), "\n") {
		row := strings.Split(line, "\t")

		// Ignore non-dependency information, such as path/mod.
		if row[0] != "dep" {
			continue
		}

		if len(row) != 4 {
			return nil, fmt.Errorf(
				"Unexpected raw dependency format: %s", line)
		}

		result = append(result, Module{
			Path:    row[1],
			Version: row[2],
			Hash:    row[3],
		})
	}

	return result, nil
}
