package cpusim

import (
	"fmt"
)

// SC714 style memory mapper
// Supports up to 8 bits (two 74LS670)

type MapSC714 struct {
	Sim           *CpuSim
	Name          string
	MapperAddress Address
	Contents      byte
	RamRomEnable  *EnableBit
	Enabler       EnablerInterface // Enabler for the Port read/writes to the mapper
	MapEnabler    EnablerInterface // Enabler for the Map function to read the contents and set the enable bits. Otherwise, always sets 0.
	MemoryFilter  string
}

func (m *MapSC714) GetName() string {
	return m.Name
}

func (m *MapSC714) HasAddress(address Address) bool {
	if !m.Enabler.Bool() {
		return false
	}
	return (address >= m.MapperAddress) && (address <= (m.MapperAddress + 3))
}

func (m *MapSC714) Write(address Address, value byte) error {
	m.Contents = value
	if m.Sim.MemDebug {
		fmt.Printf("MAP SC714 %s: Writing value %02X\n", m.Name, value)
	}
	return nil
}

func (m *MapSC714) Read(address Address) (byte, error) {
	return 0, &ErrReadOnly{Device: m}
}

func (m *MapSC714) WriteStatus(address Address, statusAddr Address, value byte) error {
	_ = address
	_ = statusAddr
	_ = value
	return &ErrNotImplemented{Device: m}

}

func (m *MapSC714) ReadStatus(address Address, statusAddr Address) (byte, error) {
	_ = address
	_ = statusAddr
	return 0, &ErrReadOnly{Device: m}
}

func (m *MapSC714) Map(address Address) (Address, error) {
	addressIn := address

	if address&0x8000 != 0 {
		// Top window is always top 32K of RAM
		if m.RamRomEnable != nil {
			m.RamRomEnable.Set(true)
		}
		address = address | 0x78000
	} else {
		address = (address & 0x7FFF) | (Address(m.Contents) << 14)

		if m.RamRomEnable != nil {
			if (m.Contents & 0x10) != 0 { // bit 5
				m.RamRomEnable.Set(true)
			} else {
				m.RamRomEnable.Set(false)
			}
		}
	}

	_ = addressIn
	if m.Sim.MemDebug {
		fmt.Printf("Mapper %s <%04X:%02X> --> %04X\n", m.Name, addressIn, m.Contents, address)
	}
	return address, nil
}

func (m *MapSC714) ConnectEnableBit(bit int, enableBit *EnableBit) {
	_ = bit
	_ = enableBit
}

func (m *MapSC714) GetKind() string {
	return KIND_MAPPER
}

func (m *MapSC714) FilterMemoryKind(kind string) {
	m.MemoryFilter = kind
}

func (m *MapSC714) MatchMemory(mem MemoryInterface) bool {
	return m.MemoryFilter == "" || mem.GetKind() == m.MemoryFilter
}

func NewSC714(sim *CpuSim, name string, address Address, ramRomEnable *EnableBit, enabler EnablerInterface, mapEnabler EnablerInterface) *MapSC714 {
	return &MapSC714{
		Sim:           sim,
		Name:          name,
		MapperAddress: address,
		RamRomEnable:  ramRomEnable,
		Enabler:       enabler,
		MapEnabler:    mapEnabler,
	}
}
