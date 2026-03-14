package cpuz80

import (
	"testing"

	"github.com/scottmbaker/gocpusim/pkg/cpusim"
)

func TestCPUZ80SaveLoadState(t *testing.T) {
	sim := cpusim.NewCPUSim()
	cpu := NewZ80(sim, "z80")
	cpu.A = 0xFF
	cpu.F = 0xD7
	cpu.B = 0x12
	cpu.C = 0x34
	cpu.D = 0x56
	cpu.E = 0x78
	cpu.H = 0x9A
	cpu.L = 0xBC
	cpu.AF_ = 0x1122
	cpu.BC_ = 0x3344
	cpu.DE_ = 0x5566
	cpu.HL_ = 0x7788
	cpu.IX = 0xAABB
	cpu.IY = 0xCCDD
	cpu.SP = 0xEEFF
	cpu.PC = 0x4000
	cpu.I = 0x3E
	cpu.R = 0x42
	cpu.IFF1 = true
	cpu.IFF2 = true
	cpu.IM = 1
	cpu.EIPending = true
	cpu.Halted = true
	cpu.WZ = 0x1234
	cpu.Q = 0x0F
	cpu.PrevQ = 0xF0

	data := cpu.SaveState()

	cpu2 := NewZ80(sim, "z80")
	if err := cpu2.LoadState(data); err != nil {
		t.Fatal(err)
	}

	if cpu2.A != 0xFF || cpu2.F != 0xD7 {
		t.Errorf("AF mismatch: A=%02X F=%02X", cpu2.A, cpu2.F)
	}
	if cpu2.B != 0x12 || cpu2.C != 0x34 {
		t.Errorf("BC mismatch: B=%02X C=%02X", cpu2.B, cpu2.C)
	}
	if cpu2.D != 0x56 || cpu2.E != 0x78 {
		t.Errorf("DE mismatch")
	}
	if cpu2.H != 0x9A || cpu2.L != 0xBC {
		t.Errorf("HL mismatch")
	}
	if cpu2.AF_ != 0x1122 || cpu2.BC_ != 0x3344 || cpu2.DE_ != 0x5566 || cpu2.HL_ != 0x7788 {
		t.Error("alternate registers mismatch")
	}
	if cpu2.IX != 0xAABB || cpu2.IY != 0xCCDD {
		t.Error("index registers mismatch")
	}
	if cpu2.SP != 0xEEFF || cpu2.PC != 0x4000 {
		t.Error("SP/PC mismatch")
	}
	if cpu2.I != 0x3E || cpu2.R != 0x42 {
		t.Error("I/R mismatch")
	}
	if !cpu2.IFF1 || !cpu2.IFF2 || cpu2.IM != 1 {
		t.Error("interrupt state mismatch")
	}
	if !cpu2.EIPending {
		t.Error("EIPending should be true")
	}
	if !cpu2.Halted {
		t.Error("Halted should be true")
	}
	if cpu2.WZ != 0x1234 {
		t.Errorf("WZ mismatch: got %04X", cpu2.WZ)
	}
	if cpu2.Q != 0x0F || cpu2.PrevQ != 0xF0 {
		t.Error("Q/PrevQ mismatch")
	}
}
