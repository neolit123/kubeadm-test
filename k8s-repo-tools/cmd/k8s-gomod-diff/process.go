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
	"sort"
	"text/tabwriter"

	"golang.org/x/mod/modfile"
	"k8s.io/kubeadm/k8s-repo-tools/pkg"
)

func process(d *pkg.Data) (*output, error) {
	dataSource, err := pkg.ReadFromFileOrURL(d.Source, d.Timeout)
	if err != nil {
		return nil, err
	}
	dataDest, err := pkg.ReadFromFileOrURL(d.Dest, d.Timeout)
	if err != nil {
		return nil, err
	}
	return processBytes(dataSource, dataDest)
}

func processBytes(dataSource, dataDest []byte) (*output, error) {
	m := pathVersionTuple{}

	// Parse the source data.
	modFileSource, err := modfile.Parse("", []byte(dataSource), nil)
	if err != nil {
		return nil, err
	}

	for _, r := range modFileSource.Require {
		if r.Indirect {
			continue // Skip indirect dependencies.
		}
		m[r.Mod.Path] = &versionTuple{Source: r.Mod.Version}
	}

	// Parse the destination data.
	modFileDest, err := modfile.Parse("", []byte(dataDest), nil)
	if err != nil {
		return nil, err
	}

	for _, r := range modFileDest.Require {
		if r.Indirect {
			continue // Skip indirect dependencies.
		}
		if _, ok := m[r.Mod.Path]; !ok {
			continue
		}
		m[r.Mod.Path].Dest = r.Mod.Version
	}

	// Add Golang to the dependency list.
	m["Golang"] = &versionTuple{
		Source: modFileSource.Go.Version,
		Dest:   modFileDest.Go.Version,
	}

	// Format output.
	o := &output{
		Dependencies: m,
	}
	return o, nil
}

type output struct {
	Dependencies pathVersionTuple `json:"dependencies"`
}

type versionTuple struct {
	Source string `json:"source"`
	Dest   string `json:"dest"`
}

type pathVersionTuple map[string]*versionTuple

func formatOutput(w io.Writer, o *output, source, dest string) {
	var header = fmt.Sprintf("Comparing Go module files:\n  Source: %s\n  Destination: %s\n"+
		"The following dependency versions differ:", source, dest)
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

		// If the destination version is empty, dest does not have this dependency.
		if v.Dest == "" || v.Source == v.Dest {
			continue
		}
		if !hasHeader {
			fmt.Fprintln(w, header)
			tabW = tabwriter.NewWriter(w, 12, 0, 2, ' ', 0)
			fmt.Fprintln(tabW, "PATH\tSOURCE\tDEST")
			hasHeader = true
		}
		fmt.Fprintf(tabW, "%s\t%s\t%s\n", k, v.Source, v.Dest)
	}

	if tabW != nil {
		tabW.Flush()
	}
}
