// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bpf_test

import (
	"net"
	"testing"
	"vendor"
)

func TestVMLoadAbsoluteOffsetOutOfBounds(t *testing.T) {
	vm, done, err := vendor.testVM(t, []vendor.Instruction{
		vendor.LoadAbsolute{
			Off:  100,
			Size: 2,
		},
		vendor.RetA{},
	})
	if err != nil {
		t.Fatalf("failed to load BPF program: %v", err)
	}
	defer done()

	out, err := vm.Run([]byte{
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		0, 1, 2, 3,
	})
	if err != nil {
		t.Fatalf("unexpected error while running program: %v", err)
	}
	if want, got := 0, out; want != got {
		t.Fatalf("unexpected number of output bytes:\n- want: %d\n-  got: %d",
			want, got)
	}
}

func TestVMLoadAbsoluteOffsetPlusSizeOutOfBounds(t *testing.T) {
	vm, done, err := vendor.testVM(t, []vendor.Instruction{
		vendor.LoadAbsolute{
			Off:  8,
			Size: 2,
		},
		vendor.RetA{},
	})
	if err != nil {
		t.Fatalf("failed to load BPF program: %v", err)
	}
	defer done()

	out, err := vm.Run([]byte{
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		0,
	})
	if err != nil {
		t.Fatalf("unexpected error while running program: %v", err)
	}
	if want, got := 0, out; want != got {
		t.Fatalf("unexpected number of output bytes:\n- want: %d\n-  got: %d",
			want, got)
	}
}

func TestVMLoadAbsoluteBadInstructionSize(t *testing.T) {
	_, _, err := vendor.testVM(t, []vendor.Instruction{
		vendor.LoadAbsolute{
			Size: 5,
		},
		vendor.RetA{},
	})
	if vendor.errStr(err) != "assembling instruction 1: invalid load byte length 0" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVMLoadConstantOK(t *testing.T) {
	vm, done, err := vendor.testVM(t, []vendor.Instruction{
		vendor.LoadConstant{
			Dst: vendor.RegX,
			Val: 9,
		},
		vendor.TXA{},
		vendor.RetA{},
	})
	if err != nil {
		t.Fatalf("failed to load BPF program: %v", err)
	}
	defer done()

	out, err := vm.Run([]byte{
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		0,
	})
	if err != nil {
		t.Fatalf("unexpected error while running program: %v", err)
	}
	if want, got := 1, out; want != got {
		t.Fatalf("unexpected number of output bytes:\n- want: %d\n-  got: %d",
			want, got)
	}
}

func TestVMLoadIndirectOutOfBounds(t *testing.T) {
	vm, done, err := vendor.testVM(t, []vendor.Instruction{
		vendor.LoadIndirect{
			Off:  100,
			Size: 1,
		},
		vendor.RetA{},
	})
	if err != nil {
		t.Fatalf("failed to load BPF program: %v", err)
	}
	defer done()

	out, err := vm.Run([]byte{
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		0,
	})
	if err != nil {
		t.Fatalf("unexpected error while running program: %v", err)
	}
	if want, got := 0, out; want != got {
		t.Fatalf("unexpected number of output bytes:\n- want: %d\n-  got: %d",
			want, got)
	}
}

func TestVMLoadMemShiftOutOfBounds(t *testing.T) {
	vm, done, err := vendor.testVM(t, []vendor.Instruction{
		vendor.LoadMemShift{
			Off: 100,
		},
		vendor.RetA{},
	})
	if err != nil {
		t.Fatalf("failed to load BPF program: %v", err)
	}
	defer done()

	out, err := vm.Run([]byte{
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		0,
	})
	if err != nil {
		t.Fatalf("unexpected error while running program: %v", err)
	}
	if want, got := 0, out; want != got {
		t.Fatalf("unexpected number of output bytes:\n- want: %d\n-  got: %d",
			want, got)
	}
}

const (
	dhcp4Port = 53
)

func TestVMLoadMemShiftLoadIndirectNoResult(t *testing.T) {
	vm, in, done := testDHCPv4(t)
	defer done()

	// Append mostly empty UDP header with incorrect DHCPv4 port
	in = append(in, []byte{
		0, 0,
		0, dhcp4Port + 1,
		0, 0,
		0, 0,
	}...)

	out, err := vm.Run(in)
	if err != nil {
		t.Fatalf("unexpected error while running program: %v", err)
	}
	if want, got := 0, out; want != got {
		t.Fatalf("unexpected number of output bytes:\n- want: %d\n-  got: %d",
			want, got)
	}
}

func TestVMLoadMemShiftLoadIndirectOK(t *testing.T) {
	vm, in, done := testDHCPv4(t)
	defer done()

	// Append mostly empty UDP header with correct DHCPv4 port
	in = append(in, []byte{
		0, 0,
		0, dhcp4Port,
		0, 0,
		0, 0,
	}...)

	out, err := vm.Run(in)
	if err != nil {
		t.Fatalf("unexpected error while running program: %v", err)
	}
	if want, got := len(in)-8, out; want != got {
		t.Fatalf("unexpected number of output bytes:\n- want: %d\n-  got: %d",
			want, got)
	}
}

func testDHCPv4(t *testing.T) (vendor.virtualMachine, []byte, func()) {
	// DHCPv4 test data courtesy of David Anderson:
	// https://github.com/google/netboot/blob/master/dhcp4/conn_linux.go#L59-L70
	vm, done, err := vendor.testVM(t, []vendor.Instruction{
		// Load IPv4 packet length
		vendor.LoadMemShift{Off: 8},
		// Get UDP dport
		vendor.LoadIndirect{Off: 8 + 2, Size: 2},
		// Correct dport?
		vendor.JumpIf{Cond: vendor.JumpEqual, Val: dhcp4Port, SkipFalse: 1},
		// Accept
		vendor.RetConstant{Val: 1500},
		// Ignore
		vendor.RetConstant{Val: 0},
	})
	if err != nil {
		t.Fatalf("failed to load BPF program: %v", err)
	}

	// Minimal requirements to make a valid IPv4 header
	h := &vendor.Header{
		Len: vendor.HeaderLen,
		Src: net.IPv4(192, 168, 1, 1),
		Dst: net.IPv4(192, 168, 1, 2),
	}
	hb, err := h.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal IPv4 header: %v", err)
	}

	hb = append([]byte{
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
	}, hb...)

	return vm, hb, done
}
