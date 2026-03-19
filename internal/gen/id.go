package gen

import (
	"sync/atomic"
	"time"
)

// Simple snowflake-ish ID generator.
// Base = unix microseconds at startup (~1.7e15), well under JS
// Number.MAX_SAFE_INTEGER (9e15). Each NextID() increments atomically.
// Unique across restarts since the base advances with wall clock.
var idSeq atomic.Int64

func init() {
	idSeq.Store(time.Now().UnixMicro())
}

func NextID() int64 {
	return idSeq.Add(1)
}
