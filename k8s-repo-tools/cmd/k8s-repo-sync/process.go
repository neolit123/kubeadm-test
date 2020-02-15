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
	"sort"
	"strings"

	"github.com/google/go-github/v29/github"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/version"
	"k8s.io/kubeadm/k8s-repo-tools/pkg"
)

// process is responsible for all operations that the application performs.
func process(d *pkg.Data) ([]*github.Reference, error) {

	// The version should be already validated at this point.
	minV := version.MustParseSemantic(d.MinVersion)
	pkg.Logf("using minimum version %q", minV.String())
	pkg.Logf("using branch prefix %q", d.PrefixBranch)

	// Obtain source repository tags and branches.
	tagsSrc, err := pkg.GitHubGetTags(d, d.Source)
	if err != nil {
		return nil, err
	}
	branchesSrc, err := pkg.GitHubGetBranches(d, d.Source)
	if err != nil {
		return nil, err
	}

	// Trim branches and tags that are not usable.
	tagsSrcTrimmed := pkg.TrimTags(tagsSrc, minV)
	branchesSrcTrimmed := pkg.TrimBranches(branchesSrc, minV, d.PrefixBranch)
	pkg.LogRefList("existing tags", d.Source, tagsSrcTrimmed)
	pkg.LogRefList("existing branches", d.Source, branchesSrcTrimmed)

	// Obtain destination repository tags and branches.
	tagsDest, err := pkg.GitHubGetTags(d, d.Dest)
	if err != nil {
		return nil, err
	}
	branchesDest, err := pkg.GitHubGetBranches(d, d.Dest)
	if err != nil {
		return nil, err
	}

	// Trim branches and tags that are not usable.
	tagsDestTrimmed := pkg.TrimTags(tagsDest, minV)
	branchesDestTrimmed := pkg.TrimBranches(branchesDest, minV, d.PrefixBranch)
	pkg.LogRefList("existing tags", d.Dest, tagsDestTrimmed)
	pkg.LogRefList("existing branches", d.Dest, tagsDestTrimmed)

	// Find new tags and branches.
	newTags := pkg.FindNewRefs(tagsSrcTrimmed, tagsDestTrimmed)
	newBranches := pkg.FindNewRefs(branchesSrcTrimmed, branchesDestTrimmed)
	if len(newTags) == 0 && len(newBranches) == 0 {
		pkg.Logf("no new branches and tags for repository %q", d.Dest)
		return newTags, nil
	}

	// Print summary of new refs.
	pkg.PrintSeparator()
	pkg.LogRefList("new tags", d.Dest, newTags)
	pkg.LogRefList("new branches", d.Dest, newBranches)
	pkg.PrintSeparator()

	var promptMessage, masterSHA string
	var yes bool

	// Skip prompt.
	if d.Force {
		goto write
	}

	// Prompt the user.
	promptMessage = fmt.Sprintf("Do you want to write these changes to repository %q?", d.Dest)
	if yes, err = pkg.ShowPrompt(promptMessage); err != nil {
		return nil, err
	} else if yes {
		goto write
	}
	goto exit

write:
	// Find the master SHA and use it for branch creation in the destination repository.
	for _, b := range branchesDest {
		if strings.TrimPrefix(b.GetRef(), "refs/heads/") == pkg.BranchMaster {
			masterSHA = b.GetObject().GetSHA()
			break
		}
	}
	if len(masterSHA) == 0 {
		return nil, errors.Errorf("the repository %q does not have a branch called %q", d.Dest, pkg.BranchMaster)
	}

	// Create branches in the destination repository.
	if err := pkg.GitHubCreateNewBranches(d, d.Dest, &branchesDest, newBranches, masterSHA); err != nil {
		return nil, err
	}

	if !d.DryRun {
		// Fetch the branches again. this is not needed in dry-run mode, because
		// pkg.GitHubCreateNewBranches() above manages that.
		branchesDest, err = pkg.GitHubGetBranches(d, d.Dest)
		if err != nil {
			return nil, err
		}
	}

	// Update the list of new branches for the destination repository.
	// This updates their SHAs, links and other properties.
	for i := range newBranches {
		for _, branch := range branchesDest {
			if newBranches[i].GetRef() == branch.GetRef() {
				*newBranches[i] = *branch
				break
			}
		}
	}

	if err := pkg.GitHubCreateNewTags(d, d.Dest, &tagsDest, branchesDest, newTags, masterSHA); err != nil {
		return nil, err
	}

	if !d.DryRun {
		// Fetch the tags again. this is not needed in dry-run mode, because
		// pkg.GitHubCreateNewTags() above manages that.
		tagsDest, err = pkg.GitHubGetTags(d, d.Dest)
		if err != nil {
			return nil, err
		}
	}

	// Update the list of new tags for the destination repository.
	// this updates their SHAs, links and other properties.
	for i := range newTags {
		for _, tag := range tagsDest {
			if newTags[i].GetRef() == tag.GetRef() {
				*newTags[i] = *tag
				break
			}
		}
	}

exit:
	// Sort and return.
	refs := append(newTags, newBranches...)
	sort.Slice(refs, func(i, j int) bool {
		return refs[i].GetRef() < refs[j].GetRef()
	})
	return refs, nil
}
