package mapper

import (
	"context"

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

	return &license.License{Name: v, SPDX: v}, nil
}
