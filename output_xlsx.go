package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/mitchellh/golicense/config"
	"github.com/mitchellh/golicense/license"
	"github.com/mitchellh/golicense/module"
)

// XLSXOutput writes the results of license lookups to an XLSX file.
type XLSXOutput struct {
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
func (o *XLSXOutput) Start(m *module.Module) {}

// Update implements Output
func (o *XLSXOutput) Update(m *module.Module, t license.StatusType, msg string) {}

// Finish implements Output
func (o *XLSXOutput) Finish(m *module.Module, l *license.License, err error) {
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
func (o *XLSXOutput) Close() error {
	o.lock.Lock()
	defer o.lock.Unlock()

	const s = "Sheet1"
	f := excelize.NewFile()

	// Create all our styles
	redStyle, _ := f.NewStyle(`{"fill":{"type":"pattern","pattern":1,"color":["#FFCCCC"]}}`)
	yellowStyle, _ := f.NewStyle(`{"fill":{"type":"pattern","pattern":1,"color":["#FFC107"]}}`)
	greenStyle, _ := f.NewStyle(`{"fill":{"type":"pattern","pattern":1,"color":["#9CCC65"]}}`)

	columns := []string{"Dependency", "Version", "SPDX ID", "License", "Allowed"}
	if len(o.Config.OutputColumns) > 0 {
		columns = o.Config.OutputColumns
	}

	// Headers
	for i, k := range columns {
		colID := string(i + 65)
		f.SetCellValue(s, colID+"1", k)
		f.SetColWidth(s, colID, colID, 40)
	}

	keys, index := o.sortModulesByName(o.modules)

	// Go through each module and output it into the spreadsheet
	for rowIdx, rowKey := range keys {
		rowID := strconv.Itoa(rowIdx + 2)

		module := index[rowKey]
		rawModule := o.modules[module]

		for colIdx, colKey := range columns {
			cellID := string(colIdx+65) + rowID

			f.SetCellValue(s, cellID, o.getModuleValue(module, rawModule, colKey))
			f.SetCellStyle(s, cellID, cellID, o.getModuleStyle(module, rawModule, redStyle, yellowStyle, greenStyle))
		}
	}

	// Save
	if err := f.SaveAs(o.Path); err != nil {
		return err
	}

	return nil
}

func (o *XLSXOutput) sortModulesByName(modules map[*module.Module]interface{}) ([]string, map[string]*module.Module) {
	keys := make([]string, 0, len(modules))
	index := map[string]*module.Module{}
	for m := range modules {
		keys = append(keys, m.Path)
		index[m.Path] = m
	}
	sort.Strings(keys)
	return keys, index
}

func (o *XLSXOutput) getModuleValue(module *module.Module, rawModule interface{}, key string) string {
	compareKey := strings.ToLower(key)
	// keys from module
	switch compareKey {
	case "dependency":
		return module.Path
	case "version":
		return module.Version
	}

	// license key
	if compareKey == "license" {
		if rawModule == nil {
			return "no"
		}
		if err, ok := rawModule.(error); ok {
			return fmt.Sprintf("ERROR: %s", err)
		}
		if lic, ok := rawModule.(*license.License); ok {
			return lic.String()
		}
		return ""
	}

	// allowed key
	if compareKey == "allowed" {
		if rawModule == nil {
			return "no"
		}
		if _, ok := rawModule.(error); ok {
			return "no"
		}
		if lic, ok := rawModule.(*license.License); ok && o.Config != nil {
			switch o.Config.Allowed(lic) {
			case config.StateAllowed:
				return "yes"
			case config.StateDenied:
				return "no"
			}
		}
		return "unknown"
	}

	// other keys
	if rawModule == nil {
		return ""
	}
	if lic, ok := rawModule.(*license.License); ok {
		switch compareKey {
		case "spdx id":
			return lic.SPDX
		case "license text":
			return lic.Text
		}
	}
	return ""
}

func (o *XLSXOutput) getModuleStyle(module *module.Module, rawModule interface{}, deny, indecisive, allow int) int {
	if rawModule == nil {
		return deny
	}
	if _, ok := rawModule.(error); ok {
		return deny
	}
	if lic, ok := rawModule.(*license.License); ok && o.Config != nil {
		switch o.Config.Allowed(lic) {
		case config.StateAllowed:
			return allow
		case config.StateDenied:
			return deny
		}
	}
	return indecisive
}
