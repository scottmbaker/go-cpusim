package cpu4004

import (
	"fmt"
)

func (cpu *CPU4004) DebugInstr(format string, args ...interface{}) {
	if cpu.Sim.Debug {
		fmt.Printf(format+"\n", args...)
	}
}

func (cpu *CPU4004) DebugJump(isCall bool, addr uint16) {
	if isCall {
		cpu.DebugInstr("JMS %02xh", addr)
	} else {
		cpu.DebugInstr("JUN %02xh", addr)
	}
}

func (cpu *CPU4004) DebugJCN(invert, checkAccum, checkCarry, checkTest bool, addr uint16) {
	conditions := ""
	if checkAccum {
		conditions += "A"
	}
	if checkCarry {
		conditions += "C"
	}
	if checkTest {
		conditions += "T"
	}
	if invert {
		conditions += "N"
	}
	cpu.DebugInstr("JCN %s, %02xh", conditions, addr)
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
	case OP_TCS:
		cpu.DebugInstr("TCS")
	case OP_INC:
		cpu.DebugInstr("IAC")
	case OP_DEC:
		cpu.DebugInstr("DAC")
	case OP_DAA:
		cpu.DebugInstr("DAA")
	case OP_TCC:
		cpu.DebugInstr("TCC")
	case OP_KBP:
		cpu.DebugInstr("KBP")
	case OP_CMA:
		cpu.DebugInstr("CMA")
	case OP_ADM:
		cpu.DebugInstr("ADM")
	case OP_SBM:
		cpu.DebugInstr("SBM")
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
	} else {
		cpu.DebugInstr("??? Move Pair %s", srcName)
	}
}

func (cpu *CPU4004) DebugFIN(dest int) {
	destName := cpu.GetPairName(dest)
	cpu.DebugInstr("FIN %s", destName)
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

func (cpu *CPU4004) DebugRead(where int) {
	switch where {
	case 1:
		cpu.DebugInstr("RDM")
	case 2:
		cpu.DebugInstr("RDR")
	case 4:
		cpu.DebugInstr("RD0")
	case 5:
		cpu.DebugInstr("RD1")
	case 6:
		cpu.DebugInstr("RD2")
	case 7:
		cpu.DebugInstr("RD3")
	}
}

func (cpu *CPU4004) DebugWrite(where int) {
	switch where {
	case 0:
		cpu.DebugInstr("WRM")
	case 1:
		cpu.DebugInstr("WMP")
	case 2:
		cpu.DebugInstr("WRR")
	case 4:
		cpu.DebugInstr("WR0")
	case 5:
		cpu.DebugInstr("WR1")
	case 6:
		cpu.DebugInstr("WR2")
	case 7:
		cpu.DebugInstr("WR3")
	}
}
