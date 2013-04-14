package oneoneq

import (
	"github.com/fmstephe/queues/fenced"
)

type Q struct {
	head fenced.Value
	tail fenced.Value
	//before [8]int64
	//head int64
	//mid [8]int64
	//tail int64
	//after [8]int64
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
	head := q.head.Get()
	tail := q.tail.Get()
	if tail - head == q.size {
		return false
	}
	q.buffer[tail & q.mask] = item
	q.tail.SetOrdered(tail+1)
	return true
}

func (q *Q) Dequeue() int64 {
	head := q.head.Get()
	tail := q.tail.Get()
	if head == tail {
		return -1
	}
	idx := head & q.mask
	item := q.buffer[idx]
	q.buffer[idx] = -1
	q.head.SetOrdered(head+1)
	return item
}
