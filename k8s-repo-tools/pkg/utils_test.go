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
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/google/go-github/v29/github"
)

func TestFindReleaseNotesSinceRef(t *testing.T) {
	// Swap these two lines to enable debug logging.
	SetLogWriters(os.Stdout, os.Stderr)
	SetLogWriters(ioutil.Discard, ioutil.Discard)

	tests := []struct {
		name          string
		ref           *github.Reference
		expectedRef   *github.Reference
		refs          []*github.Reference
		expectedError bool
	}{
		{
			name:        "valid: input is v0.0.0, return the same version",
			ref:         &github.Reference{Ref: github.String("refs/tags/v0.0.0")},
			expectedRef: &github.Reference{Ref: github.String("refs/tags/v0.0.0")},
			refs:        []*github.Reference{},
		},
		{
			name:        "valid: input is MINOR release, expect previous MINOR release",
			ref:         &github.Reference{Ref: github.String("refs/tags/v1.17.0")},
			expectedRef: &github.Reference{Ref: github.String("refs/tags/v1.16.0")},
			refs: []*github.Reference{
				&github.Reference{Ref: github.String("refs/tags/some-non-semver-ref")},
				&github.Reference{Ref: github.String("refs/tags/v1.16.0")},
				&github.Reference{Ref: github.String("refs/tags/v1.17.0")},
			},
		},
		{
			name:        "valid: input is a MAJOR release, expect previous MINOR release",
			ref:         &github.Reference{Ref: github.String("refs/tags/v2.0.0")},
			expectedRef: &github.Reference{Ref: github.String("refs/tags/v1.64.0")},
			refs: []*github.Reference{
				&github.Reference{Ref: github.String("refs/tags/some-non-semver-ref")},
				&github.Reference{Ref: github.String("refs/tags/v1.63.0")},
				&github.Reference{Ref: github.String("refs/tags/v1.64.0")},
				&github.Reference{Ref: github.String("refs/tags/v2.0.0")},
			},
		},
		{
			name:        "valid: could not find a range reference",
			ref:         &github.Reference{Ref: github.String("refs/tags/v1.23.0")},
			expectedRef: &github.Reference{Ref: github.String("refs/tags/v1.23.0")},
			refs: []*github.Reference{
				&github.Reference{Ref: github.String("refs/tags/v1.63.0")},
			},
		},
		{
			name:        "valid: return the previous PATCH",
			ref:         &github.Reference{Ref: github.String("refs/tags/v1.23.2")},
			expectedRef: &github.Reference{Ref: github.String("refs/tags/v1.23.1")},
			refs: []*github.Reference{
				&github.Reference{Ref: github.String("refs/tags/some-non-semver-ref")},
				&github.Reference{Ref: github.String("refs/tags/v1.23.2")},
				&github.Reference{Ref: github.String("refs/tags/v1.23.1")},
				&github.Reference{Ref: github.String("refs/tags/v1.23.0")},
			},
		},
		{
			name:        "valid: alpha.0 is unhandled",
			ref:         &github.Reference{Ref: github.String("refs/tags/v1.23.0-alpha.0")},
			expectedRef: &github.Reference{Ref: github.String("refs/tags/v1.23.0-alpha.0")},
			refs:        []*github.Reference{},
		},
		{
			name:        "valid: alpha.1 should return previous MINOR",
			ref:         &github.Reference{Ref: github.String("refs/tags/v1.23.0-alpha.1")},
			expectedRef: &github.Reference{Ref: github.String("refs/tags/v1.22.0")},
			refs: []*github.Reference{
				&github.Reference{Ref: github.String("refs/tags/v1.23.0-alpha.0")},
				&github.Reference{Ref: github.String("refs/tags/v1.23.0")},
				&github.Reference{Ref: github.String("refs/tags/v1.22.0")},
			},
		},
		{
			name:        "valid: alpha.1 should return previous MINOR (with MAJOR handling)",
			ref:         &github.Reference{Ref: github.String("refs/tags/v2.0.0-alpha.1")},
			expectedRef: &github.Reference{Ref: github.String("refs/tags/v1.23.0")},
			refs: []*github.Reference{
				&github.Reference{Ref: github.String("refs/tags/v2.0.0-alpha.0")},
				&github.Reference{Ref: github.String("refs/tags/v2.0.0-alpha.1")},
				&github.Reference{Ref: github.String("refs/tags/v1.23.0")},
				&github.Reference{Ref: github.String("refs/tags/v1.22.0")},
			},
		},
		{
			name:        "valid: other pre-releases should return the previous pre-release[1]",
			ref:         &github.Reference{Ref: github.String("refs/tags/v1.23.0-beta.0")},
			expectedRef: &github.Reference{Ref: github.String("refs/tags/v1.23.0-alpha.3")},
			refs: []*github.Reference{
				&github.Reference{Ref: github.String("refs/tags/some-non-semver-ref")},
				&github.Reference{Ref: github.String("refs/tags/v1.23.0-beta.0")},
				&github.Reference{Ref: github.String("refs/tags/v1.23.0-alpha.3")},
				&github.Reference{Ref: github.String("refs/tags/v1.22.0-rc.2")},
				&github.Reference{Ref: github.String("refs/tags/v1.22.0-rc.1")},
			},
		},
		{
			name:        "valid: other pre-releases should return the previous pre-release[2]",
			ref:         &github.Reference{Ref: github.String("refs/tags/v1.24.0-rc.1")},
			expectedRef: &github.Reference{Ref: github.String("refs/tags/v1.24.0-beta.1")},
			refs: []*github.Reference{
				&github.Reference{Ref: github.String("refs/tags/v1.24.0-beta.1")},
				&github.Reference{Ref: github.String("refs/tags/v1.24.0-alpha.3")},
				&github.Reference{Ref: github.String("refs/tags/v1.24.0-alpha.2")},
			},
		},
		{
			name:          "valid: bad pre-release format should return error",
			ref:           &github.Reference{Ref: github.String("refs/tags/v1.23.0-alpha:0")},
			expectedError: true,
		},
		{
			name:          "invalid: input is not SemVer",
			ref:           &github.Reference{Ref: github.String("refs/tags/v11111")},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref, err := FindReleaseNotesSinceRef(tt.ref, tt.refs)
			if (err != nil) != tt.expectedError {
				t.Errorf("expected error %v, got %v, error: %v", tt.expectedError, err != nil, err)
			}
			if !reflect.DeepEqual(ref, tt.expectedRef) {
				t.Errorf("expected ref %v, got %v", tt.expectedRef, ref)
			}
		})
	}
}
