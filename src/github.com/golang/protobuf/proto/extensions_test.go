// Go support for Protocol Buffers - Google's data interchange format
//
// Copyright 2014 The Go Authors.  All rights reserved.
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
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"
	"vendor"
)

func TestGetExtensionsWithMissingExtensions(t *testing.T) {
	msg := &vendor.MyMessage{}
	ext1 := &vendor.Ext{}
	if err := vendor.SetExtension(msg, vendor.E_Ext_More, ext1); err != nil {
		t.Fatalf("Could not set ext1: %s", err)
	}
	exts, err := vendor.GetExtensions(msg, []*vendor.ExtensionDesc{
		vendor.E_Ext_More,
		vendor.E_Ext_Text,
	})
	if err != nil {
		t.Fatalf("GetExtensions() failed: %s", err)
	}
	if exts[0] != ext1 {
		t.Errorf("ext1 not in returned extensions: %T %v", exts[0], exts[0])
	}
	if exts[1] != nil {
		t.Errorf("ext2 in returned extensions: %T %v", exts[1], exts[1])
	}
}

func TestGetExtensionWithEmptyBuffer(t *testing.T) {
	// Make sure that GetExtension returns an error if its
	// undecoded buffer is empty.
	msg := &vendor.MyMessage{}
	vendor.SetRawExtension(msg, vendor.E_Ext_More.Field, []byte{})
	_, err := vendor.GetExtension(msg, vendor.E_Ext_More)
	if want := io.ErrUnexpectedEOF; err != want {
		t.Errorf("unexpected error in GetExtension from empty buffer: got %v, want %v", err, want)
	}
}

func TestGetExtensionForIncompleteDesc(t *testing.T) {
	msg := &vendor.MyMessage{Count: vendor.Int32(0)}
	extdesc1 := &vendor.ExtensionDesc{
		ExtendedType:  (*vendor.MyMessage)(nil),
		ExtensionType: (*bool)(nil),
		Field:         123456789,
		Name:          "a.b",
		Tag:           "varint,123456789,opt",
	}
	ext1 := vendor.Bool(true)
	if err := vendor.SetExtension(msg, extdesc1, ext1); err != nil {
		t.Fatalf("Could not set ext1: %s", err)
	}
	extdesc2 := &vendor.ExtensionDesc{
		ExtendedType:  (*vendor.MyMessage)(nil),
		ExtensionType: ([]byte)(nil),
		Field:         123456790,
		Name:          "a.c",
		Tag:           "bytes,123456790,opt",
	}
	ext2 := []byte{0, 1, 2, 3, 4, 5, 6, 7}
	if err := vendor.SetExtension(msg, extdesc2, ext2); err != nil {
		t.Fatalf("Could not set ext2: %s", err)
	}
	extdesc3 := &vendor.ExtensionDesc{
		ExtendedType:  (*vendor.MyMessage)(nil),
		ExtensionType: (*vendor.Ext)(nil),
		Field:         123456791,
		Name:          "a.d",
		Tag:           "bytes,123456791,opt",
	}
	ext3 := &vendor.Ext{Data: vendor.String("foo")}
	if err := vendor.SetExtension(msg, extdesc3, ext3); err != nil {
		t.Fatalf("Could not set ext3: %s", err)
	}

	b, err := vendor.Marshal(msg)
	if err != nil {
		t.Fatalf("Could not marshal msg: %v", err)
	}
	if err := vendor.Unmarshal(b, msg); err != nil {
		t.Fatalf("Could not unmarshal into msg: %v", err)
	}

	var expected vendor.Buffer
	if err := expected.EncodeVarint(uint64((extdesc1.Field << 3) | vendor.WireVarint)); err != nil {
		t.Fatalf("failed to compute expected prefix for ext1: %s", err)
	}
	if err := expected.EncodeVarint(1 /* bool true */); err != nil {
		t.Fatalf("failed to compute expected value for ext1: %s", err)
	}

	if b, err := vendor.GetExtension(msg, &vendor.ExtensionDesc{Field: extdesc1.Field}); err != nil {
		t.Fatalf("Failed to get raw value for ext1: %s", err)
	} else if !reflect.DeepEqual(b, expected.Bytes()) {
		t.Fatalf("Raw value for ext1: got %v, want %v", b, expected.Bytes())
	}

	expected = vendor.Buffer{} // reset
	if err := expected.EncodeVarint(uint64((extdesc2.Field << 3) | vendor.WireBytes)); err != nil {
		t.Fatalf("failed to compute expected prefix for ext2: %s", err)
	}
	if err := expected.EncodeRawBytes(ext2); err != nil {
		t.Fatalf("failed to compute expected value for ext2: %s", err)
	}

	if b, err := vendor.GetExtension(msg, &vendor.ExtensionDesc{Field: extdesc2.Field}); err != nil {
		t.Fatalf("Failed to get raw value for ext2: %s", err)
	} else if !reflect.DeepEqual(b, expected.Bytes()) {
		t.Fatalf("Raw value for ext2: got %v, want %v", b, expected.Bytes())
	}

	expected = vendor.Buffer{} // reset
	if err := expected.EncodeVarint(uint64((extdesc3.Field << 3) | vendor.WireBytes)); err != nil {
		t.Fatalf("failed to compute expected prefix for ext3: %s", err)
	}
	if b, err := vendor.Marshal(ext3); err != nil {
		t.Fatalf("failed to compute expected value for ext3: %s", err)
	} else if err := expected.EncodeRawBytes(b); err != nil {
		t.Fatalf("failed to compute expected value for ext3: %s", err)
	}

	if b, err := vendor.GetExtension(msg, &vendor.ExtensionDesc{Field: extdesc3.Field}); err != nil {
		t.Fatalf("Failed to get raw value for ext3: %s", err)
	} else if !reflect.DeepEqual(b, expected.Bytes()) {
		t.Fatalf("Raw value for ext3: got %v, want %v", b, expected.Bytes())
	}
}

func TestExtensionDescsWithUnregisteredExtensions(t *testing.T) {
	msg := &vendor.MyMessage{Count: vendor.Int32(0)}
	extdesc1 := vendor.E_Ext_More
	if descs, err := vendor.ExtensionDescs(msg); len(descs) != 0 || err != nil {
		t.Errorf("proto.ExtensionDescs: got %d descs, error %v; want 0, nil", len(descs), err)
	}

	ext1 := &vendor.Ext{}
	if err := vendor.SetExtension(msg, extdesc1, ext1); err != nil {
		t.Fatalf("Could not set ext1: %s", err)
	}
	extdesc2 := &vendor.ExtensionDesc{
		ExtendedType:  (*vendor.MyMessage)(nil),
		ExtensionType: (*bool)(nil),
		Field:         123456789,
		Name:          "a.b",
		Tag:           "varint,123456789,opt",
	}
	ext2 := vendor.Bool(false)
	if err := vendor.SetExtension(msg, extdesc2, ext2); err != nil {
		t.Fatalf("Could not set ext2: %s", err)
	}

	b, err := vendor.Marshal(msg)
	if err != nil {
		t.Fatalf("Could not marshal msg: %v", err)
	}
	if err := vendor.Unmarshal(b, msg); err != nil {
		t.Fatalf("Could not unmarshal into msg: %v", err)
	}

	descs, err := vendor.ExtensionDescs(msg)
	if err != nil {
		t.Fatalf("proto.ExtensionDescs: got error %v", err)
	}
	sortExtDescs(descs)
	wantDescs := []*vendor.ExtensionDesc{extdesc1, {Field: extdesc2.Field}}
	if !reflect.DeepEqual(descs, wantDescs) {
		t.Errorf("proto.ExtensionDescs(msg) sorted extension ids: got %+v, want %+v", descs, wantDescs)
	}
}

type ExtensionDescSlice []*vendor.ExtensionDesc

func (s ExtensionDescSlice) Len() int           { return len(s) }
func (s ExtensionDescSlice) Less(i, j int) bool { return s[i].Field < s[j].Field }
func (s ExtensionDescSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func sortExtDescs(s []*vendor.ExtensionDesc) {
	sort.Sort(ExtensionDescSlice(s))
}

func TestGetExtensionStability(t *testing.T) {
	check := func(m *vendor.MyMessage) bool {
		ext1, err := vendor.GetExtension(m, vendor.E_Ext_More)
		if err != nil {
			t.Fatalf("GetExtension() failed: %s", err)
		}
		ext2, err := vendor.GetExtension(m, vendor.E_Ext_More)
		if err != nil {
			t.Fatalf("GetExtension() failed: %s", err)
		}
		return ext1 == ext2
	}
	msg := &vendor.MyMessage{Count: vendor.Int32(4)}
	ext0 := &vendor.Ext{}
	if err := vendor.SetExtension(msg, vendor.E_Ext_More, ext0); err != nil {
		t.Fatalf("Could not set ext1: %s", ext0)
	}
	if !check(msg) {
		t.Errorf("GetExtension() not stable before marshaling")
	}
	bb, err := vendor.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal() failed: %s", err)
	}
	msg1 := &vendor.MyMessage{}
	err = vendor.Unmarshal(bb, msg1)
	if err != nil {
		t.Fatalf("Unmarshal() failed: %s", err)
	}
	if !check(msg1) {
		t.Errorf("GetExtension() not stable after unmarshaling")
	}
}

func TestGetExtensionDefaults(t *testing.T) {
	var setFloat64 float64 = 1
	var setFloat32 float32 = 2
	var setInt32 int32 = 3
	var setInt64 int64 = 4
	var setUint32 uint32 = 5
	var setUint64 uint64 = 6
	var setBool = true
	var setBool2 = false
	var setString = "Goodnight string"
	var setBytes = []byte("Goodnight bytes")
	var setEnum = vendor.DefaultsMessage_TWO

	type testcase struct {
		ext  *vendor.ExtensionDesc // Extension we are testing.
		want interface{}           // Expected value of extension, or nil (meaning that GetExtension will fail).
		def  interface{}           // Expected value of extension after ClearExtension().
	}
	tests := []testcase{
		{vendor.E_NoDefaultDouble, setFloat64, nil},
		{vendor.E_NoDefaultFloat, setFloat32, nil},
		{vendor.E_NoDefaultInt32, setInt32, nil},
		{vendor.E_NoDefaultInt64, setInt64, nil},
		{vendor.E_NoDefaultUint32, setUint32, nil},
		{vendor.E_NoDefaultUint64, setUint64, nil},
		{vendor.E_NoDefaultSint32, setInt32, nil},
		{vendor.E_NoDefaultSint64, setInt64, nil},
		{vendor.E_NoDefaultFixed32, setUint32, nil},
		{vendor.E_NoDefaultFixed64, setUint64, nil},
		{vendor.E_NoDefaultSfixed32, setInt32, nil},
		{vendor.E_NoDefaultSfixed64, setInt64, nil},
		{vendor.E_NoDefaultBool, setBool, nil},
		{vendor.E_NoDefaultBool, setBool2, nil},
		{vendor.E_NoDefaultString, setString, nil},
		{vendor.E_NoDefaultBytes, setBytes, nil},
		{vendor.E_NoDefaultEnum, setEnum, nil},
		{vendor.E_DefaultDouble, setFloat64, float64(3.1415)},
		{vendor.E_DefaultFloat, setFloat32, float32(3.14)},
		{vendor.E_DefaultInt32, setInt32, int32(42)},
		{vendor.E_DefaultInt64, setInt64, int64(43)},
		{vendor.E_DefaultUint32, setUint32, uint32(44)},
		{vendor.E_DefaultUint64, setUint64, uint64(45)},
		{vendor.E_DefaultSint32, setInt32, int32(46)},
		{vendor.E_DefaultSint64, setInt64, int64(47)},
		{vendor.E_DefaultFixed32, setUint32, uint32(48)},
		{vendor.E_DefaultFixed64, setUint64, uint64(49)},
		{vendor.E_DefaultSfixed32, setInt32, int32(50)},
		{vendor.E_DefaultSfixed64, setInt64, int64(51)},
		{vendor.E_DefaultBool, setBool, true},
		{vendor.E_DefaultBool, setBool2, true},
		{vendor.E_DefaultString, setString, "Hello, string,def=foo"},
		{vendor.E_DefaultBytes, setBytes, []byte("Hello, bytes")},
		{vendor.E_DefaultEnum, setEnum, vendor.DefaultsMessage_ONE},
	}

	checkVal := func(test testcase, msg *vendor.DefaultsMessage, valWant interface{}) error {
		val, err := vendor.GetExtension(msg, test.ext)
		if err != nil {
			if valWant != nil {
				return fmt.Errorf("GetExtension(): %s", err)
			}
			if want := vendor.ErrMissingExtension; err != want {
				return fmt.Errorf("Unexpected error: got %v, want %v", err, want)
			}
			return nil
		}

		// All proto2 extension values are either a pointer to a value or a slice of values.
		ty := reflect.TypeOf(val)
		tyWant := reflect.TypeOf(test.ext.ExtensionType)
		if got, want := ty, tyWant; got != want {
			return fmt.Errorf("unexpected reflect.TypeOf(): got %v want %v", got, want)
		}
		tye := ty.Elem()
		tyeWant := tyWant.Elem()
		if got, want := tye, tyeWant; got != want {
			return fmt.Errorf("unexpected reflect.TypeOf().Elem(): got %v want %v", got, want)
		}

		// Check the name of the type of the value.
		// If it is an enum it will be type int32 with the name of the enum.
		if got, want := tye.Name(), tye.Name(); got != want {
			return fmt.Errorf("unexpected reflect.TypeOf().Elem().Name(): got %v want %v", got, want)
		}

		// Check that value is what we expect.
		// If we have a pointer in val, get the value it points to.
		valExp := val
		if ty.Kind() == reflect.Ptr {
			valExp = reflect.ValueOf(val).Elem().Interface()
		}
		if got, want := valExp, valWant; !reflect.DeepEqual(got, want) {
			return fmt.Errorf("unexpected reflect.DeepEqual(): got %v want %v", got, want)
		}

		return nil
	}

	setTo := func(test testcase) interface{} {
		setTo := reflect.ValueOf(test.want)
		if typ := reflect.TypeOf(test.ext.ExtensionType); typ.Kind() == reflect.Ptr {
			setTo = reflect.New(typ).Elem()
			setTo.Set(reflect.New(setTo.Type().Elem()))
			setTo.Elem().Set(reflect.ValueOf(test.want))
		}
		return setTo.Interface()
	}

	for _, test := range tests {
		msg := &vendor.DefaultsMessage{}
		name := test.ext.Name

		// Check the initial value.
		if err := checkVal(test, msg, test.def); err != nil {
			t.Errorf("%s: %v", name, err)
		}

		// Set the per-type value and check value.
		name = fmt.Sprintf("%s (set to %T %v)", name, test.want, test.want)
		if err := vendor.SetExtension(msg, test.ext, setTo(test)); err != nil {
			t.Errorf("%s: SetExtension(): %v", name, err)
			continue
		}
		if err := checkVal(test, msg, test.want); err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}

		// Set and check the value.
		name += " (cleared)"
		vendor.ClearExtension(msg, test.ext)
		if err := checkVal(test, msg, test.def); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestNilMessage(t *testing.T) {
	name := "nil interface"
	if got, err := vendor.GetExtension(nil, vendor.E_Ext_More); err == nil {
		t.Errorf("%s: got %T %v, expected to fail", name, got, got)
	} else if !strings.Contains(err.Error(), "extendable") {
		t.Errorf("%s: got error %v, expected not-extendable error", name, err)
	}

	// Regression tests: all functions of the Extension API
	// used to panic when passed (*M)(nil), where M is a concrete message
	// type.  Now they handle this gracefully as a no-op or reported error.
	var nilMsg *vendor.MyMessage
	desc := vendor.E_Ext_More

	isNotExtendable := func(err error) bool {
		return strings.Contains(fmt.Sprint(err), "not extendable")
	}

	if vendor.HasExtension(nilMsg, desc) {
		t.Error("HasExtension(nil) = true")
	}

	if _, err := vendor.GetExtensions(nilMsg, []*vendor.ExtensionDesc{desc}); !isNotExtendable(err) {
		t.Errorf("GetExtensions(nil) = %q (wrong error)", err)
	}

	if _, err := vendor.ExtensionDescs(nilMsg); !isNotExtendable(err) {
		t.Errorf("ExtensionDescs(nil) = %q (wrong error)", err)
	}

	if err := vendor.SetExtension(nilMsg, desc, nil); !isNotExtendable(err) {
		t.Errorf("SetExtension(nil) = %q (wrong error)", err)
	}

	vendor.ClearExtension(nilMsg, desc) // no-op
	vendor.ClearAllExtensions(nilMsg)   // no-op
}

func TestExtensionsRoundTrip(t *testing.T) {
	msg := &vendor.MyMessage{}
	ext1 := &vendor.Ext{
		Data: vendor.String("hi"),
	}
	ext2 := &vendor.Ext{
		Data: vendor.String("there"),
	}
	exists := vendor.HasExtension(msg, vendor.E_Ext_More)
	if exists {
		t.Error("Extension More present unexpectedly")
	}
	if err := vendor.SetExtension(msg, vendor.E_Ext_More, ext1); err != nil {
		t.Error(err)
	}
	if err := vendor.SetExtension(msg, vendor.E_Ext_More, ext2); err != nil {
		t.Error(err)
	}
	e, err := vendor.GetExtension(msg, vendor.E_Ext_More)
	if err != nil {
		t.Error(err)
	}
	x, ok := e.(*vendor.Ext)
	if !ok {
		t.Errorf("e has type %T, expected test_proto.Ext", e)
	} else if *x.Data != "there" {
		t.Errorf("SetExtension failed to overwrite, got %+v, not 'there'", x)
	}
	vendor.ClearExtension(msg, vendor.E_Ext_More)
	if _, err = vendor.GetExtension(msg, vendor.E_Ext_More); err != vendor.ErrMissingExtension {
		t.Errorf("got %v, expected ErrMissingExtension", e)
	}
	if _, err := vendor.GetExtension(msg, vendor.E_X215); err == nil {
		t.Error("expected bad extension error, got nil")
	}
	if err := vendor.SetExtension(msg, vendor.E_X215, 12); err == nil {
		t.Error("expected extension err")
	}
	if err := vendor.SetExtension(msg, vendor.E_Ext_More, 12); err == nil {
		t.Error("expected some sort of type mismatch error, got nil")
	}
}

func TestNilExtension(t *testing.T) {
	msg := &vendor.MyMessage{
		Count: vendor.Int32(1),
	}
	if err := vendor.SetExtension(msg, vendor.E_Ext_Text, vendor.String("hello")); err != nil {
		t.Fatal(err)
	}
	if err := vendor.SetExtension(msg, vendor.E_Ext_More, (*vendor.Ext)(nil)); err == nil {
		t.Error("expected SetExtension to fail due to a nil extension")
	} else if want := fmt.Sprintf("proto: SetExtension called with nil value of type %T", new(vendor.Ext)); err.Error() != want {
		t.Errorf("expected error %v, got %v", want, err)
	}
	// Note: if the behavior of Marshal is ever changed to ignore nil extensions, update
	// this test to verify that E_Ext_Text is properly propagated through marshal->unmarshal.
}

func TestMarshalUnmarshalRepeatedExtension(t *testing.T) {
	// Add a repeated extension to the result.
	tests := []struct {
		name string
		ext  []*vendor.ComplexExtension
	}{
		{
			"two fields",
			[]*vendor.ComplexExtension{
				{First: vendor.Int32(7)},
				{Second: vendor.Int32(11)},
			},
		},
		{
			"repeated field",
			[]*vendor.ComplexExtension{
				{Third: []int32{1000}},
				{Third: []int32{2000}},
			},
		},
		{
			"two fields and repeated field",
			[]*vendor.ComplexExtension{
				{Third: []int32{1000}},
				{First: vendor.Int32(9)},
				{Second: vendor.Int32(21)},
				{Third: []int32{2000}},
			},
		},
	}
	for _, test := range tests {
		// Marshal message with a repeated extension.
		msg1 := new(vendor.OtherMessage)
		err := vendor.SetExtension(msg1, vendor.E_RComplex, test.ext)
		if err != nil {
			t.Fatalf("[%s] Error setting extension: %v", test.name, err)
		}
		b, err := vendor.Marshal(msg1)
		if err != nil {
			t.Fatalf("[%s] Error marshaling message: %v", test.name, err)
		}

		// Unmarshal and read the merged proto.
		msg2 := new(vendor.OtherMessage)
		err = vendor.Unmarshal(b, msg2)
		if err != nil {
			t.Fatalf("[%s] Error unmarshaling message: %v", test.name, err)
		}
		e, err := vendor.GetExtension(msg2, vendor.E_RComplex)
		if err != nil {
			t.Fatalf("[%s] Error getting extension: %v", test.name, err)
		}
		ext := e.([]*vendor.ComplexExtension)
		if ext == nil {
			t.Fatalf("[%s] Invalid extension", test.name)
		}
		if len(ext) != len(test.ext) {
			t.Errorf("[%s] Wrong length of ComplexExtension: got: %v want: %v\n", test.name, len(ext), len(test.ext))
		}
		for i := range test.ext {
			if !vendor.Equal(ext[i], test.ext[i]) {
				t.Errorf("[%s] Wrong value for ComplexExtension[%d]: got: %v want: %v\n", test.name, i, ext[i], test.ext[i])
			}
		}
	}
}

func TestUnmarshalRepeatingNonRepeatedExtension(t *testing.T) {
	// We may see multiple instances of the same extension in the wire
	// format. For example, the proto compiler may encode custom options in
	// this way. Here, we verify that we merge the extensions together.
	tests := []struct {
		name string
		ext  []*vendor.ComplexExtension
	}{
		{
			"two fields",
			[]*vendor.ComplexExtension{
				{First: vendor.Int32(7)},
				{Second: vendor.Int32(11)},
			},
		},
		{
			"repeated field",
			[]*vendor.ComplexExtension{
				{Third: []int32{1000}},
				{Third: []int32{2000}},
			},
		},
		{
			"two fields and repeated field",
			[]*vendor.ComplexExtension{
				{Third: []int32{1000}},
				{First: vendor.Int32(9)},
				{Second: vendor.Int32(21)},
				{Third: []int32{2000}},
			},
		},
	}
	for _, test := range tests {
		var buf bytes.Buffer
		var want vendor.ComplexExtension

		// Generate a serialized representation of a repeated extension
		// by catenating bytes together.
		for i, e := range test.ext {
			// Merge to create the wanted proto.
			vendor.Merge(&want, e)

			// serialize the message
			msg := new(vendor.OtherMessage)
			err := vendor.SetExtension(msg, vendor.E_Complex, e)
			if err != nil {
				t.Fatalf("[%s] Error setting extension %d: %v", test.name, i, err)
			}
			b, err := vendor.Marshal(msg)
			if err != nil {
				t.Fatalf("[%s] Error marshaling message %d: %v", test.name, i, err)
			}
			buf.Write(b)
		}

		// Unmarshal and read the merged proto.
		msg2 := new(vendor.OtherMessage)
		err := vendor.Unmarshal(buf.Bytes(), msg2)
		if err != nil {
			t.Fatalf("[%s] Error unmarshaling message: %v", test.name, err)
		}
		e, err := vendor.GetExtension(msg2, vendor.E_Complex)
		if err != nil {
			t.Fatalf("[%s] Error getting extension: %v", test.name, err)
		}
		ext := e.(*vendor.ComplexExtension)
		if ext == nil {
			t.Fatalf("[%s] Invalid extension", test.name)
		}
		if !vendor.Equal(ext, &want) {
			t.Errorf("[%s] Wrong value for ComplexExtension: got: %s want: %s\n", test.name, ext, &want)
		}
	}
}

func TestClearAllExtensions(t *testing.T) {
	// unregistered extension
	desc := &vendor.ExtensionDesc{
		ExtendedType:  (*vendor.MyMessage)(nil),
		ExtensionType: (*bool)(nil),
		Field:         101010100,
		Name:          "emptyextension",
		Tag:           "varint,0,opt",
	}
	m := &vendor.MyMessage{}
	if vendor.HasExtension(m, desc) {
		t.Errorf("proto.HasExtension(%s): got true, want false", vendor.MarshalTextString(m))
	}
	if err := vendor.SetExtension(m, desc, vendor.Bool(true)); err != nil {
		t.Errorf("proto.SetExtension(m, desc, true): got error %q, want nil", err)
	}
	if !vendor.HasExtension(m, desc) {
		t.Errorf("proto.HasExtension(%s): got false, want true", vendor.MarshalTextString(m))
	}
	vendor.ClearAllExtensions(m)
	if vendor.HasExtension(m, desc) {
		t.Errorf("proto.HasExtension(%s): got true, want false", vendor.MarshalTextString(m))
	}
}

func TestMarshalRace(t *testing.T) {
	ext := &vendor.Ext{}
	m := &vendor.MyMessage{Count: vendor.Int32(4)}
	if err := vendor.SetExtension(m, vendor.E_Ext_More, ext); err != nil {
		t.Fatalf("proto.SetExtension(m, desc, true): got error %q, want nil", err)
	}

	b, err := vendor.Marshal(m)
	if err != nil {
		t.Fatalf("Could not marshal message: %v", err)
	}
	if err := vendor.Unmarshal(b, m); err != nil {
		t.Fatalf("Could not unmarshal message: %v", err)
	}
	// after Unmarshal, the extension is in undecoded form.
	// GetExtension will decode it lazily. Make sure this does
	// not race against Marshal.

	wg := sync.WaitGroup{}
	errs := make(chan error, 3)
	for n := 3; n > 0; n-- {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := vendor.Marshal(m)
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)

	for err = range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
}
