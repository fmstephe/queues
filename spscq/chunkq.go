package spscq

import (
	"fmt"
	"github.com/fmstephe/fatomic"
	"sync/atomic"
)

type ChunkQ struct {
	_1        fatomic.Padded64Int64
	head      fatomic.Padded64Int64
	headCache fatomic.Padded64Int64
	tail      fatomic.Padded64Int64
	tailCache fatomic.Padded64Int64
	_2        fatomic.Padded64Int64
	// Read only
	ringBuffer  []byte
	readBuffer  []byte
	writeBuffer []byte
	size        int64
	chunk       int64
	mask        int64
	_3          fatomic.Padded64Int64
}

func NewChunkQ(size int64, chunk int64) *ChunkQ {
	if !powerOfTwo(size) {
		panic(fmt.Sprintf("Size must be a power of two, size = %d", size))
	}
	if size%chunk != 0 {
		panic(fmt.Sprintf("Size must be neatly divisible by chunk, (size) %d rem (chunk) %d = %d", size, chunk, size%chunk))
	}
	ringBuffer := fatomic.CacheProtectedBytes(int(size))
	readBuffer := fatomic.CacheProtectedBytes(int(chunk))
	writeBuffer := fatomic.CacheProtectedBytes(int(chunk))
	q := &ChunkQ{ringBuffer: ringBuffer, readBuffer: readBuffer, writeBuffer: writeBuffer, size: size, chunk: chunk, mask: size - 1}
	return q
}

func (q *ChunkQ) ReadBuffer() []byte {
	return q.readBuffer
}

func (q *ChunkQ) Write() bool {
	chunk := q.chunk
	tail := q.tail.Value
	writeTo := tail + chunk
	headLimit := writeTo - q.size
	if headLimit > q.headCache.Value {
		q.headCache.Value = atomic.LoadInt64(&q.head.Value)
		if headLimit > q.headCache.Value {
			return false
		}
	}
	idx := tail & q.mask
	nxt := idx + chunk
	copy(q.ringBuffer[idx:nxt], q.writeBuffer)
	fatomic.LazyStore(&q.tail.Value, tail+chunk) // q.tail.LazyAdd(chunk)
	return true
}

func (q *ChunkQ) WriteBuffer() []byte {
	return q.writeBuffer
}

func (q *ChunkQ) Read() bool {
	chunk := q.chunk
	head := q.head.Value
	readTo := head + chunk
	if readTo > q.tailCache.Value {
		q.tailCache.Value = atomic.LoadInt64(&q.tail.Value)
		if readTo > q.tailCache.Value {
			return false
		}
	}
	idx := head & q.mask
	nxt := idx + chunk
	copy(q.readBuffer, q.ringBuffer[idx:nxt])
	fatomic.LazyStore(&q.head.Value, head+chunk) // q.head.LazyAdd(chunk)
	return true
}
