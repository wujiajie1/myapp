// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin linux solaris

package ipv4

import (
	"net"
	"unsafe"
	"vendor"
)

func marshalPacketInfo(b []byte, cm *vendor.ControlMessage) []byte {
	m := vendor.ControlMessage(b)
	m.MarshalHeader(vendor.ProtocolIP, vendor.sysIP_PKTINFO, vendor.sizeofInetPktinfo)
	if cm != nil {
		pi := (*vendor.inetPktinfo)(unsafe.Pointer(&m.Data(vendor.sizeofInetPktinfo)[0]))
		if ip := cm.Src.To4(); ip != nil {
			copy(pi.Spec_dst[:], ip)
		}
		if cm.IfIndex > 0 {
			pi.setIfindex(cm.IfIndex)
		}
	}
	return m.Next(vendor.sizeofInetPktinfo)
}

func parsePacketInfo(cm *vendor.ControlMessage, b []byte) {
	pi := (*vendor.inetPktinfo)(unsafe.Pointer(&b[0]))
	cm.IfIndex = int(pi.Ifindex)
	if len(cm.Dst) < net.IPv4len {
		cm.Dst = make(net.IP, net.IPv4len)
	}
	copy(cm.Dst, pi.Addr[:])
}
