package cpusim

import (
	"fmt"
	"io"
	"os"
	"sync"
)

// UART16550 implements a 16550 UART.
//
// The 16550 uses 8 consecutive I/O addresses (base through base+7):
//
//	Offset 0: RBR (Receive Buffer Register, read)  / THR (Transmit Holding Register, write)
//	          DLL (Divisor Latch Low)               when LCR DLAB bit is set
//	Offset 1: IER (Interrupt Enable Register)
//	          DLH (Divisor Latch High)              when LCR DLAB bit is set
//	Offset 2: IIR (Interrupt Identification Register, read) / FCR (FIFO Control Register, write)
//	Offset 3: LCR (Line Control Register)
//	Offset 4: MCR (Modem Control Register)
//	Offset 5: LSR (Line Status Register, read-only)
//	Offset 6: MSR (Modem Status Register, read-only)
//	Offset 7: SCR (Scratch Register)
//
// LSR bits used in simulation:
//   - Bit 0: DR   - Data Ready (receive data available)
//   - Bit 5: THRE - Transmit Holding Register Empty (always set)
//   - Bit 6: TEMT - Transmitter Empty (always set)
//
// Access to DLL/DLH vs RBR/THR/IER is controlled by the DLAB bit (bit 7) of LCR.
type UART16550 struct {
	Sim         *CpuSim
	Serial      SerialIO
	Name        string
	BaseAddress Address
	Enabler     EnablerInterface
	Keybuffer   []byte
	mu          sync.Mutex
	lastCharOut byte
	inputEOF    bool

	// Registers
	ier byte // Interrupt Enable Register
	lcr byte // Line Control Register (bit 7 = DLAB)
	mcr byte // Modem Control Register
	scr byte // Scratch Register
	dll byte // Divisor Latch Low
	dlh byte // Divisor Latch High
	fcr byte // FIFO Control Register
}

const (
	lcr16550DLAB = 0x80 // Divisor Latch Access Bit
	lsr16550DR   = 0x01 // Data Ready
	lsr16550THRE = 0x20 // Transmit Holding Register Empty
	lsr16550TEMT = 0x40 // Transmitter Empty
)

func (u *UART16550) GetName() string {
	return u.Name
}

func (u *UART16550) HasAddress(address Address) bool {
	if !u.Enabler.Bool() {
		return false
	}
	return address >= u.BaseAddress && address < u.BaseAddress+8
}

func (u *UART16550) Read(address Address) (byte, error) {
	if !u.HasAddress(address) {
		return 0, &ErrInvalidAddress{Address: address}
	}

	u.mu.Lock()
	defer u.mu.Unlock()

	if u.inputEOF && len(u.Keybuffer) == 0 {
		u.Sim.Halt()
	}

	offset := address - u.BaseAddress
	switch offset {
	case 0:
		if u.lcr&lcr16550DLAB != 0 {
			return u.dll, nil
		}
		// RBR: read a byte from the receive buffer
		if len(u.Keybuffer) > 0 {
			value := u.Keybuffer[0]
			u.Keybuffer = u.Keybuffer[1:]
			if value == 0x0A {
				value = 0x0D
			}
			return value, nil
		}
		return 0, nil
	case 1:
		if u.lcr&lcr16550DLAB != 0 {
			return u.dlh, nil
		}
		return u.ier, nil
	case 2:
		// IIR: no interrupts pending
		return 0x01, nil
	case 3:
		return u.lcr, nil
	case 4:
		return u.mcr, nil
	case 5:
		// LSR: THRE and TEMT are always set; DR set when data available
		var lsr byte = lsr16550THRE | lsr16550TEMT
		if len(u.Keybuffer) > 0 {
			lsr |= lsr16550DR
			u.Sim.IOActivity()
		} else {
			u.mu.Unlock()
			u.Sim.IOPoll()
			u.mu.Lock()
		}
		return lsr, nil
	case 6:
		// MSR: CTS, DSR, DCD asserted (active high in register)
		return 0xB0, nil
	case 7:
		return u.scr, nil
	}

	return 0, nil
}

func (u *UART16550) Write(address Address, value byte) error {
	if !u.HasAddress(address) {
		return &ErrInvalidAddress{Address: address}
	}

	offset := address - u.BaseAddress
	switch offset {
	case 0:
		if u.lcr&lcr16550DLAB != 0 {
			u.dll = value
			return nil
		}
		// THR: transmit a byte
		err := u.Serial.WriteByte(value)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to serial: %v\n", err)
		}
		u.lastCharOut = value
		u.Sim.IOActivity()
	case 1:
		if u.lcr&lcr16550DLAB != 0 {
			u.dlh = value
			return nil
		}
		u.ier = value
	case 2:
		u.fcr = value
	case 3:
		u.lcr = value
	case 4:
		u.mcr = value
	case 7:
		u.scr = value
	}

	return nil
}

func (u *UART16550) WriteStatus(address Address, statusAddr Address, value byte) error {
	return &ErrNotImplemented{Device: u}
}

func (u *UART16550) ReadStatus(address Address, statusAddr Address) (byte, error) {
	return 0, &ErrNotImplemented{Device: u}
}

func (u *UART16550) Run() error {
	for {
		b, err := u.Serial.ReadByte()
		if err != nil {
			return err
		}
		if b == 0x03 {
			u.Sim.CtrlC.Store(true)
		}
		u.mu.Lock()
		u.Keybuffer = append(u.Keybuffer, b)
		u.mu.Unlock()
	}
}

func (u *UART16550) Start(wg *sync.WaitGroup) {
	go func() {
		u.Serial.Start()
		err := u.Run()
		if err != nil {
			if err == io.EOF {
				u.mu.Lock()
				u.inputEOF = true
				u.mu.Unlock()
			} else {
				fmt.Fprintf(os.Stderr, "16550 error: %v\n", err)
			}
		}
	}()
}

func (u *UART16550) RestoreTerminal() {
	u.Serial.RestoreTerminal()
}

func (u *UART16550) GetKind() string {
	return KIND_16550
}

// NewUART16550 creates a new 16550 UART. The device occupies 8 consecutive I/O
// addresses starting at baseAddress.
func NewUART16550(sim *CpuSim, serial SerialIO, name string, baseAddress Address, enabler EnablerInterface) *UART16550 {
	return &UART16550{
		Sim:         sim,
		Serial:      serial,
		Name:        name,
		BaseAddress: baseAddress,
		Enabler:     enabler,
	}
}
