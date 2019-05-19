// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bpf_test

import (
	"testing"
	"vendor"
)

func TestVMLoadExtensionNotImplemented(t *testing.T) {
	_, _, err := vendor.testVM(t, []vendor.Instruction{
		vendor.LoadExtension{
			Num: 100,
		},
		vendor.RetA{},
	})
	if vendor.errStr(err) != "extension 100 not implemented" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVMLoadExtensionExtLen(t *testing.T) {
	vm, done, err := vendor.testVM(t, []vendor.Instruction{
		vendor.LoadExtension{
			Num: vendor.ExtLen,
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
	if want, got := 4, out; want != got {
		t.Fatalf("unexpected number of output bytes:\n- want: %d\n-  got: %d",
			want, got)
	}
}
