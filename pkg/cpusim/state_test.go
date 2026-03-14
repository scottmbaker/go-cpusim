package cpusim

import (
	"testing"
)

func TestMemorySaveLoadState(t *testing.T) {
	sim := NewCPUSim()
	mem := NewMemory(sim, "test", KIND_RAM, 0, 0xFF, 8, false, &AlwaysEnabled)
	mem.Contents[0] = 0x42
	mem.Contents[0xFF] = 0xAB

	data := mem.SaveState()

	mem2 := NewMemory(sim, "test", KIND_RAM, 0, 0xFF, 8, false, &AlwaysEnabled)
	if err := mem2.LoadState(data); err != nil {
		t.Fatal(err)
	}
	if mem2.Contents[0] != 0x42 || mem2.Contents[0xFF] != 0xAB {
		t.Errorf("contents mismatch: got [0]=%02X [FF]=%02X", mem2.Contents[0], mem2.Contents[0xFF])
	}
}

func TestMemoryWithStatusSaveLoadState(t *testing.T) {
	sim := NewCPUSim()
	mem := NewMemory(sim, "test", KIND_RAM, 0, 0xFF, 8, false, &AlwaysEnabled)
	mem.CreateStatusBytes(4, 4)
	mem.StatusContents[0][0] = 0x12
	mem.StatusContents[3][3] = 0x34

	data := mem.SaveState()

	mem2 := NewMemory(sim, "test", KIND_RAM, 0, 0xFF, 8, false, &AlwaysEnabled)
	if err := mem2.LoadState(data); err != nil {
		t.Fatal(err)
	}
	if mem2.StatusContents[0][0] != 0x12 || mem2.StatusContents[3][3] != 0x34 {
		t.Error("status contents mismatch")
	}
}

func TestMemorySizeMismatch(t *testing.T) {
	sim := NewCPUSim()
	mem := NewMemory(sim, "test", KIND_RAM, 0, 0xFF, 8, false, &AlwaysEnabled)
	data := mem.SaveState()

	mem2 := NewMemory(sim, "test", KIND_RAM, 0, 0x7F, 7, false, &AlwaysEnabled)
	if err := mem2.LoadState(data); err == nil {
		t.Error("expected size mismatch error")
	}
}

func TestUARTSaveLoadState(t *testing.T) {
	sim := NewCPUSim()
	ch := NewChannelSerial()
	uart := NewUART(sim, ch, "uart", 0, 0, 1, 1, &AlwaysEnabled)
	uart.Keybuffer = []byte{0x41, 0x42}
	uart.lastCharOut = 0x43

	data := uart.SaveState()

	uart2 := NewUART(sim, ch, "uart", 0, 0, 1, 1, &AlwaysEnabled)
	if err := uart2.LoadState(data); err != nil {
		t.Fatal(err)
	}
	if len(uart2.Keybuffer) != 2 || uart2.Keybuffer[0] != 0x41 || uart2.Keybuffer[1] != 0x42 {
		t.Error("keybuffer mismatch")
	}
	if uart2.lastCharOut != 0x43 {
		t.Error("lastCharOut mismatch")
	}
}

func TestACIASaveLoadState(t *testing.T) {
	sim := NewCPUSim()
	ch := NewChannelSerial()
	acia := NewACIA(sim, ch, "acia", 0, 1, &AlwaysEnabled)
	acia.Keybuffer = []byte{0x10}
	acia.controlReg = 0x15

	data := acia.SaveState()

	acia2 := NewACIA(sim, ch, "acia", 0, 1, &AlwaysEnabled)
	if err := acia2.LoadState(data); err != nil {
		t.Fatal(err)
	}
	if len(acia2.Keybuffer) != 1 || acia2.Keybuffer[0] != 0x10 {
		t.Error("keybuffer mismatch")
	}
	if acia2.controlReg != 0x15 {
		t.Error("controlReg mismatch")
	}
}

func TestSIOSaveLoadState(t *testing.T) {
	sim := NewCPUSim()
	ch := NewChannelSerial()
	sio := NewSIO(sim, ch, "sio", 0, 1, 2, 3, &AlwaysEnabled)
	sio.Keybuffer = []byte{0x20, 0x21}
	sio.chanA.writeRegs[1] = 0xAA
	sio.chanB.regPtr = 3

	data := sio.SaveState()

	sio2 := NewSIO(sim, ch, "sio", 0, 1, 2, 3, &AlwaysEnabled)
	if err := sio2.LoadState(data); err != nil {
		t.Fatal(err)
	}
	if sio2.chanA.writeRegs[1] != 0xAA {
		t.Error("chanA writeRegs mismatch")
	}
	if sio2.chanB.regPtr != 3 {
		t.Error("chanB regPtr mismatch")
	}
}

func TestCompactFlashSaveLoadState(t *testing.T) {
	sim := NewCPUSim()
	cf := NewCompactFlash(sim, "cf", 0x10, &AlwaysEnabled)
	cf.lba1 = 0x12
	cf.status = 0x50
	cf.dataLatch = 0xAB
	cf.eightbit = true

	data := cf.SaveState()

	cf2 := NewCompactFlash(sim, "cf", 0x10, &AlwaysEnabled)
	if err := cf2.LoadState(data); err != nil {
		t.Fatal(err)
	}
	if cf2.lba1 != 0x12 {
		t.Error("lba1 mismatch")
	}
	if cf2.status != 0x50 {
		t.Error("status mismatch")
	}
	if cf2.dataLatch != 0xAB {
		t.Error("dataLatch mismatch")
	}
	if !cf2.eightbit {
		t.Error("eightbit mismatch")
	}
}

func TestDipSwitchSaveLoadState(t *testing.T) {
	sim := NewCPUSim()
	ds := NewDipSwitch(sim, "ds", 0, 0xAA, &AlwaysEnabled)

	data := ds.SaveState()

	ds2 := NewDipSwitch(sim, "ds", 0, 0, &AlwaysEnabled)
	if err := ds2.LoadState(data); err != nil {
		t.Fatal(err)
	}
	if ds2.Value != 0xAA {
		t.Errorf("value mismatch: got %02X", ds2.Value)
	}
}

func TestGenericOutputPortSaveLoadState(t *testing.T) {
	sim := NewCPUSim()
	eb := NewEnableBit()
	port := NewGenericOutputPort(sim, "port", 0, 0, &AlwaysEnabled)
	port.ConnectEnableBit(0, eb)
	port.Value = 0x01
	port.UpdateEnableOut()

	data := port.SaveState()

	eb2 := NewEnableBit()
	port2 := NewGenericOutputPort(sim, "port", 0, 0, &AlwaysEnabled)
	port2.ConnectEnableBit(0, eb2)
	if err := port2.LoadState(data); err != nil {
		t.Fatal(err)
	}
	if port2.Value != 0x01 {
		t.Error("value mismatch")
	}
	if !eb2.Value {
		t.Error("enable bit not restored via UpdateEnableOut")
	}
}

func TestMap670SaveLoadState(t *testing.T) {
	sim := NewCPUSim()
	m := New74670(sim, "mapper", 0, 14, 0, 16, 17, 18, 19, &AlwaysEnabled, &AlwaysEnabled)
	m.Contents[0] = 0x0F
	m.Contents[3] = 0xAB

	data := m.SaveState()

	m2 := New74670(sim, "mapper", 0, 14, 0, 16, 17, 18, 19, &AlwaysEnabled, &AlwaysEnabled)
	if err := m2.LoadState(data); err != nil {
		t.Fatal(err)
	}
	if m2.Contents[0] != 0x0F || m2.Contents[3] != 0xAB {
		t.Error("contents mismatch")
	}
}

func TestMap173SaveLoadState(t *testing.T) {
	sim := NewCPUSim()
	m := New74173(sim, "mapper", 0, 16, 17, 18, 19, &AlwaysEnabled)
	m.Contents = 0x0A

	data := m.SaveState()

	m2 := New74173(sim, "mapper", 0, 16, 17, 18, 19, &AlwaysEnabled)
	if err := m2.LoadState(data); err != nil {
		t.Fatal(err)
	}
	if m2.Contents != 0x0A {
		t.Errorf("contents mismatch: got %02X", m2.Contents)
	}
}

func TestEnableBitSaveLoadState(t *testing.T) {
	eb := NewEnableBit()
	eb.Set(true)

	data := eb.SaveState()

	eb2 := NewEnableBit()
	if err := eb2.LoadState(data); err != nil {
		t.Fatal(err)
	}
	if !eb2.Value {
		t.Error("value should be true")
	}
}

func TestSimSaveLoadState(t *testing.T) {
	sim := NewCPUSim()
	mem := NewMemory(sim, "ram", KIND_RAM, 0, 0xFF, 8, false, &AlwaysEnabled)
	mem.Contents[0] = 0x42
	sim.AddMemory(mem)

	ch := NewChannelSerial()
	uart := NewUART(sim, ch, "uart", 0x100, 0x100, 0x101, 0x101, &AlwaysEnabled)
	uart.Keybuffer = []byte{0x41}
	sim.AddPort(uart)

	data := sim.SaveState()

	// Create a fresh sim with same structure
	sim2 := NewCPUSim()
	mem2 := NewMemory(sim2, "ram", KIND_RAM, 0, 0xFF, 8, false, &AlwaysEnabled)
	sim2.AddMemory(mem2)

	ch2 := NewChannelSerial()
	uart2 := NewUART(sim2, ch2, "uart", 0x100, 0x100, 0x101, 0x101, &AlwaysEnabled)
	sim2.AddPort(uart2)

	if err := sim2.LoadState(data); err != nil {
		t.Fatal(err)
	}

	if mem2.Contents[0] != 0x42 {
		t.Errorf("memory not restored: got %02X", mem2.Contents[0])
	}
	if len(uart2.Keybuffer) != 1 || uart2.Keybuffer[0] != 0x41 {
		t.Error("uart keybuffer not restored")
	}
}
