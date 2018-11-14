package main

import (
	"github.com/mitchellh/golicense/license"
	"github.com/mitchellh/golicense/module"
)

// Output represents the output format for the progress and completion
// of license lookups. This can be implemented to introduce new UI styles
// or output formats (like JSON, etc.).
type Output interface {
	// Start is called when the license lookup for a module is started.
	Start(*module.Module)

	// Update is called for each status update during the license lookup.
	Update(*module.Module, license.StatusType, string)

	// Finish is called when a module license lookup is complete with
	// the results of the lookup.
	Finish(*module.Module, *license.License, error)

	// Close is called when all modules lookups are completed. This can be
	// used to output a summary report, if any.
	Close() error
}

// StatusListener returns a license.StatusListener implementation for
// a single module to route to an Output implementation.
//
// The caller must still call Start and Finish appropriately on the
// Output while using this StatusListener.
func StatusListener(o Output, m *module.Module) license.StatusListener {
	return &outputStatusListener{m: m, o: o}
}

// outputStatusListener is a license.StatusListener implementation that
// updates the Output for a single module by calling Update.
type outputStatusListener struct {
	m *module.Module
	o Output
}

// UpdateStatus implements license.StatusListener
func (sl *outputStatusListener) UpdateStatus(t license.StatusType, msg string) {
	sl.o.Update(sl.m, t, msg)
}
