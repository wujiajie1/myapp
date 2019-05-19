// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ipv4

import (
	"vendor"
)

func setControlMessage(c *vendor.Conn, opt *vendor.rawOpt, cf vendor.ControlFlags, on bool) error {
	// TODO(mikio): implement this
	return vendor.errNotImplemented
}
