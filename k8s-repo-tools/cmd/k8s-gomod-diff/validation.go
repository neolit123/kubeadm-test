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
	"k8s.io/kubeadm/k8s-repo-tools/pkg"
)

// validateData validates the user input.
func validateData(d *pkg.Data) error {
	pkg.Logf("validating user input...")

	// Validate empty options.
	for k, v := range map[string]*string{
		pkg.FlagSource: &d.Source,
		pkg.FlagDest:   &d.Dest,
	} {
		if err := pkg.ValidateEmptyOption(k, *v); err != nil {
			return err
		}
	}

	// Validate token.
	if len(d.Token) > 0 {
		if err := pkg.ValidateToken(pkg.FlagToken, d.Token); err != nil {
			return err
		}
	}

	// Validate target issue.
	if len(d.TargetIssue) > 0 {
		if err := pkg.ValidateTargetIssue(pkg.FlagTargetIssue, d.TargetIssue); err != nil {
			return err
		}
	}

	// Both token and target issue must be set.
	if (len(d.Token) > 0) != (len(d.TargetIssue) > 0) {
		return errors.Errorf("both --%s and --%s must be set", pkg.FlagToken, pkg.FlagTargetIssue)
	}

	return nil
}
