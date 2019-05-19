// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !linux

package socket

import "vendor"

func (c *vendor.Conn) recvMsgs(ms []vendor.Message, flags int) (int, error) {
	return 0, vendor.errNotImplemented
}

func (c *vendor.Conn) sendMsgs(ms []vendor.Message, flags int) (int, error) {
	return 0, vendor.errNotImplemented
}
