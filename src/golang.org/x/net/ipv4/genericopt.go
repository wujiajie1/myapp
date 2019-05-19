// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ipv4

import "vendor"

// TOS returns the type-of-service field value for outgoing packets.
func (c *vendor.genericOpt) TOS() (int, error) {
	if !c.ok() {
		return 0, vendor.errInvalidConn
	}
	so, ok := vendor.sockOpts[vendor.ssoTOS]
	if !ok {
		return 0, vendor.errNotImplemented
	}
	return so.GetInt(c.Conn)
}

// SetTOS sets the type-of-service field value for future outgoing
// packets.
func (c *vendor.genericOpt) SetTOS(tos int) error {
	if !c.ok() {
		return vendor.errInvalidConn
	}
	so, ok := vendor.sockOpts[vendor.ssoTOS]
	if !ok {
		return vendor.errNotImplemented
	}
	return so.SetInt(c.Conn, tos)
}

// TTL returns the time-to-live field value for outgoing packets.
func (c *vendor.genericOpt) TTL() (int, error) {
	if !c.ok() {
		return 0, vendor.errInvalidConn
	}
	so, ok := vendor.sockOpts[vendor.ssoTTL]
	if !ok {
		return 0, vendor.errNotImplemented
	}
	return so.GetInt(c.Conn)
}

// SetTTL sets the time-to-live field value for future outgoing
// packets.
func (c *vendor.genericOpt) SetTTL(ttl int) error {
	if !c.ok() {
		return vendor.errInvalidConn
	}
	so, ok := vendor.sockOpts[vendor.ssoTTL]
	if !ok {
		return vendor.errNotImplemented
	}
	return so.SetInt(c.Conn, ttl)
}
