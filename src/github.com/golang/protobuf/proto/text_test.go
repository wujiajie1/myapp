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
	"errors"
	"io/ioutil"
	"math"
	"strings"
	"sync"
	"testing"
	"vendor"
)

// textMessage implements the methods that allow it to marshal and unmarshal
// itself as text.
type textMessage struct {
}

func (*textMessage) MarshalText() ([]byte, error) {
	return []byte("custom"), nil
}

func (*textMessage) UnmarshalText(bytes []byte) error {
	if string(bytes) != "custom" {
		return errors.New("expected 'custom'")
	}
	return nil
}

func (*textMessage) Reset()         {}
func (*textMessage) String() string { return "" }
func (*textMessage) ProtoMessage()  {}

func newTestMessage() *vendor.MyMessage {
	msg := &vendor.MyMessage{
		Count: vendor.Int32(42),
		Name:  vendor.String("Dave"),
		Quote: vendor.String(`"I didn't want to go."`),
		Pet:   []string{"bunny", "kitty", "horsey"},
		Inner: &vendor.InnerMessage{
			Host:      vendor.String("footrest.syd"),
			Port:      vendor.Int32(7001),
			Connected: vendor.Bool(true),
		},
		Others: []*vendor.OtherMessage{
			{
				Key:   vendor.Int64(0xdeadbeef),
				Value: []byte{1, 65, 7, 12},
			},
			{
				Weight: vendor.Float32(6.022),
				Inner: &vendor.InnerMessage{
					Host: vendor.String("lesha.mtv"),
					Port: vendor.Int32(8002),
				},
			},
		},
		Bikeshed: vendor.MyMessage_BLUE.Enum(),
		Somegroup: &vendor.MyMessage_SomeGroup{
			GroupField: vendor.Int32(8),
		},
		// One normally wouldn't do this.
		// This is an undeclared tag 13, as a varint (wire type 0) with value 4.
		XXX_unrecognized: []byte{13<<3 | 0, 4},
	}
	ext := &vendor.Ext{
		Data: vendor.String("Big gobs for big rats"),
	}
	if err := vendor.SetExtension(msg, vendor.E_Ext_More, ext); err != nil {
		panic(err)
	}
	greetings := []string{"adg", "easy", "cow"}
	if err := vendor.SetExtension(msg, vendor.E_Greeting, greetings); err != nil {
		panic(err)
	}

	// Add an unknown extension. We marshal a pb.Ext, and fake the ID.
	b, err := vendor.Marshal(&vendor.Ext{Data: vendor.String("3G skiing")})
	if err != nil {
		panic(err)
	}
	b = append(vendor.EncodeVarint(201<<3|vendor.WireBytes), b...)
	vendor.SetRawExtension(msg, 201, b)

	// Extensions can be plain fields, too, so let's test that.
	b = append(vendor.EncodeVarint(202<<3|vendor.WireVarint), 19)
	vendor.SetRawExtension(msg, 202, b)

	return msg
}

const text = `count: 42
name: "Dave"
quote: "\"I didn't want to go.\""
pet: "bunny"
pet: "kitty"
pet: "horsey"
inner: <
  host: "footrest.syd"
  port: 7001
  connected: true
>
others: <
  key: 3735928559
  value: "\001A\007\014"
>
others: <
  weight: 6.022
  inner: <
    host: "lesha.mtv"
    port: 8002
  >
>
bikeshed: BLUE
SomeGroup {
  group_field: 8
}
/* 2 unknown bytes */
13: 4
[test_proto.Ext.more]: <
  data: "Big gobs for big rats"
>
[test_proto.greeting]: "adg"
[test_proto.greeting]: "easy"
[test_proto.greeting]: "cow"
/* 13 unknown bytes */
201: "\t3G skiing"
/* 3 unknown bytes */
202: 19
`

func TestMarshalText(t *testing.T) {
	buf := new(bytes.Buffer)
	if err := vendor.MarshalText(buf, newTestMessage()); err != nil {
		t.Fatalf("proto.MarshalText: %v", err)
	}
	s := buf.String()
	if s != text {
		t.Errorf("Got:\n===\n%v===\nExpected:\n===\n%v===\n", s, text)
	}
}

func TestMarshalTextCustomMessage(t *testing.T) {
	buf := new(bytes.Buffer)
	if err := vendor.MarshalText(buf, &textMessage{}); err != nil {
		t.Fatalf("proto.MarshalText: %v", err)
	}
	s := buf.String()
	if s != "custom" {
		t.Errorf("Got %q, expected %q", s, "custom")
	}
}
func TestMarshalTextNil(t *testing.T) {
	want := "<nil>"
	tests := []vendor.Message{nil, (*vendor.MyMessage)(nil)}
	for i, test := range tests {
		buf := new(bytes.Buffer)
		if err := vendor.MarshalText(buf, test); err != nil {
			t.Fatal(err)
		}
		if got := buf.String(); got != want {
			t.Errorf("%d: got %q want %q", i, got, want)
		}
	}
}

func TestMarshalTextUnknownEnum(t *testing.T) {
	// The Color enum only specifies values 0-2.
	m := &vendor.MyMessage{Bikeshed: vendor.MyMessage_Color(3).Enum()}
	got := m.String()
	const want = `bikeshed:3 `
	if got != want {
		t.Errorf("\n got %q\nwant %q", got, want)
	}
}

func TestTextOneof(t *testing.T) {
	tests := []struct {
		m    vendor.Message
		want string
	}{
		// zero message
		{&vendor.Communique{}, ``},
		// scalar field
		{&vendor.Communique{Union: &vendor.Communique_Number{4}}, `number:4`},
		// message field
		{&vendor.Communique{Union: &vendor.Communique_Msg{
			&vendor.Strings{StringField: vendor.String("why hello!")},
		}}, `msg:<string_field:"why hello!" >`},
		// bad oneof (should not panic)
		{&vendor.Communique{Union: &vendor.Communique_Msg{nil}}, `msg:/* nil */`},
	}
	for _, test := range tests {
		got := strings.TrimSpace(test.m.String())
		if got != test.want {
			t.Errorf("\n got %s\nwant %s", got, test.want)
		}
	}
}

func BenchmarkMarshalTextBuffered(b *testing.B) {
	buf := new(bytes.Buffer)
	m := newTestMessage()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		vendor.MarshalText(buf, m)
	}
}

func BenchmarkMarshalTextUnbuffered(b *testing.B) {
	w := ioutil.Discard
	m := newTestMessage()
	for i := 0; i < b.N; i++ {
		vendor.MarshalText(w, m)
	}
}

func compact(src string) string {
	// s/[ \n]+/ /g; s/ $//;
	dst := make([]byte, len(src))
	space, comment := false, false
	j := 0
	for i := 0; i < len(src); i++ {
		if strings.HasPrefix(src[i:], "/*") {
			comment = true
			i++
			continue
		}
		if comment && strings.HasPrefix(src[i:], "*/") {
			comment = false
			i++
			continue
		}
		if comment {
			continue
		}
		c := src[i]
		if c == ' ' || c == '\n' {
			space = true
			continue
		}
		if j > 0 && (dst[j-1] == ':' || dst[j-1] == '<' || dst[j-1] == '{') {
			space = false
		}
		if c == '{' {
			space = false
		}
		if space {
			dst[j] = ' '
			j++
			space = false
		}
		dst[j] = c
		j++
	}
	if space {
		dst[j] = ' '
		j++
	}
	return string(dst[0:j])
}

var compactText = compact(text)

func TestCompactText(t *testing.T) {
	s := vendor.CompactTextString(newTestMessage())
	if s != compactText {
		t.Errorf("Got:\n===\n%v===\nExpected:\n===\n%v\n===\n", s, compactText)
	}
}

func TestStringEscaping(t *testing.T) {
	testCases := []struct {
		in  *vendor.Strings
		out string
	}{
		{
			// Test data from C++ test (TextFormatTest.StringEscape).
			// Single divergence: we don't escape apostrophes.
			&vendor.Strings{StringField: vendor.String("\"A string with ' characters \n and \r newlines and \t tabs and \001 slashes \\ and  multiple   spaces")},
			"string_field: \"\\\"A string with ' characters \\n and \\r newlines and \\t tabs and \\001 slashes \\\\ and  multiple   spaces\"\n",
		},
		{
			// Test data from the same C++ test.
			&vendor.Strings{StringField: vendor.String("\350\260\267\346\255\214")},
			"string_field: \"\\350\\260\\267\\346\\255\\214\"\n",
		},
		{
			// Some UTF-8.
			&vendor.Strings{StringField: vendor.String("\x00\x01\xff\x81")},
			`string_field: "\000\001\377\201"` + "\n",
		},
	}

	for i, tc := range testCases {
		var buf bytes.Buffer
		if err := vendor.MarshalText(&buf, tc.in); err != nil {
			t.Errorf("proto.MarsalText: %v", err)
			continue
		}
		s := buf.String()
		if s != tc.out {
			t.Errorf("#%d: Got:\n%s\nExpected:\n%s\n", i, s, tc.out)
			continue
		}

		// Check round-trip.
		pb := new(vendor.Strings)
		if err := vendor.UnmarshalText(s, pb); err != nil {
			t.Errorf("#%d: UnmarshalText: %v", i, err)
			continue
		}
		if !vendor.Equal(pb, tc.in) {
			t.Errorf("#%d: Round-trip failed:\nstart: %v\n  end: %v", i, tc.in, pb)
		}
	}
}

// A limitedWriter accepts some output before it fails.
// This is a proxy for something like a nearly-full or imminently-failing disk,
// or a network connection that is about to die.
type limitedWriter struct {
	b     bytes.Buffer
	limit int
}

var outOfSpace = errors.New("proto: insufficient space")

func (w *limitedWriter) Write(p []byte) (n int, err error) {
	var avail = w.limit - w.b.Len()
	if avail <= 0 {
		return 0, outOfSpace
	}
	if len(p) <= avail {
		return w.b.Write(p)
	}
	n, _ = w.b.Write(p[:avail])
	return n, outOfSpace
}

func TestMarshalTextFailing(t *testing.T) {
	// Try lots of different sizes to exercise more error code-paths.
	for lim := 0; lim < len(text); lim++ {
		buf := new(limitedWriter)
		buf.limit = lim
		err := vendor.MarshalText(buf, newTestMessage())
		// We expect a certain error, but also some partial results in the buffer.
		if err != outOfSpace {
			t.Errorf("Got:\n===\n%v===\nExpected:\n===\n%v===\n", err, outOfSpace)
		}
		s := buf.b.String()
		x := text[:buf.limit]
		if s != x {
			t.Errorf("Got:\n===\n%v===\nExpected:\n===\n%v===\n", s, x)
		}
	}
}

func TestFloats(t *testing.T) {
	tests := []struct {
		f    float64
		want string
	}{
		{0, "0"},
		{4.7, "4.7"},
		{math.Inf(1), "inf"},
		{math.Inf(-1), "-inf"},
		{math.NaN(), "nan"},
	}
	for _, test := range tests {
		msg := &vendor.FloatingPoint{F: &test.f}
		got := strings.TrimSpace(msg.String())
		want := `f:` + test.want
		if got != want {
			t.Errorf("f=%f: got %q, want %q", test.f, got, want)
		}
	}
}

func TestRepeatedNilText(t *testing.T) {
	m := &vendor.MessageList{
		Message: []*vendor.MessageList_Message{
			nil,
			&vendor.MessageList_Message{
				Name: vendor.String("Horse"),
			},
			nil,
		},
	}
	want := `Message <nil>
Message {
  name: "Horse"
}
Message <nil>
`
	if s := vendor.MarshalTextString(m); s != want {
		t.Errorf(" got: %s\nwant: %s", s, want)
	}
}

func TestProto3Text(t *testing.T) {
	tests := []struct {
		m    vendor.Message
		want string
	}{
		// zero message
		{&vendor.Message{}, ``},
		// zero message except for an empty byte slice
		{&vendor.Message{Data: []byte{}}, ``},
		// trivial case
		{&vendor.Message{Name: "Rob", HeightInCm: 175}, `name:"Rob" height_in_cm:175`},
		// empty map
		{&vendor.MessageWithMap{}, ``},
		// non-empty map; map format is the same as a repeated struct,
		// and they are sorted by key (numerically for numeric keys).
		{
			&vendor.MessageWithMap{NameMapping: map[int32]string{
				-1:      "Negatory",
				7:       "Lucky",
				1234:    "Feist",
				6345789: "Otis",
			}},
			`name_mapping:<key:-1 value:"Negatory" > ` +
				`name_mapping:<key:7 value:"Lucky" > ` +
				`name_mapping:<key:1234 value:"Feist" > ` +
				`name_mapping:<key:6345789 value:"Otis" >`,
		},
		// map with nil value; not well-defined, but we shouldn't crash
		{
			&vendor.MessageWithMap{MsgMapping: map[int64]*vendor.FloatingPoint{7: nil}},
			`msg_mapping:<key:7 >`,
		},
	}
	for _, test := range tests {
		got := strings.TrimSpace(test.m.String())
		if got != test.want {
			t.Errorf("\n got %s\nwant %s", got, test.want)
		}
	}
}

func TestRacyMarshal(t *testing.T) {
	// This test should be run with the race detector.

	any := &vendor.MyMessage{Count: vendor.Int32(47), Name: vendor.String("David")}
	vendor.SetExtension(any, vendor.E_Ext_Text, vendor.String("bar"))
	b, err := vendor.Marshal(any)
	if err != nil {
		panic(err)
	}
	m := &vendor.Message{
		Name:        "David",
		ResultCount: 47,
		Anything:    &vendor.Any{TypeUrl: "type.googleapis.com/" + vendor.MessageName(any), Value: b},
	}

	wantText := vendor.MarshalTextString(m)
	wantBytes, err := vendor.Marshal(m)
	if err != nil {
		t.Fatalf("proto.Marshal error: %v", err)
	}

	var wg sync.WaitGroup
	defer wg.Wait()
	wg.Add(20)
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			got := vendor.MarshalTextString(m)
			if got != wantText {
				t.Errorf("proto.MarshalTextString = %q, want %q", got, wantText)
			}
		}()
		go func() {
			defer wg.Done()
			got, err := vendor.Marshal(m)
			if !bytes.Equal(got, wantBytes) || err != nil {
				t.Errorf("proto.Marshal = (%x, %v), want (%x, nil)", got, err, wantBytes)
			}
		}()
	}
}
