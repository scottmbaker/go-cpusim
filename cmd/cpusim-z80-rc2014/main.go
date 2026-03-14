package main

// go-cpusim
// Scott Baker
//
// A Z80 CPU simulator written in Go. This emulates a simple RC2014 contifiguration with a
// Zeta-2 style memory mapper, an ACIA, and 512K each of RAM and ROM.

import (
	"fmt"
	"os"
	"sync"

	"github.com/scottmbaker/gocpusim/pkg/cpusim"
	"github.com/scottmbaker/gocpusim/pkg/cpusim/cpuz80"
	"github.com/spf13/cobra"
)

const (
	ACIA_DATA    = 0x81
	ACIA_CONTROL = 0x80

	SIO_CTRL_A = 0x80
	SIO_DATA_A = 0x81
	SIO_CTRL_B = 0x82
	SIO_DATA_B = 0x83

	CF_BASE = 0x10
)

var (
	debug       bool
	memDebug    bool
	romFilename string
	serial      string
	cfImage     string
	cfIdentify  string
	cfOffset    int64
	rootCmd     = &cobra.Command{
		Use:   "cpusimz80",
		Short: "scott's Z80 cpu simulator",
		Long:  "A simulator for the Z80 CPU.",
	}
)

func newZ80Computer() (*cpusim.CpuSim, cpusim.UartInterface) {
	sim := cpusim.NewCPUSim()
	sim.SetDebug(debug)
	sim.SetMemDebug(memDebug)

	// Create a Z80 CPU and attach it to the simulator
	cpu := cpuz80.NewZ80(sim, "cpu")
	cpu.PortAddressMask = 0xFF // Use 8-bit port addresses for Z80
	sim.AddCPU(cpu)

	mapEnable := cpusim.NewEnableBit()

	// ZETA-2 style memory mapper
	// Port 0x78-0x7B are the page select bits for the mapper. Port 0x7C bit 0 is the enable bit for the mapper.
	// D6 of the mapper output is used as the enable for the RAM/ROM. When low, ROM is selected. When high, RAM is selected.
	ramRomEnable := cpusim.NewEnableBit()
	mapper := cpusim.NewDual74670(sim, "mapper-lo", 0x78, cpusim.A14, cpusim.D0, cpusim.A14, cpusim.A15, cpusim.A16, cpusim.A17, cpusim.A18, -1, -1, -1, &cpusim.AlwaysEnabled, &mapEnable.HiEnable)
	mapper.ConnectEnableBit(5, ramRomEnable)
	sim.AddMapper(mapper)
	sim.AddPort(mapper)

	// When a "1" is written to D0 in the latch, the memory mapper should be enabled
	mapperEnableLatch := cpusim.NewGenericOutputPort(sim, "mapper-enable-latch", 0x7C, 0, &cpusim.AlwaysEnabled)
	mapperEnableLatch.ConnectEnableBit(0, mapEnable)
	sim.AddPort(mapperEnableLatch)

	// 512KB RAM
	ram := cpusim.NewMemory(sim, "ram", cpusim.KIND_RAM, 0x0000, 0x7FFFF, 19, false, &ramRomEnable.HiEnable)
	sim.AddMemory(ram)

	// 512KB ROM
	rom := cpusim.NewMemory(sim, "rom", cpusim.KIND_ROM, 0x0000, 0x7FFFF, 19, true, &ramRomEnable.LoEnable)
	sim.AddMemory(rom)

	// UART on I/O ports
	serialIO := cpusim.NewStdioSerial(true)
	var uart cpusim.UartInterface
	if serial == "acia" {
		acia := cpusim.NewACIA(sim, serialIO, "uart", ACIA_DATA, ACIA_CONTROL, &cpusim.AlwaysEnabled)
		sim.AddPort(acia)
		uart = acia
	} else if serial == "sio" {
		sio := cpusim.NewSIO(sim, serialIO, "uart", SIO_DATA_A, SIO_DATA_B, SIO_CTRL_A, SIO_CTRL_B, &cpusim.AlwaysEnabled)
		sim.AddPort(sio)
		uart = sio
	} else {
		fmt.Fprintf(os.Stderr, "Error: invalid serial device type '%s'. Valid options are 'acia' and 'sio'.\n", serial)
		os.Exit(1)
	}

	// CompactFlash on I/O ports
	if cfImage != "" {
		cf := cpusim.NewCompactFlash(sim, "cf", CF_BASE, &cpusim.AlwaysEnabled)
		err := cf.AttachImage(cfImage, cfOffset)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if cfIdentify != "" {
			err = cf.LoadIdentify(cfIdentify)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading identify from file '%s': %v\n", cfIdentify, err)
				os.Exit(1)
			}
		} else if cfOffset > 0 {
			// if the CF is offset and not identify file is given, assume it's an emulatorkit-style image with the identify block at offset 512
			err = cf.LoadIdentifyFromImage()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading identify from image: %v\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Fprintf(os.Stderr, "Error: --cf-identify is required when using a raw CF image\n")
		}

		sim.AddPort(cf)
	}

	// Load ROM file into RAM at 0x0000
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

	sim, uart := newZ80Computer()

	sim.Start(&wg)
	uart.Start(&wg)
	wg.Wait()
	uart.RestoreTerminal()
}

func main() {
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "debug messages")
	rootCmd.PersistentFlags().BoolVarP(&memDebug, "memDebug", "m", false, "memory debug messages")
	rootCmd.PersistentFlags().StringVarP(&serial, "serial", "s", "acia", "type of serial device to use (acia, sio)")
	rootCmd.PersistentFlags().StringVarP(&romFilename, "rom-file", "f", "", "rom filename")
	rootCmd.PersistentFlags().StringVar(&cfImage, "cf-image", "", "CompactFlash disk image file")
	rootCmd.PersistentFlags().StringVar(&cfIdentify, "cf-identify", "", "CompactFlash identify block file (512 bytes)")
	rootCmd.PersistentFlags().Int64Var(&cfOffset, "cf-offset", 0, "byte offset to sector 0 in CF image (1024 for emulatorkit, 0 for raw)")
	rootCmd.Run = mainCommand

	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
}
