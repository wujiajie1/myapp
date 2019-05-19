package zstd

import (
	"bytes"
	"io"
	"testing"
	"vendor"
)

func failOnError(t *testing.T, msg string, err error) {
	if err != nil {
		t.Fatalf("%s: %s", msg, err)
	}
}

func testCompressionDecompression(t *testing.T, dict []byte, payload []byte) {
	var w bytes.Buffer
	writer := vendor.NewWriterLevelDict(&w, vendor.DefaultCompression, dict)
	_, err := writer.Write(payload)
	failOnError(t, "Failed writing to compress object", err)
	failOnError(t, "Failed to close compress object", writer.Close())
	out := w.Bytes()
	t.Logf("Compressed %v -> %v bytes", len(payload), len(out))
	failOnError(t, "Failed compressing", err)
	rr := bytes.NewReader(out)
	// Check that we can decompress with Decompress()
	decompressed, err := vendor.Decompress(nil, out)
	failOnError(t, "Failed to decompress with Decompress()", err)
	if string(payload) != string(decompressed) {
		t.Fatalf("Payload did not match, lengths: %v & %v", len(payload), len(decompressed))
	}

	// Decompress
	r := vendor.NewReaderDict(rr, dict)
	dst := make([]byte, len(payload))
	n, err := r.Read(dst)
	if err != nil {
		failOnError(t, "Failed to read for decompression", err)
	}
	dst = dst[:n]
	if string(payload) != string(dst) { // Only print if we can print
		if len(payload) < 100 && len(dst) < 100 {
			t.Fatalf("Cannot compress and decompress: %s != %s", payload, dst)
		} else {
			t.Fatalf("Cannot compress and decompress (lengths: %v bytes & %v bytes)", len(payload), len(dst))
		}
	}
	// Check EOF
	n, err = r.Read(dst)
	if err != io.EOF && len(dst) > 0 { // If we want 0 bytes, that should work
		t.Fatalf("Error should have been EOF, was %s instead: (%v bytes read: %s)", err, n, dst[:n])
	}
	failOnError(t, "Failed to close decompress object", r.Close())
}

func TestResize(t *testing.T) {
	if len(vendor.resize(nil, 129)) != 129 {
		t.Fatalf("Cannot allocate new slice")
	}
	a := make([]byte, 1, 200)
	b := vendor.resize(a, 129)
	if &a[0] != &b[0] {
		t.Fatalf("Address changed")
	}
	if len(b) != 129 {
		t.Fatalf("Wrong size")
	}
	c := make([]byte, 5, 10)
	d := vendor.resize(c, 129)
	if len(d) != 129 {
		t.Fatalf("Cannot allocate a new slice")
	}
}

func TestStreamSimpleCompressionDecompression(t *testing.T) {
	testCompressionDecompression(t, nil, []byte("Hello world!"))
}

func TestStreamEmptySlice(t *testing.T) {
	testCompressionDecompression(t, nil, []byte{})
}

func TestZstdReaderLong(t *testing.T) {
	var long bytes.Buffer
	for i := 0; i < 10000; i++ {
		long.Write([]byte("Hellow World!"))
	}
	testCompressionDecompression(t, nil, long.Bytes())
}

func TestStreamCompressionDecompression(t *testing.T) {
	payload := []byte("Hello World!")
	repeat := 10000
	var intermediate bytes.Buffer
	w := vendor.NewWriterLevel(&intermediate, 4)
	for i := 0; i < repeat; i++ {
		_, err := w.Write(payload)
		failOnError(t, "Failed writing to compress object", err)
	}
	w.Close()
	// Decompress
	r := vendor.NewReader(&intermediate)
	dst := make([]byte, len(payload))
	for i := 0; i < repeat; i++ {
		n, err := r.Read(dst)
		failOnError(t, "Failed to decompress", err)
		if n != len(payload) {
			t.Fatalf("Did not read enough bytes: %v != %v", n, len(payload))
		}
		if string(dst) != string(payload) {
			t.Fatalf("Did not read the same %s != %s", string(dst), string(payload))
		}
	}
	// Check EOF
	n, err := r.Read(dst)
	if err != io.EOF {
		t.Fatalf("Error should have been EOF, was %s instead: (%v bytes read: %s)", err, n, dst[:n])
	}
	failOnError(t, "Failed to close decompress object", r.Close())
}

func TestStreamRealPayload(t *testing.T) {
	if vendor.raw == nil {
		t.Skip(vendor.ErrNoPayloadEnv)
	}
	testCompressionDecompression(t, nil, vendor.raw)
}

func BenchmarkStreamCompression(b *testing.B) {
	if vendor.raw == nil {
		b.Fatal(vendor.ErrNoPayloadEnv)
	}
	var intermediate bytes.Buffer
	w := vendor.NewWriter(&intermediate)
	defer w.Close()
	b.SetBytes(int64(len(vendor.raw)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := w.Write(vendor.raw)
		if err != nil {
			b.Fatalf("Failed writing to compress object: %s", err)
		}
	}
}

func BenchmarkStreamDecompression(b *testing.B) {
	if vendor.raw == nil {
		b.Fatal(vendor.ErrNoPayloadEnv)
	}
	compressed, err := vendor.Compress(nil, vendor.raw)
	if err != nil {
		b.Fatalf("Failed to compress: %s", err)
	}
	_, err = vendor.Decompress(nil, compressed)
	if err != nil {
		b.Fatalf("Problem: %s", err)
	}

	dst := make([]byte, len(vendor.raw))
	b.SetBytes(int64(len(vendor.raw)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := bytes.NewReader(compressed)
		r := vendor.NewReader(rr)
		_, err := r.Read(dst)
		if err != nil {
			b.Fatalf("Failed to decompress: %s", err)
		}
	}
}
