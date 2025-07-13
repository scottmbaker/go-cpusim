package cpu8008

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

const (
	ASL   = "/usr/local/bin/asl"
	P2BIN = "/usr/local/bin/p2bin"
)

type TestPort struct {
	Sim *cpusim.CpuSim
	in  [8]byte
	out [32]byte // the first 8 are not used
}

func (p *TestPort) HasAddress(address uint16) bool {
	return true
}

func (p *TestPort) Read(address uint16) (byte, error) {
	if address < 8 {
		return p.in[address], nil
	}
	return 0, &cpusim.ErrInvalidAddress{Address: address}
}

func (p *TestPort) Write(address uint16, value byte) error {
	if (address < 8) || (address >= 32) {
		return &cpusim.ErrInvalidAddress{Address: address}
	}
	p.out[address] = value
	return nil
}

func (p *TestPort) GetName() string {
	return "testport"
}

type Cpu8008Suite struct {
	suite.Suite
	sim        *cpusim.CpuSim
	cpu        *CPU8008
	ram        *cpusim.Memory
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

func (s *Cpu8008Suite) SetupTest() {
	s.sim = cpusim.NewCPUSim()

	// Create an 8008 CPU and attach it to the simulator
	s.cpu = New8008(s.sim, "cpu")
	s.sim.AddCPU(s.cpu)

	s.ram = cpusim.NewMemory(s.sim, "ram", 0x0000, 0x3FFF, 14, false, &cpusim.AlwaysEnabled)
	s.sim.AddMemory(s.ram)

	s.testPort = &TestPort{Sim: s.sim,
		in:  [8]byte{0xC0, 0xC1, 0xC2, 0xC3, 0xC4, 0xC5, 0xC6, 0xC7},
		out: [32]byte{},
	}
	s.sim.AddPort(s.testPort)

	s.testBinDir = "testbin"

	err := os.MkdirAll(s.testBinDir, os.ModePerm)
	s.Require().NoError(err, "Failed to create directory: ../../testbin")
}

func (s *Cpu8008Suite) run(command string, args ...string) (string, string, error) {
	cmd := exec.Command(command, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func (s *Cpu8008Suite) assemble(program string) string {
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

	_, stderr, err := s.run(ASL, "-cpu", "8008", "-L", asmFile, "-o", pFile)
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

func (s *Cpu8008Suite) AssembleAndLoad(program string) {
	var indentedProgram string
	header := `
cpu 8008new             ; use "new" 8008 mnemonics
radix 10                ; use base 10 for numbers
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
	err := s.ram.Load(binFile)
	s.Require().NoError(err, "Failed to load binary file: "+binFile)
}

func (s *Cpu8008Suite) TestIncrement() {
	s.cpu.Registers[REG_B] = 0 // Set register B to 0
	s.AssembleAndLoad(`
INR B
HLT
`)
	err := s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(1), s.cpu.Registers[REG_B])
	s.Equal(byte(0), s.cpu.Registers[FLAG_ZERO])
	s.Equal(byte(0), s.cpu.Registers[FLAG_PARITY])
	s.Equal(byte(0), s.cpu.Registers[FLAG_SIGN])

	s.cpu.PC = 0 // Reset program counter to start
	err = s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(2), s.cpu.Registers[REG_B])
	s.Equal(byte(0), s.cpu.Registers[FLAG_ZERO])
	s.Equal(byte(0), s.cpu.Registers[FLAG_PARITY])
	s.Equal(byte(0), s.cpu.Registers[FLAG_SIGN])

	s.cpu.PC = 0 // Reset program counter to start
	err = s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(3), s.cpu.Registers[REG_B])
	s.Equal(byte(0), s.cpu.Registers[FLAG_ZERO])
	s.Equal(byte(1), s.cpu.Registers[FLAG_PARITY])
	s.Equal(byte(0), s.cpu.Registers[FLAG_SIGN])

	s.cpu.Registers[REG_B] = 255 // Set register B to 255
	s.cpu.PC = 0                 // Reset program counter to start
	err = s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(0), s.cpu.Registers[REG_B])
	s.Equal(byte(1), s.cpu.Registers[FLAG_ZERO])
	s.Equal(byte(1), s.cpu.Registers[FLAG_PARITY])
	s.Equal(byte(0), s.cpu.Registers[FLAG_SIGN])
}

func (s *Cpu8008Suite) TestDecrement() {
	s.cpu.Registers[REG_B] = 1 // Set register B to 1
	s.AssembleAndLoad(`
DCR B
HLT
`)
	err := s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(0), s.cpu.Registers[REG_B])
	s.Equal(byte(1), s.cpu.Registers[FLAG_ZERO])
	s.Equal(byte(1), s.cpu.Registers[FLAG_PARITY])
	s.Equal(byte(0), s.cpu.Registers[FLAG_SIGN])

	s.cpu.PC = 0 // Reset program counter to start
	err = s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(255), s.cpu.Registers[REG_B])
	s.Equal(byte(0), s.cpu.Registers[FLAG_ZERO])
	s.Equal(byte(1), s.cpu.Registers[FLAG_PARITY])
	s.Equal(byte(1), s.cpu.Registers[FLAG_SIGN])

	s.cpu.PC = 0 // Reset program counter to start
	err = s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(254), s.cpu.Registers[REG_B])
	s.Equal(byte(0), s.cpu.Registers[FLAG_ZERO])
	s.Equal(byte(0), s.cpu.Registers[FLAG_PARITY])
	s.Equal(byte(1), s.cpu.Registers[FLAG_SIGN])
}

func (s *Cpu8008Suite) TestADI() {
	s.cpu.Registers[REG_A] = 0 // Set register B to 0
	s.AssembleAndLoad(`
ADI 1
HLT
`)
	err := s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(1), s.cpu.Registers[REG_A])
	s.Equal(byte(0), s.cpu.Registers[FLAG_ZERO])
	s.Equal(byte(0), s.cpu.Registers[FLAG_PARITY])
	s.Equal(byte(0), s.cpu.Registers[FLAG_SIGN])
	s.Equal(byte(0), s.cpu.Registers[FLAG_CARRY])

	s.cpu.PC = 0 // Reset program counter to start
	err = s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(2), s.cpu.Registers[REG_A])
	s.Equal(byte(0), s.cpu.Registers[FLAG_ZERO])
	s.Equal(byte(0), s.cpu.Registers[FLAG_PARITY])
	s.Equal(byte(0), s.cpu.Registers[FLAG_SIGN])
	s.Equal(byte(0), s.cpu.Registers[FLAG_CARRY])

	s.cpu.PC = 0 // Reset program counter to start
	err = s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(3), s.cpu.Registers[REG_A])
	s.Equal(byte(0), s.cpu.Registers[FLAG_ZERO])
	s.Equal(byte(1), s.cpu.Registers[FLAG_PARITY])
	s.Equal(byte(0), s.cpu.Registers[FLAG_SIGN])
	s.Equal(byte(0), s.cpu.Registers[FLAG_CARRY])

	s.cpu.Registers[REG_A] = 255 // Set register B to 255
	s.cpu.PC = 0                 // Reset program counter to start
	err = s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(0), s.cpu.Registers[REG_A])
	s.Equal(byte(1), s.cpu.Registers[FLAG_ZERO])
	s.Equal(byte(1), s.cpu.Registers[FLAG_PARITY])
	s.Equal(byte(0), s.cpu.Registers[FLAG_SIGN])
	s.Equal(byte(1), s.cpu.Registers[FLAG_CARRY])
}

func (s *Cpu8008Suite) TestSUI() {
	s.cpu.Registers[REG_A] = 1 // Set register A to 1
	s.AssembleAndLoad(`
SUI 1
HLT
`)
	err := s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(0), s.cpu.Registers[REG_A])
	s.Equal(byte(1), s.cpu.Registers[FLAG_ZERO])
	s.Equal(byte(1), s.cpu.Registers[FLAG_PARITY])
	s.Equal(byte(0), s.cpu.Registers[FLAG_SIGN])
	s.Equal(byte(0), s.cpu.Registers[FLAG_CARRY])

	s.cpu.PC = 0 // Reset program counter to start
	err = s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(255), s.cpu.Registers[REG_A])
	s.Equal(byte(0), s.cpu.Registers[FLAG_ZERO])
	s.Equal(byte(1), s.cpu.Registers[FLAG_PARITY])
	s.Equal(byte(1), s.cpu.Registers[FLAG_SIGN])
	s.Equal(byte(1), s.cpu.Registers[FLAG_CARRY])

	s.cpu.PC = 0 // Reset program counter to start
	err = s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(254), s.cpu.Registers[REG_A])
	s.Equal(byte(0), s.cpu.Registers[FLAG_ZERO])
	s.Equal(byte(0), s.cpu.Registers[FLAG_PARITY])
	s.Equal(byte(1), s.cpu.Registers[FLAG_SIGN])
	s.Equal(byte(0), s.cpu.Registers[FLAG_CARRY])
}

func (s *Cpu8008Suite) TestJZ() {
	s.AssembleAndLoad(`
ORA A
JZ L1
MVI B,2
HLT
L1:
MVI B,3
HLT
`)
	err := s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(3), s.cpu.Registers[REG_B])

	s.cpu.Registers[REG_A] = 1
	s.cpu.PC = 0
	err = s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(2), s.cpu.Registers[REG_B])
}

func (s *Cpu8008Suite) TestJNZ() {
	s.AssembleAndLoad(`
ORA A
JNZ L1
MVI B,2
HLT
L1:
MVI B,3
HLT
`)
	err := s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(2), s.cpu.Registers[REG_B])

	s.cpu.Registers[REG_A] = 1
	s.cpu.PC = 0
	err = s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(3), s.cpu.Registers[REG_B])
}

func (s *Cpu8008Suite) TestMOV() {
	s.AssembleAndLoad(`
MVI A, 12H
MOV B,A
INR B
MOV C,B
INR C
MOV D,C
INR D
MOV E,D
INR E
MOV H,E
INR H
MOV L,H
INR L
HLT
`)
	err := s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(0x12), s.cpu.Registers[REG_A])
	s.Equal(byte(0x13), s.cpu.Registers[REG_B])
	s.Equal(byte(0x14), s.cpu.Registers[REG_C])
	s.Equal(byte(0x15), s.cpu.Registers[REG_D])
	s.Equal(byte(0x16), s.cpu.Registers[REG_E])
	s.Equal(byte(0x17), s.cpu.Registers[REG_H])
	s.Equal(byte(0x18), s.cpu.Registers[REG_L])
}

func (s *Cpu8008Suite) TestMOVM() {
	err := s.ram.Write(0x1234, 0x56)
	s.NoError(err)
	s.AssembleAndLoad(`
MVI	H,12H
MVI L,34H
MOV A,M
ADI 7
MVI H,12H
MVI L,35H
MOV M,A
HLT
`)
	err = s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(0x12), s.cpu.Registers[REG_H])
	s.Equal(byte(0x35), s.cpu.Registers[REG_L])
	s.Equal(byte(0x5D), s.cpu.Registers[REG_A])
	value, err := s.ram.Read(0x1235)
	s.NoError(err)
	s.Equal(byte(0x5D), value)
}

func (s *Cpu8008Suite) TestCall() {
	s.AssembleAndLoad(`
MVI D, 45H
MVI E, 54H
CALL L1
HLT
L1:
INR D
CALL L2
INR E
RET
L2:
INR D
CALL L3
INR E
RET
L3:
INR D
CALL L4
INR E
RET
L4:
INR D
CALL L5
INR E
RET
L5:
INR D
CALL L6
INR E
RET
L6:
INR D
CALL L7
INR E
RET
L7:
MOV B,D
MOV C,E
RET
`)
	err := s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(0x4B), s.cpu.Registers[REG_B])
	s.Equal(byte(0x54), s.cpu.Registers[REG_C])
	s.Equal(byte(0x4B), s.cpu.Registers[REG_D])
	s.Equal(byte(0x5A), s.cpu.Registers[REG_E])
}

func (s *Cpu8008Suite) TestLogical() {
	s.AssembleAndLoad(`
MVI A, 12H
ANI 03H
MOV B,A
MVI A, 34H
ORI 01H
MOV C,A
MVI A, 56H
XRI 33H
MOV D,A
MVI A, 78H
CPI 78H
HLT
`)
	err := s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(0x78), s.cpu.Registers[REG_A])
	s.Equal(byte(0x02), s.cpu.Registers[REG_B])
	s.Equal(byte(0x35), s.cpu.Registers[REG_C])
	s.Equal(byte(0x65), s.cpu.Registers[REG_D])
	s.Equal(byte(1), s.cpu.Registers[FLAG_ZERO])
}

func (s *Cpu8008Suite) TestLogicalReg() {
	s.AssembleAndLoad(`
MVI A, 12H
MVI B, 03H
ANA B
MOV B,A
MVI A, 34H
MVI C, 01H
ORA C
MOV C,A
MVI A, 56H
MVI D, 33H
XRA D
MOV D,A
MVI A, 78H
CPI 78H
HLT
`)
	err := s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(0x78), s.cpu.Registers[REG_A])
	s.Equal(byte(0x02), s.cpu.Registers[REG_B])
	s.Equal(byte(0x35), s.cpu.Registers[REG_C])
	s.Equal(byte(0x65), s.cpu.Registers[REG_D])
	s.Equal(byte(1), s.cpu.Registers[FLAG_ZERO])
}

func (s *Cpu8008Suite) TestIn() {
	s.AssembleAndLoad(`
IN 3
HLT
`)
	err := s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(0xC3), s.cpu.Registers[REG_A])
}

func (s *Cpu8008Suite) TestOut() {
	s.AssembleAndLoad(`
MVI A,0A0h
OUT 08H
MVI A,0A1h
OUT 09H
MVI A,0A2h
OUT 0AH
MVI A,0A3h
OUT 0BH
MVI A,0A4h
OUT 0CH
MVI A,0A5h
OUT 0DH
MVI A,0A6h
OUT 0EH
MVI A,0A7h
OUT 0FH
MVI A,0A8h
OUT 10H
MVI A,0A9h
OUT 11H
MVI A,0AAh
OUT 12H
MVI A,0ABh
OUT 13H
MVI A,0ACh
OUT 14H
MVI A,0ADh
OUT 15H
MVI A,0AEh
OUT 16H
MVI A,0AFh
OUT 17H
MVI A,0B0h
OUT 18H
MVI A,0B1h
OUT 19H
MVI A,0B2h
OUT 1AH
MVI A,0B3h
OUT 1BH
MVI A,0B4h
OUT 1CH
MVI A,0B5h
OUT 1DH
MVI A,0B6h
OUT 1EH
MVI A,0B7h
OUT 1FH
HLT
`)
	err := s.cpu.Run()
	s.NoError(err)
	for i := 8; i < 32; i++ {
		s.Equal(byte(0xA0+(i-8)), s.testPort.out[i], "Output to port %02X should be %02X", i, byte(0xA0+(i-8)))
	}
}

func (s *Cpu8008Suite) TestRotate() {
	s.AssembleAndLoad(`
MVI	A,1
RLC
MOV	B,A
RRC
MOV C,A
MVI A,80h
RLC
MOV D,A
RRC
MOV E,A
HLT
`)
	err := s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(0x02), s.cpu.Registers[REG_B])
	s.Equal(byte(0x01), s.cpu.Registers[REG_C])
	s.Equal(byte(0x01), s.cpu.Registers[REG_D])
	s.Equal(byte(0x80), s.cpu.Registers[REG_E])
}

func (s *Cpu8008Suite) TestRAL() {
	s.AssembleAndLoad(`
MVI	A,80h
RAL
HLT
`)
	err := s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(0x00), s.cpu.Registers[REG_A])
	s.Equal(byte(1), s.cpu.Registers[FLAG_CARRY])
}

func (s *Cpu8008Suite) TestRALWithCarry() {
	s.cpu.Registers[FLAG_CARRY] = 1
	s.AssembleAndLoad(`
MVI	A,40h
RAL
HLT
`)
	err := s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(0x81), s.cpu.Registers[REG_A])
	s.Equal(byte(0), s.cpu.Registers[FLAG_CARRY])
}

func (s *Cpu8008Suite) TestRAR() {
	s.AssembleAndLoad(`
MVI	A,1h
RAR
HLT
`)
	err := s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(0x00), s.cpu.Registers[REG_A])
	s.Equal(byte(1), s.cpu.Registers[FLAG_CARRY])
}

func (s *Cpu8008Suite) TestRARWithCarrt() {
	s.cpu.Registers[FLAG_CARRY] = 1
	s.AssembleAndLoad(`
MVI	A,04h
RAR
HLT
`)
	err := s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(0x82), s.cpu.Registers[REG_A])
	s.Equal(byte(0), s.cpu.Registers[FLAG_CARRY])
}

func (s *Cpu8008Suite) TestRST() {
	s.AssembleAndLoad(`
JMP L1
ORG 8h
MVI	B,2
RET
ORG 10h
MVI C,3
RET
ORG 18h
MVI D,4
RET
ORG 20h
MVI E,5
RET
ORG 28h
MVI H,6
RET
ORG 30h
MVI L,7
RET
ORG 38h
MVI A,8
RET
L1:
RST 1
RST 2
RST 3
RST 4
RST 5
RST 6
RST 7
HLT
`)
	err := s.cpu.Run()
	s.NoError(err)
	s.Equal(byte(0x02), s.cpu.Registers[REG_B])
	s.Equal(byte(0x03), s.cpu.Registers[REG_C])
	s.Equal(byte(0x04), s.cpu.Registers[REG_D])
	s.Equal(byte(0x05), s.cpu.Registers[REG_E])
	s.Equal(byte(0x06), s.cpu.Registers[REG_H])
	s.Equal(byte(0x07), s.cpu.Registers[REG_L])
	s.Equal(byte(0x08), s.cpu.Registers[REG_A])
}

func TestCpu8008Suite(t *testing.T) {
	suite.Run(t, new(Cpu8008Suite))
}
