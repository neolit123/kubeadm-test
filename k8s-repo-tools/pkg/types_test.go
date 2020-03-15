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
	"io/ioutil"
	"testing"
)

func TestAssetMap(t *testing.T) {
	SetLogWriters(ioutil.Discard, ioutil.Discard)

	tests := []struct {
		name           string
		input          []string
		expectedOutput string
		expectedError  bool
	}{
		{
			name:           "valid: a list of valid key=value pairs",
			input:          []string{"key1=value", "key2=value"},
			expectedOutput: "key1=value,key2=value",
		},
		{
			name:           "valid: same keys override existing keys",
			input:          []string{"key1=value", "key1=value"},
			expectedOutput: "key1=value",
		},
		{
			name:           "invalid: badly separated key value pair",
			input:          []string{"key1=value", "key2-value"},
			expectedOutput: "key1=value",
			expectedError:  true,
		},
		{
			name:           "invalid: empty key or value in pairs",
			input:          []string{"=value", "key="},
			expectedOutput: "",
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			am := assetMap{}
			var foundErr error
			for _, i := range tt.input {
				if err := am.Set(i); err != nil {
					foundErr = err
				}
			}
			if (foundErr != nil) != tt.expectedError {
				t.Errorf("expected error %v, got %v, error: %v", tt.expectedError, foundErr != nil, foundErr)
			}
			out := am.String()
			if out != tt.expectedOutput {
				t.Errorf("expected output %q, got %q", tt.expectedOutput, out)
			}
		})
	}
}
