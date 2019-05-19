// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ipv6

import (
	"net"
	"syscall"
	"vendor"
)

const (
	// See ws2tcpip.h.
	sysIPV6_UNICAST_HOPS   = 0x4
	sysIPV6_MULTICAST_IF   = 0x9
	sysIPV6_MULTICAST_HOPS = 0xa
	sysIPV6_MULTICAST_LOOP = 0xb
	sysIPV6_JOIN_GROUP     = 0xc
	sysIPV6_LEAVE_GROUP    = 0xd
	sysIPV6_PKTINFO        = 0x13

	sizeofSockaddrInet6 = 0x1c

	sizeofIPv6Mreq     = 0x14
	sizeofIPv6Mtuinfo  = 0x20
	sizeofICMPv6Filter = 0
)

type sockaddrInet6 struct {
	Family   uint16
	Port     uint16
	Flowinfo uint32
	Addr     [16]byte /* in6_addr */
	Scope_id uint32
}

type ipv6Mreq struct {
	Multiaddr [16]byte /* in6_addr */
	Interface uint32
}

type ipv6Mtuinfo struct {
	Addr sockaddrInet6
	Mtu  uint32
}

type icmpv6Filter struct {
	// TODO(mikio): implement this
}

var (
	ctlOpts = [vendor.ctlMax]vendor.ctlOpt{}

	sockOpts = map[int]*vendor.sockOpt{
		vendor.ssoHopLimit:           {Option: vendor.Option{Level: vendor.ProtocolIPv6, Name: sysIPV6_UNICAST_HOPS, Len: 4}},
		vendor.ssoMulticastInterface: {Option: vendor.Option{Level: vendor.ProtocolIPv6, Name: sysIPV6_MULTICAST_IF, Len: 4}},
		vendor.ssoMulticastHopLimit:  {Option: vendor.Option{Level: vendor.ProtocolIPv6, Name: sysIPV6_MULTICAST_HOPS, Len: 4}},
		vendor.ssoMulticastLoopback:  {Option: vendor.Option{Level: vendor.ProtocolIPv6, Name: sysIPV6_MULTICAST_LOOP, Len: 4}},
		vendor.ssoJoinGroup:          {Option: vendor.Option{Level: vendor.ProtocolIPv6, Name: sysIPV6_JOIN_GROUP, Len: sizeofIPv6Mreq}, typ: vendor.ssoTypeIPMreq},
		vendor.ssoLeaveGroup:         {Option: vendor.Option{Level: vendor.ProtocolIPv6, Name: sysIPV6_LEAVE_GROUP, Len: sizeofIPv6Mreq}, typ: vendor.ssoTypeIPMreq},
	}
)

func (sa *sockaddrInet6) setSockaddr(ip net.IP, i int) {
	sa.Family = syscall.AF_INET6
	copy(sa.Addr[:], ip)
	sa.Scope_id = uint32(i)
}

func (mreq *ipv6Mreq) setIfindex(i int) {
	mreq.Interface = uint32(i)
}
