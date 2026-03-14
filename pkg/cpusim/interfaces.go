package cpusim

import "sync"

type CpuInterface interface {
	SetReg(register int, value byte) error
	GetReg(register int) (byte, error)
	String() string
	Run() error
	Halt()
}

type MemoryInterface interface {
	GetKind() string
	HasAddress(address Address) bool
	Read(address Address) (byte, error)
	Write(address Address, value byte) error
	ReadStatus(address Address, statusAddr Address) (byte, error)      // for 4004
	WriteStatus(address Address, statusAddr Address, value byte) error // for 4004
}

type MapperInterface interface {
	Map(address Address) (Address, error)
	MatchMemory(mem MemoryInterface) bool
}

type UartInterface interface {
	Start(wg *sync.WaitGroup)
	RestoreTerminal()
}

// StatefulDevice can save and restore its mutable state.
type StatefulDevice interface {
	SaveState() []byte
	LoadState([]byte) error
}

// SerialIO abstracts the byte-level I/O transport for serial devices.
type SerialIO interface {
	ReadByte() (byte, error)
	WriteByte(b byte) error
	Start()
	RestoreTerminal()
}
