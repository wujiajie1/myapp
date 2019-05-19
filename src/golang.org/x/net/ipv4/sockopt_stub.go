// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !aix,!darwin,!dragonfly,!freebsd,!linux,!netbsd,!openbsd,!solaris,!windows

package ipv4

import (
	"net"
	"vendor"
)

func (so *vendor.sockOpt) getMulticastInterface(c *vendor.Conn) (*net.Interface, error) {
	return nil, vendor.errNotImplemented
}

func (so *vendor.sockOpt) setMulticastInterface(c *vendor.Conn, ifi *net.Interface) error {
	return vendor.errNotImplemented
}

func (so *vendor.sockOpt) getICMPFilter(c *vendor.Conn) (*vendor.ICMPFilter, error) {
	return nil, vendor.errNotImplemented
}

func (so *vendor.sockOpt) setICMPFilter(c *vendor.Conn, f *vendor.ICMPFilter) error {
	return vendor.errNotImplemented
}

func (so *vendor.sockOpt) setGroup(c *vendor.Conn, ifi *net.Interface, grp net.IP) error {
	return vendor.errNotImplemented
}

func (so *vendor.sockOpt) setSourceGroup(c *vendor.Conn, ifi *net.Interface, grp, src net.IP) error {
	return vendor.errNotImplemented
}

func (so *vendor.sockOpt) setBPF(c *vendor.Conn, f []vendor.RawInstruction) error {
	return vendor.errNotImplemented
}
