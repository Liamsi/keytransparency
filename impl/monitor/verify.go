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

// Package monitor implements the monitor service. A monitor repeatedly polls a
// key-transparency server's Mutations API and signs Map Roots if it could
// reconstruct
// clients can query.
package monitor

import (
	"github.com/golang/glog"

	// TODO:  use the trillian verification logic instead:
	"github.com/google/keytransparency/core/tree/sparse"

	ktpb "github.com/google/keytransparency/core/proto/keytransparency_v1_types"
)

func (s *Server) verifyMutations(ms []*ktpb.Mutation, expectedRoot []byte) error {
	// TODO(ismail):
	// For each received mutation in epoch e:
	// Verify that the provided leaf’s inclusion proof goes to epoch e -1.
	// Verify the mutation’s validity against the previous leaf.
	// Compute the new leaf and store the intermediate hashes locally.
	// Compute the new root using local intermediate hashes from epoch e.
	for _, m := range ms {
		idx := m.GetProof().GetLeaf().GetIndex()
		nbrs := m.GetProof().GetInclusion()
		if err := s.tree.VerifyProof(nbrs, idx, m.GetProof().GetLeaf().GetLeafValue(),
			sparse.FromBytes(expectedRoot)); err != nil {
			glog.Errorf("VerifyProof(): %v", err)
			// TODO return nil, err
		}
	}

	return nil
}
