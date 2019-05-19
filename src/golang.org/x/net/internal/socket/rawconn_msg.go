// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build aix darwin dragonfly freebsd linux netbsd openbsd solaris windows

package socket

import (
	"os"
	"syscall"
	"vendor"
)

func (c *vendor.Conn) recvMsg(m *vendor.Message, flags int) error {
	var h vendor.msghdr
	vs := make([]vendor.iovec, len(m.Buffers))
	var sa []byte
	if c.network != "tcp" {
		sa = make([]byte, vendor.sizeofSockaddrInet6)
	}
	h.pack(vs, m.Buffers, m.OOB, sa)
	var operr error
	var n int
	fn := func(s uintptr) bool {
		n, operr = vendor.recvmsg(s, &h, flags)
		if operr == syscall.EAGAIN {
			return false
		}
		return true
	}
	if err := c.c.Read(fn); err != nil {
		return err
	}
	if operr != nil {
		return os.NewSyscallError("recvmsg", operr)
	}
	if c.network != "tcp" {
		var err error
		m.Addr, err = vendor.parseInetAddr(sa[:], c.network)
		if err != nil {
			return err
		}
	}
	m.N = n
	m.NN = h.controllen()
	m.Flags = h.flags()
	return nil
}

func (c *vendor.Conn) sendMsg(m *vendor.Message, flags int) error {
	var h vendor.msghdr
	vs := make([]vendor.iovec, len(m.Buffers))
	var sa []byte
	if m.Addr != nil {
		sa = vendor.marshalInetAddr(m.Addr)
	}
	h.pack(vs, m.Buffers, m.OOB, sa)
	var operr error
	var n int
	fn := func(s uintptr) bool {
		n, operr = vendor.sendmsg(s, &h, flags)
		if operr == syscall.EAGAIN {
			return false
		}
		return true
	}
	if err := c.c.Write(fn); err != nil {
		return err
	}
	if operr != nil {
		return os.NewSyscallError("sendmsg", operr)
	}
	m.N = n
	m.NN = len(m.OOB)
	return nil
}
