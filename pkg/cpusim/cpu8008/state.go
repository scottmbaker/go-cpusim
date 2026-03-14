package cpu8008

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

type cpu8008State struct {
	Registers [12]byte
	Stack     [8]uint16
	SP        byte
	PC        uint16
	Halted    bool
}

func (cpu *CPU8008) SaveState() []byte {
	return gobEncode(cpu8008State{
		Registers: cpu.Registers,
		Stack:     cpu.Stack,
		SP:        cpu.SP,
		PC:        cpu.PC,
		Halted:    cpu.Halted,
	})
}

func (cpu *CPU8008) LoadState(data []byte) error {
	var s cpu8008State
	if err := gobDecode(data, &s); err != nil {
		return fmt.Errorf("cpu8008 LoadState: %w", err)
	}
	cpu.Registers = s.Registers
	cpu.Stack = s.Stack
	cpu.SP = s.SP
	cpu.PC = s.PC
	cpu.Halted = s.Halted
	return nil
}
