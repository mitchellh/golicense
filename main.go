package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/mitchellh/golicense/module"
	"github.com/rsc/goversion/version"
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

	for _, m := range mods {
		println(fmt.Sprintf("%s\t%s", m.Path, m.Version))
	}

	return 0
}

func printError() {
}
