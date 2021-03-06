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
	"fmt"
	"net/http"

	"github.com/google/go-github/v29/github"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/version"
	"k8s.io/kubeadm/k8s-repo-tools/pkg"
)

// process is responsible for all operations that the application performs.
func process(d *pkg.Data) (*github.Reference, *github.RepositoryCommit, error) {

	pkg.Logf("using branch prefix %q", d.PrefixBranch)

	// Obtain destination repository tags and branches.
	tagsDest, err := pkg.GitHubGetTags(d, d.Dest)
	if err != nil {
		return nil, nil, &genericError{error: err}
	}
	branchesDest, err := pkg.GitHubGetBranches(d, d.Dest)
	if err != nil {
		return nil, nil, &genericError{error: err}
	}

	// Trim branches and tags that are not usable.
	pkg.LogRefList("existing tags", d.Dest, tagsDest)
	pkg.LogRefList("existing branches", d.Dest, branchesDest)

	// Find the latest versioned branch.
	latestBranch, err := pkg.FindLatestBranch(branchesDest, d.PrefixBranch)
	if err != nil {
		return nil, nil, &releaseBranchError{error: err}
	}
	pkg.Logf("found %q as the latest versioned branch", latestBranch.GetRef())

	// Find the latest tag for this versioned branch.
	latestBranchVer, _ := pkg.BranchRefToVersion(latestBranch, d.PrefixBranch)
	latestTag, err := pkg.FindLatestTag(tagsDest, latestBranchVer)
	if err != nil {
		return nil, nil, &genericError{error: err}
	}
	pkg.Logf("found %q as the latest versioned tag for branch %q", latestTag.GetRef(), latestBranch.GetRef())

	// Prepare the fast-forward window.
	// Given the latest tag x for the latest branch y:
	// - x must be >= (y.MAJOR).(y.MINOR).0-beta.0
	// - x must be <  (y.MAJOR).(y.MINOR).0-rc.1
	// https://github.com/kubernetes/sig-release/blob/d6a4a0c/release-engineering/role-handbooks/branch-manager.md#branch-fast-forward
	minVersion := version.MustParseSemantic(
		fmt.Sprintf("%d.%d.0-beta.0", latestBranchVer.Major(), latestBranchVer.Minor()),
	)
	maxVersion := version.MustParseSemantic(
		fmt.Sprintf("%d.%d.0-rc.1", latestBranchVer.Major(), latestBranchVer.Minor()),
	)
	latestTagVer, _ := pkg.TagRefToVersion(latestTag)

	if !(latestTagVer.AtLeast(minVersion) && latestTagVer.LessThan(maxVersion)) {
		return nil, nil, &fastForwardWindowError{
			error: errors.Errorf("the latest versioned tag %q for branch %q does not fall within the fast-forward window: %s <= VER < %s",
				latestTag.GetRef(), latestBranch, minVersion.String(), maxVersion.String()),
		}
	}

	// Compare the latest and the master branches.
	cmp, err := pkg.GitHubCompareBranches(d, d.Dest, latestBranch.GetRef(), pkg.BranchMaster)
	if err != nil {
		return nil, nil, &genericError{error: err}
	}
	switch cmp.GetStatus() {
	case "identical":
		return nil, nil, &identicalBranchesError{
			error: errors.Errorf("the branches %q and %q are identical",
				pkg.BranchMaster, latestBranch.GetRef()),
		}
	default:
		break
	}

	pkg.Logf("branch comparison status between %q and %q is reported as %q and there are %d different commit(s)",
		pkg.BranchMaster, latestBranch.GetRef(), cmp.GetStatus(), len(cmp.Commits))
	if len(cmp.Commits) > 0 {
		commitURLs := "list of commits from the comparison:"
		for _, c := range cmp.Commits {
			commitURLs += "\n" + c.GetHTMLURL()
		}
		pkg.Logf(commitURLs)
	}
	pkg.Logf("comparison URL:\n%s", cmp.GetHTMLURL())

	var promptMessage string
	var yes bool

	// Skip prompt.
	if d.Force {
		goto write
	}

	// Prompt the user.
	promptMessage = fmt.Sprintf("Do you want to fast-forward branch %q of repository %q?",
		latestBranch.GetRef(), d.Dest)
	if yes, err = pkg.ShowPrompt(promptMessage); err != nil {
		return nil, nil, &genericError{error: err}
	} else if yes {
		goto write
	}
	return nil, nil, nil

write:
	// Merge the branches.
	commitMessage := pkg.FormatMergeCommitMessage(latestBranch.GetRef(), pkg.BranchMaster)
	commit, resp, err := pkg.GitHubMergeBranch(d, d.Dest, latestBranch.GetRef(), pkg.BranchMaster, commitMessage)
	if err != nil {
		return nil, nil, &genericError{error: err}
	}
	mergeStatus := resp.StatusCode
	switch mergeStatus {
	case http.StatusCreated:
		break
	case http.StatusNoContent:
		return nil, nil, &noContentError{error: errors.Errorf("got status %d when merging branch %q into %q.",
			mergeStatus, pkg.BranchMaster, latestBranch.GetRef())}
	default: // Should not happen?
		return nil, nil, &genericError{error: errors.Errorf("unexpected status %d when merging branch %q into %q. "+
			"Please verify if the branch is mergeable!",
			mergeStatus, pkg.BranchMaster, latestBranch.GetRef()),
		}
	}
	pkg.Logf("created commit with SHA %q in repository %q", commit.GetSHA(), d.Dest)
	return latestBranch, commit, nil
}
