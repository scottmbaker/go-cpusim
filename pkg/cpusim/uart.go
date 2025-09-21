package cpusim

import (
	"fmt"
	"github.com/scottmbaker/gocpusim/pkg/rawmode"
	"os"
	"sync"
)

type UART struct {
	Sim                 *CpuSim
	Name                string
	DataReadAddress     Address
	DataWriteAddress    Address
	ControlReadAddress  Address
	ControlWriteAddress Address
	Enabler             EnablerInterface
	Keybuffer           []byte // Simulated key buffer for UART
	RawMode             bool   // Whether to run in raw mode
	lastCharOut         byte   // Last key pressed, for debugging or other purposes
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
		return byte(value & 0xFF), nil
	}

	return 0, nil
}

func (u *UART) Write(address Address, value byte) error {
	if !u.HasAddress(address) {
		return &ErrInvalidAddress{Address: address}
	}

	if address == u.DataWriteAddress {
		//if value == 0x0A && u.lastCharOut != 0x0D {
		//	os.Stdout.Write([]byte{0x0D}) // Add CR before LF
		//}
		_, err := os.Stdout.Write([]byte{value})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to stdout: %v", err)
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
		input := make([]byte, 1)
		_, err := os.Stdin.Read(input)
		if err != nil {
			return err
		}
		if input[0] == 0x03 {
			u.Sim.CtrlC = true // Handle Ctrl-C
		}
		u.Keybuffer = append(u.Keybuffer, input[0])
	}
}

func (u *UART) Start(wg *sync.WaitGroup) {
	//wg.Add(1)   do not add ourselves to wg. We want this to die when the cpus die.
	go func(c *UART) {
		//defer wg.Done()
		if u.RawMode {
			err := rawmode.EnableRawMode() // Use the raw mode enabling function
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error setting terminal to raw mode: %v\n", err)
			}
		}
		err := u.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "UART error: %v\n", err)
		}
	}(u)
}

func (u *UART) RestoreTerminal() {
	err := rawmode.DisableRawMode() // Use the raw mode disabling function
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error restoring terminal mode: %v\n", err)
	}
}

func NewUART(sim *CpuSim, name string, dataReadAddress, dataWriteAddress, controlReadAddress, controlWriteAddress Address, enabler EnablerInterface) *UART {
	return &UART{
		Sim:                 sim,
		Name:                name,
		DataReadAddress:     dataReadAddress,
		DataWriteAddress:    dataWriteAddress,
		ControlReadAddress:  controlReadAddress,
		ControlWriteAddress: controlWriteAddress,
		Enabler:             enabler,
		RawMode:             true,
	}
}
