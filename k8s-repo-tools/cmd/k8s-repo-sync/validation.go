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
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/version"
	"k8s.io/kubeadm/k8s-repo-tools/pkg"
)

// validateData validates the user input.
func validateData(d *pkg.Data) error {
	pkg.Logf("validating user input...")

	// Validate empty options.
	for k, v := range map[string]*string{
		pkg.FlagDest:       &d.Dest,
		pkg.FlagSource:     &d.Source,
		pkg.FlagMinVersion: &d.MinVersion,
		pkg.FlagToken:      &d.Token,
	} {
		if err := pkg.ValidateEmptyOption(k, *v); err != nil {
			return err
		}
	}

	// Validate org/repo options.
	for k, v := range map[string]*string{
		pkg.FlagDest:   &d.Dest,
		pkg.FlagSource: &d.Source,
	} {
		if err := pkg.ValidateRepo(k, *v); err != nil {
			return err
		}
	}

	// Validate version.
	if _, err := version.ParseSemantic(d.MinVersion); err != nil {
		return errors.Wrapf(err, "the option %q must be a valid semantic version", pkg.FlagMinVersion)
	}

	// Validate token.
	if err := pkg.ValidateToken(pkg.FlagToken, d.Token); err != nil {
		return err
	}

	return nil
}
