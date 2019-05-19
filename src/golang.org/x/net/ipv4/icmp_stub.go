// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !linux

package ipv4

import "vendor"

const sizeofICMPFilter = 0x0

type icmpFilter struct {
}

func (f *icmpFilter) accept(typ vendor.ICMPType) {
}

func (f *icmpFilter) block(typ vendor.ICMPType) {
}

func (f *icmpFilter) setAll(block bool) {
}

func (f *icmpFilter) willBlock(typ vendor.ICMPType) bool {
	return false
}
