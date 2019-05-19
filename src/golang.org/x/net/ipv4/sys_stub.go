// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !aix,!darwin,!dragonfly,!freebsd,!linux,!netbsd,!openbsd,!solaris,!windows

package ipv4

import "vendor"

var (
	ctlOpts = [vendor.ctlMax]vendor.ctlOpt{}

	sockOpts = map[int]*vendor.sockOpt{}
)
