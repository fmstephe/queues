package oneoneq

import (
	"unsafe"
	"fmt"
	"github.com/fmstephe/fatomic"
)

type Q struct {
	_1 fatomic.AtomicInt
	head fatomic.AtomicInt
	headCache fatomic.AtomicInt
	tail fatomic.AtomicInt
	tailCache fatomic.AtomicInt
	_2 fatomic.AtomicInt
	// Read only
	buffer []byte
	size int64
	chunk int64
	mask int64
	_3 fatomic.AtomicInt
}

func New(size int64, chunk int64) *Q {
	if size % chunk != 0 {
		panic(fmt.Sprintf("Size must be neatly divisible by chunk, (size) %d rem (chunk) %d = %d", size, chunk, size % chunk))
	}
	pow := int64(2)
	for i := 0; i < 64; i++ {
		if pow == size {
			buffer := fatomic.CacheProtectedBytes(int(size))
			q := &Q{buffer: buffer, size: size, chunk: chunk, mask: size-1}
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
		q.headCache.Value = q.head.ALoad()
		if headLimit > q.headCache.Value {
			return 0
		}
	}
	idx := tail & q.mask
	nxt := idx + chunk
	copy(q.buffer[idx : nxt], b)
	q.tail.LazyAdd(chunk)
	return chunk
}

func (q *Q) Read(b []byte) int64 {
	chunk := q.chunk
	head := q.head.Value
	readTo := head + chunk
	if readTo > q.tailCache.Value {
		q.tailCache.Value = q.tail.ALoad()
		if readTo > q.tailCache.Value {
			return 0
		}
	}
	idx := head & q.mask
	nxt := idx + chunk
	copy(b, q.buffer[idx : nxt])
	q.head.LazyAdd(chunk)
	return chunk
}
