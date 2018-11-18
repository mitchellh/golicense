package resolver

import (
	"context"
	"fmt"
	"regexp"

	"github.com/mitchellh/golicense/license"
	"github.com/mitchellh/golicense/module"
	"golang.org/x/tools/go/vcs"
)

// Translator resolves import paths to their proper VCS location. For
// example: "rsc.io/pdf" turns into "github.com/rsc/pdf".
type Translator struct{}

func (t Translator) Translate(ctx context.Context, m module.Module) (module.Module, bool) {
	root, err := vcs.RepoRootForImportPath(m.Path, false)
	if err != nil {
		return module.Module{}, false
	}

	path := hostStripRe.ReplaceAllString(root.Repo, "")
	if m.Path == path {
		return module.Module{}, false
	}

	license.UpdateStatus(ctx, license.StatusNormal, fmt.Sprintf(
		"translated %q to %q", m.Path, path))
	m.Path = path
	return m, true
}

// hostStripRe is a simple regexp to strip the schema from a URL.
var hostStripRe = regexp.MustCompile(`^\w+:\/\/`)
