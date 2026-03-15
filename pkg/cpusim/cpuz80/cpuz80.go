package cpuz80

import (
	"fmt"

	"github.com/scottmbaker/gocpusim/pkg/cpusim"
)

const (
	RegA = iota
	RegF
	RegB
	RegC
	RegD
	RegE
	RegH
	RegL
	RegAF_
	RegBC_
	RegDE_
	RegHL_
	RegIXH
	RegIXL
	RegIYH
	RegIYL
	RegSPH
	RegSPL
	RegI
	RegR
	RegIM
	RegIFF1
	RegIFF2
	RegPCH
	RegPCL
)

// Flag bit positions in the F register
const (
	FlagC  = 0 // Carry
	FlagN  = 1 // Subtract
	FlagPV = 2 // Parity/Overflow
	FlagX  = 3 // Undocumented (bit 3 of result)
	FlagH  = 4 // Half-carry
	FlagY  = 5 // Undocumented (bit 5 of result)
	FlagZ  = 6 // Zero
	FlagS  = 7 // Sign
)

const (
	MaskC  = 1 << FlagC
	MaskN  = 1 << FlagN
	MaskPV = 1 << FlagPV
	MaskX  = 1 << FlagX
	MaskH  = 1 << FlagH
	MaskY  = 1 << FlagY
	MaskZ  = 1 << FlagZ
	MaskS  = 1 << FlagS
)

// CPUZ80 implements the Zilog Z80 CPU
type CPUZ80 struct {
	Sim  *cpusim.CpuSim
	Name string

	// Main registers
	A, F byte
	B, C byte
	D, E byte
	H, L byte

	// Alternate register set
	AF_ uint16
	BC_ uint16
	DE_ uint16
	HL_ uint16

	// Index registers
	IX, IY uint16

	// Special registers
	SP uint16
	PC uint16
	I  byte
	R  byte

	// Interrupt state
	IFF1, IFF2 bool
	IM         byte
	EIPending  bool // EI delays one instruction

	// Internal
	Halted bool
	WZ     uint16 // Internal MEMPTR register

	// Q flag tracking for undocumented SCF/CCF behavior
	Q     byte
	PrevQ byte

	PortAddressMask uint16 // Mask to apply to port addresses (e.g., 0xFF for 8-bit ports)
}

// Parity lookup table: true if even number of bits set
var parityTable [256]bool

func init() {
	for i := 0; i < 256; i++ {
		bits := 0
		v := i
		for v != 0 {
			bits += v & 1
			v >>= 1
		}
		parityTable[i] = (bits % 2) == 0
	}
}

func NewZ80(sim *cpusim.CpuSim, name string) *CPUZ80 {
	return &CPUZ80{
		Sim:             sim,
		Name:            name,
		PortAddressMask: 0xFFFF, // Default to 16-bit port addresses
	}
}

func (cpu *CPUZ80) GetName() string {
	return cpu.Name
}

func (cpu *CPUZ80) Halt() {
	cpu.Halted = true
}

func (cpu *CPUZ80) SetReg(register int, value byte) error {
	switch register {
	case RegA:
		cpu.A = value
	case RegF:
		cpu.F = value
	case RegB:
		cpu.B = value
	case RegC:
		cpu.C = value
	case RegD:
		cpu.D = value
	case RegE:
		cpu.E = value
	case RegH:
		cpu.H = value
	case RegL:
		cpu.L = value
	case RegI:
		cpu.I = value
	case RegR:
		cpu.R = value
	case RegIM:
		cpu.IM = value
	case RegIFF1:
		cpu.IFF1 = value != 0
	case RegIFF2:
		cpu.IFF2 = value != 0
	default:
		return &cpusim.ErrInvalidRegister{Device: cpu, Register: register}
	}
	return nil
}

func (cpu *CPUZ80) GetReg(register int) (byte, error) {
	switch register {
	case RegA:
		return cpu.A, nil
	case RegF:
		return cpu.F, nil
	case RegB:
		return cpu.B, nil
	case RegC:
		return cpu.C, nil
	case RegD:
		return cpu.D, nil
	case RegE:
		return cpu.E, nil
	case RegH:
		return cpu.H, nil
	case RegL:
		return cpu.L, nil
	case RegI:
		return cpu.I, nil
	case RegR:
		return cpu.R, nil
	case RegIM:
		return cpu.IM, nil
	case RegIFF1:
		if cpu.IFF1 {
			return 1, nil
		}
		return 0, nil
	case RegIFF2:
		if cpu.IFF2 {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, &cpusim.ErrInvalidRegister{Device: cpu, Register: register}
	}
}

func (cpu *CPUZ80) String() string {
	return fmt.Sprintf("A=%02X F=%02X B=%02X C=%02X D=%02X E=%02X H=%02X L=%02X SP=%04X PC=%04X IX=%04X IY=%04X I=%02X R=%02X",
		cpu.A, cpu.F, cpu.B, cpu.C, cpu.D, cpu.E, cpu.H, cpu.L, cpu.SP, cpu.PC, cpu.IX, cpu.IY, cpu.I, cpu.R)
}

func (cpu *CPUZ80) Run() error {
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
		cpu.Sim.Throttle.Tick()
	}
}

func (cpu *CPUZ80) readByte(addr uint16) byte {
	val, _ := cpu.Sim.ReadMemory(cpusim.Address(addr))
	return val
}

func (cpu *CPUZ80) writeByte(addr uint16, val byte) {
	_ = cpu.Sim.WriteMemory(cpusim.Address(addr), val)
}

func (cpu *CPUZ80) readWord(addr uint16) uint16 {
	lo := cpu.readByte(addr)
	hi := cpu.readByte(addr + 1)
	return uint16(hi)<<8 | uint16(lo)
}

func (cpu *CPUZ80) writeWord(addr uint16, val uint16) {
	cpu.writeByte(addr, byte(val&0xFF))
	cpu.writeByte(addr+1, byte(val>>8))
}

func (cpu *CPUZ80) readPort(addr uint16) byte {
	val, _ := cpu.Sim.ReadPort(cpusim.Address(addr))
	return val
}

func (cpu *CPUZ80) writePort(addr uint16, val byte) {
	_ = cpu.Sim.WritePort(cpusim.Address(addr), val)
}

func (cpu *CPUZ80) fetchByte() byte {
	val := cpu.readByte(cpu.PC)
	cpu.PC++
	return val
}

func (cpu *CPUZ80) fetchWord() uint16 {
	lo := cpu.fetchByte()
	hi := cpu.fetchByte()
	return uint16(hi)<<8 | uint16(lo)
}

func (cpu *CPUZ80) push(val uint16) {
	cpu.SP--
	cpu.writeByte(cpu.SP, byte(val>>8))
	cpu.SP--
	cpu.writeByte(cpu.SP, byte(val&0xFF))
}

func (cpu *CPUZ80) pop() uint16 {
	lo := cpu.readByte(cpu.SP)
	cpu.SP++
	hi := cpu.readByte(cpu.SP)
	cpu.SP++
	return uint16(hi)<<8 | uint16(lo)
}

func (cpu *CPUZ80) getAF() uint16 {
	return uint16(cpu.A)<<8 | uint16(cpu.F)
}

func (cpu *CPUZ80) setAF(val uint16) {
	cpu.A = byte(val >> 8)
	cpu.F = byte(val & 0xFF)
}

func (cpu *CPUZ80) getBC() uint16 {
	return uint16(cpu.B)<<8 | uint16(cpu.C)
}

func (cpu *CPUZ80) setBC(val uint16) {
	cpu.B = byte(val >> 8)
	cpu.C = byte(val & 0xFF)
}

func (cpu *CPUZ80) getDE() uint16 {
	return uint16(cpu.D)<<8 | uint16(cpu.E)
}

func (cpu *CPUZ80) setDE(val uint16) {
	cpu.D = byte(val >> 8)
	cpu.E = byte(val & 0xFF)
}

func (cpu *CPUZ80) getHL() uint16 {
	return uint16(cpu.H)<<8 | uint16(cpu.L)
}

func (cpu *CPUZ80) setHL(val uint16) {
	cpu.H = byte(val >> 8)
	cpu.L = byte(val & 0xFF)
}

func (cpu *CPUZ80) getReg8(r byte) byte {
	switch r {
	case 0:
		return cpu.B
	case 1:
		return cpu.C
	case 2:
		return cpu.D
	case 3:
		return cpu.E
	case 4:
		return cpu.H
	case 5:
		return cpu.L
	case 6:
		return cpu.readByte(cpu.getHL())
	case 7:
		return cpu.A
	}
	return 0
}

func (cpu *CPUZ80) setReg8(r byte, val byte) {
	switch r {
	case 0:
		cpu.B = val
	case 1:
		cpu.C = val
	case 2:
		cpu.D = val
	case 3:
		cpu.E = val
	case 4:
		cpu.H = val
	case 5:
		cpu.L = val
	case 6:
		cpu.writeByte(cpu.getHL(), val)
	case 7:
		cpu.A = val
	}
}

func (cpu *CPUZ80) getReg8Idx(r byte, idx uint16) byte {
	switch r {
	case 4:
		return byte(idx >> 8)
	case 5:
		return byte(idx & 0xFF)
	case 6:
		return cpu.readByte(cpu.getHL())
	default:
		return cpu.getReg8(r)
	}
}

func (cpu *CPUZ80) setReg8Idx(r byte, val byte, idx *uint16) {
	switch r {
	case 4:
		*idx = (*idx & 0x00FF) | (uint16(val) << 8)
	case 5:
		*idx = (*idx & 0xFF00) | uint16(val)
	default:
		cpu.setReg8(r, val)
	}
}

func (cpu *CPUZ80) getReg16(pp byte) uint16 {
	switch pp {
	case 0:
		return cpu.getBC()
	case 1:
		return cpu.getDE()
	case 2:
		return cpu.getHL()
	case 3:
		return cpu.SP
	}
	return 0
}

func (cpu *CPUZ80) setReg16(pp byte, val uint16) {
	switch pp {
	case 0:
		cpu.setBC(val)
	case 1:
		cpu.setDE(val)
	case 2:
		cpu.setHL(val)
	case 3:
		cpu.SP = val
	}
}

func (cpu *CPUZ80) getReg16AF(pp byte) uint16 {
	switch pp {
	case 0:
		return cpu.getBC()
	case 1:
		return cpu.getDE()
	case 2:
		return cpu.getHL()
	case 3:
		return cpu.getAF()
	}
	return 0
}

func (cpu *CPUZ80) setReg16AF(pp byte, val uint16) {
	switch pp {
	case 0:
		cpu.setBC(val)
	case 1:
		cpu.setDE(val)
	case 2:
		cpu.setHL(val)
	case 3:
		cpu.setAF(val)
	}
}

func (cpu *CPUZ80) condition(cc byte) bool {
	switch cc {
	case 0: // NZ
		return cpu.F&MaskZ == 0
	case 1: // Z
		return cpu.F&MaskZ != 0
	case 2: // NC
		return cpu.F&MaskC == 0
	case 3: // C
		return cpu.F&MaskC != 0
	case 4: // PO
		return cpu.F&MaskPV == 0
	case 5: // PE
		return cpu.F&MaskPV != 0
	case 6: // P (positive)
		return cpu.F&MaskS == 0
	case 7: // M (minus)
		return cpu.F&MaskS != 0
	}
	return false
}

func (cpu *CPUZ80) szFlags(val byte) byte {
	var f byte
	if val == 0 {
		f |= MaskZ
	}
	f |= val & (MaskS | MaskY | MaskX)
	return f
}

// incR increments the lower 7 bits of R, preserving bit 7
func (cpu *CPUZ80) incR() {
	cpu.R = (cpu.R & 0x80) | ((cpu.R + 1) & 0x7F)
}

func (cpu *CPUZ80) Execute() error {
	cpu.PrevQ = cpu.Q
	cpu.Q = 0

	cpu.incR()

	opcode := cpu.fetchByte()

	if cpu.Sim.Debug {
		fmt.Printf("%04X: [%02X] %s\n", cpu.PC-1, opcode, cpu.String())
	}

	// Clear EI pending before execution; the EI instruction sets it fresh
	if cpu.EIPending {
		cpu.EIPending = false
	}

	return cpu.executeUnprefixed(opcode)
}
