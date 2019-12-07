package config

import (
	"strings"

	"github.com/mitchellh/golicense/license"
)

// Config is the configuration structure for the license checker.
type Config struct {
	// Allow and Deny are the list of licenses that are allowed or disallowed,
	// respectively. The string value here can be either the license name
	// (case insensitive) or the SPDX ID (case insensitive).
	//
	// If a license is found that isn't in either list, then a warning is
	// emitted. If a license is in both deny and allow, then deny takes
	// priority.
	Allow []string `hcl:"allow,optional"`
	Deny  []string `hcl:"deny,optional"`

	// Preapproved is a list of import paths that have been "preapproved",
	// either because the license is non-standard or approved as a one-off.
	Preapproved []string `hcl:"preapproved,optional"`

	// Override is a map that explicitly sets the license for the given
	// import path. The key is an import path (exact) and the value is
	// the name or SPDX ID of the license. Regardless, the value will
	// be set as both the name and SPDX ID, so SPDX IDs are recommended.
	Override map[string]string `hcl:"override,optional"`

	// Translate is a map that translates one import source into another.
	// For example, "gopkg.in/(.*)" => "github.com/\1" would translate
	// gopkg into github (incorrectly, but the example would work).
	Translate map[string]string `hcl:"translate,optional"`
}

// Allowed returns the allowed state of a license given the configuration.
func (c *Config) Allowed(importPath string, l *license.License) AllowState {
	if l == nil {
		return StateDenied // no license is never allowed
	}

	path := strings.ToLower(importPath)
	name := strings.ToLower(l.Name)
	spdx := strings.ToLower(l.SPDX)

	// Deny takes priority
	for _, v := range c.Deny {
		v = strings.ToLower(v)
		if name == v || spdx == v {
			return StateDenied
		}
	}

	for _, v := range c.Allow {
		v = strings.ToLower(v)
		if name == v || spdx == v {
			return StateAllowed
		}
	}

	// Allow repos that are pre-approved
	for _, v := range c.Preapproved {
		v = strings.ToLower(v)
		if path == v {
			return StateAllowed
		}
	}

	return StateUnknown
}

// AllowState - Is the repo/license allowed?
type AllowState int

// Possible AllowStates
const (
	StateUnknown AllowState = iota
	StateAllowed
	StateDenied
)
