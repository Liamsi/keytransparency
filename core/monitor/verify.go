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
// reconstruct clients can query.
package monitor

import (
	"bytes"
	"errors"
	"math/big"

	"github.com/golang/glog"

	"github.com/google/trillian/merkle"
	"github.com/google/trillian/storage"

	tcrypto "github.com/google/trillian/crypto"

	"github.com/google/keytransparency/core/mutator/entry"
	ktpb "github.com/google/keytransparency/core/proto/keytransparency_v1_types"
)

var (
	// ErrInvalidMutation occurs when verification failed because of an invalid
	// mutation.
	ErrInvalidMutation = errors.New("invalid mutation")
	// ErrNotMatchingMapRoot occurs when the reconstructed root differs from the one
	// we received from the server.
	ErrNotMatchingMapRoot = errors.New("recreated root does not match")
	// ErrInvalidMapSignature occurs if the map roots signature does not verify.
	ErrInvalidMapSignature = errors.New("invalid signature on map root")
	// ErrInvalidLogSignature occurs if the log roots signature does not verify.
	ErrInvalidLogSignature = errors.New("invalid signature on log root")
	// ErrInconsistentProofs occurs when the server returned different hashes
	// for the same inclusion proof node in the tree.
	ErrInconsistentProofs = errors.New("inconsistent inclusion proofs")
	// ErrInvalidConsistencyProof occurs when the log consistency proof does not
	// verify.
	ErrInvalidConsistencyProof = errors.New("invalid log consistency proof")
)

// verifyResponse verifies a response received by the GetMutations API.
// Additionally to the response it takes a complete list of mutations. The list
// of received mutations may differ from those included in the initial response
// because of the max. page size.
func (m *Monitor) verifyMutationsResponse(in *ktpb.GetMutationsResponse) []error {
	errList := make([]error, 0)
	// copy of singed map root
	smr := *in.GetSmr()
	// reset to the state before it was signed:
	smr.Signature = nil
	// verify signature on map root:
	if err := tcrypto.VerifyObject(m.mapPubKey, smr, in.GetSmr().GetSignature()); err != nil {
		glog.Infof("couldn't verify signature on map root: %v", err)
		errList = append(errList, ErrInvalidMapSignature)
	}

	logRoot := in.GetLogRoot()
	// Verify SignedLogRoot signature.
	hash := tcrypto.HashLogRoot(*logRoot)
	if err := tcrypto.Verify(m.logPubKey, hash, logRoot.GetSignature()); err != nil {
		glog.Infof("couldn't verify signature on log root: %v", err)
		errList = append(errList, ErrInvalidLogSignature)
	}

	if m.trusted != nil {
		// Verify consistency proof:
		err := m.logVerifier.VerifyConsistencyProof(
			m.trusted.TreeSize, logRoot.TreeSize,
			m.trusted.RootHash, logRoot.RootHash,
			in.GetLogConsistency())
		if err != nil {
			errList = append(errList, ErrInvalidConsistencyProof)
		}
	} else {
		// trust the first log root we see, don't verify anything yet
		m.trusted = in.GetLogRoot()
	}

	// m.logVerifier.VerifyInclusionProof()
	//
	// retrieve the old root hash from storage!
	monRes, err := m.store.Get(in.Epoch - 1)
	if err != nil {
		glog.Infof("Could not retrieve previous monitoring result: %v", err)
	}
	// we need the old root for verifying the inclusion of the old leafs in the
	// previous epoch. Storage always stores the mutations response independent
	// from if the checks succeeded or not.
	oldRoot := monRes.Response.GetSmr().GetRootHash()

	if err := m.verifyMutations(in.GetMutations(), oldRoot,
		in.GetSmr().GetRootHash(), in.GetSmr().GetMapId()); len(err) > 0 {
		errList = append(errList, err...)
	}

	return errList
}

func (m *Monitor) verifyMutations(muts []*ktpb.Mutation, oldRoot, expectedNewRoot []byte, mapID int64) []error {
	errList := make([]error, 0)
	mutator := entry.New()
	oldProofNodes := make(map[string][]byte)
	newLeaves := make([]merkle.HStar2LeafHash, 0, len(muts))

	for _, mut := range muts {
		oldLeaf, err := entry.FromLeafValue(mut.GetProof().GetLeaf().GetLeafValue())
		if err != nil {
			errList = append(errList, ErrInvalidMutation)
		}

		// verify that the provided leaf’s inclusion proof goes to epoch e-1:
		index := mut.GetProof().GetLeaf().GetIndex()
		leafHash := mut.GetProof().GetLeaf().GetLeafHash()
		if err := merkle.VerifyMapInclusionProof(mapID, index,
			leafHash, oldRoot, mut.GetProof().GetInclusion(), m.hasher); err != nil {
			glog.Infof("VerifyMapInclusionProof(%x): %v", index, err)
			errList = append(errList, err)
		}

		// compute the new leaf
		newLeaf, err := mutator.Mutate(oldLeaf, mut.GetUpdate())
		if err != nil {
			errList = append(errList, ErrInvalidMutation)
		}
		newLeafnID := storage.NewNodeIDFromPrefixSuffix(index, storage.Suffix{}, m.hasher.BitLen())
		newLeafHash := m.hasher.HashLeaf(mapID, index, newLeaf)
		newLeaves = append(newLeaves, merkle.HStar2LeafHash{
			Index:    newLeafnID.BigInt(),
			LeafHash: newLeafHash,
		})

		// store the proof hashes locally to recompute the tree below:
		sibIDs := newLeafnID.Siblings()
		// TODO(ismail) iterate over the sibIDs instead!
		for level, proof := range mut.GetProof().GetInclusion() {
			pID := sibIDs[level]
			if p, ok := oldProofNodes[pID.String()]; ok {
				// sanity check: for each mut overlapping proof nodes should be
				// equal:
				if !bytes.Equal(p, proof) {
					// TODO(ismail): remove this check and this error type as
					// soon as the server does not return interior proof nodes
					// multiple times
					//
					// this is really odd and should never happen
					errList = append(errList, ErrInconsistentProofs)
				}
			} else {
				oldProofNodes[pID.String()] = proof
			}
		}
	}
	if err := m.validateMapRoot(expectedNewRoot, mapID, newLeaves, oldProofNodes); err != nil {
		errList = append(errList, err)
	}

	return errList
}

func (m *Monitor) validateMapRoot(expectedRoot []byte, mapID int64, mutatedLeaves []merkle.HStar2LeafHash, oldProofNodes map[string][]byte) error {
	// compute the new root using local intermediate hashes from epoch e
	// (above proof hashes):
	hs2 := merkle.NewHStar2(mapID, m.hasher)
	newRoot, err := hs2.HStar2Nodes([]byte{}, m.hasher.Size(), mutatedLeaves,
		func(depth int, index *big.Int) ([]byte, error) {
			nID := storage.NewNodeIDFromBigInt(depth, index, m.hasher.BitLen())
			if p, ok := oldProofNodes[nID.String()]; ok {
				return p, nil
			}
			return nil, nil
		}, nil)

	if err != nil {
		glog.Errorf("hs2.HStar2Nodes(_): %v", err)
		// TODO(ismail): figure out what to return here
	}

	// verify rootHash
	if !bytes.Equal(newRoot, expectedRoot) {
		return ErrNotMatchingMapRoot
	}

	return nil
}
