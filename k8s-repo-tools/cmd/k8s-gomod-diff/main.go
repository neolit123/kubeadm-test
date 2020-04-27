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
	fmt.Fprintln(out, "k8s-gomod-diff is a tool for comparing gomod files "+
		"and optionally printing results in GitHub issues")
	fmt.Fprintln(out, "\nusage:")
	fmt.Fprintf(out, "  k8s-gomod-diff -dest=some-url-or-file -source=-dest=some-url-or-file -token=<token> <options>\n\n")
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
		pkg.FlagSource,
		pkg.FlagDest,
		pkg.FlagToken,
		pkg.FlagDryRun,
		pkg.FlagTargetIssue,
		pkg.FlagTimeout,
	}
	fd := pkg.GetDefaultFlagDescriptions()
	fd[pkg.FlagDest] = "Destination gomod file or URL"
	fd[pkg.FlagSource] = "Source gomod file or URL"
	pkg.SetupFlags(&d, flag.CommandLine, flagList, fd)
	flag.Parse()

	// Validate the user parameters.
	if err := validateData(&d); err != nil {
		pkg.PrintErrorAndExit(err)
	}

	// Create an HTTP client and process the data.
	pkg.NewClient(&d, nil)
	out, err := process(&d)
	if err != nil {
		pkg.Errorf(err.Error())
	}

	pkg.Logf("done!")

	if len(d.TargetIssue) > 0 {
		// if d.DryRun {
		// TODO
		// }
	} else {
		formatOutput(os.Stdout, out, d.Source, d.Dest)
	}
}
