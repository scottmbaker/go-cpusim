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
	if cpu.NewStyle {
		if isCall {
			if conditional {
				if isTrue {
					cpu.DebugInstr("C%s %04xh", FlagStr(flag), addr)
				} else {
					cpu.DebugInstr("CN%s %04xh", FlagStr(flag), addr)
				}
			} else {
				cpu.DebugInstr("CALL %04xh", addr)
			}
		} else {
			if conditional {
				if isTrue {
					cpu.DebugInstr("J%s %04xh", FlagStr(flag), addr)
				} else {
					cpu.DebugInstr("JN%s %04xh", FlagStr(flag), addr)
				}
			} else {
				cpu.DebugInstr("JMP %04xh", addr)
			}
		}
	} else {
		if isCall {
			if conditional {
				if isTrue {
					cpu.DebugInstr("CT%s %04xh", FlagStr(flag), addr)
				} else {
					cpu.DebugInstr("CF%s %04xh", FlagStr(flag), addr)
				}
			} else {
				cpu.DebugInstr("CAL %x", addr)
			}
		} else {
			if conditional {
				if isTrue {
					cpu.DebugInstr("JT%s %04xh", FlagStr(flag), addr)
				} else {
					cpu.DebugInstr("JF%s %04xh", FlagStr(flag), addr)
				}
			} else {
				cpu.DebugInstr("JMP %x", addr)
			}
		}
	}
}

func (cpu *CPU8008) DebugRet(conditional bool, flag int, isTrue bool) {
	if cpu.NewStyle {
		if conditional {
			if isTrue {
				cpu.DebugInstr("R%s", FlagStr(flag))
			} else {
				cpu.DebugInstr("RN%s", FlagStr(flag))
			}
		} else {
			cpu.DebugInstr("RET")
		}
	} else {
		if conditional {
			if isTrue {
				cpu.DebugInstr("RT%s", FlagStr(flag))
			} else {
				cpu.DebugInstr("RF%s", FlagStr(flag))
			}
		} else {
			cpu.DebugInstr("RET")
		}
	}
}

func (cpu *CPU8008) DebugRes(n byte) {
	cpu.DebugInstr("RST %x", n)
}

func (cpu *CPU8008) DebugPort(port byte, isRead bool, val byte) {
	if cpu.NewStyle {
		if isRead {
			cpu.DebugInstr("IN %02xh <- %02x", port, val)
		} else {
			cpu.DebugInstr("OUT %02xh -> %02x", port, val)
		}
	} else {
		if isRead {
			cpu.DebugInstr("INP %02xh <- %02x", port, val)
		} else {
			cpu.DebugInstr("OUT %02xh -> %02x", port, val)
		}
	}
}

func (cpu CPU8008) DebugRotate(op int) {
	switch op {
	case OP_RLC:
		cpu.DebugInstr("RLC A")
	case OP_RRC:
		cpu.DebugInstr("RRC A")
	case OP_RAL:
		cpu.DebugInstr("RAL A")
	case OP_RAR:
		cpu.DebugInstr("RAR A")
	default:
		cpu.DebugInstr("Unknown rotate operation: %d", op)
	}
}

// DebugAccumulator logs the accumulator operation
func (cpu CPU8008) DebugAccumulator(op int, reg int, isImmed bool, val byte) {
	if cpu.NewStyle {
		if isImmed {
			switch op {
			case OP_ADD:
				cpu.DebugInstr("ADI %02xh", val)
			case OP_AC:
				cpu.DebugInstr("ADI %02xh", val)
			case OP_SUB:
				cpu.DebugInstr("SUI %02xh", val)
			case OP_SC:
				cpu.DebugInstr("SCI %02xh", val)
			case OP_AND:
				cpu.DebugInstr("ANI %02xh", val)
			case OP_XOR:
				cpu.DebugInstr("XRI %02xh", val)
			case OP_OR:
				cpu.DebugInstr("ORI %02xh", val)
			case OP_CP:
				cpu.DebugInstr("CPI %02xh", val)
			}
		} else {
			regName := cpu.GetRegName(reg)
			switch op {
			case OP_ADD:
				cpu.DebugInstr("ADD %s", regName)
			case OP_AC:
				cpu.DebugInstr("ADC %s", regName)
			case OP_SUB:
				cpu.DebugInstr("SUB %s", regName)
			case OP_SC:
				cpu.DebugInstr("SBB %s", regName)
			case OP_AND:
				cpu.DebugInstr("ANA %s", regName)
			case OP_XOR:
				cpu.DebugInstr("XRA %s", regName)
			case OP_OR:
				cpu.DebugInstr("ORA %s", regName)
			case OP_CP:
				cpu.DebugInstr("CMP %s", regName)
			}
		}
	} else {
		if isImmed {
			switch op {
			case OP_ADD:
				cpu.DebugInstr("ADI %02xh", val)
			case OP_AC:
				cpu.DebugInstr("ACI %02xh", val)
			case OP_SUB:
				cpu.DebugInstr("SUI %02xh", val)
			case OP_SC:
				cpu.DebugInstr("SBI %02xh", val)
			case OP_AND:
				cpu.DebugInstr("NDI %02xh", val)
			case OP_XOR:
				cpu.DebugInstr("XRI %02xh", val)
			case OP_OR:
				cpu.DebugInstr("ORI %02xh", val)
			case OP_CP:
				cpu.DebugInstr("CPI %02xh", val)
			}
		} else {
			regName := cpu.GetRegName(reg)
			switch op {
			case OP_ADD:
				cpu.DebugInstr("AD%s", regName)
			case OP_AC:
				cpu.DebugInstr("AC%s", regName)
			case OP_SUB:
				cpu.DebugInstr("SU%s", regName)
			case OP_SC:
				cpu.DebugInstr("SB%s", regName)
			case OP_AND:
				cpu.DebugInstr("ND%s", regName)
			case OP_XOR:
				cpu.DebugInstr("XR%s", regName)
			case OP_OR:
				cpu.DebugInstr("OR%s", regName)
			case OP_CP:
				cpu.DebugInstr("CR%s", regName)
			}
		}
	}
}

func (cpu *CPU8008) DebugIncDec(reg int, increment int) {
	regName := cpu.GetRegName(reg)
	if cpu.NewStyle {
		if increment > 0 {
			cpu.DebugInstr("INR %s", regName)
		} else {
			cpu.DebugInstr("DCR %s", regName)
		}
	} else {
		if increment > 0 {
			cpu.DebugInstr("IN%s", regName)
		} else {
			cpu.DebugInstr("DC%s", regName)
		}
	}
}

func (cpu *CPU8008) DebugMove(dest, src int) {
	srcName := cpu.GetRegName(src)
	destName := cpu.GetRegName(dest)
	if cpu.NewStyle {
		cpu.DebugInstr("MOV %s,%s", destName, srcName)
	} else {
		cpu.DebugInstr("L%s%s", destName, srcName)
	}
}

func (cpu *CPU8008) DebugMoveImmediate(dest int, val byte) {
	destName := cpu.GetRegName(dest)
	if cpu.NewStyle {
		cpu.DebugInstr("MVI %s, %02xh", destName, val)
	} else {
		cpu.DebugInstr("L%sI %02xh", destName, val)
	}
}
