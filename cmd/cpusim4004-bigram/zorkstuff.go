package main

// Zorkinator support
//
// The whole reason for the "bigram" variant of the 4004 was to support playing Zork on the 4004.
// While implementing the Z-Machine, I added various debugging and profiling hooks to help me
// develop it. I've tried to confine those functions to this file, to make main.go a little
// less confusing.

import (
	"fmt"
	"github.com/scottmbaker/gocpusim/pkg/cpusim"
	"github.com/scottmbaker/gocpusim/pkg/cpusim/cpu4004"
	"sort"
)

type ProfileOperation struct {
	OpCode    string // use strings because opcode of multiple operand forms map to same base opcode
	Count     int
	Cycles    int
	MaxCycles int
}

const (
	OPCODE_STARTUP = "startup" // for the profiler, indicates startup
	OPCODE_BETWEEN = "between" // for the profiler, indicates time between opcodes
)

var Profiler map[string]*ProfileOperation = map[string]*ProfileOperation{}
var LastOpcode string = OPCODE_STARTUP
var LastOpcodeCycleStart int = 0

var opCodeName = map[int]string{}

func initOpcodes() {
	opCodeName[1] = "je"
	opCodeName[2] = "jl"
	opCodeName[3] = "jg"
	opCodeName[4] = "dec_chk"
	opCodeName[5] = "inc_chk"
	opCodeName[6] = "jin"
	opCodeName[7] = "test"
	opCodeName[8] = "or"
	opCodeName[9] = "and"
	opCodeName[10] = "test_attr"
	opCodeName[11] = "set_attr"
	opCodeName[12] = "clear_attr"
	opCodeName[13] = "store"
	opCodeName[14] = "insert_obj"
	opCodeName[15] = "loadw"
	opCodeName[16] = "loadb"
	opCodeName[17] = "get_prop"
	opCodeName[18] = "get_prop_addr"
	opCodeName[19] = "get_next_prop"
	opCodeName[20] = "add"
	opCodeName[21] = "sub"
	opCodeName[22] = "mul"
	opCodeName[23] = "div"
	opCodeName[24] = "mod"

	opCodeName[128] = "jz"
	opCodeName[129] = "get_sibling"
	opCodeName[130] = "get_child"
	opCodeName[131] = "get_parent"
	opCodeName[132] = "get_prop_len"
	opCodeName[133] = "inc"
	opCodeName[134] = "dec"
	opCodeName[135] = "print_addr"
	opCodeName[137] = "remove_obj"
	opCodeName[138] = "print_obj"
	opCodeName[139] = "ret"
	opCodeName[140] = "jump"
	opCodeName[141] = "print_paddr"
	opCodeName[142] = "load"
	opCodeName[143] = "not"

	opCodeName[176] = "rtrue"
	opCodeName[177] = "rfalse"
	opCodeName[178] = "print"
	opCodeName[179] = "print_ret"
	opCodeName[180] = "nop"
	opCodeName[181] = "save"
	opCodeName[182] = "restore"
	opCodeName[183] = "restart"
	opCodeName[184] = "ret_popped"
	opCodeName[185] = "pop"
	opCodeName[186] = "quit"
	opCodeName[187] = "new_line"

	opCodeName[224] = "call"
	opCodeName[225] = "storew"
	opCodeName[226] = "storeb"
	opCodeName[227] = "put_prop"
	opCodeName[228] = "read"
	opCodeName[229] = "print_char"
	opCodeName[230] = "print_num"
	opCodeName[231] = "random"
	opCodeName[232] = "push"
	opCodeName[233] = "pull"

	opCodeName[188] = "show_status"
	opCodeName[189] = "verify"
	opCodeName[234] = "split_window"
	opCodeName[235] = "set_window"
	opCodeName[243] = "output_stream"
	opCodeName[244] = "input_stream"
	opCodeName[245] = "sound_effect"

	for i := 32; i <= 127; i++ { // 2OP opcodes repeating with different operand forms.
		opCodeName[i] = opCodeName[i%32]
	}
	for i := 144; i <= 175; i++ { // 1OP opcodes repeating with different operand forms.
		opCodeName[i] = opCodeName[128+(i%16)]
	}
	for i := 192; i <= 223; i++ { // 2OP opcodes repeating with VAR operand forms.
		opCodeName[i] = opCodeName[i%32]
	}
}

func closeProfilerOpcode(sim *cpusim.CpuSim, nextOpcode byte, runSignal bool) {
	if !debugProfile {
		return
	}

	cpu := sim.CPU[0].(*cpu4004.CPU4004)

	opcode := LastOpcode

	cycles := cpu.Cycles - LastOpcodeCycleStart

	nextOpcodeName := ""
	var okay bool

	if runSignal {
		nextOpcodeName = OPCODE_BETWEEN
	} else {
		nextOpcodeName, okay = opCodeName[int(nextOpcode)]
		if !okay {
			nextOpcodeName = fmt.Sprintf("unknown_%02X", int(nextOpcode))
		}
	}

	LastOpcode = nextOpcodeName
	LastOpcodeCycleStart = cpu.Cycles

	prof, exists := Profiler[opcode]
	if !exists {
		prof = &ProfileOperation{OpCode: opcode, Count: 0, Cycles: 0, MaxCycles: 0}
		Profiler[opcode] = prof
	}
	prof.Count++
	prof.Cycles += cycles
	if cycles > prof.MaxCycles {
		prof.MaxCycles = cycles
	}
}

func printProfiler() {
	fmt.Printf("\nProfiler Results:\n")
	fmt.Printf("Opcode          Count       Cycles      Avg Cycles  Max Cycles\n")
	fmt.Printf("---------------------------------------------------------------\n")
	keys := make([]string, 0, len(Profiler))
	for key := range Profiler {
		keys = append(keys, key)
	}

	sort.Slice(keys, func(i, j int) bool {
		avgCyclesI := 0
		if Profiler[keys[i]].Count > 0 {
			avgCyclesI = Profiler[keys[i]].Cycles / Profiler[keys[i]].Count
		}
		avgCyclesJ := 0
		if Profiler[keys[j]].Count > 0 {
			avgCyclesJ = Profiler[keys[j]].Cycles / Profiler[keys[j]].Count
		}
		return avgCyclesI > avgCyclesJ
	})

	for _, key := range keys {
		prof := Profiler[key]
		avgCycles := 0
		if prof.Count > 0 {
			avgCycles = prof.Cycles / prof.Count
		}
		fmt.Printf("%-15s %-11d %-11d %-11d %-11d\n", prof.OpCode, prof.Count, prof.Cycles, avgCycles, prof.MaxCycles)
	}
}

func insDebug(sim *cpusim.CpuSim) {
	opCount, _ := BigRamLink.Read(0x20000 + 0x20 + 1)

	opcode, _ := BigRamLink.Read(0x20000 + 0x22 + 1)
	operands := []uint16{}
	opsAddr := 0x20000 + 0

	closeProfilerOpcode(sim, opcode, false)

	for i := 0; i < int(opCount); i++ {
		operand := (uint16(BigRamLink.Contents[opsAddr+i*2])<<8 | uint16(BigRamLink.Contents[opsAddr+i*2+1]))
		operands = append(operands, operand)
	}

	pc := uint32(BigRamLink.Contents[0x20000+0x25])<<16 | uint32(BigRamLink.Contents[0x20000+0x26])<<8 | uint32(BigRamLink.Contents[0x20000+0x27])

	if debug2 {
		fmt.Printf("> %04X: %02X (%s) [", pc, opcode, opCodeName[int(opcode)])
		for i, operand := range operands {
			if i > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("%04X", operand)
		}
		fmt.Printf("]\n")
	}
}

func runSignal(sim *cpusim.CpuSim) {
	closeProfilerOpcode(sim, 0, true)
}

func branchDebug(sim *cpusim.CpuSim) {
	if !debug3 {
		return
	}

	r4 := sim.CPU[0].(*cpu4004.CPU4004).Registers[4]
	r5 := sim.CPU[0].(*cpu4004.CPU4004).Registers[5]
	r6 := sim.CPU[0].(*cpu4004.CPU4004).Registers[6]
	r7 := sim.CPU[0].(*cpu4004.CPU4004).Registers[7]
	r8 := sim.CPU[0].(*cpu4004.CPU4004).Registers[8]
	r9 := sim.CPU[0].(*cpu4004.CPU4004).Registers[9]
	r10 := sim.CPU[0].(*cpu4004.CPU4004).Registers[10]
	r11 := sim.CPU[0].(*cpu4004.CPU4004).Registers[11]
	r12 := sim.CPU[0].(*cpu4004.CPU4004).Registers[12]
	r13 := sim.CPU[0].(*cpu4004.CPU4004).Registers[13]

	offset := int(r4)<<12 | int(r5)<<8 | int(r6)<<4 | int(r7)
	pc := int(r8)<<20 | int(r9)<<16 | int(r10)<<12 | int(r11)<<8 | int(r12)<<4 | int(r13)

	fmt.Printf("    branch debug: offset=%04X pc=%06X\n", offset, pc)
}

func ZSCIILookup(val int) rune {
	if val == 5 {
		return '5'
	} else {
		return rune(val + 'a' - 6)
	}
}

func getZSCII(w int) string {
	chars := []rune{}
	chars = append(chars, ZSCIILookup(w>>10&0x1F))
	chars = append(chars, ZSCIILookup(w>>5&0x1F))
	chars = append(chars, ZSCIILookup(w&0x1F))
	return string(chars)
}

func debugLine(sim *cpusim.CpuSim) {
	pch, _ := BigRamLink.Read(0x20000 + 0x11)
	pclh, _ := BigRamLink.Read(0x20000 + 0x12)
	pcll, _ := BigRamLink.Read(0x20000 + 0x13)

	mph, _ := BigRamLink.Read(0x20000 + 0x15)
	mplh, _ := BigRamLink.Read(0x20000 + 0x16)
	mpll, _ := BigRamLink.Read(0x20000 + 0x17)

	abbvh, _ := BigRamLink.Read(0x20000 + 0x18)
	abbvl, _ := BigRamLink.Read(0x20000 + 0x19)

	globh, _ := BigRamLink.Read(0x20000 + 0x1A)
	globl, _ := BigRamLink.Read(0x20000 + 0x1B)
	globAddr := int(globh)<<8 | int(globl)

	sph, _ := BigRamLink.Read(0x20000 + 0x1C)
	spl, _ := BigRamLink.Read(0x20000 + 0x1D)

	bph, _ := BigRamLink.Read(0x20000 + 0x1E)
	bpl, _ := BigRamLink.Read(0x20000 + 0x1F)

	tok0h, _ := BigRamLink.Read(0x20000 + 0x44)
	tok0l, _ := BigRamLink.Read(0x20000 + 0x45)
	tok1h, _ := BigRamLink.Read(0x20000 + 0x46)
	tok1l, _ := BigRamLink.Read(0x20000 + 0x47)

	tok0 := int(tok0h)<<8 | int(tok0l)
	tok1 := int(tok1h)<<8 | int(tok1l)

	dicth, _ := BigRamLink.Read(0x20000 + 0x0c)
	dictl, _ := BigRamLink.Read(0x20000 + 0x0d)
	dictCounth := BigRamLink.Contents[0x20000+0x0e]
	dictCountl := BigRamLink.Contents[0x20000+0x0f]

	opsAddr := 0x20000 + 0

	fmt.Printf("\n")
	fmt.Printf("> %04X: ", sim.CPU[0].(*cpu4004.CPU4004).PC)
	fmt.Printf(" %s\n", sim.CPU[0].String())
	fmt.Printf("> PC=%02X%02X%02X MP=%02X%02X%02X ABBV=%02X%02X GLOB=%02X%02X SP=%02X%02X BP=%02X%02X DICT=%02X%02X DCOUNT=%02X%02X\n", pch, pclh, pcll, mph, mplh, mpll, abbvh, abbvl, globh, globl, sph, spl, bph, bpl, dicth, dictl, dictCounth, dictCountl)

	fmt.Printf("> GLOBS: ")
	for i := 0; i < 16; i++ {
		fmt.Printf("%d:%02X%02X ", i, BigRamLink.Contents[globAddr+i*2], BigRamLink.Contents[globAddr+i*2+1])
	}
	fmt.Printf("\n")

	fmt.Printf("> OPS: ")
	for i := 0; i < 8; i++ {
		fmt.Printf("%d:%02X%02X ", i, BigRamLink.Contents[opsAddr+i*2], BigRamLink.Contents[opsAddr+i*2+1])
	}
	fmt.Printf("\n")

	stackAddr := 0x20000 + 0x100
	fmt.Printf("> STACK: ")
	for i := 0; i < 8; i++ {
		fmt.Printf("%d:%02X%02X ", i, BigRamLink.Contents[stackAddr+i*2], BigRamLink.Contents[stackAddr+i*2+1])
	}
	fmt.Printf("\n")

	fmt.Printf("> TOKENS: %04X %04X %s%s\n", tok0, tok1, getZSCII(tok0), getZSCII(tok1))

	parseh, _ := BigRamLink.Read(0x20000 + 0x48)
	parsel, _ := BigRamLink.Read(0x20000 + 0x49)
	parseAddr := int(parseh)<<8 | int(parsel)
	parseMax, _ := BigRamLink.Read(cpusim.Address(parseAddr))
	parseCount, _ := BigRamLink.Read(cpusim.Address(parseAddr + 1))
	fmt.Printf("> Parse (%04X,%02X,%02X): ", parseAddr, parseMax, parseCount)
	for i := 0; i < int(min(parseCount, 8)); i++ {
		wordh, _ := BigRamLink.Read(cpusim.Address(parseAddr + 2 + i*4))
		wordl, _ := BigRamLink.Read(cpusim.Address(parseAddr + 3 + i*4))
		word := int(wordh)<<8 | int(wordl)
		len, _ := BigRamLink.Read(cpusim.Address(parseAddr + 4 + i*4))
		pos, _ := BigRamLink.Read(cpusim.Address(parseAddr + 5 + i*4))
		fmt.Printf(" %04X/%02X/%02X", word, len, pos)
	}
	fmt.Printf("\n")

	/*
		objAddr := 0x20000 + 0x400
		fmt.Printf("> OBJS: ")
		for i := 0; i < 6; i++ {
			fmt.Printf("%d:%02X%02X ", i, BigRamLink.Contents[objAddr+i*2], BigRamLink.Contents[objAddr+i*2+1])
		}
		fmt.Printf("\n")
	*/
}
