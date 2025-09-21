package cpu4004

import (
	"github.com/scottmbaker/gocpusim/pkg/cpusim"
)

type Bus8Bit struct {
	Sim            *cpusim.CpuSim
	Name           string
	Enabler        cpusim.EnablerInterface
	Memory         []cpusim.MemoryInterface
	Ports          []cpusim.MemoryInterface
	LastReadValue  byte
	LastWriteValue byte
}

func (b *Bus8Bit) GetName() string {
	return b.Name
}

func (b *Bus8Bit) HasAddress(address cpusim.Address) bool {
	if !b.Enabler.Bool() {
		return false
	}
	return (address >= 0) && (address <= 255)
}

func (b *Bus8Bit) _write(address cpusim.Address, value byte) error {
	for _, mem := range b.Memory {
		if mem.HasAddress(address) {
			return mem.Write(address, value)
		}
	}
	return nil
}

func (b *Bus8Bit) _read(address cpusim.Address) (byte, error) {
	for _, mem := range b.Memory {
		if mem.HasAddress(address) {
			return mem.Read(address)
		}
	}
	return 0, nil
}

func (b *Bus8Bit) Read(address cpusim.Address) (byte, error) {
	_ = address
	return 0, &cpusim.ErrNotImplemented{Device: b}
}

func (b *Bus8Bit) Write(address cpusim.Address, value byte) error {
	_ = address
	_ = value
	return &cpusim.ErrNotImplemented{Device: b}
}

func (b *Bus8Bit) ReadStatus(address cpusim.Address, statusAddr cpusim.Address) (byte, error) {
	var err error
	if !b.HasAddress(address) {
		return 0, &cpusim.ErrInvalidAddress{Device: b, Address: address}
	}
	if statusAddr < 0 || statusAddr >= 4 {
		return 0, &cpusim.ErrInvalidAddress{Device: b, Address: statusAddr}
	}
	switch statusAddr {
	case 0:
		b.LastReadValue, err = b._read(address)
		if err != nil {
			return 0, err
		}
		return b.LastReadValue & 0x0F, nil
	case 1:
		return (b.LastReadValue >> 4) & 0x0F, nil
	}
	return 0, &cpusim.ErrNotImplemented{Device: b}
}

func (b *Bus8Bit) WriteStatus(address cpusim.Address, statusAddr cpusim.Address, value byte) error {
	var err error
	if !b.HasAddress(address) {
		return &cpusim.ErrInvalidAddress{Device: b, Address: address}
	}
	if statusAddr < 0 || statusAddr >= 4 {
		return &cpusim.ErrInvalidAddress{Device: b, Address: statusAddr}
	}
	switch statusAddr {
	case 0:
		b.LastWriteValue = b.LastWriteValue&0xF0 | (value & 0x0F)
		err = b._write(address, b.LastWriteValue)
		if err != nil {
			return err
		}
	case 1:
		b.LastWriteValue = b.LastWriteValue&0x0F | ((value & 0x0F) << 4)
	}
	return &cpusim.ErrNotImplemented{Device: b}
}

func (b *Bus8Bit) AddMemory(memory cpusim.MemoryInterface) {
	b.Memory = append(b.Memory, memory)
}

func (b *Bus8Bit) AddPort(port cpusim.MemoryInterface) {
	b.Ports = append(b.Ports, port)
}

func NewBus8Bit(sim *cpusim.CpuSim, name string, enabler cpusim.EnablerInterface) *Bus8Bit {
	mem := &Bus8Bit{
		Sim:     sim,
		Name:    name,
		Enabler: enabler,
		Memory:  make([]cpusim.MemoryInterface, 0),
		Ports:   make([]cpusim.MemoryInterface, 0),
	}
	return mem
}
