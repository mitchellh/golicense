package license

//go:generate mockery -all -inpkg

// License represents a software license.
type License struct {
	Name string // Name is a human-friendly name like "MIT License"
	SPDX string // SPDX ID of the license, blank if unknown or unavailable
	Text string // License text is the full content of the license file
}

// NameString - get the stringified license name
func (l *License) NameString() string {
	if l == nil {
		return "<license not found or detected>"
	}

	return l.Name
}

// TextString - get the stringified license text
func (l *License) TextString() string {
	if l == nil {
		return "<license not found or detected>"
	}

	return l.Text
}

// SPDXString - get the stringified SPDX type
func (l *License) SPDXString() string {
	if l == nil {
		return "<license not found or detected>"
	}

	return l.SPDX
}
