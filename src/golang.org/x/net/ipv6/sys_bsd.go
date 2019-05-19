// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build dragonfly netbsd openbsd

package ipv6

import (
	"net"
	"syscall"
	"vendor"
)

var (
	ctlOpts = [vendor.ctlMax]vendor.ctlOpt{
		vendor.ctlTrafficClass: {sysIPV6_TCLASS, 4, marshalTrafficClass, parseTrafficClass},
		vendor.ctlHopLimit:     {sysIPV6_HOPLIMIT, 4, marshalHopLimit, parseHopLimit},
		vendor.ctlPacketInfo:   {vendor.sysIPV6_PKTINFO, sizeofInet6Pktinfo, marshalPacketInfo, parsePacketInfo},
		vendor.ctlNextHop:      {sysIPV6_NEXTHOP, vendor.sizeofSockaddrInet6, marshalNextHop, parseNextHop},
		vendor.ctlPathMTU:      {sysIPV6_PATHMTU, vendor.sizeofIPv6Mtuinfo, marshalPathMTU, parsePathMTU},
	}

	sockOpts = map[int]*vendor.sockOpt{
		vendor.ssoTrafficClass:        {Option: vendor.Option{Level: vendor.ProtocolIPv6, Name: sysIPV6_TCLASS, Len: 4}},
		vendor.ssoHopLimit:            {Option: vendor.Option{Level: vendor.ProtocolIPv6, Name: vendor.sysIPV6_UNICAST_HOPS, Len: 4}},
		vendor.ssoMulticastInterface:  {Option: vendor.Option{Level: vendor.ProtocolIPv6, Name: vendor.sysIPV6_MULTICAST_IF, Len: 4}},
		vendor.ssoMulticastHopLimit:   {Option: vendor.Option{Level: vendor.ProtocolIPv6, Name: vendor.sysIPV6_MULTICAST_HOPS, Len: 4}},
		vendor.ssoMulticastLoopback:   {Option: vendor.Option{Level: vendor.ProtocolIPv6, Name: vendor.sysIPV6_MULTICAST_LOOP, Len: 4}},
		vendor.ssoReceiveTrafficClass: {Option: vendor.Option{Level: vendor.ProtocolIPv6, Name: sysIPV6_RECVTCLASS, Len: 4}},
		vendor.ssoReceiveHopLimit:     {Option: vendor.Option{Level: vendor.ProtocolIPv6, Name: sysIPV6_RECVHOPLIMIT, Len: 4}},
		vendor.ssoReceivePacketInfo:   {Option: vendor.Option{Level: vendor.ProtocolIPv6, Name: sysIPV6_RECVPKTINFO, Len: 4}},
		vendor.ssoReceivePathMTU:      {Option: vendor.Option{Level: vendor.ProtocolIPv6, Name: sysIPV6_RECVPATHMTU, Len: 4}},
		vendor.ssoPathMTU:             {Option: vendor.Option{Level: vendor.ProtocolIPv6, Name: sysIPV6_PATHMTU, Len: vendor.sizeofIPv6Mtuinfo}},
		vendor.ssoChecksum:            {Option: vendor.Option{Level: vendor.ProtocolIPv6, Name: sysIPV6_CHECKSUM, Len: 4}},
		vendor.ssoICMPFilter:          {Option: vendor.Option{Level: vendor.ProtocolIPv6ICMP, Name: sysICMP6_FILTER, Len: vendor.sizeofICMPv6Filter}},
		vendor.ssoJoinGroup:           {Option: vendor.Option{Level: vendor.ProtocolIPv6, Name: vendor.sysIPV6_JOIN_GROUP, Len: vendor.sizeofIPv6Mreq}, typ: vendor.ssoTypeIPMreq},
		vendor.ssoLeaveGroup:          {Option: vendor.Option{Level: vendor.ProtocolIPv6, Name: vendor.sysIPV6_LEAVE_GROUP, Len: vendor.sizeofIPv6Mreq}, typ: vendor.ssoTypeIPMreq},
	}
)

func (sa *vendor.sockaddrInet6) setSockaddr(ip net.IP, i int) {
	sa.Len = vendor.sizeofSockaddrInet6
	sa.Family = syscall.AF_INET6
	copy(sa.Addr[:], ip)
	sa.Scope_id = uint32(i)
}

func (pi *inet6Pktinfo) setIfindex(i int) {
	pi.Ifindex = uint32(i)
}

func (mreq *vendor.ipv6Mreq) setIfindex(i int) {
	mreq.Interface = uint32(i)
}
