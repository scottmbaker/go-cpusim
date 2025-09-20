package cpusim

import (
	"io"
	"os"
)

type Memory struct {
	Sim          *CpuSim
	Name         string
	StartAddress Address
	EndAddress   Address
	AddressBits  int
	ReadOnly     bool
	Contents     []byte
	Enabler      EnablerInterface
}

func (mem *Memory) GetName() string {
	return mem.Name
}

func (mem *Memory) HasAddress(address Address) bool {
	if !mem.Enabler.Bool() {
		return false
	}
	return (address >= mem.StartAddress) && (address <= mem.EndAddress)
}

func (mem *Memory) Read(address Address) (byte, error) {
	if !mem.HasAddress(address) {
		return 0, &ErrInvalidAddress{Address: address}
	}
	index := address - mem.StartAddress
	return mem.Contents[index], nil
}

func (mem *Memory) Write(address Address, value byte) error {
	if !mem.HasAddress(address) {
		return &ErrInvalidAddress{Address: address}
	}
	if mem.ReadOnly {
		return &ErrReadOnly{}
	}
	index := address - mem.StartAddress
	mem.Contents[index] = value

	return nil
}

func (mem *Memory) Load(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close() // nolint:errcheck

	_, err = file.Read(mem.Contents)
	if err != nil && err != io.EOF {
		return err
	}

	return nil
}

func NewMemory(sim *CpuSim, name string, startAddress, endAddress Address, addressBits int, readonly bool, enabler EnablerInterface) *Memory {
	mem := &Memory{
		Sim:          sim,
		Name:         name,
		StartAddress: startAddress,
		EndAddress:   endAddress,
		AddressBits:  addressBits,
		ReadOnly:     readonly,
		Contents:     make([]byte, (int(endAddress)-int(startAddress))+1),
		Enabler:      enabler,
	}
	return mem
}
