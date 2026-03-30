package cpusim

import (
	"fmt"
	"io"
	"os"
	"sync"
)

// SCC implements a Zilog Z8530 Serial Communications Controller.
//
// The Z8530 SCC has two independent channels (A and B), each with a data port
// and a control port, for a total of four I/O addresses.
//
// Control registers are accessed via a register pointer scheme:
//   - Write to control port: first byte selects the register (WR0-WR7) via
//     bits 0-2; command code 1 ("Point High") in bits 3-5 adds 8 for WR8-WR15.
//     The subsequent byte is the register value.
//   - Read from control port: RR0 is read by default. To read other registers,
//     write WR0 with bits 0-3 selecting the read register, then read.
//
// Read Register 0 (RR0) status bits:
//   - Bit 0: Rx Character Available
//   - Bit 1: Zero Count
//   - Bit 2: Tx Buffer Empty
//   - Bit 3: DCD
//   - Bit 4: Sync/Hunt
//   - Bit 5: CTS
//   - Bit 6: Tx Underrun/EOM
//   - Bit 7: Break/Abort
//
// Read Register 1 (RR1) status bits:
//   - Bit 0: All Sent
//   - Bits 1-2: Residue Code
//   - Bit 4: Parity Error
//   - Bit 5: Rx Overrun Error
//   - Bit 6: CRC/Framing Error
//   - Bit 7: End of Frame (SDLC)
//
// Channel A is the primary channel with keyboard input. Channel B is a
// secondary channel that accepts output but has no input source.
type SCC struct {
	Sim          *CpuSim
	Serial       SerialIO
	Name         string
	DataAddrA    Address
	DataAddrB    Address
	ControlAddrA Address
	ControlAddrB Address
	Enabler      EnablerInterface
	Keybuffer    []byte
	mu           sync.Mutex
	lastCharOut  byte
	inputEOF     bool
	chanA        sccChannel
	chanB        sccChannel
}

// sccChannel holds per-channel state for the SCC.
type sccChannel struct {
	writeRegs [16]byte // WR0-WR15
	readRegs  [16]byte // RR0-RR15
	regPtr    byte     // Next register to read/write (bits 0-2 of WR0, +8 if Point High)
}

func (s *SCC) GetName() string {
	return s.Name
}

func (s *SCC) HasAddress(address Address) bool {
	if !s.Enabler.Bool() {
		return false
	}
	return address == s.DataAddrA || address == s.DataAddrB ||
		address == s.ControlAddrA || address == s.ControlAddrB
}

func (s *SCC) Read(address Address) (byte, error) {
	if !s.HasAddress(address) {
		return 0, &ErrInvalidAddress{Address: address}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.inputEOF && len(s.Keybuffer) == 0 {
		s.Sim.Halt()
	}

	// Data port reads
	if address == s.DataAddrA {
		if len(s.Keybuffer) > 0 {
			value := s.Keybuffer[0]
			s.Keybuffer = s.Keybuffer[1:]
			if value == 0x0A {
				value = 0x0D
			}
			return value, nil
		}
		return 0, nil
	}
	if address == s.DataAddrB {
		return 0, nil
	}

	// Control port reads
	if address == s.ControlAddrA {
		status := s.readControl(&s.chanA, true)
		if status&0x01 != 0 {
			s.Sim.IOActivity()
		} else {
			s.mu.Unlock()
			s.Sim.IOPoll()
			s.mu.Lock()
		}
		return status, nil
	}
	if address == s.ControlAddrB {
		return s.readControl(&s.chanB, false), nil
	}

	return 0, nil
}

// readControl reads the selected read register for a channel.
func (s *SCC) readControl(ch *sccChannel, isChannelA bool) byte {
	reg := ch.regPtr
	ch.regPtr = 0

	switch reg {
	case 0:
		// RR0: status
		var status byte
		if isChannelA && len(s.Keybuffer) > 0 {
			status |= 0x01 // Rx Character Available
		}
		status |= 0x04 // Tx Buffer Empty (always ready)
		return status
	case 1:
		// RR1: All Sent, no errors
		return 0x01
	case 2:
		// RR2: Interrupt vector (channel B returns modified vector)
		return ch.readRegs[2]
	case 3:
		// RR3: Interrupt pending (channel A only)
		return ch.readRegs[3]
	default:
		return ch.readRegs[reg]
	}
}

func (s *SCC) Write(address Address, value byte) error {
	if !s.HasAddress(address) {
		return &ErrInvalidAddress{Address: address}
	}

	// Data port writes
	if address == s.DataAddrA || address == s.DataAddrB {
		err := s.Serial.WriteByte(value)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to serial: %v\n", err)
		}
		s.lastCharOut = value
		s.Sim.IOActivity()
		return nil
	}

	// Control port writes
	if address == s.ControlAddrA {
		s.writeControl(&s.chanA, value)
		return nil
	}
	if address == s.ControlAddrB {
		s.writeControl(&s.chanB, value)
		return nil
	}

	return nil
}

// writeControl handles writes to the control port using the register pointer scheme.
func (s *SCC) writeControl(ch *sccChannel, value byte) {
	reg := ch.regPtr
	ch.regPtr = 0

	if reg == 0 {
		// Writing to WR0 - bits 0-2 set the register pointer for the next access
		ch.writeRegs[0] = value
		ch.regPtr = value & 0x07

		// Handle WR0 command bits (bits 3-5)
		cmd := (value >> 3) & 0x07
		switch cmd {
		case 1:
			// Point High - add 8 to register pointer for accessing WR8-WR15/RR8-RR15
			ch.regPtr |= 0x08
		case 2:
			// Reset external/status interrupts
		case 3:
			// Channel reset
			ch.regPtr = 0
			for i := range ch.writeRegs {
				ch.writeRegs[i] = 0
			}
		case 5:
			// Reset Tx interrupt pending
		case 6:
			// Error reset
		case 7:
			// Reset highest IUS (channel A only)
		}
	} else {
		ch.writeRegs[reg] = value
	}
}

func (s *SCC) WriteStatus(address Address, statusAddr Address, value byte) error {
	return &ErrNotImplemented{Device: s}
}

func (s *SCC) ReadStatus(address Address, statusAddr Address) (byte, error) {
	return 0, &ErrNotImplemented{Device: s}
}

func (s *SCC) Run() error {
	for {
		b, err := s.Serial.ReadByte()
		if err != nil {
			return err
		}
		if b == 0x03 {
			s.Sim.CtrlC.Store(true)
		}
		s.mu.Lock()
		s.Keybuffer = append(s.Keybuffer, b)
		s.mu.Unlock()
	}
}

func (s *SCC) Start(wg *sync.WaitGroup) {
	go func() {
		s.Serial.Start()
		err := s.Run()
		if err != nil {
			if err == io.EOF {
				s.mu.Lock()
				s.inputEOF = true
				s.mu.Unlock()
			} else {
				fmt.Fprintf(os.Stderr, "SCC error: %v\n", err)
			}
		}
	}()
}

func (s *SCC) RestoreTerminal() {
	s.Serial.RestoreTerminal()
}

func (s *SCC) GetKind() string {
	return KIND_SCC
}

func NewSCC(sim *CpuSim, serial SerialIO, name string, dataAddrA, dataAddrB, controlAddrA, controlAddrB Address, enabler EnablerInterface) *SCC {
	return &SCC{
		Sim:          sim,
		Serial:       serial,
		Name:         name,
		DataAddrA:    dataAddrA,
		DataAddrB:    dataAddrB,
		ControlAddrA: controlAddrA,
		ControlAddrB: controlAddrB,
		Enabler:      enabler,
	}
}
