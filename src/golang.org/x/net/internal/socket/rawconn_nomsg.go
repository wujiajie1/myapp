// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !aix,!darwin,!dragonfly,!freebsd,!linux,!netbsd,!openbsd,!solaris,!windows

package socket

import "vendor"

func (c *vendor.Conn) recvMsg(m *vendor.Message, flags int) error {
	return vendor.errNotImplemented
}

func (c *vendor.Conn) sendMsg(m *vendor.Message, flags int) error {
	return vendor.errNotImplemented
}
