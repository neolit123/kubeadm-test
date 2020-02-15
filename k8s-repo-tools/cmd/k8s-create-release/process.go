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
	// "net/http"

	// "github.com/google/go-github/v29/github"
	// "github.com/pkg/errors"
	// "k8s.io/apimachinery/pkg/util/version"
	"k8s.io/kubeadm/k8s-repo-tools/pkg"
)

// process is responsible for all operations that the application performs.
func process(d *pkg.Data, body []byte) error {
	// Only format the body if not empty.
	var bodyStr string
	if len(body) != 0 {
		bodyStr = formatReleaseBody(body)
	}

	// Create a release for this tag.
	release, err := pkg.GitHubGetCreateRelease(d, d.Dest, d.ReleaseTag, bodyStr)
	if err != nil {
		return err
	}

	// TODO: build release

	files := []string{}
	err = pkg.GitHubUploadReleaseAssets(d, d.Dest, release, files)
	if err != nil {
		return err
	}

	return nil
}

func formatReleaseBody(body []byte) string {
	return fmt.Sprintf("<details>\n<summary>Click to see changelog</summary>\n\n%q\n\n</details>", body)
}
