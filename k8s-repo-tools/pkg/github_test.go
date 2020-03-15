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
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"testing"

	"github.com/google/go-github/v29/github"
)

func TestGitHubGetCreateRelease(t *testing.T) {
	// Swap these two lines to enable debug logging.
	SetLogWriters(os.Stdout, os.Stderr)
	SetLogWriters(ioutil.Discard, ioutil.Discard)

	tests := []struct {
		name                 string
		data                 *Data
		refs                 []*github.Reference
		releases             []*github.RepositoryRelease
		releaseBody          string
		methodErrorsRefs     map[string]bool
		methodErrorsReleases map[string]bool
		skipDryRun           bool
		expectedRelease      *github.RepositoryRelease
		expectedError        bool
	}{
		{
			name: "valid: found release by matching tag",
			data: &Data{ReleaseTag: "v1.16.0"},
			refs: []*github.Reference{
				&github.Reference{Ref: github.String("refs/tags/v1.16.0"), Object: &github.GitObject{SHA: github.String("1234567890")}},
			},
			releases: []*github.RepositoryRelease{
				&github.RepositoryRelease{TagName: github.String("v1.16.0")},
			},
			expectedRelease: &github.RepositoryRelease{TagName: github.String("v1.16.0")},
		},
		{
			name: "valid: release is missing; create it from this tag",
			data: &Data{ReleaseTag: "v1.16.0"},
			refs: []*github.Reference{
				&github.Reference{Ref: github.String("refs/tags/v1.16.0"), Object: &github.GitObject{SHA: github.String("1234567890")}},
			},
			releaseBody: "foo",
			expectedRelease: &github.RepositoryRelease{
				TagName:    github.String("v1.16.0"),
				Name:       github.String("v1.16.0"),
				Body:       github.String("foo"),
				Draft:      github.Bool(false),
				Prerelease: github.Bool(false),
			},
		},
		{
			name: "invalid: fail creating release",
			data: &Data{ReleaseTag: "v1.16.0"},
			refs: []*github.Reference{
				&github.Reference{Ref: github.String("refs/tags/v1.16.0"), Object: &github.GitObject{SHA: github.String("1234567890")}},
			},
			methodErrorsReleases: map[string]bool{http.MethodPost: true},
			expectedError:        true,
			skipDryRun:           true,
		},
		{
			name: "invalid: tag not found in the list of refs",
			data: &Data{ReleaseTag: "v1.16.0"},
			refs: []*github.Reference{
				&github.Reference{Ref: github.String("refs/tags/v1.17.0"), Object: &github.GitObject{SHA: github.String("1234567890")}},
				&github.Reference{Ref: github.String("refs/tags/v1.15.0"), Object: &github.GitObject{SHA: github.String("1234567891")}},
			},
			expectedError: true,
		},
		{
			name:             "invalid: could not get the reference for this tag",
			data:             &Data{ReleaseTag: "v1.16.0"},
			methodErrorsRefs: map[string]bool{http.MethodGet: true},
			expectedError:    true,
		},
		{
			name:                 "invalid: could not get the release for this tag",
			data:                 &Data{ReleaseTag: "v1.16.0"},
			methodErrorsReleases: map[string]bool{http.MethodGet: true},
			expectedError:        true,
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
				tt.data.PrefixBranch = PrefixBranch
				tt.data.Force = true
				tt.data.DryRun = dryRunVal

				if tt.methodErrorsRefs == nil {
					tt.methodErrorsRefs = map[string]bool{}
				}
				if tt.methodErrorsReleases == nil {
					tt.methodErrorsReleases = map[string]bool{}
				}

				// create fake client and setup endpoint handlers
				NewClient(tt.data, NewTransport())
				const (
					testRefs     = "https://api.github.com/repos/org/dest/git/refs"
					testReleases = "https://api.github.com/repos/org/dest/releases"
				)
				handlerRefs := NewReferenceHandler(&tt.refs, tt.methodErrorsRefs)
				handlerReleases := NewReleaseHandler(&tt.releases, tt.methodErrorsReleases)
				tt.data.Transport.SetHandler(testRefs, handlerRefs)
				tt.data.Transport.SetHandler(testReleases, handlerReleases)

				rel, err := GitHubGetCreateRelease(tt.data, tt.data.Dest, tt.data.ReleaseTag, tt.releaseBody, dryRunVal)
				if (err != nil) != tt.expectedError {
					t.Errorf("expected error %v, got %v, error: %v", tt.expectedError, err != nil, err)
				}
				if err != nil {
					return
				}

				if !reflect.DeepEqual(tt.expectedRelease, rel) {
					t.Errorf("expected release:\n%+v\ngot:\n%+v\n", tt.expectedRelease, rel)
				}
			})
		}
	}
}
