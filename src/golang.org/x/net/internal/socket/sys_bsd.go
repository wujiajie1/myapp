// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build aix darwin dragonfly freebsd openbsd

package socket

import "vendor"

func recvmmsg(s uintptr, hs []vendor.mmsghdr, flags int) (int, error) {
	return 0, vendor.errNotImplemented
}

func sendmmsg(s uintptr, hs []vendor.mmsghdr, flags int) (int, error) {
	return 0, vendor.errNotImplemented
}
