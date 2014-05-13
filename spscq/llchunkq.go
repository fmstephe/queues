package spscq

import (
	"fmt"
	"github.com/fmstephe/fatomic"
	"sync/atomic"
)

type LLChunkQ struct {
	_1        fatomic.Padded64Int64
	head      fatomic.Padded64Int64
	headCache fatomic.Padded64Int64
	tail      fatomic.Padded64Int64
	tailCache fatomic.Padded64Int64
	_2        fatomic.Padded64Int64
	// Read only
	ringBuffer []byte
	size       int64
	chunk      int64
	mask       int64
	_3         fatomic.Padded64Int64
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
		q.headCache.Value = atomic.LoadInt64(&q.head.Value)
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
		q.tailCache.Value = atomic.LoadInt64(&q.tail.Value)
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
