// Package mapper contains a translator using a raw map[string]string
package mapper

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/mitchellh/golicense/module"
)

type Translator struct {
	// Map is the mapping of package names to translate. If the name is
	// exact then it will map exactly to the destination. If the name begins
	// and ends with `/` (forward slash) then it will be treated like a regular
	// expression. The destination can use \1, \2, ... to reference capture
	// groups.
	//
	// The translation will run until in a loop until no translation occurs
	// anymore or len(Map) translations occur, in which case it is an error.
	Map map[string]string
}

func (t Translator) Translate(ctx context.Context, m module.Module) (module.Module, bool) {
	count := 0

RESTART:
	if count > len(t.Map) {
		// No way to error currently...
		return module.Module{}, false
	}

	for k, v := range t.Map {
		if k == m.Path {
			m.Path = v
			count++
			goto RESTART
		}

		if k[0] == '/' && k[len(k)-1] == '/' {
			// Note that this isn't super performant since we constantly
			// recompile any translations as we retry, but we don't expect
			// many translations. If this ever becomes a performance issue,
			// we can fix it then.
			re, err := regexp.Compile(k[1 : len(k)-1])
			if err != nil {
				return module.Module{}, false
			}

			ms := re.FindStringSubmatch(m.Path)
			if ms == nil {
				continue
			}

			for i, m := range ms {
				v = strings.Replace(v, fmt.Sprintf("\\%d", i), m, -1)
			}

			m.Path = v
			count++
			goto RESTART
		}
	}

	return m, count > 0
}
