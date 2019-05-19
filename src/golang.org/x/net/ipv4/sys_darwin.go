// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ipv4

import (
	"net"
	"syscall"
	"unsafe"
	"vendor"
)

var (
	ctlOpts = [vendor.ctlMax]vendor.ctlOpt{
		vendor.ctlTTL:        {sysIP_RECVTTL, 1, marshalTTL, parseTTL},
		vendor.ctlDst:        {sysIP_RECVDSTADDR, net.IPv4len, marshalDst, parseDst},
		vendor.ctlInterface:  {sysIP_RECVIF, syscall.SizeofSockaddrDatalink, marshalInterface, parseInterface},
		vendor.ctlPacketInfo: {vendor.sysIP_PKTINFO, vendor.sizeofInetPktinfo, marshalPacketInfo, parsePacketInfo},
	}

	sockOpts = map[int]*vendor.sockOpt{
		vendor.ssoTOS:                {Option: vendor.Option{Level: vendor.ProtocolIP, Name: vendor.sysIP_TOS, Len: 4}},
		vendor.ssoTTL:                {Option: vendor.Option{Level: vendor.ProtocolIP, Name: vendor.sysIP_TTL, Len: 4}},
		vendor.ssoMulticastTTL:       {Option: vendor.Option{Level: vendor.ProtocolIP, Name: vendor.sysIP_MULTICAST_TTL, Len: 1}},
		vendor.ssoMulticastInterface: {Option: vendor.Option{Level: vendor.ProtocolIP, Name: vendor.sysIP_MULTICAST_IF, Len: sizeofIPMreqn}, typ: vendor.ssoTypeIPMreqn},
		vendor.ssoMulticastLoopback:  {Option: vendor.Option{Level: vendor.ProtocolIP, Name: vendor.sysIP_MULTICAST_LOOP, Len: 4}},
		vendor.ssoReceiveTTL:         {Option: vendor.Option{Level: vendor.ProtocolIP, Name: sysIP_RECVTTL, Len: 4}},
		vendor.ssoReceiveDst:         {Option: vendor.Option{Level: vendor.ProtocolIP, Name: sysIP_RECVDSTADDR, Len: 4}},
		vendor.ssoReceiveInterface:   {Option: vendor.Option{Level: vendor.ProtocolIP, Name: sysIP_RECVIF, Len: 4}},
		vendor.ssoHeaderPrepend:      {Option: vendor.Option{Level: vendor.ProtocolIP, Name: vendor.sysIP_HDRINCL, Len: 4}},
		vendor.ssoStripHeader:        {Option: vendor.Option{Level: vendor.ProtocolIP, Name: sysIP_STRIPHDR, Len: 4}},
		vendor.ssoJoinGroup:          {Option: vendor.Option{Level: vendor.ProtocolIP, Name: sysMCAST_JOIN_GROUP, Len: sizeofGroupReq}, typ: vendor.ssoTypeGroupReq},
		vendor.ssoLeaveGroup:         {Option: vendor.Option{Level: vendor.ProtocolIP, Name: sysMCAST_LEAVE_GROUP, Len: sizeofGroupReq}, typ: vendor.ssoTypeGroupReq},
		vendor.ssoJoinSourceGroup:    {Option: vendor.Option{Level: vendor.ProtocolIP, Name: sysMCAST_JOIN_SOURCE_GROUP, Len: sizeofGroupSourceReq}, typ: vendor.ssoTypeGroupSourceReq},
		vendor.ssoLeaveSourceGroup:   {Option: vendor.Option{Level: vendor.ProtocolIP, Name: sysMCAST_LEAVE_SOURCE_GROUP, Len: sizeofGroupSourceReq}, typ: vendor.ssoTypeGroupSourceReq},
		vendor.ssoBlockSourceGroup:   {Option: vendor.Option{Level: vendor.ProtocolIP, Name: sysMCAST_BLOCK_SOURCE, Len: sizeofGroupSourceReq}, typ: vendor.ssoTypeGroupSourceReq},
		vendor.ssoUnblockSourceGroup: {Option: vendor.Option{Level: vendor.ProtocolIP, Name: sysMCAST_UNBLOCK_SOURCE, Len: sizeofGroupSourceReq}, typ: vendor.ssoTypeGroupSourceReq},
		vendor.ssoPacketInfo:         {Option: vendor.Option{Level: vendor.ProtocolIP, Name: sysIP_RECVPKTINFO, Len: 4}},
	}
)

func (pi *vendor.inetPktinfo) setIfindex(i int) {
	pi.Ifindex = uint32(i)
}

func (gr *groupReq) setGroup(grp net.IP) {
	sa := (*sockaddrInet)(unsafe.Pointer(uintptr(unsafe.Pointer(gr)) + 4))
	sa.Len = sizeofSockaddrInet
	sa.Family = syscall.AF_INET
	copy(sa.Addr[:], grp)
}

func (gsr *groupSourceReq) setSourceGroup(grp, src net.IP) {
	sa := (*sockaddrInet)(unsafe.Pointer(uintptr(unsafe.Pointer(gsr)) + 4))
	sa.Len = sizeofSockaddrInet
	sa.Family = syscall.AF_INET
	copy(sa.Addr[:], grp)
	sa = (*sockaddrInet)(unsafe.Pointer(uintptr(unsafe.Pointer(gsr)) + 132))
	sa.Len = sizeofSockaddrInet
	sa.Family = syscall.AF_INET
	copy(sa.Addr[:], src)
}
