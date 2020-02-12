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
	"encoding/json"
	"io/ioutil"

	"github.com/google/go-github/v29/github"
	"k8s.io/kubeadm/k8s-repo-tools/pkg"
)

// formatRefOutput marshals a list of Reference objects.
func formatRefOutput(refs []*github.Reference) ([]byte, error) {
	buf, err := json.Marshal(&refs)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

// writeRefOutputToFile writes the list of Reference objects to the
// given filePath.
func writeRefOutputToFile(filePath string, refs []*github.Reference) error {
	buf, err := formatRefOutput(refs)
	if err != nil {
		return err
	}

	pkg.Logf("writing the resulted tags and branches to the file %q", filePath)
	if err := ioutil.WriteFile(filePath, buf, 0600); err != nil {
		return err
	}
	return nil
}
