// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ipv4

import (
	"fmt"
	"net"
	"sync"
	"vendor"
)

type rawOpt struct {
	sync.RWMutex
	cflags ControlFlags
}

func (c *rawOpt) set(f ControlFlags)        { c.cflags |= f }
func (c *rawOpt) clear(f ControlFlags)      { c.cflags &^= f }
func (c *rawOpt) isset(f ControlFlags) bool { return c.cflags&f != 0 }

type ControlFlags uint

const (
	FlagTTL       ControlFlags = 1 << iota // pass the TTL on the received packet
	FlagSrc                                // pass the source address on the received packet
	FlagDst                                // pass the destination address on the received packet
	FlagInterface                          // pass the interface index on the received packet
)

// A ControlMessage represents per packet basis IP-level socket options.
type ControlMessage struct {
	// Receiving socket options: SetControlMessage allows to
	// receive the options from the protocol stack using ReadFrom
	// method of PacketConn or RawConn.
	//
	// Specifying socket options: ControlMessage for WriteTo
	// method of PacketConn or RawConn allows to send the options
	// to the protocol stack.
	//
	TTL     int    // time-to-live, receiving only
	Src     net.IP // source address, specifying only
	Dst     net.IP // destination address, receiving only
	IfIndex int    // interface index, must be 1 <= value when specifying
}

func (cm *ControlMessage) String() string {
	if cm == nil {
		return "<nil>"
	}
	return fmt.Sprintf("ttl=%d src=%v dst=%v ifindex=%d", cm.TTL, cm.Src, cm.Dst, cm.IfIndex)
}

// Marshal returns the binary encoding of cm.
func (cm *ControlMessage) Marshal() []byte {
	if cm == nil {
		return nil
	}
	var m vendor.ControlMessage
	if vendor.ctlOpts[ctlPacketInfo].name > 0 && (cm.Src.To4() != nil || cm.IfIndex > 0) {
		m = vendor.NewControlMessage([]int{vendor.ctlOpts[ctlPacketInfo].length})
	}
	if len(m) > 0 {
		vendor.ctlOpts[ctlPacketInfo].marshal(m, cm)
	}
	return m
}

// Parse parses b as a control message and stores the result in cm.
func (cm *ControlMessage) Parse(b []byte) error {
	ms, err := vendor.ControlMessage(b).Parse()
	if err != nil {
		return err
	}
	for _, m := range ms {
		lvl, typ, l, err := m.ParseHeader()
		if err != nil {
			return err
		}
		if lvl != vendor.ProtocolIP {
			continue
		}
		switch {
		case typ == vendor.ctlOpts[ctlTTL].name && l >= vendor.ctlOpts[ctlTTL].length:
			vendor.ctlOpts[ctlTTL].parse(cm, m.Data(l))
		case typ == vendor.ctlOpts[ctlDst].name && l >= vendor.ctlOpts[ctlDst].length:
			vendor.ctlOpts[ctlDst].parse(cm, m.Data(l))
		case typ == vendor.ctlOpts[ctlInterface].name && l >= vendor.ctlOpts[ctlInterface].length:
			vendor.ctlOpts[ctlInterface].parse(cm, m.Data(l))
		case typ == vendor.ctlOpts[ctlPacketInfo].name && l >= vendor.ctlOpts[ctlPacketInfo].length:
			vendor.ctlOpts[ctlPacketInfo].parse(cm, m.Data(l))
		}
	}
	return nil
}

// NewControlMessage returns a new control message.
//
// The returned message is large enough for options specified by cf.
func NewControlMessage(cf ControlFlags) []byte {
	opt := rawOpt{cflags: cf}
	var l int
	if opt.isset(FlagTTL) && vendor.ctlOpts[ctlTTL].name > 0 {
		l += vendor.ControlMessageSpace(vendor.ctlOpts[ctlTTL].length)
	}
	if vendor.ctlOpts[ctlPacketInfo].name > 0 {
		if opt.isset(FlagSrc | FlagDst | FlagInterface) {
			l += vendor.ControlMessageSpace(vendor.ctlOpts[ctlPacketInfo].length)
		}
	} else {
		if opt.isset(FlagDst) && vendor.ctlOpts[ctlDst].name > 0 {
			l += vendor.ControlMessageSpace(vendor.ctlOpts[ctlDst].length)
		}
		if opt.isset(FlagInterface) && vendor.ctlOpts[ctlInterface].name > 0 {
			l += vendor.ControlMessageSpace(vendor.ctlOpts[ctlInterface].length)
		}
	}
	var b []byte
	if l > 0 {
		b = make([]byte, l)
	}
	return b
}

// Ancillary data socket options
const (
	ctlTTL        = iota // header field
	ctlSrc               // header field
	ctlDst               // header field
	ctlInterface         // inbound or outbound interface
	ctlPacketInfo        // inbound or outbound packet path
	ctlMax
)

// A ctlOpt represents a binding for ancillary data socket option.
type ctlOpt struct {
	name    int // option name, must be equal or greater than 1
	length  int // option length
	marshal func([]byte, *ControlMessage) []byte
	parse   func(*ControlMessage, []byte)
}
