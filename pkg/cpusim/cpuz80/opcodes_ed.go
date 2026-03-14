package cpuz80

// executeED handles ED-prefixed opcodes (extended instructions)
func (cpu *CPUZ80) executeED() error {
	cpu.incR()
	opcode := cpu.fetchByte()

	switch opcode {
	// IN r,(C) - 0x40,0x48,0x50,0x58,0x60,0x68,0x70,0x78
	case 0x40: // IN B,(C)
		cpu.B = cpu.edIn()
		return nil
	case 0x48: // IN C,(C)
		cpu.C = cpu.edIn()
		return nil
	case 0x50: // IN D,(C)
		cpu.D = cpu.edIn()
		return nil
	case 0x58: // IN E,(C)
		cpu.E = cpu.edIn()
		return nil
	case 0x60: // IN H,(C)
		cpu.H = cpu.edIn()
		return nil
	case 0x68: // IN L,(C)
		cpu.L = cpu.edIn()
		return nil
	case 0x70: // IN (C) / IN F,(C) - result discarded, flags set
		cpu.edIn()
		return nil
	case 0x78: // IN A,(C)
		cpu.A = cpu.edIn()
		return nil

	// OUT (C),r - 0x41,0x49,0x51,0x59,0x61,0x69,0x71,0x79
	case 0x41:
		cpu.edOut(cpu.B)
		return nil
	case 0x49:
		cpu.edOut(cpu.C)
		return nil
	case 0x51:
		cpu.edOut(cpu.D)
		return nil
	case 0x59:
		cpu.edOut(cpu.E)
		return nil
	case 0x61:
		cpu.edOut(cpu.H)
		return nil
	case 0x69:
		cpu.edOut(cpu.L)
		return nil
	case 0x71: // OUT (C),0
		cpu.edOut(0)
		return nil
	case 0x79:
		cpu.edOut(cpu.A)
		return nil

	// SBC HL,rr
	case 0x42:
		cpu.sbcHL(cpu.getBC())
		return nil
	case 0x52:
		cpu.sbcHL(cpu.getDE())
		return nil
	case 0x62:
		cpu.sbcHL(cpu.getHL())
		return nil
	case 0x72:
		cpu.sbcHL(cpu.SP)
		return nil

	// ADC HL,rr
	case 0x4A:
		cpu.adcHL(cpu.getBC())
		return nil
	case 0x5A:
		cpu.adcHL(cpu.getDE())
		return nil
	case 0x6A:
		cpu.adcHL(cpu.getHL())
		return nil
	case 0x7A:
		cpu.adcHL(cpu.SP)
		return nil

	// LD (nn),rr
	case 0x43:
		addr := cpu.fetchWord()
		cpu.writeWord(addr, cpu.getBC())
		cpu.WZ = addr + 1
		return nil
	case 0x53:
		addr := cpu.fetchWord()
		cpu.writeWord(addr, cpu.getDE())
		cpu.WZ = addr + 1
		return nil
	case 0x63:
		addr := cpu.fetchWord()
		cpu.writeWord(addr, cpu.getHL())
		cpu.WZ = addr + 1
		return nil
	case 0x73:
		addr := cpu.fetchWord()
		cpu.writeWord(addr, cpu.SP)
		cpu.WZ = addr + 1
		return nil

	// LD rr,(nn)
	case 0x4B:
		addr := cpu.fetchWord()
		cpu.setBC(cpu.readWord(addr))
		cpu.WZ = addr + 1
		return nil
	case 0x5B:
		addr := cpu.fetchWord()
		cpu.setDE(cpu.readWord(addr))
		cpu.WZ = addr + 1
		return nil
	case 0x6B:
		addr := cpu.fetchWord()
		cpu.setHL(cpu.readWord(addr))
		cpu.WZ = addr + 1
		return nil
	case 0x7B:
		addr := cpu.fetchWord()
		cpu.SP = cpu.readWord(addr)
		cpu.WZ = addr + 1
		return nil

	// NEG
	case 0x44, 0x4C, 0x54, 0x5C, 0x64, 0x6C, 0x74, 0x7C:
		a := cpu.A
		cpu.A = 0
		cpu.sub8(a)
		return nil

	// RETN
	case 0x45, 0x55, 0x65, 0x75:
		cpu.IFF1 = cpu.IFF2
		cpu.PC = cpu.pop()
		cpu.WZ = cpu.PC
		return nil

	// RETI
	case 0x4D, 0x5D, 0x6D, 0x7D:
		cpu.IFF1 = cpu.IFF2
		cpu.PC = cpu.pop()
		cpu.WZ = cpu.PC
		return nil

	// IM 0
	case 0x46, 0x66:
		cpu.IM = 0
		return nil
	// IM 1
	case 0x56, 0x76:
		cpu.IM = 1
		return nil
	// IM 2
	case 0x5E, 0x7E:
		cpu.IM = 2
		return nil
	// IM 0/1 (undocumented, acts as IM 0)
	case 0x4E, 0x6E:
		cpu.IM = 0
		return nil

	// LD I,A
	case 0x47:
		cpu.I = cpu.A
		return nil

	// LD R,A
	case 0x4F:
		cpu.R = cpu.A
		return nil

	// LD A,I
	case 0x57:
		cpu.A = cpu.I
		f := cpu.F & MaskC
		f |= cpu.A & (MaskS | MaskY | MaskX)
		if cpu.A == 0 {
			f |= MaskZ
		}
		if cpu.IFF2 {
			f |= MaskPV
		}
		cpu.F = f
		cpu.Q = f
		return nil

	// LD A,R
	case 0x5F:
		cpu.A = cpu.R
		f := cpu.F & MaskC
		f |= cpu.A & (MaskS | MaskY | MaskX)
		if cpu.A == 0 {
			f |= MaskZ
		}
		if cpu.IFF2 {
			f |= MaskPV
		}
		cpu.F = f
		cpu.Q = f
		return nil

	// RRD
	case 0x67:
		addr := cpu.getHL()
		val := cpu.readByte(addr)
		result := (cpu.A << 4) | (val >> 4)
		cpu.A = (cpu.A & 0xF0) | (val & 0x0F)
		cpu.writeByte(addr, result)
		f := cpu.F & MaskC
		f |= cpu.A & (MaskS | MaskY | MaskX)
		if cpu.A == 0 {
			f |= MaskZ
		}
		if parityTable[cpu.A] {
			f |= MaskPV
		}
		cpu.F = f
		cpu.Q = f
		cpu.WZ = addr + 1
		return nil

	// RLD
	case 0x6F:
		addr := cpu.getHL()
		val := cpu.readByte(addr)
		result := (val << 4) | (cpu.A & 0x0F)
		cpu.A = (cpu.A & 0xF0) | (val >> 4)
		cpu.writeByte(addr, result)
		f := cpu.F & MaskC
		f |= cpu.A & (MaskS | MaskY | MaskX)
		if cpu.A == 0 {
			f |= MaskZ
		}
		if parityTable[cpu.A] {
			f |= MaskPV
		}
		cpu.F = f
		cpu.Q = f
		cpu.WZ = addr + 1
		return nil

	// Block transfer instructions
	case 0xA0: // LDI
		cpu.ldi()
		return nil
	case 0xB0: // LDIR
		cpu.ldi()
		if cpu.getBC() != 0 {
			cpu.PC -= 2
			cpu.WZ = cpu.PC + 1
			cpu.blockRepeatFlags()
		}
		return nil
	case 0xA8: // LDD
		cpu.ldd()
		return nil
	case 0xB8: // LDDR
		cpu.ldd()
		if cpu.getBC() != 0 {
			cpu.PC -= 2
			cpu.WZ = cpu.PC + 1
			cpu.blockRepeatFlags()
		}
		return nil

	// Block compare instructions
	case 0xA1: // CPI
		cpu.cpi()
		return nil
	case 0xB1: // CPIR
		cpu.cpi()
		if cpu.getBC() != 0 && cpu.F&MaskZ == 0 {
			cpu.PC -= 2
			cpu.WZ = cpu.PC + 1
			cpu.blockRepeatFlags()
		}
		return nil
	case 0xA9: // CPD
		cpu.cpd()
		return nil
	case 0xB9: // CPDR
		cpu.cpd()
		if cpu.getBC() != 0 && cpu.F&MaskZ == 0 {
			cpu.PC -= 2
			cpu.WZ = cpu.PC + 1
			cpu.blockRepeatFlags()
		}
		return nil

	// Block I/O instructions
	case 0xA2: // INI
		cpu.ini()
		return nil
	case 0xB2: // INIR
		cpu.ini()
		if cpu.B != 0 {
			cpu.PC -= 2
			cpu.WZ = cpu.PC + 1
			cpu.blockRepeatIOFlags()
		}
		return nil
	case 0xAA: // IND
		cpu.ind()
		return nil
	case 0xBA: // INDR
		cpu.ind()
		if cpu.B != 0 {
			cpu.PC -= 2
			cpu.WZ = cpu.PC + 1
			cpu.blockRepeatIOFlags()
		}
		return nil
	case 0xA3: // OUTI
		cpu.outi()
		return nil
	case 0xB3: // OTIR
		cpu.outi()
		if cpu.B != 0 {
			cpu.PC -= 2
			cpu.WZ = cpu.PC + 1
			cpu.blockRepeatIOFlags()
		}
		return nil
	case 0xAB: // OUTD
		cpu.outd()
		return nil
	case 0xBB: // OTDR
		cpu.outd()
		if cpu.B != 0 {
			cpu.PC -= 2
			cpu.WZ = cpu.PC + 1
			cpu.blockRepeatIOFlags()
		}
		return nil

	default:
		// Invalid ED opcodes act as 2-byte NOP
		return nil
	}
}

func (cpu *CPUZ80) blockRepeatFlags() {
	pch := byte(cpu.PC >> 8)
	cpu.F = (cpu.F & ^byte(MaskY|MaskX)) | (pch & (MaskY | MaskX))
	cpu.Q = cpu.F
}

func (cpu *CPUZ80) blockRepeatIOFlags() {
	f := cpu.F

	pch := byte(cpu.PC >> 8)
	f = (f & ^byte(MaskY|MaskX)) | (pch & (MaskY | MaskX))

	// PV and H adjustments depend on CF and N (N holds bit 7 of data value)
	if f&MaskC != 0 {
		if f&MaskN != 0 {
			p := byte(0)
			if parityTable[(cpu.B-1)&0x07] {
				p = 1
			}
			f ^= (p ^ 1) << 2
			if cpu.B&0x0F == 0x00 {
				f |= MaskH
			} else {
				f &^= MaskH
			}
		} else {
			p := byte(0)
			if parityTable[(cpu.B+1)&0x07] {
				p = 1
			}
			f ^= (p ^ 1) << 2
			if cpu.B&0x0F == 0x0F {
				f |= MaskH
			} else {
				f &^= MaskH
			}
		}
	} else {
		p := byte(0)
		if parityTable[cpu.B&0x07] {
			p = 1
		}
		f ^= (p ^ 1) << 2
	}

	cpu.F = f
	cpu.Q = f
}

func (cpu *CPUZ80) edIn() byte {
	addr := cpu.getBC() & cpu.PortAddressMask
	val := cpu.readPort(addr)
	f := cpu.F & MaskC
	f |= val & (MaskS | MaskY | MaskX)
	if val == 0 {
		f |= MaskZ
	}
	if parityTable[val] {
		f |= MaskPV
	}
	cpu.F = f
	cpu.Q = f
	cpu.WZ = addr + 1
	return val
}

func (cpu *CPUZ80) edOut(val byte) {
	addr := cpu.getBC() & cpu.PortAddressMask
	cpu.writePort(addr, val)
	cpu.WZ = addr + 1
}

func (cpu *CPUZ80) adcHL(val uint16) {
	hl := cpu.getHL()
	carry := uint32(cpu.F & MaskC)
	result := uint32(hl) + uint32(val) + carry

	var f byte
	r16 := uint16(result)
	f |= byte(r16>>8) & (MaskS | MaskY | MaskX)
	if r16 == 0 {
		f |= MaskZ
	}
	if result > 0xFFFF {
		f |= MaskC
	}
	if (hl&0x0FFF)+(val&0x0FFF)+uint16(carry) > 0x0FFF {
		f |= MaskH
	}
	if (hl^val)&0x8000 == 0 && (hl^r16)&0x8000 != 0 {
		f |= MaskPV
	}

	cpu.setHL(r16)
	cpu.F = f
	cpu.Q = f
	cpu.WZ = hl + 1
}

func (cpu *CPUZ80) sbcHL(val uint16) {
	hl := cpu.getHL()
	carry := uint32(cpu.F & MaskC)
	result := uint32(hl) - uint32(val) - carry

	var f byte
	f |= MaskN
	r16 := uint16(result)
	f |= byte(r16>>8) & (MaskS | MaskY | MaskX)
	if r16 == 0 {
		f |= MaskZ
	}
	if result > 0xFFFF {
		f |= MaskC
	}
	if int32(hl&0x0FFF)-int32(val&0x0FFF)-int32(carry) < 0 {
		f |= MaskH
	}
	if (hl^val)&0x8000 != 0 && (hl^r16)&0x8000 != 0 {
		f |= MaskPV
	}

	cpu.setHL(r16)
	cpu.F = f
	cpu.Q = f
	cpu.WZ = hl + 1
}

func (cpu *CPUZ80) ldi() {
	val := cpu.readByte(cpu.getHL())
	cpu.writeByte(cpu.getDE(), val)
	cpu.setDE(cpu.getDE() + 1)
	cpu.setHL(cpu.getHL() + 1)
	cpu.setBC(cpu.getBC() - 1)

	n := val + cpu.A
	f := cpu.F & (MaskS | MaskZ | MaskC)
	if n&0x02 != 0 {
		f |= MaskY
	}
	f |= n & MaskX
	if cpu.getBC() != 0 {
		f |= MaskPV
	}
	cpu.F = f
	cpu.Q = f
}

func (cpu *CPUZ80) ldd() {
	val := cpu.readByte(cpu.getHL())
	cpu.writeByte(cpu.getDE(), val)
	cpu.setDE(cpu.getDE() - 1)
	cpu.setHL(cpu.getHL() - 1)
	cpu.setBC(cpu.getBC() - 1)

	n := val + cpu.A
	f := cpu.F & (MaskS | MaskZ | MaskC)
	if n&0x02 != 0 {
		f |= MaskY
	}
	f |= n & MaskX
	if cpu.getBC() != 0 {
		f |= MaskPV
	}
	cpu.F = f
	cpu.Q = f
}

func (cpu *CPUZ80) cpi() {
	hl := cpu.getHL()
	val := cpu.readByte(hl)
	result := cpu.A - val

	cpu.setHL(hl + 1)
	cpu.setBC(cpu.getBC() - 1)

	f := cpu.F & MaskC
	f |= MaskN
	f |= result & MaskS
	if result == 0 {
		f |= MaskZ
	}
	if int(cpu.A&0x0F)-int(val&0x0F) < 0 {
		f |= MaskH
	}
	n := result
	if f&MaskH != 0 {
		n--
	}
	if n&0x02 != 0 {
		f |= MaskY
	}
	f |= n & MaskX
	if cpu.getBC() != 0 {
		f |= MaskPV
	}
	cpu.F = f
	cpu.Q = f
	cpu.WZ++
}

func (cpu *CPUZ80) cpd() {
	hl := cpu.getHL()
	val := cpu.readByte(hl)
	result := cpu.A - val

	cpu.setHL(hl - 1)
	cpu.setBC(cpu.getBC() - 1)

	f := cpu.F & MaskC
	f |= MaskN
	f |= result & MaskS
	if result == 0 {
		f |= MaskZ
	}
	if int(cpu.A&0x0F)-int(val&0x0F) < 0 {
		f |= MaskH
	}
	n := result
	if f&MaskH != 0 {
		n--
	}
	if n&0x02 != 0 {
		f |= MaskY
	}
	f |= n & MaskX
	if cpu.getBC() != 0 {
		f |= MaskPV
	}
	cpu.F = f
	cpu.Q = f
	cpu.WZ--
}

func (cpu *CPUZ80) ini() {
	addr := cpu.getBC()
	val := cpu.readPort(addr)
	cpu.WZ = addr + 1
	cpu.B--
	cpu.writeByte(cpu.getHL(), val)
	cpu.setHL(cpu.getHL() + 1)

	f := cpu.szFlags(cpu.B)
	f |= cpu.B & (MaskY | MaskX)
	if val&0x80 != 0 {
		f |= MaskN
	}
	t := uint16(val) + uint16((cpu.C+1)&0xFF)
	if t > 255 {
		f |= MaskC | MaskH
	}
	if parityTable[byte(t&0x07)^cpu.B] {
		f |= MaskPV
	}
	cpu.F = f
	cpu.Q = f
}

func (cpu *CPUZ80) ind() {
	addr := cpu.getBC()
	val := cpu.readPort(addr)
	cpu.WZ = addr - 1
	cpu.B--
	cpu.writeByte(cpu.getHL(), val)
	cpu.setHL(cpu.getHL() - 1)

	f := cpu.szFlags(cpu.B)
	f |= cpu.B & (MaskY | MaskX)
	if val&0x80 != 0 {
		f |= MaskN
	}
	t := uint16(val) + uint16((cpu.C-1)&0xFF)
	if t > 255 {
		f |= MaskC | MaskH
	}
	if parityTable[byte(t&0x07)^cpu.B] {
		f |= MaskPV
	}
	cpu.F = f
	cpu.Q = f
}

func (cpu *CPUZ80) outi() {
	val := cpu.readByte(cpu.getHL())
	cpu.B--
	addr := cpu.getBC()
	cpu.writePort(addr, val)
	cpu.setHL(cpu.getHL() + 1)
	cpu.WZ = addr + 1

	f := cpu.szFlags(cpu.B)
	f |= cpu.B & (MaskY | MaskX)
	if val&0x80 != 0 {
		f |= MaskN
	}
	t := uint16(val) + uint16(cpu.L)
	if t > 255 {
		f |= MaskC | MaskH
	}
	if parityTable[byte(t&0x07)^cpu.B] {
		f |= MaskPV
	}
	cpu.F = f
	cpu.Q = f
}

func (cpu *CPUZ80) outd() {
	val := cpu.readByte(cpu.getHL())
	cpu.B--
	addr := cpu.getBC()
	cpu.writePort(addr, val)
	cpu.setHL(cpu.getHL() - 1)
	cpu.WZ = addr - 1

	f := cpu.szFlags(cpu.B)
	f |= cpu.B & (MaskY | MaskX)
	if val&0x80 != 0 {
		f |= MaskN
	}
	t := uint16(val) + uint16(cpu.L)
	if t > 255 {
		f |= MaskC | MaskH
	}
	if parityTable[byte(t&0x07)^cpu.B] {
		f |= MaskPV
	}
	cpu.F = f
	cpu.Q = f
}
