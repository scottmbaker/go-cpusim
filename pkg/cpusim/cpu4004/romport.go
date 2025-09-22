package cpu4004

import (
	"github.com/scottmbaker/gocpusim/pkg/cpusim"
)

type RomPort struct {
	Sim     *cpusim.CpuSim
	Name    string
	Enabler cpusim.EnablerInterface
	Ports   []cpusim.MemoryInterface
}

func (b *RomPort) GetName() string {
	return b.Name
}

func (b *RomPort) HasAddress(address cpusim.Address) bool {
	if !b.Enabler.Bool() {
		return false
	}
	return (address >= 0) && (address <= 255)
}

func (b *RomPort) _writePort(address cpusim.Address, value byte) error {
	address = (address >> 4) & 0x0F
	for _, mem := range b.Ports {
		if mem.HasAddress(address) {
			return mem.Write(address, value)
		}
	}
	return nil
}

func (b *RomPort) _readPort(address cpusim.Address) (byte, error) {
	address = (address >> 4) & 0x0F
	for _, mem := range b.Ports {
		if mem.HasAddress(address) {
			return mem.Read(address)
		}
	}
	return 0, nil
}

func (b *RomPort) Read(address cpusim.Address) (byte, error) {
	return b._readPort(address)
}

func (b *RomPort) Write(address cpusim.Address, value byte) error {
	return b._writePort(address, value)
}

func (b *RomPort) ReadStatus(address cpusim.Address, statusAddr cpusim.Address) (byte, error) {
	_ = address
	_ = statusAddr
	return 0, &cpusim.ErrNotImplemented{Device: b}
}

func (b *RomPort) WriteStatus(address cpusim.Address, statusAddr cpusim.Address, value byte) error {
	_ = address
	_ = statusAddr
	_ = value
	return &cpusim.ErrNotImplemented{Device: b}
}

func (b *RomPort) AddPort(port cpusim.MemoryInterface) {
	b.Ports = append(b.Ports, port)
}

func (b *RomPort) GetKind() string {
	return cpusim.KIND_ROMPORT
}

func NewRomPort(sim *cpusim.CpuSim, name string, enabler cpusim.EnablerInterface) *RomPort {
	mem := &RomPort{
		Sim:     sim,
		Name:    name,
		Enabler: enabler,
		Ports:   make([]cpusim.MemoryInterface, 0),
	}
	return mem
}
