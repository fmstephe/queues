package spscq

import (
	"fmt"
	"github.com/fmstephe/fatomic"
	"sync/atomic"
)

type LLChunkQ struct {
	_1         fatomic.Padded64Int64
	read       fatomic.Padded64Int64
	readCache  fatomic.Padded64Int64
	write      fatomic.Padded64Int64
	writeCache fatomic.Padded64Int64
	_2         fatomic.Padded64Int64
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
	write := q.write.Value
	writeTo := write + chunk
	readLimit := writeTo - q.size
	if readLimit > q.readCache.Value {
		q.readCache.Value = atomic.LoadInt64(&q.read.Value)
		if readLimit > q.readCache.Value {
			return nil
		}
	}
	idx := write & q.mask
	nxt := idx + chunk
	return q.ringBuffer[idx:nxt]
}

func (q *LLChunkQ) CommitWrite() {
	fatomic.LazyStore(&q.write.Value, q.write.Value+q.chunk) // q.write.LazyAdd(q.chunk)
}

func (q *LLChunkQ) ReadBuffer() []byte {
	chunk := q.chunk
	read := q.read.Value
	readTo := read + chunk
	if readTo > q.writeCache.Value {
		q.writeCache.Value = atomic.LoadInt64(&q.write.Value)
		if readTo > q.writeCache.Value {
			return nil
		}
	}
	idx := read & q.mask
	nxt := idx + chunk
	return q.ringBuffer[idx:nxt]
}

func (q *LLChunkQ) CommitRead() {
	fatomic.LazyStore(&q.read.Value, q.read.Value+q.chunk) // q.read.LazyAdd(q.chunk)
}
