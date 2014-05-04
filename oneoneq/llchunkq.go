package oneoneq

import (
	"fmt"
	"github.com/fmstephe/fatomic"
)

type LLChunkQ struct {
	_1        fatomic.AtomicInt
	head      fatomic.AtomicInt
	headCache fatomic.AtomicInt
	tail      fatomic.AtomicInt
	tailCache fatomic.AtomicInt
	_2        fatomic.AtomicInt
	// Read only
	ringBuffer  []byte
	size        int64
	chunk       int64
	mask        int64
	_3          fatomic.AtomicInt
}

func NewLLChunkQ(size int64, chunk int64) *LLChunkQ {
	if !powerOfTwo(size) {
		panic(fmt.Sprintf("Size must be a power of two, size = %d", size))
	}
	if size%chunk != 0 {
		panic(fmt.Sprintf("Size must be neatly divisible by chunk, (size) %d rem (chunk) %d = %d", size, chunk, size%chunk))
	}
	ringBuffer := fatomic.CacheProtectedBytes(int(size))
	q := &LLChunkQ{ringBuffer: ringBuffer, size: size, chunk: chunk, mask: size - 1}
	return q
}

func (q *LLChunkQ) WriteBuffer() []byte {
	chunk := q.chunk
	tail := q.tail.Value
	writeTo := tail + chunk
	headLimit := writeTo - q.size
	if headLimit > q.headCache.Value {
		q.headCache.Value = q.head.ALoad()
		if headLimit > q.headCache.Value {
			return nil
		}
	}
	idx := tail & q.mask
	nxt := idx + chunk
	return q.ringBuffer[idx:nxt]
}

func (q *LLChunkQ) CommitWrite() {
	fatomic.LazyStore(&q.tail.Value, q.tail.Value+q.chunk) // q.tail.LazyAdd(q.chunk)
}

func (q *LLChunkQ) ReadBuffer() []byte {
	chunk := q.chunk
	head := q.head.Value
	readTo := head + chunk
	if readTo > q.tailCache.Value {
		q.tailCache.Value = q.tail.ALoad()
		if readTo > q.tailCache.Value {
			return nil
		}
	}
	idx := head & q.mask
	nxt := idx + chunk
	return q.ringBuffer[idx:nxt]
}

func (q *LLChunkQ) CommitRead() {
	fatomic.LazyStore(&q.head.Value, q.head.Value+q.chunk) // q.head.LazyAdd(q.chunk)
}
