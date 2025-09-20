package cpusim

import (
	"fmt"
	"sync"
)

const (
	NOPIN = -1

	A0  = 0
	A1  = 1
	A2  = 2
	A3  = 3
	A4  = 4
	A5  = 5
	A6  = 6
	A7  = 7
	A8  = 8
	A9  = 9
	A10 = 10
	A11 = 11
	A12 = 12
	A13 = 13
	A14 = 14
	A15 = 15

	D0 = 0
	D1 = 1
	D2 = 2
	D3 = 3
	D4 = 4
	D5 = 5
	D6 = 6
	D7 = 7
)

type CpuSim struct {
	CPU     []CpuInterface
	Memory  []MemoryInterface
	Ports   []MemoryInterface
	Mappers []MapperInterface
	CtrlC   bool
	Debug   bool
}

func NewCPUSim() *CpuSim {
	return &CpuSim{
		CPU:    make([]CpuInterface, 0),
		Memory: make([]MemoryInterface, 0),
		Ports:  make([]MemoryInterface, 0),
		Debug:  true,
	}
}

func (sim *CpuSim) SetDebug(debug bool) {
	sim.Debug = debug
}

func (sim *CpuSim) AddCPU(cpu CpuInterface) {
	sim.CPU = append(sim.CPU, cpu)
}

func (sim *CpuSim) AddMemory(memory MemoryInterface) {
	sim.Memory = append(sim.Memory, memory)
}

func (sim *CpuSim) AddPort(port MemoryInterface) {
	sim.Ports = append(sim.Ports, port)
}

func (sim *CpuSim) AddMapper(mapper MapperInterface) {
	sim.Mappers = append(sim.Mappers, mapper)
}

func (sim *CpuSim) Start(wg *sync.WaitGroup) {
	for _, cpu := range sim.CPU {
		wg.Add(1)
		go func(c CpuInterface) {
			defer wg.Done()
			err := c.Run()
			if err != nil {
				fmt.Printf("%s", err)
			}
		}(cpu)
	}
}

func (sim *CpuSim) WriteMemory(address Address, value byte) error {
	for _, mapper := range sim.Mappers {
		var err error
		address, err = mapper.Map(address)
		if err != nil {
			return err
		}
	}
	for _, mem := range sim.Memory {
		if mem.HasAddress(address) {
			return mem.Write(address, value)
		}
	}
	return nil
}

func (sim *CpuSim) ReadMemory(address Address) (byte, error) {
	for _, mapper := range sim.Mappers {
		var err error
		address, err = mapper.Map(address)
		if err != nil {
			return 0, err
		}
	}
	for _, mem := range sim.Memory {
		if mem.HasAddress(address) {
			return mem.Read(address)
		}
	}
	return 0, nil
}

func (sim *CpuSim) ReadPort(port Address) (byte, error) {
	for _, p := range sim.Ports {
		if p.HasAddress(port) {
			return p.Read(port)
		}
	}
	return 0, nil
}

func (sim *CpuSim) WritePort(port Address, value byte) error {
	for _, p := range sim.Ports {
		if p.HasAddress(port) {
			return p.Write(port, value)
		}
	}
	return nil
}
