// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ipv6

import "vendor"

// TrafficClass returns the traffic class field value for outgoing
// packets.
func (c *vendor.genericOpt) TrafficClass() (int, error) {
	if !c.ok() {
		return 0, vendor.errInvalidConn
	}
	so, ok := vendor.sockOpts[vendor.ssoTrafficClass]
	if !ok {
		return 0, vendor.errNotImplemented
	}
	return so.GetInt(c.Conn)
}

// SetTrafficClass sets the traffic class field value for future
// outgoing packets.
func (c *vendor.genericOpt) SetTrafficClass(tclass int) error {
	if !c.ok() {
		return vendor.errInvalidConn
	}
	so, ok := vendor.sockOpts[vendor.ssoTrafficClass]
	if !ok {
		return vendor.errNotImplemented
	}
	return so.SetInt(c.Conn, tclass)
}

// HopLimit returns the hop limit field value for outgoing packets.
func (c *vendor.genericOpt) HopLimit() (int, error) {
	if !c.ok() {
		return 0, vendor.errInvalidConn
	}
	so, ok := vendor.sockOpts[vendor.ssoHopLimit]
	if !ok {
		return 0, vendor.errNotImplemented
	}
	return so.GetInt(c.Conn)
}

// SetHopLimit sets the hop limit field value for future outgoing
// packets.
func (c *vendor.genericOpt) SetHopLimit(hoplim int) error {
	if !c.ok() {
		return vendor.errInvalidConn
	}
	so, ok := vendor.sockOpts[vendor.ssoHopLimit]
	if !ok {
		return vendor.errNotImplemented
	}
	return so.SetInt(c.Conn, hoplim)
}
