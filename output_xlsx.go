package main

import (
	"fmt"
	"sort"
	"strconv"
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

	// Headers
	f.SetCellValue(s, "A1", "Dependency")
	f.SetCellValue(s, "B1", "Version")
	f.SetCellValue(s, "C1", "License")
	f.SetCellValue(s, "D1", "Allowed")
	f.SetColWidth(s, "A", "A", 40)
	f.SetColWidth(s, "B", "B", 20)
	f.SetColWidth(s, "C", "C", 40)
	f.SetColWidth(s, "D", "D", 10)

	// Create all our styles
	redStyle, _ := f.NewStyle(`{"fill":{"type":"pattern","pattern":1,"color":["#FFCCCC"]}}`)
	yellowStyle, _ := f.NewStyle(`{"fill":{"type":"pattern","pattern":1,"color":["#FFC107"]}}`)

	// Sort the modules by name
	keys := make([]string, 0, len(o.modules))
	index := map[string]*module.Module{}
	for m := range o.modules {
		keys = append(keys, m.Path)
		index[m.Path] = m
	}
	sort.Strings(keys)

	// Go through each module and output it into the spreadsheet
	for i, k := range keys {
		row := strconv.FormatInt(int64(i+2), 10)

		m := index[k]
		f.SetCellValue(s, "A"+row, m.Path)
		f.SetCellValue(s, "B"+row, m.Version)
		f.SetCellValue(s, "D"+row, "unknown")
		f.SetCellStyle(s, "A"+row, "A"+row, yellowStyle)
		f.SetCellStyle(s, "B"+row, "B"+row, yellowStyle)
		f.SetCellStyle(s, "C"+row, "C"+row, yellowStyle)
		f.SetCellStyle(s, "D"+row, "D"+row, yellowStyle)

		raw := o.modules[m]
		if raw == nil {
			f.SetCellValue(s, "D"+row, "no")
			f.SetCellStyle(s, "A"+row, "A"+row, redStyle)
			f.SetCellStyle(s, "B"+row, "B"+row, redStyle)
			f.SetCellStyle(s, "C"+row, "C"+row, redStyle)
			f.SetCellStyle(s, "D"+row, "D"+row, redStyle)
			continue
		}

		// If the value is an error, then note the error
		if err, ok := raw.(error); ok {
			f.SetCellValue(s, "C"+row, fmt.Sprintf("ERROR: %s", err))
			f.SetCellValue(s, "D"+row, "no")
			f.SetCellStyle(s, "A"+row, "A"+row, redStyle)
			f.SetCellStyle(s, "B"+row, "B"+row, redStyle)
			f.SetCellStyle(s, "C"+row, "C"+row, redStyle)
			f.SetCellStyle(s, "D"+row, "D"+row, redStyle)
			continue
		}

		// If the value is a license, then mark the license
		if lic, ok := raw.(*license.License); ok {
			f.SetCellValue(s, fmt.Sprintf("C%d", i+2), lic.String())
			if o.Config != nil {
				switch o.Config.Allowed(lic) {
				case config.StateAllowed:
					f.SetCellValue(s, fmt.Sprintf("D%d", i+2), "yes")

				case config.StateDenied:
					f.SetCellValue(s, fmt.Sprintf("D%d", i+2), "no")
					f.SetCellStyle(s, "A"+row, "A"+row, redStyle)
					f.SetCellStyle(s, "B"+row, "B"+row, redStyle)
					f.SetCellStyle(s, "C"+row, "C"+row, redStyle)
					f.SetCellStyle(s, "D"+row, "D"+row, redStyle)
				}
			}
		}
	}

	// Save
	if err := f.SaveAs(o.Path); err != nil {
		return err
	}

	return nil
}
