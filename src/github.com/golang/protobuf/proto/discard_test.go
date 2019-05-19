// Go support for Protocol Buffers - Google's data interchange format
//
// Copyright 2017 The Go Authors.  All rights reserved.
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
	"testing"
	"vendor"
)

func TestDiscardUnknown(t *testing.T) {
	tests := []struct {
		desc     string
		in, want vendor.Message
	}{{
		desc: "Nil",
		in:   nil, want: nil, // Should not panic
	}, {
		desc: "NilPtr",
		in:   (*vendor.Message)(nil), want: (*vendor.Message)(nil), // Should not panic
	}, {
		desc: "Nested",
		in: &vendor.Message{
			Name:             "Aaron",
			Nested:           &vendor.Nested{Cute: true, XXX_unrecognized: []byte("blah")},
			XXX_unrecognized: []byte("blah"),
		},
		want: &vendor.Message{
			Name:   "Aaron",
			Nested: &vendor.Nested{Cute: true},
		},
	}, {
		desc: "Slice",
		in: &vendor.Message{
			Name: "Aaron",
			Children: []*vendor.Message{
				{Name: "Sarah", XXX_unrecognized: []byte("blah")},
				{Name: "Abraham", XXX_unrecognized: []byte("blah")},
			},
			XXX_unrecognized: []byte("blah"),
		},
		want: &vendor.Message{
			Name: "Aaron",
			Children: []*vendor.Message{
				{Name: "Sarah"},
				{Name: "Abraham"},
			},
		},
	}, {
		desc: "OneOf",
		in: &vendor.Communique{
			Union: &vendor.Communique_Msg{&vendor.Strings{
				StringField:      vendor.String("123"),
				XXX_unrecognized: []byte("blah"),
			}},
			XXX_unrecognized: []byte("blah"),
		},
		want: &vendor.Communique{
			Union: &vendor.Communique_Msg{&vendor.Strings{StringField: vendor.String("123")}},
		},
	}, {
		desc: "Map",
		in: &vendor.MessageWithMap{MsgMapping: map[int64]*vendor.FloatingPoint{
			0x4002: &vendor.FloatingPoint{
				Exact:            vendor.Bool(true),
				XXX_unrecognized: []byte("blah"),
			},
		}},
		want: &vendor.MessageWithMap{MsgMapping: map[int64]*vendor.FloatingPoint{
			0x4002: &vendor.FloatingPoint{Exact: vendor.Bool(true)},
		}},
	}, {
		desc: "Extension",
		in: func() vendor.Message {
			m := &vendor.MyMessage{
				Count: vendor.Int32(42),
				Somegroup: &vendor.MyMessage_SomeGroup{
					GroupField:       vendor.Int32(6),
					XXX_unrecognized: []byte("blah"),
				},
				XXX_unrecognized: []byte("blah"),
			}
			vendor.SetExtension(m, vendor.E_Ext_More, &vendor.Ext{
				Data:             vendor.String("extension"),
				XXX_unrecognized: []byte("blah"),
			})
			return m
		}(),
		want: func() vendor.Message {
			m := &vendor.MyMessage{
				Count:     vendor.Int32(42),
				Somegroup: &vendor.MyMessage_SomeGroup{GroupField: vendor.Int32(6)},
			}
			vendor.SetExtension(m, vendor.E_Ext_More, &vendor.Ext{Data: vendor.String("extension")})
			return m
		}(),
	}}

	// Test the legacy code path.
	for _, tt := range tests {
		// Clone the input so that we don't alter the original.
		in := tt.in
		if in != nil {
			in = vendor.Clone(tt.in)
		}

		var m LegacyMessage
		m.Message, _ = in.(*vendor.Message)
		m.Communique, _ = in.(*vendor.Communique)
		m.MessageWithMap, _ = in.(*vendor.MessageWithMap)
		m.MyMessage, _ = in.(*vendor.MyMessage)
		vendor.DiscardUnknown(&m)
		if !vendor.Equal(in, tt.want) {
			t.Errorf("test %s/Legacy, expected unknown fields to be discarded\ngot  %v\nwant %v", tt.desc, in, tt.want)
		}
	}

	for _, tt := range tests {
		vendor.DiscardUnknown(tt.in)
		if !vendor.Equal(tt.in, tt.want) {
			t.Errorf("test %s, expected unknown fields to be discarded\ngot  %v\nwant %v", tt.desc, tt.in, tt.want)
		}
	}
}

// LegacyMessage is a proto.Message that has several nested messages.
// This does not have the XXX_DiscardUnknown method and so forces DiscardUnknown
// to use the legacy fallback logic.
type LegacyMessage struct {
	Message        *vendor.Message
	Communique     *vendor.Communique
	MessageWithMap *vendor.MessageWithMap
	MyMessage      *vendor.MyMessage
}

func (m *LegacyMessage) Reset()         { *m = LegacyMessage{} }
func (m *LegacyMessage) String() string { return vendor.CompactTextString(m) }
func (*LegacyMessage) ProtoMessage()    {}
