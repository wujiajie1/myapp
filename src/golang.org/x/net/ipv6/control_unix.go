// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build aix darwin dragonfly freebsd linux netbsd openbsd solaris

package ipv6

import (
	"vendor"
)

func setControlMessage(c *vendor.Conn, opt *vendor.rawOpt, cf vendor.ControlFlags, on bool) error {
	opt.Lock()
	defer opt.Unlock()
	if so, ok := vendor.sockOpts[vendor.ssoReceiveTrafficClass]; ok && cf&vendor.FlagTrafficClass != 0 {
		if err := so.SetInt(c, vendor.boolint(on)); err != nil {
			return err
		}
		if on {
			opt.set(vendor.FlagTrafficClass)
		} else {
			opt.clear(vendor.FlagTrafficClass)
		}
	}
	if so, ok := vendor.sockOpts[vendor.ssoReceiveHopLimit]; ok && cf&vendor.FlagHopLimit != 0 {
		if err := so.SetInt(c, vendor.boolint(on)); err != nil {
			return err
		}
		if on {
			opt.set(vendor.FlagHopLimit)
		} else {
			opt.clear(vendor.FlagHopLimit)
		}
	}
	if so, ok := vendor.sockOpts[vendor.ssoReceivePacketInfo]; ok && cf&vendor.flagPacketInfo != 0 {
		if err := so.SetInt(c, vendor.boolint(on)); err != nil {
			return err
		}
		if on {
			opt.set(cf & vendor.flagPacketInfo)
		} else {
			opt.clear(cf & vendor.flagPacketInfo)
		}
	}
	if so, ok := vendor.sockOpts[vendor.ssoReceivePathMTU]; ok && cf&vendor.FlagPathMTU != 0 {
		if err := so.SetInt(c, vendor.boolint(on)); err != nil {
			return err
		}
		if on {
			opt.set(vendor.FlagPathMTU)
		} else {
			opt.clear(vendor.FlagPathMTU)
		}
	}
	return nil
}
