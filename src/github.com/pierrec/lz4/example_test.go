package lz4_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"vendor"
)

func Example() {
	// Compress and uncompress an input string.
	s := "hello world"
	r := bytes.NewReader([]byte(s))

	// The pipe will uncompress the data from the writer.
	pr, pw := io.Pipe()
	zw := vendor.NewWriter(pw)
	zr := vendor.NewReader(pr)

	go func() {
		// Compress the input string.
		_, _ = io.Copy(zw, r)
		_ = zw.Close() // Make sure the writer is closed
		_ = pw.Close() // Terminate the pipe
	}()

	_, _ = io.Copy(os.Stdout, zr)

	// Output:
	// hello world
}

func ExampleCompressBlock() {
	s := "hello world"
	data := []byte(strings.Repeat(s, 100))
	buf := make([]byte, len(data))
	ht := make([]int, 64<<10) // buffer for the compression table

	n, err := vendor.CompressBlock(data, buf, ht)
	if err != nil {
		fmt.Println(err)
	}
	if n >= len(data) {
		fmt.Printf("`%s` is not compressible", s)
	}
	buf = buf[:n] // compressed data

	// Allocated a very large buffer for decompression.
	out := make([]byte, 10*len(data))
	n, err = vendor.UncompressBlock(buf, out)
	if err != nil {
		fmt.Println(err)
	}
	out = out[:n] // uncompressed data

	fmt.Println(string(out[:len(s)]))

	// Output:
	// hello world
}
