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
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/version"
	"k8s.io/kubeadm/k8s-repo-tools/pkg"
)

func printUsage() {
	out := os.Stderr
	fmt.Fprintln(out, "k8s-latest-version is a tool for obtaining the latest SemVer "+
		"from a list of tags separated by \\n and passed via stdin")
	fmt.Fprintln(out, "\nusage:")
	fmt.Fprintf(out, "  git tag | k8s-latest-version -branch=release-1.17 -branch-prefix=release-\n\n")
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
		pkg.FlagBranch,
		pkg.FlagPrefixBranch,
	}
	pkg.SetupFlags(&d, flag.CommandLine, flagList, nil)
	flag.Parse()

	latestTag, err := process(os.Stdin, &d)
	if err != nil {
		pkg.PrintErrorAndExit(err)
	}

	// Print the latest tag to stdout
	pkg.Warningf("found latest tag %q", latestTag)
	fmt.Println(latestTag)
}

func process(input io.Reader, d *pkg.Data) (string, error) {
	var branchV *version.Version
	var err error

	// If the branch is defined extract a Version out of it
	if len(d.Branch) != 0 {
		if !strings.Contains(d.Branch, d.PrefixBranch) {
			return "", errors.Errorf("branch %q does not contain the branch prefix %q", d.Branch, d.PrefixBranch)
		}

		ver := strings.Trim(d.Branch, d.PrefixBranch)
		if strings.Count(ver, ".") < 2 {
			ver = ver + ".0"
		}
		branchV, err = version.ParseSemantic(ver)
		if err != nil {
			return "", errors.Wrap(err, "could not extract a SemVer from the given branch")
		}
	}

	// Read the list of tags from stdin
	var lines []string
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return "", errors.Wrap(err, "error scanning the given input")
	}
	pkg.Warningf("using the following input: %v", lines)

	// Get the latest SemVer tag
	return getLatestTag(lines, branchV)
}

func getLatestTag(lines []string, branchV *version.Version) (string, error) {
	var result string
	minV := version.MustParseSemantic("v0.0.0")

	for _, line := range lines {
		v, err := pkg.TagToVersion(line)
		if err != nil {
			pkg.Warningf(err.Error())
			continue
		}

		// If a branch is requested skip all other versioned tags
		if branchV != nil {
			if v.Major() != branchV.Major() || v.Minor() != branchV.Minor() {
				continue
			}
		}

		if minV.LessThan(v) {
			minV = v
			result = line
		}
	}

	if len(result) == 0 {
		if branchV != nil {
			return "", errors.Errorf("could not find any SemVer tag that matches branch version %d.%d",
				branchV.Major(), branchV.Minor())
		}
		return "", errors.Errorf("could not find the latest tag from the given input")
	}
	return result, nil
}
