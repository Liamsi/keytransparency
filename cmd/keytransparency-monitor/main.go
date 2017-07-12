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

package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/google/keytransparency/core/mapserver"

	ctxn "github.com/google/keytransparency/core/transaction"
	"github.com/google/keytransparency/impl/sql/sequenced"
	"github.com/google/keytransparency/impl/sql/sqlhist"
	"github.com/google/trillian"

	"github.com/golang/glog"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"time"
)

var (
	addr     = flag.String("addr", ":8099", "The ip:port combination to listen on")
	vrfPath  = flag.String("vrf", "genfiles/vrf-pubkey.pem", "Path to VRF public key")
	keyFile  = flag.String("key", "testdata/server.key", "TLS private key file")
	certFile = flag.String("cert", "testdata/server.pem", "TLS cert file")

	pollPeriod = flag.Duration("poll-period", time.Second*5, "Maximum time between polling the key-server. Expected to be equal to the min-period of paramerter of the keyserver.")

	// TODO(ismail): are the IDs actually needed for verification operations?
	mapID = flag.Int64("map-id", 0, "Trillian map ID")
	logID = flag.Int64("log-id", 0, "Trillian Log ID")

	metricsAddr = flag.String("metrics-addr", ":8081", "The ip:port to publish metrics on")
)

func grpcGatewayMux(addr string) (*runtime.ServeMux, error) {
	// TODO(ismail): create and register service end-points here:
	//ctx := context.Background()
	//creds, err := credentials.NewClientTLSFromFile(*certFile, "")
	//if err != nil {
	//	return nil, err
	//}
	//dopts := []grpc.DialOption{grpc.WithTransportCredentials(creds)}
	//gwmux := runtime.NewServeMux()
	//if err := ktpb.RegisterKeyTransparencyServiceHandlerFromEndpoint(ctx, gwmux, addr, dopts); err != nil {
	//	return nil, err
	//}
	//if err := mpb.RegisterMutationServiceHandlerFromEndpoint(ctx, gwmux, addr, dopts); err != nil {
	//	return nil, err
	//}

	//return gwmux, nil
	return nil, errors.New("TODO: not implemented yet")
}

// grpcHandlerFunc returns an http.Handler that delegates to grpcServer on incoming gRPC
// connections or otherHandler otherwise. Copied from cockroachdb.
func grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This is a partial recreation of gRPC's internal checks.
		// https://github.com/grpc/grpc-go/blob/master/transport/handler_server.go#L62
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	})
}

func newReadonlyMapServer(ctx context.Context, mapID int64, sqldb *sql.DB, factory ctxn.Factory) (trillian.TrillianMapClient, error) {
	tree, err := sqlhist.New(ctx, mapID, factory)
	if err != nil {
		return nil, fmt.Errorf("Failed to create SQL history: %v", err)
	}
	sths, err := sequenced.New(sqldb, mapID)
	if err != nil {
		return nil, fmt.Errorf("sequenced.New(%v): %v", mapID, err)
	}
	return mapserver.NewReadonly(mapID, tree, factory, sths), nil
}

func main() {
	flag.Parse()

	// Open Resources.

	creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
	if err != nil {
		glog.Exitf("Failed to load server credentials %v", err)
	}

	// Create gRPC server.
	// svr := keyserver.New(*logID, tlog, *mapID, tmap, commitments,
	//	vrfPriv, mutator, auth, factory, mutations)
	grpcServer := grpc.NewServer(
		grpc.Creds(creds),
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
	)
	// TODO(ismail): register service:
	// ktpb.RegisterKeyTransparencyServiceServer(grpcServer, svr)
	reflection.Register(grpcServer)
	grpc_prometheus.Register(grpcServer)
	grpc_prometheus.EnableHandlingTimeHistogram()

	// Create HTTP handlers and gRPC gateway.
	gwmux, err := grpcGatewayMux(*addr)
	if err != nil {
		glog.Exitf("Failed setting up REST proxy: %v", err)
	}

	// Insert handlers for other http paths here.
	mux := http.NewServeMux()
	mux.Handle("/", gwmux)

	metricMux := http.NewServeMux()
	metricMux.Handle("/metrics", prometheus.Handler())
	go func() {
		log.Printf("Hosting metrics on %v", *metricsAddr)
		if err := http.ListenAndServe(*metricsAddr, metricMux); err != nil {
			log.Fatalf("ListenAndServeTLS(%v): %v", *metricsAddr, err)
		}
	}()
	// Serve HTTP2 server over TLS.
	glog.Infof("Listening on %v", *addr)
	if err := http.ListenAndServeTLS(*addr, *certFile, *keyFile,
		grpcHandlerFunc(grpcServer, mux)); err != nil {
		glog.Errorf("ListenAndServeTLS: %v", err)
	}
}
