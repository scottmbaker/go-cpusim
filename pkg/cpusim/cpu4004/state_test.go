package cpu4004

import (
	"testing"

	"github.com/scottmbaker/gocpusim/pkg/cpusim"
)

func TestCPU4004SaveLoadState(t *testing.T) {
	sim := cpusim.NewCPUSim()
	cpu := New4004(sim, "4004")
	cpu.PC = 0x123
	cpu.SP = 2
	cpu.RC = 0xAB
	cpu.Registers[0] = 0x0F
	cpu.Registers[16] = 0x09 // accumulator
	cpu.Stack[0] = 0x456
	cpu.Cycles = 1000
	cpu.Halted = true

	data := cpu.SaveState()

	cpu2 := New4004(sim, "4004")
	if err := cpu2.LoadState(data); err != nil {
		t.Fatal(err)
	}
	if cpu2.PC != 0x123 {
		t.Errorf("PC mismatch: got %04X", cpu2.PC)
	}
	if cpu2.SP != 2 {
		t.Errorf("SP mismatch: got %d", cpu2.SP)
	}
	if cpu2.RC != 0xAB {
		t.Errorf("RC mismatch: got %02X", cpu2.RC)
	}
	if cpu2.Registers[0] != 0x0F {
		t.Errorf("R0 mismatch: got %02X", cpu2.Registers[0])
	}
	if cpu2.Registers[16] != 0x09 {
		t.Errorf("ACC mismatch: got %02X", cpu2.Registers[16])
	}
	if cpu2.Stack[0] != 0x456 {
		t.Errorf("Stack[0] mismatch: got %04X", cpu2.Stack[0])
	}
	if cpu2.Cycles != 1000 {
		t.Errorf("Cycles mismatch: got %d", cpu2.Cycles)
	}
	if !cpu2.Halted {
		t.Error("Halted should be true")
	}
}

func TestBus8BitSaveLoadState(t *testing.T) {
	sim := cpusim.NewCPUSim()
	bus := NewBus8Bit(sim, "bus", &cpusim.AlwaysEnabled)
	bus.LastReadValue = 0x42
	bus.LastWriteValue = 0xAB

	data := bus.SaveState()

	bus2 := NewBus8Bit(sim, "bus", &cpusim.AlwaysEnabled)
	if err := bus2.LoadState(data); err != nil {
		t.Fatal(err)
	}
	if bus2.LastReadValue != 0x42 {
		t.Errorf("LastReadValue mismatch: got %02X", bus2.LastReadValue)
	}
	if bus2.LastWriteValue != 0xAB {
		t.Errorf("LastWriteValue mismatch: got %02X", bus2.LastWriteValue)
	}
}
