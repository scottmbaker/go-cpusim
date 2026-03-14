package cpu4004

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

func gobEncode(v any) []byte {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(v); err != nil {
		panic("state: gob encode: " + err.Error())
	}
	return buf.Bytes()
}

func gobDecode(data []byte, dst any) error {
	return gob.NewDecoder(bytes.NewReader(data)).Decode(dst)
}

// --- CPU4004 ---

type cpu4004State struct {
	Registers [20]byte
	Stack     [3]uint16
	RC        byte
	SP        byte
	PC        uint16
	Halted    bool
	Cycles    int
}

func (cpu *CPU4004) SaveState() []byte {
	return gobEncode(cpu4004State{
		Registers: cpu.Registers,
		Stack:     cpu.Stack,
		RC:        cpu.RC,
		SP:        cpu.SP,
		PC:        cpu.PC,
		Halted:    cpu.Halted,
		Cycles:    cpu.Cycles,
	})
}

func (cpu *CPU4004) LoadState(data []byte) error {
	var s cpu4004State
	if err := gobDecode(data, &s); err != nil {
		return fmt.Errorf("cpu4004 LoadState: %w", err)
	}
	cpu.Registers = s.Registers
	cpu.Stack = s.Stack
	cpu.RC = s.RC
	cpu.SP = s.SP
	cpu.PC = s.PC
	cpu.Halted = s.Halted
	cpu.Cycles = s.Cycles
	return nil
}

// --- Bus8Bit ---

type bus8BitState struct {
	LastReadValue  byte
	LastWriteValue byte
}

func (b *Bus8Bit) SaveState() []byte {
	return gobEncode(bus8BitState{
		LastReadValue:  b.LastReadValue,
		LastWriteValue: b.LastWriteValue,
	})
}

func (b *Bus8Bit) LoadState(data []byte) error {
	var s bus8BitState
	if err := gobDecode(data, &s); err != nil {
		return fmt.Errorf("bus8bit LoadState: %w", err)
	}
	b.LastReadValue = s.LastReadValue
	b.LastWriteValue = s.LastWriteValue
	return nil
}
