package spscq

import (
	"fmt"
	"github.com/fmstephe/fatomic"
	"sync/atomic"
	"unsafe"
)

type PointerQ struct {
	_1         fatomic.Padded64Int64
	read       fatomic.Padded64Int64
	readCache  fatomic.Padded64Int64
	write      fatomic.Padded64Int64
	writeCache fatomic.Padded64Int64
	_2         fatomic.Padded64Int64
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
	write := q.write.Value
	readLimit := write - q.size
	if readLimit >= q.readCache.Value {
		q.readCache.Value = atomic.LoadInt64(&q.read.Value)
		if readLimit >= q.readCache.Value {
			return false
		}
	}
	idx := write & q.mask
	q.ringBuffer[idx] = val
	fatomic.LazyStore(&q.write.Value, q.write.Value+1)
	return true
}

func (q *PointerQ) Read() unsafe.Pointer {
	read := q.read.Value
	if read == q.writeCache.Value {
		q.writeCache.Value = atomic.LoadInt64(&q.write.Value)
		if read == q.writeCache.Value {
			return nil
		}
	}
	idx := read & q.mask
	val := q.ringBuffer[idx]
	fatomic.LazyStore(&q.read.Value, q.read.Value+1)
	return val
}
