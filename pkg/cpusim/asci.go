package cpusim

import (
	"fmt"
	"io"
	"os"
	"sync"
)

// ASCI implements a Z180 Asynchronous Serial Communication Interface.
//
// The Z180 has two ASCI channels (ASCI0 and ASCI1), with the following
// I/O registers:
//
//	Base+0x00: CNTLA0 - Control Register A, Channel 0
//	Base+0x01: CNTLA1 - Control Register A, Channel 1
//	Base+0x02: CNTLB0 - Control Register B, Channel 0
//	Base+0x03: CNTLB1 - Control Register B, Channel 1
//	Base+0x04: STAT0  - Status Register, Channel 0
//	Base+0x05: STAT1  - Status Register, Channel 1
//	Base+0x06: TDR0   - Transmit Data Register, Channel 0
//	Base+0x07: TDR1   - Transmit Data Register, Channel 1
//	Base+0x08: RDR0   - Receive Data Register, Channel 0
//	Base+0x09: RDR1   - Receive Data Register, Channel 1
//
// Status register bits (STAT0/STAT1):
//
//	Bit 7: RDRF - Receive Data Register Full
//	Bit 6: OVRN - Overrun Error
//	Bit 5: PE   - Parity Error
//	Bit 4: FE   - Framing Error
//	Bit 3: RIE  - Receive Interrupt Enable (read/write)
//	Bit 2: DCD0 - Data Carrier Detect (active low)
//	Bit 1: TDRE - Transmit Data Register Empty
//	Bit 0: TIE  - Transmit Interrupt Enable (read/write)
//
// Control Register A bits (CNTLA0/CNTLA1):
//
//	Bit 7: MPE   - Multi-Processor Enable
//	Bit 6: RE    - Receive Enable
//	Bit 5: TE    - Transmit Enable
//	Bit 4: RTS0  - Request To Send (channel 0 only)
//	Bit 3: MPBR/EFR - Multi-Processor Bit Receive / Error Flag Reset
//	Bit 2: Mode  - 0=Synchronous, 1=Asynchronous
//	Bits 1-0: Data bits (00=5, 01=6, 10=7, 11=8)
//
// Channel 0 is the primary channel with keyboard input. Channel 1
// accepts output but has no input source.
type ASCI struct {
	Sim       *CpuSim
	Serial    SerialIO
	Name      string
	BaseAddr  Address
	Enabler   EnablerInterface
	Keybuffer   []byte
	mu          sync.Mutex
	lastCharOut byte
	inputEOF    bool
	cntlA     [2]byte // CNTLA0, CNTLA1
	cntlB     [2]byte // CNTLB0, CNTLB1
	stat      [2]byte // STAT0, STAT1
}

func (a *ASCI) GetName() string {
	return a.Name
}

func (a *ASCI) HasAddress(address Address) bool {
	if !a.Enabler.Bool() {
		return false
	}
	offset := address - a.BaseAddr
	return offset <= 0x09
}

func (a *ASCI) Read(address Address) (byte, error) {
	if !a.HasAddress(address) {
		return 0, &ErrInvalidAddress{Address: address}
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.inputEOF && len(a.Keybuffer) == 0 {
		a.Sim.Halt()
	}

	offset := address - a.BaseAddr

	switch offset {
	case 0x00, 0x01: // CNTLA0, CNTLA1
		ch := offset
		return a.cntlA[ch], nil

	case 0x02, 0x03: // CNTLB0, CNTLB1
		ch := offset - 0x02
		return a.cntlB[ch], nil

	case 0x04, 0x05: // STAT0, STAT1
		ch := offset - 0x04
		var status byte
		status |= 0x02 // TDRE - always ready to transmit
		// Preserve RIE and TIE bits from software writes
		status |= a.stat[ch] & 0x09
		if ch == 0 && len(a.Keybuffer) > 0 {
			status |= 0x80 // RDRF - receive data available
			a.Sim.IOActivity()
		} else if ch == 0 {
			a.mu.Unlock()
			a.Sim.IOPoll()
			a.mu.Lock()
		}
		return status, nil

	case 0x06, 0x07: // TDR0, TDR1 (write-only, reads return 0)
		return 0, nil

	case 0x08: // RDR0 - Receive Data, Channel 0
		if len(a.Keybuffer) > 0 {
			value := a.Keybuffer[0]
			a.Keybuffer = a.Keybuffer[1:]
			if value == 0x0A {
				value = 0x0D
			}
			return value, nil
		}
		return 0, nil

	case 0x09: // RDR1 - Receive Data, Channel 1 (no input source)
		return 0, nil
	}

	return 0, nil
}

func (a *ASCI) Write(address Address, value byte) error {
	if !a.HasAddress(address) {
		return &ErrInvalidAddress{Address: address}
	}

	offset := address - a.BaseAddr

	switch offset {
	case 0x00, 0x01: // CNTLA0, CNTLA1
		ch := offset
		a.cntlA[ch] = value

	case 0x02, 0x03: // CNTLB0, CNTLB1
		ch := offset - 0x02
		a.cntlB[ch] = value

	case 0x04, 0x05: // STAT0, STAT1
		ch := offset - 0x04
		// Only RIE (bit 3) and TIE (bit 0) are writable
		a.stat[ch] = (a.stat[ch] & 0xF6) | (value & 0x09)

	case 0x06, 0x07: // TDR0, TDR1
		err := a.Serial.WriteByte(value)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to serial: %v\n", err)
		}
		a.lastCharOut = value
		a.Sim.IOActivity()

	case 0x08, 0x09: // RDR0, RDR1 (read-only, writes ignored)
	}

	return nil
}

func (a *ASCI) WriteStatus(address Address, statusAddr Address, value byte) error {
	return &ErrNotImplemented{Device: a}
}

func (a *ASCI) ReadStatus(address Address, statusAddr Address) (byte, error) {
	return 0, &ErrNotImplemented{Device: a}
}

func (a *ASCI) Run() error {
	for {
		b, err := a.Serial.ReadByte()
		if err != nil {
			return err
		}
		if b == 0x03 {
			a.Sim.CtrlC.Store(true)
		}
		a.mu.Lock()
		a.Keybuffer = append(a.Keybuffer, b)
		a.mu.Unlock()
	}
}

func (a *ASCI) Start(wg *sync.WaitGroup) {
	go func() {
		a.Serial.Start()
		err := a.Run()
		if err != nil {
			if err == io.EOF {
				a.mu.Lock()
				a.inputEOF = true
				a.mu.Unlock()
			} else {
				fmt.Fprintf(os.Stderr, "ASCI error: %v\n", err)
			}
		}
	}()
}

func (a *ASCI) RestoreTerminal() {
	a.Serial.RestoreTerminal()
}

func (a *ASCI) GetKind() string {
	return KIND_ASCI
}

func NewASCI(sim *CpuSim, serial SerialIO, name string, baseAddr Address, enabler EnablerInterface) *ASCI {
	return &ASCI{
		Sim:      sim,
		Serial:   serial,
		Name:     name,
		BaseAddr: baseAddr,
		Enabler:  enabler,
	}
}
