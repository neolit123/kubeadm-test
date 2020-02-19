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
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"testing"

	"github.com/google/go-github/v29/github"
	"k8s.io/kubeadm/k8s-repo-tools/pkg"
)

func TestProcess(t *testing.T) {
	// Swap these two lines to enable debug logging.
	pkg.SetLogWriters(os.Stdout, os.Stderr)
	pkg.SetLogWriters(ioutil.Discard, ioutil.Discard)

	tests := []struct {
		name                string
		commitsMaster       []*github.RepositoryCommit
		commitsBranch       []*github.RepositoryCommit
		refsDest            []*github.Reference
		methodErrorsRef     map[string]bool
		methodErrorsCompare map[string]bool
		methodErrorsMerge   map[string]bool
		skipDryRun          bool
		mergeStatus         int
		mergeRequest        *github.RepositoryMergeRequest
		expectedBranch      *github.Reference
		expectedCommit      *github.RepositoryCommit
		expectedError       error
	}{
		{
			name:            "invalid: return error obtaining refs",
			commitsMaster:   []*github.RepositoryCommit{},
			commitsBranch:   []*github.RepositoryCommit{},
			refsDest:        []*github.Reference{},
			methodErrorsRef: map[string]bool{http.MethodGet: true},
			expectedError:   &genericError{},
		},
		{
			name:          "invalid: cannot find a SemVer tag for the latest release branch",
			commitsMaster: []*github.RepositoryCommit{},
			commitsBranch: []*github.RepositoryCommit{},
			refsDest: []*github.Reference{
				&github.Reference{Ref: github.String("refs/tags/v1.16.0"), Object: &github.GitObject{SHA: github.String("1234567890")}},
				&github.Reference{Ref: github.String("refs/heads/master"), Object: &github.GitObject{SHA: github.String("1234567890")}},
				&github.Reference{Ref: github.String("refs/heads/release-1.17"), Object: &github.GitObject{SHA: github.String("1234567890")}},
			},
			expectedError: &genericError{},
		},
		{
			name:          "invalid: return an error if there are no release branches",
			commitsMaster: []*github.RepositoryCommit{},
			commitsBranch: []*github.RepositoryCommit{},
			refsDest:      []*github.Reference{},
			expectedError: &releaseBranchError{},
		},
		{
			name:          "invalid: there is a release branch and tags but not in the ff window",
			commitsMaster: []*github.RepositoryCommit{},
			commitsBranch: []*github.RepositoryCommit{},
			refsDest: []*github.Reference{
				&github.Reference{Ref: github.String("refs/tags/v1.17.0-alpha.3"), Object: &github.GitObject{SHA: github.String("1234567890")}},
				&github.Reference{Ref: github.String("refs/heads/master"), Object: &github.GitObject{SHA: github.String("1234567890")}},
				&github.Reference{Ref: github.String("refs/heads/release-1.17"), Object: &github.GitObject{SHA: github.String("1234567890")}},
			},
			expectedError: &fastForwardWindowError{},
		},
		{
			name:          "invalid: return error on identical branches",
			commitsMaster: []*github.RepositoryCommit{&github.RepositoryCommit{SHA: github.String("some-sha")}},
			commitsBranch: []*github.RepositoryCommit{&github.RepositoryCommit{SHA: github.String("some-sha")}},
			refsDest: []*github.Reference{
				&github.Reference{Ref: github.String("refs/tags/v1.17.0-beta.0"), Object: &github.GitObject{SHA: github.String("1234567890")}},
				&github.Reference{Ref: github.String("refs/heads/master"), Object: &github.GitObject{SHA: github.String("1234567890")}},
				&github.Reference{Ref: github.String("refs/heads/release-1.17"), Object: &github.GitObject{SHA: github.String("1234567890")}},
			},
			expectedError: &identicalBranchesError{},
		},
		{
			name:          "valid: do not return error if master is behind",
			commitsMaster: []*github.RepositoryCommit{&github.RepositoryCommit{SHA: github.String("some-sha1")}},
			commitsBranch: []*github.RepositoryCommit{
				&github.RepositoryCommit{SHA: github.String("some-sha1")},
				&github.RepositoryCommit{SHA: github.String("some-sha2")},
			},
			refsDest: []*github.Reference{
				&github.Reference{Ref: github.String("refs/tags/v1.17.0-beta.0"), Object: &github.GitObject{SHA: github.String("1234567890")}},
				&github.Reference{Ref: github.String("refs/heads/master"), Object: &github.GitObject{SHA: github.String("1234567890")}},
				&github.Reference{Ref: github.String("refs/heads/release-1.17"), Object: &github.GitObject{SHA: github.String("1234567890")}},
			},
			mergeRequest: &github.RepositoryMergeRequest{
				Base:          github.String("refs/heads/release-1.17"),
				Head:          github.String(pkg.BranchMaster),
				CommitMessage: github.String(pkg.FormatMergeCommitMessage("refs/heads/release-1.17", pkg.BranchMaster)),
			},
			expectedCommit: &github.RepositoryCommit{
				SHA:    github.String("dry-run-sha"),
				Commit: &github.Commit{Message: github.String(pkg.FormatMergeCommitMessage("refs/heads/release-1.17", pkg.BranchMaster))},
			},
			expectedBranch: &github.Reference{
				Ref:    github.String("refs/heads/release-1.17"),
				Object: &github.GitObject{SHA: github.String("1234567890")},
			},
		},
		{
			name:          "invalid: return error comparing branches",
			commitsMaster: []*github.RepositoryCommit{},
			commitsBranch: []*github.RepositoryCommit{},
			refsDest: []*github.Reference{
				&github.Reference{Ref: github.String("refs/tags/v1.17.0-beta.0"), Object: &github.GitObject{SHA: github.String("1234567890")}},
				&github.Reference{Ref: github.String("refs/heads/master"), Object: &github.GitObject{SHA: github.String("1234567890")}},
				&github.Reference{Ref: github.String("refs/heads/release-1.17"), Object: &github.GitObject{SHA: github.String("1234567890")}},
			},
			methodErrorsCompare: map[string]bool{http.MethodGet: true},
			expectedError:       &genericError{},
		},
		{
			name: "invalid: return error merging branches",
			commitsMaster: []*github.RepositoryCommit{
				&github.RepositoryCommit{SHA: github.String("some-sha")},
				&github.RepositoryCommit{SHA: github.String("some-sha")},
			},
			commitsBranch: []*github.RepositoryCommit{
				&github.RepositoryCommit{SHA: github.String("some-sha")},
			},
			refsDest: []*github.Reference{
				&github.Reference{Ref: github.String("refs/tags/v1.17.0-beta.0"), Object: &github.GitObject{SHA: github.String("1234567890")}},
				&github.Reference{Ref: github.String("refs/heads/master"), Object: &github.GitObject{SHA: github.String("1234567890")}},
				&github.Reference{Ref: github.String("refs/heads/release-1.17"), Object: &github.GitObject{SHA: github.String("1234567890")}},
			},
			methodErrorsMerge: map[string]bool{http.MethodPost: true},
			expectedError:     &genericError{},
			skipDryRun:        true,
		},
		{
			name: "invalid: return no-content status when merging branches",
			commitsMaster: []*github.RepositoryCommit{
				&github.RepositoryCommit{SHA: github.String("some-sha")},
				&github.RepositoryCommit{SHA: github.String("some-sha")},
			},
			commitsBranch: []*github.RepositoryCommit{
				&github.RepositoryCommit{SHA: github.String("some-sha")},
			},
			refsDest: []*github.Reference{
				&github.Reference{Ref: github.String("refs/tags/v1.17.0-beta.0"), Object: &github.GitObject{SHA: github.String("1234567890")}},
				&github.Reference{Ref: github.String("refs/heads/master"), Object: &github.GitObject{SHA: github.String("1234567890")}},
				&github.Reference{Ref: github.String("refs/heads/release-1.17"), Object: &github.GitObject{SHA: github.String("1234567890")}},
			},
			mergeRequest: &github.RepositoryMergeRequest{
				Base:          github.String("refs/heads/release-1.17"),
				Head:          github.String(pkg.BranchMaster),
				CommitMessage: github.String(pkg.FormatMergeCommitMessage("refs/heads/release-1.17", pkg.BranchMaster)),
			},
			mergeStatus:   http.StatusNoContent,
			expectedError: &noContentError{},
			skipDryRun:    true,
		},
		{
			name: "invalid: return unhandled status when merging branches",
			commitsMaster: []*github.RepositoryCommit{
				&github.RepositoryCommit{SHA: github.String("some-sha")},
				&github.RepositoryCommit{SHA: github.String("some-sha")},
			},
			commitsBranch: []*github.RepositoryCommit{
				&github.RepositoryCommit{SHA: github.String("some-sha")},
			},
			refsDest: []*github.Reference{
				&github.Reference{Ref: github.String("refs/tags/v1.17.0-beta.0"), Object: &github.GitObject{SHA: github.String("1234567890")}},
				&github.Reference{Ref: github.String("refs/heads/master"), Object: &github.GitObject{SHA: github.String("1234567890")}},
				&github.Reference{Ref: github.String("refs/heads/release-1.17"), Object: &github.GitObject{SHA: github.String("1234567890")}},
			},
			mergeRequest: &github.RepositoryMergeRequest{
				Base:          github.String("refs/heads/release-1.17"),
				Head:          github.String(pkg.BranchMaster),
				CommitMessage: github.String(pkg.FormatMergeCommitMessage("refs/heads/release-1.17", pkg.BranchMaster)),
			},
			mergeStatus:   http.StatusPartialContent,
			expectedError: &genericError{},
			skipDryRun:    true,
		},
		{
			name: "valid: successful merge of branches",
			commitsMaster: []*github.RepositoryCommit{
				&github.RepositoryCommit{SHA: github.String("some-sha")},
				&github.RepositoryCommit{SHA: github.String("some-sha")},
			},
			commitsBranch: []*github.RepositoryCommit{
				&github.RepositoryCommit{SHA: github.String("some-sha")},
			},
			refsDest: []*github.Reference{
				&github.Reference{Ref: github.String("refs/tags/v1.17.0-beta.0"), Object: &github.GitObject{SHA: github.String("1234567890")}},
				&github.Reference{Ref: github.String("refs/heads/master"), Object: &github.GitObject{SHA: github.String("1234567890")}},
				&github.Reference{Ref: github.String("refs/heads/release-1.17"), Object: &github.GitObject{SHA: github.String("1234567890")}},
			},
			mergeRequest: &github.RepositoryMergeRequest{
				Base:          github.String("refs/heads/release-1.17"),
				Head:          github.String(pkg.BranchMaster),
				CommitMessage: github.String(pkg.FormatMergeCommitMessage("refs/heads/release-1.17", pkg.BranchMaster)),
			},
			mergeStatus: http.StatusCreated,
			expectedCommit: &github.RepositoryCommit{
				SHA:    github.String("dry-run-sha"),
				Commit: &github.Commit{Message: github.String(pkg.FormatMergeCommitMessage("refs/heads/release-1.17", pkg.BranchMaster))},
			},
			expectedBranch: &github.Reference{
				Ref:    github.String("refs/heads/release-1.17"),
				Object: &github.GitObject{SHA: github.String("1234567890")},
			},
		},
	}

	// Make sure there are consistent results between dry-run and regular mode.
	for _, dryRunVal := range []bool{false, true} {
		for _, tt := range tests {
			t.Run(fmt.Sprintf("%s (dryRun=%v)", tt.name, dryRunVal), func(t *testing.T) {
				// Some operations like POST will always return non-error in dry-run mode.
				// Skip such tests.
				if tt.skipDryRun && dryRunVal {
					t.Skip()
				}

				data := &pkg.Data{}
				data.Dest = "org/dest"
				data.PrefixBranch = pkg.PrefixBranch
				data.Force = true
				data.DryRun = dryRunVal

				if tt.methodErrorsRef == nil {
					tt.methodErrorsRef = map[string]bool{}
				}
				if tt.methodErrorsCompare == nil {
					tt.methodErrorsCompare = map[string]bool{}
				}
				if tt.methodErrorsMerge == nil {
					tt.methodErrorsMerge = map[string]bool{}
				}

				// create fake client and setup endpoint handlers
				pkg.NewClient(data, pkg.NewTransport())
				const (
					testRefs    = "https://api.github.com/repos/org/dest/git/refs"
					testCommits = "https://api.github.com/repos/org/dest/compare"
					testMerges  = "https://api.github.com/repos/org/dest/merges"
				)

				handlerRefs := pkg.NewReferenceHandler(&tt.refsDest, tt.methodErrorsRef)
				handlerCompare := pkg.NewCompareHandler(&tt.commitsMaster, &tt.commitsBranch, tt.methodErrorsCompare)
				handlerMerge := pkg.NewMergeHandler(tt.mergeRequest, tt.mergeStatus, tt.methodErrorsMerge)

				data.Transport.SetHandler(testRefs, handlerRefs)
				data.Transport.SetHandler(testCommits, handlerCompare)
				data.Transport.SetHandler(testMerges, handlerMerge)

				ref, commit, err := process(data)
				if err != nil {
					pkg.Errorf("TEST: process error (%v): %v", reflect.TypeOf(err), err)
				}
				if (err != nil) != (tt.expectedError != nil) {
					t.Fatalf("expected error %v, got %v", tt.expectedError != nil, err != nil)
				}
				expectedErrorType := reflect.TypeOf(tt.expectedError)
				errorType := reflect.TypeOf(err)
				if expectedErrorType != errorType {
					t.Errorf("expected error type %v, got %v, error: %v", expectedErrorType, errorType, err)
				}
				if err != nil {
					return
				}

				if !reflect.DeepEqual(ref, tt.expectedBranch) {
					t.Errorf("expected ref:\n%v\ngot:\n%v\n", tt.expectedBranch, ref)
				}
				if !reflect.DeepEqual(commit, tt.expectedCommit) {
					t.Errorf("expected commit:\n%v\ngot:\n%v\n", tt.expectedCommit, commit)
				}
			})
		}
	}
}
