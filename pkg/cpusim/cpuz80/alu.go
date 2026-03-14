package cpuz80

func (cpu *CPUZ80) add8(val byte) {
	a := cpu.A
	result16 := uint16(a) + uint16(val)
	result := byte(result16)

	var f byte
	f |= result & (MaskS | MaskY | MaskX)
	if result == 0 {
		f |= MaskZ
	}
	if result16 > 0xFF {
		f |= MaskC
	}
	if (a&0x0F)+(val&0x0F) > 0x0F {
		f |= MaskH
	}
	if (a^val)&0x80 == 0 && (a^result)&0x80 != 0 {
		f |= MaskPV
	}

	cpu.A = result
	cpu.F = f
	cpu.Q = f
}

func (cpu *CPUZ80) adc8(val byte) {
	a := cpu.A
	carry := uint16(cpu.F & MaskC)
	result16 := uint16(a) + uint16(val) + carry
	result := byte(result16)

	var f byte
	f |= result & (MaskS | MaskY | MaskX)
	if result == 0 {
		f |= MaskZ
	}
	if result16 > 0xFF {
		f |= MaskC
	}
	if (a&0x0F)+(val&0x0F)+byte(carry) > 0x0F {
		f |= MaskH
	}
	if (a^val)&0x80 == 0 && (a^result)&0x80 != 0 {
		f |= MaskPV
	}

	cpu.A = result
	cpu.F = f
	cpu.Q = f
}

func (cpu *CPUZ80) sub8(val byte) {
	a := cpu.A
	result16 := uint16(a) - uint16(val)
	result := byte(result16)

	var f byte
	f |= MaskN
	f |= result & (MaskS | MaskY | MaskX)
	if result == 0 {
		f |= MaskZ
	}
	if result16 > 0xFF {
		f |= MaskC
	}
	if int(a&0x0F)-int(val&0x0F) < 0 {
		f |= MaskH
	}
	if (a^val)&0x80 != 0 && (a^result)&0x80 != 0 {
		f |= MaskPV
	}

	cpu.A = result
	cpu.F = f
	cpu.Q = f
}

func (cpu *CPUZ80) sbc8(val byte) {
	a := cpu.A
	carry := uint16(cpu.F & MaskC)
	result16 := uint16(a) - uint16(val) - carry
	result := byte(result16)

	var f byte
	f |= MaskN
	f |= result & (MaskS | MaskY | MaskX)
	if result == 0 {
		f |= MaskZ
	}
	if result16 > 0xFF {
		f |= MaskC
	}
	if int(a&0x0F)-int(val&0x0F)-int(carry) < 0 {
		f |= MaskH
	}
	if (a^val)&0x80 != 0 && (a^result)&0x80 != 0 {
		f |= MaskPV
	}

	cpu.A = result
	cpu.F = f
	cpu.Q = f
}

func (cpu *CPUZ80) and8(val byte) {
	cpu.A &= val
	var f byte
	f |= MaskH
	f |= cpu.A & (MaskS | MaskY | MaskX)
	if cpu.A == 0 {
		f |= MaskZ
	}
	if parityTable[cpu.A] {
		f |= MaskPV
	}
	cpu.F = f
	cpu.Q = f
}

func (cpu *CPUZ80) xor8(val byte) {
	cpu.A ^= val
	var f byte
	f |= cpu.A & (MaskS | MaskY | MaskX)
	if cpu.A == 0 {
		f |= MaskZ
	}
	if parityTable[cpu.A] {
		f |= MaskPV
	}
	cpu.F = f
	cpu.Q = f
}

func (cpu *CPUZ80) or8(val byte) {
	cpu.A |= val
	var f byte
	f |= cpu.A & (MaskS | MaskY | MaskX)
	if cpu.A == 0 {
		f |= MaskZ
	}
	if parityTable[cpu.A] {
		f |= MaskPV
	}
	cpu.F = f
	cpu.Q = f
}

func (cpu *CPUZ80) cp8(val byte) {
	a := cpu.A
	result16 := uint16(a) - uint16(val)
	result := byte(result16)

	var f byte
	f |= MaskN
	f |= result & MaskS
	f |= val & (MaskY | MaskX) // CP: X/Y from operand, not result
	if result == 0 {
		f |= MaskZ
	}
	if result16 > 0xFF {
		f |= MaskC
	}
	if int(a&0x0F)-int(val&0x0F) < 0 {
		f |= MaskH
	}
	if (a^val)&0x80 != 0 && (a^result)&0x80 != 0 {
		f |= MaskPV
	}

	cpu.F = f
	cpu.Q = f
}

func (cpu *CPUZ80) inc8(val byte) byte {
	result := val + 1

	f := cpu.F & MaskC
	f |= result & (MaskS | MaskY | MaskX)
	if result == 0 {
		f |= MaskZ
	}
	if val&0x0F == 0x0F {
		f |= MaskH
	}
	if val == 0x7F {
		f |= MaskPV
	}
	cpu.F = f
	cpu.Q = f
	return result
}

func (cpu *CPUZ80) dec8(val byte) byte {
	result := val - 1

	f := cpu.F & MaskC
	f |= MaskN
	f |= result & (MaskS | MaskY | MaskX)
	if result == 0 {
		f |= MaskZ
	}
	if val&0x0F == 0x00 {
		f |= MaskH
	}
	if val == 0x80 {
		f |= MaskPV
	}
	cpu.F = f
	cpu.Q = f
	return result
}

func (cpu *CPUZ80) addHL(val uint16) {
	hl := cpu.getHL()
	result := uint32(hl) + uint32(val)

	f := cpu.F & (MaskS | MaskZ | MaskPV) // Preserve S, Z, P/V
	f |= byte(result>>8) & (MaskY | MaskX)
	if result > 0xFFFF {
		f |= MaskC
	}
	if (hl&0x0FFF)+(val&0x0FFF) > 0x0FFF {
		f |= MaskH
	}

	cpu.setHL(uint16(result))
	cpu.F = f
	cpu.Q = f
	cpu.WZ = hl + 1
}

func (cpu *CPUZ80) addIdx(idx *uint16, val uint16) {
	base := *idx
	result := uint32(base) + uint32(val)

	f := cpu.F & (MaskS | MaskZ | MaskPV)
	f |= byte(result>>8) & (MaskY | MaskX)
	if result > 0xFFFF {
		f |= MaskC
	}
	if (base&0x0FFF)+(val&0x0FFF) > 0x0FFF {
		f |= MaskH
	}

	*idx = uint16(result)
	cpu.F = f
	cpu.Q = f
	cpu.WZ = base + 1
}

func (cpu *CPUZ80) daa() {
	a := cpu.A
	correction := byte(0)
	carry := cpu.F&MaskC != 0
	halfCarry := cpu.F&MaskH != 0
	nFlag := cpu.F&MaskN != 0

	if halfCarry || (a&0x0F) > 9 {
		correction |= 0x06
	}
	if carry || a > 0x99 {
		correction |= 0x60
	}

	newCarry := carry || a > 0x99

	if nFlag {
		cpu.A = a - correction
	} else {
		cpu.A = a + correction
	}

	newH := byte(0)
	if (a^correction^cpu.A)&0x10 != 0 {
		newH = MaskH
	}

	var f byte
	if nFlag {
		f |= MaskN
	}
	f |= cpu.A & (MaskS | MaskY | MaskX)
	if cpu.A == 0 {
		f |= MaskZ
	}
	if parityTable[cpu.A] {
		f |= MaskPV
	}
	if newCarry {
		f |= MaskC
	}
	f |= newH

	cpu.F = f
	cpu.Q = f
}

func (cpu *CPUZ80) rlc(val byte) byte {
	result := (val << 1) | (val >> 7)
	var f byte
	f |= result & (MaskS | MaskY | MaskX)
	if result == 0 {
		f |= MaskZ
	}
	if parityTable[result] {
		f |= MaskPV
	}
	if val&0x80 != 0 {
		f |= MaskC
	}
	cpu.F = f
	cpu.Q = f
	return result
}

func (cpu *CPUZ80) rrc(val byte) byte {
	result := (val >> 1) | (val << 7)
	var f byte
	f |= result & (MaskS | MaskY | MaskX)
	if result == 0 {
		f |= MaskZ
	}
	if parityTable[result] {
		f |= MaskPV
	}
	if val&0x01 != 0 {
		f |= MaskC
	}
	cpu.F = f
	cpu.Q = f
	return result
}

func (cpu *CPUZ80) rl(val byte) byte {
	oldCarry := cpu.F & MaskC
	result := (val << 1) | oldCarry
	var f byte
	f |= result & (MaskS | MaskY | MaskX)
	if result == 0 {
		f |= MaskZ
	}
	if parityTable[result] {
		f |= MaskPV
	}
	if val&0x80 != 0 {
		f |= MaskC
	}
	cpu.F = f
	cpu.Q = f
	return result
}

func (cpu *CPUZ80) rr(val byte) byte {
	oldCarry := cpu.F & MaskC
	result := (val >> 1) | (oldCarry << 7)
	var f byte
	f |= result & (MaskS | MaskY | MaskX)
	if result == 0 {
		f |= MaskZ
	}
	if parityTable[result] {
		f |= MaskPV
	}
	if val&0x01 != 0 {
		f |= MaskC
	}
	cpu.F = f
	cpu.Q = f
	return result
}

func (cpu *CPUZ80) sla(val byte) byte {
	result := val << 1
	var f byte
	f |= result & (MaskS | MaskY | MaskX)
	if result == 0 {
		f |= MaskZ
	}
	if parityTable[result] {
		f |= MaskPV
	}
	if val&0x80 != 0 {
		f |= MaskC
	}
	cpu.F = f
	cpu.Q = f
	return result
}

func (cpu *CPUZ80) sra(val byte) byte {
	result := (val >> 1) | (val & 0x80)
	var f byte
	f |= result & (MaskS | MaskY | MaskX)
	if result == 0 {
		f |= MaskZ
	}
	if parityTable[result] {
		f |= MaskPV
	}
	if val&0x01 != 0 {
		f |= MaskC
	}
	cpu.F = f
	cpu.Q = f
	return result
}

func (cpu *CPUZ80) sll(val byte) byte {
	result := (val << 1) | 0x01
	var f byte
	f |= result & (MaskS | MaskY | MaskX)
	if result == 0 {
		f |= MaskZ
	}
	if parityTable[result] {
		f |= MaskPV
	}
	if val&0x80 != 0 {
		f |= MaskC
	}
	cpu.F = f
	cpu.Q = f
	return result
}

func (cpu *CPUZ80) srl(val byte) byte {
	result := val >> 1
	var f byte
	f |= result & (MaskS | MaskY | MaskX)
	if result == 0 {
		f |= MaskZ
	}
	if parityTable[result] {
		f |= MaskPV
	}
	if val&0x01 != 0 {
		f |= MaskC
	}
	cpu.F = f
	cpu.Q = f
	return result
}

func (cpu *CPUZ80) bit(b byte, val byte) {
	result := val & (1 << b)
	f := cpu.F & MaskC
	f |= MaskH
	if result == 0 {
		f |= MaskZ | MaskPV
	}
	f |= result & MaskS
	f |= val & (MaskY | MaskX)
	cpu.F = f
	cpu.Q = f
}

func (cpu *CPUZ80) bitM(b byte, val byte) {
	result := val & (1 << b)
	f := cpu.F & MaskC
	f |= MaskH
	if result == 0 {
		f |= MaskZ | MaskPV
	}
	f |= result & MaskS
	f |= byte(cpu.WZ>>8) & (MaskY | MaskX)
	cpu.F = f
	cpu.Q = f
}
