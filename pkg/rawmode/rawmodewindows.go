//go:build windows
// +build windows

package rawmode

import (
	"golang.org/x/term"
	"os"
)

var oldState *term.State
var called int // in case there are nested calls to EnableRawMode

// enableRawMode switches the terminal to raw mode and returns the original state
func EnableRawMode() error {
	if called > 0 {
		called++
		return nil
	}

	var err error
	oldState, err = term.MakeRaw(int(os.Stdin.Fd()))
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

	term.Restore(int(os.Stdin.Fd()), oldState)
	return nil
}
