// Copyright 2016 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package multiWriter

import (
	"errors"
	"fmt"
	"io"
)

// MultiWriter contains a list of io.Writer objects. Its Write method tries to write to all of them, and aggregates
// the errors (if any of the write call fails).
type MultiWriter interface {
	AddWriter(w io.Writer)
	Write(p []byte) (n int, err error)
}

// New returns an implementation of the MultiWriter interface
func New(w io.Writer) MultiWriter {
	return &multiIoWriter{[]io.Writer{w}}
}

type multiIoWriter struct {
	writers []io.Writer
}

func (m *multiIoWriter) Write(p []byte) (n int, err error) {
	if len(m.writers) == 0 {
		return 0, fmt.Errorf("Tried to use a MultiIoWriter which does not contain any writers")
	}
	multiError := ""
	minBytesWritten := len(p)
	for _, w := range m.writers {
		n, err = w.Write(p)
		if err != nil {
			multiError = multiError + fmt.Sprintf("%v bytes written to %v: %v", n, w, err)
		}
		minBytesWritten = min(n, minBytesWritten)
	}

	return minBytesWritten, errors.New(multiError)
}

func (m *multiIoWriter) AddWriter(w io.Writer) {
	m.writers = append(m.writers, w)
}

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}
