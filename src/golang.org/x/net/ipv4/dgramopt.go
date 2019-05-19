// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ipv4

import (
	"net"
	"vendor"
)

// MulticastTTL returns the time-to-live field value for outgoing
// multicast packets.
func (c *vendor.dgramOpt) MulticastTTL() (int, error) {
	if !c.ok() {
		return 0, vendor.errInvalidConn
	}
	so, ok := vendor.sockOpts[vendor.ssoMulticastTTL]
	if !ok {
		return 0, vendor.errNotImplemented
	}
	return so.GetInt(c.Conn)
}

// SetMulticastTTL sets the time-to-live field value for future
// outgoing multicast packets.
func (c *vendor.dgramOpt) SetMulticastTTL(ttl int) error {
	if !c.ok() {
		return vendor.errInvalidConn
	}
	so, ok := vendor.sockOpts[vendor.ssoMulticastTTL]
	if !ok {
		return vendor.errNotImplemented
	}
	return so.SetInt(c.Conn, ttl)
}

// MulticastInterface returns the default interface for multicast
// packet transmissions.
func (c *vendor.dgramOpt) MulticastInterface() (*net.Interface, error) {
	if !c.ok() {
		return nil, vendor.errInvalidConn
	}
	so, ok := vendor.sockOpts[vendor.ssoMulticastInterface]
	if !ok {
		return nil, vendor.errNotImplemented
	}
	return so.getMulticastInterface(c.Conn)
}

// SetMulticastInterface sets the default interface for future
// multicast packet transmissions.
func (c *vendor.dgramOpt) SetMulticastInterface(ifi *net.Interface) error {
	if !c.ok() {
		return vendor.errInvalidConn
	}
	so, ok := vendor.sockOpts[vendor.ssoMulticastInterface]
	if !ok {
		return vendor.errNotImplemented
	}
	return so.setMulticastInterface(c.Conn, ifi)
}

// MulticastLoopback reports whether transmitted multicast packets
// should be copied and send back to the originator.
func (c *vendor.dgramOpt) MulticastLoopback() (bool, error) {
	if !c.ok() {
		return false, vendor.errInvalidConn
	}
	so, ok := vendor.sockOpts[vendor.ssoMulticastLoopback]
	if !ok {
		return false, vendor.errNotImplemented
	}
	on, err := so.GetInt(c.Conn)
	if err != nil {
		return false, err
	}
	return on == 1, nil
}

// SetMulticastLoopback sets whether transmitted multicast packets
// should be copied and send back to the originator.
func (c *vendor.dgramOpt) SetMulticastLoopback(on bool) error {
	if !c.ok() {
		return vendor.errInvalidConn
	}
	so, ok := vendor.sockOpts[vendor.ssoMulticastLoopback]
	if !ok {
		return vendor.errNotImplemented
	}
	return so.SetInt(c.Conn, vendor.boolint(on))
}

// JoinGroup joins the group address group on the interface ifi.
// By default all sources that can cast data to group are accepted.
// It's possible to mute and unmute data transmission from a specific
// source by using ExcludeSourceSpecificGroup and
// IncludeSourceSpecificGroup.
// JoinGroup uses the system assigned multicast interface when ifi is
// nil, although this is not recommended because the assignment
// depends on platforms and sometimes it might require routing
// configuration.
func (c *vendor.dgramOpt) JoinGroup(ifi *net.Interface, group net.Addr) error {
	if !c.ok() {
		return vendor.errInvalidConn
	}
	so, ok := vendor.sockOpts[vendor.ssoJoinGroup]
	if !ok {
		return vendor.errNotImplemented
	}
	grp := vendor.netAddrToIP4(group)
	if grp == nil {
		return vendor.errMissingAddress
	}
	return so.setGroup(c.Conn, ifi, grp)
}

// LeaveGroup leaves the group address group on the interface ifi
// regardless of whether the group is any-source group or
// source-specific group.
func (c *vendor.dgramOpt) LeaveGroup(ifi *net.Interface, group net.Addr) error {
	if !c.ok() {
		return vendor.errInvalidConn
	}
	so, ok := vendor.sockOpts[vendor.ssoLeaveGroup]
	if !ok {
		return vendor.errNotImplemented
	}
	grp := vendor.netAddrToIP4(group)
	if grp == nil {
		return vendor.errMissingAddress
	}
	return so.setGroup(c.Conn, ifi, grp)
}

// JoinSourceSpecificGroup joins the source-specific group comprising
// group and source on the interface ifi.
// JoinSourceSpecificGroup uses the system assigned multicast
// interface when ifi is nil, although this is not recommended because
// the assignment depends on platforms and sometimes it might require
// routing configuration.
func (c *vendor.dgramOpt) JoinSourceSpecificGroup(ifi *net.Interface, group, source net.Addr) error {
	if !c.ok() {
		return vendor.errInvalidConn
	}
	so, ok := vendor.sockOpts[vendor.ssoJoinSourceGroup]
	if !ok {
		return vendor.errNotImplemented
	}
	grp := vendor.netAddrToIP4(group)
	if grp == nil {
		return vendor.errMissingAddress
	}
	src := vendor.netAddrToIP4(source)
	if src == nil {
		return vendor.errMissingAddress
	}
	return so.setSourceGroup(c.Conn, ifi, grp, src)
}

// LeaveSourceSpecificGroup leaves the source-specific group on the
// interface ifi.
func (c *vendor.dgramOpt) LeaveSourceSpecificGroup(ifi *net.Interface, group, source net.Addr) error {
	if !c.ok() {
		return vendor.errInvalidConn
	}
	so, ok := vendor.sockOpts[vendor.ssoLeaveSourceGroup]
	if !ok {
		return vendor.errNotImplemented
	}
	grp := vendor.netAddrToIP4(group)
	if grp == nil {
		return vendor.errMissingAddress
	}
	src := vendor.netAddrToIP4(source)
	if src == nil {
		return vendor.errMissingAddress
	}
	return so.setSourceGroup(c.Conn, ifi, grp, src)
}

// ExcludeSourceSpecificGroup excludes the source-specific group from
// the already joined any-source groups by JoinGroup on the interface
// ifi.
func (c *vendor.dgramOpt) ExcludeSourceSpecificGroup(ifi *net.Interface, group, source net.Addr) error {
	if !c.ok() {
		return vendor.errInvalidConn
	}
	so, ok := vendor.sockOpts[vendor.ssoBlockSourceGroup]
	if !ok {
		return vendor.errNotImplemented
	}
	grp := vendor.netAddrToIP4(group)
	if grp == nil {
		return vendor.errMissingAddress
	}
	src := vendor.netAddrToIP4(source)
	if src == nil {
		return vendor.errMissingAddress
	}
	return so.setSourceGroup(c.Conn, ifi, grp, src)
}

// IncludeSourceSpecificGroup includes the excluded source-specific
// group by ExcludeSourceSpecificGroup again on the interface ifi.
func (c *vendor.dgramOpt) IncludeSourceSpecificGroup(ifi *net.Interface, group, source net.Addr) error {
	if !c.ok() {
		return vendor.errInvalidConn
	}
	so, ok := vendor.sockOpts[vendor.ssoUnblockSourceGroup]
	if !ok {
		return vendor.errNotImplemented
	}
	grp := vendor.netAddrToIP4(group)
	if grp == nil {
		return vendor.errMissingAddress
	}
	src := vendor.netAddrToIP4(source)
	if src == nil {
		return vendor.errMissingAddress
	}
	return so.setSourceGroup(c.Conn, ifi, grp, src)
}

// ICMPFilter returns an ICMP filter.
// Currently only Linux supports this.
func (c *vendor.dgramOpt) ICMPFilter() (*vendor.ICMPFilter, error) {
	if !c.ok() {
		return nil, vendor.errInvalidConn
	}
	so, ok := vendor.sockOpts[vendor.ssoICMPFilter]
	if !ok {
		return nil, vendor.errNotImplemented
	}
	return so.getICMPFilter(c.Conn)
}

// SetICMPFilter deploys the ICMP filter.
// Currently only Linux supports this.
func (c *vendor.dgramOpt) SetICMPFilter(f *vendor.ICMPFilter) error {
	if !c.ok() {
		return vendor.errInvalidConn
	}
	so, ok := vendor.sockOpts[vendor.ssoICMPFilter]
	if !ok {
		return vendor.errNotImplemented
	}
	return so.setICMPFilter(c.Conn, f)
}

// SetBPF attaches a BPF program to the connection.
//
// Only supported on Linux.
func (c *vendor.dgramOpt) SetBPF(filter []vendor.RawInstruction) error {
	if !c.ok() {
		return vendor.errInvalidConn
	}
	so, ok := vendor.sockOpts[vendor.ssoAttachFilter]
	if !ok {
		return vendor.errNotImplemented
	}
	return so.setBPF(c.Conn, filter)
}
