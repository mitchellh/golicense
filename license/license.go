package license

//go:generate mockery -all -inpkg

// License represents a software license.
type License struct {
	Name string // Name is a human-friendly name like "MIT License"
	SPDX string // SPDX ID of the license, blank if unknown or unavailable
	Text string // License text
}

func (l *License) String() string {
	if l == nil {
		return "<license not found or detected>"
	}

	return l.Name
}

func (l *License) TextString() string {
	if l == nil {
		return "<license not found or detected>"
	}

	return l.Text
}

func (l *License) SPDXString() string {
	if l == nil {
		return "<license not found or detected>"
	}

	return l.SPDX
}
