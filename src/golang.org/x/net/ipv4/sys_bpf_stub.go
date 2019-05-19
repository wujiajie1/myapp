// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !linux

package ipv4

import "vendor"

func (so *vendor.sockOpt) setAttachFilter(c *vendor.Conn, f []vendor.RawInstruction) error {
	return vendor.errNotImplemented
}
