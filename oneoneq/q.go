package oneoneq

import (
	"github.com/fmstephe/queues/atomicint"
)

type Q struct {
	head atomicint.Value
	headCache int64
	tail atomicint.Value
	tailCache int64
	buffer []int64
	size int64
	mask int64
}

func New(minSize int64) *Q {
	size := int64(2)
	for size < minSize {
		size *= 2
	}
	return &Q{buffer: make([]int64, size), size: size, mask: size-1}
}

func (q *Q) Enqueue(item int64) bool {
	tail := q.tail.NakedGet()
	wrapPoint := tail - q.size
	if q.headCache <= wrapPoint {
		q.headCache = q.head.Get()
		if q.headCache <= wrapPoint {
			return false
		}
	}
	q.buffer[tail & q.mask] = item
	q.tail.NakedSet(tail+1)
	return true
}

func (q *Q) Dequeue() int64 {
	head := q.head.NakedGet()
	if head == q.tailCache {
		q.tailCache = q.tail.Get()
		if head == q.tailCache {
			return -1
		}
	}
	idx := head & q.mask
	item := q.buffer[idx]
	q.buffer[idx] = -1
	q.head.NakedSet(head+1)
	return item
}
