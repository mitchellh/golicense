package mapper

import (
	"context"
	"fmt"

	"github.com/mitchellh/go-spdx"
	"github.com/mitchellh/golicense/license"
	"github.com/mitchellh/golicense/module"
)

// Finder implements license.Finder and sets the license type based on the
// given mapping if the path exists in the map.
type Finder struct {
	Map map[string]string
}

// License implements license.Finder
func (f *Finder) License(ctx context.Context, m module.Module) (*license.License, error) {
	v, ok := f.Map[m.Path]
	if !ok {
		return nil, nil
	}

	// Look up the license by SPDX ID
	lic, err := spdx.License(v)
	if err != nil {
		return nil, fmt.Errorf("Override license %q SPDX lookup error: %s", v, err)
	}

	return &license.License{Name: lic.Name, SPDX: lic.ID}, nil
}
