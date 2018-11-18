package github

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/google/go-github/v18/github"
	"github.com/mitchellh/golicense/license"
	"github.com/mitchellh/golicense/module"
)

// RepoAPI implements license.Finder and looks up the license of a module
// using the GitHub Repository License API[1].
//
// This API will return the detected license based on the current source code.
// Therefore it is theoretically possible for a dependency to have a different
// license based on the exact match of the SHA (the project changed licenses).
// In practice, this is quite rare so it is up to the caller to determine if
// this is an acceptable risk or not.
//
// [1]: https://developer.github.com/v3/licenses/#get-the-contents-of-a-repositorys-license
type RepoAPI struct {
	Client *github.Client
}

// License implements license.Finder
func (f *RepoAPI) License(ctx context.Context, m module.Module) (*license.License, error) {
	matches := githubRe.FindStringSubmatch(m.Path)
	if matches == nil {
		return nil, nil
	}

FETCH_RETRY:
	license.UpdateStatus(ctx, license.StatusNormal, "querying license")
	rl, _, err := f.Client.Repositories.License(ctx, matches[1], matches[2])
	if rateErr, ok := err.(*github.RateLimitError); ok {
		dur := time.Until(rateErr.Rate.Reset.Time)
		timer := time.NewTimer(dur)
		defer timer.Stop()
		license.UpdateStatus(ctx, license.StatusWarning, fmt.Sprintf(
			"rate limited by GitHub, waiting %s", dur))

		select {
		case <-ctx.Done():
			// Context cancelled or ended so return early
			return nil, ctx.Err()

		case <-timer.C:
			// Rate limit should be up, retry
			goto FETCH_RETRY
		}
	}
	if err != nil {
		return nil, err
	}

	// If the license type is "other" then we try to use go-license-detector
	// to determine the license, which seems to be accurate in these cases.
	if rl.GetLicense().GetKey() == "other" {
		return detect(rl)
	}

	return &license.License{
		Name: rl.GetLicense().GetName(),
		SPDX: rl.GetLicense().GetSPDXID(),
	}, nil
}

// githubRe is the regexp matching the package for a GitHub import.
var githubRe = regexp.MustCompile(`^github\.com/([^/]+)/([^/]+)$`)
