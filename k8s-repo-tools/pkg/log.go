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
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/v29/github"
	"github.com/pkg/errors"
)

var (
	logMutex = &sync.Mutex{}
	stdout   io.Writer
	stderr   io.Writer

	lineSeparator = strings.Repeat("*", 79)
)

// SetLogWriters ...
func SetLogWriters(out, err io.Writer) {
	logMutex.Lock()
	defer logMutex.Unlock()
	stdout = out
	stderr = err
}

// GetLogWriters ...
func GetLogWriters() (io.Writer, io.Writer) {
	logMutex.Lock()
	defer logMutex.Unlock()
	return stdout, stderr
}

func getLogPrefix(t string, f string) string {
	const layout = "15:04:05.000000"
	_, fn, line, _ := runtime.Caller(2)
	fn = filepath.Base(fn)
	return fmt.Sprintf("%s %s %s:%d %s\n", t, time.Now().Format(layout), fn, line, f)
}

// Logf ...
func Logf(f string, a ...interface{}) {
	fmt.Fprintf(stdout, getLogPrefix("I", f), a...)
}

// Warningf ...
func Warningf(f string, a ...interface{}) {
	fmt.Fprintf(stderr, getLogPrefix("W", f), a...)
}

// Errorf ...
func Errorf(f string, a ...interface{}) {
	fmt.Fprintf(stderr, getLogPrefix("E", f), a...)
}

// PrintErrorAndExit ...
func PrintErrorAndExit(err error) {
	Errorf("%+v", errors.WithStack(err))
	os.Exit(1)
}

// PrintSeparator ...
func PrintSeparator() {
	fmt.Fprintln(stderr, lineSeparator)
}

// LogRefList prints a simplified Reference object that only has "ref"
// and "sha" fields.
func LogRefList(msg, repo string, refs []*github.Reference) {
	str := make([]string, len(refs))
	for i, ref := range refs {
		r := referenceSubset{Ref: ref.GetRef(), SHA: ref.GetObject().GetSHA()}
		buf, _ := json.Marshal(&r)
		str[i] = string(buf)
	}
	Logf(msg+" for %s: %v", repo, str)
}
