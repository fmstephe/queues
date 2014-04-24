package oneoneq

import (
	"unsafe"
	"fmt"
	"github.com/fmstephe/queues/atomicint"
)

type Q struct {
	_1 atomicint.CacheLine
	head atomicint.CacheLine
	headCache atomicint.CacheLine
	tail atomicint.CacheLine
	tailCache atomicint.CacheLine
	_2 atomicint.CacheLine
	// Read only
	buffer []byte
	size int64
	chunk int64
	mask int64
	_3 atomicint.CacheLine
}

func New(size int64, chunk int64) *Q {
	if size % chunk != 0 {
		panic(fmt.Sprintf("Size must be neatly divisible by chunk, (size) %d rem (chunk) %d = %d", size, chunk, size % chunk))
	}
	pow := int64(2)
	for i := 0; i < 64; i++ {
		if pow == size {
			q := &Q{buffer: make([]byte, size), size: size, chunk: chunk, mask: size-1}
			println(unsafe.Sizeof(*q))
			println(q)
			return q
		}
		pow *= 2
	}
	panic(fmt.Sprintf("Size must be a power of two, size = %d", size))
}

func (q *Q) Write(b []byte) int64 {
	chunk := q.chunk
	tail := q.tail.Value
	writeTo := tail + chunk
	headLimit := writeTo - q.size
	if headLimit > q.headCache.Value {
		q.headCache.Value = q.head.Get()
		if headLimit > q.headCache.Value {
			return 0
		}
	}
	idx := tail & q.mask
	nxt := idx + chunk
	copy(q.buffer[idx : nxt], b)
	q.tail.Add(chunk)
	return chunk
}

func (q *Q) Read(b []byte) int64 {
	chunk := q.chunk
	head := q.head.Value
	readTo := head + chunk
	if readTo > q.tailCache.Value {
		q.tailCache.Value = q.tail.Get()
		if readTo > q.tailCache.Value {
			return 0
		}
	}
	idx := head & q.mask
	nxt := idx + chunk
	copy(b, q.buffer[idx : nxt])
	q.head.Add(chunk)
	return chunk
}
