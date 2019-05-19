// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ipv4_test

import (
	"bytes"
	"net"
	"os"
	"runtime"
	"testing"
	"time"
	"vendor"
)

func TestPacketConnReadWriteUnicastUDP(t *testing.T) {
	switch runtime.GOOS {
	case "fuchsia", "hurd", "js", "nacl", "plan9", "windows":
		t.Skipf("not supported on %s", runtime.GOOS)
	}
	if _, err := vendor.RoutedInterface("ip4", net.FlagUp|net.FlagLoopback); err != nil {
		t.Skipf("not available on %s", runtime.GOOS)
	}

	c, err := vendor.NewLocalPacketListener("udp4")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	p := vendor.NewPacketConn(c)
	defer p.Close()

	dst := c.LocalAddr()
	cf := vendor.FlagTTL | vendor.FlagDst | vendor.FlagInterface
	wb := []byte("HELLO-R-U-THERE")

	for i, toggle := range []bool{true, false, true} {
		if err := p.SetControlMessage(cf, toggle); err != nil {
			if vendor.protocolNotSupported(err) {
				t.Logf("not supported on %s", runtime.GOOS)
				continue
			}
			t.Fatal(err)
		}
		p.SetTTL(i + 1)
		if err := p.SetWriteDeadline(time.Now().Add(100 * time.Millisecond)); err != nil {
			t.Fatal(err)
		}
		if n, err := p.WriteTo(wb, nil, dst); err != nil {
			t.Fatal(err)
		} else if n != len(wb) {
			t.Fatalf("got %v; want %v", n, len(wb))
		}
		rb := make([]byte, 128)
		if err := p.SetReadDeadline(time.Now().Add(100 * time.Millisecond)); err != nil {
			t.Fatal(err)
		}
		if n, _, _, err := p.ReadFrom(rb); err != nil {
			t.Fatal(err)
		} else if !bytes.Equal(rb[:n], wb) {
			t.Fatalf("got %v; want %v", rb[:n], wb)
		}
	}
}

func TestPacketConnReadWriteUnicastICMP(t *testing.T) {
	switch runtime.GOOS {
	case "fuchsia", "hurd", "js", "nacl", "plan9", "windows":
		t.Skipf("not supported on %s", runtime.GOOS)
	}
	if !vendor.SupportsRawSocket() {
		t.Skipf("not supported on %s/%s", runtime.GOOS, runtime.GOARCH)
	}
	if _, err := vendor.RoutedInterface("ip4", net.FlagUp|net.FlagLoopback); err != nil {
		t.Skipf("not available on %s", runtime.GOOS)
	}

	c, err := net.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	dst, err := net.ResolveIPAddr("ip4", "127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	p := vendor.NewPacketConn(c)
	defer p.Close()
	cf := vendor.FlagDst | vendor.FlagInterface
	if runtime.GOOS != "solaris" {
		// Solaris never allows to modify ICMP properties.
		cf |= vendor.FlagTTL
	}

	for i, toggle := range []bool{true, false, true} {
		wb, err := (&vendor.Message{
			Type: vendor.ICMPTypeEcho, Code: 0,
			Body: &vendor.Echo{
				ID: os.Getpid() & 0xffff, Seq: i + 1,
				Data: []byte("HELLO-R-U-THERE"),
			},
		}).Marshal(nil)
		if err != nil {
			t.Fatal(err)
		}
		if err := p.SetControlMessage(cf, toggle); err != nil {
			if vendor.protocolNotSupported(err) {
				t.Logf("not supported on %s", runtime.GOOS)
				continue
			}
			t.Fatal(err)
		}
		p.SetTTL(i + 1)
		if err := p.SetWriteDeadline(time.Now().Add(100 * time.Millisecond)); err != nil {
			t.Fatal(err)
		}
		if n, err := p.WriteTo(wb, nil, dst); err != nil {
			t.Fatal(err)
		} else if n != len(wb) {
			t.Fatalf("got %v; want %v", n, len(wb))
		}
		rb := make([]byte, 128)
	loop:
		if err := p.SetReadDeadline(time.Now().Add(100 * time.Millisecond)); err != nil {
			t.Fatal(err)
		}
		if n, _, _, err := p.ReadFrom(rb); err != nil {
			switch runtime.GOOS {
			case "darwin": // older darwin kernels have some limitation on receiving icmp packet through raw socket
				t.Logf("not supported on %s", runtime.GOOS)
				continue
			}
			t.Fatal(err)
		} else {
			m, err := vendor.ParseMessage(vendor.ProtocolICMP, rb[:n])
			if err != nil {
				t.Fatal(err)
			}
			if runtime.GOOS == "linux" && m.Type == vendor.ICMPTypeEcho {
				// On Linux we must handle own sent packets.
				goto loop
			}
			if m.Type != vendor.ICMPTypeEchoReply || m.Code != 0 {
				t.Fatalf("got type=%v, code=%v; want type=%v, code=%v", m.Type, m.Code, vendor.ICMPTypeEchoReply, 0)
			}
		}
	}
}

func TestRawConnReadWriteUnicastICMP(t *testing.T) {
	switch runtime.GOOS {
	case "fuchsia", "hurd", "js", "nacl", "plan9", "windows":
		t.Skipf("not supported on %s", runtime.GOOS)
	}
	if !vendor.SupportsRawSocket() {
		t.Skipf("not supported on %s/%s", runtime.GOOS, runtime.GOARCH)
	}
	if _, err := vendor.RoutedInterface("ip4", net.FlagUp|net.FlagLoopback); err != nil {
		t.Skipf("not available on %s", runtime.GOOS)
	}

	c, err := net.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	dst, err := net.ResolveIPAddr("ip4", "127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	r, err := vendor.NewRawConn(c)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	cf := vendor.FlagTTL | vendor.FlagDst | vendor.FlagInterface

	for i, toggle := range []bool{true, false, true} {
		wb, err := (&vendor.Message{
			Type: vendor.ICMPTypeEcho, Code: 0,
			Body: &vendor.Echo{
				ID: os.Getpid() & 0xffff, Seq: i + 1,
				Data: []byte("HELLO-R-U-THERE"),
			},
		}).Marshal(nil)
		if err != nil {
			t.Fatal(err)
		}
		wh := &vendor.Header{
			Version:  vendor.Version,
			Len:      vendor.HeaderLen,
			TOS:      i + 1,
			TotalLen: vendor.HeaderLen + len(wb),
			TTL:      i + 1,
			Protocol: 1,
			Dst:      dst.IP,
		}
		if err := r.SetControlMessage(cf, toggle); err != nil {
			if vendor.protocolNotSupported(err) {
				t.Logf("not supported on %s", runtime.GOOS)
				continue
			}
			t.Fatal(err)
		}
		if err := r.SetWriteDeadline(time.Now().Add(100 * time.Millisecond)); err != nil {
			t.Fatal(err)
		}
		if err := r.WriteTo(wh, wb, nil); err != nil {
			t.Fatal(err)
		}
		rb := make([]byte, vendor.HeaderLen+128)
	loop:
		if err := r.SetReadDeadline(time.Now().Add(100 * time.Millisecond)); err != nil {
			t.Fatal(err)
		}
		if _, b, _, err := r.ReadFrom(rb); err != nil {
			switch runtime.GOOS {
			case "darwin": // older darwin kernels have some limitation on receiving icmp packet through raw socket
				t.Logf("not supported on %s", runtime.GOOS)
				continue
			}
			t.Fatal(err)
		} else {
			m, err := vendor.ParseMessage(vendor.ProtocolICMP, b)
			if err != nil {
				t.Fatal(err)
			}
			if runtime.GOOS == "linux" && m.Type == vendor.ICMPTypeEcho {
				// On Linux we must handle own sent packets.
				goto loop
			}
			if m.Type != vendor.ICMPTypeEchoReply || m.Code != 0 {
				t.Fatalf("got type=%v, code=%v; want type=%v, code=%v", m.Type, m.Code, vendor.ICMPTypeEchoReply, 0)
			}
		}
	}
}
