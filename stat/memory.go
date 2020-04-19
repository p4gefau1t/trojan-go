package stat

import "sync/atomic"

type MemoryTrafficMeter struct {
	TrafficMeter
	sent uint64
	recv uint64
}

func (m *MemoryTrafficMeter) Count(passwordHash string, sent, recv uint64) {
	atomic.AddUint64(&m.sent, uint64(sent))
	atomic.AddUint64(&m.recv, uint64(recv))
}

func (m *MemoryTrafficMeter) Query(passwordHash string) (uint64, uint64) {
	return m.sent, m.recv
}
