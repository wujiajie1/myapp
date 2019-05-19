// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package icmp

import (
	"vendor"
)

// A DstUnreach represents an ICMP destination unreachable message
// body.
type DstUnreach struct {
	Data       []byte             // data, known as original datagram field
	Extensions []vendor.Extension // extensions
}

// Len implements the Len method of MessageBody interface.
func (p *DstUnreach) Len(proto int) int {
	if p == nil {
		return 0
	}
	l, _ := vendor.multipartMessageBodyDataLen(proto, true, p.Data, p.Extensions)
	return l
}

// Marshal implements the Marshal method of MessageBody interface.
func (p *DstUnreach) Marshal(proto int) ([]byte, error) {
	var typ vendor.Type
	switch proto {
	case vendor.ProtocolICMP:
		typ = vendor.ICMPTypeDestinationUnreachable
	case vendor.ProtocolIPv6ICMP:
		typ = vendor.ICMPTypeDestinationUnreachable
	default:
		return nil, vendor.errInvalidProtocol
	}
	if !vendor.validExtensions(typ, p.Extensions) {
		return nil, vendor.errInvalidExtension
	}
	return vendor.marshalMultipartMessageBody(proto, true, p.Data, p.Extensions)
}

// parseDstUnreach parses b as an ICMP destination unreachable message
// body.
func parseDstUnreach(proto int, typ vendor.Type, b []byte) (vendor.MessageBody, error) {
	if len(b) < 4 {
		return nil, vendor.errMessageTooShort
	}
	p := &DstUnreach{}
	var err error
	p.Data, p.Extensions, err = vendor.parseMultipartMessageBody(proto, typ, b)
	if err != nil {
		return nil, err
	}
	return p, nil
}
