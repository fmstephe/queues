package spscq

import (
	"fmt"
	"github.com/fmstephe/fatomic"
)

type ChunkQ struct {
	_1        fatomic.AtomicInt
	head      fatomic.AtomicInt
	headCache fatomic.AtomicInt
	tail      fatomic.AtomicInt
	tailCache fatomic.AtomicInt
	_2        fatomic.AtomicInt
	// Read only
	ringBuffer  []byte
	readBuffer  []byte
	writeBuffer []byte
	size        int64
	chunk       int64
	mask        int64
	_3          fatomic.AtomicInt
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

func (q *ChunkQ) Write() int64 {
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
	copy(q.ringBuffer[idx:nxt], q.writeBuffer)
	fatomic.LazyStore(&q.tail.Value, tail+chunk) // q.tail.LazyAdd(chunk)
	return chunk
}

func (q *ChunkQ) WriteBuffer() []byte {
	return q.writeBuffer
}

func (q *ChunkQ) Read() int64 {
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
	copy(q.readBuffer, q.ringBuffer[idx:nxt])
	fatomic.LazyStore(&q.head.Value, head+chunk) // q.head.LazyAdd(chunk)
	return chunk
}
