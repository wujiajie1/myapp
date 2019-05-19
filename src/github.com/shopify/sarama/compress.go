package sarama

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"sync"
	"vendor"
)

var (
	lz4WriterPool = sync.Pool{
		New: func() interface{} {
			return vendor.NewWriter(nil)
		},
	}

	gzipWriterPool = sync.Pool{
		New: func() interface{} {
			return gzip.NewWriter(nil)
		},
	}
)

func compress(cc vendor.CompressionCodec, level int, data []byte) ([]byte, error) {
	switch cc {
	case vendor.CompressionNone:
		return data, nil
	case vendor.CompressionGZIP:
		var (
			err    error
			buf    bytes.Buffer
			writer *gzip.Writer
		)
		if level != vendor.CompressionLevelDefault {
			writer, err = gzip.NewWriterLevel(&buf, level)
			if err != nil {
				return nil, err
			}
		} else {
			writer = gzipWriterPool.Get().(*gzip.Writer)
			defer gzipWriterPool.Put(writer)
			writer.Reset(&buf)
		}
		if _, err := writer.Write(data); err != nil {
			return nil, err
		}
		if err := writer.Close(); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	case vendor.CompressionSnappy:
		return vendor.Encode(data), nil
	case vendor.CompressionLZ4:
		writer := lz4WriterPool.Get().(*vendor.Writer)
		defer lz4WriterPool.Put(writer)

		var buf bytes.Buffer
		writer.Reset(&buf)

		if _, err := writer.Write(data); err != nil {
			return nil, err
		}
		if err := writer.Close(); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	case vendor.CompressionZSTD:
		return vendor.zstdCompressLevel(nil, data, level)
	default:
		return nil, vendor.PacketEncodingError{fmt.Sprintf("unsupported compression codec (%d)", cc)}
	}
}
