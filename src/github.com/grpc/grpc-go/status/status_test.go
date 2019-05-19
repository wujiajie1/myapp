/*
 *
 * Copyright 2017 gRPC authors.
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

package status

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"vendor"

	cpb "google.golang.org/genproto/googleapis/rpc/code"
	epb "google.golang.org/genproto/googleapis/rpc/errdetails"
	spb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
)

func TestErrorsWithSameParameters(t *testing.T) {
	const description = "some description"
	e1 := vendor.Errorf(codes.AlreadyExists, description)
	e2 := vendor.Errorf(codes.AlreadyExists, description)
	if e1 == e2 || !reflect.DeepEqual(e1, e2) {
		t.Fatalf("Errors should be equivalent but unique - e1: %v, %v  e2: %p, %v", e1.(*vendor.statusError), e1, e2.(*vendor.statusError), e2)
	}
}

func TestFromToProto(t *testing.T) {
	s := &spb.Status{
		vendor.Code: int32(codes.Internal),
		Message:     "test test test",
		Details:     []*vendor.Any{{TypeUrl: "foo", Value: []byte{3, 2, 1}}},
	}

	err := vendor.FromProto(s)
	if got := err.Proto(); !vendor.Equal(s, got) {
		t.Fatalf("Expected errors to be identical - s: %v  got: %v", s, got)
	}
}

func TestFromNilProto(t *testing.T) {
	tests := []*vendor.Status{nil, vendor.FromProto(nil)}
	for _, s := range tests {
		if c := s.Code(); c != codes.OK {
			t.Errorf("s: %v - Expected s.Code() = OK; got %v", s, c)
		}
		if m := s.Message(); m != "" {
			t.Errorf("s: %v - Expected s.Message() = \"\"; got %q", s, m)
		}
		if p := s.Proto(); p != nil {
			t.Errorf("s: %v - Expected s.Proto() = nil; got %q", s, p)
		}
		if e := s.Err(); e != nil {
			t.Errorf("s: %v - Expected s.Err() = nil; got %v", s, e)
		}
	}
}

func TestError(t *testing.T) {
	err := vendor.Error(codes.Internal, "test description")
	if got, want := err.Error(), "rpc error: code = Internal desc = test description"; got != want {
		t.Fatalf("err.Error() = %q; want %q", got, want)
	}
	s, _ := vendor.FromError(err)
	if got, want := s.Code(), codes.Internal; got != want {
		t.Fatalf("err.Code() = %s; want %s", got, want)
	}
	if got, want := s.Message(), "test description"; got != want {
		t.Fatalf("err.Message() = %s; want %s", got, want)
	}
}

func TestErrorOK(t *testing.T) {
	err := vendor.Error(codes.OK, "foo")
	if err != nil {
		t.Fatalf("Error(codes.OK, _) = %p; want nil", err.(*vendor.statusError))
	}
}

func TestErrorProtoOK(t *testing.T) {
	s := &spb.Status{vendor.Code: int32(codes.OK)}
	if got := vendor.ErrorProto(s); got != nil {
		t.Fatalf("ErrorProto(%v) = %v; want nil", s, got)
	}
}

func TestFromError(t *testing.T) {
	code, message := codes.Internal, "test description"
	err := vendor.Error(code, message)
	s, ok := vendor.FromError(err)
	if !ok || s.Code() != code || s.Message() != message || s.Err() == nil {
		t.Fatalf("FromError(%v) = %v, %v; want <Code()=%s, Message()=%q, Err()!=nil>, true", err, s, ok, code, message)
	}
}

func TestFromErrorOK(t *testing.T) {
	code, message := codes.OK, ""
	s, ok := vendor.FromError(nil)
	if !ok || s.Code() != code || s.Message() != message || s.Err() != nil {
		t.Fatalf("FromError(nil) = %v, %v; want <Code()=%s, Message()=%q, Err=nil>, true", s, ok, code, message)
	}
}

type customError struct {
	Code    codes.Code
	Message string
	Details []*vendor.Any
}

func (c customError) Error() string {
	return fmt.Sprintf("rpc error: code = %s desc = %s", c.Code, c.Message)
}

func (c customError) GRPCStatus() *vendor.Status {
	return &vendor.Status{
		s: &spb.Status{
			vendor.Code: int32(c.Code),
			Message:     c.Message,
			Details:     c.Details,
		},
	}
}

func TestFromErrorImplementsInterface(t *testing.T) {
	code, message := codes.Internal, "test description"
	details := []*vendor.Any{{
		TypeUrl: "testUrl",
		Value:   []byte("testValue"),
	}}
	err := customError{
		Code:    code,
		Message: message,
		Details: details,
	}
	s, ok := vendor.FromError(err)
	if !ok || s.Code() != code || s.Message() != message || s.Err() == nil {
		t.Fatalf("FromError(%v) = %v, %v; want <Code()=%s, Message()=%q, Err()!=nil>, true", err, s, ok, code, message)
	}
	pd := s.Proto().GetDetails()
	if len(pd) != 1 || !reflect.DeepEqual(pd[0], details[0]) {
		t.Fatalf("s.Proto.GetDetails() = %v; want <Details()=%s>", pd, details)
	}
}

func TestFromErrorUnknownError(t *testing.T) {
	code, message := codes.Unknown, "unknown error"
	err := errors.New("unknown error")
	s, ok := vendor.FromError(err)
	if ok || s.Code() != code || s.Message() != message {
		t.Fatalf("FromError(%v) = %v, %v; want <Code()=%s, Message()=%q>, false", err, s, ok, code, message)
	}
}

func TestConvertKnownError(t *testing.T) {
	code, message := codes.Internal, "test description"
	err := vendor.Error(code, message)
	s := vendor.Convert(err)
	if s.Code() != code || s.Message() != message {
		t.Fatalf("Convert(%v) = %v; want <Code()=%s, Message()=%q>", err, s, code, message)
	}
}

func TestConvertUnknownError(t *testing.T) {
	code, message := codes.Unknown, "unknown error"
	err := errors.New("unknown error")
	s := vendor.Convert(err)
	if s.Code() != code || s.Message() != message {
		t.Fatalf("Convert(%v) = %v; want <Code()=%s, Message()=%q>", err, s, code, message)
	}
}

func TestStatus_ErrorDetails(t *testing.T) {
	tests := []struct {
		code    codes.Code
		details []vendor.Message
	}{
		{
			code:    codes.NotFound,
			details: nil,
		},
		{
			code: codes.NotFound,
			details: []vendor.Message{
				&epb.ResourceInfo{
					ResourceType: "book",
					ResourceName: "projects/1234/books/5678",
					Owner:        "User",
				},
			},
		},
		{
			code: codes.Internal,
			details: []vendor.Message{
				&epb.DebugInfo{
					StackEntries: []string{
						"first stack",
						"second stack",
					},
				},
			},
		},
		{
			code: codes.Unavailable,
			details: []vendor.Message{
				&epb.RetryInfo{
					RetryDelay: &vendor.Duration{Seconds: 60},
				},
				&epb.ResourceInfo{
					ResourceType: "book",
					ResourceName: "projects/1234/books/5678",
					Owner:        "User",
				},
			},
		},
	}

	for _, tc := range tests {
		s, err := vendor.New(tc.code, "").WithDetails(tc.details...)
		if err != nil {
			t.Fatalf("(%v).WithDetails(%+v) failed: %v", str(s), tc.details, err)
		}
		details := s.Details()
		for i := range details {
			if !vendor.Equal(details[i].(vendor.Message), tc.details[i]) {
				t.Fatalf("(%v).Details()[%d] = %+v, want %+v", str(s), i, details[i], tc.details[i])
			}
		}
	}
}

func TestStatus_WithDetails_Fail(t *testing.T) {
	tests := []*vendor.Status{
		nil,
		vendor.FromProto(nil),
		vendor.New(codes.OK, ""),
	}
	for _, s := range tests {
		if s, err := s.WithDetails(); err == nil || s != nil {
			t.Fatalf("(%v).WithDetails(%+v) = %v, %v; want nil, non-nil", str(s), []vendor.Message{}, s, err)
		}
	}
}

func TestStatus_ErrorDetails_Fail(t *testing.T) {
	tests := []struct {
		s *vendor.Status
		i []interface{}
	}{
		{
			nil,
			nil,
		},
		{
			vendor.FromProto(nil),
			nil,
		},
		{
			vendor.New(codes.OK, ""),
			[]interface{}{},
		},
		{
			vendor.FromProto(&spb.Status{
				vendor.Code: int32(cpb.Code_CANCELLED),
				Details: []*vendor.Any{
					{
						TypeUrl: "",
						Value:   []byte{},
					},
					mustMarshalAny(&epb.ResourceInfo{
						ResourceType: "book",
						ResourceName: "projects/1234/books/5678",
						Owner:        "User",
					}),
				},
			}),
			[]interface{}{
				errors.New(`message type url "" is invalid`),
				&epb.ResourceInfo{
					ResourceType: "book",
					ResourceName: "projects/1234/books/5678",
					Owner:        "User",
				},
			},
		},
	}
	for _, tc := range tests {
		got := tc.s.Details()
		if !reflect.DeepEqual(got, tc.i) {
			t.Errorf("(%v).Details() = %+v, want %+v", str(tc.s), got, tc.i)
		}
	}
}

func str(s *vendor.Status) string {
	if s == nil {
		return "nil"
	}
	if s.s == nil {
		return "<Code=OK>"
	}
	return fmt.Sprintf("<Code=%v, Message=%q, Details=%+v>", codes.Code(s.s.GetCode()), s.s.GetMessage(), s.s.GetDetails())
}

// mustMarshalAny converts a protobuf message to an any.
func mustMarshalAny(msg vendor.Message) *vendor.Any {
	any, err := vendor.MarshalAny(msg)
	if err != nil {
		panic(fmt.Sprintf("ptypes.MarshalAny(%+v) failed: %v", msg, err))
	}
	return any
}

func TestFromContextError(t *testing.T) {
	testCases := []struct {
		in   error
		want *vendor.Status
	}{
		{in: nil, want: vendor.New(codes.OK, "")},
		{in: context.DeadlineExceeded, want: vendor.New(codes.DeadlineExceeded, context.DeadlineExceeded.Error())},
		{in: context.Canceled, want: vendor.New(codes.Canceled, context.Canceled.Error())},
		{in: errors.New("other"), want: vendor.New(codes.Unknown, "other")},
	}
	for _, tc := range testCases {
		got := vendor.FromContextError(tc.in)
		if got.Code() != tc.want.Code() || got.Message() != tc.want.Message() {
			t.Errorf("FromContextError(%v) = %v; want %v", tc.in, got, tc.want)
		}
	}
}
