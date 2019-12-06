package main

import (
	"fmt"
	"os"
	"sort"
	"sync"

	"github.com/mitchellh/golicense/config"
	"github.com/mitchellh/golicense/license"
	"github.com/mitchellh/golicense/module"
)

// LicenseFileOutput writes the results of license lookups to an XLSX file.
type LicenseFileOutput struct {
	// Path is the path to the file to write. This will be overwritten if
	// it exists.
	Path string

	// Config is the configuration (if any). This will be used to check
	// if a license is allowed or not.
	Config *config.Config

	modules map[*module.Module]interface{}
	lock    sync.Mutex
}

// Start implements Output
func (o *LicenseFileOutput) Start(m *module.Module) {}

// Update implements Output
func (o *LicenseFileOutput) Update(m *module.Module, t license.StatusType, msg string) {}

// Finish implements Output
func (o *LicenseFileOutput) Finish(m *module.Module, l *license.License, err error) {
	o.lock.Lock()
	defer o.lock.Unlock()

	if o.modules == nil {
		o.modules = make(map[*module.Module]interface{})
	}

	o.modules[m] = l
	if err != nil {
		o.modules[m] = err
	}
}

// Close implements Output
func (o *LicenseFileOutput) Close() error {
	o.lock.Lock()
	defer o.lock.Unlock()

	f, err := os.Create(o.Path)

	if err != nil {
		return err
	}

	// Sort the modules by name
	keys := make([]string, 0, len(o.modules))
	index := map[string]*module.Module{}
	for m := range o.modules {
		keys = append(keys, m.Path)
		index[m.Path] = m
	}
	sort.Strings(keys)

	// Go through each module and output it into the spreadsheet
	for _, k := range keys {
		m := index[k]
		raw := o.modules[m]

		// if raw == nil {
		fmt.Fprintln(f, fmt.Sprintf(
			"%s - %s\n", m.Path, m.Version))

		// If the value is a license, then mark the license
		if lic, ok := raw.(*license.License); ok {
			if lic != nil {
				fmt.Fprintln(f, lic.SPDX)
				fmt.Fprintln(f, lic.TextString())
				fmt.Fprintln(f, "##########")
			}
		}
		// }
	}
	// Save
	if err := f.Close(); err != nil {
		return err
	}

	return nil
}
