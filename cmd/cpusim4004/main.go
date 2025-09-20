package main

// go-cpusim
// Scott Baker
//
// An 4004 CPU similator written in Go.

import (
	"fmt"
	"github.com/scottmbaker/gocpusim/pkg/cpusim"
	"github.com/scottmbaker/gocpusim/pkg/cpusim/cpu4004"
	"github.com/spf13/cobra"
	"os"
	"sync"
)

const (
	UART_DATA_R    = 2
	UART_DATA_W    = 0x12
	UART_CONTROL_R = 3
	UART_CONTROL_W = 0x13
)

var (
	debug       bool
	romFilename string
	rootCmd     = &cobra.Command{
		Use:   "cpusim8008",
		Short: "scott's 4004 cpu simulator",
		Long:  "A simulator for the 4004 CPU. For a quick demo, try \"cpusim -f roms/sbc-8251.rom\"",
	}
)

func newScottSingleBoardComputer() (*cpusim.CpuSim, *cpusim.UART) {
	sim := cpusim.NewCPUSim()
	sim.SetDebug(debug)

	// Create an 8008 CPU and attach it to the emulator
	cpu := cpu8008.New4004(sim, "cpu")
	sim.AddCPU(cpu)

	mapper := cpusim.New74670(sim, "mapper", 0x0C, cpusim.A12, cpusim.D0, cpusim.A12, cpusim.A13, cpusim.A14, cpusim.A15, &cpusim.AlwaysEnabled)
	sim.AddMapper(mapper)
	sim.AddPort(mapper)

	ram := cpusim.NewMemory(sim, "ram", 0x0000, 0xFFFF, 16, false, &ramRomEnable.HiEnable)
	rom := cpusim.NewMemory(sim, "rom", 0x0000, 0xFFFF, 16, true, &ramRomEnable.LoEnable)
	sim.AddMemory(ram)
	sim.AddMemory(rom)

	// Create an 8251 UART
	uart := cpusim.NewUART(sim, "uart", UART_DATA_R, UART_DATA_W, UART_CONTROL_R, UART_CONTROL_W, &cpusim.AlwaysEnabled)
	sim.AddPort(uart)

	// Next we load the ROM, from a file on disk.
	err := rom.Load(romFilename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load ROM file '%s': %v\n", romFilename, err)
		os.Exit(1)
	}

	return sim, uart
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

	// Wait for all goroutines to complete
	wg.Wait()

	// Stop raw mode terminal
	uart.RestoreTerminal()
}

func main() {
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "debug messages")
	rootCmd.PersistentFlags().StringVarP(&romFilename, "rom-file", "f", "", "rom filename")
	rootCmd.Run = mainCommand

	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
}
