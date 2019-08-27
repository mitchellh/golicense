package main

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/gosuri/uilive"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/mitchellh/golicense/config"
	"github.com/mitchellh/golicense/license"
	"github.com/mitchellh/golicense/module"
)

// TermOutput is an Output implementation that outputs to the terminal.
type TermOutput struct {
	// Out is the stdout to write to. If this is a TTY, TermOutput will
	// automatically use a "live" updating output mode for status updates.
	// This can be disabled by setting Plain to true below.
	Out io.Writer

	// Config is the configuration (if any). This will be used to check
	// if a license is allowed or not.
	Config *config.Config

	// Modules is the full list of modules that will be checked. This is
	// optional. If this is given in advance, then the output will be cleanly
	// aligned.
	Modules []module.Module

	// Plain, if true, will use the plain output vs the live updating output.
	// TermOutput will always use Plain output if the Out configured above
	// is not a TTY.
	Plain bool

	// Verbose will log all status updates in plain mode. This has no effect
	// in non-plain mode currently.
	Verbose bool

	modules   map[string]string
	moduleMax int
	exitCode  int
	lineMax   int
	live      *uilive.Writer
	once      sync.Once
	lock      sync.Mutex
}

func (o *TermOutput) ExitCode() int {
	return o.exitCode
}

// Start implements Output
func (o *TermOutput) Start(m *module.Module) {
	o.once.Do(o.init)

	if o.Plain {
		return
	}

	o.lock.Lock()
	defer o.lock.Unlock()
	o.modules[m.Path] = fmt.Sprintf("%s %s starting...", iconNormal, o.paddedModule(m))
	o.updateLiveOutput()
}

// Update implements Output
func (o *TermOutput) Update(m *module.Module, t license.StatusType, msg string) {
	o.once.Do(o.init)

	// In plain & verbose mode, we output every status message, but in normal
	// plain mode we ignore all status updates.
	if o.Plain && o.Verbose {
		fmt.Fprintf(o.Out, fmt.Sprintf(
			"%s %s\n", o.paddedModule(m), msg))
	}

	if o.Plain {
		return
	}

	var colorFunc func(string, ...interface{}) string = fmt.Sprintf
	icon := iconNormal
	switch t {
	case license.StatusWarning:
		icon = iconWarning
		colorFunc = color.YellowString

	case license.StatusError:
		icon = iconError
		colorFunc = color.RedString
	}
	if icon != "" {
		icon += " "
	}

	o.lock.Lock()
	defer o.lock.Unlock()
	o.modules[m.Path] = colorFunc("%s%s %s", icon, o.paddedModule(m), msg)
	o.updateLiveOutput()
}

// Finish implements Output
func (o *TermOutput) Finish(m *module.Module, l *license.License, err error) {
	o.once.Do(o.init)

	var colorFunc func(string, ...interface{}) string = fmt.Sprintf
	icon := iconNormal
	if o.Config != nil {
		state := o.Config.Allowed(l)
		switch state {
		case config.StateAllowed:
			colorFunc = color.GreenString
			icon = iconSuccess

		case config.StateDenied:
			colorFunc = color.RedString
			icon = iconError
			o.exitCode = 1

		case config.StateUnknown:
			if len(o.Config.Allow) > 0 || len(o.Config.Deny) > 0 {
				colorFunc = color.YellowString
				icon = iconWarning
				o.exitCode = 1
			}
		}
	}
	if icon != "" {
		icon += " "
	}

	if o.Plain {
		fmt.Fprintf(o.Out, fmt.Sprintf(
			"%s %s\n", o.paddedModule(m), l.String()))
		return
	}

	o.lock.Lock()
	defer o.lock.Unlock()
	delete(o.modules, m.Path)
	o.pauseLive(func() {
		o.live.Write([]byte(colorFunc(
			"%s%s %s\n", icon, o.paddedModule(m), l.String())))
	})
}

// Close implements Output
func (o *TermOutput) Close() error {
	o.lock.Lock()
	defer o.lock.Unlock()

	if o.live != nil {
		o.live.Stop()
	}

	return nil
}

// paddedModule returns the name of the module padded so that they align nicely.
func (o *TermOutput) paddedModule(m *module.Module) string {
	o.once.Do(o.init)

	if o.moduleMax == 0 {
		return m.Path
	}

	// Pad the path so that it is equivalent to the moduleMax length
	return m.Path + strings.Repeat(" ", o.moduleMax-len(m.Path))
}

// pauseLive pauses the live output for the duration of the function.
//
// lock must be held.
func (o *TermOutput) pauseLive(f func()) {
	o.live.Write([]byte(strings.Repeat(" ", o.lineMax) + "\n"))
	o.live.Flush()
	f()
	o.live.Flush()
	o.live.Stop()
	o.newLive()
	o.updateLiveOutput()
}

// updateLiveOutput updates the output buffer for live status.
//
// lock must be held when this is called
func (o *TermOutput) updateLiveOutput() {
	keys := make([]string, 0, len(o.modules))
	for k := range o.modules {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf bytes.Buffer
	for _, k := range keys {
		if v := len(o.modules[k]); v > o.lineMax {
			o.lineMax = v
		}

		buf.WriteString(o.modules[k] + strings.Repeat(" ", o.lineMax-len(o.modules[k])) + "\n")
	}

	o.live.Write(buf.Bytes())
	o.live.Flush()
}

func (o *TermOutput) newLive() {
	o.live = uilive.New()
	o.live.Out = o.Out
	o.live.Start()
}

func (o *TermOutput) init() {
	if o.modules == nil {
		o.modules = make(map[string]string)
	}

	// Calculate the maximum module length
	for _, m := range o.Modules {
		if v := len(m.Path); v > o.moduleMax {
			o.moduleMax = v
		}
	}

	// Check if the output is a TTY
	if !o.Plain {
		o.Plain = true // default to plain mode unless we can verify TTY
		if iofd, ok := o.Out.(ioFd); ok {
			o.Plain = !terminal.IsTerminal(int(iofd.Fd()))
		}

		if !o.Plain {
			o.newLive()
		}
	}
}

// ioFd is an interface that is implemented by things that have a file
// descriptor. We use this to check if the io.Writer is a TTY.
type ioFd interface {
	Fd() uintptr
}

const (
	iconNormal  = ""
	iconWarning = "⚠️ "
	iconError   = "🚫"
	iconSuccess = "✅"
	iconSpace   = "  "
)
