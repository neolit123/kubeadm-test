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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"

	"github.com/google/go-github/v29/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

// SetHandler will map a HTTPHandler to a given URL.
func (t *Transport) SetHandler(url string, fn HTTPHandler) {
	t.Lock()
	defer t.Unlock()
	t.handlers[url] = fn
}

// RoundTrip satisfies the http.RoundTripper interface adding
// means for http.Request and http.Response interception.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	url := req.URL.String()
	var fn HTTPHandler

	t.RLock()
	// Find an endpoint handler.
	for k, v := range t.handlers {
		if strings.HasPrefix(url, k) {
			fn = v
			break
		}
	}
	t.RUnlock()

	if fn != nil {
		return fn(req)
	}
	return nil, errors.Errorf("missing handler for %q", url)
}

// NewTransport will create a new custom transport with HTTPHandlers.
func NewTransport() *Transport {
	return &Transport{handlers: map[string]HTTPHandler{}}
}

// NewClient will create a new client for connecting to the GitHub API.
// If a custom transport is passed this transport will manage all
// requests and responses that the go-github library otherwise does.
func NewClient(d *Data, t *Transport) {
	// create an ouath2 client with token authorization.
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: d.Token},
	)
	httpClient := oauth2.NewClient(context.Background(), ts)

	// Override the HTTP client transport.
	if t != nil {
		httpClient.Transport = t
		d.Transport = t
	}

	d.client = github.NewClient(httpClient)
}

// NewReferenceHandler creates a HTTPHandler function that manages a list of GitHub References.
func NewReferenceHandler(refs *[]*github.Reference, methodErrors map[string]bool) HTTPHandler {
	return func(req *http.Request) (*http.Response, error) {
		// Unescape '%2F' -> '/'
		url := strings.Replace(req.URL.String(), "%2F", "/", -1)

		// Return an early error if methodErrors matches the Method of this http.Request.
		if val, ok := methodErrors[req.Method]; ok && val {
			msg := fmt.Sprintf("simulating error for method %q to URL %q", req.Method, url)
			Logf(msg)
			return nil, errors.New(msg)
		}

		switch req.Method {
		case http.MethodGet: // Handle GET
			filteredRefs := []*github.Reference{}

			// Determine if this is a call to get tags or branches.
			getTags := strings.HasSuffix(url, "tags")
			getBranches := strings.HasSuffix(url, "heads")
			specificRef := strings.Split(url, "git/")[1]
			var foundSpecificRef bool
			for _, ref := range *refs {
				r := ref.GetRef()
				if r == specificRef {
					foundSpecificRef = true
					filteredRefs = append(filteredRefs, ref)
					break
				}
				if getTags && strings.HasPrefix(r, "refs/tags") {
					filteredRefs = append(filteredRefs, ref)
					continue
				}
				if getBranches && strings.HasPrefix(r, "refs/heads") {
					filteredRefs = append(filteredRefs, ref)
					continue
				}
			}

			// If no refs are found return a 404 without an error.
			// This is something the go-github also does as per the GitHub API.
			if len(filteredRefs) == 0 {
				Logf("simulating method %q with status %d from URL %q", req.Method, http.StatusNotFound, url)
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       ioutil.NopCloser(bytes.NewBuffer([]byte(""))),
					Header:     http.Header{},
				}, nil
			}

			// Simulate a GET by writing the list of filtered refs to the response body.
			Logf("simulating method %q with status %d from URL %q", req.Method, http.StatusOK, url)
			var buf []byte
			var err error
			if foundSpecificRef {
				buf, err = json.Marshal(filteredRefs[0])
			} else {
				buf, err = json.Marshal(filteredRefs)
			}
			if err != nil {
				return nil, err
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewBuffer(buf)),
				Header:     http.Header{},
			}, nil

		case http.MethodPost: // Handle POST
			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}

			// Use referenceSubset for intermediate storage. In go-github
			// this is done with a "createRefRequest" structure.
			r := referenceSubset{}
			if err := json.Unmarshal(body, &r); err != nil {
				return nil, err
			}
			newRef := &github.Reference{
				Ref: github.String(r.Ref),
				Object: &github.GitObject{
					SHA: github.String(r.SHA),
				},
			}

			// Simulate a POST by appending to the managed list of refs.
			Logf("simulating method %q with status %d to URL %q with; ref %q with sha %q",
				req.Method, http.StatusOK, url, r.Ref, r.SHA)
			*refs = append(*refs, newRef)

			return &http.Response{
				StatusCode: http.StatusOK, // Note: this status is not 200 for some operations.
				Body:       ioutil.NopCloser(bytes.NewBuffer([]byte{})),
				Header:     http.Header{},
			}, nil

		default:
			panic(fmt.Sprintf("unhandled HTTP method %q", req.Method))
		}
	}
}

// NewCompareHandler creates a HTTPHandler function that manages RepositoryCommit comparison between
// two GitHub branches.
func NewCompareHandler(commitsA, commitsB *[]*github.RepositoryCommit, methodErrors map[string]bool) HTTPHandler {
	return func(req *http.Request) (*http.Response, error) {

		// Return an early error if methodErrors matches the Method of this http.Request.
		if val, ok := methodErrors[req.Method]; ok && val {
			msg := fmt.Sprintf("simulating error for method %q to URL %q", req.Method, req.URL.String())
			Logf(msg)
			return nil, errors.New(msg)
		}

		switch req.Method {
		case http.MethodGet: // Handle GET

			var cmp *github.CommitsComparison
			if reflect.DeepEqual(commitsA, commitsB) {
				// Branches are identical.
				cmp = &github.CommitsComparison{
					Status: github.String("identical"),
				}
			} else {
				// Check if branch is ahead. No commit comparison, only length.
				var commits []github.RepositoryCommit
				if len(*commitsA) > len(*commitsB) {
					// Grab the extra commits from A.
					for i := len(*commitsB) - 1; i < len(*commitsA); i++ {
						commits = append(commits, *(*commitsA)[i])
					}
					cmp = &github.CommitsComparison{
						Status:  github.String("ahead"),
						Commits: commits,
					}
				} else {
					for i := len(*commitsA) - 1; i < len(*commitsB); i++ {
						commits = append(commits, *(*commitsB)[i])
					}
					cmp = &github.CommitsComparison{
						Status:  github.String("behind"),
						Commits: commits,
					}
				}
			}

			buf, err := json.Marshal(cmp)
			if err != nil {
				return nil, err
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewBuffer(buf)),
				Header:     http.Header{},
			}, nil

		default:
			panic(fmt.Sprintf("unhandled HTTP method %q", req.Method))
		}
	}
}

// NewMergeHandler creates a HTTPHandler function that manages a list of GitHub References
// using a map of HTTP method errors.
func NewMergeHandler(mergeRequest *github.RepositoryMergeRequest, status int, methodErrors map[string]bool) HTTPHandler {
	return func(req *http.Request) (*http.Response, error) {

		// Return an early error if methodErrors matches the Method of this http.Request.
		if val, ok := methodErrors[req.Method]; ok && val {
			msg := fmt.Sprintf("simulating error for method %q to URL %q", req.Method, req.URL.String())
			Errorf(msg)
			return nil, errors.New(msg)
		}

		switch req.Method {
		case http.MethodPost: // Handle POST

			commit := &github.RepositoryCommit{
				SHA: github.String("dry-run-sha"),
				Commit: &github.Commit{
					Message: github.String(mergeRequest.GetCommitMessage()),
				},
			}
			buf, err := json.Marshal(commit)
			if err != nil {
				return nil, err
			}

			if status == 0 {
				status = http.StatusCreated
			}

			return &http.Response{
				StatusCode: status,
				Body:       ioutil.NopCloser(bytes.NewBuffer(buf)),
				Header:     http.Header{},
			}, nil

		default:
			panic(fmt.Sprintf("unhandled HTTP method %q", req.Method))
		}
	}
}

// NewReleaseHandler creates a HTTPHandler function that manages a list of GitHub RepositoryReleases.
func NewReleaseHandler(releases *[]*github.RepositoryRelease, methodErrors map[string]bool) HTTPHandler {
	return func(req *http.Request) (*http.Response, error) {
		// Unescape '%2F' -> '/'
		url := strings.Replace(req.URL.String(), "%2F", "/", -1)

		// Return an early error if methodErrors matches the Method of this http.Request.
		if val, ok := methodErrors[req.Method]; ok && val {
			msg := fmt.Sprintf("simulating error for method %q to URL %q", req.Method, url)
			Logf(msg)
			return nil, errors.New(msg)
		}

		switch req.Method {
		case http.MethodGet: // Handle GET

			requestedTag := strings.Split(url, "tags/")[1]

			var release *github.RepositoryRelease
			for _, rel := range *releases {
				if requestedTag == rel.GetTagName() { // Only match releases by tag
					release = rel
					break
				}
			}

			if release == nil {
				Logf("simulating method %q with status %d from URL %q", req.Method, http.StatusNotFound, url)
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       ioutil.NopCloser(bytes.NewBuffer([]byte(`{"message":"","documentation_url":""}`))),
					Header:     http.Header{},
				}, nil
			}

			buf, err := json.Marshal(release)
			if err != nil {
				return nil, err
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewBuffer(buf)),
				Header:     http.Header{},
			}, nil

		case http.MethodPost: // Handle POST

			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}

			// Use referenceSubset for intermediate storage. In go-github
			// this is done with a "repositoryReleaseRequest" structure.
			r := releaseSubset{}
			if err := json.Unmarshal(body, &r); err != nil {
				return nil, err
			}
			newRelease := &github.RepositoryRelease{
				TagName:    github.String(r.TagName),
				Name:       github.String(r.Name),
				Body:       github.String(r.Body),
				Draft:      github.Bool(r.Draft),
				Prerelease: github.Bool(r.Prerelease),
			}

			// Simulate a POST by appending to the managed list of releases.
			Logf("simulating method %q with status %d to URL %q with; release\n%+v\n",
				req.Method, http.StatusOK, url, newRelease)

			buf, err := json.Marshal(newRelease)
			if err != nil {
				return nil, err
			}

			return &http.Response{
				StatusCode: http.StatusOK, // Note: this status is not 200 for some operations.
				Body:       ioutil.NopCloser(bytes.NewBuffer(buf)),
				Header:     http.Header{},
			}, nil

		default:
			panic(fmt.Sprintf("unhandled HTTP method %q", req.Method))
		}
	}
}
