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
	"io"
	"io/ioutil"
	"net/http"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/mod/modfile"
)

type versionTuple struct {
	Local  string `json:"local"`
	Remote string `json:"remote"`
}

type pathVersionTuple map[string]*versionTuple

type output struct {
	Dependencies pathVersionTuple `json:"dependencies"`
}

func formatOutput(w io.Writer, o *output, localURL, remoteURL string) {
	var header = fmt.Sprintf("Comparing Go module files, local: %s, remote: %s\n"+
		"The following dependency versions differ:", localURL, remoteURL)
	var hasHeader bool
	var tabW *tabwriter.Writer

	// Store the Dependencies map keys and sort them.
	keys := make([]string, len(o.Dependencies))
	var i int
	for k := range o.Dependencies {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	// Loop trough the keys.
	for _, k := range keys {
		v := o.Dependencies[k]

		// If the remote version is empty this means that remote is not using this dependency.
		if v.Remote == "" || v.Local == v.Remote {
			continue
		}
		if !hasHeader {
			fmt.Fprintln(w, header)
			tabW = tabwriter.NewWriter(w, 12, 0, 2, ' ', 0)
			fmt.Fprintln(tabW, "PATH\tLOCAL\tREMOTE")
			hasHeader = true
		}
		fmt.Fprintf(tabW, "%s\t%s\t%s\n", k, v.Local, v.Remote)
	}

	if tabW != nil {
		tabW.Flush()
	}
}

func process(urlLocal, urlRemote string) (*output, error) {
	dataLocal, err := ReadFromURL(urlLocal, -1)
	if err != nil {
		return nil, err
	}
	dataRemote, err := ReadFromURL(urlLocal, -1)
	if err != nil {
		return nil, err
	}
	return processBytes(dataLocal, dataRemote)
}

func processBytes(dataLocal, dataRemote []byte) (*output, error) {
	m := pathVersionTuple{}

	// Parse the local data.
	modFileLocal, err := modfile.Parse("", []byte(dataLocal), nil)
	if err != nil {
		return nil, err
	}

	for _, r := range modFileLocal.Require {
		if r.Indirect {
			continue // Skip indirect dependencies.
		}
		m[r.Mod.Path] = &versionTuple{Local: r.Mod.Version}
	}

	// Parse the remote data.
	modFileRemote, err := modfile.Parse("", []byte(dataRemote), nil)
	if err != nil {
		return nil, err
	}

	for _, r := range modFileRemote.Require {
		if r.Indirect {
			continue // Skip indirect dependencies.
		}
		if _, ok := m[r.Mod.Path]; !ok {
			continue
		}
		m[r.Mod.Path].Remote = r.Mod.Version
	}

	// Add Golang to the dependency list.
	m["Golang"] = &versionTuple{
		Local:  modFileLocal.Go.Version,
		Remote: modFileRemote.Go.Version,
	}

	// Format output.
	o := &output{
		Dependencies: m,
	}
	return o, nil
}

// ReadFromURL reads the contents of a URL and returns the data
// as bytes. "timeout" allows passing timeout to the HTTP request.
// If -1 is passed as "timeout" a default value is used.
func ReadFromURL(url string, timeout time.Duration) ([]byte, error) {
	if timeout < 0 {
		timeout = 10 * time.Second
	}
	client := http.Client{Timeout: timeout}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Wrapf(err, "received status %d", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read response body")
	}

	return data, nil
}
