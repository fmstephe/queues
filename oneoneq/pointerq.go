package oneoneq

import (
	"fmt"
	"github.com/fmstephe/fatomic"
	"unsafe"
)

type PointerQ struct {
	_1        fatomic.AtomicInt
	head      fatomic.AtomicInt
	headCache fatomic.AtomicInt
	tail      fatomic.AtomicInt
	tailCache fatomic.AtomicInt
	_2        fatomic.AtomicInt
	// Read only
	ringBuffer []unsafe.Pointer
	size       int64
	mask       int64
	_3         fatomic.AtomicInt
}

func NewPointerQ(size int64) *PointerQ {
	pow := int64(2)
	for i := 0; i < 64; i++ {
		if pow == size {
			ringBuffer := fatomic.CacheProtectedPointers(int(size))
			q := &PointerQ{ringBuffer: ringBuffer, size: size, mask: size - 1}
			return q
		}
		pow *= 2
	}
	panic(fmt.Sprintf("Size must be a power of two, size = %d", size))
}

func (q *PointerQ) Write(val unsafe.Pointer) bool {
	tail := q.tail.Value
	headLimit := tail - q.size
	if headLimit >= q.headCache.Value {
		q.headCache.Value = q.head.ALoad()
		if headLimit >= q.headCache.Value {
			return false
		}
	}
	idx := tail & q.mask
	q.ringBuffer[idx] = val
	q.tail.LazyAdd(1)
	return true
}

func (q *PointerQ) Read() unsafe.Pointer {
	head := q.head.Value
	if head == q.tailCache.Value {
		q.tailCache.Value = q.tail.ALoad()
		if head == q.tailCache.Value {
			return nil
		}
	}
	idx := head & q.mask
	val := q.ringBuffer[idx]
	q.head.LazyAdd(1)
	return val
}
