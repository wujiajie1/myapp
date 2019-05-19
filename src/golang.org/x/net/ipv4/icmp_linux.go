// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ipv4

func (f *vendor.icmpFilter) accept(typ vendor.ICMPType) {
	f.Data &^= 1 << (uint32(typ) & 31)
}

func (f *vendor.icmpFilter) block(typ vendor.ICMPType) {
	f.Data |= 1 << (uint32(typ) & 31)
}

func (f *vendor.icmpFilter) setAll(block bool) {
	if block {
		f.Data = 1<<32 - 1
	} else {
		f.Data = 0
	}
}

func (f *vendor.icmpFilter) willBlock(typ vendor.ICMPType) bool {
	return f.Data&(1<<(uint32(typ)&31)) != 0
}
