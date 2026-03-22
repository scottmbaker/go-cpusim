package cpusim

import (
	"fmt"
	"io"
	"os"

	"github.com/scottmbaker/gocpusim/pkg/rawmode"
)

// FileSerial implements SerialIO by reading from a file, optionally falling
// through to stdin when the file is exhausted.
type FileSerial struct {
	file       io.Reader
	stdin      bool // true if we should fall through to stdin after file EOF
	fileEOF    bool // true once file is exhausted
	rawEnabled bool
}

// NewFileSerial creates a FileSerial that reads from the named file.
// If exitOnEof is true, ReadByte returns io.EOF when the file is exhausted.
// If exitOnEof is false, input falls through to os.Stdin after the file ends,
// and raw mode is enabled at that point.
func NewFileSerial(filename string, exitOnEof bool) (*FileSerial, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	return &FileSerial{
		file:  f,
		stdin: !exitOnEof,
	}, nil
}

func (s *FileSerial) ReadByte() (byte, error) {
	var buf [1]byte

	if !s.fileEOF {
		_, err := s.file.Read(buf[:])
		if err == nil {
			return buf[0], nil
		}
		if err != io.EOF {
			return 0, err
		}
		// File exhausted
		s.fileEOF = true
		if !s.stdin {
			return 0, io.EOF
		}
		// Transition to stdin — enable raw mode now
		err = rawmode.EnableRawMode()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error setting terminal to raw mode: %v\n", err)
		} else {
			s.rawEnabled = true
		}
	}

	_, err := os.Stdin.Read(buf[:])
	if err != nil {
		return 0, err
	}
	return buf[0], nil
}

func (s *FileSerial) WriteByte(b byte) error {
	var buf [1]byte
	buf[0] = b
	_, err := os.Stdout.Write(buf[:])
	return err
}

func (s *FileSerial) Start() {}

func (s *FileSerial) RestoreTerminal() {
	if !s.rawEnabled {
		return
	}
	err := rawmode.DisableRawMode()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error restoring terminal mode: %v\n", err)
	}
	s.rawEnabled = false
}
