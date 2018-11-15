package license

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/mitchellh/golicense/module"
)

// Finder implementations can find a license for a given module.
type Finder interface {
	// License looks up the license for a given module.
	License(context.Context, module.Module) (*License, error)
}

// Translator implementations can convert one module path to another
// module path that is more suitable for license lookup.
type Translator interface {
	// Translate takes a module and converts it into another module.
	// This is used to, for example, detect gopkg.in URLs as GitHub
	// repositories.
	Translate(context.Context, module.Module) (module.Module, bool)
}

// Translate translates the given module or returns the same module if
// no translation is necessary.
func Translate(ctx context.Context, m module.Module, ts []Translator) module.Module {
	for _, t := range ts {
		n, ok := t.Translate(ctx, m)
		if ok {
			m = n
		}
	}

	return m
}

// Find finds the license for the given module using a set of finders.
//
// The finders are tried in the order given. The first finder to return
// a non-nil License without an error is returned. If a finder returns
// an error, other finders are still attempted. It is possible for a non-nil
// license to be returned WITH a non-nil error meaning a different lookup
// failed.
func Find(ctx context.Context, m module.Module, fs []Finder) (r *License, rerr error) {
	for _, f := range fs {
		lic, err := f.License(ctx, m)
		if err != nil {
			rerr = multierror.Append(rerr, err)
			continue
		}
		if lic != nil {
			r = lic
			break
		}
	}

	return
}
