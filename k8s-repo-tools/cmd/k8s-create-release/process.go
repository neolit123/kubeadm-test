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
	var outputPath, bodyStr string

	// If a direct release notes path is given, do not generate them.
	if len(d.ReleaseNotesPath) != 0 {
		outputPath = d.ReleaseNotesPath
	} else if len(d.ReleaseNotesToolPath) != 0 {
		pkg.Logf("using %q as the release notes tool path", d.ReleaseNotesToolPath)

		// TODO:
		// - find the ref of the user tag (A)
		// - list all repo tag refs (B)
		// - use A and B for FindReleaseNotesSinceRef()
		// - use the SHAs for runGenerateReleaseNotes()

		endSHA := "TODO"
		startSHA := "TODO"

		outputPath, err = runGenerateReleaseNotes(d, startSHA, endSHA)
		if err != nil {
			return err
		}
	}

	// Only load the release notes if the output path was defined.
	if len(outputPath) != 0 {
		pkg.Logf("reading the release notes from %q", outputPath)
		body, err := ioutil.ReadFile(outputPath)
		if err != nil {
			return err
		}
		bodyStr = string(body)
	}

	// Create a release for this tag.
	release, err := pkg.GitHubGetCreateRelease(d, d.Dest, d.ReleaseTag, bodyStr)
	if err != nil {
		return err
	}

	// Build the release if a build command was provided.
	if len(d.BuildCommand) != 0 {
		if d.DryRun {
			pkg.Logf("%s: would execute the build command:\n%s", pkg.PrefixDryRun, d.BuildCommand)
		} else {
			if err := runCommand(d.BuildCommand); err != nil {
				return err
			}
		}
	} else {
		pkg.Warningf("empty --%s value; skipping build", pkg.FlagBuildCommand)
	}

	// Upload the release assets if such are provided.
	if len(d.ReleaseAssets) > 0 {
		if d.DryRun {
			pkg.Logf("%s: would upload the following artifacts to the release:\n%v", pkg.PrefixDryRun, d.ReleaseAssets)
		} else {
			if err = pkg.GitHubUploadReleaseAssets(d, d.Dest, release, d.ReleaseAssets); err != nil {
				return err
			}
		}
	} else {
		pkg.Warningf("no release assets were provided using --%s; skipping upload", pkg.FlagReleaseAsset)
	}
	return nil
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
	if err := runCommand(d.ReleaseNotesToolPath, args...); err != nil {
		return "", err
	}
	return outputPath, nil
}

func runCommand(cmdPath string, args ...string) error {
	pkg.Logf("running command: %s", cmdPath)
	pkg.Logf("using arguments: %v", args)
	cmd := exec.Command(cmdPath, args...)
	stdout, stderr := pkg.GetLogWriters()
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Env = os.Environ()
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
