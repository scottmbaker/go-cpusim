package cpu8008

import (
	"testing"

	"github.com/scottmbaker/gocpusim/pkg/cpusim"
)

func TestCPU8008SaveLoadState(t *testing.T) {
	sim := cpusim.NewCPUSim()
	cpu := New8008(sim, "8008")
	cpu.PC = 0x1234
	cpu.SP = 3
	cpu.Registers[REG_A] = 0xFF
	cpu.Registers[FLAG_CARRY] = 1
	cpu.Stack[0] = 0x5678
	cpu.Halted = true

	data := cpu.SaveState()

	cpu2 := New8008(sim, "8008")
	if err := cpu2.LoadState(data); err != nil {
		t.Fatal(err)
	}
	if cpu2.PC != 0x1234 {
		t.Errorf("PC mismatch: got %04X", cpu2.PC)
	}
	if cpu2.SP != 3 {
		t.Errorf("SP mismatch: got %d", cpu2.SP)
	}
	if cpu2.Registers[REG_A] != 0xFF {
		t.Errorf("A mismatch: got %02X", cpu2.Registers[REG_A])
	}
	if cpu2.Registers[FLAG_CARRY] != 1 {
		t.Error("carry flag mismatch")
	}
	if cpu2.Stack[0] != 0x5678 {
		t.Errorf("Stack[0] mismatch: got %04X", cpu2.Stack[0])
	}
	if !cpu2.Halted {
		t.Error("Halted should be true")
	}
}
