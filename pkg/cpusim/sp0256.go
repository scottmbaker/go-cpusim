package cpusim

import "fmt"

type Sp0SpeechDevice struct {
	Sim                *CpuSim
	Name               string
	dataWriteAddress   Address
	ConnectedEnableBit [8]*EnableBit
	Enabler            EnablerInterface
	ReadyValue         byte
	NotReadyValue      byte
}

var phones = []string{
	"PA1", "PA2", "PA3", "PA4", "PA5", "OY", "AY", "EH",
	"KK3", "PP", "JH", "NN1", "IH", "TT2", "RR1", "AX",
	"MM", "TT1", "DH1", "IY", "EY", "DD1", "UW1", "AO",
	"AA", "YY2", "AE", "HH1", "BB1", "TH", "UH", "UW2",
	"AW", "DD2", "GG3", "VV", "GG1", "SH", "ZH", "RR2",
	"FF", "KK2", "KK1", "ZZ", "NG", "LL", "WW", "XR",
	"WH", "YY1", "CH", "ER1", "ER2", "OW", "DH2", "SS",
	"NN2", "HH2", "OR", "AR", "YR", "GG2", "EL", "BB2",
	"", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "",
	"", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "",
	"", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "",
	"STOP",
}

func (d *Sp0SpeechDevice) GetPhoneme(value byte) string {
	if int(value) < len(phones) {
		return phones[value]
	}
	return ""
}

func (d *Sp0SpeechDevice) GetName() string {
	return d.Name
}

func (d *Sp0SpeechDevice) HasAddress(address Address) bool {
	if !d.Enabler.Bool() {
		return false
	}
	return (address == d.dataWriteAddress)
}

func (d *Sp0SpeechDevice) Read(address Address) (byte, error) {
	if address != d.dataWriteAddress {
		return 0, &ErrInvalidAddress{Device: d, Address: address}
	}
	//fmt.Printf("<%s> Read at address %04X, returning %02X\n", d.Name, address, d.ReadyValue)
	return d.ReadyValue, nil
}

func (d *Sp0SpeechDevice) Write(address Address, value byte) error {
	if address != d.dataWriteAddress {
		return &ErrInvalidAddress{Device: d, Address: address}
	}
	fmt.Printf("%s ", d.GetPhoneme(value))
	return nil
}

func (d *Sp0SpeechDevice) WriteStatus(address Address, statusAddr Address, value byte) error {
	_ = address
	_ = statusAddr
	_ = value
	return &ErrNotImplemented{Device: d}
}

func (d *Sp0SpeechDevice) ReadStatus(address Address, statusAddr Address) (byte, error) {
	_ = address
	_ = statusAddr
	return 0, &ErrReadOnly{Device: d}
}

func (d *Sp0SpeechDevice) GetKind() string {
	return KIND_SP0256_SPEECH_DEVICE
}

func NewSp0SpeechDevice(sim *CpuSim, name string, dataWriteAddress Address, enabler EnablerInterface) *Sp0SpeechDevice {
	return &Sp0SpeechDevice{
		Sim:              sim,
		Name:             name,
		dataWriteAddress: dataWriteAddress,
		Enabler:          enabler,
		ReadyValue:       0x01, // Assuming bit 0 indicates ready status
		NotReadyValue:    0x00, // Assuming 0 indicates not ready
	}
}
