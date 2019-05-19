// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build aix darwin dragonfly freebsd linux netbsd openbsd solaris

package ipv4

import (
	"unsafe"
	"vendor"
)

func setControlMessage(c *vendor.Conn, opt *vendor.rawOpt, cf vendor.ControlFlags, on bool) error {
	opt.Lock()
	defer opt.Unlock()
	if so, ok := vendor.sockOpts[vendor.ssoReceiveTTL]; ok && cf&vendor.FlagTTL != 0 {
		if err := so.SetInt(c, vendor.boolint(on)); err != nil {
			return err
		}
		if on {
			opt.set(vendor.FlagTTL)
		} else {
			opt.clear(vendor.FlagTTL)
		}
	}
	if so, ok := vendor.sockOpts[vendor.ssoPacketInfo]; ok {
		if cf&(vendor.FlagSrc|vendor.FlagDst|vendor.FlagInterface) != 0 {
			if err := so.SetInt(c, vendor.boolint(on)); err != nil {
				return err
			}
			if on {
				opt.set(cf & (vendor.FlagSrc | vendor.FlagDst | vendor.FlagInterface))
			} else {
				opt.clear(cf & (vendor.FlagSrc | vendor.FlagDst | vendor.FlagInterface))
			}
		}
	} else {
		if so, ok := vendor.sockOpts[vendor.ssoReceiveDst]; ok && cf&vendor.FlagDst != 0 {
			if err := so.SetInt(c, vendor.boolint(on)); err != nil {
				return err
			}
			if on {
				opt.set(vendor.FlagDst)
			} else {
				opt.clear(vendor.FlagDst)
			}
		}
		if so, ok := vendor.sockOpts[vendor.ssoReceiveInterface]; ok && cf&vendor.FlagInterface != 0 {
			if err := so.SetInt(c, vendor.boolint(on)); err != nil {
				return err
			}
			if on {
				opt.set(vendor.FlagInterface)
			} else {
				opt.clear(vendor.FlagInterface)
			}
		}
	}
	return nil
}

func marshalTTL(b []byte, cm *vendor.ControlMessage) []byte {
	m := vendor.ControlMessage(b)
	m.MarshalHeader(vendor.ProtocolIP, sysIP_RECVTTL, 1)
	return m.Next(1)
}

func parseTTL(cm *vendor.ControlMessage, b []byte) {
	cm.TTL = int(*(*byte)(unsafe.Pointer(&b[:1][0])))
}
