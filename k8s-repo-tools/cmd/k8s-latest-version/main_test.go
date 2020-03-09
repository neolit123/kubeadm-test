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
	"io/ioutil"
	"strings"
	"testing"

	"k8s.io/kubeadm/k8s-repo-tools/pkg"
)

func TestProcess(t *testing.T) {
	pkg.SetLogWriters(ioutil.Discard, ioutil.Discard)

	tests := []struct {
		name           string
		data           *pkg.Data
		input          []string
		expectedOutput string
		expectedError  bool
	}{
		{
			name: "valid: find the latest SemVer tag",
			input: []string{
				"v1.16.2-alpha.1",
				"v1.15.0",
				"v1.16.2-rc.1",
				"v1.14.3",
			},
			data: &pkg.Data{
				PrefixBranch: pkg.PrefixBranch,
			},
			expectedOutput: "v1.16.2-rc.1",
		},
		{
			name: "valid: find the latest SemVer tag matching a branch",
			input: []string{
				"v1.16.2",
				"v1.14.2",
				"v1.15.0",
				"v1.14.3",
			},
			data: &pkg.Data{
				Branch:       pkg.PrefixBranch + "1.14",
				PrefixBranch: pkg.PrefixBranch,
			},
			expectedOutput: "v1.14.3",
		},
		{
			name: "valid: should ignore tags that are not SemVer",
			input: []string{
				"v1.16.2",
				"foo",
				"v1.14.3",
				"bar",
			},
			data: &pkg.Data{
				PrefixBranch: pkg.PrefixBranch,
			},
			expectedOutput: "v1.16.2",
		},
		{
			name: "valid: should tollerate missing PATCH and 'v' in tags",
			input: []string{
				"1.15",
				"1.16",
				"1.14",
			},
			data: &pkg.Data{
				PrefixBranch: pkg.PrefixBranch,
			},
			expectedOutput: "1.16",
		},
		{
			name:  "invalid: cannot parse SemVer from branch",
			input: []string{},
			data: &pkg.Data{
				Branch:       pkg.PrefixBranch + "foo",
				PrefixBranch: pkg.PrefixBranch,
			},
			expectedError: true,
		},
		{
			name:  "invalid: branch does not contain the branch prefix",
			input: []string{},
			data: &pkg.Data{
				Branch:       "foo-bar",
				PrefixBranch: pkg.PrefixBranch,
			},
			expectedError: true,
		},
		{
			name: "invalid: could not find SemVer tags in the input",
			input: []string{
				"foo",
				"bar",
			},
			data: &pkg.Data{
				PrefixBranch: pkg.PrefixBranch,
			},
			expectedError: true,
		},
		{
			name: "invalid: could not find SemVer tags in the input for a given branch",
			input: []string{
				"foo",
				"bar",
			},
			data: &pkg.Data{
				Branch:       pkg.PrefixBranch + "1.14",
				PrefixBranch: pkg.PrefixBranch,
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := strings.NewReader(strings.Join(tt.input, "\n"))

			output, err := process(input, tt.data)
			if (err != nil) != tt.expectedError {
				t.Fatalf("expected error %v, got %v, error: %v", tt.expectedError, err != nil, err)
			}

			if output != tt.expectedOutput {
				t.Errorf("expected output %q, got %q", tt.expectedOutput, output)
			}
		})
	}
}
