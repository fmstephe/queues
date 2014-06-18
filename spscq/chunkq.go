package spscq

import (
	"fmt"
	"github.com/fmstephe/fatomic"
	"sync/atomic"
)

type ChunkQ struct {
	_1         fatomic.Padded64Int64
	read       fatomic.Padded64Int64
	readCache  fatomic.Padded64Int64
	write      fatomic.Padded64Int64
	writeCache fatomic.Padded64Int64
	_2         fatomic.Padded64Int64
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
	write := q.write.Value
	writeTo := write + chunk
	readLimit := writeTo - q.size
	if readLimit > q.readCache.Value {
		q.readCache.Value = atomic.LoadInt64(&q.read.Value)
		if readLimit > q.readCache.Value {
			return false
		}
	}
	idx := write & q.mask
	nxt := idx + chunk
	copy(q.ringBuffer[idx:nxt], q.writeBuffer)
	fatomic.LazyStore(&q.write.Value, write+chunk) // q.write.LazyAdd(chunk)
	return true
}

func (q *ChunkQ) WriteBuffer() []byte {
	return q.writeBuffer
}

func (q *ChunkQ) Read() bool {
	chunk := q.chunk
	read := q.read.Value
	readTo := read + chunk
	if readTo > q.writeCache.Value {
		q.writeCache.Value = atomic.LoadInt64(&q.write.Value)
		if readTo > q.writeCache.Value {
			return false
		}
	}
	idx := read & q.mask
	nxt := idx + chunk
	copy(q.readBuffer, q.ringBuffer[idx:nxt])
	fatomic.LazyStore(&q.read.Value, read+chunk) // q.read.LazyAdd(chunk)
	return true
}
