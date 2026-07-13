package cpusim

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
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
	A16 = 16
	A17 = 17
	A18 = 18
	A19 = 19

	D0 = 0
	D1 = 1
	D2 = 2
	D3 = 3
	D4 = 4
	D5 = 5
	D6 = 6
	D7 = 7

	KIND_RAM                  = "RAM"
	KIND_ROM                  = "ROM"
	KIND_MAPPER               = "MAPPER"
	KIND_UART                 = "UART"
	KIND_ACIA                 = "ACIA"
	KIND_SIO                  = "SIO"
	KIND_ASCI                 = "ASCI"
	KIND_SCC                  = "SCC"
	KIND_INPORT               = "INPORT"
	KIND_ROMPORT              = "ROMPORT"
	KIND_RAMPORT              = "RAMPORT"
	KIND_CF                   = "CF"
	KIND_FDC                  = "FDC"
	KIND_GENERIC_OUTPORT      = "GENERIC_OUTPORT"
	KIND_SP0256_SPEECH_DEVICE = "SP0256A-AL2"
	KIND_16550                = "16550"
)

type CpuSim struct {
	CPU             []CpuInterface
	Memory          []MemoryInterface
	Ports           []MemoryInterface
	Mappers         []MapperInterface
	Throttle        *Throttle
	IOPollDelay     time.Duration // sleep this long when a UART status poll finds no data; 0 = disabled
	emptyPolls      atomic.Int32
	CtrlC           atomic.Bool
	HostCtrlC       bool // when true, a 0x03 byte from serial input halts the sim (host escape); disable to deliver Ctrl-C to the emulated machine
	HaltWaitsForInt bool // when true, HALT with interrupts enabled idles waiting for an interrupt; default false ends Run() on HALT (historical behavior)
	intMu           sync.Mutex
	intSources      map[string]bool
	intLine         atomic.Bool
	Debug           bool
	MemDebug        bool
	MemoryFilter    string
	PortFilter      string
}

func NewCPUSim() *CpuSim {
	return &CpuSim{
		CPU:       make([]CpuInterface, 0),
		Memory:    make([]MemoryInterface, 0),
		Ports:     make([]MemoryInterface, 0),
		Throttle:  NewThrottle(0), // no throttling by default
		Debug:     true,
		HostCtrlC: true, // preserve historical stdio behavior
	}
}

func (sim *CpuSim) SetIPS(ips int64) {
	sim.Throttle = NewThrottle(ips)
}

// IOActivity resets the empty-poll counter. Any UART that has data available
// or is transmitting should call this so that activity on one UART prevents
// idle UARTs from triggering poll delays.
func (sim *CpuSim) IOActivity() {
	sim.emptyPolls.Store(0)
}

// IOPoll records an empty status-register read. On the second consecutive
// empty poll it sleeps for IOPollDelay, reducing host CPU usage when the
// emulated program is spin-waiting for input. The first empty poll is free
// so that a single status check during a TX loop doesn't stall output.
// Callers must release their mutex before calling this, as it may sleep.
func (sim *CpuSim) IOPoll() {
	if sim.IOPollDelay <= 0 {
		return
	}
	if sim.emptyPolls.Add(1) > 1 {
		time.Sleep(sim.IOPollDelay)
	}
}

// SetInt asserts or releases the shared maskable-interrupt line (/INT) on
// behalf of the named device. The line is the OR of all asserting sources,
// like the open-drain /INT wire on a real bus. Level-triggered: a device
// should keep its source asserted while its interrupt condition holds.
func (sim *CpuSim) SetInt(source string, asserted bool) {
	sim.intMu.Lock()
	defer sim.intMu.Unlock()
	if sim.intSources == nil {
		sim.intSources = make(map[string]bool)
	}
	if asserted {
		sim.intSources[source] = true
	} else {
		delete(sim.intSources, source)
	}
	sim.intLine.Store(len(sim.intSources) > 0)
}

// IntAsserted reports whether any device is asserting the /INT line.
func (sim *CpuSim) IntAsserted() bool {
	return sim.intLine.Load()
}

func (sim *CpuSim) SetDebug(debug bool) {
	sim.Debug = debug
}

func (sim *CpuSim) SetMemDebug(memDebug bool) {
	sim.MemDebug = memDebug
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

func (sim *CpuSim) Halt() {
	for _, cpu := range sim.CPU {
		cpu.Halt()
	}
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

func (sim *CpuSim) FilterMemoryKind(kind string) {
	sim.MemoryFilter = kind
}

func (sim *CpuSim) FilterPortKind(kind string) {
	sim.PortFilter = kind
}

func (sim *CpuSim) MatchMemory(mem MemoryInterface) bool {
	return sim.MemoryFilter == "" || mem.GetKind() == sim.MemoryFilter
}

func (sim *CpuSim) MatchPort(mem MemoryInterface) bool {
	return sim.PortFilter == "" || mem.GetKind() == sim.PortFilter
}

func (sim *CpuSim) WriteMemory(address Address, value byte) error {
	for _, mapper := range sim.Mappers {
		var err error
		if !mapper.MatchMemory(sim.Memory[0]) {
			continue
		}
		address, err = mapper.Map(address)
		if err != nil {
			return err
		}
	}
	for _, mem := range sim.Memory {
		if !sim.MatchMemory(mem) {
			continue
		}
		if mem.HasAddress(address) {
			return mem.Write(address, value)
		}
	}
	return nil
}

func (sim *CpuSim) ReadMemory(address Address) (byte, error) {
	for _, mapper := range sim.Mappers {
		var err error
		if !mapper.MatchMemory(sim.Memory[0]) {
			continue
		}
		address, err = mapper.Map(address)
		if err != nil {
			return 0, err
		}
	}
	for _, mem := range sim.Memory {
		if !sim.MatchMemory(mem) {
			continue
		}
		if mem.HasAddress(address) {
			return mem.Read(address)
		}
	}
	return 0, nil
}

func (sim *CpuSim) WriteMemoryStatus(address Address, statusAddr Address, value byte) error {
	for _, mem := range sim.Memory {
		if !sim.MatchMemory(mem) {
			continue
		}
		if mem.HasAddress(address) {
			return mem.WriteStatus(address, statusAddr, value)
		}
	}
	return nil
}

func (sim *CpuSim) ReadMemoryStatus(address Address, statusAddr Address) (byte, error) {
	for _, mem := range sim.Memory {
		if !sim.MatchMemory(mem) {
			continue
		}
		if mem.HasAddress(address) {
			return mem.ReadStatus(address, statusAddr)
		}
	}
	return 0, nil
}

func (sim *CpuSim) ReadPort(port Address) (byte, error) {
	for _, p := range sim.Ports {
		if !sim.MatchPort(p) {
			continue
		}
		if p.HasAddress(port) {
			return p.Read(port)
		}
	}
	return 0, nil
}

func (sim *CpuSim) WritePort(port Address, value byte) error {
	for _, p := range sim.Ports {
		if !sim.MatchPort(p) {
			continue
		}
		if p.HasAddress(port) {
			return p.Write(port, value)
		}
	}
	return nil
}
