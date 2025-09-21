package cpu4004

import (
	"fmt"
	"github.com/scottmbaker/gocpusim/pkg/cpusim"
)

type CPU4004 struct {
	Sim       *cpusim.CpuSim // Reference to the CPU simulation
	Name      string         // Name of the CPU
	Registers [20]byte       // 4-bit registers
	Stack     [3]uint16
	RC        byte
	SP        byte
	PC        uint16 // Program Counter
	Halted    bool   // Flag to indicate if the CPU is halted
	NewStyle  bool   // Flag to indicate if the new style debugging is used
}

const (
	REG_R0    = 0
	REG_R1    = 1
	REG_R2    = 2
	REG_R3    = 3
	REG_R4    = 4
	REG_R5    = 5
	REG_R6    = 6
	REG_R7    = 7
	REG_R8    = 8
	REG_R9    = 9
	REG_R10   = 10
	REG_R11   = 11
	REG_R12   = 12
	REG_R13   = 13
	REG_R14   = 14
	REG_R15   = 15
	REG_ACCUM = 16
	REG_CL    = 17

	FLAG_CARRY = 18
	FLAG_TEST  = 19

	PAIR_P0 = 0
	PAIR_P1 = 1
	PAIR_P2 = 2
	PAIR_P3 = 3
	PAIR_P4 = 4
	PAIR_P5 = 5
	PAIR_P6 = 6
	PAIR_P7 = 7

	PAIR_RC = 8

	OP_ADD = 0
	OP_SUB = 1
	OP_INC = 2
	OP_DEC = 3
	OP_RAL = 4
	OP_RAR = 5
	OP_TCC = 6
	OP_TCS = 7
	OP_DAA = 8
	OP_KBP = 9
	OP_CMA = 10
	OP_ADM = 11
	OP_SBM = 12
)

var KBPTable = [16]int{0, 1, 2, 3, 4, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15}

func New4004(sim *cpusim.CpuSim, name string) *CPU4004 {
	return &CPU4004{
		Name:      name,
		Registers: [20]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		Stack:     [3]uint16{0, 0, 0},
		SP:        0,
		PC:        0, // Program Counter
		Sim:       sim,
		NewStyle:  true, // Use new style for debugging
	}
}

func toBit(value bool) byte {
	if value {
		return 1
	}
	return 0
}

func (cpu *CPU4004) GetName() string {
	return cpu.Name
}

func (cpu *CPU4004) DCLEnabler(value byte) *cpusim.ByReferenceByteEnabler {
	return cpusim.NewByReferenceByteEnabler(&cpu.Registers[REG_CL], value)
}

func (cpu *CPU4004) GetRegName(reg int) string {
	switch reg {
	case REG_R0:
		return "R0"
	case REG_R1:
		return "R1"
	case REG_R2:
		return "R2"
	case REG_R3:
		return "R3"
	case REG_R4:
		return "R4"
	case REG_R5:
		return "R5"
	case REG_R6:
		return "R6"
	case REG_R7:
		return "R7"
	case REG_R8:
		return "R8"
	case REG_R9:
		return "R9"
	case REG_R10:
		return "R10"
	case REG_R11:
		return "R11"
	case REG_R12:
		return "R12"
	case REG_R13:
		return "R13"
	case REG_R14:
		return "R14"
	case REG_R15:
		return "R15"
	case REG_ACCUM:
		return "ACCUM"
	case REG_CL:
		return "CL"
	case FLAG_CARRY:
		return "C"
	case FLAG_TEST:
		return "T"
	default:
		return fmt.Sprintf("R%d", reg)
	}
}

func (cpu *CPU4004) GetPairName(pair int) string {
	switch pair {
	case PAIR_P0:
		return "P0"
	case PAIR_P1:
		return "P1"
	case PAIR_P2:
		return "P2"
	case PAIR_P3:
		return "P3"
	case PAIR_P4:
		return "P4"
	case PAIR_P5:
		return "P5"
	case PAIR_P6:
		return "P6"
	case PAIR_P7:
		return "P7"
	case PAIR_RC:
		return "RC"
	default:
		return fmt.Sprintf("P%d", pair)
	}
}

func (cpu *CPU4004) GetReg(register int) (byte, error) {
	if register < 0 || register >= len(cpu.Registers) {
		return 0, &cpusim.ErrInvalidRegister{Device: cpu, Register: register}
	}
	return cpu.Registers[register], nil
}

func (cpu *CPU4004) SetReg(register int, value byte) error {
	if register < 0 || register >= len(cpu.Registers) {
		return &cpusim.ErrInvalidRegister{Register: register}
	}
	cpu.Registers[register] = value
	return nil
}

func (cpu *CPU4004) GetPair(pair int) (byte, error) {
	if pair == PAIR_RC {
		return cpu.RC, nil
	}

	if pair < 0 || pair >= 8 {
		return 0, &cpusim.ErrInvalidRegister{Device: cpu, Register: pair}
	}
	low, err := cpu.GetReg(pair * 2)
	if err != nil {
		return 0, err
	}
	high, err := cpu.GetReg(pair*2 + 1)
	if err != nil {
		return 0, err
	}
	return byte(high)<<4 | byte(low), nil
}

func (cpu *CPU4004) SetPair(pair int, value byte) error {
	if pair == PAIR_RC {
		cpu.RC = value
		return nil
	}

	if pair < 0 || pair >= 8 {
		return &cpusim.ErrInvalidRegister{Register: pair}
	}
	low := value & 0x0F
	high := (value >> 4) & 0x0F
	err := cpu.SetReg(pair*2, low)
	if err != nil {
		return err
	}
	err = cpu.SetReg(pair*2+1, high)
	return err
}

func (cpu *CPU4004) ExecuteMovePair(destPair int, srcPair int) error {
	srcVal, err := cpu.GetPair(srcPair)
	if err != nil {
		return err
	}
	cpu.DebugMovePair(destPair, srcPair)
	return cpu.SetPair(destPair, srcVal)
}

func (cpu *CPU4004) ExecuteLoad(opCode byte) error {
	srcReg := int(opCode & 0xF)
	srcVal, err := cpu.GetReg(srcReg)
	if err != nil {
		return err
	}
	cpu.DebugMove(REG_ACCUM, srcReg)
	return cpu.SetReg(REG_ACCUM, srcVal)
}

func (cpu *CPU4004) ExecuteExchange(opCode byte) error {
	srcReg := int(opCode & 0xF)
	srcVal, err := cpu.GetReg(srcReg)
	if err != nil {
		return err
	}
	accVal, err := cpu.GetReg(REG_ACCUM)
	if err != nil {
		return err
	}
	cpu.DebugExchange(REG_ACCUM, srcReg)
	err = cpu.SetReg(srcReg, accVal)
	if err != nil {
		return err
	}
	return cpu.SetReg(REG_ACCUM, srcVal)
}

func (cpu *CPU4004) ExecuteLoadImmediate(opCode byte, value byte) error {
	cpu.DebugMoveImmediate(REG_ACCUM, value)
	return cpu.SetReg(REG_ACCUM, value)
}

func (cpu *CPU4004) ExecuteFetchImmediate(opCode byte) error {
	destPair := int((opCode >> 1) & 0x07)
	value, err := cpu.FetchOpcode()
	if err != nil {
		return err
	}
	cpu.DebugFetchImmediate(destPair, value)
	return cpu.SetPair(destPair, value)
}

func (cpu *CPU4004) FetchOpcode() (byte, error) {
	cpu.Sim.FilterMemoryKind(cpusim.KIND_ROM)
	opCode, err := cpu.Sim.ReadMemory(cpusim.Address(cpu.PC))
	if err != nil {
		return 0, err
	}
	cpu.PC++
	return opCode, nil
}

func (cpu *CPU4004) FetchAddr(opCode byte, long bool) (uint16, error) {
	addrHigh := 0
	if long {
		addrHigh = int(opCode & 0x0F)
	}
	addrLow, err := cpu.FetchOpcode()
	if err != nil {
		return 0, err
	}
	return (uint16(addrHigh)<<8 | uint16(addrLow)) & 0x0FFF, nil
}

func (cpu *CPU4004) ExecuteInc(opCode byte) error {
	var work int

	reg := int(opCode & 0x0F)
	value, err := cpu.GetReg(reg)
	if err != nil {
		return err
	}

	work = int(value) + 1
	err = cpu.UpdateIncFlags(work)
	if err != nil {
		return err
	}
	value = byte(work & 0x0F)

	cpu.DebugInc(reg)

	return cpu.SetReg(reg, value)
}

func (cpu *CPU4004) ExecuteIncSkip(opCode byte) error {
	var work int

	reg := int(opCode & 0x0F)
	value, err := cpu.GetReg(reg)
	if err != nil {
		return err
	}

	work = int(value) + 1
	err = cpu.UpdateIncFlags(work)
	if err != nil {
		return err
	}

	addrLow, err := cpu.FetchOpcode()
	if err != nil {
		return err
	}

	value = byte(work & 0x0F)
	err = cpu.SetReg(reg, value)
	if err != nil {
		return err
	}

	if value != 0 {
		cpu.PC = (cpu.PC & 0xFF00) | uint16(addrLow)
	}

	cpu.DebugIncSkip(reg, addrLow)

	return nil
}

func (cpu *CPU4004) updateArithFlags(op int, work int) error {
	var newCarry byte
	if op == OP_SUB || op == OP_SBM {
		newCarry = toBit((work & 0x10) == 0) // on exit from a SUB or SBM, 0 indicates carry and 1 indicates no carry
	} else {
		newCarry = toBit((work & 0x10) != 0)
	}
	err := cpu.SetReg(FLAG_CARRY, newCarry)
	if err != nil {
		return err
	}
	err = cpu.UpdateIncFlags(work)
	return err
}

func (cpu *CPU4004) UpdateIncFlags(work int) error {
	_ = work
	return nil
}

func (cpu *CPU4004) ExecuteWrite(opCode byte) error {
	acc, err := cpu.GetReg(REG_ACCUM)
	if err != nil {
		return err
	}
	where := int(opCode & 0x07)
	cpu.DebugWrite(where)
	switch where {
	case 0:
		cpu.Sim.FilterMemoryKind(cpusim.KIND_RAM)
		return cpu.Sim.WriteMemory(cpusim.Address(cpu.RC), acc)
	case 1:
		// ramport
		return nil
	case 2:
		// romport
		return nil
	case 4, 5, 6, 7:
		cpu.Sim.FilterMemoryKind(cpusim.KIND_RAM)
		return cpu.Sim.WriteMemoryStatus(cpusim.Address(cpu.RC), cpusim.Address(where-4), acc)
	}

	return &cpusim.ErrInvalidOpcode{Device: cpu, Opcode: opCode}
}

func (cpu *CPU4004) ExecuteRead(opCode byte) error {
	where := int(opCode & 0x07)
	var value byte
	var err error
	switch where {
	case 1:
		cpu.Sim.FilterMemoryKind(cpusim.KIND_RAM)
		value, err = cpu.Sim.ReadMemory(cpusim.Address(cpu.RC))
		if err != nil {
			return err
		}
	case 2:
		// romport
		return nil
	case 4, 5, 6, 7:
		cpu.Sim.FilterMemoryKind(cpusim.KIND_RAM)
		value, err = cpu.Sim.ReadMemoryStatus(cpusim.Address(cpu.RC), cpusim.Address(where-4))
		if err != nil {
			return err
		}
	default:
		return &cpusim.ErrInvalidOpcode{Device: cpu, Opcode: opCode}
	}
	cpu.DebugRead(where)
	return cpu.SetReg(REG_ACCUM, value)
}

func (cpu *CPU4004) ExecuteAccumulator(opCode byte, op int) error {
	var work int

	acc, err := cpu.GetReg(REG_ACCUM)
	if err != nil {
		return err
	}

	carryBit, err := cpu.GetReg(FLAG_CARRY)
	if err != nil {
		return err
	}

	reg := 0 // only ADD and SUB need the register
	var val byte
	if op == OP_ADD || op == OP_SUB {
		reg = int(opCode & 0x07)
		val, err = cpu.GetReg(reg)
		if err != nil {
			return err
		}
	}

	if op == OP_ADM || op == OP_SBM {
		cpu.Sim.FilterMemoryKind(cpusim.KIND_RAM)
		val, err = cpu.Sim.ReadMemory(cpusim.Address(cpu.RC))
		if err != nil {
			return err
		}
	}

	work = int(acc)

	switch op {
	case OP_ADD, OP_ADM:
		work = work + int(val)
		if carryBit != 0 {
			work = work + 1
		}
	case OP_SUB, OP_SBM:
		work = work - int(val)
		if carryBit != 0 {
			work = work - 1
		}
	case OP_INC:
		work = work + 1
	case OP_DEC:
		work = work - 1
	case OP_DAA:
		if (acc&0x0F) > 9 || (carryBit != 0) {
			work = work + 6
		}
	case OP_TCS:
		if carryBit != 0 {
			work = 10
		} else {
			work = 9
		}
	case OP_KBP:
		work = KBPTable[work&0x0F]
	case OP_CMA:
		work = (^work) & 0x0F
	}

	err = cpu.updateArithFlags(op, work)
	if err != nil {
		return err
	}

	acc = byte(work & 0xFF)
	cpu.DebugAccumulator(op, reg)

	return cpu.SetReg(REG_ACCUM, acc)
}

func (cpu *CPU4004) ExecuteRotate(op int) error {
	carryBit, _ := cpu.GetReg(FLAG_CARRY)

	acc, err := cpu.GetReg(REG_ACCUM)
	if err != nil {
		return err
	}

	if op == OP_RAL {
		newCarry := (acc >> 3) & 0x01
		acc = ((acc << 1) & 0x0F) | carryBit
		err := cpu.SetReg(FLAG_CARRY, newCarry)
		if err != nil {
			return err
		}
		cpu.DebugInstr("RAL")
		return cpu.SetReg(REG_ACCUM, acc)
	}

	if op == OP_RAR {
		newCarry := acc & 0x01
		acc = (acc >> 1) | ((carryBit << 3) & 0x0F)
		err := cpu.SetReg(FLAG_CARRY, newCarry)
		if err != nil {
			return err
		}
		cpu.DebugInstr("RAR")
		return cpu.SetReg(REG_ACCUM, acc)
	}

	return &cpusim.ErrInvalidOperation{Device: cpu, Operation: byte(op)}
}

func (cpu *CPU4004) PushStack(value uint16) {
	cpu.Stack[cpu.SP] = cpu.PC
	cpu.SP++
	if cpu.SP > byte(len(cpu.Stack)) {
		cpu.SP = 0 // wrap around stack pointer
	}
}

func (cpu *CPU4004) ExecuteJCN(opCode byte) error {
	addr, err := cpu.FetchAddr(opCode, false) // 8-bit address
	if err != nil {
		return err
	}

	addr = cpu.PC&0xFF00 | uint16(addr) // JCN is always an 8 bit address within the current page

	c1 := (opCode >> 3) & 0x01 // 1 == invert
	c2 := (opCode >> 2) & 0x01 // 1 == check accumulator==0
	c3 := (opCode >> 1) & 0x01 // 1 == check carry==1
	c4 := opCode & 0x01        // 1 == check test==0

	invert := (c1 == 1)
	checkAccum := (c2 == 1)
	checkCarry := (c3 == 1)
	checkTest := (c4 == 1)

	acc, err := cpu.GetReg(REG_ACCUM)
	if err != nil {
		return err
	}
	carry, err := cpu.GetReg(FLAG_CARRY)
	if err != nil {
		return err
	}
	test, err := cpu.GetReg(FLAG_TEST)
	if err != nil {
		return err
	}

	zero := (acc == 0)

	jump := (!invert && checkAccum && zero) ||
		(!invert && checkCarry && carry == 1) ||
		(!invert && checkTest && test == 0) ||
		(invert && checkAccum && !zero) ||
		(invert && checkCarry && carry == 0) ||
		(invert && checkTest && test == 0)

	if jump {
		cpu.PC = addr
	}

	cpu.DebugJCN(invert, checkAccum, checkCarry, checkTest, addr)

	return nil
}

func (cpu *CPU4004) ExecuteJump(addr uint16, iscall bool) error {
	if iscall {
		cpu.PushStack(cpu.PC)
	}

	cpu.PC = addr

	cpu.DebugJump(iscall, addr)

	return nil
}

func (cpu *CPU4004) ExecuteBBL(value byte) error {
	if cpu.SP == 0 {
		return fmt.Errorf("stack underflow")
	}
	if cpu.SP == 0 {
		cpu.SP = byte(len(cpu.Stack) - 1)
	} else {
		cpu.SP--
	}
	cpu.PC = cpu.Stack[cpu.SP]

	cpu.DebugRet(value)

	return cpu.SetReg(REG_ACCUM, value)
}

func (cpu *CPU4004) Execute() error {
	if cpu.Sim.Debug {
		fmt.Printf("%04X: ", cpu.PC)
	}

	opCode, err := cpu.FetchOpcode()
	if err != nil {
		return err
	}

	if cpu.Sim.Debug {
		fmt.Printf("[%02X %08b] ", opCode, opCode)
		fmt.Printf("%s ", cpu.String())
	}

	if opCode == 0x00 {
		cpu.DebugInstr("NOP")
		return nil
	}

	if opCode == 0x07 {
		cpu.DebugInstr("AN7-4004-NOP")
		return nil
	}

	if opCode&0xF0 == 0x10 {
		// JCN
		return cpu.ExecuteJCN(opCode)
	}

	if opCode&0xF1 == 0x20 {
		return cpu.ExecuteFetchImmediate(opCode)
	}

	if opCode&0xF1 == 0x21 {
		// SRC
		return cpu.ExecuteMovePair(PAIR_RC, int((opCode>>1)&0x07))
	}

	if opCode&0xF1 == 0x30 {
		// FIN
		return cpu.ExecuteMovePair(int((opCode>>1)&0x07), PAIR_P0)
	}

	if opCode&0xF1 == 0x31 {
		// JIN
	}

	if opCode&0xF0 == 0x40 {
		// JUN
		addr, err := cpu.FetchAddr(opCode, true) // 12-bit address
		if err != nil {
			return err
		}
		return cpu.ExecuteJump(addr, false)
	}

	if opCode&0xF0 == 0x50 {
		// JMS
		addr, err := cpu.FetchAddr(opCode, true) // 12-bit address
		if err != nil {
			return err
		}
		return cpu.ExecuteJump(addr, true)
	}

	if opCode&0xF0 == 0x60 {
		return cpu.ExecuteInc(opCode)
	}

	if opCode&0xF0 == 0x70 {
		// ISZ
		return cpu.ExecuteIncSkip(opCode)
	}

	if opCode&0xF0 == 0x80 {
		return cpu.ExecuteAccumulator(opCode, OP_ADD)
	}

	if opCode&0xF0 == 0x90 {
		return cpu.ExecuteAccumulator(opCode, OP_SUB)
	}

	if opCode&0xF0 == 0xA0 {
		return cpu.ExecuteLoad(opCode)
	}

	if opCode&0xF0 == 0xB0 {
		return cpu.ExecuteExchange(opCode)
	}

	if opCode&0xF0 == 0xC0 {
		return cpu.ExecuteBBL(opCode & 0x0F)
	}

	if opCode&0xF0 == 0xD0 {
		return cpu.ExecuteLoadImmediate(opCode, opCode&0x0F)
	}

	if opCode&0xF8 == 0xE0 {
		return cpu.ExecuteWrite(opCode)
	}

	if opCode == 0xE8 {
		return cpu.ExecuteAccumulator(opCode, OP_ADM)
	}

	if opCode == 0xEB {
		return cpu.ExecuteAccumulator(opCode, OP_SBM)
	}

	if opCode&0xF8 == 0xE8 {
		return cpu.ExecuteRead(opCode)
	}

	if opCode == 0xF0 {
		cpu.DebugInstr("CLB")
		err := cpu.SetReg(FLAG_CARRY, 0)
		if err != nil {
			return err
		}
		return cpu.SetReg(REG_ACCUM, 0)
	}

	if opCode == 0xF1 {
		cpu.DebugInstr("CLC")
		return cpu.SetReg(FLAG_CARRY, 0)
	}

	if opCode == 0xF2 {
		// IAC
		return cpu.ExecuteAccumulator(opCode, OP_INC)
	}

	if opCode == 0xF3 {
		cpu.DebugInstr("CMC")
		carry, err := cpu.GetReg(FLAG_CARRY)
		if err != nil {
			return err
		}
		if carry == 0 {
			return cpu.SetReg(FLAG_CARRY, 1)
		} else {
			return cpu.SetReg(FLAG_CARRY, 0)
		}
	}

	if opCode == 0xF4 {
		// CMA
		return cpu.ExecuteAccumulator(opCode, OP_CMA)
	}

	if opCode == 0xF5 {
		// RAL
		return cpu.ExecuteRotate(OP_RAL)
	}

	if opCode == 0xF6 {
		// RAR
		return cpu.ExecuteRotate(OP_RAR)
	}

	if opCode == 0xF7 {
		// TCC
		cpu.DebugInstr("TCC")
		carry, err := cpu.GetReg(FLAG_CARRY)
		if err != nil {
			return err
		}
		err = cpu.SetReg(FLAG_CARRY, 0)
		if err != nil {
			return err
		}
		return cpu.SetReg(REG_ACCUM, carry)
	}

	if opCode == 0xF8 {
		// DAC
		return cpu.ExecuteAccumulator(opCode, OP_DEC)
	}

	if opCode == 0xF9 {
		// TCS
		return cpu.ExecuteAccumulator(opCode, OP_TCS)
	}

	if opCode == 0xFA {
		cpu.DebugInstr("STC")
		return cpu.SetReg(FLAG_CARRY, 1)
	}

	if opCode == 0xFB {
		// DAA
		return cpu.ExecuteAccumulator(opCode, OP_DAA)
	}

	if opCode == 0xFC {
		// KBP
		return cpu.ExecuteAccumulator(opCode, OP_KBP)
	}

	if opCode == 0xFD {
		// DCL
		cpu.DebugInstr("DCL")
		acc, err := cpu.GetReg(REG_ACCUM)
		if err != nil {
			return err
		}
		return cpu.SetReg(REG_CL, acc&0x07)
	}

	return &cpusim.ErrInvalidOpcode{Device: cpu, Opcode: opCode}
}

func (cpu *CPU4004) Run() error {
	cpu.Halted = false
	for {
		if cpu.Sim.CtrlC {
			fmt.Println("CPU halted by Ctrl-C")
			return nil
		}
		if cpu.Halted {
			fmt.Println("CPU halted")
			return nil
		}
		if err := cpu.Execute(); err != nil {
			return err
		}
	}
	// never reached
}

func (cpu *CPU4004) String() string {
	s := ""
	for i := 0; i < len(cpu.Registers); i++ {
		if i == 7 {
			continue
		}
		s = s + fmt.Sprintf("%s=%X", cpu.GetRegName(i), cpu.Registers[i])
		if i < len(cpu.Registers)-1 {
			s += " "
		}
	}
	s = s + fmt.Sprintf(" RC=%02x", cpu.RC)
	s = s + fmt.Sprintf(" SP=%x", cpu.SP)
	return s
}
