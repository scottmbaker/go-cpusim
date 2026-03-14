package cpusim

import "io"

// ChannelSerial implements SerialIO using Go channels, for programmatic interaction
// with serial devices without a terminal.
//
// In may be closed by the producer to signal EOF; ReadByte will return io.EOF.
// Out must not be closed by consumers; the device is the sender per Go convention.
type ChannelSerial struct {
	In  chan byte // data flowing into the device (simulated keyboard input)
	Out chan byte // data flowing out of the device (simulated display output)
}

func NewChannelSerial() *ChannelSerial {
	return &ChannelSerial{
		In:  make(chan byte, 256),
		Out: make(chan byte, 256),
	}
}

func (c *ChannelSerial) ReadByte() (byte, error) {
	b, ok := <-c.In
	if !ok {
		return 0, io.EOF
	}
	return b, nil
}

func (c *ChannelSerial) WriteByte(b byte) error {
	c.Out <- b
	return nil
}

func (c *ChannelSerial) Start() {
}

func (c *ChannelSerial) RestoreTerminal() {
}
