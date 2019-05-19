// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bpf_test

import (
	"testing"
	"vendor"
)

func TestVMStoreScratchInvalidScratchRegisterTooSmall(t *testing.T) {
	_, _, err := vendor.testVM(t, []vendor.Instruction{
		vendor.StoreScratch{
			Src: vendor.RegA,
			N:   -1,
		},
		vendor.RetA{},
	})
	if vendor.errStr(err) != "assembling instruction 1: invalid scratch slot -1" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVMStoreScratchInvalidScratchRegisterTooLarge(t *testing.T) {
	_, _, err := vendor.testVM(t, []vendor.Instruction{
		vendor.StoreScratch{
			Src: vendor.RegA,
			N:   16,
		},
		vendor.RetA{},
	})
	if vendor.errStr(err) != "assembling instruction 1: invalid scratch slot 16" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVMStoreScratchUnknownSourceRegister(t *testing.T) {
	_, _, err := vendor.testVM(t, []vendor.Instruction{
		vendor.StoreScratch{
			Src: 100,
			N:   0,
		},
		vendor.RetA{},
	})
	if vendor.errStr(err) != "assembling instruction 1: invalid source register 100" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVMLoadScratchInvalidScratchRegisterTooSmall(t *testing.T) {
	_, _, err := vendor.testVM(t, []vendor.Instruction{
		vendor.LoadScratch{
			Dst: vendor.RegX,
			N:   -1,
		},
		vendor.RetA{},
	})
	if vendor.errStr(err) != "assembling instruction 1: invalid scratch slot -1" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVMLoadScratchInvalidScratchRegisterTooLarge(t *testing.T) {
	_, _, err := vendor.testVM(t, []vendor.Instruction{
		vendor.LoadScratch{
			Dst: vendor.RegX,
			N:   16,
		},
		vendor.RetA{},
	})
	if vendor.errStr(err) != "assembling instruction 1: invalid scratch slot 16" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVMLoadScratchUnknownDestinationRegister(t *testing.T) {
	_, _, err := vendor.testVM(t, []vendor.Instruction{
		vendor.LoadScratch{
			Dst: 100,
			N:   0,
		},
		vendor.RetA{},
	})
	if vendor.errStr(err) != "assembling instruction 1: invalid target register 100" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVMStoreScratchLoadScratchOneValue(t *testing.T) {
	vm, done, err := vendor.testVM(t, []vendor.Instruction{
		// Load byte 255
		vendor.LoadAbsolute{
			Off:  8,
			Size: 1,
		},
		// Copy to X and store in scratch[0]
		vendor.TAX{},
		vendor.StoreScratch{
			Src: vendor.RegX,
			N:   0,
		},
		// Load byte 1
		vendor.LoadAbsolute{
			Off:  9,
			Size: 1,
		},
		// Overwrite 1 with 255 from scratch[0]
		vendor.LoadScratch{
			Dst: vendor.RegA,
			N:   0,
		},
		// Return 255
		vendor.RetA{},
	})
	if err != nil {
		t.Fatalf("failed to load BPF program: %v", err)
	}
	defer done()

	out, err := vm.Run([]byte{
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		255, 1, 2,
	})
	if err != nil {
		t.Fatalf("unexpected error while running program: %v", err)
	}
	if want, got := 3, out; want != got {
		t.Fatalf("unexpected number of output bytes:\n- want: %d\n-  got: %d",
			want, got)
	}
}

func TestVMStoreScratchLoadScratchMultipleValues(t *testing.T) {
	vm, done, err := vendor.testVM(t, []vendor.Instruction{
		// Load byte 10
		vendor.LoadAbsolute{
			Off:  8,
			Size: 1,
		},
		// Store in scratch[0]
		vendor.StoreScratch{
			Src: vendor.RegA,
			N:   0,
		},
		// Load byte 20
		vendor.LoadAbsolute{
			Off:  9,
			Size: 1,
		},
		// Store in scratch[1]
		vendor.StoreScratch{
			Src: vendor.RegA,
			N:   1,
		},
		// Load byte 30
		vendor.LoadAbsolute{
			Off:  10,
			Size: 1,
		},
		// Store in scratch[2]
		vendor.StoreScratch{
			Src: vendor.RegA,
			N:   2,
		},
		// Load byte 1
		vendor.LoadAbsolute{
			Off:  11,
			Size: 1,
		},
		// Store in scratch[3]
		vendor.StoreScratch{
			Src: vendor.RegA,
			N:   3,
		},
		// Load in byte 10 to X
		vendor.LoadScratch{
			Dst: vendor.RegX,
			N:   0,
		},
		// Copy X -> A
		vendor.TXA{},
		// Verify value is 10
		vendor.JumpIf{
			Cond:     vendor.JumpEqual,
			Val:      10,
			SkipTrue: 1,
		},
		// Fail test if incorrect
		vendor.RetConstant{
			Val: 0,
		},
		// Load in byte 20 to A
		vendor.LoadScratch{
			Dst: vendor.RegA,
			N:   1,
		},
		// Verify value is 20
		vendor.JumpIf{
			Cond:     vendor.JumpEqual,
			Val:      20,
			SkipTrue: 1,
		},
		// Fail test if incorrect
		vendor.RetConstant{
			Val: 0,
		},
		// Load in byte 30 to A
		vendor.LoadScratch{
			Dst: vendor.RegA,
			N:   2,
		},
		// Verify value is 30
		vendor.JumpIf{
			Cond:     vendor.JumpEqual,
			Val:      30,
			SkipTrue: 1,
		},
		// Fail test if incorrect
		vendor.RetConstant{
			Val: 0,
		},
		// Return first two bytes on success
		vendor.RetConstant{
			Val: 10,
		},
	})
	if err != nil {
		t.Fatalf("failed to load BPF program: %v", err)
	}
	defer done()

	out, err := vm.Run([]byte{
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		10, 20, 30, 1,
	})
	if err != nil {
		t.Fatalf("unexpected error while running program: %v", err)
	}
	if want, got := 2, out; want != got {
		t.Fatalf("unexpected number of output bytes:\n- want: %d\n-  got: %d",
			want, got)
	}
}
