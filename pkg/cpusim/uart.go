package cpusim

import (
	"fmt"
	"os"
	"sync"
)

// UART implements an Intel 8251 USART.
// Status register bits: bit 0 = TxReady (always set), bit 1 = RxReady.
// Data and control/status use separate configurable addresses.
type UART struct {
	Sim                 *CpuSim
	Serial              SerialIO
	Name                string
	DataReadAddress     Address
	DataWriteAddress    Address
	ControlReadAddress  Address
	ControlWriteAddress Address
	Enabler             EnablerInterface
	Keybuffer           []byte
	lastCharOut         byte
	exitEof             bool
}

func (u *UART) LoadInputFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	u.Keybuffer = append(u.Keybuffer, data...)
	return nil
}

func (u *UART) SetExitOnEof(exitEof bool) {
	u.exitEof = exitEof
}

func (u *UART) GetName() string {
	return u.Name
}

func (u *UART) HasAddress(address Address) bool {
	if !u.Enabler.Bool() {
		return false
	}
	return (address == u.DataReadAddress || address == u.DataWriteAddress ||
		address == u.ControlReadAddress || address == u.ControlWriteAddress)
}

func (u *UART) Read(address Address) (byte, error) {
	if !u.HasAddress(address) {
		return 0, &ErrInvalidAddress{Address: address}
	}

	if u.exitEof && len(u.Keybuffer) == 0 {
		u.Sim.Halt()
	}

	if address == u.DataReadAddress {
		if len(u.Keybuffer) > 0 {
			value := u.Keybuffer[0]
			u.Keybuffer = u.Keybuffer[1:]
			if value == 0x0A {
				value = 0x0D
			}
			return value, nil
		}
	}

	if address == u.ControlReadAddress {
		value := 0x01 // TXReady is always true
		if len(u.Keybuffer) > 0 {
			value |= 0x02 // RXReady is true if there are bytes in the buffer
		}
		return byte(value), nil
	}

	return 0, nil
}

func (u *UART) Write(address Address, value byte) error {
	if !u.HasAddress(address) {
		return &ErrInvalidAddress{Address: address}
	}

	if address == u.DataWriteAddress {
		err := u.Serial.WriteByte(value)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to serial: %v", err)
		}
		u.lastCharOut = value
	}

	return nil
}

func (u *UART) WriteStatus(address Address, statusAddr Address, value byte) error {
	_ = address
	_ = statusAddr
	_ = value
	return &ErrNotImplemented{Device: u}
}

func (u *UART) ReadStatus(address Address, statusAddr Address) (byte, error) {
	_ = address
	_ = statusAddr
	return 0, &ErrNotImplemented{Device: u}
}

func (u *UART) Run() error {
	for {
		b, err := u.Serial.ReadByte()
		if err != nil {
			return err
		}
		if b == 0x03 {
			u.Sim.CtrlC = true
		}
		u.Keybuffer = append(u.Keybuffer, b)
	}
}

func (u *UART) Start(wg *sync.WaitGroup) {
	go func() {
		u.Serial.Start()
		err := u.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "UART error: %v\n", err)
		}
	}()
}

func (u *UART) RestoreTerminal() {
	u.Serial.RestoreTerminal()
}

func (u *UART) GetKind() string {
	return KIND_UART
}

func NewUART(sim *CpuSim, serial SerialIO, name string, dataReadAddress, dataWriteAddress, controlReadAddress, controlWriteAddress Address, enabler EnablerInterface) *UART {
	return &UART{
		Sim:                 sim,
		Serial:              serial,
		Name:                name,
		DataReadAddress:     dataReadAddress,
		DataWriteAddress:    dataWriteAddress,
		ControlReadAddress:  controlReadAddress,
		ControlWriteAddress: controlWriteAddress,
		Enabler:             enabler,
	}
}
