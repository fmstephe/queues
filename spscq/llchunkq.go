package spscq

import (
	"fmt"
	"github.com/fmstephe/flib/fsync/fatomic"
	"github.com/fmstephe/flib/fsync/padded"
	"sync/atomic"
)

type LLChunkQ struct {
	_1         padded.Int64
	read       padded.Int64
	readCache  padded.Int64
	write      padded.Int64
	writeCache padded.Int64
	_2         padded.Int64
	// Read only
	ringBuffer []byte
	size       int64
	chunk      int64
	mask       int64
	_3         padded.Int64
}

func NewLLChunkQ(size int64, chunk int64) *LLChunkQ {
	if !powerOfTwo(size) {
		panic(fmt.Sprintf("Size must be a power of two, size = %d", size))
	}
	if size%chunk != 0 {
		panic(fmt.Sprintf("Size must be neatly divisible by chunk, (size) %d rem (chunk) %d = %d", size, chunk, size%chunk))
	}
	ringBuffer := padded.ByteSlice(int(size))
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
	fatomic.LazyStore(&q.write.Value, q.write.Value+q.chunk)
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
	fatomic.LazyStore(&q.read.Value, q.read.Value+q.chunk)
}
