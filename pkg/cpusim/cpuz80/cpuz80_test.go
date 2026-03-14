package cpuz80

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/scottmbaker/gocpusim/pkg/cpusim"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testDataDir = "testdata/v1"

type TestPort struct {
	data map[cpusim.Address]byte
}

func NewTestPort() *TestPort {
	return &TestPort{data: make(map[cpusim.Address]byte)}
}

func (p *TestPort) HasAddress(address cpusim.Address) bool {
	return true
}

func (p *TestPort) Read(address cpusim.Address) (byte, error) {
	if val, ok := p.data[address]; ok {
		return val, nil
	}
	return 0, nil
}

func (p *TestPort) Write(address cpusim.Address, value byte) error {
	p.data[address] = value
	return nil
}

func (p *TestPort) ReadStatus(address cpusim.Address, statusAddr cpusim.Address) (byte, error) {
	return 0, nil
}

func (p *TestPort) WriteStatus(address cpusim.Address, statusAddr cpusim.Address, value byte) error {
	return nil
}

func (p *TestPort) GetKind() string {
	return "TESTPORT"
}

func (p *TestPort) GetName() string {
	return "testport"
}

type TestState struct {
	PC   uint16     `json:"pc"`
	SP   uint16     `json:"sp"`
	A    byte       `json:"a"`
	B    byte       `json:"b"`
	C    byte       `json:"c"`
	D    byte       `json:"d"`
	E    byte       `json:"e"`
	F    byte       `json:"f"`
	H    byte       `json:"h"`
	L    byte       `json:"l"`
	I    byte       `json:"i"`
	R    byte       `json:"r"`
	EI   byte       `json:"ei"`
	WZ   uint16     `json:"wz"`
	IX   uint16     `json:"ix"`
	IY   uint16     `json:"iy"`
	AF_  uint16     `json:"af_"`
	BC_  uint16     `json:"bc_"`
	DE_  uint16     `json:"de_"`
	HL_  uint16     `json:"hl_"`
	IM   byte       `json:"im"`
	IFF1 byte       `json:"iff1"`
	IFF2 byte       `json:"iff2"`
	P    byte       `json:"p"`
	Q    byte       `json:"q"`
	RAM  [][2]int   `json:"ram"`
}

type PortOp struct {
	Address int    `json:"-"`
	Value   int    `json:"-"`
	Dir     string `json:"-"`
}

func (p *PortOp) UnmarshalJSON(data []byte) error {
	var arr []interface{}
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}
	if len(arr) >= 3 {
		if v, ok := arr[0].(float64); ok {
			p.Address = int(v)
		}
		if v, ok := arr[1].(float64); ok {
			p.Value = int(v)
		}
		if v, ok := arr[2].(string); ok {
			p.Dir = v
		}
	}
	return nil
}

type TestCase struct {
	Name    string    `json:"name"`
	Initial TestState `json:"initial"`
	Final   TestState `json:"final"`
	Ports   []PortOp  `json:"ports"`
}

func setupCPU(tc *TestCase) (*CPUZ80, *cpusim.CpuSim, *cpusim.Memory, *TestPort) {
	sim := cpusim.NewCPUSim()
	sim.SetDebug(false)

	cpu := NewZ80(sim, "test-cpu")
	sim.AddCPU(cpu)

	ram := cpusim.NewMemory(sim, "ram", cpusim.KIND_RAM, 0x0000, 0xFFFF, 16, false, &cpusim.AlwaysEnabled)
	sim.AddMemory(ram)

	port := NewTestPort()
	sim.AddPort(port)

	// Set initial CPU state
	s := &tc.Initial
	cpu.A = s.A
	cpu.F = s.F
	cpu.B = s.B
	cpu.C = s.C
	cpu.D = s.D
	cpu.E = s.E
	cpu.H = s.H
	cpu.L = s.L
	cpu.I = s.I
	cpu.R = s.R
	cpu.PC = s.PC
	cpu.SP = s.SP
	cpu.IX = s.IX
	cpu.IY = s.IY
	cpu.AF_ = s.AF_
	cpu.BC_ = s.BC_
	cpu.DE_ = s.DE_
	cpu.HL_ = s.HL_
	cpu.WZ = s.WZ
	cpu.IM = s.IM
	cpu.IFF1 = s.IFF1 != 0
	cpu.IFF2 = s.IFF2 != 0
	cpu.Q = s.Q
	cpu.PrevQ = s.Q
	cpu.Halted = false

	// EI pending state
	if s.EI != 0 {
		cpu.EIPending = true
	}

	// Set initial RAM
	for _, entry := range s.RAM {
		addr := entry[0]
		val := entry[1]
		_ = ram.Write(cpusim.Address(addr), byte(val))
	}

	// Pre-load port read values
	for _, p := range tc.Ports {
		if p.Dir == "r" {
			port.data[cpusim.Address(p.Address)] = byte(p.Value)
		}
	}

	return cpu, sim, ram, port
}

func compareCPU(t *testing.T, tc *TestCase, cpu *CPUZ80, ram *cpusim.Memory, port *TestPort) {
	t.Helper()
	f := &tc.Final

	assert.Equal(t, f.A, cpu.A, "A register")
	assert.Equal(t, f.F, cpu.F, "F register (flags)")
	assert.Equal(t, f.B, cpu.B, "B register")
	assert.Equal(t, f.C, cpu.C, "C register")
	assert.Equal(t, f.D, cpu.D, "D register")
	assert.Equal(t, f.E, cpu.E, "E register")
	assert.Equal(t, f.H, cpu.H, "H register")
	assert.Equal(t, f.L, cpu.L, "L register")
	assert.Equal(t, f.I, cpu.I, "I register")
	assert.Equal(t, f.R, cpu.R, "R register")
	assert.Equal(t, f.PC, cpu.PC, "PC")
	assert.Equal(t, f.SP, cpu.SP, "SP")
	assert.Equal(t, f.IX, cpu.IX, "IX")
	assert.Equal(t, f.IY, cpu.IY, "IY")
	assert.Equal(t, f.AF_, cpu.AF_, "AF'")
	assert.Equal(t, f.BC_, cpu.BC_, "BC'")
	assert.Equal(t, f.DE_, cpu.DE_, "DE'")
	assert.Equal(t, f.HL_, cpu.HL_, "HL'")
	assert.Equal(t, f.WZ, cpu.WZ, "WZ (MEMPTR)")
	assert.Equal(t, f.IM, cpu.IM, "IM")

	expectedIFF1 := f.IFF1 != 0
	expectedIFF2 := f.IFF2 != 0
	assert.Equal(t, expectedIFF1, cpu.IFF1, "IFF1")
	assert.Equal(t, expectedIFF2, cpu.IFF2, "IFF2")

	// Check EI pending
	expectedEI := f.EI != 0
	assert.Equal(t, expectedEI, cpu.EIPending, "EI pending")

	// Check Q
	assert.Equal(t, f.Q, cpu.Q, "Q")

	// Check final RAM
	for _, entry := range f.RAM {
		addr := entry[0]
		expected := byte(entry[1])
		actual, err := ram.Read(cpusim.Address(addr))
		assert.NoError(t, err)
		assert.Equal(t, expected, actual, "RAM[0x%04X]", addr)
	}

	// Check port writes
	for _, p := range tc.Ports {
		if p.Dir == "w" {
			actual, ok := port.data[cpusim.Address(p.Address)]
			assert.True(t, ok, "Port write to 0x%04X expected", p.Address)
			assert.Equal(t, byte(p.Value), actual, "Port[0x%04X]", p.Address)
		}
	}
}

func loadTestFile(t *testing.T, filename string) []TestCase {
	t.Helper()

	file, err := os.Open(filename)
	require.NoError(t, err, "Failed to open test file: "+filename)
	defer file.Close()

	var reader *gzip.Reader
	if strings.HasSuffix(filename, ".gz") {
		reader, err = gzip.NewReader(file)
		require.NoError(t, err, "Failed to create gzip reader")
		defer reader.Close()
	}

	var tests []TestCase
	if reader != nil {
		err = json.NewDecoder(reader).Decode(&tests)
	} else {
		err = json.NewDecoder(file).Decode(&tests)
	}
	require.NoError(t, err, "Failed to decode test data from "+filename)

	return tests
}

func runTestFile(t *testing.T, filename string) {
	t.Helper()

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Skipf("Test data file not found: %s (run 'make testdata-z80' to download)", filename)
		return
	}

	tests := loadTestFile(t, filename)

	for i, tc := range tests {
		// Only show first few failures for readability
		if !t.Run(fmt.Sprintf("%d_%s", i, tc.Name), func(t *testing.T) {
			cpu, _, ram, port := setupCPU(&tc)
			err := cpu.Execute()
			if err != nil {
				t.Fatalf("Execute error: %v", err)
			}
			compareCPU(t, &tc, cpu, ram, port)
		}) {
			if i > 5 {
				t.Fatalf("Too many failures, stopping after %d", i)
			}
		}
	}
}

func testOpcodeRange(t *testing.T, prefix string) {
	for opcode := 0x00; opcode <= 0xFF; opcode++ {
		var filename string
		if prefix == "" {
			filename = fmt.Sprintf("%02x", opcode)
		} else {
			filename = fmt.Sprintf("%s %02x", prefix, opcode)
		}
		path := filepath.Join(testDataDir, filename+".json.gz")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			// Try without .gz
			path = filepath.Join(testDataDir, filename+".json")
			if _, err := os.Stat(path); os.IsNotExist(err) {
				continue
			}
		}
		t.Run(filename, func(t *testing.T) {
			runTestFile(t, path)
		})
	}
}

func TestUnprefixed(t *testing.T) {
	if _, err := os.Stat(testDataDir); os.IsNotExist(err) {
		t.Skip("Test data not downloaded. Run 'make testdata-z80'")
	}
	testOpcodeRange(t, "")
}

func TestCBPrefixed(t *testing.T) {
	if _, err := os.Stat(testDataDir); os.IsNotExist(err) {
		t.Skip("Test data not downloaded. Run 'make testdata-z80'")
	}
	testOpcodeRange(t, "cb")
}

func TestEDPrefixed(t *testing.T) {
	if _, err := os.Stat(testDataDir); os.IsNotExist(err) {
		t.Skip("Test data not downloaded. Run 'make testdata-z80'")
	}
	testOpcodeRange(t, "ed")
}

func TestDDPrefixed(t *testing.T) {
	if _, err := os.Stat(testDataDir); os.IsNotExist(err) {
		t.Skip("Test data not downloaded. Run 'make testdata-z80'")
	}
	testOpcodeRange(t, "dd")
}

func TestFDPrefixed(t *testing.T) {
	if _, err := os.Stat(testDataDir); os.IsNotExist(err) {
		t.Skip("Test data not downloaded. Run 'make testdata-z80'")
	}
	testOpcodeRange(t, "fd")
}

func TestDDCBPrefixed(t *testing.T) {
	if _, err := os.Stat(testDataDir); os.IsNotExist(err) {
		t.Skip("Test data not downloaded. Run 'make testdata-z80'")
	}
	testOpcodeRange(t, "dd cb")
}

func TestFDCBPrefixed(t *testing.T) {
	if _, err := os.Stat(testDataDir); os.IsNotExist(err) {
		t.Skip("Test data not downloaded. Run 'make testdata-z80'")
	}
	testOpcodeRange(t, "fd cb")
}
