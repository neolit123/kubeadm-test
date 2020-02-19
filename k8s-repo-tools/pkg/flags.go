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

package pkg

import (
	"flag"
	"fmt"
	"regexp"
	"time"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/version"
)

const (
	// FlagDest ...
	FlagDest = "dest"
	// FlagSource ...
	FlagSource = "source"
	// FlagMinVersion ...
	FlagMinVersion = "min-version"
	// FlagToken ...
	FlagToken = "token"
	// FlagPrefixBranch ...
	FlagPrefixBranch = "branch-prefix"
	// FlagOutput ...
	FlagOutput = "output"
	// FlagDryRun ...
	FlagDryRun = "dry-run"
	// FlagForce ...
	FlagForce = "force"
	// FlagTimeout ...
	FlagTimeout = "timeout"
	// FlagReleaseTag ...
	FlagReleaseTag = "release-tag"
	// FlagReleaseNotesToolPath ...
	FlagReleaseNotesToolPath = "release-notes-tool-path"
	// FlagReleaseNotesPath ...
	FlagReleaseNotesPath = "release-notes-path"
	// FlagBuildCommand ...
	FlagBuildCommand = "build-command"
	// FlagReleaseAsset ...
	FlagReleaseAsset = "release-asset"
)

// SetupFlags ...
func SetupFlags(d *Data, fs *flag.FlagSet, flags []string) {
	for _, f := range flags {
		switch f {
		case FlagDest:
			fs.StringVar(&d.Dest, FlagDest, "", "Destination org/repo to write tags and branches to")
		case FlagSource:
			fs.StringVar(&d.Source, FlagSource, "", "Source org/repo from which to take tags and branches")
		case FlagMinVersion:
			fs.StringVar(&d.MinVersion, FlagMinVersion, "", "All versions for tags and branches older than this SemVer will be ignored")
		case FlagToken:
			fs.StringVar(&d.Token, FlagToken, "", "Token to use for authentication with the GitHub API. Write permissions are required for the destination repository")
		case FlagPrefixBranch:
			fs.StringVar(&d.PrefixBranch, FlagPrefixBranch, PrefixBranch, "Branch name prefix. Expected format is \"prefixMAJOR.MINOR\"")
		case FlagOutput:
			fs.StringVar(&d.Output, FlagOutput, "", "Path to a file that will be written with a list of new tags and branches as GitHub API JSON objects")
		case FlagTimeout:
			fs.DurationVar(&d.Timeout, FlagTimeout, time.Second*20, "Timeout for client connections to the GitHub API")
		case FlagDryRun:
			fs.BoolVar(&d.DryRun, FlagDryRun, true, fmt.Sprintf("In %s mode repository writing operations are disabled", PrefixDryRun))
		case FlagForce:
			fs.BoolVar(&d.Force, FlagForce, false, "Skip the confirmation prompt before writing to the destination repository")
		case FlagReleaseTag:
			fs.StringVar(&d.ReleaseTag, FlagReleaseTag, "", "A SemVer tag from which to create a release")
		case FlagReleaseNotesToolPath:
			fs.StringVar(&d.ReleaseNotesToolPath, FlagReleaseNotesToolPath, "", "Path to the release notes tool binary")
		case FlagReleaseNotesPath:
			fs.StringVar(&d.ReleaseNotesPath, FlagReleaseNotesPath, "", fmt.Sprintf("Path to a text file containing release notes. Overrides the usage of %q", FlagReleaseNotesToolPath))
		case FlagBuildCommand:
			fs.StringVar(&d.BuildCommand, FlagBuildCommand, "", "A command to execute for build the release assets")
		case FlagReleaseAsset:
			fs.Var(&d.ReleaseAssets, FlagReleaseAsset, "A release asset to upload to the GitHub release. Multiple instances of the flag are allowed")
		}
	}
}

// ValidateRepo checks if a repository string is of the format 'org/repo'.
func ValidateRepo(option, repo string) error {
	const orgRepo = `[A-Za-z0-9_.-]+\/[A-Za-z0-9_.-]+`
	var regexpOrgRepo = regexp.MustCompile(orgRepo)
	if !regexpOrgRepo.MatchString(repo) {
		return errors.Errorf("the option %q must be of the format 'org/repo' %s", option, orgRepo)
	}
	return nil
}

// ValidateToken checks if a GitHub token is valid.
func ValidateToken(option, token string) error {
	const tokenFormat = `(v[0-9]\.)?[0-9a-f]{40}`
	var regexpTokenFormat = regexp.MustCompile(tokenFormat)
	if !regexpTokenFormat.MatchString(token) {
		return errors.Errorf("the option %q must be a 40 character HEX string "+
			"with an optional version prefix: %s", option, tokenFormat)
	}
	return nil
}

// ValidateEmptyOption checks if a option is empty.
func ValidateEmptyOption(option, value string) error {
	if len(value) == 0 {
		return errors.Errorf("the option %q cannot be empty", option)
	}
	return nil
}

// ValidateReleaseTag validates if a tag is SemVer.
func ValidateReleaseTag(option, value string) error {
	if _, err := version.ParseSemantic(value); err != nil {
		return err
	}
	return nil
}
