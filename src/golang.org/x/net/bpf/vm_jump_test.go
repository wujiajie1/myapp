// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bpf_test

import (
	"testing"
	"vendor"
)

func TestVMJumpOne(t *testing.T) {
	vm, done, err := vendor.testVM(t, []vendor.Instruction{
		vendor.LoadAbsolute{
			Off:  8,
			Size: 1,
		},
		vendor.Jump{
			Skip: 1,
		},
		vendor.RetConstant{
			Val: 0,
		},
		vendor.RetConstant{
			Val: 9,
		},
	})
	if err != nil {
		t.Fatalf("failed to load BPF program: %v", err)
	}
	defer done()

	out, err := vm.Run([]byte{
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		1,
	})
	if err != nil {
		t.Fatalf("unexpected error while running program: %v", err)
	}
	if want, got := 1, out; want != got {
		t.Fatalf("unexpected number of output bytes:\n- want: %d\n-  got: %d",
			want, got)
	}
}

func TestVMJumpOutOfProgram(t *testing.T) {
	_, _, err := vendor.testVM(t, []vendor.Instruction{
		vendor.Jump{
			Skip: 1,
		},
		vendor.RetA{},
	})
	if vendor.errStr(err) != "cannot jump 1 instructions; jumping past program bounds" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVMJumpIfTrueOutOfProgram(t *testing.T) {
	_, _, err := vendor.testVM(t, []vendor.Instruction{
		vendor.JumpIf{
			Cond:     vendor.JumpEqual,
			SkipTrue: 2,
		},
		vendor.RetA{},
	})
	if vendor.errStr(err) != "cannot jump 2 instructions in true case; jumping past program bounds" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVMJumpIfFalseOutOfProgram(t *testing.T) {
	_, _, err := vendor.testVM(t, []vendor.Instruction{
		vendor.JumpIf{
			Cond:      vendor.JumpEqual,
			SkipFalse: 3,
		},
		vendor.RetA{},
	})
	if vendor.errStr(err) != "cannot jump 3 instructions in false case; jumping past program bounds" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVMJumpIfXTrueOutOfProgram(t *testing.T) {
	_, _, err := vendor.testVM(t, []vendor.Instruction{
		vendor.JumpIfX{
			Cond:     vendor.JumpEqual,
			SkipTrue: 2,
		},
		vendor.RetA{},
	})
	if vendor.errStr(err) != "cannot jump 2 instructions in true case; jumping past program bounds" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVMJumpIfXFalseOutOfProgram(t *testing.T) {
	_, _, err := vendor.testVM(t, []vendor.Instruction{
		vendor.JumpIfX{
			Cond:      vendor.JumpEqual,
			SkipFalse: 3,
		},
		vendor.RetA{},
	})
	if vendor.errStr(err) != "cannot jump 3 instructions in false case; jumping past program bounds" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVMJumpIfEqual(t *testing.T) {
	vm, done, err := vendor.testVM(t, []vendor.Instruction{
		vendor.LoadAbsolute{
			Off:  8,
			Size: 1,
		},
		vendor.JumpIf{
			Cond:     vendor.JumpEqual,
			Val:      1,
			SkipTrue: 1,
		},
		vendor.RetConstant{
			Val: 0,
		},
		vendor.RetConstant{
			Val: 9,
		},
	})
	if err != nil {
		t.Fatalf("failed to load BPF program: %v", err)
	}
	defer done()

	out, err := vm.Run([]byte{
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		1,
	})
	if err != nil {
		t.Fatalf("unexpected error while running program: %v", err)
	}
	if want, got := 1, out; want != got {
		t.Fatalf("unexpected number of output bytes:\n- want: %d\n-  got: %d",
			want, got)
	}
}

func TestVMJumpIfNotEqual(t *testing.T) {
	vm, done, err := vendor.testVM(t, []vendor.Instruction{
		vendor.LoadAbsolute{
			Off:  8,
			Size: 1,
		},
		vendor.JumpIf{
			Cond:      vendor.JumpNotEqual,
			Val:       1,
			SkipFalse: 1,
		},
		vendor.RetConstant{
			Val: 0,
		},
		vendor.RetConstant{
			Val: 9,
		},
	})
	if err != nil {
		t.Fatalf("failed to load BPF program: %v", err)
	}
	defer done()

	out, err := vm.Run([]byte{
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		1,
	})
	if err != nil {
		t.Fatalf("unexpected error while running program: %v", err)
	}
	if want, got := 1, out; want != got {
		t.Fatalf("unexpected number of output bytes:\n- want: %d\n-  got: %d",
			want, got)
	}
}

func TestVMJumpIfGreaterThan(t *testing.T) {
	vm, done, err := vendor.testVM(t, []vendor.Instruction{
		vendor.LoadAbsolute{
			Off:  8,
			Size: 4,
		},
		vendor.JumpIf{
			Cond:     vendor.JumpGreaterThan,
			Val:      0x00010202,
			SkipTrue: 1,
		},
		vendor.RetConstant{
			Val: 0,
		},
		vendor.RetConstant{
			Val: 12,
		},
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
	if want, got := 4, out; want != got {
		t.Fatalf("unexpected number of output bytes:\n- want: %d\n-  got: %d",
			want, got)
	}
}

func TestVMJumpIfLessThan(t *testing.T) {
	vm, done, err := vendor.testVM(t, []vendor.Instruction{
		vendor.LoadAbsolute{
			Off:  8,
			Size: 4,
		},
		vendor.JumpIf{
			Cond:     vendor.JumpLessThan,
			Val:      0xff010203,
			SkipTrue: 1,
		},
		vendor.RetConstant{
			Val: 0,
		},
		vendor.RetConstant{
			Val: 12,
		},
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
	if want, got := 4, out; want != got {
		t.Fatalf("unexpected number of output bytes:\n- want: %d\n-  got: %d",
			want, got)
	}
}

func TestVMJumpIfGreaterOrEqual(t *testing.T) {
	vm, done, err := vendor.testVM(t, []vendor.Instruction{
		vendor.LoadAbsolute{
			Off:  8,
			Size: 4,
		},
		vendor.JumpIf{
			Cond:     vendor.JumpGreaterOrEqual,
			Val:      0x00010203,
			SkipTrue: 1,
		},
		vendor.RetConstant{
			Val: 0,
		},
		vendor.RetConstant{
			Val: 12,
		},
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
	if want, got := 4, out; want != got {
		t.Fatalf("unexpected number of output bytes:\n- want: %d\n-  got: %d",
			want, got)
	}
}

func TestVMJumpIfLessOrEqual(t *testing.T) {
	vm, done, err := vendor.testVM(t, []vendor.Instruction{
		vendor.LoadAbsolute{
			Off:  8,
			Size: 4,
		},
		vendor.JumpIf{
			Cond:     vendor.JumpLessOrEqual,
			Val:      0xff010203,
			SkipTrue: 1,
		},
		vendor.RetConstant{
			Val: 0,
		},
		vendor.RetConstant{
			Val: 12,
		},
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
	if want, got := 4, out; want != got {
		t.Fatalf("unexpected number of output bytes:\n- want: %d\n-  got: %d",
			want, got)
	}
}

func TestVMJumpIfBitsSet(t *testing.T) {
	vm, done, err := vendor.testVM(t, []vendor.Instruction{
		vendor.LoadAbsolute{
			Off:  8,
			Size: 2,
		},
		vendor.JumpIf{
			Cond:     vendor.JumpBitsSet,
			Val:      0x1122,
			SkipTrue: 1,
		},
		vendor.RetConstant{
			Val: 0,
		},
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
		0x01, 0x02,
	})
	if err != nil {
		t.Fatalf("unexpected error while running program: %v", err)
	}
	if want, got := 2, out; want != got {
		t.Fatalf("unexpected number of output bytes:\n- want: %d\n-  got: %d",
			want, got)
	}
}

func TestVMJumpIfBitsNotSet(t *testing.T) {
	vm, done, err := vendor.testVM(t, []vendor.Instruction{
		vendor.LoadAbsolute{
			Off:  8,
			Size: 2,
		},
		vendor.JumpIf{
			Cond:     vendor.JumpBitsNotSet,
			Val:      0x1221,
			SkipTrue: 1,
		},
		vendor.RetConstant{
			Val: 0,
		},
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
		0x01, 0x02,
	})
	if err != nil {
		t.Fatalf("unexpected error while running program: %v", err)
	}
	if want, got := 2, out; want != got {
		t.Fatalf("unexpected number of output bytes:\n- want: %d\n-  got: %d",
			want, got)
	}
}

func TestVMJumpIfXEqual(t *testing.T) {
	vm, done, err := vendor.testVM(t, []vendor.Instruction{
		vendor.LoadAbsolute{
			Off:  8,
			Size: 1,
		},
		vendor.LoadConstant{
			Dst: vendor.RegX,
			Val: 1,
		},
		vendor.JumpIfX{
			Cond:     vendor.JumpEqual,
			SkipTrue: 1,
		},
		vendor.RetConstant{
			Val: 0,
		},
		vendor.RetConstant{
			Val: 9,
		},
	})
	if err != nil {
		t.Fatalf("failed to load BPF program: %v", err)
	}
	defer done()

	out, err := vm.Run([]byte{
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		1,
	})
	if err != nil {
		t.Fatalf("unexpected error while running program: %v", err)
	}
	if want, got := 1, out; want != got {
		t.Fatalf("unexpected number of output bytes:\n- want: %d\n-  got: %d",
			want, got)
	}
}

func TestVMJumpIfXNotEqual(t *testing.T) {
	vm, done, err := vendor.testVM(t, []vendor.Instruction{
		vendor.LoadAbsolute{
			Off:  8,
			Size: 1,
		},
		vendor.LoadConstant{
			Dst: vendor.RegX,
			Val: 1,
		},
		vendor.JumpIfX{
			Cond:      vendor.JumpNotEqual,
			SkipFalse: 1,
		},
		vendor.RetConstant{
			Val: 0,
		},
		vendor.RetConstant{
			Val: 9,
		},
	})
	if err != nil {
		t.Fatalf("failed to load BPF program: %v", err)
	}
	defer done()

	out, err := vm.Run([]byte{
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		1,
	})
	if err != nil {
		t.Fatalf("unexpected error while running program: %v", err)
	}
	if want, got := 1, out; want != got {
		t.Fatalf("unexpected number of output bytes:\n- want: %d\n-  got: %d",
			want, got)
	}
}

func TestVMJumpIfXGreaterThan(t *testing.T) {
	vm, done, err := vendor.testVM(t, []vendor.Instruction{
		vendor.LoadAbsolute{
			Off:  8,
			Size: 4,
		},
		vendor.LoadConstant{
			Dst: vendor.RegX,
			Val: 0x00010202,
		},
		vendor.JumpIfX{
			Cond:     vendor.JumpGreaterThan,
			SkipTrue: 1,
		},
		vendor.RetConstant{
			Val: 0,
		},
		vendor.RetConstant{
			Val: 12,
		},
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
	if want, got := 4, out; want != got {
		t.Fatalf("unexpected number of output bytes:\n- want: %d\n-  got: %d",
			want, got)
	}
}

func TestVMJumpIfXLessThan(t *testing.T) {
	vm, done, err := vendor.testVM(t, []vendor.Instruction{
		vendor.LoadAbsolute{
			Off:  8,
			Size: 4,
		},
		vendor.LoadConstant{
			Dst: vendor.RegX,
			Val: 0xff010203,
		},
		vendor.JumpIfX{
			Cond:     vendor.JumpLessThan,
			SkipTrue: 1,
		},
		vendor.RetConstant{
			Val: 0,
		},
		vendor.RetConstant{
			Val: 12,
		},
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
	if want, got := 4, out; want != got {
		t.Fatalf("unexpected number of output bytes:\n- want: %d\n-  got: %d",
			want, got)
	}
}

func TestVMJumpIfXGreaterOrEqual(t *testing.T) {
	vm, done, err := vendor.testVM(t, []vendor.Instruction{
		vendor.LoadAbsolute{
			Off:  8,
			Size: 4,
		},
		vendor.LoadConstant{
			Dst: vendor.RegX,
			Val: 0x00010203,
		},
		vendor.JumpIfX{
			Cond:     vendor.JumpGreaterOrEqual,
			SkipTrue: 1,
		},
		vendor.RetConstant{
			Val: 0,
		},
		vendor.RetConstant{
			Val: 12,
		},
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
	if want, got := 4, out; want != got {
		t.Fatalf("unexpected number of output bytes:\n- want: %d\n-  got: %d",
			want, got)
	}
}

func TestVMJumpIfXLessOrEqual(t *testing.T) {
	vm, done, err := vendor.testVM(t, []vendor.Instruction{
		vendor.LoadAbsolute{
			Off:  8,
			Size: 4,
		},
		vendor.LoadConstant{
			Dst: vendor.RegX,
			Val: 0xff010203,
		},
		vendor.JumpIfX{
			Cond:     vendor.JumpLessOrEqual,
			SkipTrue: 1,
		},
		vendor.RetConstant{
			Val: 0,
		},
		vendor.RetConstant{
			Val: 12,
		},
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
	if want, got := 4, out; want != got {
		t.Fatalf("unexpected number of output bytes:\n- want: %d\n-  got: %d",
			want, got)
	}
}

func TestVMJumpIfXBitsSet(t *testing.T) {
	vm, done, err := vendor.testVM(t, []vendor.Instruction{
		vendor.LoadAbsolute{
			Off:  8,
			Size: 2,
		},
		vendor.LoadConstant{
			Dst: vendor.RegX,
			Val: 0x1122,
		},
		vendor.JumpIfX{
			Cond:     vendor.JumpBitsSet,
			SkipTrue: 1,
		},
		vendor.RetConstant{
			Val: 0,
		},
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
		0x01, 0x02,
	})
	if err != nil {
		t.Fatalf("unexpected error while running program: %v", err)
	}
	if want, got := 2, out; want != got {
		t.Fatalf("unexpected number of output bytes:\n- want: %d\n-  got: %d",
			want, got)
	}
}

func TestVMJumpIfXBitsNotSet(t *testing.T) {
	vm, done, err := vendor.testVM(t, []vendor.Instruction{
		vendor.LoadAbsolute{
			Off:  8,
			Size: 2,
		},
		vendor.LoadConstant{
			Dst: vendor.RegX,
			Val: 0x1221,
		},
		vendor.JumpIfX{
			Cond:     vendor.JumpBitsNotSet,
			SkipTrue: 1,
		},
		vendor.RetConstant{
			Val: 0,
		},
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
		0x01, 0x02,
	})
	if err != nil {
		t.Fatalf("unexpected error while running program: %v", err)
	}
	if want, got := 2, out; want != got {
		t.Fatalf("unexpected number of output bytes:\n- want: %d\n-  got: %d",
			want, got)
	}
}
