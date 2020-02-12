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
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v29/github"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/version"
)

// TrimTags goes trough a list of tags and returns a trimmed list of those that are
// SemVer and are newer or equal than the provided minimum version.
func TrimTags(refs []*github.Reference, minV *version.Version) []*github.Reference {
	result := []*github.Reference{}
	for _, ref := range refs {
		v, err := TagRefToVersion(ref)
		if err != nil {
			Warningf(err.Error())
			continue
		}

		if v.LessThan(minV) {
			Errorf("skipping ref %s; version is older than the minimum version", ref.GetRef())
			continue
		}
		result = append(result, ref)
	}
	return result
}

// TrimBranches goes trough a list of branches and returns a trimmed list of those that contain
// a SemVer newer or equal than the provided minimum version and also contain the user provided
// branch prefix.
func TrimBranches(refs []*github.Reference, minV *version.Version, prefix string) []*github.Reference {
	result := []*github.Reference{}
	for _, ref := range refs {
		v, err := BranchRefToVersion(ref, prefix)
		if err != nil {
			Warningf(err.Error())
			continue
		}

		// Only handle branches whose MAJOR.MINOR are equal or newer than the minimum version.
		if v.Major() < minV.Major() || (minV.Major() == v.Major() && v.Minor() < minV.Minor()) {
			Warningf("the MAJOR.MINOR in ref %q is older than the minimum version; skipping...", ref.GetRef())
			continue
		}
		result = append(result, ref)
	}
	return result
}

// TagRefToVersion converts a tag Reference to a Version.
func TagRefToVersion(ref *github.Reference) (*version.Version, error) {
	refStr := ref.GetRef()
	ver := strings.TrimPrefix(refStr, "refs/tags/")
	if strings.Count(ver, ".") < 2 { // a version without a .PATCH component?
		ver = ver + ".0"
	}

	v, err := version.ParseSemantic(ver)
	if err != nil {
		return nil, errors.Wrapf(err, "skipping ref %s", refStr)
	}
	return v, nil
}

// BranchRefToVersion converts a branch Reference to a Version.
func BranchRefToVersion(ref *github.Reference, prefix string) (*version.Version, error) {
	refStr := ref.GetRef()
	ver := strings.TrimPrefix(refStr, "refs/heads/")
	if !strings.HasPrefix(ver, prefix) {
		return nil, errors.Errorf("skipping non-prefixed ref %q...", refStr)
	}

	ver = strings.TrimPrefix(ver, prefix)
	if strings.Count(ver, ".") < 2 { // a version without a .PATCH component?
		ver = ver + ".0"
	}

	v, err := version.ParseSemantic(ver)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot parse versioned branch from ref %q", refStr)
	}
	return v, nil
}

// FindNewRefs goes trough two lists, src and dest and returns a list
// of elements present in dest but not in src.
func FindNewRefs(src, dest []*github.Reference) []*github.Reference {
	new := []*github.Reference{}
	for _, a := range src {
		found := false
		for _, b := range dest {
			if a.GetRef() == b.GetRef() {
				found = true
				break
			}
		}
		if !found {
			new = append(new, a)
		}
	}
	return new
}

// FindBranchHEADForTag matches a SemVer tag to a versioned branch's MAJOR.MINOR
// and returns the SHA of the match. If no branches are found it returns masterSHA.
func FindBranchHEADForTag(
	tag *github.Reference,
	prefixBranch,
	masterSHA string,
	branches []*github.Reference) string {

	tagStr := tag.GetRef()
	prefix := "refs/tags/"
	tagStr = strings.TrimPrefix(tagStr, prefix)
	tagVer := version.MustParseSemantic(tagStr)
	Logf("finding branch for tag %q", tagStr)

	for _, branch := range branches {
		branchRef := strings.TrimPrefix(branch.GetRef(), "refs/heads/")
		if branchRef == BranchMaster {
			continue
		}
		branchRef = strings.TrimPrefix(branchRef, prefixBranch)
		if strings.Count(branchRef, ".") < 2 {
			branchRef = branchRef + ".0"
		}
		branchVer := version.MustParseSemantic(branchRef)
		if tagVer.Major() == branchVer.Major() && tagVer.Minor() == branchVer.Minor() {
			sha := branch.GetObject().GetSHA()
			Logf("found matching destination branch %q for tag %q with HEAD %q", branch.GetRef(), tag, sha)
			return sha
		}
	}
	Logf("using the %q branch for new tag %q", BranchMaster, tag)
	return masterSHA
}

// FindLatestBranch goes trough a list of branches and finds the latest
// based on its prefixMAJOR.MINOR format.
func FindLatestBranch(refs []*github.Reference, prefix string) (*github.Reference, error) {
	var result *github.Reference
	minV := version.MustParseSemantic("v0.0.0")

	for _, ref := range refs {
		v, err := BranchRefToVersion(ref, prefix)
		if err != nil {
			Warningf(err.Error())
			continue
		}

		if minV.LessThan(v) {
			minV = v
			r := *ref
			result = &r
		}
	}

	if result == nil {
		return nil, errors.Errorf("could not find any branches of the format %sMAJOR.MINOR", prefix)
	}
	return result, nil
}

// FindLatestTag goes trough a list of tags and finds the latest for a given branch version.
func FindLatestTag(refs []*github.Reference, branchV *version.Version) (*github.Reference, error) {
	var result *github.Reference
	minV := version.MustParseSemantic("v0.0.0")

	for _, ref := range refs {
		v, err := TagRefToVersion(ref)
		if err != nil {
			Warningf(err.Error())
			continue
		}

		if v.Major() != branchV.Major() || v.Minor() != branchV.Minor() {
			continue
		}

		if minV.LessThan(v) {
			minV = v
			r := *ref
			result = &r
		}
	}

	if result == nil {
		return nil, errors.Errorf("could not find any SemVer tag that matches branch version %d.%d",
			branchV.Major(), branchV.Minor())
	}
	return result, nil
}

// ShowPrompt shows a confirmation prompt to the user.
func ShowPrompt(message string) (bool, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf(message + " [y/n]: ")
	resp, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}
	resp = strings.ToLower(strings.TrimSpace(resp))
	if resp == "y" || resp == "yes" {
		return true, nil
	}
	return false, nil
}
