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
	for i := 0; i < 64; i++ {
		if pow == size {
			return &Q{buffer: make([]byte, size), size: size, chunk: chunk, mask: size-1}
		}
		pow *= 2
	}
	panic(fmt.Sprintf("Size must be a power of two, size = %d", size))
}

func (q *Q) Write(b []byte) int64 {
	tail := q.tail.NakedGet()
	writeTo := tail + q.chunk
	headLimit := writeTo - q.size
	if headLimit > q.headCache {
		q.headCache = q.head.Get()
		if headLimit > q.headCache {
			return 0
		}
	}
	idx := tail & q.mask
	copy(q.buffer[idx : idx+q.chunk], b)
	q.tail.Add(q.chunk)
	return q.chunk
}

func (q *Q) Read(b []byte) int64 {
	bl := int64(len(b))
	head := q.head.NakedGet()
	readTo := head + bl
	if readTo > q.tailCache {
		q.tailCache = q.tail.Get()
		if readTo > q.tailCache {
			return 0
		}
	}
	idx := head & q.mask
	nxt := idx + bl
	if nxt > q.size {
		nxt = q.size
	}
	copy(b, q.buffer[idx : nxt])
	read := nxt - idx
	q.head.Add(read)
	return read
}
