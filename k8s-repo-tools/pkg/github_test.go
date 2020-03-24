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

package pkg

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/google/go-github/v29/github"
)

func TestGitHubGetCreateRelease(t *testing.T) {
	// Swap these two lines to enable debug logging.
	SetLogWriters(os.Stdout, os.Stderr)
	SetLogWriters(ioutil.Discard, ioutil.Discard)

	tests := []struct {
		name                 string
		data                 *Data
		refs                 []*github.Reference
		releases             []*github.RepositoryRelease
		releaseBody          string
		methodErrorsRefs     map[string]bool
		methodErrorsReleases map[string]bool
		skipDryRun           bool
		expectedRelease      *github.RepositoryRelease
		expectedError        bool
	}{
		{
			name: "valid: found release by matching tag",
			data: &Data{ReleaseTag: "v1.16.0"},
			refs: []*github.Reference{
				&github.Reference{Ref: github.String("refs/tags/v1.16.0"), Object: &github.GitObject{SHA: github.String("1234567890")}},
			},
			releases: []*github.RepositoryRelease{
				&github.RepositoryRelease{TagName: github.String("v1.16.0")},
			},
			expectedRelease: &github.RepositoryRelease{TagName: github.String("v1.16.0")},
		},
		{
			name: "valid: release is missing; create it from this tag",
			data: &Data{ReleaseTag: "v1.16.0"},
			refs: []*github.Reference{
				&github.Reference{Ref: github.String("refs/tags/v1.16.0"), Object: &github.GitObject{SHA: github.String("1234567890")}},
			},
			releaseBody: "foo",
			expectedRelease: &github.RepositoryRelease{
				TagName:    github.String("v1.16.0"),
				Name:       github.String("v1.16.0"),
				Body:       github.String("foo"),
				Draft:      github.Bool(false),
				Prerelease: github.Bool(false),
			},
		},
		{
			name: "valid: release is missing; create it from this tag (pre-release)",
			data: &Data{ReleaseTag: "v1.16.0-rc.1"},
			refs: []*github.Reference{
				&github.Reference{Ref: github.String("refs/tags/v1.16.0-rc.1"), Object: &github.GitObject{SHA: github.String("1234567890")}},
			},
			releaseBody: "foo",
			expectedRelease: &github.RepositoryRelease{
				TagName:    github.String("v1.16.0-rc.1"),
				Name:       github.String("v1.16.0-rc.1"),
				Body:       github.String("foo"),
				Draft:      github.Bool(false),
				Prerelease: github.Bool(true),
			},
		},
		{
			name: "invalid: fail creating release",
			data: &Data{ReleaseTag: "v1.16.0"},
			refs: []*github.Reference{
				&github.Reference{Ref: github.String("refs/tags/v1.16.0"), Object: &github.GitObject{SHA: github.String("1234567890")}},
			},
			methodErrorsReleases: map[string]bool{http.MethodPost: true},
			expectedError:        true,
			skipDryRun:           true,
		},
		{
			name: "invalid: tag not found in the list of refs",
			data: &Data{ReleaseTag: "v1.16.0"},
			refs: []*github.Reference{
				&github.Reference{Ref: github.String("refs/tags/v1.17.0"), Object: &github.GitObject{SHA: github.String("1234567890")}},
				&github.Reference{Ref: github.String("refs/tags/v1.15.0"), Object: &github.GitObject{SHA: github.String("1234567891")}},
			},
			expectedError: true,
		},
		{
			name:             "invalid: could not get the reference for this tag",
			data:             &Data{ReleaseTag: "v1.16.0"},
			methodErrorsRefs: map[string]bool{http.MethodGet: true},
			expectedError:    true,
		},
		{
			name:                 "invalid: could not get the release for this tag",
			data:                 &Data{ReleaseTag: "v1.16.0"},
			methodErrorsReleases: map[string]bool{http.MethodGet: true},
			expectedError:        true,
		},
	}

	// Make sure there are consistent results between dry-run and regular mode.
	for _, dryRunVal := range []bool{false, true} {
		for _, tt := range tests {
			t.Run(fmt.Sprintf("%s (dryRun=%v)", tt.name, dryRunVal), func(t *testing.T) {
				// Some operations like POST will always return non-error in dry-run mode.
				// Skip such tests.
				if tt.skipDryRun && dryRunVal {
					t.Skip()
				}

				// Override/hardcode some values.
				tt.data.Dest = "org/dest"
				tt.data.PrefixBranch = PrefixBranch
				tt.data.Force = true
				tt.data.DryRun = dryRunVal

				if tt.methodErrorsRefs == nil {
					tt.methodErrorsRefs = map[string]bool{}
				}
				if tt.methodErrorsReleases == nil {
					tt.methodErrorsReleases = map[string]bool{}
				}

				// Create fake client and setup endpoint handlers.
				NewClient(tt.data, NewTransport())
				const (
					testRefs     = "https://api.github.com/repos/org/dest/git/refs"
					testReleases = "https://api.github.com/repos/org/dest/releases"
				)
				handlerRefs := NewReferenceHandler(&tt.refs, tt.methodErrorsRefs)
				handlerReleases := NewReleaseHandler(&tt.releases, tt.methodErrorsReleases)
				tt.data.Transport.SetHandler(testRefs, handlerRefs)
				tt.data.Transport.SetHandler(testReleases, handlerReleases)

				rel, err := GitHubGetCreateRelease(tt.data, tt.data.Dest, tt.data.ReleaseTag, tt.releaseBody, dryRunVal)
				if (err != nil) != tt.expectedError {
					t.Errorf("expected error %v, got %v, error: %v", tt.expectedError, err != nil, err)
				}
				if err != nil {
					return
				}

				if !reflect.DeepEqual(tt.expectedRelease, rel) {
					t.Errorf("expected release:\n%+v\ngot:\n%+v\n", tt.expectedRelease, rel)
				}
			})
		}
	}
}

func TestGitHubUploadReleaseAssets(t *testing.T) {
	// Swap these two lines to enable debug logging.
	SetLogWriters(os.Stdout, os.Stderr)
	SetLogWriters(ioutil.Discard, ioutil.Discard)

	tests := []struct {
		name           string
		release        *github.RepositoryRelease
		am             assetMap
		skipDryRun     bool
		methodErrors   map[string]bool
		expectedAssets []*github.ReleaseAsset
		expectedError  bool
	}{
		{
			name: "valid: release has no assets; upload new ones",
			am: assetMap{
				"foo1": "bar1",
				"foo2": "bar2",
			},
			release: &github.RepositoryRelease{
				ID: github.Int64(1),
			},
			expectedAssets: []*github.ReleaseAsset{
				&github.ReleaseAsset{Name: github.String("foo1")},
				&github.ReleaseAsset{Name: github.String("foo2")},
			},
		},
		{
			name: "valid: release has existing assets; upload new ones",
			am: assetMap{
				"foo1": "bar1",
				"foo2": "bar2",
			},
			release: &github.RepositoryRelease{
				ID: github.Int64(1),
				Assets: []github.ReleaseAsset{
					github.ReleaseAsset{Name: github.String("z1")},
					github.ReleaseAsset{Name: github.String("z2")},
				},
			},
			expectedAssets: []*github.ReleaseAsset{
				&github.ReleaseAsset{Name: github.String("foo1")},
				&github.ReleaseAsset{Name: github.String("foo2")},
				&github.ReleaseAsset{Name: github.String("z1")},
				&github.ReleaseAsset{Name: github.String("z2")},
			},
		},
		{
			name: "valid: release has overlap between existing and new assets",
			am: assetMap{
				"foo1": "bar1",
				"foo2": "bar2",
				"z1":   "baz1",
			},
			release: &github.RepositoryRelease{
				ID: github.Int64(1),
				Assets: []github.ReleaseAsset{
					github.ReleaseAsset{Name: github.String("foo1")},
				},
			},
			expectedAssets: []*github.ReleaseAsset{
				&github.ReleaseAsset{Name: github.String("foo1")},
				&github.ReleaseAsset{Name: github.String("foo2")},
				&github.ReleaseAsset{Name: github.String("z1")},
			},
		},
		{
			name: "invalid: simulate error uploading an asset",
			am: assetMap{
				"foo1": "bar1",
			},
			release: &github.RepositoryRelease{
				ID: github.Int64(1),
			},
			methodErrors:  map[string]bool{http.MethodPost: true},
			skipDryRun:    true,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		am, dir, err := createTempTestFiles(tt.am)
		if err != nil {
			if len(dir) > 0 {
				os.RemoveAll(dir)
			}
			t.Fatalf("error creating temporary files: %v", err)
		}
		// Make sure there are consistent results between dry-run and regular mode.
		for _, dryRunVal := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s (dryRun=%v)", tt.name, dryRunVal), func(t *testing.T) {
				// Some operations like POST will always return non-error in dry-run mode.
				// Skip such tests.
				if tt.skipDryRun && dryRunVal {
					t.Skip()
				}

				// Override/hardcode some values.
				data := &Data{}
				data.Dest = "org/dest"
				data.PrefixBranch = PrefixBranch
				data.Force = true
				data.DryRun = dryRunVal

				if tt.methodErrors == nil {
					tt.methodErrors = map[string]bool{}
				}

				// Clone the release object, so that multiple runs on (dryRun=false/true) would store state.
				release := *(tt.release)

				// Create fake client and setup endpoint handlers.
				NewClient(data, NewTransport())
				handlerReleaseAssets := NewReleaseAssetsHandler(&release, tt.methodErrors)
				data.Transport.SetHandler(
					fmt.Sprintf("https://api.github.com/repos/org/dest/releases/%d/assets", release.GetID()),
					handlerReleaseAssets,
				)
				data.Transport.SetHandler(
					fmt.Sprintf("https://uploads.github.com/repos/org/dest/releases/%d/assets", release.GetID()),
					handlerReleaseAssets,
				)

				assets, err := GitHubUploadReleaseAssets(data, data.Dest, &release, am, dryRunVal)
				if (err != nil) != tt.expectedError {
					t.Errorf("expected error %v, got %v, error: %v", tt.expectedError, err != nil, err)
				}

				sort.Slice(tt.expectedAssets, func(i, j int) bool {
					return tt.expectedAssets[i].GetName() < tt.expectedAssets[j].GetName()
				})
				sort.Slice(assets, func(i, j int) bool {
					return assets[i].GetName() < assets[j].GetName()
				})
				if !reflect.DeepEqual(tt.expectedAssets, assets) {
					t.Errorf("expected assets:\n%+v\ngot:\n%+v\n", tt.expectedAssets, assets)
				}
			})
		}
		os.RemoveAll(dir)
	}
}

// createTempTestFiles takes an assetMap and creates a new assetMap that points
// to real files from a temporary directory.
func createTempTestFiles(am assetMap) (assetMap, string, error) {
	if len(am) == 0 {
		return am, "", nil
	}

	dir, err := ioutil.TempDir("", "assets")
	if err != nil {
		return nil, "", err
	}

	newAssetMap := assetMap{}
	for k, v := range am {
		updatedPath := filepath.Join(dir, v)
		err := ioutil.WriteFile(updatedPath, []byte("placeholder"), 0644)
		if err != nil {
			return nil, dir, err
		}
		newAssetMap[k] = updatedPath
	}
	return newAssetMap, dir, nil
}
