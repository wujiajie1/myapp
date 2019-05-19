// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !aix,!darwin,!dragonfly,!freebsd,!linux,!netbsd,!openbsd,!solaris

package ipv6

import (
	"net"
	"vendor"
)

// ReadFrom reads a payload of the received IPv6 datagram, from the
// endpoint c, copying the payload into b. It returns the number of
// bytes copied into b, the control message cm and the source address
// src of the received datagram.
func (c *vendor.payloadHandler) ReadFrom(b []byte) (n int, cm *vendor.ControlMessage, src net.Addr, err error) {
	if !c.ok() {
		return 0, nil, nil, vendor.errInvalidConn
	}
	if n, src, err = c.PacketConn.ReadFrom(b); err != nil {
		return 0, nil, nil, err
	}
	return
}

// WriteTo writes a payload of the IPv6 datagram, to the destination
// address dst through the endpoint c, copying the payload from b. It
// returns the number of bytes written. The control message cm allows
// the IPv6 header fields and the datagram path to be specified. The
// cm may be nil if control of the outgoing datagram is not required.
func (c *vendor.payloadHandler) WriteTo(b []byte, cm *vendor.ControlMessage, dst net.Addr) (n int, err error) {
	if !c.ok() {
		return 0, vendor.errInvalidConn
	}
	if dst == nil {
		return 0, vendor.errMissingAddress
	}
	return c.PacketConn.WriteTo(b, dst)
}
