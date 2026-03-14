package cpusim

import (
	"fmt"
	"os"

	"github.com/scottmbaker/gocpusim/pkg/rawmode"
)

// StdioSerial implements SerialIO using os.Stdin and os.Stdout.
type StdioSerial struct {
	rawMode bool
}

func NewStdioSerial(rawMode bool) *StdioSerial {
	return &StdioSerial{rawMode: rawMode}
}

func (s *StdioSerial) ReadByte() (byte, error) {
	input := make([]byte, 1)
	_, err := os.Stdin.Read(input)
	if err != nil {
		return 0, err
	}
	return input[0], nil
}

func (s *StdioSerial) WriteByte(b byte) error {
	_, err := os.Stdout.Write([]byte{b})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to stdout: %v", err)
	}
	return err
}

func (s *StdioSerial) Start() {
	if s.rawMode {
		err := rawmode.EnableRawMode()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error setting terminal to raw mode: %v\n", err)
		}
	}
}

func (s *StdioSerial) RestoreTerminal() {
	err := rawmode.DisableRawMode()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error restoring terminal mode: %v\n", err)
	}
}
