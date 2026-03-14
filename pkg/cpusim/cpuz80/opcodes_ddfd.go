package cpuz80

// executeDD handles DD-prefixed opcodes (IX register operations)
func (cpu *CPUZ80) executeDD() error {
	return cpu.executeIndexed(&cpu.IX)
}

// executeFD handles FD-prefixed opcodes (IY register operations)
func (cpu *CPUZ80) executeFD() error {
	return cpu.executeIndexed(&cpu.IY)
}

// executeIndexed handles DD/FD-prefixed opcodes generically
func (cpu *CPUZ80) executeIndexed(idx *uint16) error {
	cpu.incR()
	opcode := cpu.fetchByte()

	switch opcode {
	case 0x09: // ADD IX/IY,BC
		cpu.addIdx(idx, cpu.getBC())
		return nil
	case 0x19: // ADD IX/IY,DE
		cpu.addIdx(idx, cpu.getDE())
		return nil
	case 0x29: // ADD IX/IY,IX/IY
		cpu.addIdx(idx, *idx)
		return nil
	case 0x39: // ADD IX/IY,SP
		cpu.addIdx(idx, cpu.SP)
		return nil

	case 0x21: // LD IX/IY,nn
		*idx = cpu.fetchWord()
		return nil

	case 0x22: // LD (nn),IX/IY
		addr := cpu.fetchWord()
		cpu.writeWord(addr, *idx)
		cpu.WZ = addr + 1
		return nil

	case 0x2A: // LD IX/IY,(nn)
		addr := cpu.fetchWord()
		*idx = cpu.readWord(addr)
		cpu.WZ = addr + 1
		return nil

	case 0x23: // INC IX/IY
		*idx++
		return nil

	case 0x2B: // DEC IX/IY
		*idx--
		return nil

	// INC/DEC IXH/IXL (undocumented)
	case 0x24: // INC IXH/IYH
		h := byte(*idx >> 8)
		h = cpu.inc8(h)
		*idx = (*idx & 0x00FF) | (uint16(h) << 8)
		return nil
	case 0x2C: // INC IXL/IYL
		l := byte(*idx & 0xFF)
		l = cpu.inc8(l)
		*idx = (*idx & 0xFF00) | uint16(l)
		return nil
	case 0x25: // DEC IXH/IYH
		h := byte(*idx >> 8)
		h = cpu.dec8(h)
		*idx = (*idx & 0x00FF) | (uint16(h) << 8)
		return nil
	case 0x2D: // DEC IXL/IYL
		l := byte(*idx & 0xFF)
		l = cpu.dec8(l)
		*idx = (*idx & 0xFF00) | uint16(l)
		return nil

	// LD IXH/IXL,n (undocumented)
	case 0x26: // LD IXH/IYH,n
		n := cpu.fetchByte()
		*idx = (*idx & 0x00FF) | (uint16(n) << 8)
		return nil
	case 0x2E: // LD IXL/IYL,n
		n := cpu.fetchByte()
		*idx = (*idx & 0xFF00) | uint16(n)
		return nil

	case 0x34: // INC (IX/IY+d)
		d := cpu.fetchByte()
		addr := uint16(int32(*idx) + int32(int8(d)))
		cpu.WZ = addr
		cpu.writeByte(addr, cpu.inc8(cpu.readByte(addr)))
		return nil

	case 0x35: // DEC (IX/IY+d)
		d := cpu.fetchByte()
		addr := uint16(int32(*idx) + int32(int8(d)))
		cpu.WZ = addr
		cpu.writeByte(addr, cpu.dec8(cpu.readByte(addr)))
		return nil

	case 0x36: // LD (IX/IY+d),n
		d := cpu.fetchByte()
		n := cpu.fetchByte()
		addr := uint16(int32(*idx) + int32(int8(d)))
		cpu.WZ = addr
		cpu.writeByte(addr, n)
		return nil

	// LD r,(IX/IY+d) and LD (IX/IY+d),r
	case 0x46, 0x4E, 0x56, 0x5E, 0x66, 0x6E, 0x7E: // LD r,(IX/IY+d)
		d := cpu.fetchByte()
		addr := uint16(int32(*idx) + int32(int8(d)))
		cpu.WZ = addr
		r := (opcode >> 3) & 0x07
		cpu.setReg8(r, cpu.readByte(addr))
		return nil

	case 0x70, 0x71, 0x72, 0x73, 0x74, 0x75, 0x77: // LD (IX/IY+d),r
		d := cpu.fetchByte()
		addr := uint16(int32(*idx) + int32(int8(d)))
		cpu.WZ = addr
		r := opcode & 0x07
		cpu.writeByte(addr, cpu.getReg8(r))
		return nil

	// Undocumented LD r,IXH/IXL and LD IXH/IXL,r operations
	// LD B,IXH/IYH
	case 0x44:
		cpu.B = byte(*idx >> 8)
		return nil
	// LD B,IXL/IYL
	case 0x45:
		cpu.B = byte(*idx & 0xFF)
		return nil
	// LD C,IXH/IYH
	case 0x4C:
		cpu.C = byte(*idx >> 8)
		return nil
	// LD C,IXL/IYL
	case 0x4D:
		cpu.C = byte(*idx & 0xFF)
		return nil
	// LD D,IXH/IYH
	case 0x54:
		cpu.D = byte(*idx >> 8)
		return nil
	// LD D,IXL/IYL
	case 0x55:
		cpu.D = byte(*idx & 0xFF)
		return nil
	// LD E,IXH/IYH
	case 0x5C:
		cpu.E = byte(*idx >> 8)
		return nil
	// LD E,IXL/IYL
	case 0x5D:
		cpu.E = byte(*idx & 0xFF)
		return nil
	// LD IXH,B..A (undocumented)
	case 0x60:
		*idx = (*idx & 0x00FF) | (uint16(cpu.B) << 8)
		return nil
	case 0x61:
		*idx = (*idx & 0x00FF) | (uint16(cpu.C) << 8)
		return nil
	case 0x62:
		*idx = (*idx & 0x00FF) | (uint16(cpu.D) << 8)
		return nil
	case 0x63:
		*idx = (*idx & 0x00FF) | (uint16(cpu.E) << 8)
		return nil
	case 0x64: // LD IXH,IXH
		return nil
	case 0x65: // LD IXH,IXL
		*idx = (*idx & 0x00FF) | (uint16(byte(*idx&0xFF)) << 8)
		return nil
	case 0x67:
		*idx = (*idx & 0x00FF) | (uint16(cpu.A) << 8)
		return nil
	// LD IXL,B..A (undocumented)
	case 0x68:
		*idx = (*idx & 0xFF00) | uint16(cpu.B)
		return nil
	case 0x69:
		*idx = (*idx & 0xFF00) | uint16(cpu.C)
		return nil
	case 0x6A:
		*idx = (*idx & 0xFF00) | uint16(cpu.D)
		return nil
	case 0x6B:
		*idx = (*idx & 0xFF00) | uint16(cpu.E)
		return nil
	case 0x6C: // LD IXL,IXH
		*idx = (*idx & 0xFF00) | uint16(byte(*idx>>8))
		return nil
	case 0x6D: // LD IXL,IXL
		return nil
	case 0x6F:
		*idx = (*idx & 0xFF00) | uint16(cpu.A)
		return nil

	// LD A,IXH/IXL
	case 0x7C:
		cpu.A = byte(*idx >> 8)
		return nil
	case 0x7D:
		cpu.A = byte(*idx & 0xFF)
		return nil

	// ALU A,(IX/IY+d)
	case 0x86: // ADD A,(IX/IY+d)
		d := cpu.fetchByte()
		addr := uint16(int32(*idx) + int32(int8(d)))
		cpu.WZ = addr
		cpu.add8(cpu.readByte(addr))
		return nil
	case 0x8E: // ADC A,(IX/IY+d)
		d := cpu.fetchByte()
		addr := uint16(int32(*idx) + int32(int8(d)))
		cpu.WZ = addr
		cpu.adc8(cpu.readByte(addr))
		return nil
	case 0x96: // SUB (IX/IY+d)
		d := cpu.fetchByte()
		addr := uint16(int32(*idx) + int32(int8(d)))
		cpu.WZ = addr
		cpu.sub8(cpu.readByte(addr))
		return nil
	case 0x9E: // SBC A,(IX/IY+d)
		d := cpu.fetchByte()
		addr := uint16(int32(*idx) + int32(int8(d)))
		cpu.WZ = addr
		cpu.sbc8(cpu.readByte(addr))
		return nil
	case 0xA6: // AND (IX/IY+d)
		d := cpu.fetchByte()
		addr := uint16(int32(*idx) + int32(int8(d)))
		cpu.WZ = addr
		cpu.and8(cpu.readByte(addr))
		return nil
	case 0xAE: // XOR (IX/IY+d)
		d := cpu.fetchByte()
		addr := uint16(int32(*idx) + int32(int8(d)))
		cpu.WZ = addr
		cpu.xor8(cpu.readByte(addr))
		return nil
	case 0xB6: // OR (IX/IY+d)
		d := cpu.fetchByte()
		addr := uint16(int32(*idx) + int32(int8(d)))
		cpu.WZ = addr
		cpu.or8(cpu.readByte(addr))
		return nil
	case 0xBE: // CP (IX/IY+d)
		d := cpu.fetchByte()
		addr := uint16(int32(*idx) + int32(int8(d)))
		cpu.WZ = addr
		cpu.cp8(cpu.readByte(addr))
		return nil

	// Undocumented ALU A,IXH/IXL
	case 0x84: // ADD A,IXH/IYH
		cpu.add8(byte(*idx >> 8))
		return nil
	case 0x85: // ADD A,IXL/IYL
		cpu.add8(byte(*idx & 0xFF))
		return nil
	case 0x8C: // ADC A,IXH/IYH
		cpu.adc8(byte(*idx >> 8))
		return nil
	case 0x8D: // ADC A,IXL/IYL
		cpu.adc8(byte(*idx & 0xFF))
		return nil
	case 0x94: // SUB IXH/IYH
		cpu.sub8(byte(*idx >> 8))
		return nil
	case 0x95: // SUB IXL/IYL
		cpu.sub8(byte(*idx & 0xFF))
		return nil
	case 0x9C: // SBC A,IXH/IYH
		cpu.sbc8(byte(*idx >> 8))
		return nil
	case 0x9D: // SBC A,IXL/IYL
		cpu.sbc8(byte(*idx & 0xFF))
		return nil
	case 0xA4: // AND IXH/IYH
		cpu.and8(byte(*idx >> 8))
		return nil
	case 0xA5: // AND IXL/IYL
		cpu.and8(byte(*idx & 0xFF))
		return nil
	case 0xAC: // XOR IXH/IYH
		cpu.xor8(byte(*idx >> 8))
		return nil
	case 0xAD: // XOR IXL/IYL
		cpu.xor8(byte(*idx & 0xFF))
		return nil
	case 0xB4: // OR IXH/IYH
		cpu.or8(byte(*idx >> 8))
		return nil
	case 0xB5: // OR IXL/IYL
		cpu.or8(byte(*idx & 0xFF))
		return nil
	case 0xBC: // CP IXH/IYH
		cpu.cp8(byte(*idx >> 8))
		return nil
	case 0xBD: // CP IXL/IYL
		cpu.cp8(byte(*idx & 0xFF))
		return nil

	// PUSH/POP IX/IY
	case 0xE5: // PUSH IX/IY
		cpu.push(*idx)
		return nil
	case 0xE1: // POP IX/IY
		*idx = cpu.pop()
		return nil

	// EX (SP),IX/IY
	case 0xE3:
		lo := cpu.readByte(cpu.SP)
		hi := cpu.readByte(cpu.SP + 1)
		cpu.writeByte(cpu.SP, byte(*idx&0xFF))
		cpu.writeByte(cpu.SP+1, byte(*idx>>8))
		*idx = uint16(hi)<<8 | uint16(lo)
		cpu.WZ = *idx
		return nil

	// JP (IX/IY)
	case 0xE9:
		cpu.PC = *idx
		return nil

	// LD SP,IX/IY
	case 0xF9:
		cpu.SP = *idx
		return nil

	case 0xCB: // DDCB/FDCB prefix
		return cpu.executeIndexedCB(idx)

	default:
		// Unrecognized DD/FD opcodes fall through to the unprefixed handler
		return cpu.executeUnprefixed(opcode)
	}
}

// executeIndexedCB handles DDCB and FDCB prefixed opcodes
func (cpu *CPUZ80) executeIndexedCB(idx *uint16) error {
	// Format: DD CB dd oo  (displacement comes BEFORE the opcode)
	d := cpu.fetchByte()
	addr := uint16(int32(*idx) + int32(int8(d)))
	cpu.WZ = addr
	opcode := cpu.fetchByte()
	cpu.incR()

	r := opcode & 0x07
	op := opcode >> 3

	if op < 8 {
		// Rotate/shift on (IX/IY+d), result also stored in register (undocumented)
		val := cpu.readByte(addr)
		var result byte
		switch op {
		case 0:
			result = cpu.rlc(val)
		case 1:
			result = cpu.rrc(val)
		case 2:
			result = cpu.rl(val)
		case 3:
			result = cpu.rr(val)
		case 4:
			result = cpu.sla(val)
		case 5:
			result = cpu.sra(val)
		case 6:
			result = cpu.sll(val)
		case 7:
			result = cpu.srl(val)
		}
		cpu.writeByte(addr, result)
		if r != 6 {
			cpu.setReg8(r, result) // Undocumented: also store in register
		}
		return nil
	}

	if op < 16 {
		// BIT b,(IX/IY+d)
		b := op - 8
		val := cpu.readByte(addr)
		cpu.bitM(b, val)
		return nil
	}

	if op < 24 {
		// RES b,(IX/IY+d)
		b := op - 16
		val := cpu.readByte(addr)
		result := val & ^(1 << b)
		cpu.writeByte(addr, result)
		if r != 6 {
			cpu.setReg8(r, result)
		}
		return nil
	}

	// SET b,(IX/IY+d)
	b := op - 24
	val := cpu.readByte(addr)
	result := val | (1 << b)
	cpu.writeByte(addr, result)
	if r != 6 {
		cpu.setReg8(r, result)
	}
	return nil
}
