package license

//go:generate mockery -all -inpkg

// License represents a software license.
type License struct {
	Name string // Name is a human-friendly name like "MIT License"
	SPDX string // SPDX ID of the license, blank if unknown or unavailable
	URL  string // URL of license
}

func (l *License) String() string {
	if l == nil {
		return "<license not found or detected>"
	}

	return l.Name
}
