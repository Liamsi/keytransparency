// Copyright 2017 Google Inc. All Rights Reserved.
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

package monitor

import (
	"crypto"

	"github.com/google/trillian"
	"github.com/google/trillian/merkle"
	"github.com/google/trillian/merkle/hashers"
)

type Monitor struct {
	hasher      hashers.MapHasher
	logPubKey   crypto.PublicKey
	mapPubKey   crypto.PublicKey
	logVerifier merkle.LogVerifier
	trusted     trillian.SignedLogRoot
}

func New(hasher hashers.MapHasher, mapPubKey, logPubKey crypto.PublicKey, logVerifier merkle.LogVerifier) *Monitor {
	return &Monitor{
		hasher:      hasher,
		logVerifier: logVerifier,
		logPubKey:   logPubKey,
		mapPubKey:   mapPubKey,
	}
}