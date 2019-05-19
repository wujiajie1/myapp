// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Added for go1.11 compatibility
// +build aix

package ipv4

import (
	"net"
	"syscall"
	"vendor"
)

var (
	ctlOpts = [vendor.ctlMax]vendor.ctlOpt{
		vendor.ctlTTL:       {sysIP_RECVTTL, 1, marshalTTL, parseTTL},
		vendor.ctlDst:       {sysIP_RECVDSTADDR, net.IPv4len, marshalDst, parseDst},
		vendor.ctlInterface: {sysIP_RECVIF, syscall.SizeofSockaddrDatalink, marshalInterface, parseInterface},
	}

	sockOpts = map[int]*vendor.sockOpt{
		vendor.ssoTOS:                {Option: vendor.Option{Level: vendor.ProtocolIP, Name: vendor.sysIP_TOS, Len: 4}},
		vendor.ssoTTL:                {Option: vendor.Option{Level: vendor.ProtocolIP, Name: vendor.sysIP_TTL, Len: 4}},
		vendor.ssoMulticastTTL:       {Option: vendor.Option{Level: vendor.ProtocolIP, Name: vendor.sysIP_MULTICAST_TTL, Len: 1}},
		vendor.ssoMulticastInterface: {Option: vendor.Option{Level: vendor.ProtocolIP, Name: vendor.sysIP_MULTICAST_IF, Len: 4}},
		vendor.ssoMulticastLoopback:  {Option: vendor.Option{Level: vendor.ProtocolIP, Name: vendor.sysIP_MULTICAST_LOOP, Len: 1}},
		vendor.ssoReceiveTTL:         {Option: vendor.Option{Level: vendor.ProtocolIP, Name: sysIP_RECVTTL, Len: 4}},
		vendor.ssoReceiveDst:         {Option: vendor.Option{Level: vendor.ProtocolIP, Name: sysIP_RECVDSTADDR, Len: 4}},
		vendor.ssoReceiveInterface:   {Option: vendor.Option{Level: vendor.ProtocolIP, Name: sysIP_RECVIF, Len: 4}},
		vendor.ssoHeaderPrepend:      {Option: vendor.Option{Level: vendor.ProtocolIP, Name: vendor.sysIP_HDRINCL, Len: 4}},
		vendor.ssoJoinGroup:          {Option: vendor.Option{Level: vendor.ProtocolIP, Name: vendor.sysIP_ADD_MEMBERSHIP, Len: vendor.sizeofIPMreq}, typ: vendor.ssoTypeIPMreq},
		vendor.ssoLeaveGroup:         {Option: vendor.Option{Level: vendor.ProtocolIP, Name: vendor.sysIP_DROP_MEMBERSHIP, Len: vendor.sizeofIPMreq}, typ: vendor.ssoTypeIPMreq},
	}
)
