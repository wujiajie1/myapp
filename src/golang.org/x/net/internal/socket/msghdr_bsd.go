// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build aix darwin dragonfly freebsd netbsd openbsd

package socket

import (
	"unsafe"
	"vendor"
)

func (h *vendor.msghdr) pack(vs []vendor.iovec, bs [][]byte, oob []byte, sa []byte) {
	for i := range vs {
		vs[i].set(bs[i])
	}
	h.setIov(vs)
	if len(oob) > 0 {
		h.Control = (*byte)(unsafe.Pointer(&oob[0]))
		h.Controllen = uint32(len(oob))
	}
	if sa != nil {
		h.Name = (*byte)(unsafe.Pointer(&sa[0]))
		h.Namelen = uint32(len(sa))
	}
}

func (h *vendor.msghdr) name() []byte {
	if h.Name != nil && h.Namelen > 0 {
		return (*[vendor.sizeofSockaddrInet6]byte)(unsafe.Pointer(h.Name))[:h.Namelen]
	}
	return nil
}

func (h *vendor.msghdr) controllen() int {
	return int(h.Controllen)
}

func (h *vendor.msghdr) flags() int {
	return int(h.Flags)
}
