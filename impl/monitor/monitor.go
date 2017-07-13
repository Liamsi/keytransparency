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

// Package monitor implements the monitor service. A monitor repeatedly polls a
// key-transparency server's Mutations API and signs Map Roots if it could
// reconstruct
// clients can query.
package monitor

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	ktpb "github.com/google/keytransparency/core/proto/keytransparency_v1_types"
	mspb "github.com/google/keytransparency/impl/proto/monitor_v1_service"
)

// Server holds internal state for the monitor server.
type Server struct {
	// TODO
}

// New creates a new instance of the monitor server.
func New( /*TODO*/ ) *Server {
	return &Server{}
}

// GetSignedMapRoot returns the latest reconstructed using the Mutations API and
// validated signed map root.
func (s *Server) GetSignedMapRoot(ctx context.Context, in *ktpb.GetMonitoringRequest) (*ktpb.GetMonitoringResponse, error) {
	return nil, grpc.Errorf(codes.Unimplemented, "GetSignedMapRoot is unimplemented")
}

// GetSignedMapRootStream is a streaming API similar to GetSignedMapRoot.
func (s *Server) GetSignedMapRootStream(in *ktpb.GetMonitoringRequest, stream mspb.MonitorService_GetSignedMapRootStreamServer) error {
	return grpc.Errorf(codes.Unimplemented, "GetSignedMapRootStream is unimplemented")
}

func (s *Server) GetSignedMapRootByRevision(ctx context.Context, in *ktpb.GetMonitoringRequest) (*ktpb.GetMonitoringResponse, error){
	return nil, grpc.Errorf(codes.Unimplemented, "GetSignedMapRoot is unimplemented")
}
