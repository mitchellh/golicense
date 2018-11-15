package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/google/go-github/v18/github"
	"github.com/rsc/goversion/version"
	"golang.org/x/oauth2"

	"github.com/mitchellh/golicense/config"
	"github.com/mitchellh/golicense/license"
	githubFinder "github.com/mitchellh/golicense/license/github"
	"github.com/mitchellh/golicense/license/golang"
	"github.com/mitchellh/golicense/license/gopkg"
	"github.com/mitchellh/golicense/license/mapper"
	"github.com/mitchellh/golicense/module"
)

const (
	EnvGitHubToken = "GITHUB_TOKEN"
)

func main() {
	os.Exit(realMain())
}

func realMain() int {
	out := &TermOutput{Out: os.Stdout}

	var flagLicense bool
	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flags.BoolVar(&flagLicense, "license", true,
		"look up and verify license. If false, dependencies are\n"+
			"printed without licenses.")
	flags.BoolVar(&out.Plain, "plain", false, "plain simple output, no colors or live updates")
	flags.BoolVar(&out.Verbose, "verbose", false, "additional logging, requires -plain")
	flags.Parse(os.Args[1:])
	args := flags.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, color.RedString(
			"❗️ Path to file to analyze expected.\n\n"))
		printHelp(flags)
		return 1
	} else if len(args) > 2 {
		fmt.Fprintf(os.Stderr, color.RedString(
			"❗️ Exactly one or two arguments is allowed.\n\n"))
		printHelp(flags)
		return 1
	}

	// Determine the exe path and parse the configuration if given.
	var cfg config.Config
	exePath := args[0]
	if len(args) > 1 {
		exePath = args[1]

		c, err := config.ParseFile(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, color.RedString(fmt.Sprintf(
				"❗️ Error parsing configuration:\n\n%s\n", err)))
			return 1
		}

		// Store the config and set it on the output
		cfg = *c
		out.Config = &cfg
	}

	// Read the dependencies from the binary itself
	vsn, err := version.ReadExe(exePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, color.RedString(fmt.Sprintf(
			"❗️ Error reading %q: %s\n", args[0], err)))
		return 1
	}

	if vsn.ModuleInfo == "" {
		// ModuleInfo empty means that the binary didn't use Go modules
		// or it could mean that a binary has no dependencies. Either way
		// we error since we can't be sure.
		fmt.Fprintf(os.Stderr, color.YellowString(fmt.Sprintf(
			"⚠️  %q ⚠️\n\n"+
				"This executable was compiled without using Go modules or has \n"+
				"zero dependencies. golicense considers this an error (exit code 1).\n", args[0])))
		return 1
	}

	// From the raw module string from the binary, we need to parse this
	// into structured data with the module information.
	mods, err := module.ParseExeData(vsn.ModuleInfo)
	if err != nil {
		fmt.Fprintf(os.Stderr, color.RedString(fmt.Sprintf(
			"❗️ Error parsing dependencies: %s", err)))
		return 1
	}

	// Set the modules on our terminal output so that it can look prettier
	out.Modules = mods

	// Setup a context. We don't connect this to an interrupt signal or
	// anything since we just exit immediately on interrupt. No cleanup
	// necessary.
	ctx := context.Background()

	// Auth with GitHub if available
	var githubClient *http.Client
	if v := os.Getenv(EnvGitHubToken); v != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: v})
		githubClient = oauth2.NewClient(ctx, ts)
	}

	// Build our translators and license finders
	ts := []license.Translator{
		&mapper.Translator{Map: cfg.Translate},
		&golang.Translator{},
		&gopkg.Translator{},
	}
	var fs []license.Finder
	if flagLicense {
		fs = []license.Finder{
			&mapper.Finder{Map: cfg.Override},
			&githubFinder.RepoAPI{
				Client: github.NewClient(githubClient),
			},
		}
	}

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
			ctx := license.StatusWithContext(ctx, StatusListener(out, &m))

			// Lookup
			out.Start(&m)

			// We first try the untranslated version. If we can detect
			// a license then take that. Otherwise, we translate.
			lic, err := license.Find(ctx, m, fs)
			if lic == nil || err != nil {
				lic, err = license.Find(ctx, license.Translate(ctx, m, ts), fs)
			}
			out.Finish(&m, lic, err)
		}(m)
	}

	// Wait for all lookups to complete
	wg.Wait()

	// Close the output
	if err := out.Close(); err != nil {
		fmt.Fprintf(os.Stderr, color.RedString(fmt.Sprintf(
			"❗️ Error: %s\n", err)))
		return 1
	}

	return out.ExitCode()
}

func printHelp(fs *flag.FlagSet) {
	fmt.Fprintf(os.Stderr, strings.TrimSpace(help)+"\n\n", os.Args[0])
	fs.PrintDefaults()
}

const help = `
golicense analyzes the dependencies of a binary compiled from Go.

Usage: %[1]s [flags] [BINARY]
Usage: %[1]s [flags] [CONFIG] [BINARY]

One or two arguments can be given: a binary by itself which will output
all the licenses of dependencies, or a configuration file and a binary
which also notes which licenses are allowed among other settings.

For full help text, see the README in the GitHub repository:
http://github.com/mitchellh/golicense

Flags:

`
