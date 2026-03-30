package cpusim

import (
	"fmt"
	"os"
)

// WD37C65 / NEC uPD765 Floppy Disk Controller emulation.
//
// Default geometry is a standard 1.44 MB 3.5" HD floppy:
//   80 cylinders, 2 heads, 18 sectors/track, 512 bytes/sector
//
// Port map (directly addressed, no base offset):
//   0x48: DCR  - Configuration Control Register (write only)
//   0x50: MSR  - Main Status Register (read only)
//   0x51: DATA - Data Register (read/write)
//   0x58: DOR  - Digital Output Register (write only)

const (
	// Default geometry: 1.44 MB 3.5" HD floppy
	fdcDefaultCylinders      = 80
	fdcDefaultHeads          = 2
	fdcDefaultSectorsPerTrack = 18
	fdcDefaultSectorSize     = 512
	fdcDefaultSectorSizeCode = 2 // 128 << 2 = 512

	// Main Status Register bits
	fdcMsrD0B = 0x01 // Drive 0 busy (seeking)
	fdcMsrD1B = 0x02 // Drive 1 busy
	fdcMsrD2B = 0x04 // Drive 2 busy
	fdcMsrD3B = 0x08 // Drive 3 busy
	fdcMsrCB  = 0x10 // FDC busy (command in progress)
	fdcMsrEXM = 0x20 // Execution mode (in data transfer phase)
	fdcMsrDIO = 0x40 // Data direction: 1 = FDC->CPU, 0 = CPU->FDC
	fdcMsrRQM = 0x80 // Request for Master: 1 = data register ready

	// Status Register 0
	fdcSt0IC0 = 0x40 // Interrupt Code bit 0 (normal termination = 0x00)
	fdcSt0IC1 = 0x80 // Interrupt Code bit 1 (abnormal = 0x40, invalid = 0x80, drive not ready = 0xC0)
	fdcSt0SE  = 0x20 // Seek End
	fdcSt0EC  = 0x10 // Equipment Check
	fdcSt0NR  = 0x08 // Not Ready
	fdcSt0HD  = 0x04 // Head Address

	// Status Register 1
	fdcSt1EN = 0x80 // End of Cylinder
	fdcSt1DE = 0x20 // Data Error (CRC)
	fdcSt1OR = 0x10 // Overrun
	fdcSt1ND = 0x04 // No Data (sector not found)
	fdcSt1NW = 0x02 // Not Writable (write protect)
	fdcSt1MA = 0x01 // Missing Address Mark

	// Status Register 2
	fdcSt2CM = 0x40 // Control Mark (deleted data)
	fdcSt2DD = 0x20 // Data Error in Data Field
	fdcSt2WC = 0x10 // Wrong Cylinder
	fdcSt2BC = 0x02 // Bad Cylinder
	fdcSt2MD = 0x01 // Missing Data Address Mark

	// Status Register 3
	fdcSt3WP = 0x40 // Write Protected
	fdcSt3RY = 0x20 // Ready
	fdcSt3T0 = 0x10 // Track 0
	fdcSt3TS = 0x08 // Two Side
	fdcSt3HD = 0x04 // Head Address

	// Digital Output Register bits
	fdcDorDSEL  = 0x03 // Drive select mask
	fdcDorRESET = 0x04 // Reset (active low: 0 = reset, 1 = normal)
	fdcDorDMA   = 0x08 // DMA enable
	fdcDorMOT0  = 0x10 // Motor on drive 0
	fdcDorMOT1  = 0x20 // Motor on drive 1
	fdcDorMOT2  = 0x40 // Motor on drive 2
	fdcDorMOT3  = 0x80 // Motor on drive 3

	// FDC commands (low nibble, some bits are parameter flags)
	fdcCmdReadData       = 0x06 // MT/MF/SK + 0x06
	fdcCmdReadDeleted    = 0x0C // MT/MF/SK + 0x0C
	fdcCmdWriteData      = 0x05 // MT/MF + 0x05
	fdcCmdWriteDeleted   = 0x09 // MT/MF + 0x09
	fdcCmdReadTrack      = 0x02 // MF + 0x02
	fdcCmdReadID         = 0x0A // MF + 0x0A
	fdcCmdFormatTrack    = 0x0D // MF + 0x0D
	fdcCmdScanEqual      = 0x11 // MT/MF/SK + 0x11
	fdcCmdScanLowOrEqual = 0x19 // MT/MF/SK + 0x19
	fdcCmdScanHighOrEq   = 0x1D // MT/MF/SK + 0x1D
	fdcCmdRecalibrate    = 0x07
	fdcCmdSenseInterrupt = 0x08
	fdcCmdSpecify        = 0x03
	fdcCmdSenseDriveStatus = 0x04
	fdcCmdSeek           = 0x0F

	// FDC phases
	fdcPhaseIdle    = 0
	fdcPhaseCommand = 1
	fdcPhaseExec    = 2
	fdcPhaseResult  = 3

)

// FDC emulates a WD37C65 / NEC uPD765 floppy disk controller.
type FDC struct {
	Sim     *CpuSim
	Name    string
	Enabler EnablerInterface

	// Port addresses
	PortMSR  Address // Main Status Register (read)
	PortData Address // Data Register (read/write)
	PortDOR  Address // Digital Output Register (write)
	PortDCR  Address // Configuration Control Register (write)

	// Disk geometry
	Cylinders      int
	Heads          int
	SectorsPerTrack int
	SectorSize     int
	SectorSizeCode byte

	// Disk images (up to 4 drives)
	files [4]*os.File

	// Controller state
	phase     int
	msr       byte
	dor       byte
	dcr       byte
	st0       byte
	st1       byte
	st2       byte
	st3       byte

	// Command buffer
	cmdBuf  [16]byte
	cmdLen  int
	cmdPos  int

	// Result buffer
	resBuf  [16]byte
	resLen  int
	resPos  int

	// Data buffer for sector transfers
	dataBuf  [16384]byte // big enough for largest sector size
	dataLen  int
	dataPos  int
	dataDir  int // 0 = CPU->FDC (write), 1 = FDC->CPU (read)

	// Per-drive state
	pcn        [4]byte // Present Cylinder Number for each drive
	pendingSt0 [4]byte // pending ST0 for sense interrupt (0 = no pending)

	// Command parameters (parsed from command bytes)
	cmdCode   byte
	multiTrack bool
	mfm       bool
	skipDeleted bool
}

// NewFDC creates a new WD37C65 floppy disk controller with default 1.44MB geometry.
func NewFDC(sim *CpuSim, name string, portMSR, portData, portDOR, portDCR Address, enabler EnablerInterface) *FDC {
	fdc := &FDC{
		Sim:             sim,
		Name:            name,
		Enabler:         enabler,
		PortMSR:         portMSR,
		PortData:        portData,
		PortDOR:         portDOR,
		PortDCR:         portDCR,
		Cylinders:       fdcDefaultCylinders,
		Heads:           fdcDefaultHeads,
		SectorsPerTrack: fdcDefaultSectorsPerTrack,
		SectorSize:      fdcDefaultSectorSize,
		SectorSizeCode:  fdcDefaultSectorSizeCode,
	}
	fdc.dor = fdcDorRESET // start with reset de-asserted so FDC is usable
	fdc.reset()
	return fdc
}

// AttachImage opens a disk image file for the specified drive (0-3).
func (fdc *FDC) AttachImage(drive int, filename string) error {
	if drive < 0 || drive > 3 {
		return fmt.Errorf("fdc: invalid drive number %d", drive)
	}
	f, err := os.OpenFile(filename, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("fdc: open %s: %w", filename, err)
	}
	if fdc.files[drive] != nil {
		fdc.files[drive].Close()
	}
	fdc.files[drive] = f
	return nil
}

func (fdc *FDC) GetName() string {
	return fdc.Name
}

func (fdc *FDC) GetKind() string {
	return KIND_FDC
}

func (fdc *FDC) HasAddress(address Address) bool {
	if !fdc.Enabler.Bool() {
		return false
	}
	return address == fdc.PortMSR || address == fdc.PortData ||
		address == fdc.PortDOR || address == fdc.PortDCR
}

func (fdc *FDC) Read(address Address) (byte, error) {
	switch address {
	case fdc.PortMSR:
		return fdc.readMSR(), nil
	case fdc.PortData:
		return fdc.readData(), nil
	default:
		return 0xFF, nil
	}
}

func (fdc *FDC) Write(address Address, value byte) error {
	switch address {
	case fdc.PortData:
		fdc.writeData(value)
	case fdc.PortDOR:
		fdc.writeDOR(value)
	case fdc.PortDCR:
		fdc.dcr = value
	}
	return nil
}

func (fdc *FDC) ReadStatus(address Address, statusAddr Address) (byte, error) {
	return 0, &ErrNotImplemented{Device: fdc}
}

func (fdc *FDC) WriteStatus(address Address, statusAddr Address, value byte) error {
	return &ErrNotImplemented{Device: fdc}
}

func (fdc *FDC) Close() {
	for i := range fdc.files {
		if fdc.files[i] != nil {
			fdc.files[i].Close()
			fdc.files[i] = nil
		}
	}
}

// reset reinitializes the controller to its power-on state.
func (fdc *FDC) reset() {
	fdc.phase = fdcPhaseIdle
	fdc.msr = fdcMsrRQM // ready to accept commands
	fdc.st0 = 0
	fdc.st1 = 0
	fdc.st2 = 0
	fdc.cmdPos = 0
	fdc.cmdLen = 0
	fdc.resPos = 0
	fdc.resLen = 0
	fdc.dataPos = 0
	fdc.dataLen = 0
	for i := range fdc.pcn {
		fdc.pcn[i] = 0
	}
	for i := range fdc.pendingSt0 {
		fdc.pendingSt0[i] = 0
	}
}

func (fdc *FDC) readMSR() byte {
	return fdc.msr
}

func (fdc *FDC) readData() byte {
	switch fdc.phase {
	case fdcPhaseExec:
		if fdc.dataDir == 1 { // FDC->CPU read
			if fdc.dataPos < fdc.dataLen {
				b := fdc.dataBuf[fdc.dataPos]
				fdc.dataPos++
				if fdc.dataPos >= fdc.dataLen {
					// Sector transfer complete, advance to next sector or result phase
					fdc.execReadNext()
				}
				return b
			}
		}
		return 0xFF

	case fdcPhaseResult:
		if fdc.resPos < fdc.resLen {
			b := fdc.resBuf[fdc.resPos]
			fdc.resPos++
			if fdc.resPos >= fdc.resLen {
				fdc.enterIdle()
			}
			return b
		}
		return 0xFF
	}

	return 0xFF
}

func (fdc *FDC) writeData(value byte) {
	switch fdc.phase {
	case fdcPhaseIdle:
		// First byte of a new command
		fdc.cmdBuf[0] = value
		fdc.cmdPos = 1
		fdc.cmdLen = fdc.commandLength(value)
		if fdc.cmdLen <= 1 {
			// Command is complete with just the command byte
			fdc.executeCommand()
		} else {
			fdc.phase = fdcPhaseCommand
			fdc.msr = fdcMsrRQM | fdcMsrCB // ready for parameter bytes
		}

	case fdcPhaseCommand:
		if fdc.cmdPos < len(fdc.cmdBuf) {
			fdc.cmdBuf[fdc.cmdPos] = value
			fdc.cmdPos++
		}
		if fdc.cmdPos >= fdc.cmdLen {
			fdc.executeCommand()
		}

	case fdcPhaseExec:
		if fdc.dataDir == 0 { // CPU->FDC write
			if fdc.dataPos < fdc.dataLen {
				fdc.dataBuf[fdc.dataPos] = value
				fdc.dataPos++
				if fdc.dataPos >= fdc.dataLen {
					if fdc.cmdCode == fdcCmdFormatTrack {
						fdc.execFormatComplete()
					} else {
						fdc.execWriteNext()
					}
				}
			}
		}
	}
}

func (fdc *FDC) writeDOR(value byte) {
	oldDOR := fdc.dor
	fdc.dor = value

	// Check for reset transition (bit 2: 0 = reset, 1 = normal)
	wasReset := oldDOR&fdcDorRESET == 0
	isReset := value&fdcDorRESET == 0

	if isReset {
		fdc.reset()
		fdc.msr = 0 // during reset, RQM is cleared
	} else if wasReset && !isReset {
		// Coming out of reset
		fdc.reset()
		// Generate interrupt for each drive (sense interrupt should be issued 4 times)
		// IC=11 (0xC0) = ready line changed during polling after reset
		for i := 0; i < 4; i++ {
			fdc.pendingSt0[i] = 0xC0 | byte(i)
		}
	}
}

// commandLength returns the expected total number of bytes (including the command byte)
// for the given command.
func (fdc *FDC) commandLength(cmd byte) int {
	switch cmd & 0x1F {
	case fdcCmdReadData, fdcCmdReadDeleted:
		return 9
	case fdcCmdWriteData, fdcCmdWriteDeleted:
		return 9
	case fdcCmdReadTrack:
		return 9
	case fdcCmdReadID:
		return 2
	case fdcCmdFormatTrack:
		return 6
	case fdcCmdScanEqual, fdcCmdScanLowOrEqual, fdcCmdScanHighOrEq:
		return 9
	case fdcCmdRecalibrate:
		return 2
	case fdcCmdSenseInterrupt:
		return 1
	case fdcCmdSpecify:
		return 3
	case fdcCmdSenseDriveStatus:
		return 2
	case fdcCmdSeek:
		return 3
	// Note: Version (0x10) is NOT supported by WD37C65 (only uPD765B/82077AA).
	// It falls through to default and returns as an invalid command.
	default:
		return 1 // invalid command, will return error in result phase
	}
}

func (fdc *FDC) executeCommand() {
	cmd := fdc.cmdBuf[0]
	fdc.cmdCode = cmd & 0x1F
	fdc.multiTrack = cmd&0x80 != 0
	fdc.mfm = cmd&0x40 != 0
	fdc.skipDeleted = cmd&0x20 != 0

	switch fdc.cmdCode {
	case fdcCmdReadData:
		fdc.execReadData()
	case fdcCmdWriteData:
		fdc.execWriteData()
	case fdcCmdReadDeleted:
		fdc.execReadData() // treat same as read for emulation
	case fdcCmdWriteDeleted:
		fdc.execWriteData() // treat same as write for emulation
	case fdcCmdReadID:
		fdc.execReadID()
	case fdcCmdFormatTrack:
		fdc.execFormatTrack()
	case fdcCmdRecalibrate:
		fdc.execRecalibrate()
	case fdcCmdSenseInterrupt:
		fdc.execSenseInterrupt()
	case fdcCmdSpecify:
		fdc.execSpecify()
	case fdcCmdSenseDriveStatus:
		fdc.execSenseDriveStatus()
	case fdcCmdSeek:
		fdc.execSeek()
	case fdcCmdReadTrack:
		fdc.execReadData() // simplified: treat like read data
	case fdcCmdScanEqual, fdcCmdScanLowOrEqual, fdcCmdScanHighOrEq:
		// Scan commands: simplified, just return normal completion
		fdc.setupSt012(fdc.cmdBuf[1]&0x03, fdc.cmdBuf[1]&0x04 != 0)
		fdc.st2 |= 0x08 // Scan Equal Hit (SH)
		fdc.enterResultPhaseRW()
	default:
		// Invalid command
		fdc.st0 = 0x80 // Invalid command
		fdc.resBuf[0] = fdc.st0
		fdc.resLen = 1
		fdc.resPos = 0
		fdc.phase = fdcPhaseResult
		fdc.msr = fdcMsrRQM | fdcMsrDIO | fdcMsrCB
	}
}

// setupSt012 initializes ST0 with the drive/head info and clears ST1/ST2.
func (fdc *FDC) setupSt012(drive byte, head bool) {
	fdc.st0 = drive & 0x03
	if head {
		fdc.st0 |= fdcSt0HD
	}
	fdc.st1 = 0
	fdc.st2 = 0
}

// --- Read Data command ---

// execReadData starts the Read Data command.
// Command bytes: [cmd, HD/DS, C, H, R, N, EOT, GPL, DTL]
func (fdc *FDC) execReadData() {
	drive := fdc.cmdBuf[1] & 0x03
	head := fdc.cmdBuf[1] & 0x04
	fdc.setupSt012(drive, head != 0)

	if fdc.files[drive] == nil {
		// No disk in drive
		fdc.st0 |= fdcSt0IC0 | fdcSt0NR // abnormal termination, not ready
		fdc.st1 |= fdcSt1MA
		fdc.enterResultPhaseRW()
		return
	}

	// Read first sector
	if !fdc.readCurrentSector() {
		fdc.enterResultPhaseRW()
		return
	}

	fdc.dataDir = 1 // FDC->CPU
	fdc.phase = fdcPhaseExec
	fdc.msr = fdcMsrRQM | fdcMsrDIO | fdcMsrEXM | fdcMsrCB
}

// readCurrentSector reads the sector specified by the command parameters into dataBuf.
func (fdc *FDC) readCurrentSector() bool {
	drive := fdc.cmdBuf[1] & 0x03
	c := fdc.cmdBuf[2]     // Cylinder
	h := fdc.cmdBuf[3]     // Head
	r := fdc.cmdBuf[4]     // Sector (1-based)

	offset := fdc.sectorOffset(c, h, r)
	if offset < 0 {
		fdc.st0 |= fdcSt0IC0 // abnormal termination
		fdc.st1 |= fdcSt1ND  // no data
		return false
	}

	n, err := fdc.files[drive].ReadAt(fdc.dataBuf[:fdc.SectorSize], offset)
	if err != nil || n < fdc.SectorSize {
		fdc.st0 |= fdcSt0IC0 // abnormal termination
		fdc.st1 |= fdcSt1DE  // data error
		return false
	}

	fdc.dataPos = 0
	fdc.dataLen = fdc.SectorSize
	return true
}

// execReadNext is called when a sector transfer completes during a read.
func (fdc *FDC) execReadNext() {
	eot := fdc.cmdBuf[6] // End of Track (last sector number)

	// Advance sector number
	fdc.cmdBuf[4]++

	if fdc.cmdBuf[4] > eot {
		// Past end of track
		if fdc.multiTrack && fdc.cmdBuf[3] == 0 {
			// Switch to head 1
			fdc.cmdBuf[3] = 1
			fdc.cmdBuf[1] |= 0x04 // set head bit
			fdc.cmdBuf[4] = 1     // start from sector 1
			fdc.st0 |= fdcSt0HD   // update ST0 head address
		} else {
			// Done - normal termination
			fdc.st1 |= fdcSt1EN // end of cylinder
			fdc.enterResultPhaseRW()
			return
		}
	}

	// Read next sector
	if !fdc.readCurrentSector() {
		fdc.enterResultPhaseRW()
		return
	}
	// Stay in exec phase, MSR already set for read
}

// --- Write Data command ---

// execWriteData starts the Write Data command.
func (fdc *FDC) execWriteData() {
	drive := fdc.cmdBuf[1] & 0x03
	head := fdc.cmdBuf[1] & 0x04
	fdc.setupSt012(drive, head != 0)

	if fdc.files[drive] == nil {
		fdc.st0 |= fdcSt0IC0 | fdcSt0NR
		fdc.st1 |= fdcSt1MA
		fdc.enterResultPhaseRW()
		return
	}

	fdc.dataDir = 0 // CPU->FDC
	fdc.dataPos = 0
	fdc.dataLen = fdc.SectorSize
	fdc.phase = fdcPhaseExec
	fdc.msr = fdcMsrRQM | fdcMsrEXM | fdcMsrCB // DIO=0 for write
}

// writeCurrentSector writes dataBuf to the sector specified by command parameters.
func (fdc *FDC) writeCurrentSector() bool {
	drive := fdc.cmdBuf[1] & 0x03
	c := fdc.cmdBuf[2]
	h := fdc.cmdBuf[3]
	r := fdc.cmdBuf[4]

	offset := fdc.sectorOffset(c, h, r)
	if offset < 0 {
		fdc.st0 |= fdcSt0IC0
		fdc.st1 |= fdcSt1ND
		return false
	}

	_, err := fdc.files[drive].WriteAt(fdc.dataBuf[:fdc.SectorSize], offset)
	if err != nil {
		fdc.st0 |= fdcSt0IC0
		fdc.st1 |= fdcSt1DE
		return false
	}

	return true
}

// execWriteNext is called when a sector buffer is full during a write.
func (fdc *FDC) execWriteNext() {
	if !fdc.writeCurrentSector() {
		fdc.enterResultPhaseRW()
		return
	}

	eot := fdc.cmdBuf[6]

	// Advance sector
	fdc.cmdBuf[4]++

	if fdc.cmdBuf[4] > eot {
		if fdc.multiTrack && fdc.cmdBuf[3] == 0 {
			fdc.cmdBuf[3] = 1
			fdc.cmdBuf[1] |= 0x04
			fdc.cmdBuf[4] = 1
			fdc.st0 |= fdcSt0HD // update ST0 head address
		} else {
			fdc.st1 |= fdcSt1EN
			fdc.enterResultPhaseRW()
			return
		}
	}

	// Ready for next sector
	fdc.dataPos = 0
	fdc.dataLen = fdc.SectorSize
	// Stay in exec phase
}

// --- Read ID command ---

func (fdc *FDC) execReadID() {
	drive := fdc.cmdBuf[1] & 0x03
	head := fdc.cmdBuf[1] & 0x04
	fdc.setupSt012(drive, head != 0)

	h := byte(0)
	if head != 0 {
		h = 1
	}

	// Return current cylinder, head, sector 1, sector size code
	fdc.resBuf[0] = fdc.st0
	fdc.resBuf[1] = fdc.st1
	fdc.resBuf[2] = fdc.st2
	fdc.resBuf[3] = fdc.pcn[drive]   // C
	fdc.resBuf[4] = h                // H
	fdc.resBuf[5] = 1                // R (sector 1)
	fdc.resBuf[6] = fdc.SectorSizeCode // N
	fdc.resLen = 7
	fdc.resPos = 0
	fdc.phase = fdcPhaseResult
	fdc.msr = fdcMsrRQM | fdcMsrDIO | fdcMsrCB
}

// --- Format Track command ---

func (fdc *FDC) execFormatTrack() {
	// Command bytes: [cmd, HD/DS, N, SC, GPL, D]
	drive := fdc.cmdBuf[1] & 0x03
	head := fdc.cmdBuf[1] & 0x04
	fdc.setupSt012(drive, head != 0)

	if fdc.files[drive] == nil {
		fdc.st0 |= fdcSt0IC0 | fdcSt0NR
		fdc.st1 |= fdcSt1MA
		fdc.enterResultPhaseFormat(0, 0, 0, 0)
		return
	}

	sc := int(fdc.cmdBuf[3]) // sectors per track to format
	// cmdBuf[5] (fill byte) is read by execFormatComplete

	// We need to receive 4 bytes per sector (C, H, R, N) from the CPU
	fdc.dataDir = 0 // CPU->FDC
	fdc.dataPos = 0
	fdc.dataLen = sc * 4
	fdc.phase = fdcPhaseExec
	fdc.msr = fdcMsrRQM | fdcMsrEXM | fdcMsrCB
}

func (fdc *FDC) execFormatComplete() {
	drive := fdc.cmdBuf[1] & 0x03
	sc := int(fdc.cmdBuf[3])
	d := fdc.cmdBuf[5]

	// Fill buffer with the fill byte for one sector
	fillBuf := make([]byte, fdc.SectorSize)
	for i := range fillBuf {
		fillBuf[i] = d
	}

	// Track the last sector's CHRN for the result phase
	var lastC, lastH, lastR, lastN byte

	// Process each sector's CHRN from the received data
	for i := 0; i < sc; i++ {
		lastC = fdc.dataBuf[i*4+0]
		lastH = fdc.dataBuf[i*4+1]
		lastR = fdc.dataBuf[i*4+2]
		lastN = fdc.dataBuf[i*4+3]

		offset := fdc.sectorOffset(lastC, lastH, lastR)
		if offset < 0 {
			continue
		}
		fdc.files[drive].WriteAt(fillBuf, offset)
	}

	fdc.enterResultPhaseFormat(lastC, lastH, lastR, lastN)
}

// --- Recalibrate command ---

func (fdc *FDC) execRecalibrate() {
	drive := fdc.cmdBuf[1] & 0x03
	fdc.pcn[drive] = 0
	fdc.pendingSt0[drive] = fdcSt0SE | drive // SE=1, IC=00 (normal)
	fdc.enterIdle()
}

// --- Sense Interrupt command ---

func (fdc *FDC) execSenseInterrupt() {
	// Find a drive with a pending interrupt
	for i := 0; i < 4; i++ {
		if fdc.pendingSt0[i] != 0 {
			fdc.resBuf[0] = fdc.pendingSt0[i] // ST0 (varies: 0xC0|drv for reset, 0x20|hd|drv for seek)
			fdc.resBuf[1] = fdc.pcn[i]        // Present Cylinder Number
			fdc.pendingSt0[i] = 0
			fdc.resLen = 2
			fdc.resPos = 0
			fdc.phase = fdcPhaseResult
			fdc.msr = fdcMsrRQM | fdcMsrDIO | fdcMsrCB
			return
		}
	}

	// No pending interrupts - return invalid command
	fdc.resBuf[0] = 0x80 // Invalid
	fdc.resLen = 1
	fdc.resPos = 0
	fdc.phase = fdcPhaseResult
	fdc.msr = fdcMsrRQM | fdcMsrDIO | fdcMsrCB
}

// --- Specify command ---

func (fdc *FDC) execSpecify() {
	// SRT/HUT and HLT/ND parameters - accepted but not used in emulation
	fdc.enterIdle()
}

// --- Sense Drive Status command ---

func (fdc *FDC) execSenseDriveStatus() {
	drive := fdc.cmdBuf[1] & 0x03
	head := fdc.cmdBuf[1] & 0x04

	fdc.st3 = drive
	if head != 0 {
		fdc.st3 |= fdcSt3HD
	}
	fdc.st3 |= fdcSt3TS // two sided
	fdc.st3 |= fdcSt3RY // ready
	if fdc.pcn[drive] == 0 {
		fdc.st3 |= fdcSt3T0
	}
	// Note: write protect not emulated, would need per-file flag

	fdc.resBuf[0] = fdc.st3
	fdc.resLen = 1
	fdc.resPos = 0
	fdc.phase = fdcPhaseResult
	fdc.msr = fdcMsrRQM | fdcMsrDIO | fdcMsrCB
}

// --- Seek command ---

func (fdc *FDC) execSeek() {
	drive := fdc.cmdBuf[1] & 0x03
	head := fdc.cmdBuf[1] & 0x04 // HD bit is already in bit 2 position
	ncn := fdc.cmdBuf[2]         // New Cylinder Number

	fdc.pcn[drive] = ncn
	fdc.pendingSt0[drive] = fdcSt0SE | head | drive // SE=1, IC=00 (normal)
	fdc.enterIdle()
}

// --- Helpers ---

// sectorOffset calculates the byte offset into the disk image for a given C/H/S.
// Returns -1 if the address is out of range.
func (fdc *FDC) sectorOffset(c, h, r byte) int64 {
	if int(c) >= fdc.Cylinders || int(h) >= fdc.Heads || r < 1 || int(r) > fdc.SectorsPerTrack {
		return -1
	}
	lba := (int64(c)*int64(fdc.Heads)+int64(h))*int64(fdc.SectorsPerTrack) + int64(r) - 1
	return lba * int64(fdc.SectorSize)
}

// enterResultPhaseRW sets up the result phase for read/write commands (7 result bytes).
func (fdc *FDC) enterResultPhaseRW() {
	fdc.resBuf[0] = fdc.st0
	fdc.resBuf[1] = fdc.st1
	fdc.resBuf[2] = fdc.st2
	fdc.resBuf[3] = fdc.cmdBuf[2] // C
	fdc.resBuf[4] = fdc.cmdBuf[3] // H
	fdc.resBuf[5] = fdc.cmdBuf[4] // R
	fdc.resBuf[6] = fdc.cmdBuf[5] // N
	fdc.resLen = 7
	fdc.resPos = 0
	fdc.phase = fdcPhaseResult
	fdc.msr = fdcMsrRQM | fdcMsrDIO | fdcMsrCB
}

// enterResultPhaseFormat sets up the result phase for format track commands (7 result bytes).
func (fdc *FDC) enterResultPhaseFormat(c, h, r, n byte) {
	fdc.resBuf[0] = fdc.st0
	fdc.resBuf[1] = fdc.st1
	fdc.resBuf[2] = fdc.st2
	fdc.resBuf[3] = c
	fdc.resBuf[4] = h
	fdc.resBuf[5] = r
	fdc.resBuf[6] = n
	fdc.resLen = 7
	fdc.resPos = 0
	fdc.phase = fdcPhaseResult
	fdc.msr = fdcMsrRQM | fdcMsrDIO | fdcMsrCB
}

// enterIdle returns the controller to idle state.
func (fdc *FDC) enterIdle() {
	fdc.phase = fdcPhaseIdle
	fdc.msr = fdcMsrRQM
	fdc.cmdPos = 0
}
