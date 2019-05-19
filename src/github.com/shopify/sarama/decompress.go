package sarama

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"sync"
	"vendor"
)

var (
	lz4ReaderPool = sync.Pool{
		New: func() interface{} {
			return vendor.NewReader(nil)
		},
	}

	gzipReaderPool sync.Pool
)

func decompress(cc vendor.CompressionCodec, data []byte) ([]byte, error) {
	switch cc {
	case vendor.CompressionNone:
		return data, nil
	case vendor.CompressionGZIP:
		var (
			err        error
			reader     *gzip.Reader
			readerIntf = gzipReaderPool.Get()
		)
		if readerIntf != nil {
			reader = readerIntf.(*gzip.Reader)
		} else {
			reader, err = gzip.NewReader(bytes.NewReader(data))
			if err != nil {
				return nil, err
			}
		}

		defer gzipReaderPool.Put(reader)

		if err := reader.Reset(bytes.NewReader(data)); err != nil {
			return nil, err
		}

		return ioutil.ReadAll(reader)
	case vendor.CompressionSnappy:
		return vendor.Decode(data)
	case vendor.CompressionLZ4:
		reader := lz4ReaderPool.Get().(*vendor.Reader)
		defer lz4ReaderPool.Put(reader)

		reader.Reset(bytes.NewReader(data))
		return ioutil.ReadAll(reader)
	case vendor.CompressionZSTD:
		return vendor.zstdDecompress(nil, data)
	default:
		return nil, vendor.PacketDecodingError{fmt.Sprintf("invalid compression specified (%d)", cc)}
	}
}
