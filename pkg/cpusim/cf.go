package cpusim

import (
	"encoding/binary"
	"fmt"
	"os"
)

const (
	cfRegData      = 0
	cfRegError     = 1 // read
	cfRegFeature   = 1 // write
	cfRegSecCount  = 2
	cfRegLBALow    = 3
	cfRegLBAMid    = 4
	cfRegLBAHigh   = 5
	cfRegDevHead   = 6
	cfRegStatus    = 7 // read
	cfRegCommand   = 7 // write
	cfRegAltStatus = 8 // read
	cfRegDevCtrl   = 8 // write
	cfRegDataLatch = 9

	cfNumRegs = 10

	cfStBSY  = 0x80
	cfStDRDY = 0x40
	cfStDF   = 0x20
	cfStDSC  = 0x10
	cfStDRQ  = 0x08
	cfStCORR = 0x04
	cfStIDX  = 0x02
	cfStERR  = 0x01

	cfErrAMNF = 0x01
	cfErrABRT = 0x04
	cfErrIDNF = 0x10
	cfErrUNC  = 0x40

	cfDevLBA  = 0x40
	cfDevDEV  = 0x10
	cfDevHEAD = 0x0F

	cfStateIdle    = 0
	cfStateDataIn  = 1
	cfStateDataOut = 2

	cfCmdRecalibrate = 0x10
	cfCmdRead        = 0x20
	cfCmdReadNR      = 0x21
	cfCmdWrite       = 0x30
	cfCmdWriteNR     = 0x31
	cfCmdVerify      = 0x40
	cfCmdVerifyNR    = 0x41
	cfCmdSeek        = 0x70
	cfCmdEDD         = 0x90
	cfCmdInitParams  = 0x91
	cfCmdIdentify    = 0xEC
	cfCmdSetFeatures = 0xEF
)

// CompactFlash emulates an IDE/CompactFlash device accessed via 8-bit I/O with
// a data latch for 16-bit transfers.
//
// Registers (offsets from BaseAddress):
//
//	0: Data            5: LBA High / Cylinder High
//	1: Error / Feature 6: Device/Head
//	2: Sector Count    7: Status / Command
//	3: LBA Low         8: Alt Status / Device Control
//	4: LBA Mid         9: Data Latch (high byte)
type CompactFlash struct {
	Sim         *CpuSim
	Name        string
	BaseAddress Address
	Enabler     EnablerInterface

	file      *os.File
	imageOff  int64 // byte offset to sector 0 in the image file
	identify  [512]byte
	data      [512]byte
	dptr      int
	dataLatch byte

	state  int
	length int // sectors remaining

	// Task file registers
	error   byte
	feature byte
	count   byte
	lba1    byte // LBA low / sector number
	lba2    byte // LBA mid / cylinder low
	lba3    byte // LBA high / cylinder high
	lba4    byte // device/head
	status  byte
	devctrl byte

	// Drive geometry (from identify block)
	cylinders uint16
	heads     byte
	sectors   byte
	lba       bool
	eightbit  bool
}

func NewCompactFlash(sim *CpuSim, name string, baseAddress Address, enabler EnablerInterface) *CompactFlash {
	return &CompactFlash{
		Sim:         sim,
		Name:        name,
		BaseAddress: baseAddress,
		Enabler:     enabler,
		status:      cfStDRDY | cfStDSC,
	}
}

// AttachImage opens a disk image file. imageOffset is the byte offset where
// sector 0 begins (use 1024 for emulatorkit images that have a 1KB header,
// or 0 for raw images).
func (cf *CompactFlash) AttachImage(filename string, imageOffset int64) error {
	f, err := os.OpenFile(filename, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("cf: open %s: %w", filename, err)
	}
	cf.file = f
	cf.imageOff = imageOffset
	return nil
}

// LoadIdentify loads the 512-byte identify block from a separate file.
// If not called, the identify block can be loaded from the image header
// via LoadIdentifyFromImage.
func (cf *CompactFlash) LoadIdentify(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("cf: read identify %s: %w", filename, err)
	}
	if len(data) < 512 {
		return fmt.Errorf("cf: identify file too short (%d bytes)", len(data))
	}
	copy(cf.identify[:], data[:512])
	cf.parseIdentify()
	return nil
}

// LoadIdentifyFromImage reads the identify block from offset 512 in the image
// file (the second 512 bytes of the emulatorkit 1KB header).
func (cf *CompactFlash) LoadIdentifyFromImage() error {
	if cf.file == nil {
		return fmt.Errorf("cf: no image attached")
	}
	_, err := cf.file.ReadAt(cf.identify[:], 512)
	if err != nil {
		return fmt.Errorf("cf: read identify from image: %w", err)
	}
	cf.parseIdentify()
	return nil
}

func (cf *CompactFlash) parseIdentify() {
	cf.cylinders = binary.LittleEndian.Uint16(cf.identify[1*2:])
	cf.heads = byte(binary.LittleEndian.Uint16(cf.identify[3*2:]))
	cf.sectors = byte(binary.LittleEndian.Uint16(cf.identify[6*2:]))
	cap := binary.LittleEndian.Uint16(cf.identify[49*2:])
	cf.lba = cap&(1<<9) != 0
}

func (cf *CompactFlash) GetName() string {
	return cf.Name
}

func (cf *CompactFlash) GetKind() string {
	return KIND_CF
}

func (cf *CompactFlash) HasAddress(address Address) bool {
	if !cf.Enabler.Bool() {
		return false
	}
	reg := int(address) - int(cf.BaseAddress)
	return reg >= 0 && reg < cfNumRegs
}

func (cf *CompactFlash) Read(address Address) (byte, error) {
	reg := int(address) - int(cf.BaseAddress)

	switch reg {
	case cfRegData:
		return cf.dataIn(), nil
	case cfRegError:
		return cf.error, nil
	case cfRegSecCount:
		return cf.count, nil
	case cfRegLBALow:
		return cf.lba1, nil
	case cfRegLBAMid:
		return cf.lba2, nil
	case cfRegLBAHigh:
		return cf.lba3, nil
	case cfRegDevHead:
		return cf.lba4, nil
	case cfRegStatus:
		return cf.status, nil
	case cfRegAltStatus:
		return cf.status, nil
	case cfRegDataLatch:
		return cf.dataLatch, nil
	}
	return 0xFF, nil
}

func (cf *CompactFlash) Write(address Address, value byte) error {
	reg := int(address) - int(cf.BaseAddress)

	switch reg {
	case cfRegData:
		cf.dataOut(value)
	case cfRegFeature:
		cf.feature = value
	case cfRegSecCount:
		cf.count = value
	case cfRegLBALow:
		cf.lba1 = value
	case cfRegLBAMid:
		cf.lba2 = value
	case cfRegLBAHigh:
		cf.lba3 = value
	case cfRegDevHead:
		cf.lba4 = value & (cfDevHEAD | cfDevDEV | cfDevLBA)
	case cfRegCommand:
		cf.issueCommand(value)
	case cfRegDevCtrl:
		if (value^cf.devctrl)&0x04 != 0 {
			if value&0x04 != 0 {
				cf.status |= cfStBSY
			} else {
				cf.reset()
			}
		}
		cf.devctrl = value
	case cfRegDataLatch:
		cf.dataLatch = value
	}
	return nil
}

func (cf *CompactFlash) ReadStatus(address Address, statusAddr Address) (byte, error) {
	return 0, &ErrNotImplemented{Device: cf}
}

func (cf *CompactFlash) WriteStatus(address Address, statusAddr Address, value byte) error {
	return &ErrNotImplemented{Device: cf}
}

func (cf *CompactFlash) reset() {
	cf.status = cfStDRDY | cfStDSC
	cf.error = 0x01
	cf.lba1 = 0x01
	cf.lba2 = 0
	cf.lba3 = 0
	cf.lba4 = 0
	cf.count = 0x01
	cf.state = cfStateIdle
	cf.eightbit = false
}

func (cf *CompactFlash) xlateBlock() int64 {
	if cf.lba4&cfDevLBA != 0 {
		if !cf.lba {
			return -1
		}
		return int64(cf.lba4&cfDevHEAD)<<24 |
			int64(cf.lba3)<<16 |
			int64(cf.lba2)<<8 |
			int64(cf.lba1)
	}
	// CHS
	sector := cf.lba1
	if sector == 0 {
		sector = 1
	}
	cyl := uint16(cf.lba3)<<8 | uint16(cf.lba2)
	head := cf.lba4 & cfDevHEAD
	if sector > cf.sectors || head >= cf.heads || cyl >= cf.cylinders {
		return -1
	}
	return int64((uint32(cyl)*uint32(cf.heads)+uint32(head))*uint32(cf.sectors)) + int64(sector) - 1
}

func (cf *CompactFlash) sectorOffset(block int64) int64 {
	return cf.imageOff + block*512
}

func (cf *CompactFlash) readSector() bool {
	block := cf.xlateBlock()
	if block < 0 {
		cf.status |= cfStERR
		cf.error |= cfErrIDNF
		return false
	}
	off := cf.sectorOffset(block)
	_, err := cf.file.ReadAt(cf.data[:], off)
	if err != nil {
		cf.status |= cfStERR
		cf.error |= cfErrUNC
		return false
	}
	cf.dptr = 0
	return true
}

func (cf *CompactFlash) writeSector() bool {
	block := cf.xlateBlock()
	if block < 0 {
		cf.status |= cfStERR
		cf.error |= cfErrIDNF
		return false
	}
	off := cf.sectorOffset(block)
	_, err := cf.file.WriteAt(cf.data[:], off)
	if err != nil {
		cf.status |= cfStERR
		cf.error |= cfErrUNC
		return false
	}
	return true
}

func (cf *CompactFlash) advanceLBA() {
	if cf.lba4&cfDevLBA != 0 {
		lba := uint32(cf.lba4&cfDevHEAD)<<24 |
			uint32(cf.lba3)<<16 |
			uint32(cf.lba2)<<8 |
			uint32(cf.lba1)
		lba++
		cf.lba1 = byte(lba)
		cf.lba2 = byte(lba >> 8)
		cf.lba3 = byte(lba >> 16)
		cf.lba4 = (cf.lba4 & ^byte(cfDevHEAD)) | byte(lba>>24)&cfDevHEAD
	} else {
		cf.lba1++
		if cf.lba1 > cf.sectors {
			cf.lba1 = 1
			head := (cf.lba4 & cfDevHEAD) + 1
			if head >= cf.heads {
				head = 0
				cyl := uint16(cf.lba3)<<8 | uint16(cf.lba2)
				cyl++
				cf.lba2 = byte(cyl)
				cf.lba3 = byte(cyl >> 8)
			}
			cf.lba4 = (cf.lba4 & ^byte(cfDevHEAD)) | (head & cfDevHEAD)
		}
	}
}

func (cf *CompactFlash) dataIn() byte {
	if cf.state != cfStateDataIn {
		return 0xFF
	}

	if cf.dptr >= 512 {
		if !cf.readSector() {
			cf.completed()
			return 0xFF
		}
	}

	var v byte
	if cf.eightbit {
		v = cf.data[cf.dptr]
		cf.dptr++
	} else {
		v = cf.data[cf.dptr]
		cf.dataLatch = cf.data[cf.dptr+1]
		cf.dptr += 2
	}

	if cf.dptr >= 512 {
		cf.length--
		cf.advanceLBA()
		if cf.length == 0 {
			cf.completed()
		}
	}
	return v
}

func (cf *CompactFlash) dataOut(v byte) {
	if cf.state != cfStateDataOut {
		return
	}

	if cf.eightbit {
		cf.data[cf.dptr] = v
		cf.dptr++
	} else {
		cf.data[cf.dptr] = v
		cf.data[cf.dptr+1] = cf.dataLatch
		cf.dptr += 2
	}

	if cf.dptr >= 512 {
		if !cf.writeSector() {
			cf.completed()
			return
		}
		cf.length--
		cf.advanceLBA()
		if cf.length == 0 {
			cf.status |= cfStDSC
			cf.completed()
		} else {
			cf.dptr = 0
		}
	}
}

func (cf *CompactFlash) completed() {
	cf.status &= ^byte(cfStBSY | cfStDRQ)
	cf.status |= cfStDRDY
	cf.state = cfStateIdle
}

func (cf *CompactFlash) issueCommand(cmd byte) {
	cf.status &= ^byte(cfStERR | cfStDRDY)
	cf.status |= cfStBSY
	cf.error = 0

	switch {
	case cmd == cfCmdIdentify:
		copy(cf.data[:], cf.identify[:])
		cf.dptr = 0
		cf.length = 1
		cf.state = cfStateDataIn
		cf.status &= ^byte(cfStBSY)
		cf.status |= cfStDRQ | cfStDRDY

	case cmd == cfCmdRead || cmd == cfCmdReadNR:
		cf.length = int(cf.count)
		if cf.length == 0 {
			cf.length = 256
		}
		cf.dptr = 512 // force read on first data access
		cf.state = cfStateDataIn
		cf.status &= ^byte(cfStBSY)
		cf.status |= cfStDRQ | cfStDSC | cfStDRDY

	case cmd == cfCmdWrite || cmd == cfCmdWriteNR:
		cf.length = int(cf.count)
		if cf.length == 0 {
			cf.length = 256
		}
		cf.dptr = 0
		cf.state = cfStateDataOut
		cf.status &= ^byte(cfStBSY)
		cf.status |= cfStDRQ | cfStDRDY

	case cmd == cfCmdVerify || cmd == cfCmdVerifyNR:
		block := cf.xlateBlock()
		if block < 0 {
			cf.status |= cfStERR
			cf.error |= cfErrIDNF
		}
		cf.status |= cfStDSC
		cf.completed()

	case cmd == cfCmdEDD:
		cf.error = 0x01
		cf.lba1 = 0x01
		cf.lba2 = 0
		cf.lba3 = 0
		cf.lba4 = 0
		cf.count = 0x01
		cf.completed()

	case cmd == cfCmdInitParams:
		if cf.count != cf.sectors || (cf.lba4&cfDevHEAD)+1 != cf.heads {
			cf.status |= cfStERR
			cf.error |= cfErrABRT
		}
		cf.completed()

	case cmd == cfCmdSetFeatures:
		switch cf.feature {
		case 0x01:
			cf.eightbit = true
		case 0x81:
			cf.eightbit = false
		case 0x03:
			// Accept PIO mode setting
		default:
			cf.status |= cfStERR
			cf.error |= cfErrABRT
		}
		cf.completed()

	case cmd&0xF0 == cfCmdRecalibrate:
		cf.status |= cfStDSC
		cf.completed()

	case cmd&0xF0 == cfCmdSeek:
		block := cf.xlateBlock()
		if block < 0 {
			cf.status |= cfStERR
			cf.error |= cfErrIDNF
		}
		cf.status |= cfStDSC
		cf.completed()

	default:
		cf.status |= cfStERR
		cf.error |= cfErrABRT
		cf.completed()
	}
}

func (cf *CompactFlash) Close() {
	if cf.file != nil {
		cf.file.Close()
		cf.file = nil
	}
}
