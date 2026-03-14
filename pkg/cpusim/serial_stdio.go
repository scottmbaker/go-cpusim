package cpusim

import (
	"fmt"
	"os"

	"github.com/scottmbaker/gocpusim/pkg/rawmode"
)

// StdioSerial implements SerialIO using os.Stdin and os.Stdout.
type StdioSerial struct {
	rawMode    bool
	rawEnabled bool
}

func NewStdioSerial(rawMode bool) *StdioSerial {
	return &StdioSerial{rawMode: rawMode}
}

func (s *StdioSerial) ReadByte() (byte, error) {
	var buf [1]byte
	_, err := os.Stdin.Read(buf[:])
	if err != nil {
		return 0, err
	}
	return buf[0], nil
}

func (s *StdioSerial) WriteByte(b byte) error {
	var buf [1]byte
	buf[0] = b
	_, err := os.Stdout.Write(buf[:])
	return err
}

func (s *StdioSerial) Start() {
	if s.rawMode {
		err := rawmode.EnableRawMode()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error setting terminal to raw mode: %v\n", err)
		} else {
			s.rawEnabled = true
		}
	}
}

func (s *StdioSerial) RestoreTerminal() {
	if !s.rawEnabled {
		return
	}
	err := rawmode.DisableRawMode()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error restoring terminal mode: %v\n", err)
	}
	s.rawEnabled = false
}
