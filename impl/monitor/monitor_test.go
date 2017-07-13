package monitor

import (
	"google.golang.org/grpc/codes"
	"testing"
	"google.golang.org/grpc"
	"golang.org/x/net/context"
)

func TestGetSignedMapRoot(t *testing.T) {
	srv := New()
	_, err := srv.GetSignedMapRoot(context.TODO(), nil)
	if got, want := grpc.Code(err), codes.Unimplemented; got != want {
		t.Errorf("GetSignedMapRootStream(_, _): %v, want %v", got, want)
	}
}

func TestGetSignedMapRootStream(t *testing.T) {
	srv := New()
	err := srv.GetSignedMapRootStream(nil, nil)
	if got, want := grpc.Code(err), codes.Unimplemented; got != want {
		t.Errorf("GetSignedMapRootStream(_, _): %v, want %v", got, want)
	}
}

