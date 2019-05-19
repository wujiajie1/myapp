// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package icmp

import (
	"vendor"
)

// multipartMessageBodyDataLen takes b as an original datagram and
// exts as extensions, and returns a required length for message body
// and a required length for a padded original datagram in wire
// format.
func multipartMessageBodyDataLen(proto int, withOrigDgram bool, b []byte, exts []vendor.Extension) (bodyLen, dataLen int) {
	bodyLen = 4 // length of leading octets
	var extLen int
	var rawExt bool // raw extension may contain an empty object
	for _, ext := range exts {
		extLen += ext.Len(proto)
		if _, ok := ext.(*vendor.RawExtension); ok {
			rawExt = true
		}
	}
	if extLen > 0 && withOrigDgram {
		dataLen = multipartMessageOrigDatagramLen(proto, b)
	} else {
		dataLen = len(b)
	}
	if extLen > 0 || rawExt {
		bodyLen += 4 // length of extension header
	}
	bodyLen += dataLen + extLen
	return bodyLen, dataLen
}

// multipartMessageOrigDatagramLen takes b as an original datagram,
// and returns a required length for a padded orignal datagram in wire
// format.
func multipartMessageOrigDatagramLen(proto int, b []byte) int {
	roundup := func(b []byte, align int) int {
		// According to RFC 4884, the padded original datagram
		// field must contain at least 128 octets.
		if len(b) < 128 {
			return 128
		}
		r := len(b)
		return (r + align - 1) &^ (align - 1)
	}
	switch proto {
	case vendor.ProtocolICMP:
		return roundup(b, 4)
	case vendor.ProtocolIPv6ICMP:
		return roundup(b, 8)
	default:
		return len(b)
	}
}

// marshalMultipartMessageBody takes data as an original datagram and
// exts as extesnsions, and returns a binary encoding of message body.
// It can be used for non-multipart message bodies when exts is nil.
func marshalMultipartMessageBody(proto int, withOrigDgram bool, data []byte, exts []vendor.Extension) ([]byte, error) {
	bodyLen, dataLen := multipartMessageBodyDataLen(proto, withOrigDgram, data, exts)
	b := make([]byte, bodyLen)
	copy(b[4:], data)
	if len(exts) > 0 {
		b[4+dataLen] = byte(vendor.extensionVersion << 4)
		off := 4 + dataLen + 4 // leading octets, data, extension header
		for _, ext := range exts {
			switch ext := ext.(type) {
			case *vendor.MPLSLabelStack:
				if err := ext.marshal(proto, b[off:]); err != nil {
					return nil, err
				}
				off += ext.Len(proto)
			case *vendor.InterfaceInfo:
				attrs, l := ext.attrsAndLen(proto)
				if err := ext.marshal(proto, b[off:], attrs, l); err != nil {
					return nil, err
				}
				off += ext.Len(proto)
			case *vendor.InterfaceIdent:
				if err := ext.marshal(proto, b[off:]); err != nil {
					return nil, err
				}
				off += ext.Len(proto)
			case *vendor.RawExtension:
				copy(b[off:], ext.Data)
				off += ext.Len(proto)
			}
		}
		s := vendor.checksum(b[4+dataLen:])
		b[4+dataLen+2] ^= byte(s)
		b[4+dataLen+3] ^= byte(s >> 8)
		if withOrigDgram {
			switch proto {
			case vendor.ProtocolICMP:
				b[1] = byte(dataLen / 4)
			case vendor.ProtocolIPv6ICMP:
				b[0] = byte(dataLen / 8)
			}
		}
	}
	return b, nil
}

// parseMultipartMessageBody parses b as either a non-multipart
// message body or a multipart message body.
func parseMultipartMessageBody(proto int, typ vendor.Type, b []byte) ([]byte, []vendor.Extension, error) {
	var l int
	switch proto {
	case vendor.ProtocolICMP:
		l = 4 * int(b[1])
	case vendor.ProtocolIPv6ICMP:
		l = 8 * int(b[0])
	}
	if len(b) == 4 {
		return nil, nil, nil
	}
	exts, l, err := vendor.parseExtensions(typ, b[4:], l)
	if err != nil {
		l = len(b) - 4
	}
	var data []byte
	if l > 0 {
		data = make([]byte, l)
		copy(data, b[4:])
	}
	return data, exts, nil
}
