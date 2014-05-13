package spscq

import (
	"fmt"
	"github.com/fmstephe/fatomic"
	"sync/atomic"
	"unsafe"
)

type PointerQ struct {
	_1        fatomic.Padded64Int64
	head      fatomic.Padded64Int64
	headCache fatomic.Padded64Int64
	tail      fatomic.Padded64Int64
	tailCache fatomic.Padded64Int64
	_2        fatomic.Padded64Int64
	// Read only
	ringBuffer []unsafe.Pointer
	size       int64
	mask       int64
	_3         fatomic.Padded64Int64
}

func NewPointerQ(size int64) *PointerQ {
	if !powerOfTwo(size) {
		panic(fmt.Sprintf("Size must be a power of two, size = %d", size))
	}
	ringBuffer := fatomic.CacheProtectedPointers(int(size))
	q := &PointerQ{ringBuffer: ringBuffer, size: size, mask: size - 1}
	return q
}

func (q *PointerQ) Write(val unsafe.Pointer) bool {
	tail := q.tail.Value
	headLimit := tail - q.size
	if headLimit >= q.headCache.Value {
		q.headCache.Value = atomic.LoadInt64(&q.head.Value)
		if headLimit >= q.headCache.Value {
			return false
		}
	}
	idx := tail & q.mask
	q.ringBuffer[idx] = val
	fatomic.LazyStore(&q.tail.Value, q.tail.Value+1)
	return true
}

func (q *PointerQ) Read() unsafe.Pointer {
	head := q.head.Value
	if head == q.tailCache.Value {
		q.tailCache.Value = atomic.LoadInt64(&q.tail.Value)
		if head == q.tailCache.Value {
			return nil
		}
	}
	idx := head & q.mask
	val := q.ringBuffer[idx]
	fatomic.LazyStore(&q.head.Value, q.head.Value+1)
	return val
}
