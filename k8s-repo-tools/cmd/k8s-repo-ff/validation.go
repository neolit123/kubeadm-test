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
	"k8s.io/kubeadm/k8s-repo-tools/pkg"
)

// validateData validates the user input.
func validateData(d *pkg.Data) error {
	pkg.Logf("validating user input...")

	// Validate empty options.
	for k, v := range map[string]*string{
		pkg.FlagDest:  &d.Dest,
		pkg.FlagToken: &d.Token,
	} {
		if err := pkg.ValidateEmptyOption(k, *v); err != nil {
			return err
		}
	}

	// Validate org/repo.
	if err := pkg.ValidateRepo(pkg.FlagDest, d.Dest); err != nil {
		return err
	}

	// Validate token.
	if err := pkg.ValidateToken(pkg.FlagToken, d.Token); err != nil {
		return err
	}

	return nil
}
