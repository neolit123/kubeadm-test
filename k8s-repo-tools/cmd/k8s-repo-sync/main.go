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
	fmt.Fprintln(out, "k8s-repo-sync is a tool for synchronizing tags and branches\n"+
		"between GitHub repositories")
	fmt.Fprintln(out, "\nusage:")
	fmt.Fprintf(out, "  k8s-repo-sync -source=org/repo -dest=org/repo "+
		"-min-version=v1.17.0 -token <token> <options>\n\n")
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
		pkg.FlagSource,
		pkg.FlagMinVersion,
		pkg.FlagToken,
		pkg.FlagPrefixBranch,
		pkg.FlagOutput,
		pkg.FlagTimeout,
		pkg.FlagDryRun,
		pkg.FlagForce,
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

	// Create an HTTP client and process the data.
	pkg.NewClient(&d, nil)
	refs, err := process(&d)
	if err != nil {
		pkg.PrintErrorAndExit(err)
	}

	// Write the output References to disk.
	if len(d.Output) != 0 {
		if err := writeRefOutputToFile(d.Output, refs); err != nil {
			pkg.PrintErrorAndExit(err)
		}
	}
	pkg.Logf("done!")
}
