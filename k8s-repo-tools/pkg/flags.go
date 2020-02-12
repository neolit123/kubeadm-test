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
	"encoding/hex"
	"flag"
	"fmt"
	"regexp"
	"time"

	"github.com/pkg/errors"
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
	if len(token) != 40 {
		return errors.Errorf("the GitHub token (--%s) must be 40 characters long, got %d", option, len(token))
	}
	if _, err := hex.DecodeString(token); err != nil {
		return errors.Wrapf(err, "cannot parse the GitHub token (--%s)", option)
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
