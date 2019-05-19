// Go support for Protocol Buffers - Google's data interchange format
//
// Copyright 2011 The Go Authors.  All rights reserved.
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

var cloneTestMessage = &vendor.MyMessage{
	Count: vendor.Int32(42),
	Name:  vendor.String("Dave"),
	Pet:   []string{"bunny", "kitty", "horsey"},
	Inner: &vendor.InnerMessage{
		Host:      vendor.String("niles"),
		Port:      vendor.Int32(9099),
		Connected: vendor.Bool(true),
	},
	Others: []*vendor.OtherMessage{
		{
			Value: []byte("some bytes"),
		},
	},
	Somegroup: &vendor.MyMessage_SomeGroup{
		GroupField: vendor.Int32(6),
	},
	RepBytes: [][]byte{[]byte("sham"), []byte("wow")},
}

func init() {
	ext := &vendor.Ext{
		Data: vendor.String("extension"),
	}
	if err := vendor.SetExtension(cloneTestMessage, vendor.E_Ext_More, ext); err != nil {
		panic("SetExtension: " + err.Error())
	}
	if err := vendor.SetExtension(cloneTestMessage, vendor.E_Ext_Text, vendor.String("hello")); err != nil {
		panic("SetExtension: " + err.Error())
	}
	if err := vendor.SetExtension(cloneTestMessage, vendor.E_Greeting, []string{"one", "two"}); err != nil {
		panic("SetExtension: " + err.Error())
	}
}

func TestClone(t *testing.T) {
	// Create a clone using a marshal/unmarshal roundtrip.
	vanilla := new(vendor.MyMessage)
	b, err := vendor.Marshal(cloneTestMessage)
	if err != nil {
		t.Errorf("unexpected Marshal error: %v", err)
	}
	if err := vendor.Unmarshal(b, vanilla); err != nil {
		t.Errorf("unexpected Unarshal error: %v", err)
	}

	// Create a clone using Clone and verify that it is equal to the original.
	m := vendor.Clone(cloneTestMessage).(*vendor.MyMessage)
	if !vendor.Equal(m, cloneTestMessage) {
		t.Fatalf("Clone(%v) = %v", cloneTestMessage, m)
	}

	// Mutate the clone, which should not affect the original.
	x1, err := vendor.GetExtension(m, vendor.E_Ext_More)
	if err != nil {
		t.Errorf("unexpected GetExtension(%v) error: %v", vendor.E_Ext_More.Name, err)
	}
	x2, err := vendor.GetExtension(m, vendor.E_Ext_Text)
	if err != nil {
		t.Errorf("unexpected GetExtension(%v) error: %v", vendor.E_Ext_Text.Name, err)
	}
	x3, err := vendor.GetExtension(m, vendor.E_Greeting)
	if err != nil {
		t.Errorf("unexpected GetExtension(%v) error: %v", vendor.E_Greeting.Name, err)
	}
	*m.Inner.Port++
	*(x1.(*vendor.Ext)).Data = "blah blah"
	*(x2.(*string)) = "goodbye"
	x3.([]string)[0] = "zero"
	if !vendor.Equal(cloneTestMessage, vanilla) {
		t.Fatalf("mutation on original detected:\ngot  %v\nwant %v", cloneTestMessage, vanilla)
	}
}

func TestCloneNil(t *testing.T) {
	var m *vendor.MyMessage
	if c := vendor.Clone(m); !vendor.Equal(m, c) {
		t.Errorf("Clone(%v) = %v", m, c)
	}
}

var mergeTests = []struct {
	src, dst, want vendor.Message
}{
	{
		src: &vendor.MyMessage{
			Count: vendor.Int32(42),
		},
		dst: &vendor.MyMessage{
			Name: vendor.String("Dave"),
		},
		want: &vendor.MyMessage{
			Count: vendor.Int32(42),
			Name:  vendor.String("Dave"),
		},
	},
	{
		src: &vendor.MyMessage{
			Inner: &vendor.InnerMessage{
				Host:      vendor.String("hey"),
				Connected: vendor.Bool(true),
			},
			Pet: []string{"horsey"},
			Others: []*vendor.OtherMessage{
				{
					Value: []byte("some bytes"),
				},
			},
		},
		dst: &vendor.MyMessage{
			Inner: &vendor.InnerMessage{
				Host: vendor.String("niles"),
				Port: vendor.Int32(9099),
			},
			Pet: []string{"bunny", "kitty"},
			Others: []*vendor.OtherMessage{
				{
					Key: vendor.Int64(31415926535),
				},
				{
					// Explicitly test a src=nil field
					Inner: nil,
				},
			},
		},
		want: &vendor.MyMessage{
			Inner: &vendor.InnerMessage{
				Host:      vendor.String("hey"),
				Connected: vendor.Bool(true),
				Port:      vendor.Int32(9099),
			},
			Pet: []string{"bunny", "kitty", "horsey"},
			Others: []*vendor.OtherMessage{
				{
					Key: vendor.Int64(31415926535),
				},
				{},
				{
					Value: []byte("some bytes"),
				},
			},
		},
	},
	{
		src: &vendor.MyMessage{
			RepBytes: [][]byte{[]byte("wow")},
		},
		dst: &vendor.MyMessage{
			Somegroup: &vendor.MyMessage_SomeGroup{
				GroupField: vendor.Int32(6),
			},
			RepBytes: [][]byte{[]byte("sham")},
		},
		want: &vendor.MyMessage{
			Somegroup: &vendor.MyMessage_SomeGroup{
				GroupField: vendor.Int32(6),
			},
			RepBytes: [][]byte{[]byte("sham"), []byte("wow")},
		},
	},
	// Check that a scalar bytes field replaces rather than appends.
	{
		src:  &vendor.OtherMessage{Value: []byte("foo")},
		dst:  &vendor.OtherMessage{Value: []byte("bar")},
		want: &vendor.OtherMessage{Value: []byte("foo")},
	},
	{
		src: &vendor.MessageWithMap{
			NameMapping: map[int32]string{6: "Nigel"},
			MsgMapping: map[int64]*vendor.FloatingPoint{
				0x4001: &vendor.FloatingPoint{F: vendor.Float64(2.0)},
				0x4002: &vendor.FloatingPoint{
					F: vendor.Float64(2.0),
				},
			},
			ByteMapping: map[bool][]byte{true: []byte("wowsa")},
		},
		dst: &vendor.MessageWithMap{
			NameMapping: map[int32]string{
				6: "Bruce", // should be overwritten
				7: "Andrew",
			},
			MsgMapping: map[int64]*vendor.FloatingPoint{
				0x4002: &vendor.FloatingPoint{
					F:     vendor.Float64(3.0),
					Exact: vendor.Bool(true),
				}, // the entire message should be overwritten
			},
		},
		want: &vendor.MessageWithMap{
			NameMapping: map[int32]string{
				6: "Nigel",
				7: "Andrew",
			},
			MsgMapping: map[int64]*vendor.FloatingPoint{
				0x4001: &vendor.FloatingPoint{F: vendor.Float64(2.0)},
				0x4002: &vendor.FloatingPoint{
					F: vendor.Float64(2.0),
				},
			},
			ByteMapping: map[bool][]byte{true: []byte("wowsa")},
		},
	},
	// proto3 shouldn't merge zero values,
	// in the same way that proto2 shouldn't merge nils.
	{
		src: &vendor.Message{
			Name: "Aaron",
			Data: []byte(""), // zero value, but not nil
		},
		dst: &vendor.Message{
			HeightInCm: 176,
			Data:       []byte("texas!"),
		},
		want: &vendor.Message{
			Name:       "Aaron",
			HeightInCm: 176,
			Data:       []byte("texas!"),
		},
	},
	{ // Oneof fields should merge by assignment.
		src:  &vendor.Communique{Union: &vendor.Communique_Number{41}},
		dst:  &vendor.Communique{Union: &vendor.Communique_Name{"Bobby Tables"}},
		want: &vendor.Communique{Union: &vendor.Communique_Number{41}},
	},
	{ // Oneof nil is the same as not set.
		src:  &vendor.Communique{},
		dst:  &vendor.Communique{Union: &vendor.Communique_Name{"Bobby Tables"}},
		want: &vendor.Communique{Union: &vendor.Communique_Name{"Bobby Tables"}},
	},
	{
		src:  &vendor.Communique{Union: &vendor.Communique_Number{1337}},
		dst:  &vendor.Communique{},
		want: &vendor.Communique{Union: &vendor.Communique_Number{1337}},
	},
	{
		src:  &vendor.Communique{Union: &vendor.Communique_Col{vendor.MyMessage_RED}},
		dst:  &vendor.Communique{},
		want: &vendor.Communique{Union: &vendor.Communique_Col{vendor.MyMessage_RED}},
	},
	{
		src:  &vendor.Communique{Union: &vendor.Communique_Data{[]byte("hello")}},
		dst:  &vendor.Communique{},
		want: &vendor.Communique{Union: &vendor.Communique_Data{[]byte("hello")}},
	},
	{
		src:  &vendor.Communique{Union: &vendor.Communique_Msg{&vendor.Strings{BytesField: []byte{1, 2, 3}}}},
		dst:  &vendor.Communique{},
		want: &vendor.Communique{Union: &vendor.Communique_Msg{&vendor.Strings{BytesField: []byte{1, 2, 3}}}},
	},
	{
		src:  &vendor.Communique{Union: &vendor.Communique_Msg{}},
		dst:  &vendor.Communique{},
		want: &vendor.Communique{Union: &vendor.Communique_Msg{}},
	},
	{
		src:  &vendor.Communique{Union: &vendor.Communique_Msg{&vendor.Strings{StringField: vendor.String("123")}}},
		dst:  &vendor.Communique{Union: &vendor.Communique_Msg{&vendor.Strings{BytesField: []byte{1, 2, 3}}}},
		want: &vendor.Communique{Union: &vendor.Communique_Msg{&vendor.Strings{StringField: vendor.String("123"), BytesField: []byte{1, 2, 3}}}},
	},
	{
		src: &vendor.Message{
			Terrain: map[string]*vendor.Nested{
				"kay_a": &vendor.Nested{Cute: true},      // replace
				"kay_b": &vendor.Nested{Bunny: "rabbit"}, // insert
			},
		},
		dst: &vendor.Message{
			Terrain: map[string]*vendor.Nested{
				"kay_a": &vendor.Nested{Bunny: "lost"},  // replaced
				"kay_c": &vendor.Nested{Bunny: "bunny"}, // keep
			},
		},
		want: &vendor.Message{
			Terrain: map[string]*vendor.Nested{
				"kay_a": &vendor.Nested{Cute: true},
				"kay_b": &vendor.Nested{Bunny: "rabbit"},
				"kay_c": &vendor.Nested{Bunny: "bunny"},
			},
		},
	},
	{
		src: &vendor.GoTest{
			F_BoolRepeated:   []bool{},
			F_Int32Repeated:  []int32{},
			F_Int64Repeated:  []int64{},
			F_Uint32Repeated: []uint32{},
			F_Uint64Repeated: []uint64{},
			F_FloatRepeated:  []float32{},
			F_DoubleRepeated: []float64{},
			F_StringRepeated: []string{},
			F_BytesRepeated:  [][]byte{},
		},
		dst: &vendor.GoTest{},
		want: &vendor.GoTest{
			F_BoolRepeated:   []bool{},
			F_Int32Repeated:  []int32{},
			F_Int64Repeated:  []int64{},
			F_Uint32Repeated: []uint32{},
			F_Uint64Repeated: []uint64{},
			F_FloatRepeated:  []float32{},
			F_DoubleRepeated: []float64{},
			F_StringRepeated: []string{},
			F_BytesRepeated:  [][]byte{},
		},
	},
	{
		src: &vendor.GoTest{},
		dst: &vendor.GoTest{
			F_BoolRepeated:   []bool{},
			F_Int32Repeated:  []int32{},
			F_Int64Repeated:  []int64{},
			F_Uint32Repeated: []uint32{},
			F_Uint64Repeated: []uint64{},
			F_FloatRepeated:  []float32{},
			F_DoubleRepeated: []float64{},
			F_StringRepeated: []string{},
			F_BytesRepeated:  [][]byte{},
		},
		want: &vendor.GoTest{
			F_BoolRepeated:   []bool{},
			F_Int32Repeated:  []int32{},
			F_Int64Repeated:  []int64{},
			F_Uint32Repeated: []uint32{},
			F_Uint64Repeated: []uint64{},
			F_FloatRepeated:  []float32{},
			F_DoubleRepeated: []float64{},
			F_StringRepeated: []string{},
			F_BytesRepeated:  [][]byte{},
		},
	},
	{
		src: &vendor.GoTest{
			F_BytesRepeated: [][]byte{nil, []byte{}, []byte{0}},
		},
		dst: &vendor.GoTest{},
		want: &vendor.GoTest{
			F_BytesRepeated: [][]byte{nil, []byte{}, []byte{0}},
		},
	},
	{
		src: &vendor.MyMessage{
			Others: []*vendor.OtherMessage{},
		},
		dst: &vendor.MyMessage{},
		want: &vendor.MyMessage{
			Others: []*vendor.OtherMessage{},
		},
	},
}

func TestMerge(t *testing.T) {
	for _, m := range mergeTests {
		got := vendor.Clone(m.dst)
		if !vendor.Equal(got, m.dst) {
			t.Errorf("Clone()\ngot  %v\nwant %v", got, m.dst)
			continue
		}
		vendor.Merge(got, m.src)
		if !vendor.Equal(got, m.want) {
			t.Errorf("Merge(%v, %v)\ngot  %v\nwant %v", m.dst, m.src, got, m.want)
		}
	}
}
