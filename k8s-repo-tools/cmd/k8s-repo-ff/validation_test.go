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
	"testing"

	"k8s.io/kubeadm/k8s-repo-tools/pkg"
)

func TestValidateData(t *testing.T) {
	const validToken = "282ef40c7d38cbfafe7d6ebe91cdfbbcbe5d71ab"
	pkg.SetLogWriters(ioutil.Discard, ioutil.Discard)

	tests := []struct {
		name          string
		data          *pkg.Data
		expectedError bool
	}{
		{
			name: "valid: all fields are valid",
			data: &pkg.Data{
				Token: validToken,
				Dest:  "org/dest",
			},
		},
		{
			name: "valid: valid fields and version prefixed token",
			data: &pkg.Data{
				Token: "v1." + validToken,
				Dest:  "org/dest",
			},
		},
		{
			name: "invalid: empty string arguments",
			data: &pkg.Data{
				Token: "",
				Dest:  "",
			},
			expectedError: true,
		},
		{
			name: "invalid: short token hash",
			data: &pkg.Data{
				Token: "282ef40c7d38cbfa",
				Dest:  "org/dest",
			},
			expectedError: true,
		},
		{
			name: "invalid: token hash is not in hex",
			data: &pkg.Data{
				Token: "zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz",
				Dest:  "org/dest",
			},
			expectedError: true,
		},
		{
			name: "invalid: repositories are not formatted correctly",
			data: &pkg.Data{
				Token: validToken,
				Dest:  "bar/",
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateData(tt.data); (err != nil) != tt.expectedError {
				t.Errorf("expected error %v, got %v, error: %v", tt.expectedError, err != nil, err)
			}
		})
	}
}
