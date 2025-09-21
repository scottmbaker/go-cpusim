package cpusim

type CpuInterface interface {
	SetReg(register int, value byte) error
	GetReg(register int) (byte, error)
	String() string
	Run() error
}

type MemoryInterface interface {
	HasAddress(address Address) bool
	Read(address Address) (byte, error)
	Write(address Address, value byte) error
	ReadStatus(address Address, statusAddr Address) (byte, error)      // for 4004
	WriteStatus(address Address, statusAddr Address, value byte) error // for 4004
}

type MapperInterface interface {
	Map(address Address) (Address, error)
}
