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
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/v29/github"
	"k8s.io/apimachinery/pkg/util/version"
)

// GitHubGetRefs obtains a list of References from a GitHub repository.
func GitHubGetRefs(d *Data, repo string, refs string) ([]*github.Reference, error) {
	Logf("getting %q from repository %q", refs, repo)
	ownerRepo := strings.Split(repo, "/")

	ctx, cancel := d.CreateContext()
	defer cancel()
	r, resp, err := d.client.Git.GetRefs(ctx, ownerRepo[0], ownerRepo[1], refs)
	// handle not found by returning an empty list
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		return []*github.Reference{}, nil
	}
	if err != nil {
		return nil, err
	}
	return r, nil
}

// GitHubGetTags obtains tags from a GitHub repository.
func GitHubGetTags(d *Data, repo string) ([]*github.Reference, error) {
	return GitHubGetRefs(d, repo, "refs/tags")
}

// GitHubGetBranches obtains branches from a GitHub repository.
func GitHubGetBranches(d *Data, repo string) ([]*github.Reference, error) {
	return GitHubGetRefs(d, repo, "refs/heads")
}

// GitHubCreateRef creates a general Reference in a GitHub repository.
func GitHubCreateRef(d *Data, repo, ref, sha string, dryRun bool) (*github.Reference, error) {
	newRef := github.Reference{
		Ref: github.String(ref),
		Object: &github.GitObject{
			SHA: github.String(sha),
		},
	}
	if dryRun {
		Logf("%s: would create ref %q from commit %q in repository %q", PrefixDryRun, ref, sha, repo)
		return &newRef, nil
	}
	ownerRepo := strings.Split(repo, "/")
	ctx, cancel := d.CreateContext()
	defer cancel()
	Logf("creating ref %q from commit %q in repository %q", ref, sha, repo)
	_, _, err := d.client.Git.CreateRef(ctx, ownerRepo[0], ownerRepo[1], &newRef)
	return &newRef, err
}

// GitHubGetRef obtains a Reference from a GitHub repository.
func GitHubGetRef(d *Data, repo, ref string) (*github.Reference, error) {
	ownerRepo := strings.Split(repo, "/")
	ctx, cancel := d.CreateContext()
	defer cancel()
	Logf("getting ref %q from repository %q", ref, repo)
	r, _, err := d.client.Git.GetRef(ctx, ownerRepo[0], ownerRepo[1], ref)
	return r, err
}

// GitHubCreateNewBranches goes trough a list of branches and creates them
// based on the HEAD of the master branch.
func GitHubCreateNewBranches(
	d *Data,
	repo string,
	branchesDest *[]*github.Reference,
	newBranches []*github.Reference,
	masterSHA string) error {

	for _, branch := range newBranches {
		// In dry-run mode just append the new ref to the given list of destination refs.
		if d.DryRun {
			ref, _ := GitHubCreateRef(d, repo, branch.GetRef(), masterSHA, true)
			*branchesDest = append(*branchesDest, ref)
			continue
		}

		// Always create new branches from "master".
		if _, err := GitHubCreateRef(d, repo, branch.GetRef(), masterSHA, false); err != nil {
			return err
		}
	}
	return nil
}

// GitHubCreateNewTags goes trough a list of tags and creates
// them for matching versioned branch from a list of branches.
// If no matching branch is found the SHA of master is used.
func GitHubCreateNewTags(
	d *Data,
	repo string,
	tagsDest *[]*github.Reference,
	branches, newTags []*github.Reference,
	masterSHA string) error {

	for _, tag := range newTags {
		sha := FindBranchHEADForTag(tag, d.PrefixBranch, masterSHA, branches)

		// In dry-run mode just append the new ref to the given list of destination refs.
		if d.DryRun {
			ref, _ := GitHubCreateRef(d, repo, tag.GetRef(), sha, true)
			*tagsDest = append(*tagsDest, ref)
			continue
		}

		if _, err := GitHubCreateRef(d, repo, tag.GetRef(), sha, false); err != nil {
			return err
		}
	}
	return nil
}

// GitHubCompareBranches compares a couple of branches or SHAs of a GitHub repository.
func GitHubCompareBranches(d *Data, repo, base, head string) (*github.CommitsComparison, error) {
	ctx, cancel := d.CreateContext()
	defer cancel()
	ownerRepo := strings.Split(repo, "/")
	cmp, _, err := d.client.Repositories.CompareCommits(ctx, ownerRepo[0], ownerRepo[1], base, head)
	return cmp, err
}

// GitHubMergeBranch merges head into the base branch and creates a merge commit.
// TODO: add fake transport
// https://github.com/google/go-github/blob/60d040d2dafa18fa3e86cbf22fbc3208ef9ef1e0/github/repos_merging.go#L25
func GitHubMergeBranch(d *Data, repo, base, head, commitMessage string) (*github.RepositoryCommit, *github.Response, error) {
	// return fake results on dry-run
	if d.DryRun {
		Logf("%s: would create a merge commit in repository %q", PrefixDryRun, repo)
		commit := &github.RepositoryCommit{
			SHA:    github.String("dry-run-sha"),
			Commit: &github.Commit{Message: github.String(commitMessage)},
		}
		resp := &github.Response{
			Response: &http.Response{
				StatusCode: http.StatusCreated,
			},
		}
		return commit, resp, nil
	}

	ctx, cancel := d.CreateContext()
	defer cancel()
	ownerRepo := strings.Split(repo, "/")
	req := github.RepositoryMergeRequest{
		Base:          github.String(base),
		Head:          github.String(head),
		CommitMessage: github.String(commitMessage),
	}
	Logf("merging %q into %q for repository %q", head, base, repo)
	return d.client.Repositories.Merge(ctx, ownerRepo[0], ownerRepo[1], &req)
}

// GitHubGetCreateRelease first checks if a tag exists and obtains a release from this tag.
// If the tag is missing return an error. If the release is missing create it.
func GitHubGetCreateRelease(d *Data, repo, tag string, body string, dryRun bool) (*github.RepositoryRelease, error) {
	ownerRepo := strings.Split(repo, "/")

	Logf("checking if tag %q exists", tag)
	ctx, cancel := d.CreateContext()
	defer cancel()
	_, _, err := d.client.Git.GetRef(ctx, ownerRepo[0], ownerRepo[1], "refs/tags/"+tag)
	if err != nil {
		return nil, err
	}

	Logf("getting release from tag %q", tag)
	ctx, cancel = d.CreateContext()
	defer cancel()
	release, resp, err := d.client.Repositories.GetReleaseByTag(ctx, ownerRepo[0], ownerRepo[1], tag)
	if resp.StatusCode == http.StatusOK {
		return release, nil
	}
	// Don't treat "not found" as an error
	if resp.StatusCode != http.StatusNotFound && err != nil {
		return nil, err
	}

	// Check if this is a pre-release
	v, err := version.ParseSemantic(tag)
	if err != nil {
		return nil, err
	}
	var isPreRelease bool
	if len(v.PreRelease()) > 0 {
		isPreRelease = true
	}

	release = &github.RepositoryRelease{
		TagName:    github.String(tag),
		Name:       github.String(tag),
		Body:       github.String(body),
		Draft:      github.Bool(false),
		Prerelease: github.Bool(isPreRelease),
	}

	if dryRun {
		Logf("%s: would create a release for tag %q in repository %q", PrefixDryRun, tag, repo)
		return release, nil
	}

	Logf("creating release for tag %q", tag)
	ctx, cancel = d.CreateContext()
	defer cancel()
	release, _, err = d.client.Repositories.CreateRelease(ctx, ownerRepo[0], ownerRepo[1], release)
	if err != nil {
		return nil, err
	}
	return release, err
}

// GitHubUploadReleaseAssets uploads files to a GitHub repository.
func GitHubUploadReleaseAssets(d *Data, repo string, release *github.RepositoryRelease, am assetMap, dryRun bool) ([]*github.ReleaseAsset, error) {
	// Get the existing list of assets, but convert them to a list of pointers
	assets := make([]*github.ReleaseAsset, len(release.Assets))
	for i := range release.Assets {
		assets[i] = &release.Assets[i]
	}
	Logf("found %d existing assets in release", len(assets))

	// Only upload new files.
	assetsNew := map[string]string{}
	for k, v := range am {
		exists := false
		for _, a := range assets {
			if k == a.GetName() {
				exists = true
				break
			}
		}
		if exists {
			Logf("skipping existing asset %q", k)
			continue
		}
		assetsNew[k] = v
	}
	Logf("found %d new assets", len(assetsNew))

	newReleaseAssets := make([]*github.ReleaseAsset, len(assetsNew))
	var i int
	for k, v := range assetsNew {
		// Handle dry run.
		if dryRun {
			Logf("%s: would upload asset %q from path %q", PrefixDryRun, k, v)
			newReleaseAssets[i] = &github.ReleaseAsset{Name: github.String(k)}
			i++
			continue
		}

		// Open the file.
		file, err := os.Open(v)
		if err != nil {
			return nil, err
		}

		opt := github.UploadOptions{
			Name:      k,
			MediaType: "application/octet-stream",
		}

		// Upload the file as asset.
		Logf("uploading asset %q from path %q", k, v)
		ctx, cancel := d.CreateContext()
		ownerRepo := strings.Split(repo, "/")
		releaseAsset, _, err := d.client.Repositories.UploadReleaseAsset(ctx, ownerRepo[0], ownerRepo[1], release.GetID(), &opt, file)
		if err != nil {
			cancel()
			file.Close()
			return nil, err
		}
		cancel()
		file.Close()
		newReleaseAssets[i] = releaseAsset
		i++
	}
	assets = append(assets, newReleaseAssets...)
	return assets, nil
}
