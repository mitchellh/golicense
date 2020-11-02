package main

import (
	"os"
	"sync"
	"text/template"

	"github.com/mitchellh/golicense/config"
	"github.com/mitchellh/golicense/license"
	"github.com/mitchellh/golicense/module"
)

// GoTemplateOutput formats the results of license lookups using a Go template.
type GoTemplateOutput struct {
	// Path is the path to the file to write. This will be overwritten if
	// it exists.
	Path string

	// Config is the configuration (if any). This will be used to check
	// if a license is allowed or not.
	Config *config.Config

	scanningResults []scanningResult
	lock            sync.Mutex
}

type scanningResult struct {
	Module  *module.Module
	License *license.License
}

type result struct {
	Entries []scanningResult
}

// Start implements Output
func (o *GoTemplateOutput) Start(m *module.Module) {}

// Update implements Output
func (o *GoTemplateOutput) Update(m *module.Module, t license.StatusType, msg string) {}

// Finish implements Output
func (o *GoTemplateOutput) Finish(m *module.Module, l *license.License, err error) {
	o.lock.Lock()
	defer o.lock.Unlock()

	if o.scanningResults == nil {
		o.scanningResults = make([]scanningResult, 0)
	}
	result := scanningResult{}
	if m != nil {
		result.Module = m
	}
	if l != nil {
		result.License = l
	}
	o.scanningResults = append(o.scanningResults, result)
}

// Close implements Output
func (o *GoTemplateOutput) Close() error {
	o.lock.Lock()
	defer o.lock.Unlock()

	tmpl, err := template.ParseFiles(o.Path)
	if err != nil {
		return err
	}

	r := result{o.scanningResults}
	if err = tmpl.Execute(os.Stdout, r); err != nil {
		return err
	}

	return nil
}
