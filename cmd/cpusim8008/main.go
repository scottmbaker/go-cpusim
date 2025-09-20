package main

// go-cpusim
// Scott Baker
//
// An 8008 CPU similator written in Go.

import (
	"fmt"
	"github.com/scottmbaker/gocpusim/pkg/cpusim"
	"github.com/scottmbaker/gocpusim/pkg/cpusim/cpu8008"
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
		Use:   "cpusim",
		Short: "scott's 8008 cpu simulator",
		Long:  "A simulator for the 8008 CPU. For a quick demo, try \"cpusim -f roms/sbc-8251.rom\"",
	}
)

/* neewScottSingleBoardComputer
 *
 * Here is an example of creating a computer. We need to create a simulator and then
 * bind the following resources to it:
 *
 *  - CPU: So it can execute instructions
 *  - ROM: Read-only storage to hold the program to execute
 *  - RAM: Read/wri9te storage for the program to store data
 *  - UART: Serial port for input/output
 *
 * Additionally, we can add other devices, such as:
 *
 *  - Dipswitch: An input port with a fixed value
 *  - Mapper: A device that maps RAM/ROM address space to CPU address space.
 */

func newScottSingleBoardComputer() (*cpusim.CpuSim, *cpusim.UART) {
	sim := cpusim.NewCPUSim()
	sim.SetDebug(debug)

	// Create an 8008 CPU and attach it to the emulator
	cpu := cpu8008.New8008(sim, "cpu")
	sim.AddCPU(cpu)

	/* Create a 74LS670 mamemory mapper
	 *
	 * The mapper is used to allow more memory space than the CPU can address directly. The 8008 can address 16 kilobytes of memory.
	 * When I designed my single-board-computer, I wanted to support up to 64 kilobytes of ROM, and up to 16 kilobytes or RAM. The
	 * mapper lets us do that.
	 *
	 * Basically the mapper is a file of four 8-bit registers. These registers are addressed by the A12 and A13 address lines. This
	 * means the mapper has 4 KB granularity. So you have four pages:
	 *   Page 0: CPU address space 0 - 0xFFF
	 *   Page 1: CPU address space 0x1000 - 0x1FFF
	 *   Page 2: CPU address space 0x2000 - 0x2FFF
	 *   Page 3: CPU address space 0x3000 - 0x3FFF
	 *
	 * Each page has an 8-bit register associated with it. The first 4 bits of the register are A12..A15. This gives us a space of up
	 * to 64 kilobytes. The hi bit of the register is a chip select for the RAM/ROM. If the hi bit is set, then the RAM device will be
	 * selected. If the hi bit is clear, then the ROM device will be selected.
	 *
	 * When the mapper initializes, all the registers are set to 0x00. This means that for each of the 4K pages, the ROM chip select will
	 * be active, and ROM page 0 will be mapped to the page.
	 *
	 * Typically the ROM begins executing at 0x0000, then immediately jumps to an address around 0x2000. Then it initializes the mapper to
	 * map RAM to the first two pages. For example, when running the monitor,
	 *   Page 0: 0x0000 - 0x0FFF maps to RAM page 0
	 *   Page 1: 0x1000 - 0x1FFF maps to RAM page 1
	 *   Page 2: 0x2000 - 0x2FFF maps to ROM page 0
	 *   Page 3: 0x3000 - 0x3FFF maps to ROM page 1
	 *
	 * IF you press the "B" key in the monitor, it will load BASIC by remapping pages 2 and 3 to point at SCELBAL instead of the monitor.
	 */

	ramRomEnable := cpusim.NewEnableBit()
	mapper := cpusim.New74670(sim, "mapper", 0x0C, cpusim.A12, cpusim.D0, cpusim.A12, cpusim.A13, cpusim.A14, cpusim.A15, &cpusim.AlwaysEnabled)
	mapper.ConnectEnableBit(7, ramRomEnable)
	sim.AddMapper(mapper)
	sim.AddPort(mapper)

	// Create RAM and ROM. These are both 32K in size. The 74670 mapper will map the appropriate regions and toggle the
	// chip select as appropriate.
	ram := cpusim.NewMemory(sim, "ram", 0x0000, 0xFFFF, 16, false, &ramRomEnable.HiEnable)
	rom := cpusim.NewMemory(sim, "rom", 0x0000, 0xFFFF, 16, true, &ramRomEnable.LoEnable)
	sim.AddMemory(ram)
	sim.AddMemory(rom)

	// Create an 8251 UART
	uart := cpusim.NewUART(sim, "uart", UART_DATA_R, UART_DATA_W, UART_CONTROL_R, UART_CONTROL_W, &cpusim.AlwaysEnabled)
	sim.AddPort(uart)

	// There's a bug in the monitor that will abort the memory dump command if it reads a low bit 0 from port 0.
	// (it's actually a feature, intended to support bit-bang serial, but in this case it's an undesirable feature!)
	dipswitch := cpusim.NewDipSwitch(sim, "dipswitch", 0x00, 0xFF, &cpusim.AlwaysEnabled)
	sim.AddPort(dipswitch)

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
