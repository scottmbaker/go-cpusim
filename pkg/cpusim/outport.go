package cpusim

type GenericOutputPort struct {
	Sim                *CpuSim
	Name               string
	dataWriteAddress   Address
	ConnectedEnableBit [8]*EnableBit
	Enabler            EnablerInterface
	Value              byte
}

func (d *GenericOutputPort) GetName() string {
	return d.Name
}

func (d *GenericOutputPort) HasAddress(address Address) bool {
	if !d.Enabler.Bool() {
		return false
	}
	return (address == d.dataWriteAddress)
}

func (d *GenericOutputPort) Read(address Address) (byte, error) {
	return 0, &ErrWriteOnly{Device: d}
}

func (d *GenericOutputPort) UpdateEnableOut() {
	for i := 0; i < 8; i++ {
		bitIsSet := (d.Value & (1 << i)) != 0
		if d.ConnectedEnableBit[i] != nil {
			d.ConnectedEnableBit[i].Set(bitIsSet)
		}
	}
}

func (d *GenericOutputPort) Write(address Address, value byte) error {
	if address == d.dataWriteAddress {
		d.Value = value
		d.UpdateEnableOut()
		return nil
	}
	return &ErrInvalidAddress{Device: d, Address: address}
}

func (d *GenericOutputPort) WriteStatus(address Address, statusAddr Address, value byte) error {
	_ = address
	_ = statusAddr
	_ = value
	return &ErrNotImplemented{Device: d}
}

func (d *GenericOutputPort) ReadStatus(address Address, statusAddr Address) (byte, error) {
	_ = address
	_ = statusAddr
	return 0, &ErrReadOnly{Device: d}
}

func (d *GenericOutputPort) ConnectEnableBit(bit int, enableBit *EnableBit) {
	d.ConnectedEnableBit[bit] = enableBit
}

func (d *GenericOutputPort) GetKind() string {
	return KIND_GENERIC_OUTPORT
}

func NewGenericOutputPort(sim *CpuSim, name string, dataWriteAddress Address, value byte, enabler EnablerInterface) *GenericOutputPort {
	return &GenericOutputPort{
		Sim:              sim,
		Name:             name,
		dataWriteAddress: dataWriteAddress,
		Enabler:          enabler,
		Value:            value,
	}
}
