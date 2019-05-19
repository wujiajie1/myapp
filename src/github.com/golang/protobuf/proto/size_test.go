// Go support for Protocol Buffers - Google's data interchange format
//
// Copyright 2012 The Go Authors.  All rights reserved.
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
	"log"
	"strings"
	"testing"
	"vendor"
)

var messageWithExtension1 = &vendor.MyMessage{Count: vendor.Int32(7)}

// messageWithExtension2 is in equal_test.go.
var messageWithExtension3 = &vendor.MyMessage{Count: vendor.Int32(8)}

func init() {
	if err := vendor.SetExtension(messageWithExtension1, vendor.E_Ext_More, &vendor.Ext{Data: vendor.String("Abbott")}); err != nil {
		log.Panicf("SetExtension: %v", err)
	}
	if err := vendor.SetExtension(messageWithExtension3, vendor.E_Ext_More, &vendor.Ext{Data: vendor.String("Costello")}); err != nil {
		log.Panicf("SetExtension: %v", err)
	}

	// Force messageWithExtension3 to have the extension encoded.
	vendor.Marshal(messageWithExtension3)

}

// non-pointer custom message
type nonptrMessage struct{}

func (m nonptrMessage) ProtoMessage()  {}
func (m nonptrMessage) Reset()         {}
func (m nonptrMessage) String() string { return "" }

func (m nonptrMessage) Marshal() ([]byte, error) {
	return []byte{42}, nil
}

// custom message embedding a proto.Message
type messageWithEmbedding struct {
	*vendor.OtherMessage
}

func (m *messageWithEmbedding) ProtoMessage()  {}
func (m *messageWithEmbedding) Reset()         {}
func (m *messageWithEmbedding) String() string { return "" }

func (m *messageWithEmbedding) Marshal() ([]byte, error) {
	return []byte{42}, nil
}

var SizeTests = []struct {
	desc string
	pb   vendor.Message
}{
	{"empty", &vendor.OtherMessage{}},
	// Basic types.
	{"bool", &vendor.Defaults{F_Bool: vendor.Bool(true)}},
	{"int32", &vendor.Defaults{F_Int32: vendor.Int32(12)}},
	{"negative int32", &vendor.Defaults{F_Int32: vendor.Int32(-1)}},
	{"small int64", &vendor.Defaults{F_Int64: vendor.Int64(1)}},
	{"big int64", &vendor.Defaults{F_Int64: vendor.Int64(1 << 20)}},
	{"negative int64", &vendor.Defaults{F_Int64: vendor.Int64(-1)}},
	{"fixed32", &vendor.Defaults{F_Fixed32: vendor.Uint32(71)}},
	{"fixed64", &vendor.Defaults{F_Fixed64: vendor.Uint64(72)}},
	{"uint32", &vendor.Defaults{F_Uint32: vendor.Uint32(123)}},
	{"uint64", &vendor.Defaults{F_Uint64: vendor.Uint64(124)}},
	{"float", &vendor.Defaults{F_Float: vendor.Float32(12.6)}},
	{"double", &vendor.Defaults{F_Double: vendor.Float64(13.9)}},
	{"string", &vendor.Defaults{F_String: vendor.String("niles")}},
	{"bytes", &vendor.Defaults{F_Bytes: []byte("wowsa")}},
	{"bytes, empty", &vendor.Defaults{F_Bytes: []byte{}}},
	{"sint32", &vendor.Defaults{F_Sint32: vendor.Int32(65)}},
	{"sint64", &vendor.Defaults{F_Sint64: vendor.Int64(67)}},
	{"enum", &vendor.Defaults{F_Enum: vendor.Defaults_BLUE.Enum()}},
	// Repeated.
	{"empty repeated bool", &vendor.MoreRepeated{Bools: []bool{}}},
	{"repeated bool", &vendor.MoreRepeated{Bools: []bool{false, true, true, false}}},
	{"packed repeated bool", &vendor.MoreRepeated{BoolsPacked: []bool{false, true, true, false, true, true, true}}},
	{"repeated int32", &vendor.MoreRepeated{Ints: []int32{1, 12203, 1729, -1}}},
	{"repeated int32 packed", &vendor.MoreRepeated{IntsPacked: []int32{1, 12203, 1729}}},
	{"repeated int64 packed", &vendor.MoreRepeated{Int64SPacked: []int64{
		// Need enough large numbers to verify that the header is counting the number of bytes
		// for the field, not the number of elements.
		1 << 62, 1 << 62, 1 << 62, 1 << 62, 1 << 62, 1 << 62, 1 << 62, 1 << 62, 1 << 62, 1 << 62,
		1 << 62, 1 << 62, 1 << 62, 1 << 62, 1 << 62, 1 << 62, 1 << 62, 1 << 62, 1 << 62, 1 << 62,
	}}},
	{"repeated string", &vendor.MoreRepeated{Strings: []string{"r", "ken", "gri"}}},
	{"repeated fixed", &vendor.MoreRepeated{Fixeds: []uint32{1, 2, 3, 4}}},
	// Nested.
	{"nested", &vendor.OldMessage{Nested: &vendor.OldMessage_Nested{Name: vendor.String("whatever")}}},
	{"group", &vendor.GroupOld{G: &vendor.GroupOld_G{X: vendor.Int32(12345)}}},
	// Other things.
	{"unrecognized", &vendor.MoreRepeated{XXX_unrecognized: []byte{13<<3 | 0, 4}}},
	{"extension (unencoded)", messageWithExtension1},
	{"extension (encoded)", messageWithExtension3},
	// proto3 message
	{"proto3 empty", &vendor.Message{}},
	{"proto3 bool", &vendor.Message{TrueScotsman: true}},
	{"proto3 int64", &vendor.Message{ResultCount: 1}},
	{"proto3 uint32", &vendor.Message{HeightInCm: 123}},
	{"proto3 float", &vendor.Message{Score: 12.6}},
	{"proto3 string", &vendor.Message{Name: "Snezana"}},
	{"proto3 bytes", &vendor.Message{Data: []byte("wowsa")}},
	{"proto3 bytes, empty", &vendor.Message{Data: []byte{}}},
	{"proto3 enum", &vendor.Message{Hilarity: vendor.Message_PUNS}},
	{"proto3 map field with empty bytes", &vendor.MessageWithMap{ByteMapping: map[bool][]byte{false: []byte{}}}},

	{"map field", &vendor.MessageWithMap{NameMapping: map[int32]string{1: "Rob", 7: "Andrew"}}},
	{"map field with message", &vendor.MessageWithMap{MsgMapping: map[int64]*vendor.FloatingPoint{0x7001: &vendor.FloatingPoint{F: vendor.Float64(2.0)}}}},
	{"map field with bytes", &vendor.MessageWithMap{ByteMapping: map[bool][]byte{true: []byte("this time for sure")}}},
	{"map field with empty bytes", &vendor.MessageWithMap{ByteMapping: map[bool][]byte{true: []byte{}}}},

	{"map field with big entry", &vendor.MessageWithMap{NameMapping: map[int32]string{8: strings.Repeat("x", 125)}}},
	{"map field with big key and val", &vendor.MessageWithMap{StrToStr: map[string]string{strings.Repeat("x", 70): strings.Repeat("y", 70)}}},
	{"map field with big numeric key", &vendor.MessageWithMap{NameMapping: map[int32]string{0xf00d: "om nom nom"}}},

	{"oneof not set", &vendor.Oneof{}},
	{"oneof bool", &vendor.Oneof{Union: &vendor.Oneof_F_Bool{true}}},
	{"oneof zero int32", &vendor.Oneof{Union: &vendor.Oneof_F_Int32{0}}},
	{"oneof big int32", &vendor.Oneof{Union: &vendor.Oneof_F_Int32{1 << 20}}},
	{"oneof int64", &vendor.Oneof{Union: &vendor.Oneof_F_Int64{42}}},
	{"oneof fixed32", &vendor.Oneof{Union: &vendor.Oneof_F_Fixed32{43}}},
	{"oneof fixed64", &vendor.Oneof{Union: &vendor.Oneof_F_Fixed64{44}}},
	{"oneof uint32", &vendor.Oneof{Union: &vendor.Oneof_F_Uint32{45}}},
	{"oneof uint64", &vendor.Oneof{Union: &vendor.Oneof_F_Uint64{46}}},
	{"oneof float", &vendor.Oneof{Union: &vendor.Oneof_F_Float{47.1}}},
	{"oneof double", &vendor.Oneof{Union: &vendor.Oneof_F_Double{48.9}}},
	{"oneof string", &vendor.Oneof{Union: &vendor.Oneof_F_String{"Rhythmic Fman"}}},
	{"oneof bytes", &vendor.Oneof{Union: &vendor.Oneof_F_Bytes{[]byte("let go")}}},
	{"oneof sint32", &vendor.Oneof{Union: &vendor.Oneof_F_Sint32{50}}},
	{"oneof sint64", &vendor.Oneof{Union: &vendor.Oneof_F_Sint64{51}}},
	{"oneof enum", &vendor.Oneof{Union: &vendor.Oneof_F_Enum{vendor.MyMessage_BLUE}}},
	{"message for oneof", &vendor.GoTestField{Label: vendor.String("k"), Type: vendor.String("v")}},
	{"oneof message", &vendor.Oneof{Union: &vendor.Oneof_F_Message{&vendor.GoTestField{Label: vendor.String("k"), Type: vendor.String("v")}}}},
	{"oneof group", &vendor.Oneof{Union: &vendor.Oneof_FGroup{&vendor.Oneof_F_Group{X: vendor.Int32(52)}}}},
	{"oneof largest tag", &vendor.Oneof{Union: &vendor.Oneof_F_Largest_Tag{1}}},
	{"multiple oneofs", &vendor.Oneof{Union: &vendor.Oneof_F_Int32{1}, Tormato: &vendor.Oneof_Value{2}}},

	{"non-pointer message", nonptrMessage{}},
	{"custom message with embedding", &messageWithEmbedding{&vendor.OtherMessage{}}},
}

func TestSize(t *testing.T) {
	for _, tc := range SizeTests {
		size := vendor.Size(tc.pb)
		b, err := vendor.Marshal(tc.pb)
		if err != nil {
			t.Errorf("%v: Marshal failed: %v", tc.desc, err)
			continue
		}
		if size != len(b) {
			t.Errorf("%v: Size(%v) = %d, want %d", tc.desc, tc.pb, size, len(b))
			t.Logf("%v: bytes: %#v", tc.desc, b)
		}
	}
}
