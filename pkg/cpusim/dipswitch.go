package cpusim

import ()

type DipSwitch struct {
	Sim             *CpuSim
	Name            string
	DataReadAddress uint16
	Enabler         EnablerInterface
	Value           byte
}

func (d *DipSwitch) GetName() string {
	return d.Name
}

func (d *DipSwitch) HasAddress(address uint16) bool {
	if !d.Enabler.Bool() {
		return false
	}
	return (address == d.DataReadAddress)
}

func (d *DipSwitch) Read(address uint16) (byte, error) {
	if address == d.DataReadAddress {
		return d.Value, nil
	}

	return 0, &ErrInvalidAddress{Device: d, Address: address}
}

func (d *DipSwitch) Write(address uint16, value byte) error {
	return &ErrReadOnly{Device: d}
}

func NewDipSwitch(sim *CpuSim, name string, dataReadAddress uint16, value byte, enabler EnablerInterface) *DipSwitch {
	return &DipSwitch{
		Sim:             sim,
		Name:            name,
		DataReadAddress: dataReadAddress,
		Enabler:         enabler,
		Value:           value,
	}
}
