package cpuz80

import "github.com/scottmbaker/gocpusim/pkg/cpusim"

// executeUnprefixed handles all unprefixed Z80 opcodes
func (cpu *CPUZ80) executeUnprefixed(opcode byte) error {
	switch opcode {
	case 0x00: // NOP
		return nil

	case 0x01: // LD BC,nn
		cpu.setBC(cpu.fetchWord())
		return nil
	case 0x11: // LD DE,nn
		cpu.setDE(cpu.fetchWord())
		return nil
	case 0x21: // LD HL,nn
		cpu.setHL(cpu.fetchWord())
		return nil
	case 0x31: // LD SP,nn
		cpu.SP = cpu.fetchWord()
		return nil

	case 0x02: // LD (BC),A
		addr := cpu.getBC()
		cpu.writeByte(addr, cpu.A)
		cpu.WZ = (uint16(cpu.A) << 8) | ((addr + 1) & 0xFF)
		return nil
	case 0x12: // LD (DE),A
		addr := cpu.getDE()
		cpu.writeByte(addr, cpu.A)
		cpu.WZ = (uint16(cpu.A) << 8) | ((addr + 1) & 0xFF)
		return nil
	case 0x22: // LD (nn),HL
		addr := cpu.fetchWord()
		cpu.writeWord(addr, cpu.getHL())
		cpu.WZ = addr + 1
		return nil
	case 0x32: // LD (nn),A
		addr := cpu.fetchWord()
		cpu.writeByte(addr, cpu.A)
		cpu.WZ = (uint16(cpu.A) << 8) | ((addr + 1) & 0xFF)
		return nil

	case 0x03: // INC BC
		cpu.setBC(cpu.getBC() + 1)
		return nil
	case 0x13: // INC DE
		cpu.setDE(cpu.getDE() + 1)
		return nil
	case 0x23: // INC HL
		cpu.setHL(cpu.getHL() + 1)
		return nil
	case 0x33: // INC SP
		cpu.SP++
		return nil

	case 0x0B: // DEC BC
		cpu.setBC(cpu.getBC() - 1)
		return nil
	case 0x1B: // DEC DE
		cpu.setDE(cpu.getDE() - 1)
		return nil
	case 0x2B: // DEC HL
		cpu.setHL(cpu.getHL() - 1)
		return nil
	case 0x3B: // DEC SP
		cpu.SP--
		return nil

	case 0x04: // INC B
		cpu.B = cpu.inc8(cpu.B)
		return nil
	case 0x0C: // INC C
		cpu.C = cpu.inc8(cpu.C)
		return nil
	case 0x14: // INC D
		cpu.D = cpu.inc8(cpu.D)
		return nil
	case 0x1C: // INC E
		cpu.E = cpu.inc8(cpu.E)
		return nil
	case 0x24: // INC H
		cpu.H = cpu.inc8(cpu.H)
		return nil
	case 0x2C: // INC L
		cpu.L = cpu.inc8(cpu.L)
		return nil
	case 0x34: // INC (HL)
		addr := cpu.getHL()
		cpu.writeByte(addr, cpu.inc8(cpu.readByte(addr)))
		return nil
	case 0x3C: // INC A
		cpu.A = cpu.inc8(cpu.A)
		return nil

	case 0x05: // DEC B
		cpu.B = cpu.dec8(cpu.B)
		return nil
	case 0x0D: // DEC C
		cpu.C = cpu.dec8(cpu.C)
		return nil
	case 0x15: // DEC D
		cpu.D = cpu.dec8(cpu.D)
		return nil
	case 0x1D: // DEC E
		cpu.E = cpu.dec8(cpu.E)
		return nil
	case 0x25: // DEC H
		cpu.H = cpu.dec8(cpu.H)
		return nil
	case 0x2D: // DEC L
		cpu.L = cpu.dec8(cpu.L)
		return nil
	case 0x35: // DEC (HL)
		addr := cpu.getHL()
		cpu.writeByte(addr, cpu.dec8(cpu.readByte(addr)))
		return nil
	case 0x3D: // DEC A
		cpu.A = cpu.dec8(cpu.A)
		return nil

	case 0x06: // LD B,n
		cpu.B = cpu.fetchByte()
		return nil
	case 0x0E: // LD C,n
		cpu.C = cpu.fetchByte()
		return nil
	case 0x16: // LD D,n
		cpu.D = cpu.fetchByte()
		return nil
	case 0x1E: // LD E,n
		cpu.E = cpu.fetchByte()
		return nil
	case 0x26: // LD H,n
		cpu.H = cpu.fetchByte()
		return nil
	case 0x2E: // LD L,n
		cpu.L = cpu.fetchByte()
		return nil
	case 0x36: // LD (HL),n
		cpu.writeByte(cpu.getHL(), cpu.fetchByte())
		return nil
	case 0x3E: // LD A,n
		cpu.A = cpu.fetchByte()
		return nil

	case 0x07: // RLCA
		bit7 := cpu.A >> 7
		cpu.A = (cpu.A << 1) | bit7
		f := cpu.F & (MaskS | MaskZ | MaskPV)
		f |= cpu.A & (MaskY | MaskX)
		if bit7 != 0 {
			f |= MaskC
		}
		cpu.F = f
		cpu.Q = f
		return nil

	case 0x0F: // RRCA
		bit0 := cpu.A & 1
		cpu.A = (cpu.A >> 1) | (bit0 << 7)
		f := cpu.F & (MaskS | MaskZ | MaskPV)
		f |= cpu.A & (MaskY | MaskX)
		if bit0 != 0 {
			f |= MaskC
		}
		cpu.F = f
		cpu.Q = f
		return nil

	case 0x17: // RLA
		oldCarry := cpu.F & MaskC
		bit7 := cpu.A >> 7
		cpu.A = (cpu.A << 1) | oldCarry
		f := cpu.F & (MaskS | MaskZ | MaskPV)
		f |= cpu.A & (MaskY | MaskX)
		if bit7 != 0 {
			f |= MaskC
		}
		cpu.F = f
		cpu.Q = f
		return nil

	case 0x1F: // RRA
		oldCarry := cpu.F & MaskC
		bit0 := cpu.A & 1
		cpu.A = (cpu.A >> 1) | (oldCarry << 7)
		f := cpu.F & (MaskS | MaskZ | MaskPV)
		f |= cpu.A & (MaskY | MaskX)
		if bit0 != 0 {
			f |= MaskC
		}
		cpu.F = f
		cpu.Q = f
		return nil

	case 0x08: // EX AF,AF'
		af := cpu.getAF()
		cpu.setAF(cpu.AF_)
		cpu.AF_ = af
		return nil

	case 0x09: // ADD HL,BC
		cpu.addHL(cpu.getBC())
		return nil
	case 0x19: // ADD HL,DE
		cpu.addHL(cpu.getDE())
		return nil
	case 0x29: // ADD HL,HL
		cpu.addHL(cpu.getHL())
		return nil
	case 0x39: // ADD HL,SP
		cpu.addHL(cpu.SP)
		return nil

	case 0x0A: // LD A,(BC)
		addr := cpu.getBC()
		cpu.A = cpu.readByte(addr)
		cpu.WZ = addr + 1
		return nil
	case 0x1A: // LD A,(DE)
		addr := cpu.getDE()
		cpu.A = cpu.readByte(addr)
		cpu.WZ = addr + 1
		return nil
	case 0x2A: // LD HL,(nn)
		addr := cpu.fetchWord()
		cpu.setHL(cpu.readWord(addr))
		cpu.WZ = addr + 1
		return nil
	case 0x3A: // LD A,(nn)
		addr := cpu.fetchWord()
		cpu.A = cpu.readByte(addr)
		cpu.WZ = addr + 1
		return nil

	case 0x10: // DJNZ d
		d := cpu.fetchByte()
		cpu.B--
		if cpu.B != 0 {
			offset := int8(d)
			cpu.PC = uint16(int32(cpu.PC) + int32(offset))
			cpu.WZ = cpu.PC
		}
		return nil

	case 0x18: // JR d
		d := cpu.fetchByte()
		offset := int8(d)
		cpu.PC = uint16(int32(cpu.PC) + int32(offset))
		cpu.WZ = cpu.PC
		return nil

	case 0x20: // JR NZ,d
		d := cpu.fetchByte()
		if cpu.F&MaskZ == 0 {
			offset := int8(d)
			cpu.PC = uint16(int32(cpu.PC) + int32(offset))
			cpu.WZ = cpu.PC
		}
		return nil

	case 0x28: // JR Z,d
		d := cpu.fetchByte()
		if cpu.F&MaskZ != 0 {
			offset := int8(d)
			cpu.PC = uint16(int32(cpu.PC) + int32(offset))
			cpu.WZ = cpu.PC
		}
		return nil

	case 0x30: // JR NC,d
		d := cpu.fetchByte()
		if cpu.F&MaskC == 0 {
			offset := int8(d)
			cpu.PC = uint16(int32(cpu.PC) + int32(offset))
			cpu.WZ = cpu.PC
		}
		return nil

	case 0x38: // JR C,d
		d := cpu.fetchByte()
		if cpu.F&MaskC != 0 {
			offset := int8(d)
			cpu.PC = uint16(int32(cpu.PC) + int32(offset))
			cpu.WZ = cpu.PC
		}
		return nil

	case 0x27: // DAA
		cpu.daa()
		return nil

	case 0x2F: // CPL
		cpu.A = ^cpu.A
		f := cpu.F & (MaskS | MaskZ | MaskPV | MaskC)
		f |= MaskH | MaskN
		f |= cpu.A & (MaskY | MaskX)
		cpu.F = f
		cpu.Q = f
		return nil

	case 0x37: // SCF
		f := cpu.F & (MaskS | MaskZ | MaskPV)
		f |= MaskC
		f |= (cpu.A | (cpu.PrevQ ^ cpu.F)) & (MaskY | MaskX)
		cpu.F = f
		cpu.Q = f
		return nil

	case 0x3F: // CCF
		f := cpu.F & (MaskS | MaskZ | MaskPV)
		if cpu.F&MaskC != 0 {
			f |= MaskH
		}
		f |= (cpu.F ^ MaskC) & MaskC
		f |= (cpu.A | (cpu.PrevQ ^ cpu.F)) & (MaskY | MaskX)
		cpu.F = f
		cpu.Q = f
		return nil

	// LD r,r' block: 0x40-0x7F (except 0x76 which is HALT)
	case 0x76: // HALT
		cpu.Halted.Store(true)
		return nil

	case 0x40: // LD B,B
		return nil
	case 0x41: // LD B,C
		cpu.B = cpu.C
		return nil
	case 0x42: // LD B,D
		cpu.B = cpu.D
		return nil
	case 0x43: // LD B,E
		cpu.B = cpu.E
		return nil
	case 0x44: // LD B,H
		cpu.B = cpu.H
		return nil
	case 0x45: // LD B,L
		cpu.B = cpu.L
		return nil
	case 0x46: // LD B,(HL)
		cpu.B = cpu.readByte(cpu.getHL())
		return nil
	case 0x47: // LD B,A
		cpu.B = cpu.A
		return nil

	case 0x48: // LD C,B
		cpu.C = cpu.B
		return nil
	case 0x49: // LD C,C
		return nil
	case 0x4A: // LD C,D
		cpu.C = cpu.D
		return nil
	case 0x4B: // LD C,E
		cpu.C = cpu.E
		return nil
	case 0x4C: // LD C,H
		cpu.C = cpu.H
		return nil
	case 0x4D: // LD C,L
		cpu.C = cpu.L
		return nil
	case 0x4E: // LD C,(HL)
		cpu.C = cpu.readByte(cpu.getHL())
		return nil
	case 0x4F: // LD C,A
		cpu.C = cpu.A
		return nil

	case 0x50: // LD D,B
		cpu.D = cpu.B
		return nil
	case 0x51: // LD D,C
		cpu.D = cpu.C
		return nil
	case 0x52: // LD D,D
		return nil
	case 0x53: // LD D,E
		cpu.D = cpu.E
		return nil
	case 0x54: // LD D,H
		cpu.D = cpu.H
		return nil
	case 0x55: // LD D,L
		cpu.D = cpu.L
		return nil
	case 0x56: // LD D,(HL)
		cpu.D = cpu.readByte(cpu.getHL())
		return nil
	case 0x57: // LD D,A
		cpu.D = cpu.A
		return nil

	case 0x58: // LD E,B
		cpu.E = cpu.B
		return nil
	case 0x59: // LD E,C
		cpu.E = cpu.C
		return nil
	case 0x5A: // LD E,D
		cpu.E = cpu.D
		return nil
	case 0x5B: // LD E,E
		return nil
	case 0x5C: // LD E,H
		cpu.E = cpu.H
		return nil
	case 0x5D: // LD E,L
		cpu.E = cpu.L
		return nil
	case 0x5E: // LD E,(HL)
		cpu.E = cpu.readByte(cpu.getHL())
		return nil
	case 0x5F: // LD E,A
		cpu.E = cpu.A
		return nil

	case 0x60: // LD H,B
		cpu.H = cpu.B
		return nil
	case 0x61: // LD H,C
		cpu.H = cpu.C
		return nil
	case 0x62: // LD H,D
		cpu.H = cpu.D
		return nil
	case 0x63: // LD H,E
		cpu.H = cpu.E
		return nil
	case 0x64: // LD H,H
		return nil
	case 0x65: // LD H,L
		cpu.H = cpu.L
		return nil
	case 0x66: // LD H,(HL)
		cpu.H = cpu.readByte(cpu.getHL())
		return nil
	case 0x67: // LD H,A
		cpu.H = cpu.A
		return nil

	case 0x68: // LD L,B
		cpu.L = cpu.B
		return nil
	case 0x69: // LD L,C
		cpu.L = cpu.C
		return nil
	case 0x6A: // LD L,D
		cpu.L = cpu.D
		return nil
	case 0x6B: // LD L,E
		cpu.L = cpu.E
		return nil
	case 0x6C: // LD L,H
		cpu.L = cpu.H
		return nil
	case 0x6D: // LD L,L
		return nil
	case 0x6E: // LD L,(HL)
		cpu.L = cpu.readByte(cpu.getHL())
		return nil
	case 0x6F: // LD L,A
		cpu.L = cpu.A
		return nil

	case 0x70: // LD (HL),B
		cpu.writeByte(cpu.getHL(), cpu.B)
		return nil
	case 0x71: // LD (HL),C
		cpu.writeByte(cpu.getHL(), cpu.C)
		return nil
	case 0x72: // LD (HL),D
		cpu.writeByte(cpu.getHL(), cpu.D)
		return nil
	case 0x73: // LD (HL),E
		cpu.writeByte(cpu.getHL(), cpu.E)
		return nil
	case 0x74: // LD (HL),H
		cpu.writeByte(cpu.getHL(), cpu.H)
		return nil
	case 0x75: // LD (HL),L
		cpu.writeByte(cpu.getHL(), cpu.L)
		return nil
	// 0x76 is HALT (handled above)
	case 0x77: // LD (HL),A
		cpu.writeByte(cpu.getHL(), cpu.A)
		return nil

	case 0x78: // LD A,B
		cpu.A = cpu.B
		return nil
	case 0x79: // LD A,C
		cpu.A = cpu.C
		return nil
	case 0x7A: // LD A,D
		cpu.A = cpu.D
		return nil
	case 0x7B: // LD A,E
		cpu.A = cpu.E
		return nil
	case 0x7C: // LD A,H
		cpu.A = cpu.H
		return nil
	case 0x7D: // LD A,L
		cpu.A = cpu.L
		return nil
	case 0x7E: // LD A,(HL)
		cpu.A = cpu.readByte(cpu.getHL())
		return nil
	case 0x7F: // LD A,A
		return nil

	// ALU A,r: 0x80-0xBF
	case 0x80, 0x81, 0x82, 0x83, 0x84, 0x85, 0x86, 0x87:
		cpu.add8(cpu.getReg8(opcode & 0x07))
		return nil
	case 0x88, 0x89, 0x8A, 0x8B, 0x8C, 0x8D, 0x8E, 0x8F:
		cpu.adc8(cpu.getReg8(opcode & 0x07))
		return nil
	case 0x90, 0x91, 0x92, 0x93, 0x94, 0x95, 0x96, 0x97:
		cpu.sub8(cpu.getReg8(opcode & 0x07))
		return nil
	case 0x98, 0x99, 0x9A, 0x9B, 0x9C, 0x9D, 0x9E, 0x9F:
		cpu.sbc8(cpu.getReg8(opcode & 0x07))
		return nil
	case 0xA0, 0xA1, 0xA2, 0xA3, 0xA4, 0xA5, 0xA6, 0xA7:
		cpu.and8(cpu.getReg8(opcode & 0x07))
		return nil
	case 0xA8, 0xA9, 0xAA, 0xAB, 0xAC, 0xAD, 0xAE, 0xAF:
		cpu.xor8(cpu.getReg8(opcode & 0x07))
		return nil
	case 0xB0, 0xB1, 0xB2, 0xB3, 0xB4, 0xB5, 0xB6, 0xB7:
		cpu.or8(cpu.getReg8(opcode & 0x07))
		return nil
	case 0xB8, 0xB9, 0xBA, 0xBB, 0xBC, 0xBD, 0xBE, 0xBF:
		cpu.cp8(cpu.getReg8(opcode & 0x07))
		return nil

	// ALU A,n immediate: 0xC6, 0xCE, 0xD6, 0xDE, 0xE6, 0xEE, 0xF6, 0xFE
	case 0xC6: // ADD A,n
		cpu.add8(cpu.fetchByte())
		return nil
	case 0xCE: // ADC A,n
		cpu.adc8(cpu.fetchByte())
		return nil
	case 0xD6: // SUB n
		cpu.sub8(cpu.fetchByte())
		return nil
	case 0xDE: // SBC A,n
		cpu.sbc8(cpu.fetchByte())
		return nil
	case 0xE6: // AND n
		cpu.and8(cpu.fetchByte())
		return nil
	case 0xEE: // XOR n
		cpu.xor8(cpu.fetchByte())
		return nil
	case 0xF6: // OR n
		cpu.or8(cpu.fetchByte())
		return nil
	case 0xFE: // CP n
		cpu.cp8(cpu.fetchByte())
		return nil

	// JP nn
	case 0xC3:
		addr := cpu.fetchWord()
		cpu.PC = addr
		cpu.WZ = addr
		return nil

	// JP cc,nn
	case 0xC2, 0xCA, 0xD2, 0xDA, 0xE2, 0xEA, 0xF2, 0xFA:
		addr := cpu.fetchWord()
		cc := (opcode >> 3) & 0x07
		if cpu.condition(cc) {
			cpu.PC = addr
		}
		cpu.WZ = addr
		return nil

	// CALL nn
	case 0xCD:
		addr := cpu.fetchWord()
		cpu.push(cpu.PC)
		cpu.PC = addr
		cpu.WZ = addr
		return nil

	// CALL cc,nn
	case 0xC4, 0xCC, 0xD4, 0xDC, 0xE4, 0xEC, 0xF4, 0xFC:
		addr := cpu.fetchWord()
		cc := (opcode >> 3) & 0x07
		if cpu.condition(cc) {
			cpu.push(cpu.PC)
			cpu.PC = addr
		}
		cpu.WZ = addr
		return nil

	// RET
	case 0xC9:
		cpu.PC = cpu.pop()
		cpu.WZ = cpu.PC
		return nil

	// RET cc
	case 0xC0, 0xC8, 0xD0, 0xD8, 0xE0, 0xE8, 0xF0, 0xF8:
		cc := (opcode >> 3) & 0x07
		if cpu.condition(cc) {
			cpu.PC = cpu.pop()
			cpu.WZ = cpu.PC
		}
		return nil

	// RST
	case 0xC7, 0xCF, 0xD7, 0xDF, 0xE7, 0xEF, 0xF7, 0xFF:
		addr := uint16(opcode & 0x38)
		cpu.push(cpu.PC)
		cpu.PC = addr
		cpu.WZ = addr
		return nil

	// PUSH/POP
	case 0xC5: // PUSH BC
		cpu.push(cpu.getBC())
		return nil
	case 0xD5: // PUSH DE
		cpu.push(cpu.getDE())
		return nil
	case 0xE5: // PUSH HL
		cpu.push(cpu.getHL())
		return nil
	case 0xF5: // PUSH AF
		cpu.push(cpu.getAF())
		return nil
	case 0xC1: // POP BC
		cpu.setBC(cpu.pop())
		return nil
	case 0xD1: // POP DE
		cpu.setDE(cpu.pop())
		return nil
	case 0xE1: // POP HL
		cpu.setHL(cpu.pop())
		return nil
	case 0xF1: // POP AF
		cpu.setAF(cpu.pop())
		return nil

	// Exchange instructions
	case 0xD9: // EXX
		bc := cpu.getBC()
		de := cpu.getDE()
		hl := cpu.getHL()
		cpu.setBC(cpu.BC_)
		cpu.setDE(cpu.DE_)
		cpu.setHL(cpu.HL_)
		cpu.BC_ = bc
		cpu.DE_ = de
		cpu.HL_ = hl
		return nil

	case 0xEB: // EX DE,HL
		de := cpu.getDE()
		cpu.setDE(cpu.getHL())
		cpu.setHL(de)
		return nil

	case 0xE3: // EX (SP),HL
		lo := cpu.readByte(cpu.SP)
		hi := cpu.readByte(cpu.SP + 1)
		cpu.writeByte(cpu.SP, cpu.L)
		cpu.writeByte(cpu.SP+1, cpu.H)
		cpu.L = lo
		cpu.H = hi
		cpu.WZ = cpu.getHL()
		return nil

	// I/O
	case 0xDB: // IN A,(n)
		port := cpu.fetchByte()
		addr := (uint16(cpu.A)<<8 | uint16(port)) & cpu.PortAddressMask
		cpu.A = cpu.readPort(addr)
		cpu.WZ = addr + 1
		return nil

	case 0xD3: // OUT (n),A
		port := cpu.fetchByte()
		addr := (uint16(cpu.A)<<8 | uint16(port)) & cpu.PortAddressMask
		cpu.writePort(addr, cpu.A)
		cpu.WZ = (uint16(cpu.A) << 8) | ((uint16(port) + 1) & 0xFF)
		return nil

	// Misc
	case 0xF3: // DI
		cpu.IFF1 = false
		cpu.IFF2 = false
		return nil

	case 0xFB: // EI
		cpu.IFF1 = true
		cpu.IFF2 = true
		cpu.EIPending = true // Interrupts delayed until after next instruction
		return nil

	case 0xF9: // LD SP,HL
		cpu.SP = cpu.getHL()
		return nil

	case 0xE9: // JP (HL) - actually JP HL, not indirect
		cpu.PC = cpu.getHL()
		return nil

	// Prefix bytes - dispatch to sub-handlers
	case 0xCB:
		return cpu.executeCB()
	case 0xDD:
		return cpu.executeDD()
	case 0xED:
		return cpu.executeED()
	case 0xFD:
		return cpu.executeFD()
	}

	// Should not reach here
	return &cpusim.ErrInvalidOpcode{Device: cpu, Opcode: opcode}
}
