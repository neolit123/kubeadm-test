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
	"testing"

	"github.com/google/go-github/v29/github"
	"k8s.io/kubeadm/k8s-repo-tools/pkg"
)

func TestGetReleaseNotesToolSHAs(t *testing.T) {
	// Swap these two lines to enable debug logging.
	pkg.SetLogWriters(os.Stdout, os.Stderr)
	pkg.SetLogWriters(ioutil.Discard, ioutil.Discard)

	tests := []struct {
		name             string
		data             *pkg.Data
		refs             []*github.Reference
		expectedStartSHA string
		expectedEndSHA   string
		methodErrors     map[string]bool
		skipDryRun       bool // Not really needed here, but can still catch future divergence
		expectedError    bool
	}{
		{
			name: "valid: found start SHA and end SHA for a MINOR release",
			data: &pkg.Data{ReleaseTag: "v1.16.0"},
			refs: []*github.Reference{
				&github.Reference{Ref: github.String("refs/tags/v1.17.0"), Object: &github.GitObject{SHA: github.String("1234567890")}},
				&github.Reference{Ref: github.String("refs/tags/v1.16.0"), Object: &github.GitObject{SHA: github.String("1234567891")}},
				&github.Reference{Ref: github.String("refs/tags/v1.15.0"), Object: &github.GitObject{SHA: github.String("1234567892")}},
			},
			expectedStartSHA: "1234567891",
			expectedEndSHA:   "1234567892",
		},
		{
			name: "valid: expect start and end SHA to match for an alpha.0 release",
			data: &pkg.Data{ReleaseTag: "v1.17.0-alpha.0"},
			refs: []*github.Reference{
				&github.Reference{Ref: github.String("refs/tags/v1.17.0-alpha.0"), Object: &github.GitObject{SHA: github.String("1234567890")}},
			},
			expectedStartSHA: "1234567890",
			expectedEndSHA:   "1234567890",
		},
		{
			name: "invalid: no matching ref for tag",
			data: &pkg.Data{ReleaseTag: "v1.16.0"},
			refs: []*github.Reference{
				&github.Reference{Ref: github.String("refs/tags/v1.15.0"), Object: &github.GitObject{SHA: github.String("1234567892")}},
			},
			expectedError: true,
		},
		{
			name: "invalid: expect error on non-SemVer tag",
			data: &pkg.Data{ReleaseTag: "foo"},
			refs: []*github.Reference{
				&github.Reference{Ref: github.String("refs/tags/v1.17.0"), Object: &github.GitObject{SHA: github.String("1234567890")}},
			},
			expectedError: true,
		},
		{
			name: "invalid: expect error on simulated 404 GET",
			data: &pkg.Data{ReleaseTag: "v1.17.0"},
			refs: []*github.Reference{
				&github.Reference{Ref: github.String("refs/tags/v1.17.0"), Object: &github.GitObject{SHA: github.String("1234567890")}},
			},
			methodErrors:  map[string]bool{http.MethodGet: true},
			expectedError: true,
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
				tt.data.Dest = "org/dest"
				tt.data.PrefixBranch = pkg.PrefixBranch
				tt.data.Force = true
				tt.data.DryRun = dryRunVal

				if tt.methodErrors == nil {
					tt.methodErrors = map[string]bool{}
				}

				// create fake client and setup endpoint handlers
				pkg.NewClient(tt.data, pkg.NewTransport())
				const testRefs = "https://api.github.com/repos/org/dest/git/refs"
				handler := pkg.NewReferenceHandler(&tt.refs, tt.methodErrors)
				tt.data.Transport.SetHandler(testRefs, handler)

				startSHA, endSHA, err := getReleaseNotesToolSHAs(tt.data)
				if (err != nil) != tt.expectedError {
					t.Errorf("expected error %v, got %v, error: %v", tt.expectedError, err != nil, err)
				}
				if err != nil {
					return
				}

				if startSHA != tt.expectedStartSHA {
					t.Errorf("expected start SHA %s, got %s", tt.expectedStartSHA, startSHA)
				}
				if endSHA != tt.expectedEndSHA {
					t.Errorf("expected end SHA %s, got %s", tt.expectedEndSHA, endSHA)
				}
			})
		}
	}
}
