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
	// "fmt"
	"io/ioutil"
	// "net/http"
	"os"
	// "reflect"
	"testing"

	// "github.com/google/go-github/v29/github"
	"k8s.io/kubeadm/k8s-repo-tools/pkg"
)

func TestGetReleaseNotesToolSHAs(t *testing.T) {
	// Swap these two lines to enable debug logging.
	pkg.SetLogWriters(os.Stdout, os.Stderr)
	pkg.SetLogWriters(ioutil.Discard, ioutil.Discard)

	/*
		tests := []struct {
			name             string
			data             *pkg.Data
			refsSrc          []*github.Reference
			refsDest         []*github.Reference
			expectedRefs     []*github.Reference
			methodErrorsSrc  map[string]bool
			methodErrorsDest map[string]bool
			skipDryRun       bool
			expectedError    bool
		}{
			{
				name: "valid: new branches and tags",
				data: &pkg.Data{MinVersion: "v1.16.1"},
				refsSrc: []*github.Reference{
					&github.Reference{Ref: github.String("refs/tags/v1.15.1"), Object: &github.GitObject{SHA: github.String("1234567890")}},
					&github.Reference{Ref: github.String("refs/tags/v1.16.2"), Object: &github.GitObject{SHA: github.String("1234567890")}},
					&github.Reference{Ref: github.String("refs/tags/v1.16.3"), Object: &github.GitObject{SHA: github.String("1234567890")}},
					&github.Reference{Ref: github.String("refs/tags/v1.17.1"), Object: &github.GitObject{SHA: github.String("1234567890")}},
					&github.Reference{Ref: github.String("refs/tags/v1.17.2"), Object: &github.GitObject{SHA: github.String("1234567890")}},
					&github.Reference{Ref: github.String("refs/heads/master"), Object: &github.GitObject{SHA: github.String("1234567890")}},
					&github.Reference{Ref: github.String("refs/heads/release-1.16"), Object: &github.GitObject{SHA: github.String("1234567890")}},
					&github.Reference{Ref: github.String("refs/heads/release-1.17"), Object: &github.GitObject{SHA: github.String("1234567890")}},
				},
				refsDest: []*github.Reference{
					&github.Reference{Ref: github.String("refs/heads/master"), Object: &github.GitObject{SHA: github.String("0000")}},
				},
				expectedRefs: []*github.Reference{
					&github.Reference{Ref: github.String("refs/heads/release-1.16"), Object: &github.GitObject{SHA: github.String("0000")}},
					&github.Reference{Ref: github.String("refs/heads/release-1.17"), Object: &github.GitObject{SHA: github.String("0000")}},
					&github.Reference{Ref: github.String("refs/tags/v1.16.2"), Object: &github.GitObject{SHA: github.String("0000")}},
					&github.Reference{Ref: github.String("refs/tags/v1.16.3"), Object: &github.GitObject{SHA: github.String("0000")}},
					&github.Reference{Ref: github.String("refs/tags/v1.17.1"), Object: &github.GitObject{SHA: github.String("0000")}},
					&github.Reference{Ref: github.String("refs/tags/v1.17.2"), Object: &github.GitObject{SHA: github.String("0000")}},
				},
			},
			{
				name: "valid: new tag with min version",
				data: &pkg.Data{MinVersion: "v1.17.2"},
				refsSrc: []*github.Reference{
					&github.Reference{Ref: github.String("refs/tags/v1.17.2"), Object: &github.GitObject{SHA: github.String("1234567890")}},
					&github.Reference{Ref: github.String("refs/tags/v1.17.2-rc.1"), Object: &github.GitObject{SHA: github.String("1234567890")}},
					&github.Reference{Ref: github.String("refs/heads/master"), Object: &github.GitObject{SHA: github.String("1234567890")}},
					&github.Reference{Ref: github.String("refs/heads/release-1.17"), Object: &github.GitObject{SHA: github.String("1234567890")}},
					&github.Reference{Ref: github.String("refs/heads/release-1.16"), Object: &github.GitObject{SHA: github.String("1234567890")}},
				},
				refsDest: []*github.Reference{
					&github.Reference{Ref: github.String("refs/heads/master"), Object: &github.GitObject{SHA: github.String("1234567890")}},
					&github.Reference{Ref: github.String("refs/heads/release-1.17"), Object: &github.GitObject{SHA: github.String("17")}},
				},
				expectedRefs: []*github.Reference{
					&github.Reference{Ref: github.String("refs/tags/v1.17.2"), Object: &github.GitObject{SHA: github.String("17")}},
				},
			},
			{
				name: "valid: no new tags due to min version filtering",
				data: &pkg.Data{MinVersion: "v1.17.3"},
				refsSrc: []*github.Reference{
					&github.Reference{Ref: github.String("refs/tags/v1.17.2"), Object: &github.GitObject{SHA: github.String("1234567890")}},
					&github.Reference{Ref: github.String("refs/tags/v1.17.1"), Object: &github.GitObject{SHA: github.String("1234567890")}},
					&github.Reference{Ref: github.String("refs/heads/master"), Object: &github.GitObject{SHA: github.String("1234567890")}},
					&github.Reference{Ref: github.String("refs/heads/release-1.17"), Object: &github.GitObject{SHA: github.String("1234567890")}},
				},
				refsDest: []*github.Reference{
					&github.Reference{Ref: github.String("refs/heads/master"), Object: &github.GitObject{SHA: github.String("1234567890")}},
					&github.Reference{Ref: github.String("refs/heads/release-1.17"), Object: &github.GitObject{SHA: github.String("1234567890")}},
				},
				expectedRefs: []*github.Reference{},
			},
			{
				name: "valid: non-version tags and branches are ignored",
				data: &pkg.Data{MinVersion: "v1.17.0"},
				refsSrc: []*github.Reference{
					&github.Reference{Ref: github.String("refs/tags/foo"), Object: &github.GitObject{SHA: github.String("1234567890")}},
					&github.Reference{Ref: github.String("refs/tags/bar"), Object: &github.GitObject{SHA: github.String("1234567890")}},
					&github.Reference{Ref: github.String("refs/tags/v1.17.1"), Object: &github.GitObject{SHA: github.String("1234567890")}},
					&github.Reference{Ref: github.String("refs/heads/master"), Object: &github.GitObject{SHA: github.String("1234567890")}},
					&github.Reference{Ref: github.String("refs/heads/foo"), Object: &github.GitObject{SHA: github.String("1234567890")}},
					&github.Reference{Ref: github.String("refs/heads/release-foo"), Object: &github.GitObject{SHA: github.String("1234567890")}},
					&github.Reference{Ref: github.String("refs/heads/release-1.17"), Object: &github.GitObject{SHA: github.String("1234567890")}},
				},
				refsDest: []*github.Reference{
					&github.Reference{Ref: github.String("refs/heads/master"), Object: &github.GitObject{SHA: github.String("1234567890")}},
					&github.Reference{Ref: github.String("refs/heads/test-branch"), Object: &github.GitObject{SHA: github.String("1234567890")}},
				},
				expectedRefs: []*github.Reference{
					&github.Reference{Ref: github.String("refs/heads/release-1.17"), Object: &github.GitObject{SHA: github.String("1234567890")}},
					&github.Reference{Ref: github.String("refs/tags/v1.17.1"), Object: &github.GitObject{SHA: github.String("1234567890")}},
				},
			},
			{
				name: "invalid: dest repo missing master branch",
				data: &pkg.Data{MinVersion: "v1.17.0"},
				refsSrc: []*github.Reference{
					&github.Reference{Ref: github.String("refs/tags/v1.17.2"), Object: &github.GitObject{SHA: github.String("1234567890")}},
					&github.Reference{Ref: github.String("refs/tags/v1.17.1"), Object: &github.GitObject{SHA: github.String("1234567890")}},
					&github.Reference{Ref: github.String("refs/heads/master"), Object: &github.GitObject{SHA: github.String("1234567890")}},
					&github.Reference{Ref: github.String("refs/heads/release-1.17"), Object: &github.GitObject{SHA: github.String("1234567890")}},
				},
				refsDest:      []*github.Reference{},
				expectedError: true,
			},
			{
				name:            "invalid: cannot get refs from source repo",
				data:            &pkg.Data{MinVersion: "v1.17.0"},
				refsSrc:         []*github.Reference{},
				refsDest:        []*github.Reference{},
				methodErrorsSrc: map[string]bool{http.MethodGet: true},
				expectedError:   true,
			},
			{
				name:             "invalid: cannot get refs from destination repo",
				data:             &pkg.Data{MinVersion: "v1.17.0"},
				refsSrc:          []*github.Reference{},
				refsDest:         []*github.Reference{},
				methodErrorsDest: map[string]bool{http.MethodGet: true},
				expectedError:    true,
			},
			{
				name: "invalid: cannot post refs to destination repo",
				data: &pkg.Data{MinVersion: "v1.17.0"},
				refsSrc: []*github.Reference{
					&github.Reference{Ref: github.String("refs/tags/v1.17.2"), Object: &github.GitObject{SHA: github.String("1234567890")}},
				},
				refsDest: []*github.Reference{
					&github.Reference{Ref: github.String("refs/heads/master"), Object: &github.GitObject{SHA: github.String("0000")}},
				},
				methodErrorsDest: map[string]bool{http.MethodPost: true},
				expectedError:    true,
				skipDryRun:       true,
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

					// Override/hardcode some values.
					tt.data.Source = "org/src"
					tt.data.Dest = "org/dest"
					tt.data.PrefixBranch = pkg.PrefixBranch
					tt.data.Force = true
					tt.data.DryRun = dryRunVal

					if tt.methodErrorsSrc == nil {
						tt.methodErrorsSrc = map[string]bool{}
					}
					if tt.methodErrorsDest == nil {
						tt.methodErrorsDest = map[string]bool{}
					}

					// create fake client and setup endpoint handlers
					pkg.NewClient(tt.data, pkg.NewTransport())
					const (
						testRefsSrc  = "https://api.github.com/repos/org/src/git/refs"
						testRefsDest = "https://api.github.com/repos/org/dest/git/refs"
					)
					handlerSrc := pkg.NewReferenceHandler(&tt.refsSrc, tt.methodErrorsSrc)
					handlerDest := pkg.NewReferenceHandler(&tt.refsDest, tt.methodErrorsDest)
					tt.data.Transport.SetHandler(testRefsSrc, handlerSrc)
					tt.data.Transport.SetHandler(testRefsDest, handlerDest)

					refs, err := process(tt.data)
					if (err != nil) != tt.expectedError {
						t.Errorf("expected error %v, got %v, error: %v", tt.expectedError, err != nil, err)
					}
					if err != nil {
						return
					}

					if !reflect.DeepEqual(refs, tt.expectedRefs) {
						t.Errorf("expected tags:\n%v\ngot:\n%v\n", tt.expectedRefs, refs)
					}
				})
			}
		}
	*/
}
