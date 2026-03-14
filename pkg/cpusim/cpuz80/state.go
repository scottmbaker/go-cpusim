package cpuz80

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

type cpuz80State struct {
	A, F byte
	B, C byte
	D, E byte
	H, L byte
	AF_  uint16
	BC_  uint16
	DE_  uint16
	HL_  uint16
	IX   uint16
	IY   uint16
	SP   uint16
	PC   uint16
	I    byte
	R    byte
	IFF1 bool
	IFF2 bool
	IM   byte

	EIPending bool
	Halted    bool
	WZ        uint16
	Q         byte
	PrevQ     byte
}

func (cpu *CPUZ80) SaveState() []byte {
	return gobEncode(cpuz80State{
		A: cpu.A, F: cpu.F,
		B: cpu.B, C: cpu.C,
		D: cpu.D, E: cpu.E,
		H: cpu.H, L: cpu.L,
		AF_: cpu.AF_, BC_: cpu.BC_, DE_: cpu.DE_, HL_: cpu.HL_,
		IX: cpu.IX, IY: cpu.IY,
		SP: cpu.SP, PC: cpu.PC,
		I: cpu.I, R: cpu.R,
		IFF1: cpu.IFF1, IFF2: cpu.IFF2, IM: cpu.IM,
		EIPending: cpu.EIPending, Halted: cpu.Halted,
		WZ: cpu.WZ, Q: cpu.Q, PrevQ: cpu.PrevQ,
	})
}

func (cpu *CPUZ80) LoadState(data []byte) error {
	var s cpuz80State
	if err := gobDecode(data, &s); err != nil {
		return fmt.Errorf("cpuz80 LoadState: %w", err)
	}
	cpu.A = s.A
	cpu.F = s.F
	cpu.B = s.B
	cpu.C = s.C
	cpu.D = s.D
	cpu.E = s.E
	cpu.H = s.H
	cpu.L = s.L
	cpu.AF_ = s.AF_
	cpu.BC_ = s.BC_
	cpu.DE_ = s.DE_
	cpu.HL_ = s.HL_
	cpu.IX = s.IX
	cpu.IY = s.IY
	cpu.SP = s.SP
	cpu.PC = s.PC
	cpu.I = s.I
	cpu.R = s.R
	cpu.IFF1 = s.IFF1
	cpu.IFF2 = s.IFF2
	cpu.IM = s.IM
	cpu.EIPending = s.EIPending
	cpu.Halted = s.Halted
	cpu.WZ = s.WZ
	cpu.Q = s.Q
	cpu.PrevQ = s.PrevQ
	return nil
}
