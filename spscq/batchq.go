package spscq

import (
	"fmt"
	"github.com/fmstephe/fatomic"
	"sync/atomic"
	"unsafe"
)

type BatchQ struct {
	_1         fatomic.Padded64Int64
	read       int64
	writeCache int64
	_2         fatomic.Padded64Int64
	write      int64
	readCache  int64
	_3         fatomic.Padded64Int64
	// Read only
	ringBuffer []unsafe.Pointer
	size       int64
	mask       int64
	_4         fatomic.Padded64Int64
}

func NewBatchQ(size int64) *BatchQ {
	if !powerOfTwo(size) {
		panic(fmt.Sprintf("Size must be a power of two, size = %d", size))
	}
	ringBuffer := fatomic.CacheProtectedPointers(int(size))
	q := &BatchQ{ringBuffer: ringBuffer, size: size, mask: size - 1}
	return q
}

func (q *BatchQ) WriteBuffer(bufferSize int64) []unsafe.Pointer {
	write := q.write
	idx := write & q.mask
	bufferSize = min(bufferSize, q.size-idx)
	writeTo := write + bufferSize
	readLimit := writeTo - q.size
	nxt := idx + bufferSize
	if readLimit > q.readCache {
		q.readCache = atomic.LoadInt64(&q.read)
		if readLimit > q.readCache {
			nxt = q.readCache & q.mask
		}
	}
	return q.ringBuffer[idx:nxt]
}

func (q *BatchQ) CommitWrite(writeSize int64) {
	atomic.StoreInt64(&q.write, q.write+writeSize)
}

func (q *BatchQ) ReadBuffer(bufferSize int64) []unsafe.Pointer {
	read := q.read
	idx := read & q.mask
	bufferSize = min(bufferSize, q.size-idx)
	readTo := read + bufferSize
	nxt := idx + bufferSize
	if readTo > q.writeCache {
		q.writeCache = atomic.LoadInt64(&q.write)
		if readTo > q.writeCache {
			nxt = q.writeCache & q.mask
		}
	}
	return q.ringBuffer[idx:nxt]
}

func (q *BatchQ) CommitRead(readSize int64) {
	atomic.StoreInt64(&q.read, q.read+readSize)
}
