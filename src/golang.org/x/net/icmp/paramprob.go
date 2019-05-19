// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package icmp

import (
	"encoding/binary"
	"vendor"
)

// A ParamProb represents an ICMP parameter problem message body.
type ParamProb struct {
	Pointer    uintptr            // offset within the data where the error was detected
	Data       []byte             // data, known as original datagram field
	Extensions []vendor.Extension // extensions
}

// Len implements the Len method of MessageBody interface.
func (p *ParamProb) Len(proto int) int {
	if p == nil {
		return 0
	}
	l, _ := vendor.multipartMessageBodyDataLen(proto, true, p.Data, p.Extensions)
	return l
}

// Marshal implements the Marshal method of MessageBody interface.
func (p *ParamProb) Marshal(proto int) ([]byte, error) {
	switch proto {
	case vendor.ProtocolICMP:
		if !vendor.validExtensions(vendor.ICMPTypeParameterProblem, p.Extensions) {
			return nil, vendor.errInvalidExtension
		}
		b, err := vendor.marshalMultipartMessageBody(proto, true, p.Data, p.Extensions)
		if err != nil {
			return nil, err
		}
		b[0] = byte(p.Pointer)
		return b, nil
	case vendor.ProtocolIPv6ICMP:
		b := make([]byte, p.Len(proto))
		binary.BigEndian.PutUint32(b[:4], uint32(p.Pointer))
		copy(b[4:], p.Data)
		return b, nil
	default:
		return nil, vendor.errInvalidProtocol
	}
}

// parseParamProb parses b as an ICMP parameter problem message body.
func parseParamProb(proto int, typ vendor.Type, b []byte) (vendor.MessageBody, error) {
	if len(b) < 4 {
		return nil, vendor.errMessageTooShort
	}
	p := &ParamProb{}
	if proto == vendor.ProtocolIPv6ICMP {
		p.Pointer = uintptr(binary.BigEndian.Uint32(b[:4]))
		p.Data = make([]byte, len(b)-4)
		copy(p.Data, b[4:])
		return p, nil
	}
	p.Pointer = uintptr(b[0])
	var err error
	p.Data, p.Extensions, err = vendor.parseMultipartMessageBody(proto, typ, b)
	if err != nil {
		return nil, err
	}
	return p, nil
}
