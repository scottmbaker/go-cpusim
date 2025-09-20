package cpu8008

import (
	"fmt"
)

func (cpu *CPU8008) DebugInstr(format string, args ...interface{}) {
	if cpu.Sim.Debug {
		fmt.Printf(format+"\n", args...)
	}
}

func (cpu *CPU8008) DebugJump(conditional bool, flag int, isTrue bool, isCall bool, addr uint16) {
	_ = contidional
	_ = flag
	_ = isTrue
	_ = isCall
	_ = addr
}

func (cpu *CPU8008) DebugRet(value byte) {
	cpu.DebugInstr("BBL %02X", value)
}

// DebugAccumulator logs the accumulator operation
func (cpu CPU8008) DebugAccumulator(op int, reg int) {
	regName := cpu.GetRegName(reg)
	switch op {
	case OP_ADD:
		cpu.DebugInstr("ADD %s", regName)
	case OP_SUB:
		cpu.DebugInstr("SUB %s", regName)
	}
}

func (cpu *CPU8008) DebugIncDec(reg int, increment int) {
	regName := cpu.GetRegName(reg)
	if increment > 0 {
		cpu.DebugInstr("INC %s", regName)
	} else {
		cpu.DebugInstr("DCR %s", regName)
	}
}

func (cpu *CPU8008) DebugMove(dest, src int) {
	srcName := cpu.GetRegName(src)
	_ = dest
	cpu.DebugInstr("LD %s", srcName)
}

func (cpu *CPU8008) DebugExchange(dest, src int) {
	srcName := cpu.GetRegName(src)
	_ = dest
	cpu.DebugInstr("XCH %s", srcName)
}

func (cpu *CPU8008) DebugMoveImmediate(dest int, val byte) {
	_ = dest
	cpu.DebugInstr("LDM %02xh", val)
}

func (cpu *CPU8008) DebugFetchImmediate(dest int, val byte) {
	destName := cpu.GetPairName(dest)
	cpu.DebugInstr("FIM %s, %02xh", val)
}
