// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build aix darwin dragonfly freebsd linux netbsd openbsd solaris windows

package ipv6

import (
	"net"
	"runtime"
	"unsafe"
	"vendor"
)

func (so *vendor.sockOpt) getMulticastInterface(c *vendor.Conn) (*net.Interface, error) {
	n, err := so.GetInt(c)
	if err != nil {
		return nil, err
	}
	return net.InterfaceByIndex(n)
}

func (so *vendor.sockOpt) setMulticastInterface(c *vendor.Conn, ifi *net.Interface) error {
	var n int
	if ifi != nil {
		n = ifi.Index
	}
	return so.SetInt(c, n)
}

func (so *vendor.sockOpt) getICMPFilter(c *vendor.Conn) (*vendor.ICMPFilter, error) {
	b := make([]byte, so.Len)
	n, err := so.Get(c, b)
	if err != nil {
		return nil, err
	}
	if n != vendor.sizeofICMPv6Filter {
		return nil, vendor.errNotImplemented
	}
	return (*vendor.ICMPFilter)(unsafe.Pointer(&b[0])), nil
}

func (so *vendor.sockOpt) setICMPFilter(c *vendor.Conn, f *vendor.ICMPFilter) error {
	b := (*[vendor.sizeofICMPv6Filter]byte)(unsafe.Pointer(f))[:vendor.sizeofICMPv6Filter]
	return so.Set(c, b)
}

func (so *vendor.sockOpt) getMTUInfo(c *vendor.Conn) (*net.Interface, int, error) {
	b := make([]byte, so.Len)
	n, err := so.Get(c, b)
	if err != nil {
		return nil, 0, err
	}
	if n != vendor.sizeofIPv6Mtuinfo {
		return nil, 0, vendor.errNotImplemented
	}
	mi := (*vendor.ipv6Mtuinfo)(unsafe.Pointer(&b[0]))
	if mi.Addr.Scope_id == 0 || runtime.GOOS == "aix" {
		// AIX kernel might return a wrong address.
		return nil, int(mi.Mtu), nil
	}
	ifi, err := net.InterfaceByIndex(int(mi.Addr.Scope_id))
	if err != nil {
		return nil, 0, err
	}
	return ifi, int(mi.Mtu), nil
}

func (so *vendor.sockOpt) setGroup(c *vendor.Conn, ifi *net.Interface, grp net.IP) error {
	switch so.typ {
	case vendor.ssoTypeIPMreq:
		return so.setIPMreq(c, ifi, grp)
	case vendor.ssoTypeGroupReq:
		return so.setGroupReq(c, ifi, grp)
	default:
		return vendor.errNotImplemented
	}
}

func (so *vendor.sockOpt) setSourceGroup(c *vendor.Conn, ifi *net.Interface, grp, src net.IP) error {
	return so.setGroupSourceReq(c, ifi, grp, src)
}

func (so *vendor.sockOpt) setBPF(c *vendor.Conn, f []vendor.RawInstruction) error {
	return so.setAttachFilter(c, f)
}
