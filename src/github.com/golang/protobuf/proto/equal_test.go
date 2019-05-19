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

// Four identical base messages.
// The init function adds extensions to some of them.
var messageWithoutExtension = &vendor.MyMessage{Count: vendor.Int32(7)}
var messageWithExtension1a = &vendor.MyMessage{Count: vendor.Int32(7)}
var messageWithExtension1b = &vendor.MyMessage{Count: vendor.Int32(7)}
var messageWithExtension2 = &vendor.MyMessage{Count: vendor.Int32(7)}
var messageWithExtension3a = &vendor.MyMessage{Count: vendor.Int32(7)}
var messageWithExtension3b = &vendor.MyMessage{Count: vendor.Int32(7)}
var messageWithExtension3c = &vendor.MyMessage{Count: vendor.Int32(7)}

// Two messages with non-message extensions.
var messageWithInt32Extension1 = &vendor.MyMessage{Count: vendor.Int32(8)}
var messageWithInt32Extension2 = &vendor.MyMessage{Count: vendor.Int32(8)}

func init() {
	ext1 := &vendor.Ext{Data: vendor.String("Kirk")}
	ext2 := &vendor.Ext{Data: vendor.String("Picard")}

	// messageWithExtension1a has ext1, but never marshals it.
	if err := vendor.SetExtension(messageWithExtension1a, vendor.E_Ext_More, ext1); err != nil {
		panic("SetExtension on 1a failed: " + err.Error())
	}

	// messageWithExtension1b is the unmarshaled form of messageWithExtension1a.
	if err := vendor.SetExtension(messageWithExtension1b, vendor.E_Ext_More, ext1); err != nil {
		panic("SetExtension on 1b failed: " + err.Error())
	}
	buf, err := vendor.Marshal(messageWithExtension1b)
	if err != nil {
		panic("Marshal of 1b failed: " + err.Error())
	}
	messageWithExtension1b.Reset()
	if err := vendor.Unmarshal(buf, messageWithExtension1b); err != nil {
		panic("Unmarshal of 1b failed: " + err.Error())
	}

	// messageWithExtension2 has ext2.
	if err := vendor.SetExtension(messageWithExtension2, vendor.E_Ext_More, ext2); err != nil {
		panic("SetExtension on 2 failed: " + err.Error())
	}

	if err := vendor.SetExtension(messageWithInt32Extension1, vendor.E_Ext_Number, vendor.Int32(23)); err != nil {
		panic("SetExtension on Int32-1 failed: " + err.Error())
	}
	if err := vendor.SetExtension(messageWithInt32Extension1, vendor.E_Ext_Number, vendor.Int32(24)); err != nil {
		panic("SetExtension on Int32-2 failed: " + err.Error())
	}

	// messageWithExtension3{a,b,c} has unregistered extension.
	if vendor.RegisteredExtensions(messageWithExtension3a)[200] != nil {
		panic("expect extension 200 unregistered")
	}
	bytes := []byte{
		0xc0, 0x0c, 0x01, // id=200, wiretype=0 (varint), data=1
	}
	bytes2 := []byte{
		0xc0, 0x0c, 0x02, // id=200, wiretype=0 (varint), data=2
	}
	vendor.SetRawExtension(messageWithExtension3a, 200, bytes)
	vendor.SetRawExtension(messageWithExtension3b, 200, bytes)
	vendor.SetRawExtension(messageWithExtension3c, 200, bytes2)
}

var EqualTests = []struct {
	desc string
	a, b vendor.Message
	exp  bool
}{
	{"different types", &vendor.GoEnum{}, &vendor.GoTestField{}, false},
	{"equal empty", &vendor.GoEnum{}, &vendor.GoEnum{}, true},
	{"nil vs nil", nil, nil, true},
	{"typed nil vs typed nil", (*vendor.GoEnum)(nil), (*vendor.GoEnum)(nil), true},
	{"typed nil vs empty", (*vendor.GoEnum)(nil), &vendor.GoEnum{}, false},
	{"different typed nil", (*vendor.GoEnum)(nil), (*vendor.GoTestField)(nil), false},

	{"one set field, one unset field", &vendor.GoTestField{Label: vendor.String("foo")}, &vendor.GoTestField{}, false},
	{"one set field zero, one unset field", &vendor.GoTest{Param: vendor.Int32(0)}, &vendor.GoTest{}, false},
	{"different set fields", &vendor.GoTestField{Label: vendor.String("foo")}, &vendor.GoTestField{Label: vendor.String("bar")}, false},
	{"equal set", &vendor.GoTestField{Label: vendor.String("foo")}, &vendor.GoTestField{Label: vendor.String("foo")}, true},

	{"repeated, one set", &vendor.GoTest{F_Int32Repeated: []int32{2, 3}}, &vendor.GoTest{}, false},
	{"repeated, different length", &vendor.GoTest{F_Int32Repeated: []int32{2, 3}}, &vendor.GoTest{F_Int32Repeated: []int32{2}}, false},
	{"repeated, different value", &vendor.GoTest{F_Int32Repeated: []int32{2}}, &vendor.GoTest{F_Int32Repeated: []int32{3}}, false},
	{"repeated, equal", &vendor.GoTest{F_Int32Repeated: []int32{2, 4}}, &vendor.GoTest{F_Int32Repeated: []int32{2, 4}}, true},
	{"repeated, nil equal nil", &vendor.GoTest{F_Int32Repeated: nil}, &vendor.GoTest{F_Int32Repeated: nil}, true},
	{"repeated, nil equal empty", &vendor.GoTest{F_Int32Repeated: nil}, &vendor.GoTest{F_Int32Repeated: []int32{}}, true},
	{"repeated, empty equal nil", &vendor.GoTest{F_Int32Repeated: []int32{}}, &vendor.GoTest{F_Int32Repeated: nil}, true},

	{
		"nested, different",
		&vendor.GoTest{RequiredField: &vendor.GoTestField{Label: vendor.String("foo")}},
		&vendor.GoTest{RequiredField: &vendor.GoTestField{Label: vendor.String("bar")}},
		false,
	},
	{
		"nested, equal",
		&vendor.GoTest{RequiredField: &vendor.GoTestField{Label: vendor.String("wow")}},
		&vendor.GoTest{RequiredField: &vendor.GoTestField{Label: vendor.String("wow")}},
		true,
	},

	{"bytes", &vendor.OtherMessage{Value: []byte("foo")}, &vendor.OtherMessage{Value: []byte("foo")}, true},
	{"bytes, empty", &vendor.OtherMessage{Value: []byte{}}, &vendor.OtherMessage{Value: []byte{}}, true},
	{"bytes, empty vs nil", &vendor.OtherMessage{Value: []byte{}}, &vendor.OtherMessage{Value: nil}, false},
	{
		"repeated bytes",
		&vendor.MyMessage{RepBytes: [][]byte{[]byte("sham"), []byte("wow")}},
		&vendor.MyMessage{RepBytes: [][]byte{[]byte("sham"), []byte("wow")}},
		true,
	},
	// In proto3, []byte{} and []byte(nil) are equal.
	{"proto3 bytes, empty vs nil", &vendor.Message{Data: []byte{}}, &vendor.Message{Data: nil}, true},

	{"extension vs. no extension", messageWithoutExtension, messageWithExtension1a, false},
	{"extension vs. same extension", messageWithExtension1a, messageWithExtension1b, true},
	{"extension vs. different extension", messageWithExtension1a, messageWithExtension2, false},

	{"int32 extension vs. itself", messageWithInt32Extension1, messageWithInt32Extension1, true},
	{"int32 extension vs. a different int32", messageWithInt32Extension1, messageWithInt32Extension2, false},

	{"unregistered extension same", messageWithExtension3a, messageWithExtension3b, true},
	{"unregistered extension different", messageWithExtension3a, messageWithExtension3c, false},

	{
		"message with group",
		&vendor.MyMessage{
			Count: vendor.Int32(1),
			Somegroup: &vendor.MyMessage_SomeGroup{
				GroupField: vendor.Int32(5),
			},
		},
		&vendor.MyMessage{
			Count: vendor.Int32(1),
			Somegroup: &vendor.MyMessage_SomeGroup{
				GroupField: vendor.Int32(5),
			},
		},
		true,
	},

	{
		"map same",
		&vendor.MessageWithMap{NameMapping: map[int32]string{1: "Ken"}},
		&vendor.MessageWithMap{NameMapping: map[int32]string{1: "Ken"}},
		true,
	},
	{
		"map different entry",
		&vendor.MessageWithMap{NameMapping: map[int32]string{1: "Ken"}},
		&vendor.MessageWithMap{NameMapping: map[int32]string{2: "Rob"}},
		false,
	},
	{
		"map different key only",
		&vendor.MessageWithMap{NameMapping: map[int32]string{1: "Ken"}},
		&vendor.MessageWithMap{NameMapping: map[int32]string{2: "Ken"}},
		false,
	},
	{
		"map different value only",
		&vendor.MessageWithMap{NameMapping: map[int32]string{1: "Ken"}},
		&vendor.MessageWithMap{NameMapping: map[int32]string{1: "Rob"}},
		false,
	},
	{
		"zero-length maps same",
		&vendor.MessageWithMap{NameMapping: map[int32]string{}},
		&vendor.MessageWithMap{NameMapping: nil},
		true,
	},
	{
		"orders in map don't matter",
		&vendor.MessageWithMap{NameMapping: map[int32]string{1: "Ken", 2: "Rob"}},
		&vendor.MessageWithMap{NameMapping: map[int32]string{2: "Rob", 1: "Ken"}},
		true,
	},
	{
		"oneof same",
		&vendor.Communique{Union: &vendor.Communique_Number{41}},
		&vendor.Communique{Union: &vendor.Communique_Number{41}},
		true,
	},
	{
		"oneof one nil",
		&vendor.Communique{Union: &vendor.Communique_Number{41}},
		&vendor.Communique{},
		false,
	},
	{
		"oneof different",
		&vendor.Communique{Union: &vendor.Communique_Number{41}},
		&vendor.Communique{Union: &vendor.Communique_Name{"Bobby Tables"}},
		false,
	},
}

func TestEqual(t *testing.T) {
	for _, tc := range EqualTests {
		if res := vendor.Equal(tc.a, tc.b); res != tc.exp {
			t.Errorf("%v: Equal(%v, %v) = %v, want %v", tc.desc, tc.a, tc.b, res, tc.exp)
		}
	}
}
