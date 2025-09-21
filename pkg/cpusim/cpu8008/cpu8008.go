package cpu8008

import (
	"fmt"
	"github.com/scottmbaker/gocpusim/pkg/cpusim"
)

type CPU8008 struct {
	Sim       *cpusim.CpuSim // Reference to the CPU simulation
	Name      string         // Name of the CPU
	Registers [12]byte       // A, B, C, D, E, H, L, MEM-placholder, CF, ZF, SF, PF
	Stack     [8]uint16
	SP        byte
	PC        uint16 // Program Counter
	Halted    bool   // Flag to indicate if the CPU is halted
	NewStyle  bool   // Flag to indicate if the new style debugging is used
}

const (
	REG_A = 0
	REG_B = 1
	REG_C = 2
	REG_D = 3
	REG_E = 4
	REG_H = 5
	REG_L = 6
	REG_M = 7 // for LrM and LMr

	FLAG_CARRY  = 8
	FLAG_ZERO   = 9
	FLAG_SIGN   = 10
	FLAG_PARITY = 11

	OP_ADD = 0
	OP_AC  = 1
	OP_SUB = 2
	OP_SC  = 3
	OP_AND = 4
	OP_XOR = 5
	OP_OR  = 6
	OP_CP  = 7

	OP_RLC = 0
	OP_RRC = 1
	OP_RAL = 2
	OP_RAR = 3
)

func New8008(sim *cpusim.CpuSim, name string) *CPU8008 {
	return &CPU8008{
		Name:      name,
		Registers: [12]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		Stack:     [8]uint16{0, 0, 0, 0, 0, 0, 0, 0},
		SP:        0,
		PC:        0, // Program Counter
		Sim:       sim,
		NewStyle:  true, // Use new style for debuggingt
	}
}

func toBit(value bool) byte {
	if value {
		return 1
	}
	return 0
}

func OperationFromOpcode(opcode byte) int {
	return int(opcode>>3) & 0x07
}

func RotFromOpcode(opcode byte) int {
	return int(opcode>>3) & 0x07
}

func FlagFromOpcode(opcode byte) int {
	switch (opcode >> 3) & 0x03 {
	case 0:
		return FLAG_CARRY
	case 1:
		return FLAG_ZERO
	case 2:
		return FLAG_SIGN
	default:
		return FLAG_PARITY
	}
}

func FlagStr(flag int) string {
	switch flag {
	case FLAG_CARRY:
		return "C"
	case FLAG_ZERO:
		return "Z"
	case FLAG_SIGN:
		return "S"
	case FLAG_PARITY:
		return "P"
	default:
		return "?"
	}
}

func (cpu *CPU8008) GetName() string {
	return cpu.Name
}

func (cpu *CPU8008) GetRegName(reg int) string {
	switch reg {
	case REG_A:
		return "A"
	case REG_B:
		return "B"
	case REG_C:
		return "C"
	case REG_D:
		return "D"
	case REG_E:
		return "E"
	case REG_H:
		return "H"
	case REG_L:
		return "L"
	case REG_M:
		return "M"
	case FLAG_CARRY:
		return "CF"
	case FLAG_ZERO:
		return "ZF"
	case FLAG_SIGN:
		return "SF"
	case FLAG_PARITY:
		return "PF"
	default:
		return fmt.Sprintf("R%d", reg)
	}
}

func (cpu *CPU8008) GetReg(register int) (byte, error) {
	if register < 0 || register >= len(cpu.Registers) {
		return 0, &cpusim.ErrInvalidRegister{Device: cpu, Register: register}
	}
	if register == REG_M {
		return cpu.Sim.ReadMemory((cpusim.Address(cpu.Registers[REG_H]) << 8) | cpusim.Address(cpu.Registers[REG_L]))
	}
	return cpu.Registers[register], nil
}

func (cpu *CPU8008) SetReg(register int, value byte) error {
	if register < 0 || register >= len(cpu.Registers) {
		return &cpusim.ErrInvalidRegister{Register: register}
	}
	if register == REG_M {
		return cpu.Sim.WriteMemory((cpusim.Address(cpu.Registers[REG_H])<<8)|cpusim.Address(cpu.Registers[REG_L]), value)
	}
	cpu.Registers[register] = value
	return nil
}

func (cpu *CPU8008) ExecuteMove(opCode byte) error {
	destReg := int(opCode>>3) & 0x7
	srcReg := int(opCode & 0x7)
	srcVal, err := cpu.GetReg(srcReg)
	if err != nil {
		return err
	}
	cpu.DebugMove(destReg, srcReg)
	return cpu.SetReg(destReg, srcVal)
}

func (cpu *CPU8008) ExecuteLoadImmediate(opCode byte, value byte) error {
	destReg := int(opCode>>3) & 0x7
	cpu.DebugMoveImmediate(destReg, value)
	return cpu.SetReg(destReg, value)
}

func (cpu *CPU8008) FetchOpcode() (byte, error) {
	opCode, err := cpu.Sim.ReadMemory(cpusim.Address(cpu.PC))
	if err != nil {
		return 0, err
	}
	cpu.PC++
	return opCode, nil
}

func (cpu *CPU8008) FetchAddr() (uint16, error) {
	addrLow, err := cpu.Sim.ReadMemory(cpusim.Address(cpu.PC))
	if err != nil {
		return 0, err
	}
	cpu.PC++
	addrHigh, err := cpu.Sim.ReadMemory(cpusim.Address(cpu.PC))
	if err != nil {
		return 0, err
	}
	cpu.PC++
	return (uint16(addrHigh)<<8 | uint16(addrLow)) & 0x3FFF, nil
}

func (cpu *CPU8008) ExecuteIncDec(opCode byte, increment int) error {
	var work int

	reg := int(opCode>>3) & 0x07
	if (reg == REG_M) || (reg == REG_A) {
		return &cpusim.ErrInvalidRegister{Register: reg}
	}
	value, err := cpu.GetReg(reg)
	if err != nil {
		return err
	}

	work = int(value)
	work += int(increment)
	err = cpu.UpdateIncFlags(work)
	if err != nil {
		return err
	}
	value = byte(work & 0xFF)

	cpu.DebugIncDec(reg, increment)

	return cpu.SetReg(reg, value)
}

func (cpu *CPU8008) SetParity(value int) error {
	value = value & 0xFF
	// Parity is set if the number of set bits is even
	count := 0
	for i := 0; i < 8; i++ {
		if (value & (1 << i)) != 0 {
			count++
		}
	}
	return cpu.SetReg(FLAG_PARITY, toBit((count%2) == 0))
}

func (cpu *CPU8008) updateArithFlags(work int) error {
	err := cpu.SetReg(FLAG_CARRY, toBit((work&0x100) != 0))
	if err != nil {
		return err
	}
	err = cpu.UpdateIncFlags(work)
	return err
}

func (cpu *CPU8008) UpdateIncFlags(work int) error {
	// everything but carry
	err := cpu.SetReg(FLAG_ZERO, toBit((work&0xFF) == 0))
	if err != nil {
		return err
	}
	err = cpu.SetReg(FLAG_SIGN, toBit((work&0x80) != 0))
	if err != nil {
		return err
	}
	err = cpu.SetParity(work)
	return err
}

func (cpu *CPU8008) UpdateLogicalFlags(work int) error {
	err := cpu.SetReg(FLAG_CARRY, 0) // Logical operations always reset carry
	if err != nil {
		return err
	}
	err = cpu.SetReg(FLAG_ZERO, toBit((work&0xFF) == 0))
	if err != nil {
		return err
	}
	err = cpu.SetReg(FLAG_SIGN, toBit((work&0x80) != 0))
	return err
}

func (cpu *CPU8008) ExecuteAccumulator(opCode byte, op int, isImmed bool, val byte) error {
	var work int

	acc, err := cpu.GetReg(REG_A)
	if err != nil {
		return err
	}

	if !isImmed {
		val, err = cpu.GetReg(int(opCode & 0x07))
		if err != nil {
			return err
		}
	}

	work = int(acc)

	switch op {
	case OP_ADD, OP_AC: // Add
		work = work + int(val)
		carryBit, _ := cpu.GetReg(FLAG_CARRY)
		if (op == OP_AC) && (carryBit != 0) {
			work = work + 1
		}
	case OP_SUB, OP_SC: // Subtract
		work = work - int(val)
		carryBit, _ := cpu.GetReg(FLAG_CARRY)
		if (op == OP_SC) && (carryBit != 0) {
			work = work - 1
		}
	case OP_AND:
		work = work & int(val)
	case OP_XOR:
		work = work ^ int(val)
	case OP_OR:
		work = work | int(val)
	case OP_CP: // Compare
		work = work - int(val)
	}

	if (op == OP_ADD) || (op == OP_AC) || (op == OP_SUB) || (op == OP_SC) || (op == OP_CP) {
		err = cpu.updateArithFlags(work)
		if err != nil {
			return err
		}
	} else {
		err = cpu.UpdateLogicalFlags(work)
		if err != nil {
			return err
		}
	}

	acc = byte(work & 0xFF)
	cpu.DebugAccumulator(op, int(opCode&0x07), isImmed, val)

	if op == OP_CP {
		// do not set the accumulator
		return nil
	}

	return cpu.SetReg(REG_A, acc)
}

func (cpu *CPU8008) ExecuteRotate(op int) error {
	acc, err := cpu.GetReg(REG_A)
	if err != nil {
		return err
	}

	switch op {
	case OP_RLC: // Rotate Left through Carry
		leftBit := (acc & 0x80) >> 7
		acc = (acc << 1) | leftBit
		err = cpu.SetReg(FLAG_CARRY, leftBit)
		if err != nil {
			return err
		}
	case OP_RRC: // Rotate Right through Carry
		rightBit := acc & 0x01
		acc = (acc >> 1) | (rightBit << 7)
		err = cpu.SetReg(FLAG_CARRY, rightBit)
		if err != nil {
			return err
		}
	case OP_RAL:
		carryBit, _ := cpu.GetReg(FLAG_CARRY)
		leftBit := (acc & 0x80) >> 7
		acc = (acc << 1) | carryBit
		err = cpu.SetReg(FLAG_CARRY, leftBit)
		if err != nil {
			return err
		}
	case OP_RAR:
		carryBit, _ := cpu.GetReg(FLAG_CARRY)
		rightBit := acc & 0x01
		acc = (acc >> 1) | (carryBit << 7)
		err = cpu.SetReg(FLAG_CARRY, rightBit)
		if err != nil {
			return err
		}
	}
	cpu.DebugRotate(op)

	return cpu.SetReg(REG_A, acc)
}

func (cpu *CPU8008) PushStack(value uint16) {
	cpu.Stack[cpu.SP] = cpu.PC
	cpu.SP++
	if cpu.SP >= byte(len(cpu.Stack)) { // XXXX possible bug here
		cpu.SP = 0 // wrap around stack pointer
	}
}

func (cpu *CPU8008) ExecuteJump(addr uint16, conditional bool, flag int, istrue bool, iscall bool) error {
	if conditional {
		flagValue, err := cpu.GetReg(flag)
		if err != nil {
			return err
		}
		if (flagValue != 0) != istrue {
			cpu.DebugJump(conditional, flag, istrue, iscall, addr)
			return nil // Jump not taken
		}
	}

	if iscall {
		cpu.PushStack(cpu.PC)
	}

	cpu.PC = addr

	cpu.DebugJump(conditional, flag, istrue, iscall, addr)

	return nil
}

func (cpu *CPU8008) ExecuteRet(conditional bool, flag int, istrue bool) error {
	if conditional {
		flagValue, err := cpu.GetReg(flag)
		if err != nil {
			return err
		}
		if (flagValue != 0) != istrue {
			cpu.DebugRet(conditional, flag, istrue)
			return nil // Return not taken
		}
	}

	if cpu.SP == 0 {
		return fmt.Errorf("stack underflow")
	}
	if cpu.SP == 0 {
		cpu.SP = byte(len(cpu.Stack) - 1)
	} else {
		cpu.SP--
	}
	cpu.PC = cpu.Stack[cpu.SP]

	cpu.DebugRet(conditional, flag, istrue)

	return nil
}

func (cpu *CPU8008) ExecuteRes(n byte) error {
	addr := uint16(n) << 3
	cpu.PushStack(cpu.PC)
	cpu.PC = addr
	cpu.DebugRes(n)
	return nil
}

func (cpu *CPU8008) ExecutePort(port byte) error {
	if (port & 0x18) == 0 {
		val, err := cpu.Sim.ReadPort(cpusim.Address(port))
		if err != nil {
			return err
		}
		err = cpu.SetReg(REG_A, val)
		if err != nil {
			return err
		}
		cpu.DebugPort(port, true, val)
	} else {
		val, err := cpu.GetReg(REG_A)
		if err != nil {
			return err
		}
		err = cpu.Sim.WritePort(cpusim.Address(port), val)
		if err != nil {
			return err
		}
		cpu.DebugPort(port, false, val)
	}

	return nil
}

func (cpu *CPU8008) Execute() error {
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

	if opCode == 0xFF || opCode == 0x00 || opCode == 0x01 {
		// make sure to check HALT before other operations because
		// it overlaps some other opcodes
		cpu.Halted = true
		cpu.DebugInstr("HALT")
		return nil
	}
	if opCode&0xC0 == 0xC0 {
		return cpu.ExecuteMove(opCode)
	}
	if opCode&0xC7 == 0x06 {
		val, err := cpu.FetchOpcode()
		if err != nil {
			return err
		}
		return cpu.ExecuteLoadImmediate(opCode, val)
	}
	if opCode&0xC7 == 0 {
		return cpu.ExecuteIncDec(opCode, 1)
	}
	if opCode&0xC7 == 1 {
		return cpu.ExecuteIncDec(opCode, -1)
	}
	if (opCode&0xE0 == 0x80) || (opCode&0xE0 == 0xA0) {
		return cpu.ExecuteAccumulator(opCode, OperationFromOpcode(opCode), false, 0)
	}
	if (opCode&0xE7 == 0x04) || (opCode&0xE7 == 0x24) {
		val, err := cpu.FetchOpcode()
		if err != nil {
			return err
		}
		return cpu.ExecuteAccumulator(opCode, OperationFromOpcode(opCode), true, val)
	}
	if opCode&0xE7 == 0x02 { // RLC, RRC, RAL, RAR
		return cpu.ExecuteRotate(RotFromOpcode(opCode))
	}
	if opCode&0xC7 == 0x44 { // JMP
		addr, err := cpu.FetchAddr()
		if err != nil {
			return err
		}
		return cpu.ExecuteJump(addr, false, 0, false, false)
	}
	if opCode&0xC7 == 0x40 { // JFc, JTc
		addr, err := cpu.FetchAddr()
		if err != nil {
			return err
		}
		return cpu.ExecuteJump(addr, true, FlagFromOpcode(opCode), (opCode&0x20) != 0, false)
	}
	if opCode&0xC7 == 0x46 { // CALL
		addr, err := cpu.FetchAddr()
		if err != nil {
			return err
		}
		return cpu.ExecuteJump(addr, false, 0, false, true)
	}
	if opCode&0xC7 == 0x42 { // CFc, CTc
		addr, err := cpu.FetchAddr()
		if err != nil {
			return err
		}
		return cpu.ExecuteJump(addr, true, FlagFromOpcode(opCode), (opCode&0x20) != 0, true)
	}
	if opCode&0xC7 == 0x07 { // RET
		return cpu.ExecuteRet(false, 0, false)
	}
	if opCode&0xC7 == 0x03 { // RFc, RTc
		return cpu.ExecuteRet(true, FlagFromOpcode(opCode), (opCode&0x20) != 0)
	}
	if opCode&0xC7 == 0x05 { // RST
		return cpu.ExecuteRes(opCode >> 3 & 0x07)
	}
	if opCode&0xC1 == 0x41 { // IN
		return cpu.ExecutePort(opCode >> 1 & 0x1F)
	}
	return &cpusim.ErrInvalidOpcode{Device: cpu, Opcode: opCode}
}

func (cpu *CPU8008) Run() error {
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

func (cpu *CPU8008) String() string {
	s := ""
	for i := 0; i < len(cpu.Registers); i++ {
		if i == 7 {
			continue
		}
		s = s + fmt.Sprintf("%s=%02X", cpu.GetRegName(i), cpu.Registers[i])
		if i < len(cpu.Registers)-1 {
			s += " "
		}
	}
	return s
}
