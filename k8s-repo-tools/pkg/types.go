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
	"fmt"
	"github.com/pkg/errors"
	"net/http"
	"strings"
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

// assetMap is a type that implements the flag.Value interface
// for supporting user input of 'name=path' for assets.
type assetMap map[string]string

func (m *assetMap) String() string {
	var list []string
	for k, v := range *m {
		list = append(list, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(list, ",")
}

func (m *assetMap) Set(value string) error {
	kv := strings.Split(value, "=")
	err := errors.Errorf("invalid asset format %q. Value must be formatted as 'name=path'", value)
	if len(kv) < 2 {
		return err
	}
	if len(kv[0]) == 0 || len(kv[1]) == 0 {
		return err
	}
	(*m)[kv[0]] = kv[1]
	return nil
}

// Data is the main data structure of the application.
type Data struct {
	// From flags
	Dest                 string
	Source               string
	MinVersion           string
	Token                string
	Branch               string
	PrefixBranch         string
	Output               string
	ReleaseTag           string
	ReleaseNotesToolPath string
	ReleaseNotesPath     string
	ReleaseAssets        assetMap
	BuildCommand         string
	Timeout              time.Duration
	DryRun               bool
	Force                bool

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
