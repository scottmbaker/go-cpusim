package cpusim

import (
	"fmt"
)

// 74LS670 style memory mapper
// Supports up to 8 bits (two 74LS670)

type Map670 struct {
	Sim                *CpuSim
	Name               string
	MapperAddress      Address
	SourceMask         Address
	SourceBit          int
	SourceData         int
	DestBit            [8]int
	Contents           [16]byte
	ConnectedEnableBit [8]*EnableBit
	Enabler            EnablerInterface
	MemoryFilter       string
}

func (m *Map670) GetName() string {
	return m.Name
}

func (m *Map670) HasAddress(address Address) bool {
	if !m.Enabler.Bool() {
		return false
	}
	return (address >= m.MapperAddress) && (address <= (m.MapperAddress + 3))
}

func (m *Map670) Write(address Address, value byte) error {
	index := (address - m.MapperAddress) & m.SourceMask
	m.Contents[index] = value
	if m.Sim.MemDebug {
		fmt.Printf("MAP 670 %s: Writing value %02X to address %04X index %04X\n", m.Name, value, address, index)
	}
	return nil
}

func (m *Map670) Read(address Address) (byte, error) {
	return 0, &ErrReadOnly{Device: m}
}

func (m *Map670) WriteStatus(address Address, statusAddr Address, value byte) error {
	_ = address
	_ = statusAddr
	_ = value
	return &ErrNotImplemented{Device: m}

}

func (m *Map670) ReadStatus(address Address, statusAddr Address) (byte, error) {
	_ = address
	_ = statusAddr
	return 0, &ErrReadOnly{Device: m}
}

func (m *Map670) Map(address Address) (Address, error) {
	addressIn := address
	index := (address >> m.SourceBit) & m.SourceMask
	value := m.Contents[index]
	for i := 0; i < 8; i++ {
		bitIsSet := (value & (1 << i)) != 0
		if m.DestBit[i] >= 0 {
			bitMask := Address(1 << m.DestBit[i])
			address &= ^bitMask
			if bitIsSet {
				address |= bitMask
			}
		}
		if m.ConnectedEnableBit[i] != nil {
			m.ConnectedEnableBit[i].Set(bitIsSet)
		}
	}
	_ = addressIn
	if m.Sim.MemDebug {
		fmt.Printf("Mapper %s <%04X:%02X> index %02X --> %04X\n", m.Name, addressIn, value, index, address)
	}
	return address, nil
}

func (m *Map670) ConnectEnableBit(bit int, enableBit *EnableBit) {
	m.ConnectedEnableBit[bit] = enableBit
}

func (m *Map670) GetKind() string {
	return KIND_MAPPER
}

func (m *Map670) FilterMemoryKind(kind string) {
	m.MemoryFilter = kind
}

func (m *Map670) MatchMemory(mem MemoryInterface) bool {
	return m.MemoryFilter == "" || mem.GetKind() == m.MemoryFilter
}

func New74670(sim *CpuSim, name string, address Address, sourceBit, sourceData, destBit0, destBit1, destBit2, destBit3 int, enabler EnablerInterface) *Map670 {
	return &Map670{
		Sim:           sim,
		Name:          name,
		MapperAddress: address,
		SourceMask:    0x03,
		SourceBit:     sourceBit,
		SourceData:    sourceData,
		DestBit:       [8]int{destBit0, destBit1, destBit2, destBit3, -1, -1, -1, -1},
		Enabler:       enabler,
	}
}
