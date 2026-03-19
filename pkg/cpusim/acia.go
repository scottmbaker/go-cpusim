package cpusim

import (
	"fmt"
	"io"
	"os"
	"sync"
)

// ACIA implements a Motorola MC6850 Asynchronous Communications Interface Adapter.
//
// The 6850 uses two addresses:
//   - Control/Status address: write selects control register, read returns status register
//   - Data address: write transmits data, read receives data
//
// Status register (read):
//   - Bit 0: RDRF - Receive Data Register Full (data available to read)
//   - Bit 1: TDRE - Transmit Data Register Empty (ready to accept data)
//   - Bit 2: DCD  - Data Carrier Detect (active low)
//   - Bit 3: CTS  - Clear To Send (active low)
//   - Bit 4: FE   - Framing Error
//   - Bit 5: OVRN - Receiver Overrun
//   - Bit 6: PE   - Parity Error
//   - Bit 7: IRQ  - Interrupt Request
//
// Control register (write):
//   - Bits 0-1: Counter Divide Select (00=÷1, 01=÷16, 10=÷64, 11=master reset)
//   - Bits 2-4: Word Select (data bits, parity, stop bits)
//   - Bits 5-6: Transmit Control (RTS, TX interrupt enable)
//   - Bit 7: Receive Interrupt Enable
type ACIA struct {
	Sim            *CpuSim
	Serial         SerialIO
	Name           string
	DataAddress    Address
	ControlAddress Address
	Enabler        EnablerInterface
	Keybuffer      []byte
	mu             sync.Mutex
	lastCharOut    byte
	inputEOF       bool
	controlReg     byte
}

func (a *ACIA) GetName() string {
	return a.Name
}

func (a *ACIA) HasAddress(address Address) bool {
	if !a.Enabler.Bool() {
		return false
	}
	return address == a.DataAddress || address == a.ControlAddress
}

func (a *ACIA) Read(address Address) (byte, error) {
	if !a.HasAddress(address) {
		return 0, &ErrInvalidAddress{Address: address}
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.inputEOF && len(a.Keybuffer) == 0 {
		a.Sim.Halt()
	}

	if address == a.DataAddress {
		if len(a.Keybuffer) > 0 {
			value := a.Keybuffer[0]
			a.Keybuffer = a.Keybuffer[1:]
			if value == 0x0A {
				value = 0x0D
			}
			return value, nil
		}
		return 0, nil
	}

	if address == a.ControlAddress {
		// Status register
		var status byte
		status |= 0x02 // TDRE - always ready to transmit
		if len(a.Keybuffer) > 0 {
			status |= 0x01 // RDRF - receive data available
			a.Sim.IOActivity()
		} else {
			a.mu.Unlock()
			a.Sim.IOPoll()
			a.mu.Lock()
		}
		// DCD=0 (asserted/active), CTS=0 (asserted/active), no errors, no IRQ
		return status, nil
	}

	return 0, nil
}

func (a *ACIA) Write(address Address, value byte) error {
	if !a.HasAddress(address) {
		return &ErrInvalidAddress{Address: address}
	}

	if address == a.DataAddress {
		err := a.Serial.WriteByte(value)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to serial: %v\n", err)
		}
		a.lastCharOut = value
		a.Sim.IOActivity()
	}

	if address == a.ControlAddress {
		a.controlReg = value
		// Bits 0-1 == 11 means master reset; we accept but ignore it
	}

	return nil
}

func (a *ACIA) WriteStatus(address Address, statusAddr Address, value byte) error {
	return &ErrNotImplemented{Device: a}
}

func (a *ACIA) ReadStatus(address Address, statusAddr Address) (byte, error) {
	return 0, &ErrNotImplemented{Device: a}
}

func (a *ACIA) Run() error {
	for {
		b, err := a.Serial.ReadByte()
		if err != nil {
			return err
		}
		if b == 0x03 {
			a.Sim.CtrlC = true
		}
		a.mu.Lock()
		a.Keybuffer = append(a.Keybuffer, b)
		a.mu.Unlock()
	}
}

func (a *ACIA) Start(wg *sync.WaitGroup) {
	go func() {
		a.Serial.Start()
		err := a.Run()
		if err != nil {
			if err == io.EOF {
				a.mu.Lock()
				a.inputEOF = true
				a.mu.Unlock()
			} else {
				fmt.Fprintf(os.Stderr, "ACIA error: %v\n", err)
			}
		}
	}()
}

func (a *ACIA) RestoreTerminal() {
	a.Serial.RestoreTerminal()
}

func (a *ACIA) GetKind() string {
	return KIND_ACIA
}

func NewACIA(sim *CpuSim, serial SerialIO, name string, dataAddress, controlAddress Address, enabler EnablerInterface) *ACIA {
	return &ACIA{
		Sim:            sim,
		Serial:         serial,
		Name:           name,
		DataAddress:    dataAddress,
		ControlAddress: controlAddress,
		Enabler:        enabler,
	}
}
