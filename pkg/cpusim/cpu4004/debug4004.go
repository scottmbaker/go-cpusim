package cpu4004

import (
	"fmt"
)

func (cpu *CPU4004) DebugInstr(format string, args ...interface{}) {
	if cpu.Sim.Debug {
		fmt.Printf(format+"\n", args...)
	}
}

func (cpu *CPU4004) DebugJump(conditional bool, flag int, isTrue bool, isCall bool, addr uint16) {
	_ = conditional
	_ = flag
	_ = isTrue
	_ = isCall
	_ = addr
}

func (cpu *CPU4004) DebugRet(value byte) {
	cpu.DebugInstr("BBL %02X", value)
}

// DebugAccumulator logs the accumulator operation
func (cpu CPU4004) DebugAccumulator(op int, reg int) {
	regName := cpu.GetRegName(reg)
	switch op {
	case OP_ADD:
		cpu.DebugInstr("ADD %s", regName)
	case OP_SUB:
		cpu.DebugInstr("SUB %s", regName)
	}
}

func (cpu *CPU4004) DebugInc(reg int) {
	regName := cpu.GetRegName(reg)
	cpu.DebugInstr("INC %s", regName)
}

func (cpu *CPU4004) DebugIncSkip(reg int, addr byte) {
	regName := cpu.GetRegName(reg)
	cpu.DebugInstr("ISZ %s, %02xh", regName, addr)
}

func (cpu *CPU4004) DebugMovePair(dest, src int) {
	srcName := cpu.GetPairName(src)
	if dest == PAIR_RC {
		cpu.DebugInstr("SRC %s", srcName)
	} else if dest == PAIR_P0 {
		cpu.DebugInstr("FIM %s", srcName)
	} else {
		cpu.DebugInstr("??? Move Pair %s", srcName)
	}
}

func (cpu *CPU4004) DebugMove(dest, src int) {
	srcName := cpu.GetRegName(src)
	_ = dest
	cpu.DebugInstr("LD %s", srcName)
}

func (cpu *CPU4004) DebugExchange(dest, src int) {
	srcName := cpu.GetRegName(src)
	_ = dest
	cpu.DebugInstr("XCH %s", srcName)
}

func (cpu *CPU4004) DebugMoveImmediate(dest int, val byte) {
	_ = dest
	cpu.DebugInstr("LDM %02xh", val)
}

func (cpu *CPU4004) DebugFetchImmediate(dest int, val byte) {
	destName := cpu.GetPairName(dest)
	cpu.DebugInstr("FIM %s, %02xh", destName, val)
}
