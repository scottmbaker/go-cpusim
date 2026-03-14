package cpuz80

// executeCB handles CB-prefixed opcodes (bit operations, rotates, shifts)
func (cpu *CPUZ80) executeCB() error {
	cpu.incR()
	opcode := cpu.fetchByte()

	r := opcode & 0x07
	op := opcode >> 3

	if op < 8 {
		// Rotate/shift operations
		val := cpu.getReg8(r)
		var result byte
		switch op {
		case 0: // RLC
			result = cpu.rlc(val)
		case 1: // RRC
			result = cpu.rrc(val)
		case 2: // RL
			result = cpu.rl(val)
		case 3: // RR
			result = cpu.rr(val)
		case 4: // SLA
			result = cpu.sla(val)
		case 5: // SRA
			result = cpu.sra(val)
		case 6: // SLL (undocumented)
			result = cpu.sll(val)
		case 7: // SRL
			result = cpu.srl(val)
		}
		cpu.setReg8(r, result)
		return nil
	}

	if op < 16 {
		// BIT b,r
		b := op - 8
		val := cpu.getReg8(r)
		if r == 6 {
			cpu.bitM(b, val)
		} else {
			cpu.bit(b, val)
		}
		return nil
	}

	if op < 24 {
		// RES b,r
		b := op - 16
		val := cpu.getReg8(r)
		val &= ^(1 << b)
		cpu.setReg8(r, val)
		return nil
	}

	// SET b,r
	b := op - 24
	val := cpu.getReg8(r)
	val |= 1 << b
	cpu.setReg8(r, val)
	return nil
}
