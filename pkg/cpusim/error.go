package cpusim

import (
	"fmt"
)

type DeviceInterface interface {
	GetName() string
}

type ErrInvalidRegister struct {
	Register int
	Device   DeviceInterface
}

func (e *ErrInvalidRegister) Error() string {
	if e.Device != nil {
		return fmt.Sprintf("Device %s Invalid register: %d", e.Device.GetName(), e.Register)
	} else {
		return fmt.Sprintf("Invalid register: %d", e.Register)
	}
}

type ErrInvalidAddress struct {
	Address Address
	Device  DeviceInterface
}

func (e *ErrInvalidAddress) Error() string {
	if e.Device != nil {
		return fmt.Sprintf("Device %s Invalid address: %04X", e.Device.GetName(), e.Address)
	} else {
		return fmt.Sprintf("Invalid address: %04X", e.Address)
	}
}

type ErrReadOnly struct {
	Device DeviceInterface
}

func (e *ErrReadOnly) Error() string {
	if e.Device != nil {
		return fmt.Sprintf("Device %s is read-only", e.Device.GetName())
	} else {
		return "Device is read-only"
	}
}

type ErrInvalidOpcode struct {
	Device DeviceInterface
	Opcode byte
}

func (e *ErrInvalidOpcode) Error() string {
	if e.Device != nil {
		return fmt.Sprintf("Device %s Invalid opcode: %02X", e.Device.GetName(), e.Opcode)
	} else {
		return fmt.Sprintf("Invalid opcode: %02X", e.Opcode)
	}
}
