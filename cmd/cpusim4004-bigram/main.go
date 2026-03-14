package main

// go-cpusim
// Scott Baker
//
// A 4004 CPU similator written in Go.

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/scottmbaker/gocpusim/pkg/cpusim"
	"github.com/scottmbaker/gocpusim/pkg/cpusim/cpu4004"
	"github.com/spf13/cobra"
)

const (
	UART_DATA_R    = 0xE0
	UART_DATA_W    = 0xE0
	UART_CONTROL_R = 0xE1
	UART_CONTROL_W = 0xE1
)

var (
	debug        bool
	debug2       bool
	debug3       bool
	debugProfile bool
	memDebug     bool
	exitEof      bool
	startAddr    int
	romFilename  string
	z3Filename   string
	inFilename   string
	rootCmd      = &cobra.Command{
		Use:   "cpusim4004",
		Short: "scott's 4004 cpu simulator",
		Long:  "A simulator for the 4004 CPU. For a quick demo, try \"cpusim -f roms/sbc-8251.rom\"",
	}
)

var BigRamLink *cpusim.Memory

func newScottSingleBoardComputer() (*cpusim.CpuSim, *cpusim.UART) {
	sim := cpusim.NewCPUSim()
	sim.SetDebug(debug)
	sim.SetMemDebug(memDebug)

	initOpcodes()

	// Create an 8008 CPU and attach it to the emulator
	cpu := cpu4004.New4004(sim, "cpu")
	cpu.SetDebugLine(debugLine)
	cpu.SetDebugTwo(insDebug)
	cpu.SetDebugThree(branchDebug)
	cpu.SetDebugFour(runSignal)
	sim.AddCPU(cpu)

	if startAddr != 0 {
		cpu.PC = uint16(startAddr)
	}

	// Setup a mapper for the ROM. It will only filter KIND_ROM devices.
	// We will attach it to the 4289's ROM port.

	// Hi mapper for A14..A17. It uses addresses 0x04 - 0x07 (is this shift-lefted? why? shouldn't it be 0x40 to 0x70?).
	mapper2 := cpusim.New74670(sim, "mapper2", 0x04, cpusim.A10, cpusim.D0, cpusim.A14, cpusim.A15, cpusim.A16, cpusim.A17, &cpusim.AlwaysEnabled, &cpusim.AlwaysEnabled)
	mapper2.FilterMemoryKind(cpusim.KIND_ROM)
	sim.AddMapper(mapper2)

	// Lo mapper for A10..A13. Do this after the hi mapper, otherwise lo mapper changing A10 will break hi mapper
	mapper := cpusim.New74670(sim, "mapper", 0x00, cpusim.A10, cpusim.D0, cpusim.A10, cpusim.A11, cpusim.A12, cpusim.A13, &cpusim.AlwaysEnabled, &cpusim.AlwaysEnabled)
	mapper.FilterMemoryKind(cpusim.KIND_ROM)
	sim.AddMapper(mapper)

	rom := cpusim.NewMemory(sim, "rom", cpusim.KIND_ROM, 0x0000, 0x3FFFF, 12, true, &cpusim.TrueEnabler{}) // 256 KB of ROM on the bigramboard
	sim.AddMemory(rom)

	ram := cpusim.NewMemory(sim, "ram", cpusim.KIND_RAM, 0x0000, 0x7F, 7, false, cpu.DCLEnabler(0))
	ram.CreateStatusBytes(0x08, 0x04)
	sim.AddMemory(ram)

	b8b := cpu4004.NewBus8Bit(sim, "bus8", cpu.DCLEnabler(4))
	sim.AddMemory(b8b)

	romPort := cpu4004.NewRomPort(sim, "romport_4289", &cpusim.TrueEnabler{})
	romPort.AddPort(mapper)
	romPort.AddPort(mapper2)
	sim.AddPort(romPort)

	// Create an 8251 UART
	uart := cpusim.NewUART(sim, "uart", UART_DATA_R, UART_DATA_W, UART_CONTROL_R, UART_CONTROL_W, &cpusim.AlwaysEnabled)
	b8b.AddPort(uart)

	// Add the bigram
	bigram := cpusim.NewMemory(sim, "ram", cpusim.KIND_RAM, 0x0000, 0x3FFFF, 16, false, &cpusim.AlwaysEnabled)
	b8b.AddMemory(bigram)

	BigRamLink = bigram

	// mappers for bigram
	mapperBigRam0 := cpusim.New74173(sim, "bigram_mapper_A8", 0x08, cpusim.A8, cpusim.A9, cpusim.A10, cpusim.A11, &cpusim.AlwaysEnabled)
	b8b.AddMapper(mapperBigRam0)
	romPort.AddPort(mapperBigRam0)

	mapperBigRam1 := cpusim.New74173(sim, "bigram_mapper_A12", 0x09, cpusim.A12, cpusim.A13, cpusim.A14, cpusim.A15, &cpusim.AlwaysEnabled)
	b8b.AddMapper(mapperBigRam1)
	romPort.AddPort(mapperBigRam1)

	mapperBigRam2 := cpusim.New74173(sim, "bigram_mapper_A16", 0x0A, cpusim.A16, cpusim.A17, cpusim.A18, cpusim.A19, &cpusim.AlwaysEnabled)
	b8b.AddMapper(mapperBigRam2)
	romPort.AddPort(mapperBigRam2)

	// Next we load the ROM, from a file on disk.
	err := rom.Load(romFilename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load ROM file '%s': %v\n", romFilename, err)
		os.Exit(1)
	}

	if z3Filename != "" {
		err := bigram.Load(z3Filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to load Z3 file '%s': %v\n", z3Filename, err)
			os.Exit(1)
		}
		rom.Contents[81920] = 0xC0 // BBL 0 to disable loader
	}

	if inFilename != "" {
		err := uart.LoadInputFile(inFilename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to load uart input file '%s': %v\n", inFilename, err)
			os.Exit(1)
		}
	}

	if exitEof {
		uart.SetExitOnEof(true)
	}

	return sim, uart
}

func cleanup(sim *cpusim.CpuSim, uart *cpusim.UART) {
	fmt.Printf(">>>>> Terminated <<<<<\n")
	debugLine(sim)

	if debugProfile {
		printProfiler()
	}

	// Stop raw mode terminal
	uart.RestoreTerminal()
}

func mainCommand(cmd *cobra.Command, args []string) {
	var wg sync.WaitGroup

	if romFilename == "" {
		fmt.Fprintf(os.Stderr, "Error: --rom-file is required\n")
		_ = cmd.Help()
		return
	}

	sim, uart := newScottSingleBoardComputer()

	// start the simulator. It will start executing code immadiately.
	sim.Start(&wg)

	//. Start the UART. It will switch the ternminal to raw input and start processing keystrokes.
	uart.Start(&wg)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		// Block until a signal is received.
		<-signalChan
		fmt.Println("Received interrupt signal.")
		cleanup(sim, uart) // Call the desired function
		os.Exit(0)         // Exit gracefully after cleanup
	}()

	// Wait for all goroutines to complete
	wg.Wait()

	cleanup(sim, uart)
}

func main() {
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "debug messages")
	rootCmd.PersistentFlags().BoolVarP(&memDebug, "memDebug", "m", false, "memory debug messages")
	rootCmd.PersistentFlags().BoolVar(&debug2, "debug2", false, "debug2 messages")
	rootCmd.PersistentFlags().BoolVar(&debug3, "debug3", false, "debug3 messages")
	rootCmd.PersistentFlags().BoolVar(&debugProfile, "debugProfile", false, "enable profiling")
	rootCmd.PersistentFlags().IntVarP(&startAddr, "startAddr", "s", 0, "start address")
	rootCmd.PersistentFlags().StringVarP(&romFilename, "rom-file", "f", "", "rom filename")
	rootCmd.PersistentFlags().StringVarP(&z3Filename, "z3-file", "z", "", "z3 filename")
	rootCmd.PersistentFlags().StringVarP(&inFilename, "in-file", "i", "", "text input filename")
	rootCmd.PersistentFlags().BoolVarP(&exitEof, "exitEof", "e", false, "exit on text eof")
	rootCmd.Run = mainCommand

	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
}
