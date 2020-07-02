package golang

import (
	"context"
	"fmt"
	"regexp"

	"github.com/cloudentity/golicense/module"
)

type Translator struct{}

func (t Translator) Translate(ctx context.Context, m module.Module) (module.Module, bool) {
	ms := re.FindStringSubmatch(m.Path)
	if ms == nil {
		return module.Module{}, false
	}

	// Matches, convert to github
	m.Path = fmt.Sprintf("github.com/golang/%s", ms[1])
	return m, true
}

// re is the regexp matching the package for a GoPkg import. This is taken
// almost directly from the GoPkg source code itself so it should match
// perfectly.
var re = regexp.MustCompile(`^go\.googlesource\.com/([^/]+)$`)
