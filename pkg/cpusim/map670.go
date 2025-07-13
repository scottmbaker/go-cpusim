package cpusim

import (
//"fmt"
)

// 74LS670 style memory mapper
// Supports up to 8 bits (two 74LS670)

type Map670 struct {
	Sim                *CpuSim
	Name               string
	Address            uint16
	SourceBit          int
	SourceData         int
	DestBit            [8]int
	Contents           [16]byte
	ConnectedEnableBit [8]*EnableBit
	Enabler            EnablerInterface
}

func (m *Map670) GetName() string {
	return m.Name
}

func (m *Map670) HasAddress(address uint16) bool {
	if !m.Enabler.Bool() {
		return false
	}
	return (address >= m.Address) && (address <= (m.Address + 3))
}

func (m *Map670) Write(address uint16, value byte) error {
	index := (address - m.Address) & 0x03
	m.Contents[index] = value
	//fmt.Printf("MAP 670: Writing value %02X to address %04X\n", value, address)
	return nil
}

func (m *Map670) Read(address uint16) (byte, error) {
	return 0, &ErrReadOnly{}
}

func (m *Map670) Map(address uint16) (uint16, error) {
	index := address >> m.SourceBit
	value := m.Contents[index]
	//fmt.Printf("<%04X:%02X>", address, value)
	for i := 0; i < 8; i++ {
		bitIsSet := (value & (1 << i)) != 0
		if m.DestBit[i] >= 0 {
			bitMask := uint16(1 << m.DestBit[i])
			address &= ^bitMask
			if bitIsSet {
				address |= bitMask
			}
		}
		if m.ConnectedEnableBit[i] != nil {
			m.ConnectedEnableBit[i].Set(bitIsSet)
		}
	}
	return address, nil
}

func (m *Map670) ConnectEnableBit(bit int, enableBit *EnableBit) {
	m.ConnectedEnableBit[bit] = enableBit
}

func New74670(sim *CpuSim, name string, address uint16, sourceBit, sourceData, destBit0, destBit1, destBit2, destBit3 int, enabler EnablerInterface) *Map670 {
	return &Map670{
		Sim:        sim,
		Name:       name,
		Address:    address,
		SourceBit:  sourceBit,
		SourceData: sourceData,
		DestBit:    [8]int{destBit0, destBit1, destBit2, destBit3, -1, -1, -1, -1},
		Enabler:    enabler,
	}
}
