package main

import (
	"io"

	"github.com/mitchellh/golicense/license"
	"github.com/mitchellh/golicense/module"
)

// TermOutput is an Output implementation that outputs to the terminal.
type TermOutput struct {
	// Out is the stdout to write to. If this is a TTY, TermOutput will
	// automatically use a "live" updating output mode for status updates.
	// This can be disabled by setting Plain to true below.
	Out io.Writer

	modules map[string]struct{}
}

// Start implements Output
func (o *TermOutput) Start(m *module.Module) {}

// Update implements Output
func (o *TermOutput) Update(m *module.Module, t license.StatusType, msg string) {}

// Finish implements Output
func (o *TermOutput) Finish(m *module.Module, l *license.License, err error) {}

// Close implements Output
func (o *TermOutput) Close() error { return nil }
