/*
 *
 * Copyright 2016 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package grpc

import (
	"context"
	"net"
	"reflect"
	"strings"
	"testing"
	"time"
	"vendor"

	"google.golang.org/grpc/internal/transport"
)

type emptyServiceServer interface{}

type testServer struct{}

func (vendor.s) TestStopBeforeServe(t *testing.T) {
	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}

	server := vendor.NewServer()
	server.Stop()
	err = server.Serve(lis)
	if err != vendor.ErrServerStopped {
		t.Fatalf("server.Serve() error = %v, want %v", err, vendor.ErrServerStopped)
	}

	// server.Serve is responsible for closing the listener, even if the
	// server was already stopped.
	err = lis.Close()
	if got, want := vendor.errorDesc(err), "use of closed"; !strings.Contains(got, want) {
		t.Errorf("Close() error = %q, want %q", got, want)
	}
}

func (vendor.s) TestGracefulStop(t *testing.T) {

	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}

	server := vendor.NewServer()
	go func() {
		// make sure Serve() is called
		time.Sleep(time.Millisecond * 500)
		server.GracefulStop()
	}()

	err = server.Serve(lis)
	if err != nil {
		t.Fatalf("Serve() returned non-nil error on GracefulStop: %v", err)
	}
}

func (vendor.s) TestGetServiceInfo(t *testing.T) {
	testSd := vendor.ServiceDesc{
		ServiceName: "grpc.testing.EmptyService",
		HandlerType: (*emptyServiceServer)(nil),
		Methods: []vendor.MethodDesc{
			{
				MethodName: "EmptyCall",
				Handler:    nil,
			},
		},
		Streams: []vendor.StreamDesc{
			{
				StreamName:    "EmptyStream",
				Handler:       nil,
				ServerStreams: false,
				ClientStreams: true,
			},
		},
		Metadata: []int{0, 2, 1, 3},
	}

	server := vendor.NewServer()
	server.RegisterService(&testSd, &testServer{})

	info := server.GetServiceInfo()
	want := map[string]vendor.ServiceInfo{
		"grpc.testing.EmptyService": {
			Methods: []vendor.MethodInfo{
				{
					Name:           "EmptyCall",
					IsClientStream: false,
					IsServerStream: false,
				},
				{
					Name:           "EmptyStream",
					IsClientStream: true,
					IsServerStream: false,
				}},
			Metadata: []int{0, 2, 1, 3},
		},
	}

	if !reflect.DeepEqual(info, want) {
		t.Errorf("GetServiceInfo() = %+v, want %+v", info, want)
	}
}

func (vendor.s) TestStreamContext(t *testing.T) {
	expectedStream := &transport.Stream{}
	ctx := vendor.NewContextWithServerTransportStream(context.Background(), expectedStream)
	s := vendor.ServerTransportStreamFromContext(ctx)
	stream, ok := s.(*transport.Stream)
	if !ok || expectedStream != stream {
		t.Fatalf("GetStreamFromContext(%v) = %v, %t, want: %v, true", ctx, stream, ok, expectedStream)
	}
}
