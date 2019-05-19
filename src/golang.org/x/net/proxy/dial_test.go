// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package proxy

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"
	"time"
	"vendor"
)

func TestDial(t *testing.T) {
	vendor.ResetProxyEnv()
	t.Run("DirectWithCancel", func(t *testing.T) {
		defer vendor.ResetProxyEnv()
		l, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		defer l.Close()
		_, port, err := net.SplitHostPort(l.Addr().String())
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		c, err := vendor.Dial(ctx, l.Addr().Network(), net.JoinHostPort("", port))
		if err != nil {
			t.Fatal(err)
		}
		c.Close()
	})
	t.Run("DirectWithTimeout", func(t *testing.T) {
		defer vendor.ResetProxyEnv()
		l, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		defer l.Close()
		_, port, err := net.SplitHostPort(l.Addr().String())
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		c, err := vendor.Dial(ctx, l.Addr().Network(), net.JoinHostPort("", port))
		if err != nil {
			t.Fatal(err)
		}
		c.Close()
	})
	t.Run("DirectWithTimeoutExceeded", func(t *testing.T) {
		defer vendor.ResetProxyEnv()
		l, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		defer l.Close()
		_, port, err := net.SplitHostPort(l.Addr().String())
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		time.Sleep(time.Millisecond)
		defer cancel()
		c, err := vendor.Dial(ctx, l.Addr().Network(), net.JoinHostPort("", port))
		if err == nil {
			defer c.Close()
			t.Fatal("failed to timeout")
		}
	})
	t.Run("SOCKS5", func(t *testing.T) {
		defer vendor.ResetProxyEnv()
		s, err := vendor.NewServer(vendor.NoAuthRequired, vendor.NoProxyRequired)
		if err != nil {
			t.Fatal(err)
		}
		defer s.Close()
		if err = os.Setenv("ALL_PROXY", fmt.Sprintf("socks5://%s", s.Addr().String())); err != nil {
			t.Fatal(err)
		}
		c, err := vendor.Dial(context.Background(), s.TargetAddr().Network(), s.TargetAddr().String())
		if err != nil {
			t.Fatal(err)
		}
		c.Close()
	})
	t.Run("SOCKS5WithTimeout", func(t *testing.T) {
		defer vendor.ResetProxyEnv()
		s, err := vendor.NewServer(vendor.NoAuthRequired, vendor.NoProxyRequired)
		if err != nil {
			t.Fatal(err)
		}
		defer s.Close()
		if err = os.Setenv("ALL_PROXY", fmt.Sprintf("socks5://%s", s.Addr().String())); err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		c, err := vendor.Dial(ctx, s.TargetAddr().Network(), s.TargetAddr().String())
		if err != nil {
			t.Fatal(err)
		}
		c.Close()
	})
	t.Run("SOCKS5WithTimeoutExceeded", func(t *testing.T) {
		defer vendor.ResetProxyEnv()
		s, err := vendor.NewServer(vendor.NoAuthRequired, vendor.NoProxyRequired)
		if err != nil {
			t.Fatal(err)
		}
		defer s.Close()
		if err = os.Setenv("ALL_PROXY", fmt.Sprintf("socks5://%s", s.Addr().String())); err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		time.Sleep(time.Millisecond)
		defer cancel()
		c, err := vendor.Dial(ctx, s.TargetAddr().Network(), s.TargetAddr().String())
		if err == nil {
			defer c.Close()
			t.Fatal("failed to timeout")
		}
	})
}
