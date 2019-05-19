// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ipv4

import (
	"net"
	"runtime"
	"strings"
	"syscall"
	"unsafe"
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
		vendor.ssoMulticastLoopback:  {Option: vendor.Option{Level: vendor.ProtocolIP, Name: vendor.sysIP_MULTICAST_LOOP, Len: 4}},
		vendor.ssoReceiveTTL:         {Option: vendor.Option{Level: vendor.ProtocolIP, Name: sysIP_RECVTTL, Len: 4}},
		vendor.ssoReceiveDst:         {Option: vendor.Option{Level: vendor.ProtocolIP, Name: sysIP_RECVDSTADDR, Len: 4}},
		vendor.ssoReceiveInterface:   {Option: vendor.Option{Level: vendor.ProtocolIP, Name: sysIP_RECVIF, Len: 4}},
		vendor.ssoHeaderPrepend:      {Option: vendor.Option{Level: vendor.ProtocolIP, Name: vendor.sysIP_HDRINCL, Len: 4}},
		vendor.ssoJoinGroup:          {Option: vendor.Option{Level: vendor.ProtocolIP, Name: sysMCAST_JOIN_GROUP, Len: sizeofGroupReq}, typ: vendor.ssoTypeGroupReq},
		vendor.ssoLeaveGroup:         {Option: vendor.Option{Level: vendor.ProtocolIP, Name: sysMCAST_LEAVE_GROUP, Len: sizeofGroupReq}, typ: vendor.ssoTypeGroupReq},
		vendor.ssoJoinSourceGroup:    {Option: vendor.Option{Level: vendor.ProtocolIP, Name: sysMCAST_JOIN_SOURCE_GROUP, Len: sizeofGroupSourceReq}, typ: vendor.ssoTypeGroupSourceReq},
		vendor.ssoLeaveSourceGroup:   {Option: vendor.Option{Level: vendor.ProtocolIP, Name: sysMCAST_LEAVE_SOURCE_GROUP, Len: sizeofGroupSourceReq}, typ: vendor.ssoTypeGroupSourceReq},
		vendor.ssoBlockSourceGroup:   {Option: vendor.Option{Level: vendor.ProtocolIP, Name: sysMCAST_BLOCK_SOURCE, Len: sizeofGroupSourceReq}, typ: vendor.ssoTypeGroupSourceReq},
		vendor.ssoUnblockSourceGroup: {Option: vendor.Option{Level: vendor.ProtocolIP, Name: sysMCAST_UNBLOCK_SOURCE, Len: sizeofGroupSourceReq}, typ: vendor.ssoTypeGroupSourceReq},
	}
)

func init() {
	vendor.freebsdVersion, _ = syscall.SysctlUint32("kern.osreldate")
	if vendor.freebsdVersion >= 1000000 {
		sockOpts[vendor.ssoMulticastInterface] = &vendor.sockOpt{Option: vendor.Option{Level: vendor.ProtocolIP, Name: vendor.sysIP_MULTICAST_IF, Len: sizeofIPMreqn}, typ: vendor.ssoTypeIPMreqn}
	}
	if runtime.GOOS == "freebsd" && runtime.GOARCH == "386" {
		archs, _ := syscall.Sysctl("kern.supported_archs")
		for _, s := range strings.Fields(archs) {
			if s == "amd64" {
				vendor.compatFreeBSD32 = true
				break
			}
		}
	}
}

func (gr *groupReq) setGroup(grp net.IP) {
	sa := (*sockaddrInet)(unsafe.Pointer(&gr.Group))
	sa.Len = sizeofSockaddrInet
	sa.Family = syscall.AF_INET
	copy(sa.Addr[:], grp)
}

func (gsr *groupSourceReq) setSourceGroup(grp, src net.IP) {
	sa := (*sockaddrInet)(unsafe.Pointer(&gsr.Group))
	sa.Len = sizeofSockaddrInet
	sa.Family = syscall.AF_INET
	copy(sa.Addr[:], grp)
	sa = (*sockaddrInet)(unsafe.Pointer(&gsr.Source))
	sa.Len = sizeofSockaddrInet
	sa.Family = syscall.AF_INET
	copy(sa.Addr[:], src)
}
