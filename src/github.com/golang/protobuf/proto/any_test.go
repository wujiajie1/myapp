// Go support for Protocol Buffers - Google's data interchange format
//
// Copyright 2016 The Go Authors.  All rights reserved.
// https://github.com/golang/protobuf
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//     * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//     * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//     * Neither the name of Google Inc. nor the names of its
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package proto_test

import (
	"strings"
	"testing"
	"vendor"
)

var (
	expandedMarshaler        = vendor.TextMarshaler{ExpandAny: true}
	expandedCompactMarshaler = vendor.TextMarshaler{Compact: true, ExpandAny: true}
)

// anyEqual reports whether two messages which may be google.protobuf.Any or may
// contain google.protobuf.Any fields are equal. We can't use proto.Equal for
// comparison, because semantically equivalent messages may be marshaled to
// binary in different tag order. Instead, trust that TextMarshaler with
// ExpandAny option works and compare the text marshaling results.
func anyEqual(got, want vendor.Message) bool {
	// if messages are proto.Equal, no need to marshal.
	if vendor.Equal(got, want) {
		return true
	}
	g := expandedMarshaler.Text(got)
	w := expandedMarshaler.Text(want)
	return g == w
}

type golden struct {
	m    vendor.Message
	t, c string
}

var goldenMessages = makeGolden()

func makeGolden() []golden {
	nested := &vendor.Nested{Bunny: "Monty"}
	nb, err := vendor.Marshal(nested)
	if err != nil {
		panic(err)
	}
	m1 := &vendor.Message{
		Name:        "David",
		ResultCount: 47,
		Anything:    &vendor.Any{TypeUrl: "type.googleapis.com/" + vendor.MessageName(nested), Value: nb},
	}
	m2 := &vendor.Message{
		Name:        "David",
		ResultCount: 47,
		Anything:    &vendor.Any{TypeUrl: "http://[::1]/type.googleapis.com/" + vendor.MessageName(nested), Value: nb},
	}
	m3 := &vendor.Message{
		Name:        "David",
		ResultCount: 47,
		Anything:    &vendor.Any{TypeUrl: `type.googleapis.com/"/` + vendor.MessageName(nested), Value: nb},
	}
	m4 := &vendor.Message{
		Name:        "David",
		ResultCount: 47,
		Anything:    &vendor.Any{TypeUrl: "type.googleapis.com/a/path/" + vendor.MessageName(nested), Value: nb},
	}
	m5 := &vendor.Any{TypeUrl: "type.googleapis.com/" + vendor.MessageName(nested), Value: nb}

	any1 := &vendor.MyMessage{Count: vendor.Int32(47), Name: vendor.String("David")}
	vendor.SetExtension(any1, vendor.E_Ext_More, &vendor.Ext{Data: vendor.String("foo")})
	vendor.SetExtension(any1, vendor.E_Ext_Text, vendor.String("bar"))
	any1b, err := vendor.Marshal(any1)
	if err != nil {
		panic(err)
	}
	any2 := &vendor.MyMessage{Count: vendor.Int32(42), Bikeshed: vendor.MyMessage_GREEN.Enum(), RepBytes: [][]byte{[]byte("roboto")}}
	vendor.SetExtension(any2, vendor.E_Ext_More, &vendor.Ext{Data: vendor.String("baz")})
	any2b, err := vendor.Marshal(any2)
	if err != nil {
		panic(err)
	}
	m6 := &vendor.Message{
		Name:        "David",
		ResultCount: 47,
		Anything:    &vendor.Any{TypeUrl: "type.googleapis.com/" + vendor.MessageName(any1), Value: any1b},
		ManyThings: []*vendor.Any{
			&vendor.Any{TypeUrl: "type.googleapis.com/" + vendor.MessageName(any2), Value: any2b},
			&vendor.Any{TypeUrl: "type.googleapis.com/" + vendor.MessageName(any1), Value: any1b},
		},
	}

	const (
		m1Golden = `
name: "David"
result_count: 47
anything: <
  [type.googleapis.com/proto3_proto.Nested]: <
    bunny: "Monty"
  >
>
`
		m2Golden = `
name: "David"
result_count: 47
anything: <
  ["http://[::1]/type.googleapis.com/proto3_proto.Nested"]: <
    bunny: "Monty"
  >
>
`
		m3Golden = `
name: "David"
result_count: 47
anything: <
  ["type.googleapis.com/\"/proto3_proto.Nested"]: <
    bunny: "Monty"
  >
>
`
		m4Golden = `
name: "David"
result_count: 47
anything: <
  [type.googleapis.com/a/path/proto3_proto.Nested]: <
    bunny: "Monty"
  >
>
`
		m5Golden = `
[type.googleapis.com/proto3_proto.Nested]: <
  bunny: "Monty"
>
`
		m6Golden = `
name: "David"
result_count: 47
anything: <
  [type.googleapis.com/test_proto.MyMessage]: <
    count: 47
    name: "David"
    [test_proto.Ext.more]: <
      data: "foo"
    >
    [test_proto.Ext.text]: "bar"
  >
>
many_things: <
  [type.googleapis.com/test_proto.MyMessage]: <
    count: 42
    bikeshed: GREEN
    rep_bytes: "roboto"
    [test_proto.Ext.more]: <
      data: "baz"
    >
  >
>
many_things: <
  [type.googleapis.com/test_proto.MyMessage]: <
    count: 47
    name: "David"
    [test_proto.Ext.more]: <
      data: "foo"
    >
    [test_proto.Ext.text]: "bar"
  >
>
`
	)
	return []golden{
		{m1, strings.TrimSpace(m1Golden) + "\n", strings.TrimSpace(vendor.compact(m1Golden)) + " "},
		{m2, strings.TrimSpace(m2Golden) + "\n", strings.TrimSpace(vendor.compact(m2Golden)) + " "},
		{m3, strings.TrimSpace(m3Golden) + "\n", strings.TrimSpace(vendor.compact(m3Golden)) + " "},
		{m4, strings.TrimSpace(m4Golden) + "\n", strings.TrimSpace(vendor.compact(m4Golden)) + " "},
		{m5, strings.TrimSpace(m5Golden) + "\n", strings.TrimSpace(vendor.compact(m5Golden)) + " "},
		{m6, strings.TrimSpace(m6Golden) + "\n", strings.TrimSpace(vendor.compact(m6Golden)) + " "},
	}
}

func TestMarshalGolden(t *testing.T) {
	for _, tt := range goldenMessages {
		if got, want := expandedMarshaler.Text(tt.m), tt.t; got != want {
			t.Errorf("message %v: got:\n%s\nwant:\n%s", tt.m, got, want)
		}
		if got, want := expandedCompactMarshaler.Text(tt.m), tt.c; got != want {
			t.Errorf("message %v: got:\n`%s`\nwant:\n`%s`", tt.m, got, want)
		}
	}
}

func TestUnmarshalGolden(t *testing.T) {
	for _, tt := range goldenMessages {
		want := tt.m
		got := vendor.Clone(tt.m)
		got.Reset()
		if err := vendor.UnmarshalText(tt.t, got); err != nil {
			t.Errorf("failed to unmarshal\n%s\nerror: %v", tt.t, err)
		}
		if !anyEqual(got, want) {
			t.Errorf("message:\n%s\ngot:\n%s\nwant:\n%s", tt.t, got, want)
		}
		got.Reset()
		if err := vendor.UnmarshalText(tt.c, got); err != nil {
			t.Errorf("failed to unmarshal\n%s\nerror: %v", tt.c, err)
		}
		if !anyEqual(got, want) {
			t.Errorf("message:\n%s\ngot:\n%s\nwant:\n%s", tt.c, got, want)
		}
	}
}

func TestMarshalUnknownAny(t *testing.T) {
	m := &vendor.Message{
		Anything: &vendor.Any{
			TypeUrl: "foo",
			Value:   []byte("bar"),
		},
	}
	want := `anything: <
  type_url: "foo"
  value: "bar"
>
`
	got := expandedMarshaler.Text(m)
	if got != want {
		t.Errorf("got\n`%s`\nwant\n`%s`", got, want)
	}
}

func TestAmbiguousAny(t *testing.T) {
	pb := &vendor.Any{}
	err := vendor.UnmarshalText(`
	type_url: "ttt/proto3_proto.Nested"
	value: "\n\x05Monty"
	`, pb)
	t.Logf("result: %v (error: %v)", expandedMarshaler.Text(pb), err)
	if err != nil {
		t.Errorf("failed to parse ambiguous Any message: %v", err)
	}
}

func TestUnmarshalOverwriteAny(t *testing.T) {
	pb := &vendor.Any{}
	err := vendor.UnmarshalText(`
  [type.googleapis.com/a/path/proto3_proto.Nested]: <
    bunny: "Monty"
  >
  [type.googleapis.com/a/path/proto3_proto.Nested]: <
    bunny: "Rabbit of Caerbannog"
  >
	`, pb)
	want := `line 7: Any message unpacked multiple times, or "type_url" already set`
	if err.Error() != want {
		t.Errorf("incorrect error.\nHave: %v\nWant: %v", err.Error(), want)
	}
}

func TestUnmarshalAnyMixAndMatch(t *testing.T) {
	pb := &vendor.Any{}
	err := vendor.UnmarshalText(`
	value: "\n\x05Monty"
  [type.googleapis.com/a/path/proto3_proto.Nested]: <
    bunny: "Rabbit of Caerbannog"
  >
	`, pb)
	want := `line 5: Any message unpacked multiple times, or "value" already set`
	if err.Error() != want {
		t.Errorf("incorrect error.\nHave: %v\nWant: %v", err.Error(), want)
	}
}
