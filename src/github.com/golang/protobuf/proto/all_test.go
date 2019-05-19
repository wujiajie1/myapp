// Go support for Protocol Buffers - Google's data interchange format
//
// Copyright 2010 The Go Authors.  All rights reserved.
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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"runtime/debug"
	"strings"
	"sync"
	"testing"
	"time"
	"vendor"

	. "github.com/golang/protobuf/proto/test_proto"
)

var globalO *vendor.Buffer

func old() *vendor.Buffer {
	if globalO == nil {
		globalO = vendor.NewBuffer(nil)
	}
	globalO.Reset()
	return globalO
}

func equalbytes(b1, b2 []byte, t *testing.T) {
	if len(b1) != len(b2) {
		t.Errorf("wrong lengths: 2*%d != %d", len(b1), len(b2))
		return
	}
	for i := 0; i < len(b1); i++ {
		if b1[i] != b2[i] {
			t.Errorf("bad byte[%d]:%x %x: %s %s", i, b1[i], b2[i], b1, b2)
		}
	}
}

func initGoTestField() *vendor.GoTestField {
	f := new(vendor.GoTestField)
	f.Label = vendor.String("label")
	f.Type = vendor.String("type")
	return f
}

// These are all structurally equivalent but the tag numbers differ.
// (It's remarkable that required, optional, and repeated all have
// 8 letters.)
func initGoTest_RequiredGroup() *vendor.GoTest_RequiredGroup {
	return &vendor.GoTest_RequiredGroup{
		RequiredField: vendor.String("required"),
	}
}

func initGoTest_OptionalGroup() *vendor.GoTest_OptionalGroup {
	return &vendor.GoTest_OptionalGroup{
		RequiredField: vendor.String("optional"),
	}
}

func initGoTest_RepeatedGroup() *vendor.GoTest_RepeatedGroup {
	return &vendor.GoTest_RepeatedGroup{
		RequiredField: vendor.String("repeated"),
	}
}

func initGoTest(setdefaults bool) *vendor.GoTest {
	pb := new(vendor.GoTest)
	if setdefaults {
		pb.F_BoolDefaulted = vendor.Bool(vendor.Default_GoTest_F_BoolDefaulted)
		pb.F_Int32Defaulted = vendor.Int32(vendor.Default_GoTest_F_Int32Defaulted)
		pb.F_Int64Defaulted = vendor.Int64(vendor.Default_GoTest_F_Int64Defaulted)
		pb.F_Fixed32Defaulted = vendor.Uint32(vendor.Default_GoTest_F_Fixed32Defaulted)
		pb.F_Fixed64Defaulted = vendor.Uint64(vendor.Default_GoTest_F_Fixed64Defaulted)
		pb.F_Uint32Defaulted = vendor.Uint32(vendor.Default_GoTest_F_Uint32Defaulted)
		pb.F_Uint64Defaulted = vendor.Uint64(vendor.Default_GoTest_F_Uint64Defaulted)
		pb.F_FloatDefaulted = vendor.Float32(vendor.Default_GoTest_F_FloatDefaulted)
		pb.F_DoubleDefaulted = vendor.Float64(vendor.Default_GoTest_F_DoubleDefaulted)
		pb.F_StringDefaulted = vendor.String(vendor.Default_GoTest_F_StringDefaulted)
		pb.F_BytesDefaulted = vendor.Default_GoTest_F_BytesDefaulted
		pb.F_Sint32Defaulted = vendor.Int32(vendor.Default_GoTest_F_Sint32Defaulted)
		pb.F_Sint64Defaulted = vendor.Int64(vendor.Default_GoTest_F_Sint64Defaulted)
		pb.F_Sfixed32Defaulted = vendor.Int32(vendor.Default_GoTest_F_Sfixed32Defaulted)
		pb.F_Sfixed64Defaulted = vendor.Int64(vendor.Default_GoTest_F_Sfixed64Defaulted)
	}

	pb.Kind = vendor.GoTest_TIME.Enum()
	pb.RequiredField = initGoTestField()
	pb.F_BoolRequired = vendor.Bool(true)
	pb.F_Int32Required = vendor.Int32(3)
	pb.F_Int64Required = vendor.Int64(6)
	pb.F_Fixed32Required = vendor.Uint32(32)
	pb.F_Fixed64Required = vendor.Uint64(64)
	pb.F_Uint32Required = vendor.Uint32(3232)
	pb.F_Uint64Required = vendor.Uint64(6464)
	pb.F_FloatRequired = vendor.Float32(3232)
	pb.F_DoubleRequired = vendor.Float64(6464)
	pb.F_StringRequired = vendor.String("string")
	pb.F_BytesRequired = []byte("bytes")
	pb.F_Sint32Required = vendor.Int32(-32)
	pb.F_Sint64Required = vendor.Int64(-64)
	pb.F_Sfixed32Required = vendor.Int32(-32)
	pb.F_Sfixed64Required = vendor.Int64(-64)
	pb.Requiredgroup = initGoTest_RequiredGroup()

	return pb
}

func hex(c uint8) uint8 {
	if '0' <= c && c <= '9' {
		return c - '0'
	}
	if 'a' <= c && c <= 'f' {
		return 10 + c - 'a'
	}
	if 'A' <= c && c <= 'F' {
		return 10 + c - 'A'
	}
	return 0
}

func equal(b []byte, s string, t *testing.T) bool {
	if 2*len(b) != len(s) {
		//		fail(fmt.Sprintf("wrong lengths: 2*%d != %d", len(b), len(s)), b, s, t)
		fmt.Printf("wrong lengths: 2*%d != %d\n", len(b), len(s))
		return false
	}
	for i, j := 0, 0; i < len(b); i, j = i+1, j+2 {
		x := hex(s[j])*16 + hex(s[j+1])
		if b[i] != x {
			//			fail(fmt.Sprintf("bad byte[%d]:%x %x", i, b[i], x), b, s, t)
			fmt.Printf("bad byte[%d]:%x %x", i, b[i], x)
			return false
		}
	}
	return true
}

func overify(t *testing.T, pb *vendor.GoTest, expected string) {
	o := old()
	err := o.Marshal(pb)
	if err != nil {
		fmt.Printf("overify marshal-1 err = %v", err)
		o.DebugPrint("", o.Bytes())
		t.Fatalf("expected = %s", expected)
	}
	if !equal(o.Bytes(), expected, t) {
		o.DebugPrint("overify neq 1", o.Bytes())
		t.Fatalf("expected = %s", expected)
	}

	// Now test Unmarshal by recreating the original buffer.
	pbd := new(vendor.GoTest)
	err = o.Unmarshal(pbd)
	if err != nil {
		t.Fatalf("overify unmarshal err = %v", err)
		o.DebugPrint("", o.Bytes())
		t.Fatalf("string = %s", expected)
	}
	o.Reset()
	err = o.Marshal(pbd)
	if err != nil {
		t.Errorf("overify marshal-2 err = %v", err)
		o.DebugPrint("", o.Bytes())
		t.Fatalf("string = %s", expected)
	}
	if !equal(o.Bytes(), expected, t) {
		o.DebugPrint("overify neq 2", o.Bytes())
		t.Fatalf("string = %s", expected)
	}
}

// Simple tests for numeric encode/decode primitives (varint, etc.)
func TestNumericPrimitives(t *testing.T) {
	for i := uint64(0); i < 1e6; i += 111 {
		o := old()
		if o.EncodeVarint(i) != nil {
			t.Error("EncodeVarint")
			break
		}
		x, e := o.DecodeVarint()
		if e != nil {
			t.Fatal("DecodeVarint")
		}
		if x != i {
			t.Fatal("varint decode fail:", i, x)
		}

		o = old()
		if o.EncodeFixed32(i) != nil {
			t.Fatal("encFixed32")
		}
		x, e = o.DecodeFixed32()
		if e != nil {
			t.Fatal("decFixed32")
		}
		if x != i {
			t.Fatal("fixed32 decode fail:", i, x)
		}

		o = old()
		if o.EncodeFixed64(i*1234567) != nil {
			t.Error("encFixed64")
			break
		}
		x, e = o.DecodeFixed64()
		if e != nil {
			t.Error("decFixed64")
			break
		}
		if x != i*1234567 {
			t.Error("fixed64 decode fail:", i*1234567, x)
			break
		}

		o = old()
		i32 := int32(i - 12345)
		if o.EncodeZigzag32(uint64(i32)) != nil {
			t.Fatal("EncodeZigzag32")
		}
		x, e = o.DecodeZigzag32()
		if e != nil {
			t.Fatal("DecodeZigzag32")
		}
		if x != uint64(uint32(i32)) {
			t.Fatal("zigzag32 decode fail:", i32, x)
		}

		o = old()
		i64 := int64(i - 12345)
		if o.EncodeZigzag64(uint64(i64)) != nil {
			t.Fatal("EncodeZigzag64")
		}
		x, e = o.DecodeZigzag64()
		if e != nil {
			t.Fatal("DecodeZigzag64")
		}
		if x != uint64(i64) {
			t.Fatal("zigzag64 decode fail:", i64, x)
		}
	}
}

// fakeMarshaler is a simple struct implementing Marshaler and Message interfaces.
type fakeMarshaler struct {
	b   []byte
	err error
}

func (f *fakeMarshaler) Marshal() ([]byte, error) { return f.b, f.err }
func (f *fakeMarshaler) String() string           { return fmt.Sprintf("Bytes: %v Error: %v", f.b, f.err) }
func (f *fakeMarshaler) ProtoMessage()            {}
func (f *fakeMarshaler) Reset()                   {}

type msgWithFakeMarshaler struct {
	M *fakeMarshaler `protobuf:"bytes,1,opt,name=fake"`
}

func (m *msgWithFakeMarshaler) String() string { return vendor.CompactTextString(m) }
func (m *msgWithFakeMarshaler) ProtoMessage()  {}
func (m *msgWithFakeMarshaler) Reset()         {}

// Simple tests for proto messages that implement the Marshaler interface.
func TestMarshalerEncoding(t *testing.T) {
	tests := []struct {
		name    string
		m       vendor.Message
		want    []byte
		errType reflect.Type
	}{
		{
			name: "Marshaler that fails",
			m: &fakeMarshaler{
				err: errors.New("some marshal err"),
				b:   []byte{5, 6, 7},
			},
			// Since the Marshal method returned bytes, they should be written to the
			// buffer.  (For efficiency, we assume that Marshal implementations are
			// always correct w.r.t. RequiredNotSetError and output.)
			want:    []byte{5, 6, 7},
			errType: reflect.TypeOf(errors.New("some marshal err")),
		},
		{
			name: "Marshaler that fails with RequiredNotSetError",
			m: &msgWithFakeMarshaler{
				M: &fakeMarshaler{
					err: &vendor.RequiredNotSetError{},
					b:   []byte{5, 6, 7},
				},
			},
			// Since there's an error that can be continued after,
			// the buffer should be written.
			want: []byte{
				10, 3, // for &msgWithFakeMarshaler
				5, 6, 7, // for &fakeMarshaler
			},
			errType: reflect.TypeOf(&vendor.RequiredNotSetError{}),
		},
		{
			name: "Marshaler that succeeds",
			m: &fakeMarshaler{
				b: []byte{0, 1, 2, 3, 4, 127, 255},
			},
			want: []byte{0, 1, 2, 3, 4, 127, 255},
		},
	}
	for _, test := range tests {
		b := vendor.NewBuffer(nil)
		err := b.Marshal(test.m)
		if reflect.TypeOf(err) != test.errType {
			t.Errorf("%s: got err %T(%v) wanted %T", test.name, err, err, test.errType)
		}
		if !reflect.DeepEqual(test.want, b.Bytes()) {
			t.Errorf("%s: got bytes %v wanted %v", test.name, b.Bytes(), test.want)
		}
		if size := vendor.Size(test.m); size != len(b.Bytes()) {
			t.Errorf("%s: Size(_) = %v, but marshaled to %v bytes", test.name, size, len(b.Bytes()))
		}

		m, mErr := vendor.Marshal(test.m)
		if !bytes.Equal(b.Bytes(), m) {
			t.Errorf("%s: Marshal returned %v, but (*Buffer).Marshal wrote %v", test.name, m, b.Bytes())
		}
		if !reflect.DeepEqual(err, mErr) {
			t.Errorf("%s: Marshal err = %q, but (*Buffer).Marshal returned %q",
				test.name, fmt.Sprint(mErr), fmt.Sprint(err))
		}
	}
}

// Ensure that Buffer.Marshal uses O(N) memory for N messages
func TestBufferMarshalAllocs(t *testing.T) {
	value := &vendor.OtherMessage{Key: vendor.Int64(1)}
	msg := &vendor.MyMessage{Count: vendor.Int32(1), Others: []*vendor.OtherMessage{value}}

	reallocSize := func(t *testing.T, items int, prealloc int) (int64, int64) {
		var b vendor.Buffer
		b.SetBuf(make([]byte, 0, prealloc))

		var allocSpace int64
		prevCap := cap(b.Bytes())
		for i := 0; i < items; i++ {
			err := b.Marshal(msg)
			if err != nil {
				t.Errorf("Marshal err = %q", err)
				break
			}
			if c := cap(b.Bytes()); prevCap != c {
				allocSpace += int64(c)
				prevCap = c
			}
		}
		needSpace := int64(len(b.Bytes()))
		return allocSpace, needSpace
	}

	for _, prealloc := range []int{0, 100, 10000} {
		for _, items := range []int{1, 2, 5, 10, 20, 50, 100, 200, 500, 1000} {
			runtimeSpace, need := reallocSize(t, items, prealloc)
			totalSpace := int64(prealloc) + runtimeSpace

			runtimeRatio := float64(runtimeSpace) / float64(need)
			totalRatio := float64(totalSpace) / float64(need)

			if totalRatio < 1 || runtimeRatio > 4 {
				t.Errorf("needed %dB, allocated %dB total (ratio %.1f), allocated %dB at runtime (ratio %.1f)",
					need, totalSpace, totalRatio, runtimeSpace, runtimeRatio)
			}
		}
	}
}

// Simple tests for bytes
func TestBytesPrimitives(t *testing.T) {
	o := old()
	bytes := []byte{'n', 'o', 'w', ' ', 'i', 's', ' ', 't', 'h', 'e', ' ', 't', 'i', 'm', 'e'}
	if o.EncodeRawBytes(bytes) != nil {
		t.Error("EncodeRawBytes")
	}
	decb, e := o.DecodeRawBytes(false)
	if e != nil {
		t.Error("DecodeRawBytes")
	}
	equalbytes(bytes, decb, t)
}

// Simple tests for strings
func TestStringPrimitives(t *testing.T) {
	o := old()
	s := "now is the time"
	if o.EncodeStringBytes(s) != nil {
		t.Error("enc_string")
	}
	decs, e := o.DecodeStringBytes()
	if e != nil {
		t.Error("dec_string")
	}
	if s != decs {
		t.Error("string encode/decode fail:", s, decs)
	}
}

// Do we catch the "required bit not set" case?
func TestRequiredBit(t *testing.T) {
	o := old()
	pb := new(vendor.GoTest)
	err := o.Marshal(pb)
	if err == nil {
		t.Error("did not catch missing required fields")
	} else if !strings.Contains(err.Error(), "Kind") {
		t.Error("wrong error type:", err)
	}
}

// Check that all fields are nil.
// Clearly silly, and a residue from a more interesting test with an earlier,
// different initialization property, but it once caught a compiler bug so
// it lives.
func checkInitialized(pb *vendor.GoTest, t *testing.T) {
	if pb.F_BoolDefaulted != nil {
		t.Error("New or Reset did not set boolean:", *pb.F_BoolDefaulted)
	}
	if pb.F_Int32Defaulted != nil {
		t.Error("New or Reset did not set int32:", *pb.F_Int32Defaulted)
	}
	if pb.F_Int64Defaulted != nil {
		t.Error("New or Reset did not set int64:", *pb.F_Int64Defaulted)
	}
	if pb.F_Fixed32Defaulted != nil {
		t.Error("New or Reset did not set fixed32:", *pb.F_Fixed32Defaulted)
	}
	if pb.F_Fixed64Defaulted != nil {
		t.Error("New or Reset did not set fixed64:", *pb.F_Fixed64Defaulted)
	}
	if pb.F_Uint32Defaulted != nil {
		t.Error("New or Reset did not set uint32:", *pb.F_Uint32Defaulted)
	}
	if pb.F_Uint64Defaulted != nil {
		t.Error("New or Reset did not set uint64:", *pb.F_Uint64Defaulted)
	}
	if pb.F_FloatDefaulted != nil {
		t.Error("New or Reset did not set float:", *pb.F_FloatDefaulted)
	}
	if pb.F_DoubleDefaulted != nil {
		t.Error("New or Reset did not set double:", *pb.F_DoubleDefaulted)
	}
	if pb.F_StringDefaulted != nil {
		t.Error("New or Reset did not set string:", *pb.F_StringDefaulted)
	}
	if pb.F_BytesDefaulted != nil {
		t.Error("New or Reset did not set bytes:", string(pb.F_BytesDefaulted))
	}
	if pb.F_Sint32Defaulted != nil {
		t.Error("New or Reset did not set int32:", *pb.F_Sint32Defaulted)
	}
	if pb.F_Sint64Defaulted != nil {
		t.Error("New or Reset did not set int64:", *pb.F_Sint64Defaulted)
	}
}

// Does Reset() reset?
func TestReset(t *testing.T) {
	pb := initGoTest(true)
	// muck with some values
	pb.F_BoolDefaulted = vendor.Bool(false)
	pb.F_Int32Defaulted = vendor.Int32(237)
	pb.F_Int64Defaulted = vendor.Int64(12346)
	pb.F_Fixed32Defaulted = vendor.Uint32(32000)
	pb.F_Fixed64Defaulted = vendor.Uint64(666)
	pb.F_Uint32Defaulted = vendor.Uint32(323232)
	pb.F_Uint64Defaulted = nil
	pb.F_FloatDefaulted = nil
	pb.F_DoubleDefaulted = vendor.Float64(0)
	pb.F_StringDefaulted = vendor.String("gotcha")
	pb.F_BytesDefaulted = []byte("asdfasdf")
	pb.F_Sint32Defaulted = vendor.Int32(123)
	pb.F_Sint64Defaulted = vendor.Int64(789)
	pb.Reset()
	checkInitialized(pb, t)
}

// All required fields set, no defaults provided.
func TestEncodeDecode1(t *testing.T) {
	pb := initGoTest(false)
	overify(t, pb,
		"0807"+ // field 1, encoding 0, value 7
			"220d"+"0a056c6162656c120474797065"+ // field 4, encoding 2 (GoTestField)
			"5001"+ // field 10, encoding 0, value 1
			"5803"+ // field 11, encoding 0, value 3
			"6006"+ // field 12, encoding 0, value 6
			"6d20000000"+ // field 13, encoding 5, value 0x20
			"714000000000000000"+ // field 14, encoding 1, value 0x40
			"78a019"+ // field 15, encoding 0, value 0xca0 = 3232
			"8001c032"+ // field 16, encoding 0, value 0x1940 = 6464
			"8d0100004a45"+ // field 17, encoding 5, value 3232.0
			"9101000000000040b940"+ // field 18, encoding 1, value 6464.0
			"9a0106"+"737472696e67"+ // field 19, encoding 2, string "string"
			"b304"+ // field 70, encoding 3, start group
			"ba0408"+"7265717569726564"+ // field 71, encoding 2, string "required"
			"b404"+ // field 70, encoding 4, end group
			"aa0605"+"6279746573"+ // field 101, encoding 2, string "bytes"
			"b0063f"+ // field 102, encoding 0, 0x3f zigzag32
			"b8067f"+ // field 103, encoding 0, 0x7f zigzag64
			"c506e0ffffff"+ // field 104, encoding 5, -32 fixed32
			"c906c0ffffffffffffff") // field 105, encoding 1, -64 fixed64
}

// All required fields set, defaults provided.
func TestEncodeDecode2(t *testing.T) {
	pb := initGoTest(true)
	overify(t, pb,
		"0807"+ // field 1, encoding 0, value 7
			"220d"+"0a056c6162656c120474797065"+ // field 4, encoding 2 (GoTestField)
			"5001"+ // field 10, encoding 0, value 1
			"5803"+ // field 11, encoding 0, value 3
			"6006"+ // field 12, encoding 0, value 6
			"6d20000000"+ // field 13, encoding 5, value 32
			"714000000000000000"+ // field 14, encoding 1, value 64
			"78a019"+ // field 15, encoding 0, value 3232
			"8001c032"+ // field 16, encoding 0, value 6464
			"8d0100004a45"+ // field 17, encoding 5, value 3232.0
			"9101000000000040b940"+ // field 18, encoding 1, value 6464.0
			"9a0106"+"737472696e67"+ // field 19, encoding 2 string "string"
			"c00201"+ // field 40, encoding 0, value 1
			"c80220"+ // field 41, encoding 0, value 32
			"d00240"+ // field 42, encoding 0, value 64
			"dd0240010000"+ // field 43, encoding 5, value 320
			"e1028002000000000000"+ // field 44, encoding 1, value 640
			"e8028019"+ // field 45, encoding 0, value 3200
			"f0028032"+ // field 46, encoding 0, value 6400
			"fd02e0659948"+ // field 47, encoding 5, value 314159.0
			"81030000000050971041"+ // field 48, encoding 1, value 271828.0
			"8a0310"+"68656c6c6f2c2022776f726c6421220a"+ // field 49, encoding 2 string "hello, \"world!\"\n"
			"b304"+ // start group field 70 level 1
			"ba0408"+"7265717569726564"+ // field 71, encoding 2, string "required"
			"b404"+ // end group field 70 level 1
			"aa0605"+"6279746573"+ // field 101, encoding 2 string "bytes"
			"b0063f"+ // field 102, encoding 0, 0x3f zigzag32
			"b8067f"+ // field 103, encoding 0, 0x7f zigzag64
			"c506e0ffffff"+ // field 104, encoding 5, -32 fixed32
			"c906c0ffffffffffffff"+ // field 105, encoding 1, -64 fixed64
			"8a1907"+"4269676e6f7365"+ // field 401, encoding 2, string "Bignose"
			"90193f"+ // field 402, encoding 0, value 63
			"98197f"+ // field 403, encoding 0, value 127
			"a519e0ffffff"+ // field 404, encoding 5, -32 fixed32
			"a919c0ffffffffffffff") // field 405, encoding 1, -64 fixed64

}

// All default fields set to their default value by hand
func TestEncodeDecode3(t *testing.T) {
	pb := initGoTest(false)
	pb.F_BoolDefaulted = vendor.Bool(true)
	pb.F_Int32Defaulted = vendor.Int32(32)
	pb.F_Int64Defaulted = vendor.Int64(64)
	pb.F_Fixed32Defaulted = vendor.Uint32(320)
	pb.F_Fixed64Defaulted = vendor.Uint64(640)
	pb.F_Uint32Defaulted = vendor.Uint32(3200)
	pb.F_Uint64Defaulted = vendor.Uint64(6400)
	pb.F_FloatDefaulted = vendor.Float32(314159)
	pb.F_DoubleDefaulted = vendor.Float64(271828)
	pb.F_StringDefaulted = vendor.String("hello, \"world!\"\n")
	pb.F_BytesDefaulted = []byte("Bignose")
	pb.F_Sint32Defaulted = vendor.Int32(-32)
	pb.F_Sint64Defaulted = vendor.Int64(-64)
	pb.F_Sfixed32Defaulted = vendor.Int32(-32)
	pb.F_Sfixed64Defaulted = vendor.Int64(-64)

	overify(t, pb,
		"0807"+ // field 1, encoding 0, value 7
			"220d"+"0a056c6162656c120474797065"+ // field 4, encoding 2 (GoTestField)
			"5001"+ // field 10, encoding 0, value 1
			"5803"+ // field 11, encoding 0, value 3
			"6006"+ // field 12, encoding 0, value 6
			"6d20000000"+ // field 13, encoding 5, value 32
			"714000000000000000"+ // field 14, encoding 1, value 64
			"78a019"+ // field 15, encoding 0, value 3232
			"8001c032"+ // field 16, encoding 0, value 6464
			"8d0100004a45"+ // field 17, encoding 5, value 3232.0
			"9101000000000040b940"+ // field 18, encoding 1, value 6464.0
			"9a0106"+"737472696e67"+ // field 19, encoding 2 string "string"
			"c00201"+ // field 40, encoding 0, value 1
			"c80220"+ // field 41, encoding 0, value 32
			"d00240"+ // field 42, encoding 0, value 64
			"dd0240010000"+ // field 43, encoding 5, value 320
			"e1028002000000000000"+ // field 44, encoding 1, value 640
			"e8028019"+ // field 45, encoding 0, value 3200
			"f0028032"+ // field 46, encoding 0, value 6400
			"fd02e0659948"+ // field 47, encoding 5, value 314159.0
			"81030000000050971041"+ // field 48, encoding 1, value 271828.0
			"8a0310"+"68656c6c6f2c2022776f726c6421220a"+ // field 49, encoding 2 string "hello, \"world!\"\n"
			"b304"+ // start group field 70 level 1
			"ba0408"+"7265717569726564"+ // field 71, encoding 2, string "required"
			"b404"+ // end group field 70 level 1
			"aa0605"+"6279746573"+ // field 101, encoding 2 string "bytes"
			"b0063f"+ // field 102, encoding 0, 0x3f zigzag32
			"b8067f"+ // field 103, encoding 0, 0x7f zigzag64
			"c506e0ffffff"+ // field 104, encoding 5, -32 fixed32
			"c906c0ffffffffffffff"+ // field 105, encoding 1, -64 fixed64
			"8a1907"+"4269676e6f7365"+ // field 401, encoding 2, string "Bignose"
			"90193f"+ // field 402, encoding 0, value 63
			"98197f"+ // field 403, encoding 0, value 127
			"a519e0ffffff"+ // field 404, encoding 5, -32 fixed32
			"a919c0ffffffffffffff") // field 405, encoding 1, -64 fixed64

}

// All required fields set, defaults provided, all non-defaulted optional fields have values.
func TestEncodeDecode4(t *testing.T) {
	pb := initGoTest(true)
	pb.Table = vendor.String("hello")
	pb.Param = vendor.Int32(7)
	pb.OptionalField = initGoTestField()
	pb.F_BoolOptional = vendor.Bool(true)
	pb.F_Int32Optional = vendor.Int32(32)
	pb.F_Int64Optional = vendor.Int64(64)
	pb.F_Fixed32Optional = vendor.Uint32(3232)
	pb.F_Fixed64Optional = vendor.Uint64(6464)
	pb.F_Uint32Optional = vendor.Uint32(323232)
	pb.F_Uint64Optional = vendor.Uint64(646464)
	pb.F_FloatOptional = vendor.Float32(32.)
	pb.F_DoubleOptional = vendor.Float64(64.)
	pb.F_StringOptional = vendor.String("hello")
	pb.F_BytesOptional = []byte("Bignose")
	pb.F_Sint32Optional = vendor.Int32(-32)
	pb.F_Sint64Optional = vendor.Int64(-64)
	pb.F_Sfixed32Optional = vendor.Int32(-32)
	pb.F_Sfixed64Optional = vendor.Int64(-64)
	pb.Optionalgroup = initGoTest_OptionalGroup()

	overify(t, pb,
		"0807"+ // field 1, encoding 0, value 7
			"1205"+"68656c6c6f"+ // field 2, encoding 2, string "hello"
			"1807"+ // field 3, encoding 0, value 7
			"220d"+"0a056c6162656c120474797065"+ // field 4, encoding 2 (GoTestField)
			"320d"+"0a056c6162656c120474797065"+ // field 6, encoding 2 (GoTestField)
			"5001"+ // field 10, encoding 0, value 1
			"5803"+ // field 11, encoding 0, value 3
			"6006"+ // field 12, encoding 0, value 6
			"6d20000000"+ // field 13, encoding 5, value 32
			"714000000000000000"+ // field 14, encoding 1, value 64
			"78a019"+ // field 15, encoding 0, value 3232
			"8001c032"+ // field 16, encoding 0, value 6464
			"8d0100004a45"+ // field 17, encoding 5, value 3232.0
			"9101000000000040b940"+ // field 18, encoding 1, value 6464.0
			"9a0106"+"737472696e67"+ // field 19, encoding 2 string "string"
			"f00101"+ // field 30, encoding 0, value 1
			"f80120"+ // field 31, encoding 0, value 32
			"800240"+ // field 32, encoding 0, value 64
			"8d02a00c0000"+ // field 33, encoding 5, value 3232
			"91024019000000000000"+ // field 34, encoding 1, value 6464
			"9802a0dd13"+ // field 35, encoding 0, value 323232
			"a002c0ba27"+ // field 36, encoding 0, value 646464
			"ad0200000042"+ // field 37, encoding 5, value 32.0
			"b1020000000000005040"+ // field 38, encoding 1, value 64.0
			"ba0205"+"68656c6c6f"+ // field 39, encoding 2, string "hello"
			"c00201"+ // field 40, encoding 0, value 1
			"c80220"+ // field 41, encoding 0, value 32
			"d00240"+ // field 42, encoding 0, value 64
			"dd0240010000"+ // field 43, encoding 5, value 320
			"e1028002000000000000"+ // field 44, encoding 1, value 640
			"e8028019"+ // field 45, encoding 0, value 3200
			"f0028032"+ // field 46, encoding 0, value 6400
			"fd02e0659948"+ // field 47, encoding 5, value 314159.0
			"81030000000050971041"+ // field 48, encoding 1, value 271828.0
			"8a0310"+"68656c6c6f2c2022776f726c6421220a"+ // field 49, encoding 2 string "hello, \"world!\"\n"
			"b304"+ // start group field 70 level 1
			"ba0408"+"7265717569726564"+ // field 71, encoding 2, string "required"
			"b404"+ // end group field 70 level 1
			"d305"+ // start group field 90 level 1
			"da0508"+"6f7074696f6e616c"+ // field 91, encoding 2, string "optional"
			"d405"+ // end group field 90 level 1
			"aa0605"+"6279746573"+ // field 101, encoding 2 string "bytes"
			"b0063f"+ // field 102, encoding 0, 0x3f zigzag32
			"b8067f"+ // field 103, encoding 0, 0x7f zigzag64
			"c506e0ffffff"+ // field 104, encoding 5, -32 fixed32
			"c906c0ffffffffffffff"+ // field 105, encoding 1, -64 fixed64
			"ea1207"+"4269676e6f7365"+ // field 301, encoding 2, string "Bignose"
			"f0123f"+ // field 302, encoding 0, value 63
			"f8127f"+ // field 303, encoding 0, value 127
			"8513e0ffffff"+ // field 304, encoding 5, -32 fixed32
			"8913c0ffffffffffffff"+ // field 305, encoding 1, -64 fixed64
			"8a1907"+"4269676e6f7365"+ // field 401, encoding 2, string "Bignose"
			"90193f"+ // field 402, encoding 0, value 63
			"98197f"+ // field 403, encoding 0, value 127
			"a519e0ffffff"+ // field 404, encoding 5, -32 fixed32
			"a919c0ffffffffffffff") // field 405, encoding 1, -64 fixed64

}

// All required fields set, defaults provided, all repeated fields given two values.
func TestEncodeDecode5(t *testing.T) {
	pb := initGoTest(true)
	pb.RepeatedField = []*vendor.GoTestField{initGoTestField(), initGoTestField()}
	pb.F_BoolRepeated = []bool{false, true}
	pb.F_Int32Repeated = []int32{32, 33}
	pb.F_Int64Repeated = []int64{64, 65}
	pb.F_Fixed32Repeated = []uint32{3232, 3333}
	pb.F_Fixed64Repeated = []uint64{6464, 6565}
	pb.F_Uint32Repeated = []uint32{323232, 333333}
	pb.F_Uint64Repeated = []uint64{646464, 656565}
	pb.F_FloatRepeated = []float32{32., 33.}
	pb.F_DoubleRepeated = []float64{64., 65.}
	pb.F_StringRepeated = []string{"hello", "sailor"}
	pb.F_BytesRepeated = [][]byte{[]byte("big"), []byte("nose")}
	pb.F_Sint32Repeated = []int32{32, -32}
	pb.F_Sint64Repeated = []int64{64, -64}
	pb.F_Sfixed32Repeated = []int32{32, -32}
	pb.F_Sfixed64Repeated = []int64{64, -64}
	pb.Repeatedgroup = []*vendor.GoTest_RepeatedGroup{initGoTest_RepeatedGroup(), initGoTest_RepeatedGroup()}

	overify(t, pb,
		"0807"+ // field 1, encoding 0, value 7
			"220d"+"0a056c6162656c120474797065"+ // field 4, encoding 2 (GoTestField)
			"2a0d"+"0a056c6162656c120474797065"+ // field 5, encoding 2 (GoTestField)
			"2a0d"+"0a056c6162656c120474797065"+ // field 5, encoding 2 (GoTestField)
			"5001"+ // field 10, encoding 0, value 1
			"5803"+ // field 11, encoding 0, value 3
			"6006"+ // field 12, encoding 0, value 6
			"6d20000000"+ // field 13, encoding 5, value 32
			"714000000000000000"+ // field 14, encoding 1, value 64
			"78a019"+ // field 15, encoding 0, value 3232
			"8001c032"+ // field 16, encoding 0, value 6464
			"8d0100004a45"+ // field 17, encoding 5, value 3232.0
			"9101000000000040b940"+ // field 18, encoding 1, value 6464.0
			"9a0106"+"737472696e67"+ // field 19, encoding 2 string "string"
			"a00100"+ // field 20, encoding 0, value 0
			"a00101"+ // field 20, encoding 0, value 1
			"a80120"+ // field 21, encoding 0, value 32
			"a80121"+ // field 21, encoding 0, value 33
			"b00140"+ // field 22, encoding 0, value 64
			"b00141"+ // field 22, encoding 0, value 65
			"bd01a00c0000"+ // field 23, encoding 5, value 3232
			"bd01050d0000"+ // field 23, encoding 5, value 3333
			"c1014019000000000000"+ // field 24, encoding 1, value 6464
			"c101a519000000000000"+ // field 24, encoding 1, value 6565
			"c801a0dd13"+ // field 25, encoding 0, value 323232
			"c80195ac14"+ // field 25, encoding 0, value 333333
			"d001c0ba27"+ // field 26, encoding 0, value 646464
			"d001b58928"+ // field 26, encoding 0, value 656565
			"dd0100000042"+ // field 27, encoding 5, value 32.0
			"dd0100000442"+ // field 27, encoding 5, value 33.0
			"e1010000000000005040"+ // field 28, encoding 1, value 64.0
			"e1010000000000405040"+ // field 28, encoding 1, value 65.0
			"ea0105"+"68656c6c6f"+ // field 29, encoding 2, string "hello"
			"ea0106"+"7361696c6f72"+ // field 29, encoding 2, string "sailor"
			"c00201"+ // field 40, encoding 0, value 1
			"c80220"+ // field 41, encoding 0, value 32
			"d00240"+ // field 42, encoding 0, value 64
			"dd0240010000"+ // field 43, encoding 5, value 320
			"e1028002000000000000"+ // field 44, encoding 1, value 640
			"e8028019"+ // field 45, encoding 0, value 3200
			"f0028032"+ // field 46, encoding 0, value 6400
			"fd02e0659948"+ // field 47, encoding 5, value 314159.0
			"81030000000050971041"+ // field 48, encoding 1, value 271828.0
			"8a0310"+"68656c6c6f2c2022776f726c6421220a"+ // field 49, encoding 2 string "hello, \"world!\"\n"
			"b304"+ // start group field 70 level 1
			"ba0408"+"7265717569726564"+ // field 71, encoding 2, string "required"
			"b404"+ // end group field 70 level 1
			"8305"+ // start group field 80 level 1
			"8a0508"+"7265706561746564"+ // field 81, encoding 2, string "repeated"
			"8405"+ // end group field 80 level 1
			"8305"+ // start group field 80 level 1
			"8a0508"+"7265706561746564"+ // field 81, encoding 2, string "repeated"
			"8405"+ // end group field 80 level 1
			"aa0605"+"6279746573"+ // field 101, encoding 2 string "bytes"
			"b0063f"+ // field 102, encoding 0, 0x3f zigzag32
			"b8067f"+ // field 103, encoding 0, 0x7f zigzag64
			"c506e0ffffff"+ // field 104, encoding 5, -32 fixed32
			"c906c0ffffffffffffff"+ // field 105, encoding 1, -64 fixed64
			"ca0c03"+"626967"+ // field 201, encoding 2, string "big"
			"ca0c04"+"6e6f7365"+ // field 201, encoding 2, string "nose"
			"d00c40"+ // field 202, encoding 0, value 32
			"d00c3f"+ // field 202, encoding 0, value -32
			"d80c8001"+ // field 203, encoding 0, value 64
			"d80c7f"+ // field 203, encoding 0, value -64
			"e50c20000000"+ // field 204, encoding 5, 32 fixed32
			"e50ce0ffffff"+ // field 204, encoding 5, -32 fixed32
			"e90c4000000000000000"+ // field 205, encoding 1, 64 fixed64
			"e90cc0ffffffffffffff"+ // field 205, encoding 1, -64 fixed64
			"8a1907"+"4269676e6f7365"+ // field 401, encoding 2, string "Bignose"
			"90193f"+ // field 402, encoding 0, value 63
			"98197f"+ // field 403, encoding 0, value 127
			"a519e0ffffff"+ // field 404, encoding 5, -32 fixed32
			"a919c0ffffffffffffff") // field 405, encoding 1, -64 fixed64

}

// All required fields set, all packed repeated fields given two values.
func TestEncodeDecode6(t *testing.T) {
	pb := initGoTest(false)
	pb.F_BoolRepeatedPacked = []bool{false, true}
	pb.F_Int32RepeatedPacked = []int32{32, 33}
	pb.F_Int64RepeatedPacked = []int64{64, 65}
	pb.F_Fixed32RepeatedPacked = []uint32{3232, 3333}
	pb.F_Fixed64RepeatedPacked = []uint64{6464, 6565}
	pb.F_Uint32RepeatedPacked = []uint32{323232, 333333}
	pb.F_Uint64RepeatedPacked = []uint64{646464, 656565}
	pb.F_FloatRepeatedPacked = []float32{32., 33.}
	pb.F_DoubleRepeatedPacked = []float64{64., 65.}
	pb.F_Sint32RepeatedPacked = []int32{32, -32}
	pb.F_Sint64RepeatedPacked = []int64{64, -64}
	pb.F_Sfixed32RepeatedPacked = []int32{32, -32}
	pb.F_Sfixed64RepeatedPacked = []int64{64, -64}

	overify(t, pb,
		"0807"+ // field 1, encoding 0, value 7
			"220d"+"0a056c6162656c120474797065"+ // field 4, encoding 2 (GoTestField)
			"5001"+ // field 10, encoding 0, value 1
			"5803"+ // field 11, encoding 0, value 3
			"6006"+ // field 12, encoding 0, value 6
			"6d20000000"+ // field 13, encoding 5, value 32
			"714000000000000000"+ // field 14, encoding 1, value 64
			"78a019"+ // field 15, encoding 0, value 3232
			"8001c032"+ // field 16, encoding 0, value 6464
			"8d0100004a45"+ // field 17, encoding 5, value 3232.0
			"9101000000000040b940"+ // field 18, encoding 1, value 6464.0
			"9a0106"+"737472696e67"+ // field 19, encoding 2 string "string"
			"9203020001"+ // field 50, encoding 2, 2 bytes, value 0, value 1
			"9a03022021"+ // field 51, encoding 2, 2 bytes, value 32, value 33
			"a203024041"+ // field 52, encoding 2, 2 bytes, value 64, value 65
			"aa0308"+ // field 53, encoding 2, 8 bytes
			"a00c0000050d0000"+ // value 3232, value 3333
			"b20310"+ // field 54, encoding 2, 16 bytes
			"4019000000000000a519000000000000"+ // value 6464, value 6565
			"ba0306"+ // field 55, encoding 2, 6 bytes
			"a0dd1395ac14"+ // value 323232, value 333333
			"c20306"+ // field 56, encoding 2, 6 bytes
			"c0ba27b58928"+ // value 646464, value 656565
			"ca0308"+ // field 57, encoding 2, 8 bytes
			"0000004200000442"+ // value 32.0, value 33.0
			"d20310"+ // field 58, encoding 2, 16 bytes
			"00000000000050400000000000405040"+ // value 64.0, value 65.0
			"b304"+ // start group field 70 level 1
			"ba0408"+"7265717569726564"+ // field 71, encoding 2, string "required"
			"b404"+ // end group field 70 level 1
			"aa0605"+"6279746573"+ // field 101, encoding 2 string "bytes"
			"b0063f"+ // field 102, encoding 0, 0x3f zigzag32
			"b8067f"+ // field 103, encoding 0, 0x7f zigzag64
			"c506e0ffffff"+ // field 104, encoding 5, -32 fixed32
			"c906c0ffffffffffffff"+ // field 105, encoding 1, -64 fixed64
			"b21f02"+ // field 502, encoding 2, 2 bytes
			"403f"+ // value 32, value -32
			"ba1f03"+ // field 503, encoding 2, 3 bytes
			"80017f"+ // value 64, value -64
			"c21f08"+ // field 504, encoding 2, 8 bytes
			"20000000e0ffffff"+ // value 32, value -32
			"ca1f10"+ // field 505, encoding 2, 16 bytes
			"4000000000000000c0ffffffffffffff") // value 64, value -64

}

// Test that we can encode empty bytes fields.
func TestEncodeDecodeBytes1(t *testing.T) {
	pb := initGoTest(false)

	// Create our bytes
	pb.F_BytesRequired = []byte{}
	pb.F_BytesRepeated = [][]byte{{}}
	pb.F_BytesOptional = []byte{}

	d, err := vendor.Marshal(pb)
	if err != nil {
		t.Error(err)
	}

	pbd := new(vendor.GoTest)
	if err := vendor.Unmarshal(d, pbd); err != nil {
		t.Error(err)
	}

	if pbd.F_BytesRequired == nil || len(pbd.F_BytesRequired) != 0 {
		t.Error("required empty bytes field is incorrect")
	}
	if pbd.F_BytesRepeated == nil || len(pbd.F_BytesRepeated) == 1 && pbd.F_BytesRepeated[0] == nil {
		t.Error("repeated empty bytes field is incorrect")
	}
	if pbd.F_BytesOptional == nil || len(pbd.F_BytesOptional) != 0 {
		t.Error("optional empty bytes field is incorrect")
	}
}

// Test that we encode nil-valued fields of a repeated bytes field correctly.
// Since entries in a repeated field cannot be nil, nil must mean empty value.
func TestEncodeDecodeBytes2(t *testing.T) {
	pb := initGoTest(false)

	// Create our bytes
	pb.F_BytesRepeated = [][]byte{nil}

	d, err := vendor.Marshal(pb)
	if err != nil {
		t.Error(err)
	}

	pbd := new(vendor.GoTest)
	if err := vendor.Unmarshal(d, pbd); err != nil {
		t.Error(err)
	}

	if len(pbd.F_BytesRepeated) != 1 || pbd.F_BytesRepeated[0] == nil {
		t.Error("Unexpected value for repeated bytes field")
	}
}

// All required fields set, defaults provided, all repeated fields given two values.
func TestSkippingUnrecognizedFields(t *testing.T) {
	o := old()
	pb := initGoTestField()

	// Marshal it normally.
	o.Marshal(pb)

	// Now new a GoSkipTest record.
	skip := &vendor.GoSkipTest{
		SkipInt32:   vendor.Int32(32),
		SkipFixed32: vendor.Uint32(3232),
		SkipFixed64: vendor.Uint64(6464),
		SkipString:  vendor.String("skipper"),
		Skipgroup: &vendor.GoSkipTest_SkipGroup{
			GroupInt32:  vendor.Int32(75),
			GroupString: vendor.String("wxyz"),
		},
	}

	// Marshal it into same buffer.
	o.Marshal(skip)

	pbd := new(vendor.GoTestField)
	o.Unmarshal(pbd)

	// The __unrecognized field should be a marshaling of GoSkipTest
	skipd := new(vendor.GoSkipTest)

	o.SetBuf(pbd.XXX_unrecognized)
	o.Unmarshal(skipd)

	if *skipd.SkipInt32 != *skip.SkipInt32 {
		t.Error("skip int32", skipd.SkipInt32)
	}
	if *skipd.SkipFixed32 != *skip.SkipFixed32 {
		t.Error("skip fixed32", skipd.SkipFixed32)
	}
	if *skipd.SkipFixed64 != *skip.SkipFixed64 {
		t.Error("skip fixed64", skipd.SkipFixed64)
	}
	if *skipd.SkipString != *skip.SkipString {
		t.Error("skip string", *skipd.SkipString)
	}
	if *skipd.Skipgroup.GroupInt32 != *skip.Skipgroup.GroupInt32 {
		t.Error("skip group int32", skipd.Skipgroup.GroupInt32)
	}
	if *skipd.Skipgroup.GroupString != *skip.Skipgroup.GroupString {
		t.Error("skip group string", *skipd.Skipgroup.GroupString)
	}
}

// Check that unrecognized fields of a submessage are preserved.
func TestSubmessageUnrecognizedFields(t *testing.T) {
	nm := &vendor.NewMessage{
		Nested: &vendor.NewMessage_Nested{
			Name:      vendor.String("Nigel"),
			FoodGroup: vendor.String("carbs"),
		},
	}
	b, err := vendor.Marshal(nm)
	if err != nil {
		t.Fatalf("Marshal of NewMessage: %v", err)
	}

	// Unmarshal into an OldMessage.
	om := new(vendor.OldMessage)
	if err := vendor.Unmarshal(b, om); err != nil {
		t.Fatalf("Unmarshal to OldMessage: %v", err)
	}
	exp := &vendor.OldMessage{
		Nested: &vendor.OldMessage_Nested{
			Name: vendor.String("Nigel"),
			// normal protocol buffer users should not do this
			XXX_unrecognized: []byte("\x12\x05carbs"),
		},
	}
	if !vendor.Equal(om, exp) {
		t.Errorf("om = %v, want %v", om, exp)
	}

	// Clone the OldMessage.
	om = vendor.Clone(om).(*vendor.OldMessage)
	if !vendor.Equal(om, exp) {
		t.Errorf("Clone(om) = %v, want %v", om, exp)
	}

	// Marshal the OldMessage, then unmarshal it into an empty NewMessage.
	if b, err = vendor.Marshal(om); err != nil {
		t.Fatalf("Marshal of OldMessage: %v", err)
	}
	t.Logf("Marshal(%v) -> %q", om, b)
	nm2 := new(vendor.NewMessage)
	if err := vendor.Unmarshal(b, nm2); err != nil {
		t.Fatalf("Unmarshal to NewMessage: %v", err)
	}
	if !vendor.Equal(nm, nm2) {
		t.Errorf("NewMessage round-trip: %v => %v", nm, nm2)
	}
}

// Check that an int32 field can be upgraded to an int64 field.
func TestNegativeInt32(t *testing.T) {
	om := &vendor.OldMessage{
		Num: vendor.Int32(-1),
	}
	b, err := vendor.Marshal(om)
	if err != nil {
		t.Fatalf("Marshal of OldMessage: %v", err)
	}

	// Check the size. It should be 11 bytes;
	// 1 for the field/wire type, and 10 for the negative number.
	if len(b) != 11 {
		t.Errorf("%v marshaled as %q, wanted 11 bytes", om, b)
	}

	// Unmarshal into a NewMessage.
	nm := new(vendor.NewMessage)
	if err := vendor.Unmarshal(b, nm); err != nil {
		t.Fatalf("Unmarshal to NewMessage: %v", err)
	}
	want := &vendor.NewMessage{
		Num: vendor.Int64(-1),
	}
	if !vendor.Equal(nm, want) {
		t.Errorf("nm = %v, want %v", nm, want)
	}
}

// Check that we can grow an array (repeated field) to have many elements.
// This test doesn't depend only on our encoding; for variety, it makes sure
// we create, encode, and decode the correct contents explicitly.  It's therefore
// a bit messier.
// This test also uses (and hence tests) the Marshal/Unmarshal functions
// instead of the methods.
func TestBigRepeated(t *testing.T) {
	pb := initGoTest(true)

	// Create the arrays
	const N = 50 // Internally the library starts much smaller.
	pb.Repeatedgroup = make([]*vendor.GoTest_RepeatedGroup, N)
	pb.F_Sint64Repeated = make([]int64, N)
	pb.F_Sint32Repeated = make([]int32, N)
	pb.F_BytesRepeated = make([][]byte, N)
	pb.F_StringRepeated = make([]string, N)
	pb.F_DoubleRepeated = make([]float64, N)
	pb.F_FloatRepeated = make([]float32, N)
	pb.F_Uint64Repeated = make([]uint64, N)
	pb.F_Uint32Repeated = make([]uint32, N)
	pb.F_Fixed64Repeated = make([]uint64, N)
	pb.F_Fixed32Repeated = make([]uint32, N)
	pb.F_Int64Repeated = make([]int64, N)
	pb.F_Int32Repeated = make([]int32, N)
	pb.F_BoolRepeated = make([]bool, N)
	pb.RepeatedField = make([]*vendor.GoTestField, N)

	// Fill in the arrays with checkable values.
	igtf := initGoTestField()
	igtrg := initGoTest_RepeatedGroup()
	for i := 0; i < N; i++ {
		pb.Repeatedgroup[i] = igtrg
		pb.F_Sint64Repeated[i] = int64(i)
		pb.F_Sint32Repeated[i] = int32(i)
		s := fmt.Sprint(i)
		pb.F_BytesRepeated[i] = []byte(s)
		pb.F_StringRepeated[i] = s
		pb.F_DoubleRepeated[i] = float64(i)
		pb.F_FloatRepeated[i] = float32(i)
		pb.F_Uint64Repeated[i] = uint64(i)
		pb.F_Uint32Repeated[i] = uint32(i)
		pb.F_Fixed64Repeated[i] = uint64(i)
		pb.F_Fixed32Repeated[i] = uint32(i)
		pb.F_Int64Repeated[i] = int64(i)
		pb.F_Int32Repeated[i] = int32(i)
		pb.F_BoolRepeated[i] = i%2 == 0
		pb.RepeatedField[i] = igtf
	}

	// Marshal.
	buf, _ := vendor.Marshal(pb)

	// Now test Unmarshal by recreating the original buffer.
	pbd := new(vendor.GoTest)
	vendor.Unmarshal(buf, pbd)

	// Check the checkable values
	for i := uint64(0); i < N; i++ {
		if pbd.Repeatedgroup[i] == nil { // TODO: more checking?
			t.Error("pbd.Repeatedgroup bad")
		}
		if x := uint64(pbd.F_Sint64Repeated[i]); x != i {
			t.Error("pbd.F_Sint64Repeated bad", x, i)
		}
		if x := uint64(pbd.F_Sint32Repeated[i]); x != i {
			t.Error("pbd.F_Sint32Repeated bad", x, i)
		}
		s := fmt.Sprint(i)
		equalbytes(pbd.F_BytesRepeated[i], []byte(s), t)
		if pbd.F_StringRepeated[i] != s {
			t.Error("pbd.F_Sint32Repeated bad", pbd.F_StringRepeated[i], i)
		}
		if x := uint64(pbd.F_DoubleRepeated[i]); x != i {
			t.Error("pbd.F_DoubleRepeated bad", x, i)
		}
		if x := uint64(pbd.F_FloatRepeated[i]); x != i {
			t.Error("pbd.F_FloatRepeated bad", x, i)
		}
		if x := pbd.F_Uint64Repeated[i]; x != i {
			t.Error("pbd.F_Uint64Repeated bad", x, i)
		}
		if x := uint64(pbd.F_Uint32Repeated[i]); x != i {
			t.Error("pbd.F_Uint32Repeated bad", x, i)
		}
		if x := pbd.F_Fixed64Repeated[i]; x != i {
			t.Error("pbd.F_Fixed64Repeated bad", x, i)
		}
		if x := uint64(pbd.F_Fixed32Repeated[i]); x != i {
			t.Error("pbd.F_Fixed32Repeated bad", x, i)
		}
		if x := uint64(pbd.F_Int64Repeated[i]); x != i {
			t.Error("pbd.F_Int64Repeated bad", x, i)
		}
		if x := uint64(pbd.F_Int32Repeated[i]); x != i {
			t.Error("pbd.F_Int32Repeated bad", x, i)
		}
		if x := pbd.F_BoolRepeated[i]; x != (i%2 == 0) {
			t.Error("pbd.F_BoolRepeated bad", x, i)
		}
		if pbd.RepeatedField[i] == nil { // TODO: more checking?
			t.Error("pbd.RepeatedField bad")
		}
	}
}

func TestBadWireTypeUnknown(t *testing.T) {
	var b []byte
	fmt.Sscanf("0a01780d00000000080b101612036161611521000000202c220362626225370000002203636363214200000000000000584d5a036464645900000000000056405d63000000", "%x", &b)

	m := new(vendor.MyMessage)
	if err := vendor.Unmarshal(b, m); err != nil {
		t.Errorf("unexpected Unmarshal error: %v", err)
	}

	var unknown []byte
	fmt.Sscanf("0a01780d0000000010161521000000202c2537000000214200000000000000584d5a036464645d63000000", "%x", &unknown)
	if !bytes.Equal(m.XXX_unrecognized, unknown) {
		t.Errorf("unknown bytes mismatch:\ngot  %x\nwant %x", m.XXX_unrecognized, unknown)
	}
	vendor.DiscardUnknown(m)

	want := &vendor.MyMessage{Count: vendor.Int32(11), Name: vendor.String("aaa"), Pet: []string{"bbb", "ccc"}, Bigfloat: vendor.Float64(88)}
	if !vendor.Equal(m, want) {
		t.Errorf("message mismatch:\ngot  %v\nwant %v", m, want)
	}
}

func encodeDecode(t *testing.T, in, out vendor.Message, msg string) {
	buf, err := vendor.Marshal(in)
	if err != nil {
		t.Fatalf("failed marshaling %v: %v", msg, err)
	}
	if err := vendor.Unmarshal(buf, out); err != nil {
		t.Fatalf("failed unmarshaling %v: %v", msg, err)
	}
}

func TestPackedNonPackedDecoderSwitching(t *testing.T) {
	np, p := new(vendor.NonPackedTest), new(vendor.PackedTest)

	// non-packed -> packed
	np.A = []int32{0, 1, 1, 2, 3, 5}
	encodeDecode(t, np, p, "non-packed -> packed")
	if !reflect.DeepEqual(np.A, p.B) {
		t.Errorf("failed non-packed -> packed; np.A=%+v, p.B=%+v", np.A, p.B)
	}

	// packed -> non-packed
	np.Reset()
	p.B = []int32{3, 1, 4, 1, 5, 9}
	encodeDecode(t, p, np, "packed -> non-packed")
	if !reflect.DeepEqual(p.B, np.A) {
		t.Errorf("failed packed -> non-packed; p.B=%+v, np.A=%+v", p.B, np.A)
	}
}

func TestProto1RepeatedGroup(t *testing.T) {
	pb := &vendor.MessageList{
		Message: []*vendor.MessageList_Message{
			{
				Name:  vendor.String("blah"),
				Count: vendor.Int32(7),
			},
			// NOTE: pb.Message[1] is a nil
			nil,
		},
	}

	o := old()
	err := o.Marshal(pb)
	if err == nil || !strings.Contains(err.Error(), "repeated field Message has nil") {
		t.Fatalf("unexpected or no error when marshaling: %v", err)
	}
}

// Test that enums work.  Checks for a bug introduced by making enums
// named types instead of int32: newInt32FromUint64 would crash with
// a type mismatch in reflect.PointTo.
func TestEnum(t *testing.T) {
	pb := new(vendor.GoEnum)
	pb.Foo = vendor.FOO_FOO1.Enum()
	o := old()
	if err := o.Marshal(pb); err != nil {
		t.Fatal("error encoding enum:", err)
	}
	pb1 := new(vendor.GoEnum)
	if err := o.Unmarshal(pb1); err != nil {
		t.Fatal("error decoding enum:", err)
	}
	if *pb1.Foo != vendor.FOO_FOO1 {
		t.Error("expected 7 but got ", *pb1.Foo)
	}
}

// Enum types have String methods. Check that enum fields can be printed.
// We don't care what the value actually is, just as long as it doesn't crash.
func TestPrintingNilEnumFields(t *testing.T) {
	pb := new(vendor.GoEnum)
	_ = fmt.Sprintf("%+v", pb)
}

// Verify that absent required fields cause Marshal/Unmarshal to return errors.
func TestRequiredFieldEnforcement(t *testing.T) {
	pb := new(vendor.GoTestField)
	_, err := vendor.Marshal(pb)
	if err == nil {
		t.Error("marshal: expected error, got nil")
	} else if _, ok := err.(*vendor.RequiredNotSetError); !ok || !strings.Contains(err.Error(), "Label") {
		t.Errorf("marshal: bad error type: %v", err)
	}

	// A slightly sneaky, yet valid, proto. It encodes the same required field twice,
	// so simply counting the required fields is insufficient.
	// field 1, encoding 2, value "hi"
	buf := []byte("\x0A\x02hi\x0A\x02hi")
	err = vendor.Unmarshal(buf, pb)
	if err == nil {
		t.Error("unmarshal: expected error, got nil")
	} else if _, ok := err.(*vendor.RequiredNotSetError); !ok || !strings.Contains(err.Error(), "Type") && !strings.Contains(err.Error(), "{Unknown}") {
		// TODO: remove unknown cases once we commit to the new unmarshaler.
		t.Errorf("unmarshal: bad error type: %v", err)
	}
}

// Verify that absent required fields in groups cause Marshal/Unmarshal to return errors.
func TestRequiredFieldEnforcementGroups(t *testing.T) {
	pb := &vendor.GoTestRequiredGroupField{Group: &vendor.GoTestRequiredGroupField_Group{}}
	if _, err := vendor.Marshal(pb); err == nil {
		t.Error("marshal: expected error, got nil")
	} else if _, ok := err.(*vendor.RequiredNotSetError); !ok || !strings.Contains(err.Error(), "Group.Field") {
		t.Errorf("marshal: bad error type: %v", err)
	}

	buf := []byte{11, 12}
	if err := vendor.Unmarshal(buf, pb); err == nil {
		t.Error("unmarshal: expected error, got nil")
	} else if _, ok := err.(*vendor.RequiredNotSetError); !ok || !strings.Contains(err.Error(), "Group.Field") && !strings.Contains(err.Error(), "Group.{Unknown}") {
		t.Errorf("unmarshal: bad error type: %v", err)
	}
}

func TestTypedNilMarshal(t *testing.T) {
	// A typed nil should return ErrNil and not crash.
	{
		var m *vendor.GoEnum
		if _, err := vendor.Marshal(m); err != vendor.ErrNil {
			t.Errorf("Marshal(%#v): got %v, want ErrNil", m, err)
		}
	}

	{
		m := &vendor.Communique{Union: &vendor.Communique_Msg{nil}}
		if _, err := vendor.Marshal(m); err == nil || err == vendor.ErrNil {
			t.Errorf("Marshal(%#v): got %v, want errOneofHasNil", m, err)
		}
	}
}

// A type that implements the Marshaler interface, but is not nillable.
type nonNillableInt uint64

func (nni nonNillableInt) Marshal() ([]byte, error) {
	return vendor.EncodeVarint(uint64(nni)), nil
}

type NNIMessage struct {
	nni nonNillableInt
}

func (*NNIMessage) Reset()         {}
func (*NNIMessage) String() string { return "" }
func (*NNIMessage) ProtoMessage()  {}

type NMMessage struct{}

func (*NMMessage) Reset()         {}
func (*NMMessage) String() string { return "" }
func (*NMMessage) ProtoMessage()  {}

// Verify a type that uses the Marshaler interface, but has a nil pointer.
func TestNilMarshaler(t *testing.T) {
	// Try a struct with a Marshaler field that is nil.
	// It should be directly marshable.
	nmm := new(NMMessage)
	if _, err := vendor.Marshal(nmm); err != nil {
		t.Error("unexpected error marshaling nmm: ", err)
	}

	// Try a struct with a Marshaler field that is not nillable.
	nnim := new(NNIMessage)
	nnim.nni = 7
	var _ vendor.Marshaler = nnim.nni // verify it is truly a Marshaler
	if _, err := vendor.Marshal(nnim); err != nil {
		t.Error("unexpected error marshaling nnim: ", err)
	}
}

func TestAllSetDefaults(t *testing.T) {
	// Exercise SetDefaults with all scalar field types.
	m := &vendor.Defaults{
		// NaN != NaN, so override that here.
		F_Nan: vendor.Float32(1.7),
	}
	expected := &vendor.Defaults{
		F_Bool:    vendor.Bool(true),
		F_Int32:   vendor.Int32(32),
		F_Int64:   vendor.Int64(64),
		F_Fixed32: vendor.Uint32(320),
		F_Fixed64: vendor.Uint64(640),
		F_Uint32:  vendor.Uint32(3200),
		F_Uint64:  vendor.Uint64(6400),
		F_Float:   vendor.Float32(314159),
		F_Double:  vendor.Float64(271828),
		F_String:  vendor.String(`hello, "world!"` + "\n"),
		F_Bytes:   []byte("Bignose"),
		F_Sint32:  vendor.Int32(-32),
		F_Sint64:  vendor.Int64(-64),
		F_Enum:    vendor.Defaults_GREEN.Enum(),
		F_Pinf:    vendor.Float32(float32(math.Inf(1))),
		F_Ninf:    vendor.Float32(float32(math.Inf(-1))),
		F_Nan:     vendor.Float32(1.7),
		StrZero:   vendor.String(""),
	}
	vendor.SetDefaults(m)
	if !vendor.Equal(m, expected) {
		t.Errorf("SetDefaults failed\n got %v\nwant %v", m, expected)
	}
}

func TestSetDefaultsWithSetField(t *testing.T) {
	// Check that a set value is not overridden.
	m := &vendor.Defaults{
		F_Int32: vendor.Int32(12),
	}
	vendor.SetDefaults(m)
	if v := m.GetF_Int32(); v != 12 {
		t.Errorf("m.FInt32 = %v, want 12", v)
	}
}

func TestSetDefaultsWithSubMessage(t *testing.T) {
	m := &vendor.OtherMessage{
		Key: vendor.Int64(123),
		Inner: &vendor.InnerMessage{
			Host: vendor.String("gopher"),
		},
	}
	expected := &vendor.OtherMessage{
		Key: vendor.Int64(123),
		Inner: &vendor.InnerMessage{
			Host: vendor.String("gopher"),
			Port: vendor.Int32(4000),
		},
	}
	vendor.SetDefaults(m)
	if !vendor.Equal(m, expected) {
		t.Errorf("\n got %v\nwant %v", m, expected)
	}
}

func TestSetDefaultsWithRepeatedSubMessage(t *testing.T) {
	m := &vendor.MyMessage{
		RepInner: []*vendor.InnerMessage{{}},
	}
	expected := &vendor.MyMessage{
		RepInner: []*vendor.InnerMessage{{
			Port: vendor.Int32(4000),
		}},
	}
	vendor.SetDefaults(m)
	if !vendor.Equal(m, expected) {
		t.Errorf("\n got %v\nwant %v", m, expected)
	}
}

func TestSetDefaultWithRepeatedNonMessage(t *testing.T) {
	m := &vendor.MyMessage{
		Pet: []string{"turtle", "wombat"},
	}
	expected := vendor.Clone(m)
	vendor.SetDefaults(m)
	if !vendor.Equal(m, expected) {
		t.Errorf("\n got %v\nwant %v", m, expected)
	}
}

func TestMaximumTagNumber(t *testing.T) {
	m := &vendor.MaxTag{
		LastField: vendor.String("natural goat essence"),
	}
	buf, err := vendor.Marshal(m)
	if err != nil {
		t.Fatalf("proto.Marshal failed: %v", err)
	}
	m2 := new(vendor.MaxTag)
	if err := vendor.Unmarshal(buf, m2); err != nil {
		t.Fatalf("proto.Unmarshal failed: %v", err)
	}
	if got, want := m2.GetLastField(), *m.LastField; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestJSON(t *testing.T) {
	m := &vendor.MyMessage{
		Count: vendor.Int32(4),
		Pet:   []string{"bunny", "kitty"},
		Inner: &vendor.InnerMessage{
			Host: vendor.String("cauchy"),
		},
		Bikeshed: vendor.MyMessage_GREEN.Enum(),
	}
	const expected = `{"count":4,"pet":["bunny","kitty"],"inner":{"host":"cauchy"},"bikeshed":1}`

	b, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}
	s := string(b)
	if s != expected {
		t.Errorf("got  %s\nwant %s", s, expected)
	}

	received := new(vendor.MyMessage)
	if err := json.Unmarshal(b, received); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}
	if !vendor.Equal(received, m) {
		t.Fatalf("got %s, want %s", received, m)
	}

	// Test unmarshalling of JSON with symbolic enum name.
	const old = `{"count":4,"pet":["bunny","kitty"],"inner":{"host":"cauchy"},"bikeshed":"GREEN"}`
	received.Reset()
	if err := json.Unmarshal([]byte(old), received); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}
	if !vendor.Equal(received, m) {
		t.Fatalf("got %s, want %s", received, m)
	}
}

func TestBadWireType(t *testing.T) {
	b := []byte{7<<3 | 6} // field 7, wire type 6
	pb := new(vendor.OtherMessage)
	if err := vendor.Unmarshal(b, pb); err == nil {
		t.Errorf("Unmarshal did not fail")
	} else if !strings.Contains(err.Error(), "unknown wire type") {
		t.Errorf("wrong error: %v", err)
	}
}

func TestBytesWithInvalidLength(t *testing.T) {
	// If a byte sequence has an invalid (negative) length, Unmarshal should not panic.
	b := []byte{2<<3 | vendor.WireBytes, 0xff, 0xff, 0xff, 0xff, 0xff, 0}
	vendor.Unmarshal(b, new(vendor.MyMessage))
}

func TestLengthOverflow(t *testing.T) {
	// Overflowing a length should not panic.
	b := []byte{2<<3 | vendor.WireBytes, 1, 1, 3<<3 | vendor.WireBytes, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f, 0x01}
	vendor.Unmarshal(b, new(vendor.MyMessage))
}

func TestVarintOverflow(t *testing.T) {
	// Overflowing a 64-bit length should not be allowed.
	b := []byte{1<<3 | vendor.WireVarint, 0x01, 3<<3 | vendor.WireBytes, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x01}
	if err := vendor.Unmarshal(b, new(vendor.MyMessage)); err == nil {
		t.Fatalf("Overflowed uint64 length without error")
	}
}

func TestBytesWithInvalidLengthInGroup(t *testing.T) {
	// Overflowing a 64-bit length should not be allowed.
	b := []byte{0xbb, 0x30, 0xb2, 0x30, 0xb0, 0xb2, 0x83, 0xf1, 0xb0, 0xb2, 0xef, 0xbf, 0xbd, 0x01}
	if err := vendor.Unmarshal(b, new(vendor.MyMessage)); err == nil {
		t.Fatalf("Overflowed uint64 length without error")
	}
}

func TestUnmarshalFuzz(t *testing.T) {
	const N = 1000
	seed := time.Now().UnixNano()
	t.Logf("RNG seed is %d", seed)
	rng := rand.New(rand.NewSource(seed))
	buf := make([]byte, 20)
	for i := 0; i < N; i++ {
		for j := range buf {
			buf[j] = byte(rng.Intn(256))
		}
		fuzzUnmarshal(t, buf)
	}
}

func TestMergeMessages(t *testing.T) {
	pb := &vendor.MessageList{Message: []*vendor.MessageList_Message{{Name: vendor.String("x"), Count: vendor.Int32(1)}}}
	data, err := vendor.Marshal(pb)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	pb1 := new(vendor.MessageList)
	if err := vendor.Unmarshal(data, pb1); err != nil {
		t.Fatalf("first Unmarshal: %v", err)
	}
	if err := vendor.Unmarshal(data, pb1); err != nil {
		t.Fatalf("second Unmarshal: %v", err)
	}
	if len(pb1.Message) != 1 {
		t.Errorf("two Unmarshals produced %d Messages, want 1", len(pb1.Message))
	}

	pb2 := new(vendor.MessageList)
	if err := vendor.UnmarshalMerge(data, pb2); err != nil {
		t.Fatalf("first UnmarshalMerge: %v", err)
	}
	if err := vendor.UnmarshalMerge(data, pb2); err != nil {
		t.Fatalf("second UnmarshalMerge: %v", err)
	}
	if len(pb2.Message) != 2 {
		t.Errorf("two UnmarshalMerges produced %d Messages, want 2", len(pb2.Message))
	}
}

func TestExtensionMarshalOrder(t *testing.T) {
	m := &vendor.MyMessage{Count: vendor.Int(123)}
	if err := vendor.SetExtension(m, vendor.E_Ext_More, &vendor.Ext{Data: vendor.String("alpha")}); err != nil {
		t.Fatalf("SetExtension: %v", err)
	}
	if err := vendor.SetExtension(m, vendor.E_Ext_Text, vendor.String("aleph")); err != nil {
		t.Fatalf("SetExtension: %v", err)
	}
	if err := vendor.SetExtension(m, vendor.E_Ext_Number, vendor.Int32(1)); err != nil {
		t.Fatalf("SetExtension: %v", err)
	}

	// Serialize m several times, and check we get the same bytes each time.
	var orig []byte
	for i := 0; i < 100; i++ {
		b, err := vendor.Marshal(m)
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		if i == 0 {
			orig = b
			continue
		}
		if !bytes.Equal(b, orig) {
			t.Errorf("Bytes differ on attempt #%d", i)
		}
	}
}

func TestExtensionMapFieldMarshalDeterministic(t *testing.T) {
	m := &vendor.MyMessage{Count: vendor.Int(123)}
	if err := vendor.SetExtension(m, vendor.E_Ext_More, &vendor.Ext{MapField: map[int32]int32{1: 1, 2: 2, 3: 3, 4: 4}}); err != nil {
		t.Fatalf("SetExtension: %v", err)
	}
	marshal := func(m vendor.Message) []byte {
		var b vendor.Buffer
		b.SetDeterministic(true)
		if err := b.Marshal(m); err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}
		return b.Bytes()
	}

	want := marshal(m)
	for i := 0; i < 100; i++ {
		if got := marshal(m); !bytes.Equal(got, want) {
			t.Errorf("Marshal produced inconsistent output with determinism enabled (pass %d).\n got %v\nwant %v", i, got, want)
		}
	}
}

// Many extensions, because small maps might not iterate differently on each iteration.
var exts = []*vendor.ExtensionDesc{
	vendor.E_X201,
	vendor.E_X202,
	vendor.E_X203,
	vendor.E_X204,
	vendor.E_X205,
	vendor.E_X206,
	vendor.E_X207,
	vendor.E_X208,
	vendor.E_X209,
	vendor.E_X210,
	vendor.E_X211,
	vendor.E_X212,
	vendor.E_X213,
	vendor.E_X214,
	vendor.E_X215,
	vendor.E_X216,
	vendor.E_X217,
	vendor.E_X218,
	vendor.E_X219,
	vendor.E_X220,
	vendor.E_X221,
	vendor.E_X222,
	vendor.E_X223,
	vendor.E_X224,
	vendor.E_X225,
	vendor.E_X226,
	vendor.E_X227,
	vendor.E_X228,
	vendor.E_X229,
	vendor.E_X230,
	vendor.E_X231,
	vendor.E_X232,
	vendor.E_X233,
	vendor.E_X234,
	vendor.E_X235,
	vendor.E_X236,
	vendor.E_X237,
	vendor.E_X238,
	vendor.E_X239,
	vendor.E_X240,
	vendor.E_X241,
	vendor.E_X242,
	vendor.E_X243,
	vendor.E_X244,
	vendor.E_X245,
	vendor.E_X246,
	vendor.E_X247,
	vendor.E_X248,
	vendor.E_X249,
	vendor.E_X250,
}

func TestMessageSetMarshalOrder(t *testing.T) {
	m := &vendor.MyMessageSet{}
	for _, x := range exts {
		if err := vendor.SetExtension(m, x, &vendor.Empty{}); err != nil {
			t.Fatalf("SetExtension: %v", err)
		}
	}

	buf, err := vendor.Marshal(m)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	// Serialize m several times, and check we get the same bytes each time.
	for i := 0; i < 10; i++ {
		b1, err := vendor.Marshal(m)
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		if !bytes.Equal(b1, buf) {
			t.Errorf("Bytes differ on re-Marshal #%d", i)
		}

		m2 := &vendor.MyMessageSet{}
		if err := vendor.Unmarshal(buf, m2); err != nil {
			t.Errorf("Unmarshal: %v", err)
		}
		b2, err := vendor.Marshal(m2)
		if err != nil {
			t.Errorf("re-Marshal: %v", err)
		}
		if !bytes.Equal(b2, buf) {
			t.Errorf("Bytes differ on round-trip #%d", i)
		}
	}
}

func TestUnmarshalMergesMessages(t *testing.T) {
	// If a nested message occurs twice in the input,
	// the fields should be merged when decoding.
	a := &vendor.OtherMessage{
		Key: vendor.Int64(123),
		Inner: &vendor.InnerMessage{
			Host: vendor.String("polhode"),
			Port: vendor.Int32(1234),
		},
	}
	aData, err := vendor.Marshal(a)
	if err != nil {
		t.Fatalf("Marshal(a): %v", err)
	}
	b := &vendor.OtherMessage{
		Weight: vendor.Float32(1.2),
		Inner: &vendor.InnerMessage{
			Host:      vendor.String("herpolhode"),
			Connected: vendor.Bool(true),
		},
	}
	bData, err := vendor.Marshal(b)
	if err != nil {
		t.Fatalf("Marshal(b): %v", err)
	}
	want := &vendor.OtherMessage{
		Key:    vendor.Int64(123),
		Weight: vendor.Float32(1.2),
		Inner: &vendor.InnerMessage{
			Host:      vendor.String("herpolhode"),
			Port:      vendor.Int32(1234),
			Connected: vendor.Bool(true),
		},
	}
	got := new(vendor.OtherMessage)
	if err := vendor.Unmarshal(append(aData, bData...), got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !vendor.Equal(got, want) {
		t.Errorf("\n got %v\nwant %v", got, want)
	}
}

func TestUnmarshalMergesGroups(t *testing.T) {
	// If a nested group occurs twice in the input,
	// the fields should be merged when decoding.
	a := &vendor.GroupNew{
		G: &vendor.GroupNew_G{
			X: vendor.Int32(7),
			Y: vendor.Int32(8),
		},
	}
	aData, err := vendor.Marshal(a)
	if err != nil {
		t.Fatalf("Marshal(a): %v", err)
	}
	b := &vendor.GroupNew{
		G: &vendor.GroupNew_G{
			X: vendor.Int32(9),
		},
	}
	bData, err := vendor.Marshal(b)
	if err != nil {
		t.Fatalf("Marshal(b): %v", err)
	}
	want := &vendor.GroupNew{
		G: &vendor.GroupNew_G{
			X: vendor.Int32(9),
			Y: vendor.Int32(8),
		},
	}
	got := new(vendor.GroupNew)
	if err := vendor.Unmarshal(append(aData, bData...), got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !vendor.Equal(got, want) {
		t.Errorf("\n got %v\nwant %v", got, want)
	}
}

func TestEncodingSizes(t *testing.T) {
	tests := []struct {
		m vendor.Message
		n int
	}{
		{&vendor.Defaults{F_Int32: vendor.Int32(math.MaxInt32)}, 6},
		{&vendor.Defaults{F_Int32: vendor.Int32(math.MinInt32)}, 11},
		{&vendor.Defaults{F_Uint32: vendor.Uint32(uint32(math.MaxInt32) + 1)}, 6},
		{&vendor.Defaults{F_Uint32: vendor.Uint32(math.MaxUint32)}, 6},
	}
	for _, test := range tests {
		b, err := vendor.Marshal(test.m)
		if err != nil {
			t.Errorf("Marshal(%v): %v", test.m, err)
			continue
		}
		if len(b) != test.n {
			t.Errorf("Marshal(%v) yielded %d bytes, want %d bytes", test.m, len(b), test.n)
		}
	}
}

func TestRequiredNotSetError(t *testing.T) {
	pb := initGoTest(false)
	pb.RequiredField.Label = nil
	pb.F_Int32Required = nil
	pb.F_Int64Required = nil

	expected := "0807" + // field 1, encoding 0, value 7
		"2206" + "120474797065" + // field 4, encoding 2 (GoTestField)
		"5001" + // field 10, encoding 0, value 1
		"6d20000000" + // field 13, encoding 5, value 0x20
		"714000000000000000" + // field 14, encoding 1, value 0x40
		"78a019" + // field 15, encoding 0, value 0xca0 = 3232
		"8001c032" + // field 16, encoding 0, value 0x1940 = 6464
		"8d0100004a45" + // field 17, encoding 5, value 3232.0
		"9101000000000040b940" + // field 18, encoding 1, value 6464.0
		"9a0106" + "737472696e67" + // field 19, encoding 2, string "string"
		"b304" + // field 70, encoding 3, start group
		"ba0408" + "7265717569726564" + // field 71, encoding 2, string "required"
		"b404" + // field 70, encoding 4, end group
		"aa0605" + "6279746573" + // field 101, encoding 2, string "bytes"
		"b0063f" + // field 102, encoding 0, 0x3f zigzag32
		"b8067f" + // field 103, encoding 0, 0x7f zigzag64
		"c506e0ffffff" + // field 104, encoding 5, -32 fixed32
		"c906c0ffffffffffffff" // field 105, encoding 1, -64 fixed64

	o := old()
	bytes, err := vendor.Marshal(pb)
	if _, ok := err.(*vendor.RequiredNotSetError); !ok {
		fmt.Printf("marshal-1 err = %v, want *RequiredNotSetError", err)
		o.DebugPrint("", bytes)
		t.Fatalf("expected = %s", expected)
	}
	if !strings.Contains(err.Error(), "RequiredField.Label") {
		t.Errorf("marshal-1 wrong err msg: %v", err)
	}
	if !equal(bytes, expected, t) {
		o.DebugPrint("neq 1", bytes)
		t.Fatalf("expected = %s", expected)
	}

	// Now test Unmarshal by recreating the original buffer.
	pbd := new(vendor.GoTest)
	err = vendor.Unmarshal(bytes, pbd)
	if _, ok := err.(*vendor.RequiredNotSetError); !ok {
		t.Fatalf("unmarshal err = %v, want *RequiredNotSetError", err)
		o.DebugPrint("", bytes)
		t.Fatalf("string = %s", expected)
	}
	if !strings.Contains(err.Error(), "RequiredField.Label") && !strings.Contains(err.Error(), "RequiredField.{Unknown}") {
		t.Errorf("unmarshal wrong err msg: %v", err)
	}
	bytes, err = vendor.Marshal(pbd)
	if _, ok := err.(*vendor.RequiredNotSetError); !ok {
		t.Errorf("marshal-2 err = %v, want *RequiredNotSetError", err)
		o.DebugPrint("", bytes)
		t.Fatalf("string = %s", expected)
	}
	if !strings.Contains(err.Error(), "RequiredField.Label") {
		t.Errorf("marshal-2 wrong err msg: %v", err)
	}
	if !equal(bytes, expected, t) {
		o.DebugPrint("neq 2", bytes)
		t.Fatalf("string = %s", expected)
	}
}

func TestRequiredNotSetErrorWithBadWireTypes(t *testing.T) {
	// Required field expects a varint, and properly found a varint.
	if err := vendor.Unmarshal([]byte{0x08, 0x00}, new(vendor.GoEnum)); err != nil {
		t.Errorf("Unmarshal = %v, want nil", err)
	}
	// Required field expects a varint, but found a fixed32 instead.
	if err := vendor.Unmarshal([]byte{0x0d, 0x00, 0x00, 0x00, 0x00}, new(vendor.GoEnum)); err == nil {
		t.Errorf("Unmarshal = nil, want RequiredNotSetError")
	}
	// Required field expects a varint, and found both a varint and fixed32 (ignored).
	m := new(vendor.GoEnum)
	if err := vendor.Unmarshal([]byte{0x08, 0x00, 0x0d, 0x00, 0x00, 0x00, 0x00}, m); err != nil {
		t.Errorf("Unmarshal = %v, want nil", err)
	}
	if !bytes.Equal(m.XXX_unrecognized, []byte{0x0d, 0x00, 0x00, 0x00, 0x00}) {
		t.Errorf("expected fixed32 to appear as unknown bytes: %x", m.XXX_unrecognized)
	}
}

func fuzzUnmarshal(t *testing.T, data []byte) {
	defer func() {
		if e := recover(); e != nil {
			t.Errorf("These bytes caused a panic: %+v", data)
			t.Logf("Stack:\n%s", debug.Stack())
			t.FailNow()
		}
	}()

	pb := new(vendor.MyMessage)
	vendor.Unmarshal(data, pb)
}

func TestMapFieldMarshal(t *testing.T) {
	m := &vendor.MessageWithMap{
		NameMapping: map[int32]string{
			1: "Rob",
			4: "Ian",
			8: "Dave",
		},
	}
	b, err := vendor.Marshal(m)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	// b should be the concatenation of these three byte sequences in some order.
	parts := []string{
		"\n\a\b\x01\x12\x03Rob",
		"\n\a\b\x04\x12\x03Ian",
		"\n\b\b\x08\x12\x04Dave",
	}
	ok := false
	for i := range parts {
		for j := range parts {
			if j == i {
				continue
			}
			for k := range parts {
				if k == i || k == j {
					continue
				}
				try := parts[i] + parts[j] + parts[k]
				if bytes.Equal(b, []byte(try)) {
					ok = true
					break
				}
			}
		}
	}
	if !ok {
		t.Fatalf("Incorrect Marshal output.\n got %q\nwant %q (or a permutation of that)", b, parts[0]+parts[1]+parts[2])
	}
	t.Logf("FYI b: %q", b)

	(new(vendor.Buffer)).DebugPrint("Dump of b", b)
}

func TestMapFieldDeterministicMarshal(t *testing.T) {
	m := &vendor.MessageWithMap{
		NameMapping: map[int32]string{
			1: "Rob",
			4: "Ian",
			8: "Dave",
		},
	}

	marshal := func(m vendor.Message) []byte {
		var b vendor.Buffer
		b.SetDeterministic(true)
		if err := b.Marshal(m); err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}
		return b.Bytes()
	}

	want := marshal(m)
	for i := 0; i < 10; i++ {
		if got := marshal(m); !bytes.Equal(got, want) {
			t.Errorf("Marshal produced inconsistent output with determinism enabled (pass %d).\n got %v\nwant %v", i, got, want)
		}
	}
}

func TestMapFieldRoundTrips(t *testing.T) {
	m := &vendor.MessageWithMap{
		NameMapping: map[int32]string{
			1: "Rob",
			4: "Ian",
			8: "Dave",
		},
		MsgMapping: map[int64]*vendor.FloatingPoint{
			0x7001: {F: vendor.Float64(2.0)},
		},
		ByteMapping: map[bool][]byte{
			false: []byte("that's not right!"),
			true:  []byte("aye, 'tis true!"),
		},
	}
	b, err := vendor.Marshal(m)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	t.Logf("FYI b: %q", b)
	m2 := new(vendor.MessageWithMap)
	if err := vendor.Unmarshal(b, m2); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !vendor.Equal(m, m2) {
		t.Errorf("Map did not survive a round trip.\ninitial: %v\n  final: %v", m, m2)
	}
}

func TestMapFieldWithNil(t *testing.T) {
	m1 := &vendor.MessageWithMap{
		MsgMapping: map[int64]*vendor.FloatingPoint{
			1: nil,
		},
	}
	b, err := vendor.Marshal(m1)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	m2 := new(vendor.MessageWithMap)
	if err := vendor.Unmarshal(b, m2); err != nil {
		t.Fatalf("Unmarshal: %v, got these bytes: %v", err, b)
	}
	if v, ok := m2.MsgMapping[1]; !ok {
		t.Error("msg_mapping[1] not present")
	} else if v != nil {
		t.Errorf("msg_mapping[1] not nil: %v", v)
	}
}

func TestMapFieldWithNilBytes(t *testing.T) {
	m1 := &vendor.MessageWithMap{
		ByteMapping: map[bool][]byte{
			false: {},
			true:  nil,
		},
	}
	n := vendor.Size(m1)
	b, err := vendor.Marshal(m1)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if n != len(b) {
		t.Errorf("Size(m1) = %d; want len(Marshal(m1)) = %d", n, len(b))
	}
	m2 := new(vendor.MessageWithMap)
	if err := vendor.Unmarshal(b, m2); err != nil {
		t.Fatalf("Unmarshal: %v, got these bytes: %v", err, b)
	}
	if v, ok := m2.ByteMapping[false]; !ok {
		t.Error("byte_mapping[false] not present")
	} else if len(v) != 0 {
		t.Errorf("byte_mapping[false] not empty: %#v", v)
	}
	if v, ok := m2.ByteMapping[true]; !ok {
		t.Error("byte_mapping[true] not present")
	} else if len(v) != 0 {
		t.Errorf("byte_mapping[true] not empty: %#v", v)
	}
}

func TestDecodeMapFieldMissingKey(t *testing.T) {
	b := []byte{
		0x0A, 0x03, // message, tag 1 (name_mapping), of length 3 bytes
		// no key
		0x12, 0x01, 0x6D, // string value of length 1 byte, value "m"
	}
	got := &vendor.MessageWithMap{}
	err := vendor.Unmarshal(b, got)
	if err != nil {
		t.Fatalf("failed to marshal map with missing key: %v", err)
	}
	want := &vendor.MessageWithMap{NameMapping: map[int32]string{0: "m"}}
	if !vendor.Equal(got, want) {
		t.Errorf("Unmarshaled map with no key was not as expected. got: %v, want %v", got, want)
	}
}

func TestDecodeMapFieldMissingValue(t *testing.T) {
	b := []byte{
		0x0A, 0x02, // message, tag 1 (name_mapping), of length 2 bytes
		0x08, 0x01, // varint key, value 1
		// no value
	}
	got := &vendor.MessageWithMap{}
	err := vendor.Unmarshal(b, got)
	if err != nil {
		t.Fatalf("failed to marshal map with missing value: %v", err)
	}
	want := &vendor.MessageWithMap{NameMapping: map[int32]string{1: ""}}
	if !vendor.Equal(got, want) {
		t.Errorf("Unmarshaled map with no value was not as expected. got: %v, want %v", got, want)
	}
}

func TestOneof(t *testing.T) {
	m := &vendor.Communique{}
	b, err := vendor.Marshal(m)
	if err != nil {
		t.Fatalf("Marshal of empty message with oneof: %v", err)
	}
	if len(b) != 0 {
		t.Errorf("Marshal of empty message yielded too many bytes: %v", b)
	}

	m = &vendor.Communique{
		Union: &vendor.Communique_Name{"Barry"},
	}

	// Round-trip.
	b, err = vendor.Marshal(m)
	if err != nil {
		t.Fatalf("Marshal of message with oneof: %v", err)
	}
	if len(b) != 7 { // name tag/wire (1) + name len (1) + name (5)
		t.Errorf("Incorrect marshal of message with oneof: %v", b)
	}
	m.Reset()
	if err := vendor.Unmarshal(b, m); err != nil {
		t.Fatalf("Unmarshal of message with oneof: %v", err)
	}
	if x, ok := m.Union.(*vendor.Communique_Name); !ok || x.Name != "Barry" {
		t.Errorf("After round trip, Union = %+v", m.Union)
	}
	if name := m.GetName(); name != "Barry" {
		t.Errorf("After round trip, GetName = %q, want %q", name, "Barry")
	}

	// Let's try with a message in the oneof.
	m.Union = &vendor.Communique_Msg{&vendor.Strings{StringField: vendor.String("deep deep string")}}
	b, err = vendor.Marshal(m)
	if err != nil {
		t.Fatalf("Marshal of message with oneof set to message: %v", err)
	}
	if len(b) != 20 { // msg tag/wire (1) + msg len (1) + msg (1 + 1 + 16)
		t.Errorf("Incorrect marshal of message with oneof set to message: %v", b)
	}
	m.Reset()
	if err := vendor.Unmarshal(b, m); err != nil {
		t.Fatalf("Unmarshal of message with oneof set to message: %v", err)
	}
	ss, ok := m.Union.(*vendor.Communique_Msg)
	if !ok || ss.Msg.GetStringField() != "deep deep string" {
		t.Errorf("After round trip with oneof set to message, Union = %+v", m.Union)
	}
}

func TestOneofNilBytes(t *testing.T) {
	// A oneof with nil byte slice should marshal to tag + 0 (size), with no error.
	m := &vendor.Communique{Union: &vendor.Communique_Data{Data: nil}}
	b, err := vendor.Marshal(m)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	want := []byte{
		7<<3 | 2, // tag 7, wire type 2
		0,        // size
	}
	if !bytes.Equal(b, want) {
		t.Errorf("Wrong result of Marshal: got %x, want %x", b, want)
	}
}

func TestInefficientPackedBool(t *testing.T) {
	// https://github.com/golang/protobuf/issues/76
	inp := []byte{
		0x12, 0x02, // 0x12 = 2<<3|2; 2 bytes
		// Usually a bool should take a single byte,
		// but it is permitted to be any varint.
		0xb9, 0x30,
	}
	if err := vendor.Unmarshal(inp, new(vendor.MoreRepeated)); err != nil {
		t.Error(err)
	}
}

// Make sure pure-reflect-based implementation handles
// []int32-[]enum conversion correctly.
func TestRepeatedEnum2(t *testing.T) {
	pb := &vendor.RepeatedEnum{
		Color: []vendor.RepeatedEnum_Color{vendor.RepeatedEnum_RED},
	}
	b, err := vendor.Marshal(pb)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	x := new(vendor.RepeatedEnum)
	err = vendor.Unmarshal(b, x)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if !vendor.Equal(pb, x) {
		t.Errorf("Incorrect result: want: %v got: %v", pb, x)
	}
}

// TestConcurrentMarshal makes sure that it is safe to marshal
// same message in multiple goroutines concurrently.
func TestConcurrentMarshal(t *testing.T) {
	pb := initGoTest(true)
	const N = 100
	b := make([][]byte, N)

	var wg sync.WaitGroup
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			var err error
			b[i], err = vendor.Marshal(pb)
			if err != nil {
				t.Errorf("marshal error: %v", err)
			}
		}(i)
	}

	wg.Wait()
	for i := 1; i < N; i++ {
		if !bytes.Equal(b[0], b[i]) {
			t.Errorf("concurrent marshal result not same: b[0] = %v, b[%d] = %v", b[0], i, b[i])
		}
	}
}

func TestInvalidUTF8(t *testing.T) {
	const invalidUTF8 = "\xde\xad\xbe\xef\x80\x00\xff"
	tests := []struct {
		label  string
		proto2 vendor.Message
		proto3 vendor.Message
		want   []byte
	}{{
		label:  "Scalar",
		proto2: &vendor.TestUTF8{Scalar: vendor.String(invalidUTF8)},
		proto3: &vendor.TestUTF8{Scalar: invalidUTF8},
		want:   []byte{0x0a, 0x07, 0xde, 0xad, 0xbe, 0xef, 0x80, 0x00, 0xff},
	}, {
		label:  "Vector",
		proto2: &vendor.TestUTF8{Vector: []string{invalidUTF8}},
		proto3: &vendor.TestUTF8{Vector: []string{invalidUTF8}},
		want:   []byte{0x12, 0x07, 0xde, 0xad, 0xbe, 0xef, 0x80, 0x00, 0xff},
	}, {
		label:  "Oneof",
		proto2: &vendor.TestUTF8{Oneof: &vendor.TestUTF8_Field{invalidUTF8}},
		proto3: &vendor.TestUTF8{Oneof: &vendor.TestUTF8_Field{invalidUTF8}},
		want:   []byte{0x1a, 0x07, 0xde, 0xad, 0xbe, 0xef, 0x80, 0x00, 0xff},
	}, {
		label:  "MapKey",
		proto2: &vendor.TestUTF8{MapKey: map[string]int64{invalidUTF8: 0}},
		proto3: &vendor.TestUTF8{MapKey: map[string]int64{invalidUTF8: 0}},
		want:   []byte{0x22, 0x0b, 0x0a, 0x07, 0xde, 0xad, 0xbe, 0xef, 0x80, 0x00, 0xff, 0x10, 0x00},
	}, {
		label:  "MapValue",
		proto2: &vendor.TestUTF8{MapValue: map[int64]string{0: invalidUTF8}},
		proto3: &vendor.TestUTF8{MapValue: map[int64]string{0: invalidUTF8}},
		want:   []byte{0x2a, 0x0b, 0x08, 0x00, 0x12, 0x07, 0xde, 0xad, 0xbe, 0xef, 0x80, 0x00, 0xff},
	}}

	for _, tt := range tests {
		// Proto2 should not validate UTF-8.
		b, err := vendor.Marshal(tt.proto2)
		if err != nil {
			t.Errorf("Marshal(proto2.%s) = %v, want nil", tt.label, err)
		}
		if !bytes.Equal(b, tt.want) {
			t.Errorf("Marshal(proto2.%s) = %x, want %x", tt.label, b, tt.want)
		}

		m := vendor.Clone(tt.proto2)
		m.Reset()
		if err = vendor.Unmarshal(tt.want, m); err != nil {
			t.Errorf("Unmarshal(proto2.%s) = %v, want nil", tt.label, err)
		}
		if !vendor.Equal(m, tt.proto2) {
			t.Errorf("proto2.%s: output mismatch:\ngot  %v\nwant %v", tt.label, m, tt.proto2)
		}

		// Proto3 should validate UTF-8.
		b, err = vendor.Marshal(tt.proto3)
		if err == nil {
			t.Errorf("Marshal(proto3.%s) = %v, want non-nil", tt.label, err)
		}
		if !bytes.Equal(b, tt.want) {
			t.Errorf("Marshal(proto3.%s) = %x, want %x", tt.label, b, tt.want)
		}

		m = vendor.Clone(tt.proto3)
		m.Reset()
		err = vendor.Unmarshal(tt.want, m)
		if err == nil {
			t.Errorf("Unmarshal(proto3.%s) = %v, want non-nil", tt.label, err)
		}
		if !vendor.Equal(m, tt.proto3) {
			t.Errorf("proto3.%s: output mismatch:\ngot  %v\nwant %v", tt.label, m, tt.proto2)
		}
	}
}

func TestRequired(t *testing.T) {
	// The F_BoolRequired field appears after all of the required fields.
	// It should still be handled even after multiple required field violations.
	m := &vendor.GoTest{F_BoolRequired: vendor.Bool(true)}
	got, err := vendor.Marshal(m)
	if _, ok := err.(*vendor.RequiredNotSetError); !ok {
		t.Errorf("Marshal() = %v, want RequiredNotSetError error", err)
	}
	if want := []byte{0x50, 0x01}; !bytes.Equal(got, want) {
		t.Errorf("Marshal() = %x, want %x", got, want)
	}

	m = new(vendor.GoTest)
	err = vendor.Unmarshal(got, m)
	if _, ok := err.(*vendor.RequiredNotSetError); !ok {
		t.Errorf("Marshal() = %v, want RequiredNotSetError error", err)
	}
	if !m.GetF_BoolRequired() {
		t.Error("m.F_BoolRequired = false, want true")
	}
}

// Benchmarks

func testMsg() *vendor.GoTest {
	pb := initGoTest(true)
	const N = 1000 // Internally the library starts much smaller.
	pb.F_Int32Repeated = make([]int32, N)
	pb.F_DoubleRepeated = make([]float64, N)
	for i := 0; i < N; i++ {
		pb.F_Int32Repeated[i] = int32(i)
		pb.F_DoubleRepeated[i] = float64(i)
	}
	return pb
}

func bytesMsg() *vendor.GoTest {
	pb := initGoTest(true)
	buf := make([]byte, 4000)
	for i := range buf {
		buf[i] = byte(i)
	}
	pb.F_BytesDefaulted = buf
	return pb
}

func benchmarkMarshal(b *testing.B, pb vendor.Message, marshal func(vendor.Message) ([]byte, error)) {
	d, _ := marshal(pb)
	b.SetBytes(int64(len(d)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		marshal(pb)
	}
}

func benchmarkBufferMarshal(b *testing.B, pb vendor.Message) {
	p := vendor.NewBuffer(nil)
	benchmarkMarshal(b, pb, func(pb0 vendor.Message) ([]byte, error) {
		p.Reset()
		err := p.Marshal(pb0)
		return p.Bytes(), err
	})
}

func benchmarkSize(b *testing.B, pb vendor.Message) {
	benchmarkMarshal(b, pb, func(pb0 vendor.Message) ([]byte, error) {
		vendor.Size(pb)
		return nil, nil
	})
}

func newOf(pb vendor.Message) vendor.Message {
	in := reflect.ValueOf(pb)
	if in.IsNil() {
		return pb
	}
	return reflect.New(in.Type().Elem()).Interface().(vendor.Message)
}

func benchmarkUnmarshal(b *testing.B, pb vendor.Message, unmarshal func([]byte, vendor.Message) error) {
	d, _ := vendor.Marshal(pb)
	b.SetBytes(int64(len(d)))
	pbd := newOf(pb)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		unmarshal(d, pbd)
	}
}

func benchmarkBufferUnmarshal(b *testing.B, pb vendor.Message) {
	p := vendor.NewBuffer(nil)
	benchmarkUnmarshal(b, pb, func(d []byte, pb0 vendor.Message) error {
		p.SetBuf(d)
		return p.Unmarshal(pb0)
	})
}

// Benchmark{Marshal,BufferMarshal,Size,Unmarshal,BufferUnmarshal}{,Bytes}

func BenchmarkMarshal(b *testing.B) {
	benchmarkMarshal(b, testMsg(), vendor.Marshal)
}

func BenchmarkBufferMarshal(b *testing.B) {
	benchmarkBufferMarshal(b, testMsg())
}

func BenchmarkSize(b *testing.B) {
	benchmarkSize(b, testMsg())
}

func BenchmarkUnmarshal(b *testing.B) {
	benchmarkUnmarshal(b, testMsg(), vendor.Unmarshal)
}

func BenchmarkBufferUnmarshal(b *testing.B) {
	benchmarkBufferUnmarshal(b, testMsg())
}

func BenchmarkMarshalBytes(b *testing.B) {
	benchmarkMarshal(b, bytesMsg(), vendor.Marshal)
}

func BenchmarkBufferMarshalBytes(b *testing.B) {
	benchmarkBufferMarshal(b, bytesMsg())
}

func BenchmarkSizeBytes(b *testing.B) {
	benchmarkSize(b, bytesMsg())
}

func BenchmarkUnmarshalBytes(b *testing.B) {
	benchmarkUnmarshal(b, bytesMsg(), vendor.Unmarshal)
}

func BenchmarkBufferUnmarshalBytes(b *testing.B) {
	benchmarkBufferUnmarshal(b, bytesMsg())
}

func BenchmarkUnmarshalUnrecognizedFields(b *testing.B) {
	b.StopTimer()
	pb := initGoTestField()
	skip := &vendor.GoSkipTest{
		SkipInt32:   vendor.Int32(32),
		SkipFixed32: vendor.Uint32(3232),
		SkipFixed64: vendor.Uint64(6464),
		SkipString:  vendor.String("skipper"),
		Skipgroup: &vendor.GoSkipTest_SkipGroup{
			GroupInt32:  vendor.Int32(75),
			GroupString: vendor.String("wxyz"),
		},
	}

	pbd := new(vendor.GoTestField)
	p := vendor.NewBuffer(nil)
	p.Marshal(pb)
	p.Marshal(skip)
	p2 := vendor.NewBuffer(nil)

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		p2.SetBuf(p.Bytes())
		p2.Unmarshal(pbd)
	}
}
