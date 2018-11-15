package github

import (
	"encoding/base64"
	"fmt"

	"github.com/google/go-github/v18/github"
	"github.com/mitchellh/golicense/license"
	"gopkg.in/src-d/go-license-detector.v2/licensedb"
	"gopkg.in/src-d/go-license-detector.v2/licensedb/filer"
)

// detect uses go-license-detector as a fallback.
func detect(rl *github.RepositoryLicense) (*license.License, error) {
	ms, err := licensedb.Detect(&filerImpl{License: rl})
	if err != nil {
		return nil, err
	}

	// Find the highest matching license
	var highest float32
	current := ""
	for id, v := range ms {
		if v > 0.90 && v > highest {
			highest = v
			current = id
		}
	}

	if current == "" {
		return nil, nil
	}

	return &license.License{
		Name: current,
		SPDX: current,
	}, nil
}

// filerImpl implements filer.Filer to return the license text directly
// from the github.RepositoryLicense structure.
type filerImpl struct {
	License *github.RepositoryLicense
}

func (f *filerImpl) ReadFile(name string) ([]byte, error) {
	if name != "LICENSE" {
		return nil, fmt.Errorf("unknown file: %s", name)
	}

	return base64.StdEncoding.DecodeString(f.License.GetContent())
}

func (f *filerImpl) ReadDir(dir string) ([]filer.File, error) {
	// We only support root
	if dir != "" {
		return nil, nil
	}

	return []filer.File{filer.File{Name: "LICENSE"}}, nil
}

func (f *filerImpl) Close() {}
