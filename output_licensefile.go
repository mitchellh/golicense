package main

import (
	"fmt"
	"os"
	"sort"
	"sync"

	"github.com/omnisci/golicense/config"
	"github.com/omnisci/golicense/license"
	"github.com/omnisci/golicense/module"
)

// LicenseFileOutput writes the results of license lookups to a text file.
// The output file will contain the import path, version, and full license text for each dependency.
type LicenseFileOutput struct {
	// Path - the path to the file to write. This will be overwritten if it exists.
	Path string

	// Config - the configuration (if any). This will be used to check if a license is allowed or not.
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

	// Go through each module and write the data to the licensefile
	for _, k := range keys {
		m := index[k]
		raw := o.modules[m]

		fmt.Fprintln(f, fmt.Sprintf(
			"%s - %s\n", m.Path, m.Version))

		// Extract the license data and write to the licensefile
		if lic, ok := raw.(*license.License); ok {
			if lic != nil {
				fmt.Fprintln(f, lic.SPDX)
				fmt.Fprintln(f, lic.TextString())
			}
		} else {
			fmt.Fprintln(f, "**LICENSE NOT FOUND!**")
		}
		fmt.Fprintln(f, "##########")
	}
	// Save
	if err := f.Close(); err != nil {
		return err
	}

	return nil
}
