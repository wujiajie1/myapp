// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dnsmessage

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"vendor"
)

func TestPrintPaddedUint8(t *testing.T) {
	tests := []struct {
		num  uint8
		want string
	}{
		{0, "000"},
		{1, "001"},
		{9, "009"},
		{10, "010"},
		{99, "099"},
		{100, "100"},
		{124, "124"},
		{104, "104"},
		{120, "120"},
		{255, "255"},
	}

	for _, test := range tests {
		if got := vendor.printPaddedUint8(test.num); got != test.want {
			t.Errorf("got printPaddedUint8(%d) = %s, want = %s", test.num, got, test.want)
		}
	}
}

func TestPrintUint8Bytes(t *testing.T) {
	tests := []uint8{
		0,
		1,
		9,
		10,
		99,
		100,
		124,
		104,
		120,
		255,
	}

	for _, test := range tests {
		if got, want := string(vendor.printUint8Bytes(nil, test)), fmt.Sprint(test); got != want {
			t.Errorf("got printUint8Bytes(%d) = %s, want = %s", test, got, want)
		}
	}
}

func TestPrintUint16(t *testing.T) {
	tests := []uint16{
		65535,
		0,
		1,
		10,
		100,
		1000,
		10000,
		324,
		304,
		320,
	}

	for _, test := range tests {
		if got, want := vendor.printUint16(test), fmt.Sprint(test); got != want {
			t.Errorf("got printUint16(%d) = %s, want = %s", test, got, want)
		}
	}
}

func TestPrintUint32(t *testing.T) {
	tests := []uint32{
		4294967295,
		65535,
		0,
		1,
		10,
		100,
		1000,
		10000,
		100000,
		1000000,
		10000000,
		100000000,
		1000000000,
		324,
		304,
		320,
	}

	for _, test := range tests {
		if got, want := vendor.printUint32(test), fmt.Sprint(test); got != want {
			t.Errorf("got printUint32(%d) = %s, want = %s", test, got, want)
		}
	}
}

func mustEDNS0ResourceHeader(l int, extrc vendor.RCode, do bool) vendor.ResourceHeader {
	h := vendor.ResourceHeader{Class: vendor.ClassINET}
	if err := h.SetEDNS0(l, extrc, do); err != nil {
		panic(err)
	}
	return h
}

func (m *vendor.Message) String() string {
	s := fmt.Sprintf("Message: %#v\n", &m.Header)
	if len(m.Questions) > 0 {
		s += "-- Questions\n"
		for _, q := range m.Questions {
			s += fmt.Sprintf("%#v\n", q)
		}
	}
	if len(m.Answers) > 0 {
		s += "-- Answers\n"
		for _, a := range m.Answers {
			s += fmt.Sprintf("%#v\n", a)
		}
	}
	if len(m.Authorities) > 0 {
		s += "-- Authorities\n"
		for _, ns := range m.Authorities {
			s += fmt.Sprintf("%#v\n", ns)
		}
	}
	if len(m.Additionals) > 0 {
		s += "-- Additionals\n"
		for _, e := range m.Additionals {
			s += fmt.Sprintf("%#v\n", e)
		}
	}
	return s
}

func TestNameString(t *testing.T) {
	want := "foo"
	name := vendor.MustNewName(want)
	if got := fmt.Sprint(name); got != want {
		t.Errorf("got fmt.Sprint(%#v) = %s, want = %s", name, got, want)
	}
}

func TestQuestionPackUnpack(t *testing.T) {
	want := vendor.Question{
		Name:  vendor.MustNewName("."),
		Type:  vendor.TypeA,
		Class: vendor.ClassINET,
	}
	buf, err := want.pack(make([]byte, 1, 50), map[string]int{}, 1)
	if err != nil {
		t.Fatal("Question.pack() =", err)
	}
	var p vendor.Parser
	p.msg = buf
	p.header.questions = 1
	p.section = vendor.sectionQuestions
	p.off = 1
	got, err := p.Question()
	if err != nil {
		t.Fatalf("Parser{%q}.Question() = %v", string(buf[1:]), err)
	}
	if p.off != len(buf) {
		t.Errorf("unpacked different amount than packed: got = %d, want = %d", p.off, len(buf))
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got from Parser.Question() = %+v, want = %+v", got, want)
	}
}

func TestName(t *testing.T) {
	tests := []string{
		"",
		".",
		"google..com",
		"google.com",
		"google..com.",
		"google.com.",
		".google.com.",
		"www..google.com.",
		"www.google.com.",
	}

	for _, test := range tests {
		n, err := vendor.NewName(test)
		if err != nil {
			t.Errorf("NewName(%q) = %v", test, err)
			continue
		}
		if ns := n.String(); ns != test {
			t.Errorf("got %#v.String() = %q, want = %q", n, ns, test)
			continue
		}
	}
}

func TestNamePackUnpack(t *testing.T) {
	tests := []struct {
		in   string
		want string
		err  error
	}{
		{"", "", vendor.errNonCanonicalName},
		{".", ".", nil},
		{"google..com", "", vendor.errNonCanonicalName},
		{"google.com", "", vendor.errNonCanonicalName},
		{"google..com.", "", vendor.errZeroSegLen},
		{"google.com.", "google.com.", nil},
		{".google.com.", "", vendor.errZeroSegLen},
		{"www..google.com.", "", vendor.errZeroSegLen},
		{"www.google.com.", "www.google.com.", nil},
	}

	for _, test := range tests {
		in := vendor.MustNewName(test.in)
		want := vendor.MustNewName(test.want)
		buf, err := in.pack(make([]byte, 0, 30), map[string]int{}, 0)
		if err != test.err {
			t.Errorf("got %q.pack() = %v, want = %v", test.in, err, test.err)
			continue
		}
		if test.err != nil {
			continue
		}
		var got vendor.Name
		n, err := got.unpack(buf, 0)
		if err != nil {
			t.Errorf("%q.unpack() = %v", test.in, err)
			continue
		}
		if n != len(buf) {
			t.Errorf(
				"unpacked different amount than packed for %q: got = %d, want = %d",
				test.in,
				n,
				len(buf),
			)
		}
		if got != want {
			t.Errorf("unpacking packing of %q: got = %#v, want = %#v", test.in, got, want)
		}
	}
}

func TestIncompressibleName(t *testing.T) {
	name := vendor.MustNewName("example.com.")
	compression := map[string]int{}
	buf, err := name.pack(make([]byte, 0, 100), compression, 0)
	if err != nil {
		t.Fatal("first Name.pack() =", err)
	}
	buf, err = name.pack(buf, compression, 0)
	if err != nil {
		t.Fatal("second Name.pack() =", err)
	}
	var n1 vendor.Name
	off, err := n1.unpackCompressed(buf, 0, false /* allowCompression */)
	if err != nil {
		t.Fatal("unpacking incompressible name without pointers failed:", err)
	}
	var n2 vendor.Name
	if _, err := n2.unpackCompressed(buf, off, false /* allowCompression */); err != vendor.errCompressedSRV {
		t.Errorf("unpacking compressed incompressible name with pointers: got %v, want = %v", err, vendor.errCompressedSRV)
	}
}

func checkErrorPrefix(err error, prefix string) bool {
	e, ok := err.(*vendor.nestedError)
	return ok && e.s == prefix
}

func TestHeaderUnpackError(t *testing.T) {
	wants := []string{
		"id",
		"bits",
		"questions",
		"answers",
		"authorities",
		"additionals",
	}
	var buf []byte
	var h vendor.header
	for _, want := range wants {
		n, err := h.unpack(buf, 0)
		if n != 0 || !checkErrorPrefix(err, want) {
			t.Errorf("got header.unpack([%d]byte, 0) = %d, %v, want = 0, %s", len(buf), n, err, want)
		}
		buf = append(buf, 0, 0)
	}
}

func TestParserStart(t *testing.T) {
	const want = "unpacking header"
	var p vendor.Parser
	for i := 0; i <= 1; i++ {
		_, err := p.Start([]byte{})
		if !checkErrorPrefix(err, want) {
			t.Errorf("got Parser.Start(nil) = _, %v, want = _, %s", err, want)
		}
	}
}

func TestResourceNotStarted(t *testing.T) {
	tests := []struct {
		name string
		fn   func(*vendor.Parser) error
	}{
		{"CNAMEResource", func(p *vendor.Parser) error { _, err := p.CNAMEResource(); return err }},
		{"MXResource", func(p *vendor.Parser) error { _, err := p.MXResource(); return err }},
		{"NSResource", func(p *vendor.Parser) error { _, err := p.NSResource(); return err }},
		{"PTRResource", func(p *vendor.Parser) error { _, err := p.PTRResource(); return err }},
		{"SOAResource", func(p *vendor.Parser) error { _, err := p.SOAResource(); return err }},
		{"TXTResource", func(p *vendor.Parser) error { _, err := p.TXTResource(); return err }},
		{"SRVResource", func(p *vendor.Parser) error { _, err := p.SRVResource(); return err }},
		{"AResource", func(p *vendor.Parser) error { _, err := p.AResource(); return err }},
		{"AAAAResource", func(p *vendor.Parser) error { _, err := p.AAAAResource(); return err }},
	}

	for _, test := range tests {
		if err := test.fn(&vendor.Parser{}); err != vendor.ErrNotStarted {
			t.Errorf("got Parser.%s() = _ , %v, want = _, %v", test.name, err, vendor.ErrNotStarted)
		}
	}
}

func TestDNSPackUnpack(t *testing.T) {
	wants := []vendor.Message{
		{
			Questions: []vendor.Question{
				{
					Name:  vendor.MustNewName("."),
					Type:  vendor.TypeAAAA,
					Class: vendor.ClassINET,
				},
			},
			Answers:     []vendor.Resource{},
			Authorities: []vendor.Resource{},
			Additionals: []vendor.Resource{},
		},
		largeTestMsg(),
	}
	for i, want := range wants {
		b, err := want.Pack()
		if err != nil {
			t.Fatalf("%d: Message.Pack() = %v", i, err)
		}
		var got vendor.Message
		err = got.Unpack(b)
		if err != nil {
			t.Fatalf("%d: Message.Unapck() = %v", i, err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%d: Message.Pack/Unpack() roundtrip: got = %+v, want = %+v", i, &got, &want)
		}
	}
}

func TestDNSAppendPackUnpack(t *testing.T) {
	wants := []vendor.Message{
		{
			Questions: []vendor.Question{
				{
					Name:  vendor.MustNewName("."),
					Type:  vendor.TypeAAAA,
					Class: vendor.ClassINET,
				},
			},
			Answers:     []vendor.Resource{},
			Authorities: []vendor.Resource{},
			Additionals: []vendor.Resource{},
		},
		largeTestMsg(),
	}
	for i, want := range wants {
		b := make([]byte, 2, 514)
		b, err := want.AppendPack(b)
		if err != nil {
			t.Fatalf("%d: Message.AppendPack() = %v", i, err)
		}
		b = b[2:]
		var got vendor.Message
		err = got.Unpack(b)
		if err != nil {
			t.Fatalf("%d: Message.Unapck() = %v", i, err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%d: Message.AppendPack/Unpack() roundtrip: got = %+v, want = %+v", i, &got, &want)
		}
	}
}

func TestSkipAll(t *testing.T) {
	msg := largeTestMsg()
	buf, err := msg.Pack()
	if err != nil {
		t.Fatal("Message.Pack() =", err)
	}
	var p vendor.Parser
	if _, err := p.Start(buf); err != nil {
		t.Fatal("Parser.Start(non-nil) =", err)
	}

	tests := []struct {
		name string
		f    func() error
	}{
		{"SkipAllQuestions", p.SkipAllQuestions},
		{"SkipAllAnswers", p.SkipAllAnswers},
		{"SkipAllAuthorities", p.SkipAllAuthorities},
		{"SkipAllAdditionals", p.SkipAllAdditionals},
	}
	for _, test := range tests {
		for i := 1; i <= 3; i++ {
			if err := test.f(); err != nil {
				t.Errorf("%d: Parser.%s() = %v", i, test.name, err)
			}
		}
	}
}

func TestSkipEach(t *testing.T) {
	msg := smallTestMsg()

	buf, err := msg.Pack()
	if err != nil {
		t.Fatal("Message.Pack() =", err)
	}
	var p vendor.Parser
	if _, err := p.Start(buf); err != nil {
		t.Fatal("Parser.Start(non-nil) =", err)
	}

	tests := []struct {
		name string
		f    func() error
	}{
		{"SkipQuestion", p.SkipQuestion},
		{"SkipAnswer", p.SkipAnswer},
		{"SkipAuthority", p.SkipAuthority},
		{"SkipAdditional", p.SkipAdditional},
	}
	for _, test := range tests {
		if err := test.f(); err != nil {
			t.Errorf("first Parser.%s() = %v, want = nil", test.name, err)
		}
		if err := test.f(); err != vendor.ErrSectionDone {
			t.Errorf("second Parser.%s() = %v, want = %v", test.name, err, vendor.ErrSectionDone)
		}
	}
}

func TestSkipAfterRead(t *testing.T) {
	msg := smallTestMsg()

	buf, err := msg.Pack()
	if err != nil {
		t.Fatal("Message.Pack() =", err)
	}
	var p vendor.Parser
	if _, err := p.Start(buf); err != nil {
		t.Fatal("Parser.Srart(non-nil) =", err)
	}

	tests := []struct {
		name string
		skip func() error
		read func() error
	}{
		{"Question", p.SkipQuestion, func() error { _, err := p.Question(); return err }},
		{"Answer", p.SkipAnswer, func() error { _, err := p.Answer(); return err }},
		{"Authority", p.SkipAuthority, func() error { _, err := p.Authority(); return err }},
		{"Additional", p.SkipAdditional, func() error { _, err := p.Additional(); return err }},
	}
	for _, test := range tests {
		if err := test.read(); err != nil {
			t.Errorf("got Parser.%s() = _, %v, want = _, nil", test.name, err)
		}
		if err := test.skip(); err != vendor.ErrSectionDone {
			t.Errorf("got Parser.Skip%s() = %v, want = %v", test.name, err, vendor.ErrSectionDone)
		}
	}
}

func TestSkipNotStarted(t *testing.T) {
	var p vendor.Parser

	tests := []struct {
		name string
		f    func() error
	}{
		{"SkipAllQuestions", p.SkipAllQuestions},
		{"SkipAllAnswers", p.SkipAllAnswers},
		{"SkipAllAuthorities", p.SkipAllAuthorities},
		{"SkipAllAdditionals", p.SkipAllAdditionals},
	}
	for _, test := range tests {
		if err := test.f(); err != vendor.ErrNotStarted {
			t.Errorf("got Parser.%s() = %v, want = %v", test.name, err, vendor.ErrNotStarted)
		}
	}
}

func TestTooManyRecords(t *testing.T) {
	const recs = int(^uint16(0)) + 1
	tests := []struct {
		name string
		msg  vendor.Message
		want error
	}{
		{
			"Questions",
			vendor.Message{
				Questions: make([]vendor.Question, recs),
			},
			vendor.errTooManyQuestions,
		},
		{
			"Answers",
			vendor.Message{
				Answers: make([]vendor.Resource, recs),
			},
			vendor.errTooManyAnswers,
		},
		{
			"Authorities",
			vendor.Message{
				Authorities: make([]vendor.Resource, recs),
			},
			vendor.errTooManyAuthorities,
		},
		{
			"Additionals",
			vendor.Message{
				Additionals: make([]vendor.Resource, recs),
			},
			vendor.errTooManyAdditionals,
		},
	}

	for _, test := range tests {
		if _, got := test.msg.Pack(); got != test.want {
			t.Errorf("got Message.Pack() for %d %s = %v, want = %v", recs, test.name, got, test.want)
		}
	}
}

func TestVeryLongTxt(t *testing.T) {
	want := vendor.Resource{
		vendor.ResourceHeader{
			Name:  vendor.MustNewName("foo.bar.example.com."),
			Type:  vendor.TypeTXT,
			Class: vendor.ClassINET,
		},
		&vendor.TXTResource{[]string{
			"",
			"",
			"foo bar",
			"",
			"www.example.com",
			"www.example.com.",
			strings.Repeat(".", 255),
		}},
	}
	buf, err := want.pack(make([]byte, 0, 8000), map[string]int{}, 0)
	if err != nil {
		t.Fatal("Resource.pack() =", err)
	}
	var got vendor.Resource
	off, err := got.Header.unpack(buf, 0)
	if err != nil {
		t.Fatal("ResourceHeader.unpack() =", err)
	}
	body, n, err := vendor.unpackResourceBody(buf, off, got.Header)
	if err != nil {
		t.Fatal("unpackResourceBody() =", err)
	}
	got.Body = body
	if n != len(buf) {
		t.Errorf("unpacked different amount than packed: got = %d, want = %d", n, len(buf))
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Resource.pack/unpack() roundtrip: got = %#v, want = %#v", got, want)
	}
}

func TestTooLongTxt(t *testing.T) {
	rb := vendor.TXTResource{[]string{strings.Repeat(".", 256)}}
	if _, err := rb.pack(make([]byte, 0, 8000), map[string]int{}, 0); err != vendor.errStringTooLong {
		t.Errorf("packing TXTResource with 256 character string: got err = %v, want = %v", err, vendor.errStringTooLong)
	}
}

func TestStartAppends(t *testing.T) {
	buf := make([]byte, 2, 514)
	wantBuf := []byte{4, 44}
	copy(buf, wantBuf)

	b := vendor.NewBuilder(buf, vendor.Header{})
	b.EnableCompression()

	buf, err := b.Finish()
	if err != nil {
		t.Fatal("Builder.Finish() =", err)
	}
	if got, want := len(buf), vendor.headerLen+2; got != want {
		t.Errorf("got len(buf) = %d, want = %d", got, want)
	}
	if string(buf[:2]) != string(wantBuf) {
		t.Errorf("original data not preserved, got = %#v, want = %#v", buf[:2], wantBuf)
	}
}

func TestStartError(t *testing.T) {
	tests := []struct {
		name string
		fn   func(*vendor.Builder) error
	}{
		{"Questions", func(b *vendor.Builder) error { return b.StartQuestions() }},
		{"Answers", func(b *vendor.Builder) error { return b.StartAnswers() }},
		{"Authorities", func(b *vendor.Builder) error { return b.StartAuthorities() }},
		{"Additionals", func(b *vendor.Builder) error { return b.StartAdditionals() }},
	}

	envs := []struct {
		name string
		fn   func() *vendor.Builder
		want error
	}{
		{"sectionNotStarted", func() *vendor.Builder { return &vendor.Builder{section: vendor.sectionNotStarted} }, vendor.ErrNotStarted},
		{"sectionDone", func() *vendor.Builder { return &vendor.Builder{section: vendor.sectionDone} }, vendor.ErrSectionDone},
	}

	for _, env := range envs {
		for _, test := range tests {
			if got := test.fn(env.fn()); got != env.want {
				t.Errorf("got Builder{%s}.Start%s() = %v, want = %v", env.name, test.name, got, env.want)
			}
		}
	}
}

func TestBuilderResourceError(t *testing.T) {
	tests := []struct {
		name string
		fn   func(*vendor.Builder) error
	}{
		{"CNAMEResource", func(b *vendor.Builder) error { return b.CNAMEResource(vendor.ResourceHeader{}, vendor.CNAMEResource{}) }},
		{"MXResource", func(b *vendor.Builder) error { return b.MXResource(vendor.ResourceHeader{}, vendor.MXResource{}) }},
		{"NSResource", func(b *vendor.Builder) error { return b.NSResource(vendor.ResourceHeader{}, vendor.NSResource{}) }},
		{"PTRResource", func(b *vendor.Builder) error { return b.PTRResource(vendor.ResourceHeader{}, vendor.PTRResource{}) }},
		{"SOAResource", func(b *vendor.Builder) error { return b.SOAResource(vendor.ResourceHeader{}, vendor.SOAResource{}) }},
		{"TXTResource", func(b *vendor.Builder) error { return b.TXTResource(vendor.ResourceHeader{}, vendor.TXTResource{}) }},
		{"SRVResource", func(b *vendor.Builder) error { return b.SRVResource(vendor.ResourceHeader{}, vendor.SRVResource{}) }},
		{"AResource", func(b *vendor.Builder) error { return b.AResource(vendor.ResourceHeader{}, vendor.AResource{}) }},
		{"AAAAResource", func(b *vendor.Builder) error { return b.AAAAResource(vendor.ResourceHeader{}, vendor.AAAAResource{}) }},
		{"OPTResource", func(b *vendor.Builder) error { return b.OPTResource(vendor.ResourceHeader{}, vendor.OPTResource{}) }},
	}

	envs := []struct {
		name string
		fn   func() *vendor.Builder
		want error
	}{
		{"sectionNotStarted", func() *vendor.Builder { return &vendor.Builder{section: vendor.sectionNotStarted} }, vendor.ErrNotStarted},
		{"sectionHeader", func() *vendor.Builder { return &vendor.Builder{section: vendor.sectionHeader} }, vendor.ErrNotStarted},
		{"sectionQuestions", func() *vendor.Builder { return &vendor.Builder{section: vendor.sectionQuestions} }, vendor.ErrNotStarted},
		{"sectionDone", func() *vendor.Builder { return &vendor.Builder{section: vendor.sectionDone} }, vendor.ErrSectionDone},
	}

	for _, env := range envs {
		for _, test := range tests {
			if got := test.fn(env.fn()); got != env.want {
				t.Errorf("got Builder{%s}.%s() = %v, want = %v", env.name, test.name, got, env.want)
			}
		}
	}
}

func TestFinishError(t *testing.T) {
	var b vendor.Builder
	want := vendor.ErrNotStarted
	if _, got := b.Finish(); got != want {
		t.Errorf("got Builder.Finish() = %v, want = %v", got, want)
	}
}

func TestBuilder(t *testing.T) {
	msg := largeTestMsg()
	want, err := msg.Pack()
	if err != nil {
		t.Fatal("Message.Pack() =", err)
	}

	b := vendor.NewBuilder(nil, msg.Header)
	b.EnableCompression()

	if err := b.StartQuestions(); err != nil {
		t.Fatal("Builder.StartQuestions() =", err)
	}
	for _, q := range msg.Questions {
		if err := b.Question(q); err != nil {
			t.Fatalf("Builder.Question(%#v) = %v", q, err)
		}
	}

	if err := b.StartAnswers(); err != nil {
		t.Fatal("Builder.StartAnswers() =", err)
	}
	for _, a := range msg.Answers {
		switch a.Header.Type {
		case vendor.TypeA:
			if err := b.AResource(a.Header, *a.Body.(*vendor.AResource)); err != nil {
				t.Fatalf("Builder.AResource(%#v) = %v", a, err)
			}
		case vendor.TypeNS:
			if err := b.NSResource(a.Header, *a.Body.(*vendor.NSResource)); err != nil {
				t.Fatalf("Builder.NSResource(%#v) = %v", a, err)
			}
		case vendor.TypeCNAME:
			if err := b.CNAMEResource(a.Header, *a.Body.(*vendor.CNAMEResource)); err != nil {
				t.Fatalf("Builder.CNAMEResource(%#v) = %v", a, err)
			}
		case vendor.TypeSOA:
			if err := b.SOAResource(a.Header, *a.Body.(*vendor.SOAResource)); err != nil {
				t.Fatalf("Builder.SOAResource(%#v) = %v", a, err)
			}
		case vendor.TypePTR:
			if err := b.PTRResource(a.Header, *a.Body.(*vendor.PTRResource)); err != nil {
				t.Fatalf("Builder.PTRResource(%#v) = %v", a, err)
			}
		case vendor.TypeMX:
			if err := b.MXResource(a.Header, *a.Body.(*vendor.MXResource)); err != nil {
				t.Fatalf("Builder.MXResource(%#v) = %v", a, err)
			}
		case vendor.TypeTXT:
			if err := b.TXTResource(a.Header, *a.Body.(*vendor.TXTResource)); err != nil {
				t.Fatalf("Builder.TXTResource(%#v) = %v", a, err)
			}
		case vendor.TypeAAAA:
			if err := b.AAAAResource(a.Header, *a.Body.(*vendor.AAAAResource)); err != nil {
				t.Fatalf("Builder.AAAAResource(%#v) = %v", a, err)
			}
		case vendor.TypeSRV:
			if err := b.SRVResource(a.Header, *a.Body.(*vendor.SRVResource)); err != nil {
				t.Fatalf("Builder.SRVResource(%#v) = %v", a, err)
			}
		}
	}

	if err := b.StartAuthorities(); err != nil {
		t.Fatal("Builder.StartAuthorities() =", err)
	}
	for _, a := range msg.Authorities {
		if err := b.NSResource(a.Header, *a.Body.(*vendor.NSResource)); err != nil {
			t.Fatalf("Builder.NSResource(%#v) = %v", a, err)
		}
	}

	if err := b.StartAdditionals(); err != nil {
		t.Fatal("Builder.StartAdditionals() =", err)
	}
	for _, a := range msg.Additionals {
		switch a.Body.(type) {
		case *vendor.TXTResource:
			if err := b.TXTResource(a.Header, *a.Body.(*vendor.TXTResource)); err != nil {
				t.Fatalf("Builder.TXTResource(%#v) = %v", a, err)
			}
		case *vendor.OPTResource:
			if err := b.OPTResource(a.Header, *a.Body.(*vendor.OPTResource)); err != nil {
				t.Fatalf("Builder.OPTResource(%#v) = %v", a, err)
			}
		}
	}

	got, err := b.Finish()
	if err != nil {
		t.Fatal("Builder.Finish() =", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("got from Builder.Finish() = %#v\nwant = %#v", got, want)
	}
}

func TestResourcePack(t *testing.T) {
	for _, tt := range []struct {
		m   vendor.Message
		err error
	}{
		{
			vendor.Message{
				Questions: []vendor.Question{
					{
						Name:  vendor.MustNewName("."),
						Type:  vendor.TypeAAAA,
						Class: vendor.ClassINET,
					},
				},
				Answers: []vendor.Resource{{vendor.ResourceHeader{}, nil}},
			},
			&vendor.nestedError{"packing Answer", vendor.errNilResouceBody},
		},
		{
			vendor.Message{
				Questions: []vendor.Question{
					{
						Name:  vendor.MustNewName("."),
						Type:  vendor.TypeAAAA,
						Class: vendor.ClassINET,
					},
				},
				Authorities: []vendor.Resource{{vendor.ResourceHeader{}, (*vendor.NSResource)(nil)}},
			},
			&vendor.nestedError{"packing Authority",
				&vendor.nestedError{"ResourceHeader",
					&vendor.nestedError{"Name", vendor.errNonCanonicalName},
				},
			},
		},
		{
			vendor.Message{
				Questions: []vendor.Question{
					{
						Name:  vendor.MustNewName("."),
						Type:  vendor.TypeA,
						Class: vendor.ClassINET,
					},
				},
				Additionals: []vendor.Resource{{vendor.ResourceHeader{}, nil}},
			},
			&vendor.nestedError{"packing Additional", vendor.errNilResouceBody},
		},
	} {
		_, err := tt.m.Pack()
		if !reflect.DeepEqual(err, tt.err) {
			t.Errorf("got Message{%v}.Pack() = %v, want %v", tt.m, err, tt.err)
		}
	}
}

func TestResourcePackLength(t *testing.T) {
	r := vendor.Resource{
		vendor.ResourceHeader{
			Name:  vendor.MustNewName("."),
			Type:  vendor.TypeA,
			Class: vendor.ClassINET,
		},
		&vendor.AResource{[4]byte{127, 0, 0, 2}},
	}

	hb, _, err := r.Header.pack(nil, nil, 0)
	if err != nil {
		t.Fatal("ResourceHeader.pack() =", err)
	}
	buf := make([]byte, 0, len(hb))
	buf, err = r.pack(buf, nil, 0)
	if err != nil {
		t.Fatal("Resource.pack() =", err)
	}

	var hdr vendor.ResourceHeader
	if _, err := hdr.unpack(buf, 0); err != nil {
		t.Fatal("ResourceHeader.unpack() =", err)
	}

	if got, want := int(hdr.Length), len(buf)-len(hb); got != want {
		t.Errorf("got hdr.Length = %d, want = %d", got, want)
	}
}

func TestOptionPackUnpack(t *testing.T) {
	for _, tt := range []struct {
		name     string
		w        []byte // wire format of m.Additionals
		m        vendor.Message
		dnssecOK bool
		extRCode vendor.RCode
	}{
		{
			name: "without EDNS(0) options",
			w: []byte{
				0x00, 0x00, 0x29, 0x10, 0x00, 0xfe, 0x00, 0x80,
				0x00, 0x00, 0x00,
			},
			m: vendor.Message{
				Header: vendor.Header{RCode: vendor.RCodeFormatError},
				Questions: []vendor.Question{
					{
						Name:  vendor.MustNewName("."),
						Type:  vendor.TypeA,
						Class: vendor.ClassINET,
					},
				},
				Additionals: []vendor.Resource{
					{
						mustEDNS0ResourceHeader(4096, 0xfe0|vendor.RCodeFormatError, true),
						&vendor.OPTResource{},
					},
				},
			},
			dnssecOK: true,
			extRCode: 0xfe0 | vendor.RCodeFormatError,
		},
		{
			name: "with EDNS(0) options",
			w: []byte{
				0x00, 0x00, 0x29, 0x10, 0x00, 0xff, 0x00, 0x00,
				0x00, 0x00, 0x0c, 0x00, 0x0c, 0x00, 0x02, 0x00,
				0x00, 0x00, 0x0b, 0x00, 0x02, 0x12, 0x34,
			},
			m: vendor.Message{
				Header: vendor.Header{RCode: vendor.RCodeServerFailure},
				Questions: []vendor.Question{
					{
						Name:  vendor.MustNewName("."),
						Type:  vendor.TypeAAAA,
						Class: vendor.ClassINET,
					},
				},
				Additionals: []vendor.Resource{
					{
						mustEDNS0ResourceHeader(4096, 0xff0|vendor.RCodeServerFailure, false),
						&vendor.OPTResource{
							Options: []vendor.Option{
								{
									Code: 12, // see RFC 7828
									Data: []byte{0x00, 0x00},
								},
								{
									Code: 11, // see RFC 7830
									Data: []byte{0x12, 0x34},
								},
							},
						},
					},
				},
			},
			dnssecOK: false,
			extRCode: 0xff0 | vendor.RCodeServerFailure,
		},
		{
			// Containing multiple OPT resources in a
			// message is invalid, but it's necessary for
			// protocol conformance testing.
			name: "with multiple OPT resources",
			w: []byte{
				0x00, 0x00, 0x29, 0x10, 0x00, 0xff, 0x00, 0x00,
				0x00, 0x00, 0x06, 0x00, 0x0b, 0x00, 0x02, 0x12,
				0x34, 0x00, 0x00, 0x29, 0x10, 0x00, 0xff, 0x00,
				0x00, 0x00, 0x00, 0x06, 0x00, 0x0c, 0x00, 0x02,
				0x00, 0x00,
			},
			m: vendor.Message{
				Header: vendor.Header{RCode: vendor.RCodeNameError},
				Questions: []vendor.Question{
					{
						Name:  vendor.MustNewName("."),
						Type:  vendor.TypeAAAA,
						Class: vendor.ClassINET,
					},
				},
				Additionals: []vendor.Resource{
					{
						mustEDNS0ResourceHeader(4096, 0xff0|vendor.RCodeNameError, false),
						&vendor.OPTResource{
							Options: []vendor.Option{
								{
									Code: 11, // see RFC 7830
									Data: []byte{0x12, 0x34},
								},
							},
						},
					},
					{
						mustEDNS0ResourceHeader(4096, 0xff0|vendor.RCodeNameError, false),
						&vendor.OPTResource{
							Options: []vendor.Option{
								{
									Code: 12, // see RFC 7828
									Data: []byte{0x00, 0x00},
								},
							},
						},
					},
				},
			},
		},
	} {
		w, err := tt.m.Pack()
		if err != nil {
			t.Errorf("Message.Pack() for %s = %v", tt.name, err)
			continue
		}
		if !bytes.Equal(w[len(w)-len(tt.w):], tt.w) {
			t.Errorf("got Message.Pack() for %s = %#v, want %#v", tt.name, w[len(w)-len(tt.w):], tt.w)
			continue
		}
		var m vendor.Message
		if err := m.Unpack(w); err != nil {
			t.Errorf("Message.Unpack() for %s = %v", tt.name, err)
			continue
		}
		if !reflect.DeepEqual(m.Additionals, tt.m.Additionals) {
			t.Errorf("got Message.Pack/Unpack() roundtrip for %s = %+v, want %+v", tt.name, m, tt.m)
			continue
		}
	}
}

// TestGoString tests that Message.GoString produces Go code that compiles to
// reproduce the Message.
//
// This test was produced as follows:
// 1. Run (*Message).GoString on largeTestMsg().
// 2. Remove "dnsmessage." from the output.
// 3. Paste the result in the test to store it in msg.
// 4. Also put the original output in the test to store in want.
func TestGoString(t *testing.T) {
	msg := vendor.Message{Header: vendor.Header{ID: 0, Response: true, OpCode: 0, Authoritative: true, Truncated: false, RecursionDesired: false, RecursionAvailable: false, RCode: vendor.RCodeSuccess}, Questions: []vendor.Question{{Name: vendor.MustNewName("foo.bar.example.com."), Type: vendor.TypeA, Class: vendor.ClassINET}}, Answers: []vendor.Resource{{Header: vendor.ResourceHeader{Name: vendor.MustNewName("foo.bar.example.com."), Type: vendor.TypeA, Class: vendor.ClassINET, TTL: 0, Length: 0}, Body: &vendor.AResource{A: [4]byte{127, 0, 0, 1}}}, {Header: vendor.ResourceHeader{Name: vendor.MustNewName("foo.bar.example.com."), Type: vendor.TypeA, Class: vendor.ClassINET, TTL: 0, Length: 0}, Body: &vendor.AResource{A: [4]byte{127, 0, 0, 2}}}, {Header: vendor.ResourceHeader{Name: vendor.MustNewName("foo.bar.example.com."), Type: vendor.TypeAAAA, Class: vendor.ClassINET, TTL: 0, Length: 0}, Body: &vendor.AAAAResource{AAAA: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}}}, {Header: vendor.ResourceHeader{Name: vendor.MustNewName("foo.bar.example.com."), Type: vendor.TypeCNAME, Class: vendor.ClassINET, TTL: 0, Length: 0}, Body: &vendor.CNAMEResource{CNAME: vendor.MustNewName("alias.example.com.")}}, {Header: vendor.ResourceHeader{Name: vendor.MustNewName("foo.bar.example.com."), Type: vendor.TypeSOA, Class: vendor.ClassINET, TTL: 0, Length: 0}, Body: &vendor.SOAResource{NS: vendor.MustNewName("ns1.example.com."), MBox: vendor.MustNewName("mb.example.com."), Serial: 1, Refresh: 2, Retry: 3, Expire: 4, MinTTL: 5}}, {Header: vendor.ResourceHeader{Name: vendor.MustNewName("foo.bar.example.com."), Type: vendor.TypePTR, Class: vendor.ClassINET, TTL: 0, Length: 0}, Body: &vendor.PTRResource{PTR: vendor.MustNewName("ptr.example.com.")}}, {Header: vendor.ResourceHeader{Name: vendor.MustNewName("foo.bar.example.com."), Type: vendor.TypeMX, Class: vendor.ClassINET, TTL: 0, Length: 0}, Body: &vendor.MXResource{Pref: 7, MX: vendor.MustNewName("mx.example.com.")}}, {Header: vendor.ResourceHeader{Name: vendor.MustNewName("foo.bar.example.com."), Type: vendor.TypeSRV, Class: vendor.ClassINET, TTL: 0, Length: 0}, Body: &vendor.SRVResource{Priority: 8, Weight: 9, Port: 11, Target: vendor.MustNewName("srv.example.com.")}}}, Authorities: []vendor.Resource{{Header: vendor.ResourceHeader{Name: vendor.MustNewName("foo.bar.example.com."), Type: vendor.TypeNS, Class: vendor.ClassINET, TTL: 0, Length: 0}, Body: &vendor.NSResource{NS: vendor.MustNewName("ns1.example.com.")}}, {Header: vendor.ResourceHeader{Name: vendor.MustNewName("foo.bar.example.com."), Type: vendor.TypeNS, Class: vendor.ClassINET, TTL: 0, Length: 0}, Body: &vendor.NSResource{NS: vendor.MustNewName("ns2.example.com.")}}}, Additionals: []vendor.Resource{{Header: vendor.ResourceHeader{Name: vendor.MustNewName("foo.bar.example.com."), Type: vendor.TypeTXT, Class: vendor.ClassINET, TTL: 0, Length: 0}, Body: &vendor.TXTResource{TXT: []string{"So Long\x2c and Thanks for All the Fish"}}}, {Header: vendor.ResourceHeader{Name: vendor.MustNewName("foo.bar.example.com."), Type: vendor.TypeTXT, Class: vendor.ClassINET, TTL: 0, Length: 0}, Body: &vendor.TXTResource{TXT: []string{"Hamster Huey and the Gooey Kablooie"}}}, {Header: vendor.ResourceHeader{Name: vendor.MustNewName("."), Type: vendor.TypeOPT, Class: 4096, TTL: 4261412864, Length: 0}, Body: &vendor.OPTResource{Options: []vendor.Option{{Code: 10, Data: []byte{1, 35, 69, 103, 137, 171, 205, 239}}}}}}}
	if !reflect.DeepEqual(msg, largeTestMsg()) {
		t.Error("Message.GoString lost information or largeTestMsg changed: msg != largeTestMsg()")
	}
	got := msg.GoString()
	want := `dnsmessage.Message{Header: dnsmessage.Header{ID: 0, Response: true, OpCode: 0, Authoritative: true, Truncated: false, RecursionDesired: false, RecursionAvailable: false, RCode: dnsmessage.RCodeSuccess}, Questions: []dnsmessage.Question{dnsmessage.Question{Name: dnsmessage.MustNewName("foo.bar.example.com."), Type: dnsmessage.TypeA, Class: dnsmessage.ClassINET}}, Answers: []dnsmessage.Resource{dnsmessage.Resource{Header: dnsmessage.ResourceHeader{Name: dnsmessage.MustNewName("foo.bar.example.com."), Type: dnsmessage.TypeA, Class: dnsmessage.ClassINET, TTL: 0, Length: 0}, Body: &dnsmessage.AResource{A: [4]byte{127, 0, 0, 1}}}, dnsmessage.Resource{Header: dnsmessage.ResourceHeader{Name: dnsmessage.MustNewName("foo.bar.example.com."), Type: dnsmessage.TypeA, Class: dnsmessage.ClassINET, TTL: 0, Length: 0}, Body: &dnsmessage.AResource{A: [4]byte{127, 0, 0, 2}}}, dnsmessage.Resource{Header: dnsmessage.ResourceHeader{Name: dnsmessage.MustNewName("foo.bar.example.com."), Type: dnsmessage.TypeAAAA, Class: dnsmessage.ClassINET, TTL: 0, Length: 0}, Body: &dnsmessage.AAAAResource{AAAA: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}}}, dnsmessage.Resource{Header: dnsmessage.ResourceHeader{Name: dnsmessage.MustNewName("foo.bar.example.com."), Type: dnsmessage.TypeCNAME, Class: dnsmessage.ClassINET, TTL: 0, Length: 0}, Body: &dnsmessage.CNAMEResource{CNAME: dnsmessage.MustNewName("alias.example.com.")}}, dnsmessage.Resource{Header: dnsmessage.ResourceHeader{Name: dnsmessage.MustNewName("foo.bar.example.com."), Type: dnsmessage.TypeSOA, Class: dnsmessage.ClassINET, TTL: 0, Length: 0}, Body: &dnsmessage.SOAResource{NS: dnsmessage.MustNewName("ns1.example.com."), MBox: dnsmessage.MustNewName("mb.example.com."), Serial: 1, Refresh: 2, Retry: 3, Expire: 4, MinTTL: 5}}, dnsmessage.Resource{Header: dnsmessage.ResourceHeader{Name: dnsmessage.MustNewName("foo.bar.example.com."), Type: dnsmessage.TypePTR, Class: dnsmessage.ClassINET, TTL: 0, Length: 0}, Body: &dnsmessage.PTRResource{PTR: dnsmessage.MustNewName("ptr.example.com.")}}, dnsmessage.Resource{Header: dnsmessage.ResourceHeader{Name: dnsmessage.MustNewName("foo.bar.example.com."), Type: dnsmessage.TypeMX, Class: dnsmessage.ClassINET, TTL: 0, Length: 0}, Body: &dnsmessage.MXResource{Pref: 7, MX: dnsmessage.MustNewName("mx.example.com.")}}, dnsmessage.Resource{Header: dnsmessage.ResourceHeader{Name: dnsmessage.MustNewName("foo.bar.example.com."), Type: dnsmessage.TypeSRV, Class: dnsmessage.ClassINET, TTL: 0, Length: 0}, Body: &dnsmessage.SRVResource{Priority: 8, Weight: 9, Port: 11, Target: dnsmessage.MustNewName("srv.example.com.")}}}, Authorities: []dnsmessage.Resource{dnsmessage.Resource{Header: dnsmessage.ResourceHeader{Name: dnsmessage.MustNewName("foo.bar.example.com."), Type: dnsmessage.TypeNS, Class: dnsmessage.ClassINET, TTL: 0, Length: 0}, Body: &dnsmessage.NSResource{NS: dnsmessage.MustNewName("ns1.example.com.")}}, dnsmessage.Resource{Header: dnsmessage.ResourceHeader{Name: dnsmessage.MustNewName("foo.bar.example.com."), Type: dnsmessage.TypeNS, Class: dnsmessage.ClassINET, TTL: 0, Length: 0}, Body: &dnsmessage.NSResource{NS: dnsmessage.MustNewName("ns2.example.com.")}}}, Additionals: []dnsmessage.Resource{dnsmessage.Resource{Header: dnsmessage.ResourceHeader{Name: dnsmessage.MustNewName("foo.bar.example.com."), Type: dnsmessage.TypeTXT, Class: dnsmessage.ClassINET, TTL: 0, Length: 0}, Body: &dnsmessage.TXTResource{TXT: []string{"So Long\x2c and Thanks for All the Fish"}}}, dnsmessage.Resource{Header: dnsmessage.ResourceHeader{Name: dnsmessage.MustNewName("foo.bar.example.com."), Type: dnsmessage.TypeTXT, Class: dnsmessage.ClassINET, TTL: 0, Length: 0}, Body: &dnsmessage.TXTResource{TXT: []string{"Hamster Huey and the Gooey Kablooie"}}}, dnsmessage.Resource{Header: dnsmessage.ResourceHeader{Name: dnsmessage.MustNewName("."), Type: dnsmessage.TypeOPT, Class: 4096, TTL: 4261412864, Length: 0}, Body: &dnsmessage.OPTResource{Options: []dnsmessage.Option{dnsmessage.Option{Code: 10, Data: []byte{1, 35, 69, 103, 137, 171, 205, 239}}}}}}}`
	if got != want {
		t.Errorf("got msg1.GoString() = %s\nwant = %s", got, want)
	}
}

func benchmarkParsingSetup() ([]byte, error) {
	name := vendor.MustNewName("foo.bar.example.com.")
	msg := vendor.Message{
		Header: vendor.Header{Response: true, Authoritative: true},
		Questions: []vendor.Question{
			{
				Name:  name,
				Type:  vendor.TypeA,
				Class: vendor.ClassINET,
			},
		},
		Answers: []vendor.Resource{
			{
				vendor.ResourceHeader{
					Name:  name,
					Class: vendor.ClassINET,
				},
				&vendor.AResource{[4]byte{}},
			},
			{
				vendor.ResourceHeader{
					Name:  name,
					Class: vendor.ClassINET,
				},
				&vendor.AAAAResource{[16]byte{}},
			},
			{
				vendor.ResourceHeader{
					Name:  name,
					Class: vendor.ClassINET,
				},
				&vendor.CNAMEResource{name},
			},
			{
				vendor.ResourceHeader{
					Name:  name,
					Class: vendor.ClassINET,
				},
				&vendor.NSResource{name},
			},
		},
	}

	buf, err := msg.Pack()
	if err != nil {
		return nil, fmt.Errorf("Message.Pack() = %v", err)
	}
	return buf, nil
}

func benchmarkParsing(tb testing.TB, buf []byte) {
	var p vendor.Parser
	if _, err := p.Start(buf); err != nil {
		tb.Fatal("Parser.Start(non-nil) =", err)
	}

	for {
		_, err := p.Question()
		if err == vendor.ErrSectionDone {
			break
		}
		if err != nil {
			tb.Fatal("Parser.Question() =", err)
		}
	}

	for {
		h, err := p.AnswerHeader()
		if err == vendor.ErrSectionDone {
			break
		}
		if err != nil {
			tb.Fatal("Parser.AnswerHeader() =", err)
		}

		switch h.Type {
		case vendor.TypeA:
			if _, err := p.AResource(); err != nil {
				tb.Fatal("Parser.AResource() =", err)
			}
		case vendor.TypeAAAA:
			if _, err := p.AAAAResource(); err != nil {
				tb.Fatal("Parser.AAAAResource() =", err)
			}
		case vendor.TypeCNAME:
			if _, err := p.CNAMEResource(); err != nil {
				tb.Fatal("Parser.CNAMEResource() =", err)
			}
		case vendor.TypeNS:
			if _, err := p.NSResource(); err != nil {
				tb.Fatal("Parser.NSResource() =", err)
			}
		case vendor.TypeOPT:
			if _, err := p.OPTResource(); err != nil {
				tb.Fatal("Parser.OPTResource() =", err)
			}
		default:
			tb.Fatalf("got unknown type: %T", h)
		}
	}
}

func BenchmarkParsing(b *testing.B) {
	buf, err := benchmarkParsingSetup()
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchmarkParsing(b, buf)
	}
}

func TestParsingAllocs(t *testing.T) {
	buf, err := benchmarkParsingSetup()
	if err != nil {
		t.Fatal(err)
	}

	if allocs := testing.AllocsPerRun(100, func() { benchmarkParsing(t, buf) }); allocs > 0.5 {
		t.Errorf("allocations during parsing: got = %f, want ~0", allocs)
	}
}

func benchmarkBuildingSetup() (vendor.Name, []byte) {
	name := vendor.MustNewName("foo.bar.example.com.")
	buf := make([]byte, 0, vendor.packStartingCap)
	return name, buf
}

func benchmarkBuilding(tb testing.TB, name vendor.Name, buf []byte) {
	bld := vendor.NewBuilder(buf, vendor.Header{Response: true, Authoritative: true})

	if err := bld.StartQuestions(); err != nil {
		tb.Fatal("Builder.StartQuestions() =", err)
	}
	q := vendor.Question{
		Name:  name,
		Type:  vendor.TypeA,
		Class: vendor.ClassINET,
	}
	if err := bld.Question(q); err != nil {
		tb.Fatalf("Builder.Question(%+v) = %v", q, err)
	}

	hdr := vendor.ResourceHeader{
		Name:  name,
		Class: vendor.ClassINET,
	}
	if err := bld.StartAnswers(); err != nil {
		tb.Fatal("Builder.StartQuestions() =", err)
	}

	ar := vendor.AResource{[4]byte{}}
	if err := bld.AResource(hdr, ar); err != nil {
		tb.Fatalf("Builder.AResource(%+v, %+v) = %v", hdr, ar, err)
	}

	aaar := vendor.AAAAResource{[16]byte{}}
	if err := bld.AAAAResource(hdr, aaar); err != nil {
		tb.Fatalf("Builder.AAAAResource(%+v, %+v) = %v", hdr, aaar, err)
	}

	cnr := vendor.CNAMEResource{name}
	if err := bld.CNAMEResource(hdr, cnr); err != nil {
		tb.Fatalf("Builder.CNAMEResource(%+v, %+v) = %v", hdr, cnr, err)
	}

	nsr := vendor.NSResource{name}
	if err := bld.NSResource(hdr, nsr); err != nil {
		tb.Fatalf("Builder.NSResource(%+v, %+v) = %v", hdr, nsr, err)
	}

	extrc := 0xfe0 | vendor.RCodeNotImplemented
	if err := (&hdr).SetEDNS0(4096, extrc, true); err != nil {
		tb.Fatalf("ResourceHeader.SetEDNS0(4096, %#x, true) = %v", extrc, err)
	}
	optr := vendor.OPTResource{}
	if err := bld.OPTResource(hdr, optr); err != nil {
		tb.Fatalf("Builder.OPTResource(%+v, %+v) = %v", hdr, optr, err)
	}

	if _, err := bld.Finish(); err != nil {
		tb.Fatal("Builder.Finish() =", err)
	}
}

func BenchmarkBuilding(b *testing.B) {
	name, buf := benchmarkBuildingSetup()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchmarkBuilding(b, name, buf)
	}
}

func TestBuildingAllocs(t *testing.T) {
	name, buf := benchmarkBuildingSetup()
	if allocs := testing.AllocsPerRun(100, func() { benchmarkBuilding(t, name, buf) }); allocs > 0.5 {
		t.Errorf("allocations during building: got = %f, want ~0", allocs)
	}
}

func smallTestMsg() vendor.Message {
	name := vendor.MustNewName("example.com.")
	return vendor.Message{
		Header: vendor.Header{Response: true, Authoritative: true},
		Questions: []vendor.Question{
			{
				Name:  name,
				Type:  vendor.TypeA,
				Class: vendor.ClassINET,
			},
		},
		Answers: []vendor.Resource{
			{
				vendor.ResourceHeader{
					Name:  name,
					Type:  vendor.TypeA,
					Class: vendor.ClassINET,
				},
				&vendor.AResource{[4]byte{127, 0, 0, 1}},
			},
		},
		Authorities: []vendor.Resource{
			{
				vendor.ResourceHeader{
					Name:  name,
					Type:  vendor.TypeA,
					Class: vendor.ClassINET,
				},
				&vendor.AResource{[4]byte{127, 0, 0, 1}},
			},
		},
		Additionals: []vendor.Resource{
			{
				vendor.ResourceHeader{
					Name:  name,
					Type:  vendor.TypeA,
					Class: vendor.ClassINET,
				},
				&vendor.AResource{[4]byte{127, 0, 0, 1}},
			},
		},
	}
}

func BenchmarkPack(b *testing.B) {
	msg := largeTestMsg()

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if _, err := msg.Pack(); err != nil {
			b.Fatal("Message.Pack() =", err)
		}
	}
}

func BenchmarkAppendPack(b *testing.B) {
	msg := largeTestMsg()
	buf := make([]byte, 0, vendor.packStartingCap)

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if _, err := msg.AppendPack(buf[:0]); err != nil {
			b.Fatal("Message.AppendPack() = ", err)
		}
	}
}

func largeTestMsg() vendor.Message {
	name := vendor.MustNewName("foo.bar.example.com.")
	return vendor.Message{
		Header: vendor.Header{Response: true, Authoritative: true},
		Questions: []vendor.Question{
			{
				Name:  name,
				Type:  vendor.TypeA,
				Class: vendor.ClassINET,
			},
		},
		Answers: []vendor.Resource{
			{
				vendor.ResourceHeader{
					Name:  name,
					Type:  vendor.TypeA,
					Class: vendor.ClassINET,
				},
				&vendor.AResource{[4]byte{127, 0, 0, 1}},
			},
			{
				vendor.ResourceHeader{
					Name:  name,
					Type:  vendor.TypeA,
					Class: vendor.ClassINET,
				},
				&vendor.AResource{[4]byte{127, 0, 0, 2}},
			},
			{
				vendor.ResourceHeader{
					Name:  name,
					Type:  vendor.TypeAAAA,
					Class: vendor.ClassINET,
				},
				&vendor.AAAAResource{[16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}},
			},
			{
				vendor.ResourceHeader{
					Name:  name,
					Type:  vendor.TypeCNAME,
					Class: vendor.ClassINET,
				},
				&vendor.CNAMEResource{vendor.MustNewName("alias.example.com.")},
			},
			{
				vendor.ResourceHeader{
					Name:  name,
					Type:  vendor.TypeSOA,
					Class: vendor.ClassINET,
				},
				&vendor.SOAResource{
					NS:      vendor.MustNewName("ns1.example.com."),
					MBox:    vendor.MustNewName("mb.example.com."),
					Serial:  1,
					Refresh: 2,
					Retry:   3,
					Expire:  4,
					MinTTL:  5,
				},
			},
			{
				vendor.ResourceHeader{
					Name:  name,
					Type:  vendor.TypePTR,
					Class: vendor.ClassINET,
				},
				&vendor.PTRResource{vendor.MustNewName("ptr.example.com.")},
			},
			{
				vendor.ResourceHeader{
					Name:  name,
					Type:  vendor.TypeMX,
					Class: vendor.ClassINET,
				},
				&vendor.MXResource{
					7,
					vendor.MustNewName("mx.example.com."),
				},
			},
			{
				vendor.ResourceHeader{
					Name:  name,
					Type:  vendor.TypeSRV,
					Class: vendor.ClassINET,
				},
				&vendor.SRVResource{
					8,
					9,
					11,
					vendor.MustNewName("srv.example.com."),
				},
			},
		},
		Authorities: []vendor.Resource{
			{
				vendor.ResourceHeader{
					Name:  name,
					Type:  vendor.TypeNS,
					Class: vendor.ClassINET,
				},
				&vendor.NSResource{vendor.MustNewName("ns1.example.com.")},
			},
			{
				vendor.ResourceHeader{
					Name:  name,
					Type:  vendor.TypeNS,
					Class: vendor.ClassINET,
				},
				&vendor.NSResource{vendor.MustNewName("ns2.example.com.")},
			},
		},
		Additionals: []vendor.Resource{
			{
				vendor.ResourceHeader{
					Name:  name,
					Type:  vendor.TypeTXT,
					Class: vendor.ClassINET,
				},
				&vendor.TXTResource{[]string{"So Long, and Thanks for All the Fish"}},
			},
			{
				vendor.ResourceHeader{
					Name:  name,
					Type:  vendor.TypeTXT,
					Class: vendor.ClassINET,
				},
				&vendor.TXTResource{[]string{"Hamster Huey and the Gooey Kablooie"}},
			},
			{
				mustEDNS0ResourceHeader(4096, 0xfe0|vendor.RCodeSuccess, false),
				&vendor.OPTResource{
					Options: []vendor.Option{
						{
							Code: 10, // see RFC 7873
							Data: []byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
						},
					},
				},
			},
		},
	}
}
