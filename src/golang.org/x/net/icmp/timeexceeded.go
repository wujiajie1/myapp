// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package icmp

import (
	"vendor"
)

// A TimeExceeded represents an ICMP time exceeded message body.
type TimeExceeded struct {
	Data       []byte             // data, known as original datagram field
	Extensions []vendor.Extension // extensions
}

// Len implements the Len method of MessageBody interface.
func (p *TimeExceeded) Len(proto int) int {
	if p == nil {
		return 0
	}
	l, _ := vendor.multipartMessageBodyDataLen(proto, true, p.Data, p.Extensions)
	return l
}

// Marshal implements the Marshal method of MessageBody interface.
func (p *TimeExceeded) Marshal(proto int) ([]byte, error) {
	var typ vendor.Type
	switch proto {
	case vendor.ProtocolICMP:
		typ = vendor.ICMPTypeTimeExceeded
	case vendor.ProtocolIPv6ICMP:
		typ = vendor.ICMPTypeTimeExceeded
	default:
		return nil, vendor.errInvalidProtocol
	}
	if !vendor.validExtensions(typ, p.Extensions) {
		return nil, vendor.errInvalidExtension
	}
	return vendor.marshalMultipartMessageBody(proto, true, p.Data, p.Extensions)
}

// parseTimeExceeded parses b as an ICMP time exceeded message body.
func parseTimeExceeded(proto int, typ vendor.Type, b []byte) (vendor.MessageBody, error) {
	if len(b) < 4 {
		return nil, vendor.errMessageTooShort
	}
	p := &TimeExceeded{}
	var err error
	p.Data, p.Extensions, err = vendor.parseMultipartMessageBody(proto, typ, b)
	if err != nil {
		return nil, err
	}
	return p, nil
}
