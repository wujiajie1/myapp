// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bpf

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"vendor"
)

// This is a direct translation of the program in
// testdata/all_instructions.txt.
var allInstructions = []vendor.Instruction{
	vendor.LoadConstant{Dst: vendor.RegA, Val: 42},
	vendor.LoadConstant{Dst: vendor.RegX, Val: 42},

	vendor.LoadScratch{Dst: vendor.RegA, N: 3},
	vendor.LoadScratch{Dst: vendor.RegX, N: 3},

	vendor.LoadAbsolute{Off: 42, Size: 1},
	vendor.LoadAbsolute{Off: 42, Size: 2},
	vendor.LoadAbsolute{Off: 42, Size: 4},

	vendor.LoadIndirect{Off: 42, Size: 1},
	vendor.LoadIndirect{Off: 42, Size: 2},
	vendor.LoadIndirect{Off: 42, Size: 4},

	vendor.LoadMemShift{Off: 42},

	vendor.LoadExtension{Num: vendor.ExtLen},
	vendor.LoadExtension{Num: vendor.ExtProto},
	vendor.LoadExtension{Num: vendor.ExtType},
	vendor.LoadExtension{Num: vendor.ExtRand},

	vendor.StoreScratch{Src: vendor.RegA, N: 3},
	vendor.StoreScratch{Src: vendor.RegX, N: 3},

	vendor.ALUOpConstant{Op: vendor.ALUOpAdd, Val: 42},
	vendor.ALUOpConstant{Op: vendor.ALUOpSub, Val: 42},
	vendor.ALUOpConstant{Op: vendor.ALUOpMul, Val: 42},
	vendor.ALUOpConstant{Op: vendor.ALUOpDiv, Val: 42},
	vendor.ALUOpConstant{Op: vendor.ALUOpOr, Val: 42},
	vendor.ALUOpConstant{Op: vendor.ALUOpAnd, Val: 42},
	vendor.ALUOpConstant{Op: vendor.ALUOpShiftLeft, Val: 42},
	vendor.ALUOpConstant{Op: vendor.ALUOpShiftRight, Val: 42},
	vendor.ALUOpConstant{Op: vendor.ALUOpMod, Val: 42},
	vendor.ALUOpConstant{Op: vendor.ALUOpXor, Val: 42},

	vendor.ALUOpX{Op: vendor.ALUOpAdd},
	vendor.ALUOpX{Op: vendor.ALUOpSub},
	vendor.ALUOpX{Op: vendor.ALUOpMul},
	vendor.ALUOpX{Op: vendor.ALUOpDiv},
	vendor.ALUOpX{Op: vendor.ALUOpOr},
	vendor.ALUOpX{Op: vendor.ALUOpAnd},
	vendor.ALUOpX{Op: vendor.ALUOpShiftLeft},
	vendor.ALUOpX{Op: vendor.ALUOpShiftRight},
	vendor.ALUOpX{Op: vendor.ALUOpMod},
	vendor.ALUOpX{Op: vendor.ALUOpXor},

	vendor.NegateA{},

	vendor.Jump{Skip: 17},
	vendor.JumpIf{Cond: vendor.JumpEqual, Val: 42, SkipTrue: 15, SkipFalse: 16},
	vendor.JumpIf{Cond: vendor.JumpNotEqual, Val: 42, SkipTrue: 15},
	vendor.JumpIf{Cond: vendor.JumpLessThan, Val: 42, SkipTrue: 14},
	vendor.JumpIf{Cond: vendor.JumpLessOrEqual, Val: 42, SkipTrue: 13},
	vendor.JumpIf{Cond: vendor.JumpGreaterThan, Val: 42, SkipTrue: 11, SkipFalse: 12},
	vendor.JumpIf{Cond: vendor.JumpGreaterOrEqual, Val: 42, SkipTrue: 10, SkipFalse: 11},
	vendor.JumpIf{Cond: vendor.JumpBitsSet, Val: 42, SkipTrue: 9, SkipFalse: 10},

	vendor.JumpIfX{Cond: vendor.JumpEqual, SkipTrue: 8, SkipFalse: 9},
	vendor.JumpIfX{Cond: vendor.JumpNotEqual, SkipTrue: 8},
	vendor.JumpIfX{Cond: vendor.JumpLessThan, SkipTrue: 7},
	vendor.JumpIfX{Cond: vendor.JumpLessOrEqual, SkipTrue: 6},
	vendor.JumpIfX{Cond: vendor.JumpGreaterThan, SkipTrue: 4, SkipFalse: 5},
	vendor.JumpIfX{Cond: vendor.JumpGreaterOrEqual, SkipTrue: 3, SkipFalse: 4},
	vendor.JumpIfX{Cond: vendor.JumpBitsSet, SkipTrue: 2, SkipFalse: 3},

	vendor.TAX{},
	vendor.TXA{},

	vendor.RetA{},
	vendor.RetConstant{Val: 42},
}
var allInstructionsExpected = "testdata/all_instructions.bpf"

// Check that we produce the same output as the canonical bpf_asm
// linux kernel tool.
func TestInterop(t *testing.T) {
	out, err := vendor.Assemble(allInstructions)
	if err != nil {
		t.Fatalf("assembly of allInstructions program failed: %s", err)
	}
	t.Logf("Assembled program is %d instructions long", len(out))

	bs, err := ioutil.ReadFile(allInstructionsExpected)
	if err != nil {
		t.Fatalf("reading %s: %s", allInstructionsExpected, err)
	}
	// First statement is the number of statements, last statement is
	// empty. We just ignore both and rely on slice length.
	stmts := strings.Split(string(bs), ",")
	if len(stmts)-2 != len(out) {
		t.Fatalf("test program lengths don't match: %s has %d, Go implementation has %d", allInstructionsExpected, len(stmts)-2, len(allInstructions))
	}

	for i, stmt := range stmts[1 : len(stmts)-2] {
		nums := strings.Split(stmt, " ")
		if len(nums) != 4 {
			t.Fatalf("malformed instruction %d in %s: %s", i+1, allInstructionsExpected, stmt)
		}

		actual := out[i]

		op, err := strconv.ParseUint(nums[0], 10, 16)
		if err != nil {
			t.Fatalf("malformed opcode %s in instruction %d of %s", nums[0], i+1, allInstructionsExpected)
		}
		if actual.Op != uint16(op) {
			t.Errorf("opcode mismatch on instruction %d (%#v): got 0x%02x, want 0x%02x", i+1, allInstructions[i], actual.Op, op)
		}

		jt, err := strconv.ParseUint(nums[1], 10, 8)
		if err != nil {
			t.Fatalf("malformed jt offset %s in instruction %d of %s", nums[1], i+1, allInstructionsExpected)
		}
		if actual.Jt != uint8(jt) {
			t.Errorf("jt mismatch on instruction %d (%#v): got %d, want %d", i+1, allInstructions[i], actual.Jt, jt)
		}

		jf, err := strconv.ParseUint(nums[2], 10, 8)
		if err != nil {
			t.Fatalf("malformed jf offset %s in instruction %d of %s", nums[2], i+1, allInstructionsExpected)
		}
		if actual.Jf != uint8(jf) {
			t.Errorf("jf mismatch on instruction %d (%#v): got %d, want %d", i+1, allInstructions[i], actual.Jf, jf)
		}

		k, err := strconv.ParseUint(nums[3], 10, 32)
		if err != nil {
			t.Fatalf("malformed constant %s in instruction %d of %s", nums[3], i+1, allInstructionsExpected)
		}
		if actual.K != uint32(k) {
			t.Errorf("constant mismatch on instruction %d (%#v): got %d, want %d", i+1, allInstructions[i], actual.K, k)
		}
	}
}

// Check that assembly and disassembly match each other.
func TestAsmDisasm(t *testing.T) {
	prog1, err := vendor.Assemble(allInstructions)
	if err != nil {
		t.Fatalf("assembly of allInstructions program failed: %s", err)
	}
	t.Logf("Assembled program is %d instructions long", len(prog1))

	got, allDecoded := vendor.Disassemble(prog1)
	if !allDecoded {
		t.Errorf("Disassemble(Assemble(allInstructions)) produced unrecognized instructions:")
		for i, inst := range got {
			if r, ok := inst.(vendor.RawInstruction); ok {
				t.Logf("  insn %d, %#v --> %#v", i+1, allInstructions[i], r)
			}
		}
	}

	if len(allInstructions) != len(got) {
		t.Fatalf("disassembly changed program size: %d insns before, %d insns after", len(allInstructions), len(got))
	}
	if !reflect.DeepEqual(allInstructions, got) {
		t.Errorf("program mutated by disassembly:")
		for i := range got {
			if !reflect.DeepEqual(allInstructions[i], got[i]) {
				t.Logf("  insn %d, s: %#v, p1: %#v, got: %#v", i+1, allInstructions[i], prog1[i], got[i])
			}
		}
	}
}

type InvalidInstruction struct{}

func (a InvalidInstruction) Assemble() (vendor.RawInstruction, error) {
	return vendor.RawInstruction{}, fmt.Errorf("Invalid Instruction")
}

func (a InvalidInstruction) String() string {
	return fmt.Sprintf("unknown instruction: %#v", a)
}

func TestString(t *testing.T) {
	testCases := []struct {
		instruction vendor.Instruction
		assembler   string
	}{
		{
			instruction: vendor.LoadConstant{Dst: vendor.RegA, Val: 42},
			assembler:   "ld #42",
		},
		{
			instruction: vendor.LoadConstant{Dst: vendor.RegX, Val: 42},
			assembler:   "ldx #42",
		},
		{
			instruction: vendor.LoadConstant{Dst: 0xffff, Val: 42},
			assembler:   "unknown instruction: bpf.LoadConstant{Dst:0xffff, Val:0x2a}",
		},
		{
			instruction: vendor.LoadScratch{Dst: vendor.RegA, N: 3},
			assembler:   "ld M[3]",
		},
		{
			instruction: vendor.LoadScratch{Dst: vendor.RegX, N: 3},
			assembler:   "ldx M[3]",
		},
		{
			instruction: vendor.LoadScratch{Dst: 0xffff, N: 3},
			assembler:   "unknown instruction: bpf.LoadScratch{Dst:0xffff, N:3}",
		},
		{
			instruction: vendor.LoadAbsolute{Off: 42, Size: 1},
			assembler:   "ldb [42]",
		},
		{
			instruction: vendor.LoadAbsolute{Off: 42, Size: 2},
			assembler:   "ldh [42]",
		},
		{
			instruction: vendor.LoadAbsolute{Off: 42, Size: 4},
			assembler:   "ld [42]",
		},
		{
			instruction: vendor.LoadAbsolute{Off: 42, Size: -1},
			assembler:   "unknown instruction: bpf.LoadAbsolute{Off:0x2a, Size:-1}",
		},
		{
			instruction: vendor.LoadIndirect{Off: 42, Size: 1},
			assembler:   "ldb [x + 42]",
		},
		{
			instruction: vendor.LoadIndirect{Off: 42, Size: 2},
			assembler:   "ldh [x + 42]",
		},
		{
			instruction: vendor.LoadIndirect{Off: 42, Size: 4},
			assembler:   "ld [x + 42]",
		},
		{
			instruction: vendor.LoadIndirect{Off: 42, Size: -1},
			assembler:   "unknown instruction: bpf.LoadIndirect{Off:0x2a, Size:-1}",
		},
		{
			instruction: vendor.LoadMemShift{Off: 42},
			assembler:   "ldx 4*([42]&0xf)",
		},
		{
			instruction: vendor.LoadExtension{Num: vendor.ExtLen},
			assembler:   "ld #len",
		},
		{
			instruction: vendor.LoadExtension{Num: vendor.ExtProto},
			assembler:   "ld #proto",
		},
		{
			instruction: vendor.LoadExtension{Num: vendor.ExtType},
			assembler:   "ld #type",
		},
		{
			instruction: vendor.LoadExtension{Num: vendor.ExtPayloadOffset},
			assembler:   "ld #poff",
		},
		{
			instruction: vendor.LoadExtension{Num: vendor.ExtInterfaceIndex},
			assembler:   "ld #ifidx",
		},
		{
			instruction: vendor.LoadExtension{Num: vendor.ExtNetlinkAttr},
			assembler:   "ld #nla",
		},
		{
			instruction: vendor.LoadExtension{Num: vendor.ExtNetlinkAttrNested},
			assembler:   "ld #nlan",
		},
		{
			instruction: vendor.LoadExtension{Num: vendor.ExtMark},
			assembler:   "ld #mark",
		},
		{
			instruction: vendor.LoadExtension{Num: vendor.ExtQueue},
			assembler:   "ld #queue",
		},
		{
			instruction: vendor.LoadExtension{Num: vendor.ExtLinkLayerType},
			assembler:   "ld #hatype",
		},
		{
			instruction: vendor.LoadExtension{Num: vendor.ExtRXHash},
			assembler:   "ld #rxhash",
		},
		{
			instruction: vendor.LoadExtension{Num: vendor.ExtCPUID},
			assembler:   "ld #cpu",
		},
		{
			instruction: vendor.LoadExtension{Num: vendor.ExtVLANTag},
			assembler:   "ld #vlan_tci",
		},
		{
			instruction: vendor.LoadExtension{Num: vendor.ExtVLANTagPresent},
			assembler:   "ld #vlan_avail",
		},
		{
			instruction: vendor.LoadExtension{Num: vendor.ExtVLANProto},
			assembler:   "ld #vlan_tpid",
		},
		{
			instruction: vendor.LoadExtension{Num: vendor.ExtRand},
			assembler:   "ld #rand",
		},
		{
			instruction: vendor.LoadAbsolute{Off: 0xfffff038, Size: 4},
			assembler:   "ld #rand",
		},
		{
			instruction: vendor.LoadExtension{Num: 0xfff},
			assembler:   "unknown instruction: bpf.LoadExtension{Num:4095}",
		},
		{
			instruction: vendor.StoreScratch{Src: vendor.RegA, N: 3},
			assembler:   "st M[3]",
		},
		{
			instruction: vendor.StoreScratch{Src: vendor.RegX, N: 3},
			assembler:   "stx M[3]",
		},
		{
			instruction: vendor.StoreScratch{Src: 0xffff, N: 3},
			assembler:   "unknown instruction: bpf.StoreScratch{Src:0xffff, N:3}",
		},
		{
			instruction: vendor.ALUOpConstant{Op: vendor.ALUOpAdd, Val: 42},
			assembler:   "add #42",
		},
		{
			instruction: vendor.ALUOpConstant{Op: vendor.ALUOpSub, Val: 42},
			assembler:   "sub #42",
		},
		{
			instruction: vendor.ALUOpConstant{Op: vendor.ALUOpMul, Val: 42},
			assembler:   "mul #42",
		},
		{
			instruction: vendor.ALUOpConstant{Op: vendor.ALUOpDiv, Val: 42},
			assembler:   "div #42",
		},
		{
			instruction: vendor.ALUOpConstant{Op: vendor.ALUOpOr, Val: 42},
			assembler:   "or #42",
		},
		{
			instruction: vendor.ALUOpConstant{Op: vendor.ALUOpAnd, Val: 42},
			assembler:   "and #42",
		},
		{
			instruction: vendor.ALUOpConstant{Op: vendor.ALUOpShiftLeft, Val: 42},
			assembler:   "lsh #42",
		},
		{
			instruction: vendor.ALUOpConstant{Op: vendor.ALUOpShiftRight, Val: 42},
			assembler:   "rsh #42",
		},
		{
			instruction: vendor.ALUOpConstant{Op: vendor.ALUOpMod, Val: 42},
			assembler:   "mod #42",
		},
		{
			instruction: vendor.ALUOpConstant{Op: vendor.ALUOpXor, Val: 42},
			assembler:   "xor #42",
		},
		{
			instruction: vendor.ALUOpConstant{Op: 0xffff, Val: 42},
			assembler:   "unknown instruction: bpf.ALUOpConstant{Op:0xffff, Val:0x2a}",
		},
		{
			instruction: vendor.ALUOpX{Op: vendor.ALUOpAdd},
			assembler:   "add x",
		},
		{
			instruction: vendor.ALUOpX{Op: vendor.ALUOpSub},
			assembler:   "sub x",
		},
		{
			instruction: vendor.ALUOpX{Op: vendor.ALUOpMul},
			assembler:   "mul x",
		},
		{
			instruction: vendor.ALUOpX{Op: vendor.ALUOpDiv},
			assembler:   "div x",
		},
		{
			instruction: vendor.ALUOpX{Op: vendor.ALUOpOr},
			assembler:   "or x",
		},
		{
			instruction: vendor.ALUOpX{Op: vendor.ALUOpAnd},
			assembler:   "and x",
		},
		{
			instruction: vendor.ALUOpX{Op: vendor.ALUOpShiftLeft},
			assembler:   "lsh x",
		},
		{
			instruction: vendor.ALUOpX{Op: vendor.ALUOpShiftRight},
			assembler:   "rsh x",
		},
		{
			instruction: vendor.ALUOpX{Op: vendor.ALUOpMod},
			assembler:   "mod x",
		},
		{
			instruction: vendor.ALUOpX{Op: vendor.ALUOpXor},
			assembler:   "xor x",
		},
		{
			instruction: vendor.ALUOpX{Op: 0xffff},
			assembler:   "unknown instruction: bpf.ALUOpX{Op:0xffff}",
		},
		{
			instruction: vendor.NegateA{},
			assembler:   "neg",
		},
		{
			instruction: vendor.Jump{Skip: 10},
			assembler:   "ja 10",
		},
		{
			instruction: vendor.JumpIf{Cond: vendor.JumpEqual, Val: 42, SkipTrue: 8, SkipFalse: 9},
			assembler:   "jeq #42,8,9",
		},
		{
			instruction: vendor.JumpIf{Cond: vendor.JumpEqual, Val: 42, SkipTrue: 8},
			assembler:   "jeq #42,8",
		},
		{
			instruction: vendor.JumpIf{Cond: vendor.JumpEqual, Val: 42, SkipFalse: 8},
			assembler:   "jneq #42,8",
		},
		{
			instruction: vendor.JumpIf{Cond: vendor.JumpNotEqual, Val: 42, SkipTrue: 8},
			assembler:   "jneq #42,8",
		},
		{
			instruction: vendor.JumpIf{Cond: vendor.JumpLessThan, Val: 42, SkipTrue: 7},
			assembler:   "jlt #42,7",
		},
		{
			instruction: vendor.JumpIf{Cond: vendor.JumpLessOrEqual, Val: 42, SkipTrue: 6},
			assembler:   "jle #42,6",
		},
		{
			instruction: vendor.JumpIf{Cond: vendor.JumpGreaterThan, Val: 42, SkipTrue: 4, SkipFalse: 5},
			assembler:   "jgt #42,4,5",
		},
		{
			instruction: vendor.JumpIf{Cond: vendor.JumpGreaterThan, Val: 42, SkipTrue: 4},
			assembler:   "jgt #42,4",
		},
		{
			instruction: vendor.JumpIf{Cond: vendor.JumpGreaterOrEqual, Val: 42, SkipTrue: 3, SkipFalse: 4},
			assembler:   "jge #42,3,4",
		},
		{
			instruction: vendor.JumpIf{Cond: vendor.JumpGreaterOrEqual, Val: 42, SkipTrue: 3},
			assembler:   "jge #42,3",
		},
		{
			instruction: vendor.JumpIf{Cond: vendor.JumpBitsSet, Val: 42, SkipTrue: 2, SkipFalse: 3},
			assembler:   "jset #42,2,3",
		},
		{
			instruction: vendor.JumpIf{Cond: vendor.JumpBitsSet, Val: 42, SkipTrue: 2},
			assembler:   "jset #42,2",
		},
		{
			instruction: vendor.JumpIf{Cond: vendor.JumpBitsNotSet, Val: 42, SkipTrue: 2, SkipFalse: 3},
			assembler:   "jset #42,3,2",
		},
		{
			instruction: vendor.JumpIf{Cond: vendor.JumpBitsNotSet, Val: 42, SkipTrue: 2},
			assembler:   "jset #42,0,2",
		},
		{
			instruction: vendor.JumpIf{Cond: 0xffff, Val: 42, SkipTrue: 1, SkipFalse: 2},
			assembler:   "unknown JumpTest 0xffff",
		},
		{
			instruction: vendor.JumpIfX{Cond: vendor.JumpEqual, SkipTrue: 8, SkipFalse: 9},
			assembler:   "jeq x,8,9",
		},
		{
			instruction: vendor.JumpIfX{Cond: vendor.JumpEqual, SkipTrue: 8},
			assembler:   "jeq x,8",
		},
		{
			instruction: vendor.JumpIfX{Cond: vendor.JumpEqual, SkipFalse: 8},
			assembler:   "jneq x,8",
		},
		{
			instruction: vendor.JumpIfX{Cond: vendor.JumpNotEqual, SkipTrue: 8},
			assembler:   "jneq x,8",
		},
		{
			instruction: vendor.JumpIfX{Cond: vendor.JumpLessThan, SkipTrue: 7},
			assembler:   "jlt x,7",
		},
		{
			instruction: vendor.JumpIfX{Cond: vendor.JumpLessOrEqual, SkipTrue: 6},
			assembler:   "jle x,6",
		},
		{
			instruction: vendor.JumpIfX{Cond: vendor.JumpGreaterThan, SkipTrue: 4, SkipFalse: 5},
			assembler:   "jgt x,4,5",
		},
		{
			instruction: vendor.JumpIfX{Cond: vendor.JumpGreaterThan, SkipTrue: 4},
			assembler:   "jgt x,4",
		},
		{
			instruction: vendor.JumpIfX{Cond: vendor.JumpGreaterOrEqual, SkipTrue: 3, SkipFalse: 4},
			assembler:   "jge x,3,4",
		},
		{
			instruction: vendor.JumpIfX{Cond: vendor.JumpGreaterOrEqual, SkipTrue: 3},
			assembler:   "jge x,3",
		},
		{
			instruction: vendor.JumpIfX{Cond: vendor.JumpBitsSet, SkipTrue: 2, SkipFalse: 3},
			assembler:   "jset x,2,3",
		},
		{
			instruction: vendor.JumpIfX{Cond: vendor.JumpBitsSet, SkipTrue: 2},
			assembler:   "jset x,2",
		},
		{
			instruction: vendor.JumpIfX{Cond: vendor.JumpBitsNotSet, SkipTrue: 2, SkipFalse: 3},
			assembler:   "jset x,3,2",
		},
		{
			instruction: vendor.JumpIfX{Cond: vendor.JumpBitsNotSet, SkipTrue: 2},
			assembler:   "jset x,0,2",
		},
		{
			instruction: vendor.JumpIfX{Cond: 0xffff, SkipTrue: 1, SkipFalse: 2},
			assembler:   "unknown JumpTest 0xffff",
		},
		{
			instruction: vendor.TAX{},
			assembler:   "tax",
		},
		{
			instruction: vendor.TXA{},
			assembler:   "txa",
		},
		{
			instruction: vendor.RetA{},
			assembler:   "ret a",
		},
		{
			instruction: vendor.RetConstant{Val: 42},
			assembler:   "ret #42",
		},
		// Invalid instruction
		{
			instruction: InvalidInstruction{},
			assembler:   "unknown instruction: bpf.InvalidInstruction{}",
		},
	}

	for _, testCase := range testCases {
		if input, ok := testCase.instruction.(fmt.Stringer); ok {
			got := input.String()
			if got != testCase.assembler {
				t.Errorf("String did not return expected assembler notation, expected: %s, got: %s", testCase.assembler, got)
			}
		} else {
			t.Errorf("Instruction %#v is not a fmt.Stringer", testCase.instruction)
		}
	}
}
