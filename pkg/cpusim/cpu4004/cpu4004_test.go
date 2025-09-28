package cpu4004

import (
	"bytes"
	"github.com/scottmbaker/gocpusim/pkg/cpusim"
	"github.com/stretchr/testify/suite"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

/*
 * Lacking coverage:
 *   - RDR (read rom port)
 *   - WRR (write rom port)
 *   - WMP (write to memory port)
 *   - ADM (add from memory)
 *   - SBM (subtract from memory)
 *   - KBP
 */

const (
	ASL   = "/usr/local/bin/asl"
	P2BIN = "/usr/local/bin/p2bin"
)

type TestPort struct {
	Sim *cpusim.CpuSim
	in  [8]byte
	out [32]byte // the first 8 are not used
}

func (p *TestPort) HasAddress(address cpusim.Address) bool {
	return true
}

func (p *TestPort) Read(address cpusim.Address) (byte, error) {
	if address < 8 {
		return p.in[address], nil
	}
	return 0, &cpusim.ErrInvalidAddress{Address: address}
}

func (p *TestPort) Write(address cpusim.Address, value byte) error {
	if (address < 8) || (address >= 32) {
		return &cpusim.ErrInvalidAddress{Address: address}
	}
	p.out[address] = value
	return nil
}

func (p *TestPort) ReadStatus(address cpusim.Address, statusAddr cpusim.Address) (byte, error) {
	_ = address
	_ = statusAddr
	return 0, nil
}

func (p *TestPort) WriteStatus(address cpusim.Address, statusAddr cpusim.Address, value byte) error {
	_ = address
	_ = statusAddr
	_ = value
	return nil
}

func (p *TestPort) GetKind() string {
	return "TESTPORT"
}

func (p *TestPort) GetName() string {
	return "testport"
}

type Cpu4004Suite struct {
	suite.Suite
	sim        *cpusim.CpuSim
	cpu        *CPU4004
	ram        *cpusim.Memory
	rom        *cpusim.Memory
	testPort   *TestPort
	testBinDir string
}

func getTestName() string {
	counter, _, _, success := runtime.Caller(3)
	if !success {
		return "program"
	}
	name := runtime.FuncForPC(counter).Name()
	parts := strings.Split(name, ".")
	return parts[len(parts)-1]
}

func (s *Cpu4004Suite) SetupTest() {
	s.sim = cpusim.NewCPUSim()

	// Create an 8008 CPU and attach it to the simulator
	s.cpu = New4004(s.sim, "cpu")
	s.sim.AddCPU(s.cpu)

	s.rom = cpusim.NewMemory(s.sim, "rom", cpusim.KIND_ROM, 0x0000, 0x3FFF, 12, true, &cpusim.TrueEnabler{})
	s.sim.AddMemory(s.rom)

	s.ram = cpusim.NewMemory(s.sim, "ram", cpusim.KIND_RAM, 0x0000, 0x3F, 6, false, s.cpu.DCLEnabler(0))
	s.ram.CreateStatusBytes(0x04, 0x04)
	s.sim.AddMemory(s.ram)

	b8b := NewBus8Bit(s.sim, "bus8", s.cpu.DCLEnabler(4))
	s.sim.AddMemory(b8b)

	s.testBinDir = "testbin"

	err := os.MkdirAll(s.testBinDir, os.ModePerm)
	s.Require().NoError(err, "Failed to create directory: ../../testbin")
}

func (s *Cpu4004Suite) run(command string, args ...string) (string, string, error) {
	cmd := exec.Command(command, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func (s *Cpu4004Suite) assemble(program string) string {
	stem := filepath.Join(s.testBinDir, getTestName())
	asmFile := stem + ".asm"
	pFile := stem + ".p"
	binFile := stem + ".bin"

	// check to see if the file was already assembled.
	if existingContent, err := os.ReadFile(asmFile); err == nil {
		if string(existingContent) == program {
			if _, err := os.Stat(binFile); err == nil {
				return binFile
			}
		} else {
			// Testify makes it very difficult to log a message outside of verbose mode.
			// It's useful to see when things reassemble, so I made this an error rather than a s.T().Logf().
			s.Equal(0, 1, "File content mismatch for "+asmFile+". Re-assembling. Re-run tests and this error will go away.")
		}
	}

	// remove existing p and bin file
	_ = os.Remove(pFile)
	_ = os.Remove(binFile)

	file, err := os.OpenFile(asmFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	s.Require().NoError(err, "Failed to open file for writing: "+asmFile)
	defer file.Close() // nolint:errcheck
	_, err = file.WriteString(program)
	s.Require().NoError(err, "Failed to write to file: "+asmFile)

	_, stderr, err := s.run(ASL, "-cpu", "4004", "-L", asmFile, "-o", pFile)
	if err != nil {
		_ = os.Remove(pFile)
	}
	s.Require().NoError(err, "Assembly failed: "+stderr)

	_, stderr, err = s.run(P2BIN, pFile, binFile)
	if err != nil {
		_ = os.Remove(binFile)
	}
	s.Require().NoError(err, "Conversion to binary failed: "+stderr)

	return binFile
}

func (s *Cpu4004Suite) AssembleAndLoad(program string) {
	var indentedProgram string
	header := `
cpu 4040                ; use 4040 for halt instruction
radix 10                ; use base 10 for numbers

include "../testdata/reg4004.inc"   ; Include 4004 register definitions.

org 0
    `
	for _, line := range strings.Split(header+program, "\n") {
		if line != "" {
			if strings.Contains(line, ":") {
				indentedProgram += line + "\n"
			} else {
				indentedProgram += "\t" + line + "\n"
			}
		}
	}

	binFile := s.assemble(indentedProgram)
	err := s.rom.Load(binFile)
	s.Require().NoError(err, "Failed to load binary file: "+binFile)
}

func (s *Cpu4004Suite) TestIncrement() {
	s.cpu.Registers[REG_R3] = 0     // Set register R3 to 0
	s.cpu.Registers[FLAG_CARRY] = 0 // Make sure carry is unset
	s.AssembleAndLoad(`
INC R3
HLT
`)
	err := s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(1), s.cpu.Registers[REG_R3])
	s.Equal(byte(0), s.cpu.Registers[FLAG_CARRY])

	s.cpu.PC = 0 // Reset program counter to start
	err = s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(2), s.cpu.Registers[REG_R3])
	s.Equal(byte(0), s.cpu.Registers[FLAG_CARRY])

	s.cpu.PC = 0 // Reset program counter to start
	err = s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(3), s.cpu.Registers[REG_R3])
	s.Equal(byte(0), s.cpu.Registers[FLAG_CARRY])

	s.cpu.Registers[REG_R3] = 255 // Set register R3 to 255
	s.cpu.PC = 0                  // Reset program counter to start
	err = s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(0), s.cpu.Registers[REG_R3])
	s.Equal(byte(0), s.cpu.Registers[FLAG_CARRY]) // carry is not set during increment

	s.cpu.PC = 0 // Reset program counter to start
	s.cpu.Registers[REG_R3] = 0
	s.cpu.Registers[FLAG_CARRY] = 1
	err = s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(1), s.cpu.Registers[REG_R3])
	s.Equal(byte(1), s.cpu.Registers[FLAG_CARRY]) // carry should be unaffected
}

func (s *Cpu4004Suite) TestIAC() {
	s.cpu.Registers[REG_ACCUM] = 0 // Set register R3 to 0
	s.AssembleAndLoad(`
IAC
HLT
`)
	err := s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(1), s.cpu.Registers[REG_ACCUM])
	s.Equal(byte(0), s.cpu.Registers[FLAG_CARRY])

	s.cpu.PC = 0 // Reset program counter to start
	err = s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(2), s.cpu.Registers[REG_ACCUM])
	s.Equal(byte(0), s.cpu.Registers[FLAG_CARRY])

	s.cpu.Registers[FLAG_CARRY] = 1
	s.cpu.PC = 0 // Reset program counter to start
	err = s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(3), s.cpu.Registers[REG_ACCUM])
	s.Equal(byte(1), s.cpu.Registers[FLAG_CARRY])

	s.cpu.PC = 0 // Reset program counter to start
	s.cpu.Registers[REG_ACCUM] = 0x0F
	s.cpu.Registers[FLAG_CARRY] = 0
	err = s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(0), s.cpu.Registers[REG_ACCUM])
	s.Equal(byte(0), s.cpu.Registers[FLAG_CARRY])
}

func (s *Cpu4004Suite) TestDAC() {
	s.cpu.Registers[REG_ACCUM] = 0x0F // Set register R3 to 0
	s.AssembleAndLoad(`
DAC
HLT
`)
	err := s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(0x0E), s.cpu.Registers[REG_ACCUM])
	s.Equal(byte(0), s.cpu.Registers[FLAG_CARRY])

	s.cpu.PC = 0 // Reset program counter to start
	err = s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(0x0D), s.cpu.Registers[REG_ACCUM])
	s.Equal(byte(0), s.cpu.Registers[FLAG_CARRY])

	s.cpu.Registers[FLAG_CARRY] = 1
	s.cpu.PC = 0 // Reset program counter to start
	err = s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(0x0C), s.cpu.Registers[REG_ACCUM])
	s.Equal(byte(1), s.cpu.Registers[FLAG_CARRY])

	s.cpu.PC = 0 // Reset program counter to start
	s.cpu.Registers[REG_ACCUM] = 0x00
	s.cpu.Registers[FLAG_CARRY] = 0
	err = s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(0x0F), s.cpu.Registers[REG_ACCUM])
	s.Equal(byte(0), s.cpu.Registers[FLAG_CARRY])
}

func (s *Cpu4004Suite) TestRAL() {
	s.cpu.Registers[REG_ACCUM] = 0x01 // Set register ACCUM to 1
	s.AssembleAndLoad(`
RAL
HLT
`)
	err := s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(0x02), s.cpu.Registers[REG_ACCUM])
	s.Equal(byte(0), s.cpu.Registers[FLAG_CARRY])

	s.cpu.PC = 0 // Reset program counter to start
	s.cpu.Registers[FLAG_CARRY] = 0x01
	err = s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(0x05), s.cpu.Registers[REG_ACCUM])
	s.Equal(byte(0), s.cpu.Registers[FLAG_CARRY])

	s.cpu.PC = 0 // Reset program counter to start
	s.cpu.Registers[FLAG_CARRY] = 0x00
	err = s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(0x0A), s.cpu.Registers[REG_ACCUM])
	s.Equal(byte(0), s.cpu.Registers[FLAG_CARRY])

	s.cpu.PC = 0 // Reset program counter to start
	s.cpu.Registers[FLAG_CARRY] = 0x00
	err = s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(0x04), s.cpu.Registers[REG_ACCUM])
	s.Equal(byte(1), s.cpu.Registers[FLAG_CARRY])
}

func (s *Cpu4004Suite) TestRAR() {
	s.cpu.Registers[REG_ACCUM] = 0x08 // Set register ACCUM to 0x08
	s.AssembleAndLoad(`
RAR
HLT
`)
	err := s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(0x04), s.cpu.Registers[REG_ACCUM])
	s.Equal(byte(0), s.cpu.Registers[FLAG_CARRY])

	s.cpu.PC = 0 // Reset program counter to start
	s.cpu.Registers[FLAG_CARRY] = 0x01
	err = s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(0x0A), s.cpu.Registers[REG_ACCUM])
	s.Equal(byte(0), s.cpu.Registers[FLAG_CARRY])

	s.cpu.PC = 0 // Reset program counter to start
	s.cpu.Registers[FLAG_CARRY] = 0x00
	err = s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(0x05), s.cpu.Registers[REG_ACCUM])
	s.Equal(byte(0), s.cpu.Registers[FLAG_CARRY])

	s.cpu.PC = 0 // Reset program counter to start
	s.cpu.Registers[FLAG_CARRY] = 0x00
	err = s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(0x02), s.cpu.Registers[REG_ACCUM])
	s.Equal(byte(1), s.cpu.Registers[FLAG_CARRY])
}

func (s *Cpu4004Suite) TestTCC() {
	s.cpu.Registers[REG_ACCUM] = 0x06 // Set register ACCUM to 0x06
	s.AssembleAndLoad(`
TCC
HLT
`)
	err := s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(0), s.cpu.Registers[REG_ACCUM])
	s.Equal(byte(0), s.cpu.Registers[FLAG_CARRY])

	s.cpu.PC = 0 // Reset program counter to start
	s.cpu.Registers[FLAG_CARRY] = 0x01
	err = s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(1), s.cpu.Registers[REG_ACCUM])
	s.Equal(byte(0), s.cpu.Registers[FLAG_CARRY])
}

func (s *Cpu4004Suite) TestTCS() {
	s.cpu.Registers[REG_ACCUM] = 0x06 // Set register ACCUM to 0x06
	s.AssembleAndLoad(`
TCS
HLT
`)
	err := s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(9), s.cpu.Registers[REG_ACCUM])
	s.Equal(byte(0), s.cpu.Registers[FLAG_CARRY])

	s.cpu.PC = 0 // Reset program counter to start
	s.cpu.Registers[FLAG_CARRY] = 0x01
	err = s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(10), s.cpu.Registers[REG_ACCUM])
	s.Equal(byte(0), s.cpu.Registers[FLAG_CARRY])
}

func (s *Cpu4004Suite) TestExchange() {
	s.AssembleAndLoad(`
LDM 0
XCH R0
LDM 1
XCH R1
LDM 2
XCH R2
LDM 3
XCH R3
LDM 4
XCH R4
LDM 5
XCH R5
LDM 6
XCH R6
LDM 7
XCH R7
LDM 8
XCH R8
LDM 9
XCH R9
LDM 10
XCH R10
LDM 11
XCH R11
LDM 12
XCH R12
LDM 13
XCH R13
LDM 14
XCH R14
LDM 15
XCH R15
HLT
`)
	err := s.cpu.Run()
	s.NoError(err)

	for i := 0; i <= 15; i++ {
		s.Equal(byte(i), s.cpu.Registers[REG_R0+i], "Register R%02d should be %02X", i, byte(i))
	}
}

func (s *Cpu4004Suite) TestLoad() {
	s.AssembleAndLoad(`
LDM 13
XCH R7
LDM 0
LD R7
HLT
`)
	err := s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(13), s.cpu.Registers[REG_ACCUM], "Accumulator should be %02X", byte(13))
	s.Equal(byte(13), s.cpu.Registers[REG_R7], "Register R07 should be %02X", byte(13))
}

func (s *Cpu4004Suite) TestRegs2() {
	s.AssembleAndLoad(`
LDM 0
XCH R0	; R0 = 0
LDM 9
XCH R9	; R9 = 9
LDM 12
XCH R12	; R12 = 12
LDM 7
XCH R9   ; accum = R9 (9), R9 = 7
XCH R0	 ; accum = R0 (0), R0 = 9
XCH R12  ; accum = R12 (12), R12 = 9
HLT
`)
	s.cpu.PC = 0 // Reset program counter to start
	err := s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(9), s.cpu.Registers[REG_R0], "Register R00 should be %02X", byte(9))
	s.Equal(byte(7), s.cpu.Registers[REG_R9], "Register R09 should be %02X", byte(7))
	s.Equal(byte(0), s.cpu.Registers[REG_R12], "Register R12 should be %02X", byte(0))
	s.Equal(byte(12), s.cpu.Registers[REG_ACCUM], "Register A should be %02X", byte(12))
}

func (s *Cpu4004Suite) TestPairs() {
	s.AssembleAndLoad(`
FIM P0, 01H
FIM P1, 23H
FIM P2, 45H
FIM P3, 67H
FIM P4, 89H
FIM P5, 0ABH
FIM P6, 0CDH
FIM P7, 0EFH
HLT
`)
	err := s.cpu.Run()
	s.NoError(err)

	for i := 0; i <= 15; i++ {
		s.Equal(byte(i), s.cpu.Registers[REG_R0+i], "Register R%02d should be %02X", i, byte(i))
	}
}

func (s *Cpu4004Suite) TestAdd() {
	s.AssembleAndLoad(`
ADD R3
HLT
`)
	s.cpu.Registers[REG_R3] = 7
	s.cpu.Registers[REG_ACCUM] = 5
	err := s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(12), s.cpu.Registers[REG_ACCUM], "Accumulator should be 0x0C")
	s.Equal(byte(0), s.cpu.Registers[FLAG_CARRY], "Carry should be unset")

	s.cpu.PC = 0                // Reset program counter to start
	s.cpu.Registers[REG_R3] = 6 // should now be 6 + 12 = 18, overflow and leaves 2 in accumulator
	err = s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(2), s.cpu.Registers[REG_ACCUM], "Accumulator should be 0x02")
	s.Equal(byte(1), s.cpu.Registers[FLAG_CARRY], "Carry should be set")

	s.cpu.PC = 0 // Reset program counter to start
	s.cpu.Registers[REG_R3] = 7
	s.cpu.Registers[REG_ACCUM] = 5
	s.cpu.Registers[FLAG_CARRY] = 1 // carry should add 1 to the sum
	err = s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(13), s.cpu.Registers[REG_ACCUM], "Accumulator should be 0x0D")
	s.Equal(byte(0), s.cpu.Registers[FLAG_CARRY], "Carry should be unset")
}

func (s *Cpu4004Suite) TestSub() {
	s.AssembleAndLoad(`
SUB R3
HLT
`)
	s.cpu.Registers[REG_R3] = 5
	s.cpu.Registers[REG_ACCUM] = 7
	err := s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(2), s.cpu.Registers[REG_ACCUM], "Accumulator should be 0x02")
	s.Equal(byte(1), s.cpu.Registers[FLAG_CARRY], "Carry should be set")

	s.cpu.PC = 0 // Reset program counter to start
	s.cpu.Registers[FLAG_CARRY] = 0
	s.cpu.Registers[REG_R3] = 6 // should now be 2-6 = -4, overflow and leaves 4 in accumulator
	err = s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(12), s.cpu.Registers[REG_ACCUM], "Accumulator should be 0x0C")
	s.Equal(byte(0), s.cpu.Registers[FLAG_CARRY], "Carry should be unset")

	s.cpu.PC = 0 // Reset program counter to start
	s.cpu.Registers[REG_R3] = 5
	s.cpu.Registers[REG_ACCUM] = 7
	s.cpu.Registers[FLAG_CARRY] = 1 // carry should subtract 1 from the result
	err = s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(1), s.cpu.Registers[REG_ACCUM], "Accumulator should be 0x01")
	s.Equal(byte(1), s.cpu.Registers[FLAG_CARRY], "Carry should be set")
}

func (s *Cpu4004Suite) TestSubBug() {
	s.AssembleAndLoad(`
LDM 15
XCH R9
LDM 15
CLC
SUB R9
HLT
`)
	err := s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(0), s.cpu.Registers[REG_ACCUM], "Accumulator should be 0x02")
	s.Equal(byte(1), s.cpu.Registers[FLAG_CARRY], "Carry should be set")
}

func (s *Cpu4004Suite) TestSBM() {
	s.AssembleAndLoad(`
FIM P7, 00H
SRC P7
LD R1
WRM
LD R2
CLC
SBM
HLT
`)
	s.cpu.Registers[REG_R1] = 7
	s.cpu.Registers[REG_R2] = 6
	err := s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(15), s.cpu.Registers[REG_ACCUM], "Accumulator should be 0x0F")
	s.Equal(byte(0), s.cpu.Registers[FLAG_CARRY], "Carry should be unset")

	s.cpu.PC = 0 // Reset program counter to start
	s.cpu.Registers[REG_R1] = 6
	s.cpu.Registers[REG_R2] = 7
	err = s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(1), s.cpu.Registers[REG_ACCUM], "Accumulator should be 0x01")
	s.Equal(byte(1), s.cpu.Registers[FLAG_CARRY], "Carry should be set")
}

func (s *Cpu4004Suite) TestSTC() {
	s.AssembleAndLoad(`
STC
HLT
`)
	s.cpu.Registers[FLAG_CARRY] = 0
	err := s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(1), s.cpu.Registers[FLAG_CARRY], "Carry should be set")
}

func (s *Cpu4004Suite) TestCLC() {
	s.AssembleAndLoad(`
CLC
HLT
`)
	s.cpu.Registers[FLAG_CARRY] = 1
	err := s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(0), s.cpu.Registers[FLAG_CARRY], "Carry should be unset")
}

func (s *Cpu4004Suite) TestCLB() {
	s.AssembleAndLoad(`
CLB
HLT
`)
	s.cpu.Registers[FLAG_CARRY] = 1
	s.cpu.Registers[REG_ACCUM] = 0x03
	err := s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(0), s.cpu.Registers[FLAG_CARRY], "Carry should be unset")
	s.Equal(byte(0), s.cpu.Registers[REG_ACCUM], "Accumulator should be 0")
}

func (s *Cpu4004Suite) TestCMC() {
	s.AssembleAndLoad(`
CMC
HLT
`)
	s.cpu.Registers[FLAG_CARRY] = 1
	err := s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(0), s.cpu.Registers[FLAG_CARRY], "Carry should be unset")

	s.cpu.PC = 0 // Reset program counter to start
	s.cpu.Registers[FLAG_CARRY] = 0
	err = s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(1), s.cpu.Registers[FLAG_CARRY], "Carry should be unset")
}

func (s *Cpu4004Suite) TestCMA() {
	s.AssembleAndLoad(`
CMA
HLT
`)
	s.cpu.Registers[REG_ACCUM] = 0x03
	err := s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(0x0C), s.cpu.Registers[REG_ACCUM], "Accumulator should be 0x0C")

	s.cpu.PC = 0 // Reset program counter to start
	s.cpu.Registers[FLAG_CARRY] = 0
	s.cpu.Registers[REG_ACCUM] = 0x0C
	err = s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(0x03), s.cpu.Registers[REG_ACCUM], "Accumulator should be 0x03")
	s.Equal(byte(0), s.cpu.Registers[FLAG_CARRY], "Carry should be unset")

	s.cpu.PC = 0 // Reset program counter to start
	s.cpu.Registers[FLAG_CARRY] = 1
	s.cpu.Registers[REG_ACCUM] = 0x0C
	err = s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(0x03), s.cpu.Registers[REG_ACCUM], "Accumulator should be 0x03")
	s.Equal(byte(1), s.cpu.Registers[FLAG_CARRY], "Carry should be set")
}

func (s *Cpu4004Suite) TestRam() {
	s.AssembleAndLoad(`
FIM P7, 00H
SRC P7
LDM 1
WR0
LDM 2
WR1
LDM 3
WR2
LDM 4
WR3
FIM P7, 10H
SRC P7
LDM 5
WR0
LDM 6
WR1
LDM 7
WR2
LDM 8
WR3

FIM P7,00H
SRC P7
LDM 9
WRM
FIM P7,01H
SRC P7
LDM 10
WRM
FIM P7,02H
SRC P7
LDM 11
WRM
FIM P7,10H
SRC P7
LDM 12
WRM
FIM P7,11H
SRC P7
LDM 13
WRM
FIM P7,12H
SRC P7
LDM 14
WRM

FIM P7, 00H
SRC P7
RD0
XCH R0
RD1
XCH R1
RD2
XCH R2
RD3
XCH R3
FIM P7, 10H
SRC P7
RD0
XCH R4
RD1
XCH R5
RD2
XCH R6
RD3
XCH R7

FIM P7,00H
SRC P7
RDM
XCH R8
FIM P7,01H
SRC P7
RDM
XCH R9
FIM P7,02H
SRC P7
RDM
XCH R10
FIM P7,10H
SRC P7
RDM
XCH R11
FIM P7,11H
SRC P7
RDM
XCH R12
FIM P7,12H
SRC P7
RDM
XCH R13

HLT
`)
	err := s.cpu.Run()
	s.NoError(err)

	for i := 0; i < 14; i++ {
		s.Equal(byte(i+1), s.cpu.Registers[REG_R0+i], "Register R%02d should be %02X", i, byte(i+1))
	}
}

func (s *Cpu4004Suite) TestDAA() {
	s.AssembleAndLoad(`
DAA
HLT
`)
	s.cpu.Registers[REG_ACCUM] = 5
	s.cpu.Registers[FLAG_CARRY] = 0
	err := s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(5), s.cpu.Registers[REG_ACCUM], "Accumulator should be 0x05")
	s.Equal(byte(0), s.cpu.Registers[FLAG_CARRY], "Carry should be unset")

	s.cpu.PC = 0 // Reset program counter to start
	s.cpu.Registers[REG_ACCUM] = 5
	s.cpu.Registers[FLAG_CARRY] = 1
	err = s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(0x0B), s.cpu.Registers[REG_ACCUM], "Accumulator should be 0x0B")
	s.Equal(byte(1), s.cpu.Registers[FLAG_CARRY], "Carry should be set")

	s.cpu.PC = 0 // Reset program counter to start
	s.cpu.Registers[REG_ACCUM] = 9
	s.cpu.Registers[FLAG_CARRY] = 0
	err = s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(0x09), s.cpu.Registers[REG_ACCUM], "Accumulator should be 0x09")
	s.Equal(byte(0), s.cpu.Registers[FLAG_CARRY], "Carry should be unset")

	s.cpu.PC = 0 // Reset program counter to start
	s.cpu.Registers[REG_ACCUM] = 9
	s.cpu.Registers[FLAG_CARRY] = 1
	err = s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(0x0F), s.cpu.Registers[REG_ACCUM], "Accumulator should be 0x0B")
	s.Equal(byte(1), s.cpu.Registers[FLAG_CARRY], "Carry should be set")

	s.cpu.PC = 0 // Reset program counter to start
	s.cpu.Registers[REG_ACCUM] = 0x0A
	s.cpu.Registers[FLAG_CARRY] = 0
	err = s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(0x00), s.cpu.Registers[REG_ACCUM], "Accumulator should be 0x00")
	s.Equal(byte(1), s.cpu.Registers[FLAG_CARRY], "Carry should be set")

	s.cpu.PC = 0 // Reset program counter to start
	s.cpu.Registers[REG_ACCUM] = 0x0F
	s.cpu.Registers[FLAG_CARRY] = 0
	err = s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(5), s.cpu.Registers[REG_ACCUM], "Accumulator should be 0x05")
	s.Equal(byte(1), s.cpu.Registers[FLAG_CARRY], "Carry should be set")
}

func (s *Cpu4004Suite) TestJMSBBL() {
	s.AssembleAndLoad(`
LDM 01H
XCH R1
JMS SUB1
XCH R7   ; contains 7 from the BBL from SUB1
HLT

SUB1:
LDM 02H
XCH R2
JMS SUB2
XCH R6	; contains 6 from the BBL from SUB2
BBL 7

SUB2:
LDM 03
XCH R3
JMS SUB3
XCH R5	; contains 5 from the BBL from SUB3
BBL 6

SUB3:
LDM 04
XCH R4
BBL 5
`)
	err := s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(1), s.cpu.Registers[REG_R1], "R1 should be 1")
	s.Equal(byte(2), s.cpu.Registers[REG_R2], "R2 should be 2")
	s.Equal(byte(3), s.cpu.Registers[REG_R3], "R3 should be 3")
	s.Equal(byte(4), s.cpu.Registers[REG_R4], "R4 should be 4")
	s.Equal(byte(5), s.cpu.Registers[REG_R5], "R5 should be 5")
	s.Equal(byte(6), s.cpu.Registers[REG_R6], "R6 should be 6")
	s.Equal(byte(7), s.cpu.Registers[REG_R7], "R7 should be 7")
}

func (s *Cpu4004Suite) TestJUN() {
	s.AssembleAndLoad(`
LDM 01H
XCH R1
JUN LAB1
LDM 03H
XCH R3	; should never be executed

LAB1:
LDM 02H
XCH R2
HLT
`)
	err := s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(1), s.cpu.Registers[REG_R1], "R1 should be 1")
	s.Equal(byte(2), s.cpu.Registers[REG_R2], "R2 should be 2")
	s.Equal(byte(0), s.cpu.Registers[REG_R3], "R3 should be 0")
}

func (s *Cpu4004Suite) TestJIN() {
	s.AssembleAndLoad(`
JIN P1     ; 0
HLT	   	   ; 1
L1:
LDM 1H     ; 2
HLT		   ; 3
L2:
LDM 2H	   ; 4
HLT		   ; 5
L3:
LDM 3H     ; 6
HLT		   ; 7
`)
	s.cpu.Registers[REG_R2] = 0
	s.cpu.Registers[REG_R3] = 2
	err := s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(1), s.cpu.Registers[REG_ACCUM], "ACCUM should be 1")

	s.cpu.PC = 0 // Reset program counter to start
	s.cpu.Registers[REG_R2] = 0
	s.cpu.Registers[REG_R3] = 4
	err = s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(2), s.cpu.Registers[REG_ACCUM], "ACCUM should be 2")

	s.cpu.Registers[REG_R2] = 0
	s.cpu.Registers[REG_R3] = 6
	err = s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(3), s.cpu.Registers[REG_ACCUM], "ACCUM should be 3")
}

func (s *Cpu4004Suite) TestJIN_boundary() {
	s.AssembleAndLoad(`
JUN START
org 0FFh
START:
JIN P1     ; FF
HLT	   	   ; 100
L1:
LDM 1H     ; 101
HLT		   ; 102
L2:
LDM 2H	   ; 103
HLT		   ; 104
L3:
LDM 3H     ; 105
HLT		   ; 106
`)
	s.cpu.Registers[REG_R2] = 0
	s.cpu.Registers[REG_R3] = 1
	err := s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(1), s.cpu.Registers[REG_ACCUM], "ACCUM should be 1")

	s.cpu.PC = 0 // Reset program counter to start
	s.cpu.Registers[REG_R2] = 0
	s.cpu.Registers[REG_R3] = 3
	err = s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(2), s.cpu.Registers[REG_ACCUM], "ACCUM should be 2")

	s.cpu.Registers[REG_R2] = 0
	s.cpu.Registers[REG_R3] = 5
	err = s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(3), s.cpu.Registers[REG_ACCUM], "ACCUM should be 3")
}

func (s *Cpu4004Suite) TestNOP() {
	s.AssembleAndLoad(`
NOP
NOP
NOP
HLT
`)
	err := s.cpu.Run()
	s.NoError(err)
}

func (s *Cpu4004Suite) TestFIN() {
	s.AssembleAndLoad(`
FIM P0, DAT1
FIN P1
FIM P0, DAT0
FIN P0
HLT
DAT1: data 12h
DAT0: data 34h
`)
	err := s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(1), s.cpu.Registers[REG_R2], "R2 should be 1")
	s.Equal(byte(2), s.cpu.Registers[REG_R3], "R3 should be 2")
	s.Equal(byte(3), s.cpu.Registers[REG_R0], "R0 should be 3")
	s.Equal(byte(4), s.cpu.Registers[REG_R1], "R1 should be 4")
}

func (s *Cpu4004Suite) TestFINBoundary() {
	s.AssembleAndLoad(`
JUN START
org 04h
WRONG1: data 56h
WRONG0: data 78h
org 0FDh
START:
FIM P0, 04H
FIN P1
org 0100h
FIM P0, 05H
FIN P0
HLT
DAT1: data 12h
DAT0: data 34h
`)
	err := s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(1), s.cpu.Registers[REG_R2], "R2 should be 1")
	s.Equal(byte(2), s.cpu.Registers[REG_R3], "R3 should be 2")
	s.Equal(byte(3), s.cpu.Registers[REG_R0], "R0 should be 3")
	s.Equal(byte(4), s.cpu.Registers[REG_R1], "R1 should be 4")
}

func (s *Cpu4004Suite) TestJCN_C() {
	s.AssembleAndLoad(`
LDM 1
XCH R0
JCN C, L1
HLT

L1:
LDM 2
XCH R0
HLT
`)
	err := s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(1), s.cpu.Registers[REG_R0], "R0 should be 1")

	s.cpu.PC = 0 // Reset program counter to start
	s.cpu.Registers[FLAG_CARRY] = 1
	err = s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(2), s.cpu.Registers[REG_R0], "R0 should be 2")
}

func (s *Cpu4004Suite) TestJCN_CN() {
	s.AssembleAndLoad(`
LDM 1
XCH R0
JCN CN, L1
HLT

L1:
LDM 2
XCH R0
HLT
`)
	err := s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(2), s.cpu.Registers[REG_R0], "R0 should be 2")

	s.cpu.PC = 0 // Reset program counter to start
	s.cpu.Registers[FLAG_CARRY] = 1
	err = s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(1), s.cpu.Registers[REG_R0], "R0 should be 1")
}

func (s *Cpu4004Suite) TestJCN_Z() {
	s.AssembleAndLoad(`
JCN Z, L1
HLT

L1:
LDM 2
XCH R0
HLT
`)
	s.cpu.Registers[REG_R0] = 1
	s.cpu.Registers[REG_ACCUM] = 2
	err := s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(1), s.cpu.Registers[REG_R0], "R0 should be 1")

	s.cpu.PC = 0 // Reset program counter to start
	s.cpu.Registers[REG_R0] = 1
	s.cpu.Registers[REG_ACCUM] = 0
	err = s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(2), s.cpu.Registers[REG_R0], "R0 should be 2")
}

func (s *Cpu4004Suite) TestJCN_ZN() {
	s.AssembleAndLoad(`
JCN ZN, L1
HLT

L1:
LDM 2
XCH R0
HLT
`)
	s.cpu.Registers[REG_R0] = 1
	s.cpu.Registers[REG_ACCUM] = 0
	err := s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(1), s.cpu.Registers[REG_R0], "R0 should be 1")

	s.cpu.PC = 0 // Reset program counter to start
	s.cpu.Registers[REG_R0] = 1
	s.cpu.Registers[REG_ACCUM] = 2
	err = s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(2), s.cpu.Registers[REG_R0], "R0 should be 2")
}

func (s *Cpu4004Suite) TestJCN_C_Boundary() {
	s.AssembleAndLoad(`
org 00h
JUN START

org 0FCh
START:
LDM 1
XCH R0
data 012H, 01H ;  assembler refuses to assemble JCN C, 01h  ; FE, FF
HLT		    ; 00

L1:
LDM 2		; 01
XCH R0
HLT
`)
	err := s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(1), s.cpu.Registers[REG_R0], "R0 should be 1")

	s.cpu.PC = 0 // Reset program counter to start
	s.cpu.Registers[FLAG_CARRY] = 1
	err = s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(2), s.cpu.Registers[REG_R0], "R0 should be 2")
}

func (s *Cpu4004Suite) TestISZ() {
	s.AssembleAndLoad(`
L1:
INC R3
ISZ R4, L1
HLT
`)
	s.cpu.Registers[REG_R3] = 1
	s.cpu.Registers[REG_R4] = 16 - 7
	err := s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(8), s.cpu.Registers[REG_R3], "R3 should be 8")

	s.cpu.PC = 0 // Reset program counter to start
	s.cpu.Registers[REG_R3] = 1
	s.cpu.Registers[REG_R4] = 16 - 16 // loop 16 times
	err = s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(1), s.cpu.Registers[REG_R3], "R3 should be 1")
}

func (s *Cpu4004Suite) TestISZ_boundary() {
	s.AssembleAndLoad(`
JUN START
org 0FDh
START:
L1:
INC R3		; FD
ISZ R4, L1  ; FE, FF
HLT		    ; 100
org 1FDh	; since ISZ is at 0FEh, cpu will wrap to 01FD rather than 0FD
INC R2
JUN L1
`)
	s.cpu.Registers[REG_R2] = 7
	s.cpu.Registers[REG_R3] = 1
	s.cpu.Registers[REG_R4] = 16 - 7
	err := s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(13), s.cpu.Registers[REG_R2], "R2 should be 13")
	s.Equal(byte(8), s.cpu.Registers[REG_R3], "R3 should be 8")

	s.cpu.PC = 0 // Reset program counter to start
	s.cpu.Registers[REG_R2] = 7
	s.cpu.Registers[REG_R3] = 1
	s.cpu.Registers[REG_R4] = 16 - 16 // loop 16 times
	err = s.cpu.Run()
	s.NoError(err)

	s.Equal(byte(6), s.cpu.Registers[REG_R2], "R2 should be 5")
	s.Equal(byte(1), s.cpu.Registers[REG_R3], "R3 should be 1")
}

func TestCpu4004Suite(t *testing.T) {
	suite.Run(t, new(Cpu4004Suite))
}
