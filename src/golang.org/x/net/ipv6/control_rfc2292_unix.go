// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin

package ipv6

import (
	"unsafe"
	"vendor"
)

func marshal2292HopLimit(b []byte, cm *vendor.ControlMessage) []byte {
	m := vendor.ControlMessage(b)
	m.MarshalHeader(vendor.ProtocolIPv6, sysIPV6_2292HOPLIMIT, 4)
	if cm != nil {
		vendor.NativeEndian.PutUint32(m.Data(4), uint32(cm.HopLimit))
	}
	return m.Next(4)
}

func marshal2292PacketInfo(b []byte, cm *vendor.ControlMessage) []byte {
	m := vendor.ControlMessage(b)
	m.MarshalHeader(vendor.ProtocolIPv6, sysIPV6_2292PKTINFO, sizeofInet6Pktinfo)
	if cm != nil {
		pi := (*inet6Pktinfo)(unsafe.Pointer(&m.Data(sizeofInet6Pktinfo)[0]))
		if ip := cm.Src.To16(); ip != nil && ip.To4() == nil {
			copy(pi.Addr[:], ip)
		}
		if cm.IfIndex > 0 {
			pi.setIfindex(cm.IfIndex)
		}
	}
	return m.Next(sizeofInet6Pktinfo)
}

func marshal2292NextHop(b []byte, cm *vendor.ControlMessage) []byte {
	m := vendor.ControlMessage(b)
	m.MarshalHeader(vendor.ProtocolIPv6, sysIPV6_2292NEXTHOP, vendor.sizeofSockaddrInet6)
	if cm != nil {
		sa := (*vendor.sockaddrInet6)(unsafe.Pointer(&m.Data(vendor.sizeofSockaddrInet6)[0]))
		sa.setSockaddr(cm.NextHop, cm.IfIndex)
	}
	return m.Next(vendor.sizeofSockaddrInet6)
}
