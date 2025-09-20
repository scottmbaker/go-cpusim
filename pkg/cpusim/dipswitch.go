package cpusim

import ()

type DipSwitch struct {
	Sim             *CpuSim
	Name            string
	DataReadAddress Address
	Enabler         EnablerInterface
	Value           byte
}

func (d *DipSwitch) GetName() string {
	return d.Name
}

func (d *DipSwitch) HasAddress(address Address) bool {
	if !d.Enabler.Bool() {
		return false
	}
	return (address == d.DataReadAddress)
}

func (d *DipSwitch) Read(address Address) (byte, error) {
	if address == d.DataReadAddress {
		return d.Value, nil
	}

	return 0, &ErrInvalidAddress{Device: d, Address: address}
}

func (d *DipSwitch) Write(address Address, value byte) error {
	return &ErrReadOnly{Device: d}
}

func NewDipSwitch(sim *CpuSim, name string, dataReadAddress Address, value byte, enabler EnablerInterface) *DipSwitch {
	return &DipSwitch{
		Sim:             sim,
		Name:            name,
		DataReadAddress: dataReadAddress,
		Enabler:         enabler,
		Value:           value,
	}
}
