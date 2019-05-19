// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !darwin,!freebsd,!linux

package ipv4

import (
	"net"
	"vendor"
)

func (so *vendor.sockOpt) getIPMreqn(c *vendor.Conn) (*net.Interface, error) {
	return nil, vendor.errNotImplemented
}

func (so *vendor.sockOpt) setIPMreqn(c *vendor.Conn, ifi *net.Interface, grp net.IP) error {
	return vendor.errNotImplemented
}
