package main

import (
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/mitchellh/golicense/config"
	"github.com/mitchellh/golicense/license"
	"github.com/mitchellh/golicense/module"
)

// MDOutput writes the results of license lookups to an MD file.
type MDOutput struct {
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
func (o *MDOutput) Start(m *module.Module) {}

// Update implements Output
func (o *MDOutput) Update(m *module.Module, t license.StatusType, msg string) {}

// Finish implements Output
func (o *MDOutput) Finish(m *module.Module, l *license.License, err error) {
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
func (o *MDOutput) Close() error {
	o.lock.Lock()
	defer o.lock.Unlock()

	f, err := os.Create(o.Path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Headers
	_, err = f.WriteString("# Licenses\n\n|Dependency|Version|SPDX ID|License|Allowed|\n|----------|-------|-------|-------|-------|\n")
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

		SPDX := ""
		License := "no"
		Allowed := "unknown"

		raw := o.modules[m]
		if raw == nil {
		} else if err, ok := raw.(error); ok {
			// If the value is an error, then note the error
			License = strings.ReplaceAll(err.Error(), "\n", "<br/>")
			Allowed = "no"
		} else if lic, ok := raw.(*license.License); ok {
			// If the value is a license, then mark the license
			if lic != nil {
				SPDX = lic.SPDX
			}
			License = strings.ReplaceAll(strings.ReplaceAll(lic.String(), ">", "*"), "<", "*")
			if o.Config != nil {
				switch o.Config.Allowed(lic) {
				case config.StateAllowed:
					Allowed = "yes"
				case config.StateDenied:
					Allowed = "no"
				}
			}
		}
		_, err = f.WriteString("|[" + m.Path + "](https://" + m.Path + ")|" + m.Version + "|" + SPDX + "|" + License + "|" + Allowed + "|\n")
		if err != nil {
			return err
		}
	}

	return nil
}
