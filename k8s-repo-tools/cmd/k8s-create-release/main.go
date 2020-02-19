/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"fmt"
	"os"

	"k8s.io/kubeadm/k8s-repo-tools/pkg"
)

func printUsage() {
	out := os.Stderr
	fmt.Fprintln(out, "k8s-create-release is a tool for creating a GitHub release\n"+
		"from a tag with a changelog and assets")
	fmt.Fprintln(out, "\nusage:")
	fmt.Fprintf(out, "  k8s-create-release -dest=org/repo -token=<token> -release-tag=<tag> <options>\n\n")
	flag.CommandLine.PrintDefaults()
}

func main() {
	// Set the default output writers.
	pkg.SetLogWriters(os.Stdout, os.Stderr)

	// Initialize the main data structure.
	d := pkg.Data{}

	// Manage flags and source.
	flag.Usage = printUsage
	flag.CommandLine.SetOutput(os.Stderr)
	flagList := []string{
		pkg.FlagDest,
		pkg.FlagToken,
		pkg.FlagTimeout,
		pkg.FlagDryRun,
		pkg.FlagForce,
		pkg.FlagBuildCommand,
		pkg.FlagReleaseTag,
		pkg.FlagReleaseNotesToolPath,
		pkg.FlagReleaseAsset,
	}
	pkg.SetupFlags(&d, flag.CommandLine, flagList)
	flag.Parse()

	// Validate the user parameters.
	if err := validateData(&d); err != nil {
		pkg.PrintErrorAndExit(err)
	}

	// Print a warning in dry-run mode.
	if d.DryRun {
		pkg.PrintSeparator()
		pkg.Warningf("running in %s mode. To enable repository writing operations pass --%s=false", pkg.PrefixDryRun, pkg.FlagDryRun)
		pkg.PrintSeparator()
	}

	pkg.NewClient(&d, nil)
	err := process(&d)
	if err != nil {
		pkg.PrintErrorAndExit(err)
	}
	pkg.Logf("done!")
}
