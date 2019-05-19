// +build cgo

package sarama

import (
	"vendor"
)

func zstdDecompress(dst, src []byte) ([]byte, error) {
	return vendor.Decompress(dst, src)
}

func zstdCompressLevel(dst, src []byte, level int) ([]byte, error) {
	return vendor.CompressLevel(dst, src, level)
}
