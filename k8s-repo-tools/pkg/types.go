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
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/google/go-github/v29/github"
)

const (
	// BranchMaster ...
	BranchMaster = "master"
	// PrefixBranch ...
	PrefixBranch = "release-"
	// PrefixDryRun ...
	PrefixDryRun = "DRY-RUN"
)

// Data is the main data structure of the application.
type Data struct {
	// From flags
	Dest         string
	Source       string
	MinVersion   string
	Token        string
	PrefixBranch string
	Output       string
	Timeout      time.Duration
	DryRun       bool
	Force        bool

	// Dynamic fields
	client    *github.Client
	Transport *Transport
}

// CreateContext can be used to create a new Go context with a timeout
// from data#timeout.
func (d *Data) CreateContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), d.Timeout)
}

// HTTPHandler has the same signature as the only function
// in the http.RoundTripper interface.
type HTTPHandler func(*http.Request) (*http.Response, error)

// Transport maps URLs to HTTPHandlers.
type Transport struct {
	sync.RWMutex

	handlers map[string]HTTPHandler
}

var _ http.RoundTripper = &Transport{}

// referenceSubset is a subset of the go-github Reference object.
type referenceSubset struct {
	Ref string `json:"ref"`
	SHA string `json:"sha"`
}
