package cpusim

import "time"

const throttleBatchSize = 1000

// Throttle limits CPU execution to a target instructions-per-second rate.
// An ips of zero means no throttling (full speed). To avoid per-instruction
// overhead, timing is checked every throttleBatchSize instructions.
type Throttle struct {
	ips        int64
	count      int64
	batchStart time.Time
}

func NewThrottle(ips int64) *Throttle {
	return &Throttle{
		ips:        ips,
		batchStart: time.Now(),
	}
}

// Tick is called once per executed instruction. Every throttleBatchSize
// instructions, it compares elapsed wall time against the expected duration
// and sleeps the difference if the CPU is running ahead of schedule.
func (t *Throttle) Tick() {
	if t.ips <= 0 {
		return
	}

	t.count++
	if t.count < throttleBatchSize {
		return
	}

	expected := time.Duration(t.count) * time.Second / time.Duration(t.ips)
	elapsed := time.Since(t.batchStart)

	if elapsed < expected {
		time.Sleep(expected - elapsed)
	}

	t.count = 0
	t.batchStart = time.Now()
}
