// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build aix darwin dragonfly freebsd linux netbsd openbsd solaris

package ipv6

import (
	"net"
	"unsafe"
	"vendor"
)

func marshalTrafficClass(b []byte, cm *vendor.ControlMessage) []byte {
	m := vendor.ControlMessage(b)
	m.MarshalHeader(vendor.ProtocolIPv6, sysIPV6_TCLASS, 4)
	if cm != nil {
		vendor.NativeEndian.PutUint32(m.Data(4), uint32(cm.TrafficClass))
	}
	return m.Next(4)
}

func parseTrafficClass(cm *vendor.ControlMessage, b []byte) {
	cm.TrafficClass = int(vendor.NativeEndian.Uint32(b[:4]))
}

func marshalHopLimit(b []byte, cm *vendor.ControlMessage) []byte {
	m := vendor.ControlMessage(b)
	m.MarshalHeader(vendor.ProtocolIPv6, sysIPV6_HOPLIMIT, 4)
	if cm != nil {
		vendor.NativeEndian.PutUint32(m.Data(4), uint32(cm.HopLimit))
	}
	return m.Next(4)
}

func parseHopLimit(cm *vendor.ControlMessage, b []byte) {
	cm.HopLimit = int(vendor.NativeEndian.Uint32(b[:4]))
}

func marshalPacketInfo(b []byte, cm *vendor.ControlMessage) []byte {
	m := vendor.ControlMessage(b)
	m.MarshalHeader(vendor.ProtocolIPv6, vendor.sysIPV6_PKTINFO, sizeofInet6Pktinfo)
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

func parsePacketInfo(cm *vendor.ControlMessage, b []byte) {
	pi := (*inet6Pktinfo)(unsafe.Pointer(&b[0]))
	if len(cm.Dst) < net.IPv6len {
		cm.Dst = make(net.IP, net.IPv6len)
	}
	copy(cm.Dst, pi.Addr[:])
	cm.IfIndex = int(pi.Ifindex)
}

func marshalNextHop(b []byte, cm *vendor.ControlMessage) []byte {
	m := vendor.ControlMessage(b)
	m.MarshalHeader(vendor.ProtocolIPv6, sysIPV6_NEXTHOP, vendor.sizeofSockaddrInet6)
	if cm != nil {
		sa := (*vendor.sockaddrInet6)(unsafe.Pointer(&m.Data(vendor.sizeofSockaddrInet6)[0]))
		sa.setSockaddr(cm.NextHop, cm.IfIndex)
	}
	return m.Next(vendor.sizeofSockaddrInet6)
}

func parseNextHop(cm *vendor.ControlMessage, b []byte) {
}

func marshalPathMTU(b []byte, cm *vendor.ControlMessage) []byte {
	m := vendor.ControlMessage(b)
	m.MarshalHeader(vendor.ProtocolIPv6, sysIPV6_PATHMTU, vendor.sizeofIPv6Mtuinfo)
	return m.Next(vendor.sizeofIPv6Mtuinfo)
}

func parsePathMTU(cm *vendor.ControlMessage, b []byte) {
	mi := (*vendor.ipv6Mtuinfo)(unsafe.Pointer(&b[0]))
	if len(cm.Dst) < net.IPv6len {
		cm.Dst = make(net.IP, net.IPv6len)
	}
	copy(cm.Dst, mi.Addr.Addr[:])
	cm.IfIndex = int(mi.Addr.Scope_id)
	cm.MTU = int(mi.Mtu)
}
