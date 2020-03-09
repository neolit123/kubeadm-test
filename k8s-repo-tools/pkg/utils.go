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
			Warningf("skipping ref %s; version is older than the minimum version", ref.GetRef())
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
	return TagToVersion(ref.GetRef())
}

// TagToVersion converts a tag string to Version.
func TagToVersion(tag string) (*version.Version, error) {
	ver := strings.TrimPrefix(tag, "refs/tags/")
	if strings.Count(ver, ".") < 2 { // a version without a .PATCH component?
		ver = ver + ".0"
	}

	v, err := version.ParseSemantic(ver)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot parse tag reference %q", tag)
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
	tagVer, err := version.ParseSemantic(tagStr)
	if err != nil {
		Warningf("skipping non-versioned input ref %s", tag.GetRef(), err)
		goto exit
	}
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
		branchVer, err := version.ParseSemantic(branchRef)
		if err != nil {
			Warningf("skipping ref %s: %v", branch.GetRef(), err)
			continue
		}
		if tagVer.Major() == branchVer.Major() && tagVer.Minor() == branchVer.Minor() {
			sha := branch.GetObject().GetSHA()
			Logf("found matching destination branch %q for tag %q with HEAD %q", branch.GetRef(), tag, sha)
			return sha
		}
	}
exit:
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

// FindReleaseNotesSinceRef takes a k8s release SemVer tag reference and determines
// the release SemVer tag which to use for a release notes range from a list of
// tag references.
// Note that nil can be returned even for err != nil.
//
// This logic needs to be adapted if the k8s release process changes.
//
// tag               | returned tag    | comment
// -------------------------------------------------
// v1.17.0-alpha.0   | v1.17.0-alpha.0 | no changelog
// v1.17.0-alpha.1   | v1.16.0         | use previous minor
// v1.17.0-<pre>     | v1.17.0-<pre-1> | previous pre-release
// v1.17.0           | v1.16.0         | previous MINOR
// v1.17.1           | v1.17.0         | previous PATCH
//
func FindReleaseNotesSinceRef(ref *github.Reference, refs []*github.Reference) (*github.Reference, error) {
	var err error
	ver, err := version.ParseSemantic(ref.GetRef())
	if err != nil {
		return nil, err
	}

	var result *github.Reference

	// Not a pre-release.
	if len(ver.PreRelease()) == 0 {
		// Handle MINOR release.
		if ver.Patch() == 0 {
			major := int(ver.Major())
			minor := int(ver.Minor()) - 1
			// Handle MAJOR release.
			if minor < 0 {
				major--
				if major < 0 {
					goto exit
				}
				largest := version.MustParseSemantic(fmt.Sprintf("v%d.0.0", major))
				result = findLargestMinorForMajorRef(largest, refs)
			} else {
				target := version.MustParseSemantic(fmt.Sprintf("v%d.%d.0", major, minor))
				result = findExactVersionRef(target, refs)
			}
		} else {
			// Handle PATCH release.
			target := version.MustParseSemantic(ver.String()).WithPatch(ver.Patch() - 1)
			result = findExactVersionRef(target, refs)
		}
	} else {
		// Split the pre-release into "pre[0].pre[1]" e.g. "alpha.1".
		// This format is guaranteed by the initial ParseSemantic() for "ver".
		pre := strings.Split(ver.PreRelease(), ".")
		// Handle alpha.0.
		if pre[0] == "alpha" && pre[1] == "0" {
			goto exit
		}
		// Handle alpha.1.
		if pre[0] == "alpha" && pre[1] == "1" {
			major := int(ver.Major())
			minor := int(ver.Minor()) - 1
			if minor < 0 {
				major--
				if major < 0 {
					goto exit
				}
				largest := version.MustParseSemantic(fmt.Sprintf("v%d.0.0", major))
				result = findLargestMinorForMajorRef(largest, refs)
			} else {
				target := version.MustParseSemantic(fmt.Sprintf("v%d.%d.0", major, minor))
				result = findExactVersionRef(target, refs)
			}
			goto exit
		}
		// Handle other pre-releases.
		// k8s does not have pre-releases for PATCH releases.
		result = findPreviousPreRelease(ver, refs)
	}
exit:
	if result == nil {
		Warningf("could not find a release notes range reference for %v; returning the same reference", ref)
		result = ref
	}
	return result, nil
}

func findLargestMinorForMajorRef(largest *version.Version, refs []*github.Reference) *github.Reference {
	var result *github.Reference

	for i := range refs {
		tag := refs[i].GetRef()
		tag = strings.TrimPrefix(tag, "refs/tags/")
		if strings.Count(tag, ".") < 2 { // a version without a .PATCH component?
			tag = tag + ".0"
		}
		ver, err := version.ParseSemantic(tag)
		if err != nil {
			Warningf("skipping ref %s: %v", refs[i], err)
			continue
		}
		if largest.LessThan(ver) && largest.Major() == ver.Major() {
			largest = ver
			result = refs[i]
		}
	}
	return result
}

func findExactVersionRef(target *version.Version, refs []*github.Reference) *github.Reference {
	for i := range refs {
		tag := refs[i].GetRef()
		tag = strings.TrimPrefix(tag, "refs/tags/")
		if strings.Count(tag, ".") < 2 { // a version without a .PATCH component?
			tag = tag + ".0"
		}
		ver, err := version.ParseSemantic(tag)
		if err != nil {
			Warningf("skipping ref %s: %v", refs[i], err)
			continue
		}
		if target.String() == ver.String() {
			return refs[i]
		}
	}
	return nil
}

func findPreviousPreRelease(target *version.Version, refs []*github.Reference) *github.Reference {
	largest := version.MustParseSemantic("v0.0.0")
	var result *github.Reference

	for i := range refs {
		tag := refs[i].GetRef()
		tag = strings.TrimPrefix(tag, "refs/tags/")
		if strings.Count(tag, ".") < 2 { // a version without a .PATCH component?
			tag = tag + ".0"
		}
		ver, err := version.ParseSemantic(tag)
		if err != nil {
			Warningf("skipping ref %s: %v", refs[i], err)
			continue
		}
		if largest.LessThan(ver) && ver.LessThan(target) {
			largest = ver
			result = refs[i]
		}
	}
	return result
}

// FormatMergeCommitMessage creates a commit message that
// indicates which branches are being merged.
func FormatMergeCommitMessage(base, head string) string {
	head = strings.TrimPrefix(head, "refs/heads/")
	base = strings.TrimPrefix(base, "refs/heads/")
	return fmt.Sprintf("Merge branch %q into %s", head, base)
}
