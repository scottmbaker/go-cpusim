//go:build linux
// +build linux

package rawmode

/* Raw input for Linux platform.
 *
 * The canonical solution is to use term.MakeRaw(int(os.Stdin.Fd())). However, this affects both
 * standard input and standard output. For example, it will cause `\n` to output linefeeds but
 * not carriage returns.
 *
 * Instead I used this solution for Linux, which only affects input:
 *
 * https://mzunino.com.uy/til/2025/03/building-a-terminal-raw-mode-input-reader-in-go/
 */

import (
	"syscall"
	"unsafe"
)

// termios holds the terminal attributes
type Termios struct {
	Iflag  uint32
	Oflag  uint32
	Cflag  uint32
	Lflag  uint32
	Cc     [20]byte
	Ispeed uint32
	Ospeed uint32
}

var oldState Termios
var called int // in case there are nested calls to EnableRawMode

// enableRawMode switches the terminal to raw mode and returns the original state
func EnableRawMode() error {
	if called > 0 {
		called++
		return nil
	}

	fd := int(syscall.Stdin)

	// Get the current terminal attributes
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(fd), uintptr(syscall.TCGETS),
		uintptr(unsafe.Pointer(&oldState)))
	if errno != 0 {
		return errno
	}

	// Modify the attributes to enable raw mode
	newState := oldState
	// Disable canonical mode (ICANON) and echo (ECHO)
	newState.Lflag &^= syscall.ICANON | syscall.ECHO

	// Set the new terminal attributes
	_, _, errno = syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(fd), uintptr(syscall.TCSETS),
		uintptr(unsafe.Pointer(&newState)))
	if errno != 0 {
		return errno
	}

	called++
	return nil
}

// disableRawMode restores the terminal to its original state
func DisableRawMode() error {
	called--
	if called != 0 {
		return nil
	}

	fd := int(syscall.Stdin)
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(fd), uintptr(syscall.TCSETS),
		uintptr(unsafe.Pointer(&oldState)))
	if errno != 0 {
		return errno
	}
	return nil
}
