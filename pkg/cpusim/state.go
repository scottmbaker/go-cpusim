package cpusim

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

// gobEncode serializes a value using gob. Panics on failure, which would
// indicate a bug (e.g. an unencodable type) rather than a runtime condition.
func gobEncode(v any) []byte {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(v); err != nil {
		panic("state: gob encode: " + err.Error())
	}
	return buf.Bytes()
}

// gobDecode deserializes gob-encoded data into the value pointed to by dst.
func gobDecode(data []byte, dst any) error {
	return gob.NewDecoder(bytes.NewReader(data)).Decode(dst)
}

// --- Memory ---

type memoryState struct {
	Contents       []byte
	StatusContents [][]byte
}

func (mem *Memory) SaveState() []byte {
	return gobEncode(memoryState{
		Contents:       mem.Contents,
		StatusContents: mem.StatusContents,
	})
}

func (mem *Memory) LoadState(data []byte) error {
	var s memoryState
	if err := gobDecode(data, &s); err != nil {
		return fmt.Errorf("memory LoadState: %w", err)
	}
	if len(s.Contents) != len(mem.Contents) {
		return fmt.Errorf("memory LoadState: contents size mismatch (got %d, want %d)", len(s.Contents), len(mem.Contents))
	}
	copy(mem.Contents, s.Contents)
	mem.StatusContents = s.StatusContents
	return nil
}

// --- UART ---

type uartState struct {
	Keybuffer   []byte
	LastCharOut byte
}

func (u *UART) SaveState() []byte {
	u.mu.Lock()
	defer u.mu.Unlock()
	return gobEncode(uartState{
		Keybuffer:   u.Keybuffer,
		LastCharOut: u.lastCharOut,
	})
}

func (u *UART) LoadState(data []byte) error {
	var s uartState
	if err := gobDecode(data, &s); err != nil {
		return fmt.Errorf("uart LoadState: %w", err)
	}
	u.mu.Lock()
	defer u.mu.Unlock()
	u.Keybuffer = s.Keybuffer
	u.lastCharOut = s.LastCharOut
	return nil
}

// --- ACIA ---

type aciaState struct {
	Keybuffer   []byte
	LastCharOut byte
	ControlReg  byte
}

func (a *ACIA) SaveState() []byte {
	a.mu.Lock()
	defer a.mu.Unlock()
	return gobEncode(aciaState{
		Keybuffer:   a.Keybuffer,
		LastCharOut: a.lastCharOut,
		ControlReg:  a.controlReg,
	})
}

func (a *ACIA) LoadState(data []byte) error {
	var s aciaState
	if err := gobDecode(data, &s); err != nil {
		return fmt.Errorf("acia LoadState: %w", err)
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Keybuffer = s.Keybuffer
	a.lastCharOut = s.LastCharOut
	a.controlReg = s.ControlReg
	return nil
}

// --- SIO ---

type sioChannelState struct {
	WriteRegs [8]byte
	ReadRegs  [3]byte
	RegPtr    byte
}

type sioState struct {
	Keybuffer   []byte
	LastCharOut byte
	ChanA       sioChannelState
	ChanB       sioChannelState
}

func (s *SIO) SaveState() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	return gobEncode(sioState{
		Keybuffer:   s.Keybuffer,
		LastCharOut: s.lastCharOut,
		ChanA: sioChannelState{
			WriteRegs: s.chanA.writeRegs,
			ReadRegs:  s.chanA.readRegs,
			RegPtr:    s.chanA.regPtr,
		},
		ChanB: sioChannelState{
			WriteRegs: s.chanB.writeRegs,
			ReadRegs:  s.chanB.readRegs,
			RegPtr:    s.chanB.regPtr,
		},
	})
}

func (s *SIO) LoadState(data []byte) error {
	var st sioState
	if err := gobDecode(data, &st); err != nil {
		return fmt.Errorf("sio LoadState: %w", err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Keybuffer = st.Keybuffer
	s.lastCharOut = st.LastCharOut
	s.chanA.writeRegs = st.ChanA.WriteRegs
	s.chanA.readRegs = st.ChanA.ReadRegs
	s.chanA.regPtr = st.ChanA.RegPtr
	s.chanB.writeRegs = st.ChanB.WriteRegs
	s.chanB.readRegs = st.ChanB.ReadRegs
	s.chanB.regPtr = st.ChanB.RegPtr
	return nil
}

// --- CompactFlash ---

// Note: Does not persist the contents of the CF, just the state of the interface.

type cfState struct {
	Data      [512]byte
	Dptr      int
	DataLatch byte
	State     int
	Length    int
	Error     byte
	Feature   byte
	Count     byte
	Lba1      byte
	Lba2      byte
	Lba3      byte
	Lba4      byte
	Status    byte
	Devctrl   byte
	EightBit  bool
}

func (cf *CompactFlash) SaveState() []byte {
	return gobEncode(cfState{
		Data: cf.data, Dptr: cf.dptr, DataLatch: cf.dataLatch,
		State: cf.state, Length: cf.length,
		Error: cf.error, Feature: cf.feature, Count: cf.count,
		Lba1: cf.lba1, Lba2: cf.lba2, Lba3: cf.lba3, Lba4: cf.lba4,
		Status: cf.status, Devctrl: cf.devctrl, EightBit: cf.eightbit,
	})
}

func (cf *CompactFlash) LoadState(data []byte) error {
	var s cfState
	if err := gobDecode(data, &s); err != nil {
		return fmt.Errorf("cf LoadState: %w", err)
	}
	cf.data = s.Data
	cf.dptr = s.Dptr
	cf.dataLatch = s.DataLatch
	cf.state = s.State
	cf.length = s.Length
	cf.error = s.Error
	cf.feature = s.Feature
	cf.count = s.Count
	cf.lba1 = s.Lba1
	cf.lba2 = s.Lba2
	cf.lba3 = s.Lba3
	cf.lba4 = s.Lba4
	cf.status = s.Status
	cf.devctrl = s.Devctrl
	cf.eightbit = s.EightBit
	return nil
}

// --- DipSwitch ---

type dipSwitchState struct {
	Value byte
}

func (d *DipSwitch) SaveState() []byte {
	return gobEncode(dipSwitchState{Value: d.Value})
}

func (d *DipSwitch) LoadState(data []byte) error {
	var s dipSwitchState
	if err := gobDecode(data, &s); err != nil {
		return fmt.Errorf("dipswitch LoadState: %w", err)
	}
	d.Value = s.Value
	return nil
}

// --- GenericOutputPort ---

type outPortState struct {
	Value byte
}

func (d *GenericOutputPort) SaveState() []byte {
	return gobEncode(outPortState{Value: d.Value})
}

func (d *GenericOutputPort) LoadState(data []byte) error {
	var s outPortState
	if err := gobDecode(data, &s); err != nil {
		return fmt.Errorf("outport LoadState: %w", err)
	}
	d.Value = s.Value
	d.UpdateEnableOut()
	return nil
}

// --- Map670 ---

type map670State struct {
	Contents [16]byte
}

func (m *Map670) SaveState() []byte {
	return gobEncode(map670State{Contents: m.Contents})
}

func (m *Map670) LoadState(data []byte) error {
	var s map670State
	if err := gobDecode(data, &s); err != nil {
		return fmt.Errorf("map670 LoadState: %w", err)
	}
	m.Contents = s.Contents
	return nil
}

// --- Map173 ---

type map173State struct {
	Contents byte
}

func (m *Map173) SaveState() []byte {
	return gobEncode(map173State{Contents: m.Contents})
}

func (m *Map173) LoadState(data []byte) error {
	var s map173State
	if err := gobDecode(data, &s); err != nil {
		return fmt.Errorf("map173 LoadState: %w", err)
	}
	m.Contents = s.Contents
	return nil
}

// --- EnableBit ---

type enableBitState struct {
	Value bool
}

func (e *EnableBit) SaveState() []byte {
	return gobEncode(enableBitState{Value: e.Value})
}

func (e *EnableBit) LoadState(data []byte) error {
	var s enableBitState
	if err := gobDecode(data, &s); err != nil {
		return fmt.Errorf("enablebit LoadState: %w", err)
	}
	e.Value = s.Value
	return nil
}

// --- CpuSim (whole-simulator envelope) ---

type simDeviceState struct {
	Category string // "cpu", "memory", "port", "mapper"
	Index    int
	Data     []byte
}

type simState struct {
	Devices []simDeviceState
}

func (sim *CpuSim) SaveState() []byte {
	var ss simState

	// saveDevice appends one device's state, keyed by its original index in
	// the simulator's slice so LoadState maps back to the correct device.
	saveDevice := func(category string, index int, device any) {
		if sd, ok := device.(StatefulDevice); ok {
			ss.Devices = append(ss.Devices, simDeviceState{
				Category: category,
				Index:    index,
				Data:     sd.SaveState(),
			})
		}
	}

	for i, c := range sim.CPU {
		saveDevice("cpu", i, c)
	}
	for i, m := range sim.Memory {
		saveDevice("memory", i, m)
	}
	for i, p := range sim.Ports {
		saveDevice("port", i, p)
	}
	for i, m := range sim.Mappers {
		saveDevice("mapper", i, m)
	}

	return gobEncode(ss)
}

func (sim *CpuSim) LoadState(data []byte) error {
	var ss simState
	if err := gobDecode(data, &ss); err != nil {
		return fmt.Errorf("sim LoadState: %w", err)
	}

	for _, ds := range ss.Devices {
		var device any
		switch ds.Category {
		case "cpu":
			if ds.Index >= len(sim.CPU) {
				return fmt.Errorf("sim LoadState: cpu index %d out of range", ds.Index)
			}
			device = sim.CPU[ds.Index]
		case "memory":
			if ds.Index >= len(sim.Memory) {
				return fmt.Errorf("sim LoadState: memory index %d out of range", ds.Index)
			}
			device = sim.Memory[ds.Index]
		case "port":
			if ds.Index >= len(sim.Ports) {
				return fmt.Errorf("sim LoadState: port index %d out of range", ds.Index)
			}
			device = sim.Ports[ds.Index]
		case "mapper":
			if ds.Index >= len(sim.Mappers) {
				return fmt.Errorf("sim LoadState: mapper index %d out of range", ds.Index)
			}
			device = sim.Mappers[ds.Index]
		default:
			return fmt.Errorf("sim LoadState: unknown category %q", ds.Category)
		}
		if sd, ok := device.(StatefulDevice); ok {
			if err := sd.LoadState(ds.Data); err != nil {
				return err
			}
		}
	}
	return nil
}
