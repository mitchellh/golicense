package license

// License represents a software license.
type License struct {
	Name string // Name is a human-friendly name like "MIT License"
	SPDX string // SPDX ID of the license, blank if unknown or unavailable
}

func (l *License) String() string {
	if l == nil {
		return "<nil license>"
	}

	return l.Name
}
