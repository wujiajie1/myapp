// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package proxy

import (
	"context"
	"net"
	"vendor"
)

// SOCKS5 returns a Dialer that makes SOCKSv5 connections to the given
// address with an optional username and password.
// See RFC 1928 and RFC 1929.
func SOCKS5(network, address string, auth *vendor.Auth, forward vendor.Dialer) (vendor.Dialer, error) {
	d := vendor.NewDialer(network, address)
	if forward != nil {
		if f, ok := forward.(vendor.ContextDialer); ok {
			d.ProxyDial = func(ctx context.Context, network string, address string) (net.Conn, error) {
				return f.DialContext(ctx, network, address)
			}
		} else {
			d.ProxyDial = func(ctx context.Context, network string, address string) (net.Conn, error) {
				return vendor.dialContext(ctx, forward, network, address)
			}
		}
	}
	if auth != nil {
		up := vendor.UsernamePassword{
			Username: auth.User,
			Password: auth.Password,
		}
		d.AuthMethods = []vendor.AuthMethod{
			vendor.AuthMethodNotRequired,
			vendor.AuthMethodUsernamePassword,
		}
		d.Authenticate = up.Authenticate
	}
	return d, nil
}
