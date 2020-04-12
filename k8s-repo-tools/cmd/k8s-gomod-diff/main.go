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
)

func main() {
	const (
		urlA = "https://raw.githubusercontent.com/kubernetes/kubeadm/master/kinder/go.mod"
		urlB = "https://raw.githubusercontent.com/kubernetes/kubernetes/master/go.mod"
	)
	if _, err := process(urlA, urlB); err != nil {
		fmt.Println("err", err)
	}
	fmt.Println("done")
}
