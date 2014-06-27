package spscq

import (
	"fmt"
	"github.com/fmstephe/flib/fsync/padded"
	"sync/atomic"
)

type ExtQ struct {
	_1         padded.Int64
	read       int64
	writeCache int64
	_2         padded.Int64
	write      int64
	readCache  int64
	_3         padded.Int64
	// Read only
	size int64
	mask int64
	_4   padded.Int64
}

func NewExtQ(size int64) *ExtQ {
	if !powerOfTwo(size) {
		panic(fmt.Sprintf("Size must be a power of two, size = %d", size))
	}
	q := &ExtQ{size: size, mask: size - 1}
	return q
}

func (q *ExtQ) WriteBuffer(bufferSize int64) (low, high int64) {
	low = q.write
	high = low + bufferSize
	writeLimit := q.readCache + q.size
	if high > writeLimit {
		q.readCache = atomic.LoadInt64(&q.read)
		writeLimit = q.readCache + q.size
		if high > writeLimit {
			high = writeLimit
		}
	}
	return low, high
}

func (q *ExtQ) CommitWriteBuffer(writeSize int64) {
	atomic.AddInt64(&q.write, writeSize)
}

func (q *ExtQ) ReadBuffer(bufferSize int64) (low, high int64) {
	low = q.read
	high = low + bufferSize
	if high > q.writeCache {
		q.writeCache = atomic.LoadInt64(&q.write)
		if high > q.writeCache {
			high = q.writeCache
		}
	}
	return low, high
}

func (q *ExtQ) CommitReadBuffer(readSize int64) {
	atomic.AddInt64(&q.read, readSize)
}
