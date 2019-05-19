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

// Test that we can use protocol buffers that use extensions.

package testdata

import (
	"bytes"
	"regexp"
	"testing"
	"vendor"
)

func TestSingleFieldExtension(t *testing.T) {
	bm := &vendor.BaseMessage{
		Height: vendor.Int32(178),
	}

	// Use extension within scope of another type.
	vol := vendor.Uint32(11)
	err := vendor.SetExtension(bm, vendor.E_LoudMessage_Volume, vol)
	if err != nil {
		t.Fatal("Failed setting extension:", err)
	}
	buf, err := vendor.Marshal(bm)
	if err != nil {
		t.Fatal("Failed encoding message with extension:", err)
	}
	bm_new := new(vendor.BaseMessage)
	if err := vendor.Unmarshal(buf, bm_new); err != nil {
		t.Fatal("Failed decoding message with extension:", err)
	}
	if !vendor.HasExtension(bm_new, vendor.E_LoudMessage_Volume) {
		t.Fatal("Decoded message didn't contain extension.")
	}
	vol_out, err := vendor.GetExtension(bm_new, vendor.E_LoudMessage_Volume)
	if err != nil {
		t.Fatal("Failed getting extension:", err)
	}
	if v := vol_out.(*uint32); *v != *vol {
		t.Errorf("vol_out = %v, expected %v", *v, *vol)
	}
	vendor.ClearExtension(bm_new, vendor.E_LoudMessage_Volume)
	if vendor.HasExtension(bm_new, vendor.E_LoudMessage_Volume) {
		t.Fatal("Failed clearing extension.")
	}
}

func TestMessageExtension(t *testing.T) {
	bm := &vendor.BaseMessage{
		Height: vendor.Int32(179),
	}

	// Use extension that is itself a message.
	um := &vendor.UserMessage{
		Name: vendor.String("Dave"),
		Rank: vendor.String("Major"),
	}
	err := vendor.SetExtension(bm, vendor.E_LoginMessage_UserMessage, um)
	if err != nil {
		t.Fatal("Failed setting extension:", err)
	}
	buf, err := vendor.Marshal(bm)
	if err != nil {
		t.Fatal("Failed encoding message with extension:", err)
	}
	bm_new := new(vendor.BaseMessage)
	if err := vendor.Unmarshal(buf, bm_new); err != nil {
		t.Fatal("Failed decoding message with extension:", err)
	}
	if !vendor.HasExtension(bm_new, vendor.E_LoginMessage_UserMessage) {
		t.Fatal("Decoded message didn't contain extension.")
	}
	um_out, err := vendor.GetExtension(bm_new, vendor.E_LoginMessage_UserMessage)
	if err != nil {
		t.Fatal("Failed getting extension:", err)
	}
	if n := um_out.(*vendor.UserMessage).Name; *n != *um.Name {
		t.Errorf("um_out.Name = %q, expected %q", *n, *um.Name)
	}
	if r := um_out.(*vendor.UserMessage).Rank; *r != *um.Rank {
		t.Errorf("um_out.Rank = %q, expected %q", *r, *um.Rank)
	}
	vendor.ClearExtension(bm_new, vendor.E_LoginMessage_UserMessage)
	if vendor.HasExtension(bm_new, vendor.E_LoginMessage_UserMessage) {
		t.Fatal("Failed clearing extension.")
	}
}

func TestTopLevelExtension(t *testing.T) {
	bm := &vendor.BaseMessage{
		Height: vendor.Int32(179),
	}

	width := vendor.Int32(17)
	err := vendor.SetExtension(bm, vendor.E_Width, width)
	if err != nil {
		t.Fatal("Failed setting extension:", err)
	}
	buf, err := vendor.Marshal(bm)
	if err != nil {
		t.Fatal("Failed encoding message with extension:", err)
	}
	bm_new := new(vendor.BaseMessage)
	if err := vendor.Unmarshal(buf, bm_new); err != nil {
		t.Fatal("Failed decoding message with extension:", err)
	}
	if !vendor.HasExtension(bm_new, vendor.E_Width) {
		t.Fatal("Decoded message didn't contain extension.")
	}
	width_out, err := vendor.GetExtension(bm_new, vendor.E_Width)
	if err != nil {
		t.Fatal("Failed getting extension:", err)
	}
	if w := width_out.(*int32); *w != *width {
		t.Errorf("width_out = %v, expected %v", *w, *width)
	}
	vendor.ClearExtension(bm_new, vendor.E_Width)
	if vendor.HasExtension(bm_new, vendor.E_Width) {
		t.Fatal("Failed clearing extension.")
	}
}

func TestMessageSetWireFormat(t *testing.T) {
	osm := new(vendor.OldStyleMessage)
	osp := &vendor.OldStyleParcel{
		Name:   vendor.String("Dave"),
		Height: vendor.Int32(178),
	}

	err := vendor.SetExtension(osm, vendor.E_OldStyleParcel_MessageSetExtension, osp)
	if err != nil {
		t.Fatal("Failed setting extension:", err)
	}

	buf, err := vendor.Marshal(osm)
	if err != nil {
		t.Fatal("Failed encoding message:", err)
	}

	// Data generated from Python implementation.
	expected := []byte{
		11, 16, 209, 15, 26, 9, 10, 4, 68, 97, 118, 101, 16, 178, 1, 12,
	}

	if !bytes.Equal(expected, buf) {
		t.Errorf("Encoding mismatch.\nwant %+v\n got %+v", expected, buf)
	}

	// Check that it is restored correctly.
	osm = new(vendor.OldStyleMessage)
	if err := vendor.Unmarshal(buf, osm); err != nil {
		t.Fatal("Failed decoding message:", err)
	}
	osp_out, err := vendor.GetExtension(osm, vendor.E_OldStyleParcel_MessageSetExtension)
	if err != nil {
		t.Fatal("Failed getting extension:", err)
	}
	osp = osp_out.(*vendor.OldStyleParcel)
	if *osp.Name != "Dave" || *osp.Height != 178 {
		t.Errorf("Retrieved extension from decoded message is not correct: %+v", osp)
	}
}

func main() {
	// simpler than rigging up gotest
	testing.Main(regexp.MatchString, []testing.InternalTest{
		{"TestSingleFieldExtension", TestSingleFieldExtension},
		{"TestMessageExtension", TestMessageExtension},
		{"TestTopLevelExtension", TestTopLevelExtension},
	},
		[]testing.InternalBenchmark{},
		[]testing.InternalExample{})
}
