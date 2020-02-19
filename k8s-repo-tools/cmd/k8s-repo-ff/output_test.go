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
	"testing"

	"github.com/google/go-github/v29/github"
)

func TestFormatOutput(t *testing.T) {
	tests := []struct {
		name        string
		out         *output
		expectedBuf []byte
	}{
		{
			name: "with error",
			out: &output{
				OutputError: github.String("test-error"),
				Reference:   &github.Reference{},
				Commit:      &github.RepositoryCommit{},
			},
			expectedBuf: []byte(`{"outputError":"test-error","reference":{"ref":null,"url":null,"object":null},"commit":{}}`),
		},
		{
			name: "without error",
			out: &output{
				Reference: &github.Reference{},
				Commit:    &github.RepositoryCommit{},
			},
			expectedBuf: []byte(`{"outputError":null,"reference":{"ref":null,"url":null,"object":null},"commit":{}}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf, err := formatOutput(tt.out, false)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !bytes.Equal(buf, tt.expectedBuf) {
				t.Errorf("expected output:\n%s\n, got:\n%s\n", tt.expectedBuf, buf)
			}
		})
	}
}
