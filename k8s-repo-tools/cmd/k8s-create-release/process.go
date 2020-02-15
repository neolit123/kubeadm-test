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
	"os"
	"path/filepath"
	"strings"
	// "net/http"
	"io/ioutil"
	"math/rand"
	"os/exec"
	"time"

	// "github.com/google/go-github/v29/github"
	// "github.com/pkg/errors"
	// "k8s.io/apimachinery/pkg/util/version"
	"k8s.io/kubeadm/k8s-repo-tools/pkg"
)

// process is responsible for all operations that the application performs.
func process(d *pkg.Data) error {

	// Handle release notes.
	var err error
	var bodyStr string
	if len(d.ReleaseNotesToolPath) != 0 {

		// Determine endSHA from the user supplied tag.
		// TODO
		endSHA := "TODO"

		// Determine startSHA:
		// TODO
		startSHA := "TODO"

		bodyStr, err = runGenerateReleaseNotes(d, startSHA, endSHA)
		if err != nil {
			return err
		}
	}

	// Create a release for this tag.
	release, err := pkg.GitHubGetCreateRelease(d, d.Dest, d.ReleaseTag, bodyStr)
	if err != nil {
		return err
	}

	// TODO: build release

	files := []string{}
	// TODO: user-provided assets, verify that the build created them.

	err = pkg.GitHubUploadReleaseAssets(d, d.Dest, release, files)
	if err != nil {
		return err
	}

	return nil
}

func formatReleaseBody(body []byte) string {
	return fmt.Sprintf("<details>\n<summary>Click to see changelog</summary>\n\n%q\n\n</details>", body)
}

func runGenerateReleaseNotes(d *pkg.Data, startSHA, endSHA string) (string, error) {
	pkg.Logf("will now run the release notes tool at %q", d.ReleaseNotesToolPath)

	// Allocate a temporary file path.
	rand.Seed(time.Now().UTC().UnixNano())
	var outputPath string
	for {
		outputPath = filepath.Join(os.TempDir(),
			fmt.Sprintf("release-notes-%d", rand.Intn(10000)))
		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			break
		}
	}
	pkg.Logf("using output path %q", outputPath)
	defer os.Remove(outputPath)

	// Prepare arguments.
	ownerRepo := strings.Split(d.Dest, "/")
	args := []string{
		"--start-sha=" + startSHA,
		"--end-sha=" + endSHA,
		"--output=" + outputPath,
		"--github-org=" + ownerRepo[0],
		"--github-repo=" + ownerRepo[1],
		"--toc",
	}
	pkg.Logf("using arguments: %v", args)

	// Prepare the command and run it.
	cmd := exec.Command(d.ReleaseNotesToolPath, args...)
	stdout, stderr := pkg.GetLogWriters()
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Env = os.Environ()
	if err := cmd.Run(); err != nil {
		return "", err
	}

	// Read the results.
	pkg.Logf("reading the release notes from  %q", outputPath)
	str, err := ioutil.ReadFile(outputPath)
	if err != nil {
		return "", err
	}
	return string(str), nil
}
