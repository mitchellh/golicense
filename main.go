package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/fatih/color"
	"github.com/google/go-github/v18/github"
	"github.com/rsc/goversion/version"
	"golang.org/x/oauth2"

	"github.com/mitchellh/golicense/license"
	githubFinder "github.com/mitchellh/golicense/license/github"
	"github.com/mitchellh/golicense/module"
)

const (
	EnvGitHubToken = "GITHUB_TOKEN"
)

func main() {
	os.Exit(realMain())
}

func realMain() int {
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, color.RedString(
			"❗️ Path to file to analyze expected.\n\n"))
		flag.Usage()
		return 1
	} else if len(args) > 1 {
		fmt.Fprintf(os.Stderr, color.RedString(
			"❗️ Exactly one argument is allowed at a time.\n\n"))
		flag.Usage()
		return 1
	}

	vsn, err := version.ReadExe(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, color.RedString(fmt.Sprintf(
			"❗️ Error reading %q: %s\n", args[0], err)))
		return 1
	}

	if vsn.ModuleInfo == "" {
		fmt.Fprintf(os.Stderr, color.YellowString(fmt.Sprintf(
			"⚠️  %q ⚠️\n\n"+
				"This executable was compiled without using Go modules or has \n"+
				"zero dependencies. golicense considers this an error (exit code 1).\n", args[0])))
		return 1
	}

	mods, err := module.ParseExeData(vsn.ModuleInfo)
	if err != nil {
		fmt.Fprintf(os.Stderr, color.RedString(fmt.Sprintf(
			"❗️ Error parsing dependencies: %s", err)))
		return 1
	}

	ctx := context.Background()

	var githubClient *http.Client
	if v := os.Getenv(EnvGitHubToken); v != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: v})
		githubClient = oauth2.NewClient(ctx, ts)
	}

	fs := []license.Finder{
		&githubFinder.RepoAPI{
			Client: github.NewClient(githubClient),
		},
	}

	o := &TermOutput{Out: os.Stdout, Modules: mods}

	// Kick off all the license lookups.
	var wg sync.WaitGroup
	sem := NewSemaphore(5)
	for _, m := range mods {
		wg.Add(1)
		go func(m module.Module) {
			defer wg.Done()

			// Acquire a semaphore so that we can limit concurrency
			sem.Acquire()
			defer sem.Release()

			// Build the context
			ctx := license.StatusWithContext(ctx, StatusListener(o, &m))

			// Lookup
			o.Start(&m)
			lic, err := license.Find(ctx, m, fs)
			o.Finish(&m, lic, err)
		}(m)
	}

	// Wait for all lookups to complete
	wg.Wait()

	// Close the output
	if err := o.Close(); err != nil {
		fmt.Fprintf(os.Stderr, color.RedString(fmt.Sprintf(
			"❗️ Error: %s\n", err)))
		return 1
	}

	return 0
}
