//go:build darwin
// +build darwin

package rawmode

/* Raw input for macOS (darwin) platform.
 *
 * The canonical solution is to use term.MakeRaw(int(os.Stdin.Fd())). However, this affects both
 * standard input and standard output. For example, it will cause `\n` to output linefeeds but
 * not carriage returns.
 *
 * Instead we use this solution for macOS, which only affects input, similar to the Linux
 * implementation but using golang.org/x/sys/unix for macOS-specific constants.
 */

import (
	"os"

	"golang.org/x/sys/unix"
)

var oldState *unix.Termios
var called int // in case there are nested calls to EnableRawMode

// enableRawMode switches the terminal to raw mode and returns the original state
func EnableRawMode() error {
	if called > 0 {
		called++
		return nil
	}

	fd := int(os.Stdin.Fd())

	// Get the current terminal attributes
	termios, err := unix.IoctlGetTermios(fd, unix.TIOCGETA)
	if err != nil {
		return err
	}

	// Save the original state
	oldState = termios

	// Modify the attributes to enable raw mode
	newState := *termios
	// Disable canonical mode (ICANON) and echo (ECHO)
	newState.Lflag &^= unix.ICANON | unix.ECHO

	// Set the new terminal attributes
	err = unix.IoctlSetTermios(fd, unix.TIOCSETA, &newState)
	if err != nil {
		return err
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

	if oldState == nil {
		return nil
	}

	fd := int(os.Stdin.Fd())
	err := unix.IoctlSetTermios(fd, unix.TIOCSETA, oldState)
	if err != nil {
		return err
	}
	return nil
}
