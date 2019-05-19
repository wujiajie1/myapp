// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package icmp_test

import (
	"errors"
	"fmt"
	"net"
	"reflect"
	"testing"
	"vendor"
)

func TestMarshalAndParseMultipartMessage(t *testing.T) {
	fn := func(t *testing.T, proto int, tm vendor.Message) error {
		b, err := tm.Marshal(nil)
		if err != nil {
			return err
		}
		switch tm.Type {
		case vendor.ICMPTypeExtendedEchoRequest, vendor.ICMPTypeExtendedEchoRequest:
		default:
			switch proto {
			case vendor.ProtocolICMP:
				if b[5] != 32 {
					return fmt.Errorf("got %d; want 32", b[5])
				}
			case vendor.ProtocolIPv6ICMP:
				if b[4] != 16 {
					return fmt.Errorf("got %d; want 16", b[4])
				}
			default:
				return fmt.Errorf("unknown protocol: %d", proto)
			}
		}
		m, err := vendor.ParseMessage(proto, b)
		if err != nil {
			return err
		}
		if m.Type != tm.Type || m.Code != tm.Code {
			return fmt.Errorf("got %v; want %v", m, &tm)
		}
		switch m.Type {
		case vendor.ICMPTypeExtendedEchoRequest, vendor.ICMPTypeExtendedEchoRequest:
			got, want := m.Body.(*vendor.ExtendedEchoRequest), tm.Body.(*vendor.ExtendedEchoRequest)
			if !reflect.DeepEqual(got.Extensions, want.Extensions) {
				return errors.New(dumpExtensions(got.Extensions, want.Extensions))
			}
		case vendor.ICMPTypeDestinationUnreachable:
			got, want := m.Body.(*vendor.DstUnreach), tm.Body.(*vendor.DstUnreach)
			if !reflect.DeepEqual(got.Extensions, want.Extensions) {
				return errors.New(dumpExtensions(got.Extensions, want.Extensions))
			}
			if len(got.Data) != 128 {
				return fmt.Errorf("got %d; want 128", len(got.Data))
			}
		case vendor.ICMPTypeTimeExceeded:
			got, want := m.Body.(*vendor.TimeExceeded), tm.Body.(*vendor.TimeExceeded)
			if !reflect.DeepEqual(got.Extensions, want.Extensions) {
				return errors.New(dumpExtensions(got.Extensions, want.Extensions))
			}
			if len(got.Data) != 128 {
				return fmt.Errorf("got %d; want 128", len(got.Data))
			}
		case vendor.ICMPTypeParameterProblem:
			got, want := m.Body.(*vendor.ParamProb), tm.Body.(*vendor.ParamProb)
			if !reflect.DeepEqual(got.Extensions, want.Extensions) {
				return errors.New(dumpExtensions(got.Extensions, want.Extensions))
			}
			if len(got.Data) != 128 {
				return fmt.Errorf("got %d; want 128", len(got.Data))
			}
		case vendor.ICMPTypeDestinationUnreachable:
			got, want := m.Body.(*vendor.DstUnreach), tm.Body.(*vendor.DstUnreach)
			if !reflect.DeepEqual(got.Extensions, want.Extensions) {
				return errors.New(dumpExtensions(got.Extensions, want.Extensions))
			}
			if len(got.Data) != 128 {
				return fmt.Errorf("got %d; want 128", len(got.Data))
			}
		case vendor.ICMPTypeTimeExceeded:
			got, want := m.Body.(*vendor.TimeExceeded), tm.Body.(*vendor.TimeExceeded)
			if !reflect.DeepEqual(got.Extensions, want.Extensions) {
				return errors.New(dumpExtensions(got.Extensions, want.Extensions))
			}
			if len(got.Data) != 128 {
				return fmt.Errorf("got %d; want 128", len(got.Data))
			}
		default:
			return fmt.Errorf("unknown message type: %v", m.Type)
		}
		return nil
	}

	t.Run("IPv4", func(t *testing.T) {
		for i, tm := range []vendor.Message{
			{
				Type: vendor.ICMPTypeDestinationUnreachable, Code: 15,
				Body: &vendor.DstUnreach{
					Data: []byte("ERROR-INVOKING-PACKET"),
					Extensions: []vendor.Extension{
						&vendor.MPLSLabelStack{
							Class: 1,
							Type:  1,
							Labels: []vendor.MPLSLabel{
								{
									Label: 16014,
									TC:    0x4,
									S:     true,
									TTL:   255,
								},
							},
						},
						&vendor.InterfaceInfo{
							Class: 2,
							Type:  0x0f,
							Interface: &net.Interface{
								Index: 15,
								Name:  "en101",
								MTU:   8192,
							},
							Addr: &net.IPAddr{
								IP: net.IPv4(192, 168, 0, 1).To4(),
							},
						},
					},
				},
			},
			{
				Type: vendor.ICMPTypeTimeExceeded, Code: 1,
				Body: &vendor.TimeExceeded{
					Data: []byte("ERROR-INVOKING-PACKET"),
					Extensions: []vendor.Extension{
						&vendor.InterfaceInfo{
							Class: 2,
							Type:  0x0f,
							Interface: &net.Interface{
								Index: 15,
								Name:  "en101",
								MTU:   8192,
							},
							Addr: &net.IPAddr{
								IP: net.IPv4(192, 168, 0, 1).To4(),
							},
						},
						&vendor.MPLSLabelStack{
							Class: 1,
							Type:  1,
							Labels: []vendor.MPLSLabel{
								{
									Label: 16014,
									TC:    0x4,
									S:     true,
									TTL:   255,
								},
							},
						},
					},
				},
			},
			{
				Type: vendor.ICMPTypeParameterProblem, Code: 2,
				Body: &vendor.ParamProb{
					Pointer: 8,
					Data:    []byte("ERROR-INVOKING-PACKET"),
					Extensions: []vendor.Extension{
						&vendor.MPLSLabelStack{
							Class: 1,
							Type:  1,
							Labels: []vendor.MPLSLabel{
								{
									Label: 16014,
									TC:    0x4,
									S:     true,
									TTL:   255,
								},
							},
						},
						&vendor.InterfaceInfo{
							Class: 2,
							Type:  0x0f,
							Interface: &net.Interface{
								Index: 15,
								Name:  "en101",
								MTU:   8192,
							},
							Addr: &net.IPAddr{
								IP: net.IPv4(192, 168, 0, 1).To4(),
							},
						},
						&vendor.InterfaceInfo{
							Class: 2,
							Type:  0x2f,
							Interface: &net.Interface{
								Index: 16,
								Name:  "en102",
								MTU:   8192,
							},
							Addr: &net.IPAddr{
								IP: net.IPv4(192, 168, 0, 2).To4(),
							},
						},
					},
				},
			},
			{
				Type: vendor.ICMPTypeExtendedEchoRequest, Code: 0,
				Body: &vendor.ExtendedEchoRequest{
					ID: 1, Seq: 2, Local: true,
					Extensions: []vendor.Extension{
						&vendor.InterfaceIdent{
							Class: 3,
							Type:  1,
							Name:  "en101",
						},
					},
				},
			},
			{
				Type: vendor.ICMPTypeExtendedEchoRequest, Code: 0,
				Body: &vendor.ExtendedEchoRequest{
					ID: 1, Seq: 2, Local: true,
					Extensions: []vendor.Extension{
						&vendor.InterfaceIdent{
							Class: 3,
							Type:  2,
							Index: 911,
						},
					},
				},
			},
			{
				Type: vendor.ICMPTypeExtendedEchoRequest, Code: 0,
				Body: &vendor.ExtendedEchoRequest{
					ID: 1, Seq: 2,
					Extensions: []vendor.Extension{
						&vendor.InterfaceIdent{
							Class: 3,
							Type:  3,
							AFI:   vendor.AddrFamily48bitMAC,
							Addr:  []byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab},
						},
					},
				},
			},
		} {
			if err := fn(t, vendor.ProtocolICMP, tm); err != nil {
				t.Errorf("#%d: %v", i, err)
			}
		}
	})
	t.Run("IPv6", func(t *testing.T) {
		for i, tm := range []vendor.Message{
			{
				Type: vendor.ICMPTypeDestinationUnreachable, Code: 6,
				Body: &vendor.DstUnreach{
					Data: []byte("ERROR-INVOKING-PACKET"),
					Extensions: []vendor.Extension{
						&vendor.MPLSLabelStack{
							Class: 1,
							Type:  1,
							Labels: []vendor.MPLSLabel{
								{
									Label: 16014,
									TC:    0x4,
									S:     true,
									TTL:   255,
								},
							},
						},
						&vendor.InterfaceInfo{
							Class: 2,
							Type:  0x0f,
							Interface: &net.Interface{
								Index: 15,
								Name:  "en101",
								MTU:   8192,
							},
							Addr: &net.IPAddr{
								IP:   net.ParseIP("fe80::1"),
								Zone: "en101",
							},
						},
					},
				},
			},
			{
				Type: vendor.ICMPTypeTimeExceeded, Code: 1,
				Body: &vendor.TimeExceeded{
					Data: []byte("ERROR-INVOKING-PACKET"),
					Extensions: []vendor.Extension{
						&vendor.InterfaceInfo{
							Class: 2,
							Type:  0x0f,
							Interface: &net.Interface{
								Index: 15,
								Name:  "en101",
								MTU:   8192,
							},
							Addr: &net.IPAddr{
								IP:   net.ParseIP("fe80::1"),
								Zone: "en101",
							},
						},
						&vendor.MPLSLabelStack{
							Class: 1,
							Type:  1,
							Labels: []vendor.MPLSLabel{
								{
									Label: 16014,
									TC:    0x4,
									S:     true,
									TTL:   255,
								},
							},
						},
						&vendor.InterfaceInfo{
							Class: 2,
							Type:  0x2f,
							Interface: &net.Interface{
								Index: 16,
								Name:  "en102",
								MTU:   8192,
							},
							Addr: &net.IPAddr{
								IP:   net.ParseIP("fe80::1"),
								Zone: "en102",
							},
						},
					},
				},
			},
			{
				Type: vendor.ICMPTypeExtendedEchoRequest, Code: 0,
				Body: &vendor.ExtendedEchoRequest{
					ID: 1, Seq: 2, Local: true,
					Extensions: []vendor.Extension{
						&vendor.InterfaceIdent{
							Class: 3,
							Type:  1,
							Name:  "en101",
						},
					},
				},
			},
			{
				Type: vendor.ICMPTypeExtendedEchoRequest, Code: 0,
				Body: &vendor.ExtendedEchoRequest{
					ID: 1, Seq: 2, Local: true,
					Extensions: []vendor.Extension{
						&vendor.InterfaceIdent{
							Class: 3,
							Type:  2,
							Index: 911,
						},
					},
				},
			},
			{
				Type: vendor.ICMPTypeExtendedEchoRequest, Code: 0,
				Body: &vendor.ExtendedEchoRequest{
					ID: 1, Seq: 2,
					Extensions: []vendor.Extension{
						&vendor.InterfaceIdent{
							Class: 3,
							Type:  3,
							AFI:   vendor.AddrFamilyIPv4,
							Addr:  []byte{192, 0, 2, 1},
						},
					},
				},
			},
		} {
			if err := fn(t, vendor.ProtocolIPv6ICMP, tm); err != nil {
				t.Errorf("#%d: %v", i, err)
			}
		}
	})
}

func dumpExtensions(gotExts, wantExts []vendor.Extension) string {
	var s string
	for i, got := range gotExts {
		switch got := got.(type) {
		case *vendor.MPLSLabelStack:
			want := wantExts[i].(*vendor.MPLSLabelStack)
			if !reflect.DeepEqual(got, want) {
				s += fmt.Sprintf("#%d: got %#v; want %#v\n", i, got, want)
			}
		case *vendor.InterfaceInfo:
			want := wantExts[i].(*vendor.InterfaceInfo)
			if !reflect.DeepEqual(got, want) {
				s += fmt.Sprintf("#%d: got %#v, %#v, %#v; want %#v, %#v, %#v\n", i, got, got.Interface, got.Addr, want, want.Interface, want.Addr)
			}
		case *vendor.InterfaceIdent:
			want := wantExts[i].(*vendor.InterfaceIdent)
			if !reflect.DeepEqual(got, want) {
				s += fmt.Sprintf("#%d: got %#v; want %#v\n", i, got, want)
			}
		case *vendor.RawExtension:
			s += fmt.Sprintf("#%d: raw extension\n", i)
		}
	}
	if len(s) == 0 {
		s += "empty extension"
	}
	return s[:len(s)-1]
}

func TestMultipartMessageBodyLen(t *testing.T) {
	for i, tt := range []struct {
		proto int
		in    vendor.MessageBody
		out   int
	}{
		{
			vendor.ProtocolICMP,
			&vendor.DstUnreach{
				Data: make([]byte, vendor.HeaderLen),
			},
			4 + vendor.HeaderLen, // unused and original datagram
		},
		{
			vendor.ProtocolICMP,
			&vendor.TimeExceeded{
				Data: make([]byte, vendor.HeaderLen),
			},
			4 + vendor.HeaderLen, // unused and original datagram
		},
		{
			vendor.ProtocolICMP,
			&vendor.ParamProb{
				Data: make([]byte, vendor.HeaderLen),
			},
			4 + vendor.HeaderLen, // [pointer, unused] and original datagram
		},

		{
			vendor.ProtocolICMP,
			&vendor.ParamProb{
				Data: make([]byte, vendor.HeaderLen),
				Extensions: []vendor.Extension{
					&vendor.MPLSLabelStack{},
				},
			},
			4 + 4 + 4 + 0 + 128, // [pointer, length, unused], extension header, object header, object payload, original datagram
		},
		{
			vendor.ProtocolICMP,
			&vendor.ParamProb{
				Data: make([]byte, 128),
				Extensions: []vendor.Extension{
					&vendor.MPLSLabelStack{},
				},
			},
			4 + 4 + 4 + 0 + 128, // [pointer, length, unused], extension header, object header, object payload and original datagram
		},
		{
			vendor.ProtocolICMP,
			&vendor.ParamProb{
				Data: make([]byte, 129),
				Extensions: []vendor.Extension{
					&vendor.MPLSLabelStack{},
				},
			},
			4 + 4 + 4 + 0 + 132, // [pointer, length, unused], extension header, object header, object payload and original datagram
		},

		{
			vendor.ProtocolIPv6ICMP,
			&vendor.DstUnreach{
				Data: make([]byte, vendor.HeaderLen),
			},
			4 + vendor.HeaderLen, // unused and original datagram
		},
		{
			vendor.ProtocolIPv6ICMP,
			&vendor.PacketTooBig{
				Data: make([]byte, vendor.HeaderLen),
			},
			4 + vendor.HeaderLen, // mtu and original datagram
		},
		{
			vendor.ProtocolIPv6ICMP,
			&vendor.TimeExceeded{
				Data: make([]byte, vendor.HeaderLen),
			},
			4 + vendor.HeaderLen, // unused and original datagram
		},
		{
			vendor.ProtocolIPv6ICMP,
			&vendor.ParamProb{
				Data: make([]byte, vendor.HeaderLen),
			},
			4 + vendor.HeaderLen, // pointer and original datagram
		},

		{
			vendor.ProtocolIPv6ICMP,
			&vendor.DstUnreach{
				Data: make([]byte, 127),
				Extensions: []vendor.Extension{
					&vendor.MPLSLabelStack{},
				},
			},
			4 + 4 + 4 + 0 + 128, // [length, unused], extension header, object header, object payload and original datagram
		},
		{
			vendor.ProtocolIPv6ICMP,
			&vendor.DstUnreach{
				Data: make([]byte, 128),
				Extensions: []vendor.Extension{
					&vendor.MPLSLabelStack{},
				},
			},
			4 + 4 + 4 + 0 + 128, // [length, unused], extension header, object header, object payload and original datagram
		},
		{
			vendor.ProtocolIPv6ICMP,
			&vendor.DstUnreach{
				Data: make([]byte, 129),
				Extensions: []vendor.Extension{
					&vendor.MPLSLabelStack{},
				},
			},
			4 + 4 + 4 + 0 + 136, // [length, unused], extension header, object header, object payload and original datagram
		},

		{
			vendor.ProtocolICMP,
			&vendor.ExtendedEchoRequest{},
			4, // [id, seq, l-bit]
		},
		{
			vendor.ProtocolICMP,
			&vendor.ExtendedEchoRequest{
				Extensions: []vendor.Extension{
					&vendor.InterfaceIdent{},
				},
			},
			4 + 4 + 4, // [id, seq, l-bit], extension header, object header
		},
		{
			vendor.ProtocolIPv6ICMP,
			&vendor.ExtendedEchoRequest{
				Extensions: []vendor.Extension{
					&vendor.InterfaceIdent{
						Type: 3,
						AFI:  vendor.AddrFamilyNSAP,
						Addr: []byte{0x49, 0x00, 0x01, 0xaa, 0xaa, 0xbb, 0xbb, 0xcc, 0xcc, 0x00},
					},
				},
			},
			4 + 4 + 4 + 16, // [id, seq, l-bit], extension header, object header, object payload
		},
	} {
		if out := tt.in.Len(tt.proto); out != tt.out {
			t.Errorf("#%d: got %d; want %d", i, out, tt.out)
		}
	}
}
