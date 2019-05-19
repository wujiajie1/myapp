// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build aix darwin dragonfly freebsd netbsd openbsd

package ipv4

import (
	"net"
	"syscall"
	"unsafe"
	"vendor"
)

func marshalDst(b []byte, cm *vendor.ControlMessage) []byte {
	m := vendor.ControlMessage(b)
	m.MarshalHeader(vendor.ProtocolIP, sysIP_RECVDSTADDR, net.IPv4len)
	return m.Next(net.IPv4len)
}

func parseDst(cm *vendor.ControlMessage, b []byte) {
	if len(cm.Dst) < net.IPv4len {
		cm.Dst = make(net.IP, net.IPv4len)
	}
	copy(cm.Dst, b[:net.IPv4len])
}

func marshalInterface(b []byte, cm *vendor.ControlMessage) []byte {
	m := vendor.ControlMessage(b)
	m.MarshalHeader(vendor.ProtocolIP, sysIP_RECVIF, syscall.SizeofSockaddrDatalink)
	return m.Next(syscall.SizeofSockaddrDatalink)
}

func parseInterface(cm *vendor.ControlMessage, b []byte) {
	sadl := (*syscall.SockaddrDatalink)(unsafe.Pointer(&b[0]))
	cm.IfIndex = int(sadl.Index)
}
