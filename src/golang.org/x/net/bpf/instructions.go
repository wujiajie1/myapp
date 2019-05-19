// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bpf

import (
	"fmt"
	"vendor"
)

// An Instruction is one instruction executed by the BPF virtual
// machine.
type Instruction interface {
	// Assemble assembles the Instruction into a RawInstruction.
	Assemble() (RawInstruction, error)
}

// A RawInstruction is a raw BPF virtual machine instruction.
type RawInstruction struct {
	// Operation to execute.
	Op uint16
	// For conditional jump instructions, the number of instructions
	// to skip if the condition is true/false.
	Jt uint8
	Jf uint8
	// Constant parameter. The meaning depends on the Op.
	K uint32
}

// Assemble implements the Instruction Assemble method.
func (ri RawInstruction) Assemble() (RawInstruction, error) { return ri, nil }

// Disassemble parses ri into an Instruction and returns it. If ri is
// not recognized by this package, ri itself is returned.
func (ri RawInstruction) Disassemble() Instruction {
	switch ri.Op & vendor.opMaskCls {
	case vendor.opClsLoadA, vendor.opClsLoadX:
		reg := vendor.Register(ri.Op & vendor.opMaskLoadDest)
		sz := 0
		switch ri.Op & vendor.opMaskLoadWidth {
		case vendor.opLoadWidth4:
			sz = 4
		case vendor.opLoadWidth2:
			sz = 2
		case vendor.opLoadWidth1:
			sz = 1
		default:
			return ri
		}
		switch ri.Op & vendor.opMaskLoadMode {
		case vendor.opAddrModeImmediate:
			if sz != 4 {
				return ri
			}
			return LoadConstant{Dst: reg, Val: ri.K}
		case vendor.opAddrModeScratch:
			if sz != 4 || ri.K > 15 {
				return ri
			}
			return LoadScratch{Dst: reg, N: int(ri.K)}
		case vendor.opAddrModeAbsolute:
			if ri.K > vendor.extOffset+0xffffffff {
				return LoadExtension{Num: vendor.Extension(-vendor.extOffset + ri.K)}
			}
			return LoadAbsolute{Size: sz, Off: ri.K}
		case vendor.opAddrModeIndirect:
			return LoadIndirect{Size: sz, Off: ri.K}
		case vendor.opAddrModePacketLen:
			if sz != 4 {
				return ri
			}
			return LoadExtension{Num: vendor.ExtLen}
		case vendor.opAddrModeMemShift:
			return LoadMemShift{Off: ri.K}
		default:
			return ri
		}

	case vendor.opClsStoreA:
		if ri.Op != vendor.opClsStoreA || ri.K > 15 {
			return ri
		}
		return StoreScratch{Src: vendor.RegA, N: int(ri.K)}

	case vendor.opClsStoreX:
		if ri.Op != vendor.opClsStoreX || ri.K > 15 {
			return ri
		}
		return StoreScratch{Src: vendor.RegX, N: int(ri.K)}

	case vendor.opClsALU:
		switch op := vendor.ALUOp(ri.Op & vendor.opMaskOperator); op {
		case vendor.ALUOpAdd, vendor.ALUOpSub, vendor.ALUOpMul, vendor.ALUOpDiv, vendor.ALUOpOr, vendor.ALUOpAnd, vendor.ALUOpShiftLeft, vendor.ALUOpShiftRight, vendor.ALUOpMod, vendor.ALUOpXor:
			switch operand := vendor.opOperand(ri.Op & vendor.opMaskOperand); operand {
			case vendor.opOperandX:
				return ALUOpX{Op: op}
			case vendor.opOperandConstant:
				return ALUOpConstant{Op: op, Val: ri.K}
			default:
				return ri
			}
		case vendor.aluOpNeg:
			return NegateA{}
		default:
			return ri
		}

	case vendor.opClsJump:
		switch op := vendor.jumpOp(ri.Op & vendor.opMaskOperator); op {
		case vendor.opJumpAlways:
			return Jump{Skip: ri.K}
		case vendor.opJumpEqual, vendor.opJumpGT, vendor.opJumpGE, vendor.opJumpSet:
			cond, skipTrue, skipFalse := jumpOpToTest(op, ri.Jt, ri.Jf)
			switch operand := vendor.opOperand(ri.Op & vendor.opMaskOperand); operand {
			case vendor.opOperandX:
				return JumpIfX{Cond: cond, SkipTrue: skipTrue, SkipFalse: skipFalse}
			case vendor.opOperandConstant:
				return JumpIf{Cond: cond, Val: ri.K, SkipTrue: skipTrue, SkipFalse: skipFalse}
			default:
				return ri
			}
		default:
			return ri
		}

	case vendor.opClsReturn:
		switch ri.Op {
		case vendor.opClsReturn | vendor.opRetSrcA:
			return RetA{}
		case vendor.opClsReturn | vendor.opRetSrcConstant:
			return RetConstant{Val: ri.K}
		default:
			return ri
		}

	case vendor.opClsMisc:
		switch ri.Op {
		case vendor.opClsMisc | vendor.opMiscTAX:
			return TAX{}
		case vendor.opClsMisc | vendor.opMiscTXA:
			return TXA{}
		default:
			return ri
		}

	default:
		panic("unreachable") // switch is exhaustive on the bit pattern
	}
}

func jumpOpToTest(op vendor.jumpOp, skipTrue uint8, skipFalse uint8) (vendor.JumpTest, uint8, uint8) {
	var test vendor.JumpTest

	// Decode "fake" jump conditions that don't appear in machine code
	// Ensures the Assemble -> Disassemble stage recreates the same instructions
	// See https://github.com/golang/go/issues/18470
	if skipTrue == 0 {
		switch op {
		case vendor.opJumpEqual:
			test = vendor.JumpNotEqual
		case vendor.opJumpGT:
			test = vendor.JumpLessOrEqual
		case vendor.opJumpGE:
			test = vendor.JumpLessThan
		case vendor.opJumpSet:
			test = vendor.JumpBitsNotSet
		}

		return test, skipFalse, 0
	}

	switch op {
	case vendor.opJumpEqual:
		test = vendor.JumpEqual
	case vendor.opJumpGT:
		test = vendor.JumpGreaterThan
	case vendor.opJumpGE:
		test = vendor.JumpGreaterOrEqual
	case vendor.opJumpSet:
		test = vendor.JumpBitsSet
	}

	return test, skipTrue, skipFalse
}

// LoadConstant loads Val into register Dst.
type LoadConstant struct {
	Dst vendor.Register
	Val uint32
}

// Assemble implements the Instruction Assemble method.
func (a LoadConstant) Assemble() (RawInstruction, error) {
	return assembleLoad(a.Dst, 4, vendor.opAddrModeImmediate, a.Val)
}

// String returns the instruction in assembler notation.
func (a LoadConstant) String() string {
	switch a.Dst {
	case vendor.RegA:
		return fmt.Sprintf("ld #%d", a.Val)
	case vendor.RegX:
		return fmt.Sprintf("ldx #%d", a.Val)
	default:
		return fmt.Sprintf("unknown instruction: %#v", a)
	}
}

// LoadScratch loads scratch[N] into register Dst.
type LoadScratch struct {
	Dst vendor.Register
	N   int // 0-15
}

// Assemble implements the Instruction Assemble method.
func (a LoadScratch) Assemble() (RawInstruction, error) {
	if a.N < 0 || a.N > 15 {
		return RawInstruction{}, fmt.Errorf("invalid scratch slot %d", a.N)
	}
	return assembleLoad(a.Dst, 4, vendor.opAddrModeScratch, uint32(a.N))
}

// String returns the instruction in assembler notation.
func (a LoadScratch) String() string {
	switch a.Dst {
	case vendor.RegA:
		return fmt.Sprintf("ld M[%d]", a.N)
	case vendor.RegX:
		return fmt.Sprintf("ldx M[%d]", a.N)
	default:
		return fmt.Sprintf("unknown instruction: %#v", a)
	}
}

// LoadAbsolute loads packet[Off:Off+Size] as an integer value into
// register A.
type LoadAbsolute struct {
	Off  uint32
	Size int // 1, 2 or 4
}

// Assemble implements the Instruction Assemble method.
func (a LoadAbsolute) Assemble() (RawInstruction, error) {
	return assembleLoad(vendor.RegA, a.Size, vendor.opAddrModeAbsolute, a.Off)
}

// String returns the instruction in assembler notation.
func (a LoadAbsolute) String() string {
	switch a.Size {
	case 1: // byte
		return fmt.Sprintf("ldb [%d]", a.Off)
	case 2: // half word
		return fmt.Sprintf("ldh [%d]", a.Off)
	case 4: // word
		if a.Off > vendor.extOffset+0xffffffff {
			return LoadExtension{Num: vendor.Extension(a.Off + 0x1000)}.String()
		}
		return fmt.Sprintf("ld [%d]", a.Off)
	default:
		return fmt.Sprintf("unknown instruction: %#v", a)
	}
}

// LoadIndirect loads packet[X+Off:X+Off+Size] as an integer value
// into register A.
type LoadIndirect struct {
	Off  uint32
	Size int // 1, 2 or 4
}

// Assemble implements the Instruction Assemble method.
func (a LoadIndirect) Assemble() (RawInstruction, error) {
	return assembleLoad(vendor.RegA, a.Size, vendor.opAddrModeIndirect, a.Off)
}

// String returns the instruction in assembler notation.
func (a LoadIndirect) String() string {
	switch a.Size {
	case 1: // byte
		return fmt.Sprintf("ldb [x + %d]", a.Off)
	case 2: // half word
		return fmt.Sprintf("ldh [x + %d]", a.Off)
	case 4: // word
		return fmt.Sprintf("ld [x + %d]", a.Off)
	default:
		return fmt.Sprintf("unknown instruction: %#v", a)
	}
}

// LoadMemShift multiplies the first 4 bits of the byte at packet[Off]
// by 4 and stores the result in register X.
//
// This instruction is mainly useful to load into X the length of an
// IPv4 packet header in a single instruction, rather than have to do
// the arithmetic on the header's first byte by hand.
type LoadMemShift struct {
	Off uint32
}

// Assemble implements the Instruction Assemble method.
func (a LoadMemShift) Assemble() (RawInstruction, error) {
	return assembleLoad(vendor.RegX, 1, vendor.opAddrModeMemShift, a.Off)
}

// String returns the instruction in assembler notation.
func (a LoadMemShift) String() string {
	return fmt.Sprintf("ldx 4*([%d]&0xf)", a.Off)
}

// LoadExtension invokes a linux-specific extension and stores the
// result in register A.
type LoadExtension struct {
	Num vendor.Extension
}

// Assemble implements the Instruction Assemble method.
func (a LoadExtension) Assemble() (RawInstruction, error) {
	if a.Num == vendor.ExtLen {
		return assembleLoad(vendor.RegA, 4, vendor.opAddrModePacketLen, 0)
	}
	return assembleLoad(vendor.RegA, 4, vendor.opAddrModeAbsolute, uint32(vendor.extOffset+a.Num))
}

// String returns the instruction in assembler notation.
func (a LoadExtension) String() string {
	switch a.Num {
	case vendor.ExtLen:
		return "ld #len"
	case vendor.ExtProto:
		return "ld #proto"
	case vendor.ExtType:
		return "ld #type"
	case vendor.ExtPayloadOffset:
		return "ld #poff"
	case vendor.ExtInterfaceIndex:
		return "ld #ifidx"
	case vendor.ExtNetlinkAttr:
		return "ld #nla"
	case vendor.ExtNetlinkAttrNested:
		return "ld #nlan"
	case vendor.ExtMark:
		return "ld #mark"
	case vendor.ExtQueue:
		return "ld #queue"
	case vendor.ExtLinkLayerType:
		return "ld #hatype"
	case vendor.ExtRXHash:
		return "ld #rxhash"
	case vendor.ExtCPUID:
		return "ld #cpu"
	case vendor.ExtVLANTag:
		return "ld #vlan_tci"
	case vendor.ExtVLANTagPresent:
		return "ld #vlan_avail"
	case vendor.ExtVLANProto:
		return "ld #vlan_tpid"
	case vendor.ExtRand:
		return "ld #rand"
	default:
		return fmt.Sprintf("unknown instruction: %#v", a)
	}
}

// StoreScratch stores register Src into scratch[N].
type StoreScratch struct {
	Src vendor.Register
	N   int // 0-15
}

// Assemble implements the Instruction Assemble method.
func (a StoreScratch) Assemble() (RawInstruction, error) {
	if a.N < 0 || a.N > 15 {
		return RawInstruction{}, fmt.Errorf("invalid scratch slot %d", a.N)
	}
	var op uint16
	switch a.Src {
	case vendor.RegA:
		op = vendor.opClsStoreA
	case vendor.RegX:
		op = vendor.opClsStoreX
	default:
		return RawInstruction{}, fmt.Errorf("invalid source register %v", a.Src)
	}

	return RawInstruction{
		Op: op,
		K:  uint32(a.N),
	}, nil
}

// String returns the instruction in assembler notation.
func (a StoreScratch) String() string {
	switch a.Src {
	case vendor.RegA:
		return fmt.Sprintf("st M[%d]", a.N)
	case vendor.RegX:
		return fmt.Sprintf("stx M[%d]", a.N)
	default:
		return fmt.Sprintf("unknown instruction: %#v", a)
	}
}

// ALUOpConstant executes A = A <Op> Val.
type ALUOpConstant struct {
	Op  vendor.ALUOp
	Val uint32
}

// Assemble implements the Instruction Assemble method.
func (a ALUOpConstant) Assemble() (RawInstruction, error) {
	return RawInstruction{
		Op: vendor.opClsALU | uint16(vendor.opOperandConstant) | uint16(a.Op),
		K:  a.Val,
	}, nil
}

// String returns the instruction in assembler notation.
func (a ALUOpConstant) String() string {
	switch a.Op {
	case vendor.ALUOpAdd:
		return fmt.Sprintf("add #%d", a.Val)
	case vendor.ALUOpSub:
		return fmt.Sprintf("sub #%d", a.Val)
	case vendor.ALUOpMul:
		return fmt.Sprintf("mul #%d", a.Val)
	case vendor.ALUOpDiv:
		return fmt.Sprintf("div #%d", a.Val)
	case vendor.ALUOpMod:
		return fmt.Sprintf("mod #%d", a.Val)
	case vendor.ALUOpAnd:
		return fmt.Sprintf("and #%d", a.Val)
	case vendor.ALUOpOr:
		return fmt.Sprintf("or #%d", a.Val)
	case vendor.ALUOpXor:
		return fmt.Sprintf("xor #%d", a.Val)
	case vendor.ALUOpShiftLeft:
		return fmt.Sprintf("lsh #%d", a.Val)
	case vendor.ALUOpShiftRight:
		return fmt.Sprintf("rsh #%d", a.Val)
	default:
		return fmt.Sprintf("unknown instruction: %#v", a)
	}
}

// ALUOpX executes A = A <Op> X
type ALUOpX struct {
	Op vendor.ALUOp
}

// Assemble implements the Instruction Assemble method.
func (a ALUOpX) Assemble() (RawInstruction, error) {
	return RawInstruction{
		Op: vendor.opClsALU | uint16(vendor.opOperandX) | uint16(a.Op),
	}, nil
}

// String returns the instruction in assembler notation.
func (a ALUOpX) String() string {
	switch a.Op {
	case vendor.ALUOpAdd:
		return "add x"
	case vendor.ALUOpSub:
		return "sub x"
	case vendor.ALUOpMul:
		return "mul x"
	case vendor.ALUOpDiv:
		return "div x"
	case vendor.ALUOpMod:
		return "mod x"
	case vendor.ALUOpAnd:
		return "and x"
	case vendor.ALUOpOr:
		return "or x"
	case vendor.ALUOpXor:
		return "xor x"
	case vendor.ALUOpShiftLeft:
		return "lsh x"
	case vendor.ALUOpShiftRight:
		return "rsh x"
	default:
		return fmt.Sprintf("unknown instruction: %#v", a)
	}
}

// NegateA executes A = -A.
type NegateA struct{}

// Assemble implements the Instruction Assemble method.
func (a NegateA) Assemble() (RawInstruction, error) {
	return RawInstruction{
		Op: vendor.opClsALU | uint16(vendor.aluOpNeg),
	}, nil
}

// String returns the instruction in assembler notation.
func (a NegateA) String() string {
	return fmt.Sprintf("neg")
}

// Jump skips the following Skip instructions in the program.
type Jump struct {
	Skip uint32
}

// Assemble implements the Instruction Assemble method.
func (a Jump) Assemble() (RawInstruction, error) {
	return RawInstruction{
		Op: vendor.opClsJump | uint16(vendor.opJumpAlways),
		K:  a.Skip,
	}, nil
}

// String returns the instruction in assembler notation.
func (a Jump) String() string {
	return fmt.Sprintf("ja %d", a.Skip)
}

// JumpIf skips the following Skip instructions in the program if A
// <Cond> Val is true.
type JumpIf struct {
	Cond      vendor.JumpTest
	Val       uint32
	SkipTrue  uint8
	SkipFalse uint8
}

// Assemble implements the Instruction Assemble method.
func (a JumpIf) Assemble() (RawInstruction, error) {
	return jumpToRaw(a.Cond, vendor.opOperandConstant, a.Val, a.SkipTrue, a.SkipFalse)
}

// String returns the instruction in assembler notation.
func (a JumpIf) String() string {
	return jumpToString(a.Cond, fmt.Sprintf("#%d", a.Val), a.SkipTrue, a.SkipFalse)
}

// JumpIfX skips the following Skip instructions in the program if A
// <Cond> X is true.
type JumpIfX struct {
	Cond      vendor.JumpTest
	SkipTrue  uint8
	SkipFalse uint8
}

// Assemble implements the Instruction Assemble method.
func (a JumpIfX) Assemble() (RawInstruction, error) {
	return jumpToRaw(a.Cond, vendor.opOperandX, 0, a.SkipTrue, a.SkipFalse)
}

// String returns the instruction in assembler notation.
func (a JumpIfX) String() string {
	return jumpToString(a.Cond, "x", a.SkipTrue, a.SkipFalse)
}

// jumpToRaw assembles a jump instruction into a RawInstruction
func jumpToRaw(test vendor.JumpTest, operand vendor.opOperand, k uint32, skipTrue, skipFalse uint8) (RawInstruction, error) {
	var (
		cond vendor.jumpOp
		flip bool
	)
	switch test {
	case vendor.JumpEqual:
		cond = vendor.opJumpEqual
	case vendor.JumpNotEqual:
		cond, flip = vendor.opJumpEqual, true
	case vendor.JumpGreaterThan:
		cond = vendor.opJumpGT
	case vendor.JumpLessThan:
		cond, flip = vendor.opJumpGE, true
	case vendor.JumpGreaterOrEqual:
		cond = vendor.opJumpGE
	case vendor.JumpLessOrEqual:
		cond, flip = vendor.opJumpGT, true
	case vendor.JumpBitsSet:
		cond = vendor.opJumpSet
	case vendor.JumpBitsNotSet:
		cond, flip = vendor.opJumpSet, true
	default:
		return RawInstruction{}, fmt.Errorf("unknown JumpTest %v", test)
	}
	jt, jf := skipTrue, skipFalse
	if flip {
		jt, jf = jf, jt
	}
	return RawInstruction{
		Op: vendor.opClsJump | uint16(cond) | uint16(operand),
		Jt: jt,
		Jf: jf,
		K:  k,
	}, nil
}

// jumpToString converts a jump instruction to assembler notation
func jumpToString(cond vendor.JumpTest, operand string, skipTrue, skipFalse uint8) string {
	switch cond {
	// K == A
	case vendor.JumpEqual:
		return conditionalJump(operand, skipTrue, skipFalse, "jeq", "jneq")
	// K != A
	case vendor.JumpNotEqual:
		return fmt.Sprintf("jneq %s,%d", operand, skipTrue)
	// K > A
	case vendor.JumpGreaterThan:
		return conditionalJump(operand, skipTrue, skipFalse, "jgt", "jle")
	// K < A
	case vendor.JumpLessThan:
		return fmt.Sprintf("jlt %s,%d", operand, skipTrue)
	// K >= A
	case vendor.JumpGreaterOrEqual:
		return conditionalJump(operand, skipTrue, skipFalse, "jge", "jlt")
	// K <= A
	case vendor.JumpLessOrEqual:
		return fmt.Sprintf("jle %s,%d", operand, skipTrue)
	// K & A != 0
	case vendor.JumpBitsSet:
		if skipFalse > 0 {
			return fmt.Sprintf("jset %s,%d,%d", operand, skipTrue, skipFalse)
		}
		return fmt.Sprintf("jset %s,%d", operand, skipTrue)
	// K & A == 0, there is no assembler instruction for JumpBitNotSet, use JumpBitSet and invert skips
	case vendor.JumpBitsNotSet:
		return jumpToString(vendor.JumpBitsSet, operand, skipFalse, skipTrue)
	default:
		return fmt.Sprintf("unknown JumpTest %#v", cond)
	}
}

func conditionalJump(operand string, skipTrue, skipFalse uint8, positiveJump, negativeJump string) string {
	if skipTrue > 0 {
		if skipFalse > 0 {
			return fmt.Sprintf("%s %s,%d,%d", positiveJump, operand, skipTrue, skipFalse)
		}
		return fmt.Sprintf("%s %s,%d", positiveJump, operand, skipTrue)
	}
	return fmt.Sprintf("%s %s,%d", negativeJump, operand, skipFalse)
}

// RetA exits the BPF program, returning the value of register A.
type RetA struct{}

// Assemble implements the Instruction Assemble method.
func (a RetA) Assemble() (RawInstruction, error) {
	return RawInstruction{
		Op: vendor.opClsReturn | vendor.opRetSrcA,
	}, nil
}

// String returns the instruction in assembler notation.
func (a RetA) String() string {
	return fmt.Sprintf("ret a")
}

// RetConstant exits the BPF program, returning a constant value.
type RetConstant struct {
	Val uint32
}

// Assemble implements the Instruction Assemble method.
func (a RetConstant) Assemble() (RawInstruction, error) {
	return RawInstruction{
		Op: vendor.opClsReturn | vendor.opRetSrcConstant,
		K:  a.Val,
	}, nil
}

// String returns the instruction in assembler notation.
func (a RetConstant) String() string {
	return fmt.Sprintf("ret #%d", a.Val)
}

// TXA copies the value of register X to register A.
type TXA struct{}

// Assemble implements the Instruction Assemble method.
func (a TXA) Assemble() (RawInstruction, error) {
	return RawInstruction{
		Op: vendor.opClsMisc | vendor.opMiscTXA,
	}, nil
}

// String returns the instruction in assembler notation.
func (a TXA) String() string {
	return fmt.Sprintf("txa")
}

// TAX copies the value of register A to register X.
type TAX struct{}

// Assemble implements the Instruction Assemble method.
func (a TAX) Assemble() (RawInstruction, error) {
	return RawInstruction{
		Op: vendor.opClsMisc | vendor.opMiscTAX,
	}, nil
}

// String returns the instruction in assembler notation.
func (a TAX) String() string {
	return fmt.Sprintf("tax")
}

func assembleLoad(dst vendor.Register, loadSize int, mode uint16, k uint32) (RawInstruction, error) {
	var (
		cls uint16
		sz  uint16
	)
	switch dst {
	case vendor.RegA:
		cls = vendor.opClsLoadA
	case vendor.RegX:
		cls = vendor.opClsLoadX
	default:
		return RawInstruction{}, fmt.Errorf("invalid target register %v", dst)
	}
	switch loadSize {
	case 1:
		sz = vendor.opLoadWidth1
	case 2:
		sz = vendor.opLoadWidth2
	case 4:
		sz = vendor.opLoadWidth4
	default:
		return RawInstruction{}, fmt.Errorf("invalid load byte length %d", sz)
	}
	return RawInstruction{
		Op: cls | sz | mode,
		K:  k,
	}, nil
}
