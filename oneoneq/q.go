package oneoneq

import (
	"fmt"
	"github.com/fmstephe/queues/atomicint"
)

type Q struct {
	head atomicint.Value
	headCache int64
	tail atomicint.Value
	tailCache int64
	buffer []byte
	size int64
	chunk int64
	mask int64
}

func New(size int64, chunk int64) *Q {
	if size % chunk != 0 {
		panic(fmt.Sprintf("Size must be neatly divisible by chunk, (size) %d rem (chunk) %d = %d", size, chunk, size % chunk))
	}
	pow := int64(2)
	for i := 0; i < 64; i++ { // TODO this isn't a very clever way to determine if something is a power of two
		if pow == size {
			return &Q{buffer: make([]byte, size), size: size, chunk: chunk, mask: size-1}
		}
		pow *= 2
	}
	panic(fmt.Sprintf("Size must be a power of two, size = %d", size))
}

func (q *Q) StartWrite() []byte {
	tail := q.tail.NakedGet()
	wrapPoint := tail - q.size
	if q.headCache <= wrapPoint {
		q.headCache = q.head.Get()
		if q.headCache <= wrapPoint {
			return nil
		}
	}
	idx := tail & q.mask
	return q.buffer[idx : idx+q.chunk]
}

func (q *Q) FinishWrite() {
	q.tail.Add(q.chunk)
}

func (q *Q) StartRead() []byte {
	head := q.head.NakedGet()
	if head == q.tailCache {
		q.tailCache = q.tail.Get()
		if head == q.tailCache {
			return nil
		}
	}
	idx := head & q.mask
	return q.buffer[idx : idx+q.chunk]
}

func (q *Q) FinishRead() {
	q.tail.Add(q.chunk)
}
