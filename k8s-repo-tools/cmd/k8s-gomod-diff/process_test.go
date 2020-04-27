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
	"bytes"
	"encoding/json"
	"testing"
)

func TestProcessBytes(t *testing.T) {
	tests := []struct {
		name               string
		dataSource         []byte
		dataDest           []byte
		expectedOutputJSON string
		expectedError      bool
	}{
		{
			name: "valid: dest has newer components",
			dataSource: []byte(`
			module k8s.io/kubeadm
			go 1.12
			require (
				k8s.io/klog v0.8.0
				sigs.k8s.io/yaml v1.0.0
			)
			`),
			dataDest: []byte(`
			module k8s.io/kubernetes
			go 1.13
			require (
				k8s.io/klog v0.9.0
				sigs.k8s.io/yaml v1.1.0
				k8s.io/api v1.0.0
			)
			`),
			expectedOutputJSON: `{"dependencies":{"Golang":{"source":"1.12","dest":"1.13"},"k8s.io/klog":{"source":"v0.8.0","dest":"v0.9.0"},"sigs.k8s.io/yaml":{"source":"v1.0.0","dest":"v1.1.0"}}}`,
		},
		{
			name: "valid: dependency versions match",
			dataSource: []byte(`
			module k8s.io/kubeadm
			go 1.13
			require (
				k8s.io/klog v0.8.0
			)
			`),
			dataDest: []byte(`
			module k8s.io/kubernetes
			go 1.13
			require (
				k8s.io/klog v0.8.0
			)
			`),
			expectedOutputJSON: `{"dependencies":{"Golang":{"source":"1.13","dest":"1.13"},"k8s.io/klog":{"source":"v0.8.0","dest":"v0.8.0"}}}`,
		},
		{
			name: "valid: no matching dependencies in dest",
			dataSource: []byte(`
			module k8s.io/kubeadm
			go 1.13
			require (
				k8s.io/klog v0.8.0
			)
			`),
			dataDest: []byte(`
			module k8s.io/kubernetes
			go 1.13
			require (
				k8s.io/api v1.0.0
			)
			`),
			expectedOutputJSON: `{"dependencies":{"Golang":{"source":"1.13","dest":"1.13"},"k8s.io/klog":{"source":"v0.8.0","dest":""}}}`,
		},
		{
			name: "valid: indirect dependencies are skipped",
			dataSource: []byte(`
			module k8s.io/kubeadm
			go 1.13
			require (
				k8s.io/klog v0.8.0 // indirect
			)
			`),
			dataDest: []byte(`
			module k8s.io/kubernetes
			go 1.13
			require (
				k8s.io/klog v0.8.0 // indirect
			)
			`),
			expectedOutputJSON: `{"dependencies":{"Golang":{"source":"1.13","dest":"1.13"}}}`,
		},
		{
			name:          "invalid: error parsing input",
			dataSource:    []byte(`foo`),
			dataDest:      []byte(`bar`),
			expectedError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			output, err := processBytes(tc.dataSource, tc.dataDest)
			if (err != nil) != tc.expectedError {
				t.Errorf("expected error: %v, got: %v, error: %v", tc.expectedError, err != nil, err)
			}
			if err != nil {
				return
			}
			outputJSON, err := json.Marshal(output)
			if err != nil {
				t.Fatalf("could not marshal output: %v", err)
			}
			outputJSONString := string(outputJSON)
			if tc.expectedOutputJSON != outputJSONString {
				t.Errorf("expected output:\n%s\ngot:\n%s\n", tc.expectedOutputJSON, outputJSON)
			}
		})
	}
}

func TestFormatOutput(t *testing.T) {
	tests := []struct {
		name           string
		output         *output
		expectedOutput string
	}{
		{
			name: "valid: output with go version and two dependencies",
			output: &output{
				Dependencies: pathVersionTuple{
					"Golang":      &versionTuple{Source: "1.12", Dest: "1.13"},
					"k8s.io/klog": &versionTuple{Source: "v1.0.0", Dest: "v1.1.0"},
					"github.com/someorg/someverylongnamegoeshere": &versionTuple{Source: "v1.0.0", Dest: "v1.1.0"},
				},
			},
			expectedOutput: `Comparing Go module files:
  Source: https://foo
  Destination: https://bar
The following dependency versions differ:
PATH                                         SOURCE      DEST
Golang                                       1.12        1.13
github.com/someorg/someverylongnamegoeshere  v1.0.0      v1.1.0
k8s.io/klog                                  v1.0.0      v1.1.0
`,
		},
		{
			name: "valid: only one dependency differs",
			output: &output{
				Dependencies: pathVersionTuple{
					"Golang":      &versionTuple{Source: "1.12", Dest: "1.12"},
					"k8s.io/klog": &versionTuple{Source: "v1.0.0", Dest: "v1.1.0"},
				},
			},
			expectedOutput: `Comparing Go module files:
  Source: https://foo
  Destination: https://bar
The following dependency versions differ:
PATH         SOURCE      DEST
k8s.io/klog  v1.0.0      v1.1.0
`,
		},
		{
			name: "valid: only go version differs",
			output: &output{
				Dependencies: pathVersionTuple{
					"Golang": &versionTuple{Source: "1.12", Dest: "1.13"},
				},
			},
			expectedOutput: `Comparing Go module files:
  Source: https://foo
  Destination: https://bar
The following dependency versions differ:
PATH        SOURCE      DEST
Golang      1.12        1.13
`,
		},
		{
			name: "valid: no differences",
			output: &output{
				Dependencies: pathVersionTuple{},
			},
			expectedOutput: ``,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var b bytes.Buffer
			formatOutput(&b, tc.output, "https://foo", "https://bar")
			bStr := b.String()
			if bStr != tc.expectedOutput {
				t.Errorf("expected output:\n%s\ngot:\n%s\n", tc.expectedOutput, bStr)
			}
		})
	}
}
