package ratelimiter

import (
	"sync/atomic"
	"time"
)

type Bucket struct {
	tokens     atomic.Uint32
	capacity   uint32
	refillRate uint32
	lastRefill atomic.Int64
}

func NewBucket(initialTokens, capacity, refillRate uint32) *Bucket {
	b := &Bucket{
		capacity:   capacity,
		refillRate: refillRate,
	}

	b.tokens.Store(initialTokens)
	b.lastRefill.Store(time.Now().UnixNano())
	return b
}

func (b *Bucket) TryTake() bool {
	for {
		current := b.tokens.Load()

		if current > 0 {
			if b.tokens.CompareAndSwap(current, current-1) {
				return true
			}
		}
		if current == 0 {
			return false
		}
	}
}

func (b *Bucket) RefillIfNeeded(now time.Time) {
	lastRefill := time.Unix(0, b.lastRefill.Load())
	elapsed := now.Sub(lastRefill)

	tokensToAdd := uint32(elapsed.Seconds()) * b.refillRate
	if tokensToAdd == 0 {
		return
	}

	current := b.tokens.Load()
	newTokens := min(current+tokensToAdd, b.capacity)

	b.tokens.Store(newTokens)
	b.lastRefill.Store(now.UnixNano())
}
