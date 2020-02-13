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
	refs := []*github.Reference{
		&github.Reference{Ref: github.String("/refs/tags/v1.16.0"), Object: &github.GitObject{SHA: github.String("123456780")}},
		&github.Reference{Ref: github.String("/refs/tags/v1.17.0"), Object: &github.GitObject{SHA: github.String("123456780")}},
		&github.Reference{Ref: github.String("/refs/heads/release-1.16"), Object: &github.GitObject{SHA: github.String("123456780")}},
		&github.Reference{Ref: github.String("/refs/heads/release-1.17"), Object: &github.GitObject{SHA: github.String("123456780")}},
	}

	out, err := formatOutput(refs, false)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expectedOut := []byte(`[{"ref":"/refs/tags/v1.16.0","url":null,"object":{"type":null,"sha":"123456780","url":null}},` +
		`{"ref":"/refs/tags/v1.17.0","url":null,"object":{"type":null,"sha":"123456780","url":null}},` +
		`{"ref":"/refs/heads/release-1.16","url":null,"object":{"type":null,"sha":"123456780","url":null}},` +
		`{"ref":"/refs/heads/release-1.17","url":null,"object":{"type":null,"sha":"123456780","url":null}}]`)

	if !bytes.Equal(out, expectedOut) {
		t.Errorf("expected output:\n%s\n, got:\n%s\n", expectedOut, out)
	}
}
