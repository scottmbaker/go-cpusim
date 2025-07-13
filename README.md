# An 8008 CPU Emulator in Go

Scott Baker, https://medium.com/@smbaker

This repository contains a CPU emulator in Go. The emulator is
intended to be extensible to different CPUs, though for now my
intention is only to emulate the Intel 8008.

This project is the result of a personal challenge to myself to
write a CPU emulator, in go, in one day. I pulled it off, but it was
one really long day... :)

## Hardware Emulated

* CPU. The CPU reads instructions from memory and executes them. This
  emulation is necessarily specific to a CPU, as CPUs generally have
  different instruction sets.

* Memory. Memory may be RAM (Random Access Memory, Read/Write) or ROM
  (Read Only Memory). Generally the emulator would be configured with
  one ROM device, to hold program contents and one RAM device to service
  as transient data storage space.

* UART. UART is a Universal Asynchronour Receiver Transmitter, and its
  job is basically to provide serial IO. This lets us interact with the
  running program with a keyboard and screen.

* Memory Mapper. The memory mapper allows you to have more physical memory
  than the CPU's address space, via a bank-switching scheme. It also
  serves a useful function in bootstrapping -- people like to locate their
  ROM high (for example at 0x2000) and their RAM low (for example at
  0x0000) yet the 8008 begins execution at address 0. So we use the
  memory mapper locate ROM at 0x0000 on bootstrap and then later remap
  that space as RAM.

I should stree that this emulates a real, functional computer. I have
assembled an 8008-based single board computer with exactly the hardware
elements listed here. 

## Software Supported

Basically anything that can be started from ROM is supported.

I have included the ROM image from my single-board computer, which has
the following built into it:

* Monitor. A machine-language monitor that lets you dump memory, examine
  registers, etc. The Monitor was adapted from Jim Loos's 8008 single
  board computer project. It's an excellent full-featured monitor.

* Scelbi Basic ("Scelbal"). A small basic interpreter that you can use
  to interactively program.

* Forth. I wrote my own 8008 forth, based on Jonesforth.

* Star Trek. The usual Star Trek game from the 8-bit era.

* Pi-100-digits. Pi computed to 100 digits.

* Pi-1000-digits. Pi computed to 1000 digits. It gets the answer wrong
  but I haven't taken the time to figure out why yet.

* Hangman. Your usual word-guessing game.

All of the programs above are accessible from the monitor by using the
"S" command. For example S-1 to select Basic, S-2 to select Forth, etc.

## Terminal Emulation

The UART is interacted directly from the console. The emulator is a 
command-line tool. Launch the emulator from a Linux shell, and the UART
output and input use stdout and stdin respectively.

## Example

To run use the monitor rom, do the following:

```bash
$ make build
$ build/_output/cpusim -f roms/sbc-8251.rom
```

The output should look like this:

```bash
Serial Monitor for Intel 8008 H8 CPU Board V2.0
Original by Jim Loos
Modified by Scott Baker, https://www.smbaker.com/ for h8 project
Assembled on 07/12/2025 at 01:48:05 AM

B - Basic
C - Call subroutine
D - Dump RAM
E - Examine/Modify RAM
F - Fill RAM
H - Hex file download
G - Go to address
I - Input byte from port
O - Output byte to port
R - Raw binary file download
S - Switch bank and load rom

>>
```

## Makefile

* `make build`. This will build the emulator as a go binary.

* `make test`. This will execute unit tests. The unit tests were
  essential in making sure the emulator correctly executed the
  instructions. Unit test coverage is not 100%, but I tried to
  implement a reasonable representative of the instructions.

  The unit tests use an assembler called `asl` to assemble code
  that it will test. The unit test will try to cache the results
  and not reassemble, but if you modify the tests, you will need
  to have `asl` instealled to assemble the changes.

## Roms

* roms/sbc-8251.rom. Scott's single board computer, using Jim Loos's
  monitor, with Scelbal, Forth, etc.

* roms/map.rom. A short test of the mapper. It will print two strings,
  one from ROM and one from RAM.

* roms/uart.rom. A noninteractive UART test that prints some characters
  out the UART.

* roms/uartin.rom. An interactive UART test the reads a character from
  the console.

* roms/ops.rom. A handful of operations, one of my early sanity checks.
