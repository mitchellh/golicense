package main

import (
	"github.com/hashicorp/go-multierror"
	"github.com/cloudentity/golicense/license"
	"github.com/cloudentity/golicense/module"
)

// MultiOutput calls the functions of multiple Output implementations.
type MultiOutput struct {
	Outputs []Output
}

// Start implements Output
func (o *MultiOutput) Start(m *module.Module) {
	for _, out := range o.Outputs {
		out.Start(m)
	}
}

// Update implements Output
func (o *MultiOutput) Update(m *module.Module, t license.StatusType, msg string) {
	for _, out := range o.Outputs {
		out.Update(m, t, msg)
	}
}

// Finish implements Output
func (o *MultiOutput) Finish(m *module.Module, l *license.License, err error) {
	for _, out := range o.Outputs {
		out.Finish(m, l, err)
	}
}

// Close implements Output
func (o *MultiOutput) Close() error {
	var err error
	for _, out := range o.Outputs {
		if e := out.Close(); e != nil {
			err = multierror.Append(err, e)
		}
	}

	return err
}
