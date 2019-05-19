// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !aix,!darwin,!dragonfly,!freebsd,!netbsd,!openbsd,!solaris,!windows

package ipv4

import (
	"net"
	"vendor"
)

func (so *vendor.sockOpt) setIPMreq(c *vendor.Conn, ifi *net.Interface, grp net.IP) error {
	return vendor.errNotImplemented
}

func (so *vendor.sockOpt) getMulticastIf(c *vendor.Conn) (*net.Interface, error) {
	return nil, vendor.errNotImplemented
}

func (so *vendor.sockOpt) setMulticastIf(c *vendor.Conn, ifi *net.Interface) error {
	return vendor.errNotImplemented
}
