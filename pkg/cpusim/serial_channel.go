package cpusim

import (
	"io"
	"sync"
)

// ChannelSerial implements SerialIO using Go channels, for programmatic interaction
// with serial devices without a terminal.
//
// In may be closed by the producer to signal EOF; ReadByte will return io.EOF.
// Out must never be closed; call Close() to shut the transport down. Close()
// unblocks any goroutine inside ReadByte or WriteByte, which is required to
// safely stop a running simulation: the CPU may be blocked mid-instruction in
// WriteByte when the consumer goes away.
type ChannelSerial struct {
	In  chan byte // data flowing into the device (simulated keyboard input)
	Out chan byte // data flowing out of the device (simulated display output)

	done      chan struct{}
	closeOnce sync.Once
}

func NewChannelSerial() *ChannelSerial {
	return &ChannelSerial{
		In:   make(chan byte, 256),
		Out:  make(chan byte, 256),
		done: make(chan struct{}),
	}
}

// Close shuts down the transport. Pending and future ReadByte calls return
// io.EOF; pending and future WriteByte calls return io.ErrClosedPipe. Safe to
// call multiple times and from any goroutine. Neither In nor Out is closed.
func (c *ChannelSerial) Close() {
	c.closeOnce.Do(func() { close(c.done) })
}

// Done is closed when the transport has been shut down via Close. Consumers
// draining Out should select on it to avoid blocking forever after shutdown.
func (c *ChannelSerial) Done() <-chan struct{} {
	return c.done
}

func (c *ChannelSerial) ReadByte() (byte, error) {
	select {
	case b, ok := <-c.In:
		if !ok {
			return 0, io.EOF
		}
		return b, nil
	case <-c.done:
		return 0, io.EOF
	}
}

// WriteByte sends a byte to the output channel. This deliberately blocks when
// the buffer is full, providing natural backpressure analogous to a real UART's
// TDRE (Transmit Data Register Empty) flow control. Dropping bytes would corrupt
// terminal output (e.g., partial escape sequences), which is worse than briefly
// stalling the simulated CPU until the consumer catches up. Close() unblocks a
// blocked WriteByte, which then returns io.ErrClosedPipe.
func (c *ChannelSerial) WriteByte(b byte) error {
	select {
	case c.Out <- b:
		return nil
	case <-c.done:
		return io.ErrClosedPipe
	}
}

func (c *ChannelSerial) Start() {
}

func (c *ChannelSerial) RestoreTerminal() {
}
