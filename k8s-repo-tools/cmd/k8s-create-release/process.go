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
	"io/ioutil"
	"k8s.io/apimachinery/pkg/util/version"
	"os"
	"os/exec"
	"strings"

	"k8s.io/kubeadm/k8s-repo-tools/pkg"
)

// process is responsible for all operations that the application performs.
func process(d *pkg.Data) error {

	// Handle release notes.
	var err error
	var outputPath string

	// If a direct release notes path is given read them from the file.
	// If a path to a release notes tools is given use it to generate the release notes.
	if len(d.ReleaseNotesPath) != 0 {
		outputPath = d.ReleaseNotesPath
	} else if len(d.ReleaseNotesToolPath) != 0 {

		// Get the start and end SHA to use for the release notes tool.
		startSHA, endSHA, err := getReleaseNotesToolSHAs(d)
		if err != nil {
			return err
		}

		// If a branch does not exist for this tag use "master"
		v := version.MustParseSemantic(d.ReleaseTag)
		branch := fmt.Sprintf("%s%d.%d", d.PrefixBranch, v.Major(), v.Minor())
		if _, err := pkg.GitHubGetRef(d, d.Dest, "refs/heads/"+branch); err != nil {
			branch = "master"
		}

		// Run the release notes tool.
		outputPath, err = runGenerateReleaseNotes(d, branch, startSHA, endSHA)
		if len(outputPath) != 0 {
			defer os.Remove(outputPath)
		}
		if err != nil {
			return err
		}
	}

	// Only load the release notes if the output path was defined.
	var bodyStr string
	if len(outputPath) != 0 {
		bodyStr, err = readReleaseNotes(outputPath, d.DryRun)
		if err != nil {
			return err
		}
	}

	var promptMessage string
	var yes bool

	// Skip prompt.
	if d.Force {
		goto createRelease
	}

	// Prompt the user about creating a release.
	promptMessage = fmt.Sprintf("Do you want to create a release for tag %q if it does not exist already?",
		d.ReleaseTag)
	if yes, err = pkg.ShowPrompt(promptMessage); err != nil {
		return err
	} else if yes {
		goto createRelease
	}
	return nil

createRelease:

	// Create a release for this tag.
	// Note: bodyStr can be empty if the release notes process was skipped.
	release, err := pkg.GitHubGetCreateRelease(d, d.Dest, d.ReleaseTag, bodyStr, d.DryRun)
	if err != nil {
		return err
	}

	// Build the release if a build command was provided.
	if len(d.BuildCommand) != 0 {
		buildCommmandSplit := strings.Split(d.BuildCommand, " ")
		var args []string
		if len(buildCommmandSplit) > 1 {
			args = buildCommmandSplit[1:]
		}
		if err := runCommand(buildCommmandSplit[0], []string{}, d.DryRun, args...); err != nil {
			return err
		}
	} else {
		pkg.Warningf("empty --%s value; skipping build", pkg.FlagBuildCommand)
	}

	// Skip prompt.
	if d.Force {
		goto uploadAssets
	}

	// Prompt the user about uploading the assets.
	promptMessage = fmt.Sprintf("Do you want to upload the given assets to release %q?",
		d.ReleaseTag)
	if yes, err = pkg.ShowPrompt(promptMessage); err != nil {
		return err
	} else if yes {
		goto uploadAssets
	}
	return nil

uploadAssets:

	// Upload the release assets if such are provided.
	if len(d.ReleaseAssets) > 0 {
		if _, err = pkg.GitHubUploadReleaseAssets(d, d.Dest, release, d.ReleaseAssets, d.DryRun); err != nil {
			return err
		}
	} else {
		pkg.Warningf("no release assets were provided using --%s; skipping upload", pkg.FlagReleaseAsset)
	}
	return nil
}

func getReleaseNotesToolSHAs(d *pkg.Data) (string, string, error) {
	pkg.Logf("finding which commits to use for the release notes tool")

	// Find the reference of the user provided release tag.
	endRef, err := pkg.GitHubGetRef(d, d.Dest, "refs/tags/"+d.ReleaseTag)
	if err != nil {
		return "", "", err
	}

	// Fetch all tag references for the destination repository.
	refs, err := pkg.GitHubGetTags(d, d.Dest)
	if err != nil {
		return "", "", err
	}

	// Find which reference to use for the end tag.
	startRef, err := pkg.FindReleaseNotesSinceRef(endRef, refs)
	if err != nil {
		return "", "", err
	}

	startSHA := startRef.GetObject().GetSHA()
	endSHA := endRef.GetObject().GetSHA()
	pkg.Logf("found start SHA %s and end SHA %s", startSHA, endSHA)
	return startSHA, endSHA, nil
}

func runGenerateReleaseNotes(d *pkg.Data, branch, startSHA, endSHA string) (string, error) {
	pkg.Logf("will now run the release notes tool at %q", d.ReleaseNotesToolPath)

	// Allocate a temporary file path.
	file, err := ioutil.TempFile("", "release-notes")
	if err != nil {
		return "", err
	}
	outputPath := file.Name()
	file.Close()
	pkg.Logf("using output path %q", outputPath)

	// Prepare arguments.
	ownerRepo := strings.Split(d.Dest, "/")
	args := []string{
		"--start-sha=" + startSHA,
		"--end-sha=" + endSHA,
		"--output=" + outputPath,
		"--github-org=" + ownerRepo[0],
		"--github-repo=" + ownerRepo[1],
		"--required-author=" + `""`,
		"--branch=" + branch,
		"--toc",
	}
	if err := runCommand(d.ReleaseNotesToolPath, []string{"GITHUB_TOKEN=" + d.Token}, d.DryRun, args...); err != nil {
		return "", err
	}
	return outputPath, nil
}

func runCommand(cmdPath string, environment []string, dryRun bool, args ...string) error {
	if dryRun {
		pkg.Logf("%s: would run command: %s", pkg.PrefixDryRun, cmdPath)
		pkg.Logf("%s: using arguments: %v", pkg.PrefixDryRun, args)
		return nil
	}
	pkg.Logf("running command: %s", cmdPath)
	pkg.Logf("using arguments: %v", args)

	cmd := exec.Command(cmdPath, args...)
	stdout, stderr := pkg.GetLogWriters()
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Env = append(os.Environ(), environment...)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func readReleaseNotes(outputPath string, dryRun bool) (string, error) {
	if dryRun {
		pkg.Logf("%s: would read the release notes from %q", pkg.PrefixDryRun, outputPath)
		return "dry-run-release-notes", nil
	}
	pkg.Logf("reading the release notes from %q", outputPath)
	body, err := ioutil.ReadFile(outputPath)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
