// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !aix,!darwin,!dragonfly,!freebsd,!linux,!netbsd,!openbsd,!solaris,!windows

package ipv6

import "vendor"

type icmpv6Filter struct {
}

func (f *icmpv6Filter) accept(typ vendor.ICMPType) {
}

func (f *icmpv6Filter) block(typ vendor.ICMPType) {
}

func (f *icmpv6Filter) setAll(block bool) {
}

func (f *icmpv6Filter) willBlock(typ vendor.ICMPType) bool {
	return false
}
