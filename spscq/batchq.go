package spscq

import (
	"fmt"
	"github.com/fmstephe/fatomic"
	"sync/atomic"
	"unsafe"
)

type BatchQ struct {
	_1        fatomic.Padded64Int64
	head      int64
	tailCache int64
	_2        fatomic.Padded64Int64
	tail      int64
	headCache int64
	_3        fatomic.Padded64Int64
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
	tail := q.tail
	idx := tail & q.mask
	bufferSize = min(bufferSize, q.size-idx)
	writeTo := tail + bufferSize
	headLimit := writeTo - q.size
	nxt := idx + bufferSize
	if headLimit > q.headCache {
		q.headCache = atomic.LoadInt64(&q.head)
		if headLimit > q.headCache {
			nxt = q.headCache &q.mask
		}
	}
	return q.ringBuffer[idx:nxt]
}

func (q *BatchQ) CommitWrite(writeSize int64) {
	atomic.StoreInt64(&q.tail, q.tail+writeSize)
}

func (q *BatchQ) ReadBuffer(bufferSize int64) []unsafe.Pointer {
	head := q.head
	idx := head & q.mask
	bufferSize = min(bufferSize, q.size-idx)
	readTo := head + bufferSize
	nxt := idx + bufferSize
	if readTo > q.tailCache {
		q.tailCache = atomic.LoadInt64(&q.tail)
		if readTo > q.tailCache {
			nxt = q.tailCache & q.mask
		}
	}
	return q.ringBuffer[idx:nxt]
}

func (q *BatchQ) CommitRead(readSize int64) {
	atomic.StoreInt64(&q.head, q.head+readSize)
}
