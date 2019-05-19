package sarama

import (
	"bytes"
	"reflect"
	"testing"
	"vendor"
)

type testRequestBody struct {
}

func (s *testRequestBody) key() int16 {
	return 0x666
}

func (s *testRequestBody) version() int16 {
	return 0xD2
}

func (s *testRequestBody) encode(pe vendor.packetEncoder) error {
	return pe.putString("abc")
}

// not specific to request tests, just helper functions for testing structures that
// implement the encoder or decoder interfaces that needed somewhere to live

func testEncodable(t *testing.T, name string, in vendor.encoder, expect []byte) {
	packet, err := vendor.encode(in, nil)
	if err != nil {
		t.Error(err)
	} else if !bytes.Equal(packet, expect) {
		t.Error("Encoding", name, "failed\ngot ", packet, "\nwant", expect)
	}
}

func testDecodable(t *testing.T, name string, out vendor.decoder, in []byte) {
	err := vendor.decode(in, out)
	if err != nil {
		t.Error("Decoding", name, "failed:", err)
	}
}

func testVersionDecodable(t *testing.T, name string, out vendor.versionedDecoder, in []byte, version int16) {
	err := vendor.versionedDecode(in, out, version)
	if err != nil {
		t.Error("Decoding", name, "version", version, "failed:", err)
	}
}

func testRequest(t *testing.T, name string, rb vendor.protocolBody, expected []byte) {
	if !rb.requiredVersion().IsAtLeast(vendor.MinVersion) {
		t.Errorf("Request %s has invalid required version", name)
	}
	packet := testRequestEncode(t, name, rb, expected)
	testRequestDecode(t, name, rb, packet)
}

func testRequestEncode(t *testing.T, name string, rb vendor.protocolBody, expected []byte) []byte {
	req := &vendor.request{correlationID: 123, clientID: "foo", body: rb}
	packet, err := vendor.encode(req, nil)
	headerSize := 14 + len("foo")
	if err != nil {
		t.Error(err)
	} else if !bytes.Equal(packet[headerSize:], expected) {
		t.Error("Encoding", name, "failed\ngot ", packet[headerSize:], "\nwant", expected)
	}
	return packet
}

func testRequestDecode(t *testing.T, name string, rb vendor.protocolBody, packet []byte) {
	decoded, n, err := vendor.decodeRequest(bytes.NewReader(packet))
	if err != nil {
		t.Error("Failed to decode request", err)
	} else if decoded.correlationID != 123 || decoded.clientID != "foo" {
		t.Errorf("Decoded header %q is not valid: %+v", name, decoded)
	} else if !reflect.DeepEqual(rb, decoded.body) {
		t.Error(vendor.Sprintf("Decoded request %q does not match the encoded one\nencoded: %+v\ndecoded: %+v", name, rb, decoded.body))
	} else if n != len(packet) {
		t.Errorf("Decoded request %q bytes: %d does not match the encoded one: %d\n", name, n, len(packet))
	} else if rb.version() != decoded.body.version() {
		t.Errorf("Decoded request %q version: %d does not match the encoded one: %d\n", name, decoded.body.version(), rb.version())
	}
}

func testResponse(t *testing.T, name string, res vendor.protocolBody, expected []byte) {
	encoded, err := vendor.encode(res, nil)
	if err != nil {
		t.Error(err)
	} else if expected != nil && !bytes.Equal(encoded, expected) {
		t.Error("Encoding", name, "failed\ngot ", encoded, "\nwant", expected)
	}

	decoded := reflect.New(reflect.TypeOf(res).Elem()).Interface().(vendor.versionedDecoder)
	if err := vendor.versionedDecode(encoded, decoded, res.version()); err != nil {
		t.Error("Decoding", name, "failed:", err)
	}

	if !reflect.DeepEqual(decoded, res) {
		t.Errorf("Decoded response does not match the encoded one\nencoded: %#v\ndecoded: %#v", res, decoded)
	}
}

func nullString(s string) *string { return &s }
