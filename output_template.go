package main

import (
	"fmt"
	"os"
	"sort"
	"sync"
	"text/template"

	"github.com/mitchellh/golicense/config"
	"github.com/mitchellh/golicense/license"
	"github.com/mitchellh/golicense/module"
)

// TemplateOutput writes the results of license lookups to a tempalte file.
type TemplateOutput struct {
	// Path is the path to the file to write. This will be overwritten if
	// it exists.
	Path string

	// The template file to use for rendering.
	Template string

	// Config is the configuration (if any). This will be used to check
	// if a license is allowed or not.
	Config *config.Config

	modules map[*module.Module]interface{}
	lock    sync.Mutex
}

// Start implements Output
func (o *TemplateOutput) Start(m *module.Module) {}

// Update implements Output
func (o *TemplateOutput) Update(m *module.Module, t license.StatusType, msg string) {}

// Finish implements Output
func (o *TemplateOutput) Finish(m *module.Module, l *license.License, err error) {
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

type libraryEntry struct {
	Dependency string
	Version    string
	Spdx       string
	License    string
	Allowed    string
}

// Close implements Output
func (o *TemplateOutput) Close() error {
	o.lock.Lock()
	defer o.lock.Unlock()

	keys := make([]string, 0, len(o.modules))
	index := map[string]*module.Module{}
	for m := range o.modules {
		keys = append(keys, m.Path)
		index[m.Path] = m
	}
	sort.Strings(keys)

	entries := []*libraryEntry{}

	// Go through each module and output it into the spreadsheet
	for _, k := range keys {
		m := index[k]
		entry := libraryEntry{
			Dependency: m.Path,
			Version:    m.Version,
			Spdx:       "",
			License:    "",
			Allowed:    "unknown",
		}

		raw := o.modules[m]
		if raw == nil {
			entry.License = "no"
			continue
		}

		if err, ok := raw.(error); ok {
			entry.License = fmt.Sprintf("ERROR: %s", err)
			continue
		}

		if lic, ok := raw.(*license.License); ok {
			if lic != nil {
				entry.Spdx = lic.SPDX
			}
			entry.License = fmt.Sprintf(lic.String())

			if o.Config != nil {
				switch o.Config.Allowed(lic) {
				case config.StateAllowed:
					entry.Allowed = "yes"
				case config.StateDenied:
					entry.Allowed = "no"
				}
			}
		}
		entries = append(entries, &entry)
	}

	t := template.Must(template.ParseFiles(o.Template))
	writer, _ := os.Create(o.Path)
	t.Execute(writer, entries)

	return nil
}
