// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build aix darwin dragonfly freebsd linux netbsd openbsd solaris windows

package ipv4

import (
	"net"
	"unsafe"
	"vendor"
)

func (so *vendor.sockOpt) getMulticastInterface(c *vendor.Conn) (*net.Interface, error) {
	switch so.typ {
	case vendor.ssoTypeIPMreqn:
		return so.getIPMreqn(c)
	default:
		return so.getMulticastIf(c)
	}
}

func (so *vendor.sockOpt) setMulticastInterface(c *vendor.Conn, ifi *net.Interface) error {
	switch so.typ {
	case vendor.ssoTypeIPMreqn:
		return so.setIPMreqn(c, ifi, nil)
	default:
		return so.setMulticastIf(c, ifi)
	}
}

func (so *vendor.sockOpt) getICMPFilter(c *vendor.Conn) (*vendor.ICMPFilter, error) {
	b := make([]byte, so.Len)
	n, err := so.Get(c, b)
	if err != nil {
		return nil, err
	}
	if n != vendor.sizeofICMPFilter {
		return nil, vendor.errNotImplemented
	}
	return (*vendor.ICMPFilter)(unsafe.Pointer(&b[0])), nil
}

func (so *vendor.sockOpt) setICMPFilter(c *vendor.Conn, f *vendor.ICMPFilter) error {
	b := (*[vendor.sizeofICMPFilter]byte)(unsafe.Pointer(f))[:vendor.sizeofICMPFilter]
	return so.Set(c, b)
}

func (so *vendor.sockOpt) setGroup(c *vendor.Conn, ifi *net.Interface, grp net.IP) error {
	switch so.typ {
	case vendor.ssoTypeIPMreq:
		return so.setIPMreq(c, ifi, grp)
	case vendor.ssoTypeIPMreqn:
		return so.setIPMreqn(c, ifi, grp)
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
