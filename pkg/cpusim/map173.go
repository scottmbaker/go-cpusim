package cpusim

import (
// "fmt"
)

// 74LS173 style memory mapper

type Map173 struct {
	Sim           *CpuSim
	Name          string
	MapperAddress Address
	SourceMask    Address
	DestBit       [4]int
	Contents      byte
	Enabler       EnablerInterface
	MemoryFilter  string
}

func (m *Map173) GetName() string {
	return m.Name
}

func (m *Map173) HasAddress(address Address) bool {
	if !m.Enabler.Bool() {
		return false
	}
	return address == m.MapperAddress
}

func (m *Map173) Write(address Address, value byte) error {
	m.Contents = value
	//fmt.Printf("MAP 173: Writing value %02X to address %04X\n", value, address)
	return nil
}

func (m *Map173) Read(address Address) (byte, error) {
	return 0, &ErrReadOnly{Device: m}
}

func (m *Map173) WriteStatus(address Address, statusAddr Address, value byte) error {
	_ = address
	_ = statusAddr
	_ = value
	return &ErrNotImplemented{Device: m}

}

func (m *Map173) ReadStatus(address Address, statusAddr Address) (byte, error) {
	_ = address
	_ = statusAddr
	return 0, &ErrReadOnly{Device: m}
}

func (m *Map173) Map(address Address) (Address, error) {
	value := m.Contents
	for i := 0; i < 4; i++ {
		bitIsSet := (value & (1 << i)) != 0
		if m.DestBit[i] >= 0 {
			bitMask := Address(1 << m.DestBit[i])
			address &= ^bitMask
			if bitIsSet {
				address |= bitMask
			}
		}
		//if m.ConnectedEnableBit[i] != nil {
		//	m.ConnectedEnableBit[i].Set(bitIsSet)
		//}
	}
	return address, nil
}

func (m *Map173) GetKind() string {
	return KIND_MAPPER
}

func (m *Map173) FilterMemoryKind(kind string) {
	m.MemoryFilter = kind
}

func (m *Map173) MatchMemory(mem MemoryInterface) bool {
	return m.MemoryFilter == "" || mem.GetKind() == m.MemoryFilter
}

func New74173(sim *CpuSim, name string, address Address, destBit0, destBit1, destBit2, destBit3 int, enabler EnablerInterface) *Map173 {
	return &Map173{
		Sim:           sim,
		Name:          name,
		MapperAddress: address,
		DestBit:       [4]int{destBit0, destBit1, destBit2, destBit3},
		Enabler:       enabler,
	}
}
