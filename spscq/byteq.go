package spscq

import (
	"fmt"
	"github.com/fmstephe/fatomic"
)

type ByteQ struct {
	_1        fatomic.AtomicInt
	head      fatomic.AtomicInt
	headCache fatomic.AtomicInt
	tail      fatomic.AtomicInt
	tailCache fatomic.AtomicInt
	_2        fatomic.AtomicInt
	// Read only
	ringBuffer []byte
	size       int64
	mask       int64
	_3         fatomic.AtomicInt
}

func NewByteQ(size int64) *ByteQ {
	if !powerOfTwo(size) {
		panic(fmt.Sprintf("Size (%d) must be a power of two", size))
	}
	ringBuffer := fatomic.CacheProtectedBytes(int(size))
	q := &ByteQ{ringBuffer: ringBuffer, size: size, mask: size - 1}
	return q
}

func (q *ByteQ) Write(writeBuffer []byte) int64 {
	chunk := int64(len(writeBuffer))
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
	if nxt <= q.size {
		copy(q.ringBuffer[idx:nxt], writeBuffer)
	} else {
		mid := q.size - idx
		copy(q.ringBuffer[idx:], writeBuffer[:mid])
		copy(q.ringBuffer, writeBuffer[mid:])
	}
	q.tail.LazyAdd(chunk)
	return chunk
}

func (q *ByteQ) Read(readBuffer []byte) int64 {
	head := q.head.Value
	tail := q.tailCache.Value
	if head == tail {
		q.tailCache.Value = q.tail.ALoad()
		tail = q.tailCache.Value
		if head == tail {
			return 0
		}
	}
	chunk := min(tail-head, int64(len(readBuffer)))
	idx := head & q.mask
	nxt := idx + chunk
	if nxt <= q.size {
		copy(readBuffer, q.ringBuffer[idx:nxt])
	} else {
		mid := q.size - idx
		copy(readBuffer[:mid], q.ringBuffer[idx:])
		copy(readBuffer[mid:], q.ringBuffer)
	}
	q.head.LazyAdd(chunk)
	return chunk
}
