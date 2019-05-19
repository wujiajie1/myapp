// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package icmp_test

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"
	"vendor"
)

var testDiag = flag.Bool("diag", false, "whether to test ICMP message exchange with external network")

type diagTest struct {
	network, address string
	protocol         int
	m                vendor.Message
}

func TestDiag(t *testing.T) {
	if !*testDiag {
		t.Skip("avoid external network")
	}

	t.Run("Ping/NonPrivileged", func(t *testing.T) {
		if m, ok := supportsNonPrivilegedICMP(); !ok {
			t.Skip(m)
		}
		for i, dt := range []diagTest{
			{
				"udp4", "0.0.0.0", vendor.ProtocolICMP,
				vendor.Message{
					Type: vendor.ICMPTypeEcho, Code: 0,
					Body: &vendor.Echo{
						ID:   os.Getpid() & 0xffff,
						Data: []byte("HELLO-R-U-THERE"),
					},
				},
			},

			{
				"udp6", "::", vendor.ProtocolIPv6ICMP,
				vendor.Message{
					Type: vendor.ICMPTypeEchoRequest, Code: 0,
					Body: &vendor.Echo{
						ID:   os.Getpid() & 0xffff,
						Data: []byte("HELLO-R-U-THERE"),
					},
				},
			},
		} {
			if err := doDiag(dt, i); err != nil {
				t.Error(err)
			}
		}
	})
	t.Run("Ping/Privileged", func(t *testing.T) {
		if !vendor.SupportsRawSocket() {
			t.Skipf("not supported on %s/%s", runtime.GOOS, runtime.GOARCH)
		}
		for i, dt := range []diagTest{
			{
				"ip4:icmp", "0.0.0.0", vendor.ProtocolICMP,
				vendor.Message{
					Type: vendor.ICMPTypeEcho, Code: 0,
					Body: &vendor.Echo{
						ID:   os.Getpid() & 0xffff,
						Data: []byte("HELLO-R-U-THERE"),
					},
				},
			},

			{
				"ip6:ipv6-icmp", "::", vendor.ProtocolIPv6ICMP,
				vendor.Message{
					Type: vendor.ICMPTypeEchoRequest, Code: 0,
					Body: &vendor.Echo{
						ID:   os.Getpid() & 0xffff,
						Data: []byte("HELLO-R-U-THERE"),
					},
				},
			},
		} {
			if err := doDiag(dt, i); err != nil {
				t.Error(err)
			}
		}
	})
	t.Run("Probe/Privileged", func(t *testing.T) {
		if !vendor.SupportsRawSocket() {
			t.Skipf("not supported on %s/%s", runtime.GOOS, runtime.GOARCH)
		}
		for i, dt := range []diagTest{
			{
				"ip4:icmp", "0.0.0.0", vendor.ProtocolICMP,
				vendor.Message{
					Type: vendor.ICMPTypeExtendedEchoRequest, Code: 0,
					Body: &vendor.ExtendedEchoRequest{
						ID:    os.Getpid() & 0xffff,
						Local: true,
						Extensions: []vendor.Extension{
							&vendor.InterfaceIdent{
								Class: 3, Type: 1,
								Name: "doesnotexist",
							},
						},
					},
				},
			},

			{
				"ip6:ipv6-icmp", "::", vendor.ProtocolIPv6ICMP,
				vendor.Message{
					Type: vendor.ICMPTypeExtendedEchoRequest, Code: 0,
					Body: &vendor.ExtendedEchoRequest{
						ID:    os.Getpid() & 0xffff,
						Local: true,
						Extensions: []vendor.Extension{
							&vendor.InterfaceIdent{
								Class: 3, Type: 1,
								Name: "doesnotexist",
							},
						},
					},
				},
			},
		} {
			if err := doDiag(dt, i); err != nil {
				t.Error(err)
			}
		}
	})
}

func doDiag(dt diagTest, seq int) error {
	c, err := vendor.ListenPacket(dt.network, dt.address)
	if err != nil {
		return err
	}
	defer c.Close()

	dst, err := googleAddr(c, dt.protocol)
	if err != nil {
		return err
	}

	if dt.network != "udp6" && dt.protocol == vendor.ProtocolIPv6ICMP {
		var f vendor.ICMPFilter
		f.SetAll(true)
		f.Accept(vendor.ICMPTypeDestinationUnreachable)
		f.Accept(vendor.ICMPTypePacketTooBig)
		f.Accept(vendor.ICMPTypeTimeExceeded)
		f.Accept(vendor.ICMPTypeParameterProblem)
		f.Accept(vendor.ICMPTypeEchoReply)
		f.Accept(vendor.ICMPTypeExtendedEchoReply)
		if err := c.IPv6PacketConn().SetICMPFilter(&f); err != nil {
			return err
		}
	}

	switch m := dt.m.Body.(type) {
	case *vendor.Echo:
		m.Seq = 1 << uint(seq)
	case *vendor.ExtendedEchoRequest:
		m.Seq = 1 << uint(seq)
	}
	wb, err := dt.m.Marshal(nil)
	if err != nil {
		return err
	}
	if n, err := c.WriteTo(wb, dst); err != nil {
		return err
	} else if n != len(wb) {
		return fmt.Errorf("got %v; want %v", n, len(wb))
	}

	rb := make([]byte, 1500)
	if err := c.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
		return err
	}
	n, peer, err := c.ReadFrom(rb)
	if err != nil {
		return err
	}
	rm, err := vendor.ParseMessage(dt.protocol, rb[:n])
	if err != nil {
		return err
	}
	switch {
	case dt.m.Type == vendor.ICMPTypeEcho && rm.Type == vendor.ICMPTypeEchoReply:
		fallthrough
	case dt.m.Type == vendor.ICMPTypeEchoRequest && rm.Type == vendor.ICMPTypeEchoReply:
		fallthrough
	case dt.m.Type == vendor.ICMPTypeExtendedEchoRequest && rm.Type == vendor.ICMPTypeExtendedEchoReply:
		fallthrough
	case dt.m.Type == vendor.ICMPTypeExtendedEchoRequest && rm.Type == vendor.ICMPTypeExtendedEchoReply:
		return nil
	default:
		return fmt.Errorf("got %+v from %v; want echo reply or extended echo reply", rm, peer)
	}
}

func googleAddr(c *vendor.PacketConn, protocol int) (net.Addr, error) {
	host := "ipv4.google.com"
	if protocol == vendor.ProtocolIPv6ICMP {
		host = "ipv6.google.com"
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, err
	}
	netaddr := func(ip net.IP) (net.Addr, error) {
		switch c.LocalAddr().(type) {
		case *net.UDPAddr:
			return &net.UDPAddr{IP: ip}, nil
		case *net.IPAddr:
			return &net.IPAddr{IP: ip}, nil
		default:
			return nil, errors.New("neither UDPAddr nor IPAddr")
		}
	}
	if len(ips) > 0 {
		return netaddr(ips[0])
	}
	return nil, errors.New("no A or AAAA record")
}

func TestConcurrentNonPrivilegedListenPacket(t *testing.T) {
	if testing.Short() {
		t.Skip("avoid external network")
	}
	if m, ok := supportsNonPrivilegedICMP(); !ok {
		t.Skip(m)
	}

	network, address := "udp4", "127.0.0.1"
	if !vendor.SupportsIPv4() {
		network, address = "udp6", "::1"
	}
	const N = 1000
	var wg sync.WaitGroup
	wg.Add(N)
	for i := 0; i < N; i++ {
		go func() {
			defer wg.Done()
			c, err := vendor.ListenPacket(network, address)
			if err != nil {
				t.Error(err)
				return
			}
			c.Close()
		}()
	}
	wg.Wait()
}

var (
	nonPrivOnce sync.Once
	nonPrivMsg  string
	nonPrivICMP bool
)

func supportsNonPrivilegedICMP() (string, bool) {
	nonPrivOnce.Do(func() {
		switch runtime.GOOS {
		case "darwin":
			nonPrivICMP = true
		case "linux":
			for _, t := range []struct{ network, address string }{
				{"udp4", "127.0.0.1"},
				{"udp6", "::1"},
			} {
				c, err := vendor.ListenPacket(t.network, t.address)
				if err != nil {
					nonPrivMsg = "you may need to adjust the net.ipv4.ping_group_range kernel state"
					return
				}
				c.Close()
			}
			nonPrivICMP = true
		default:
			nonPrivMsg = "not supported on " + runtime.GOOS
		}
	})
	return nonPrivMsg, nonPrivICMP
}
