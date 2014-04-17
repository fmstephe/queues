package atomicint

import (
	"sync/atomic"
)

type CacheLine struct {
	before [7]int64
	Value int64
	after [8]int64
}

func (v *CacheLine) Get() int64 {
	return atomic.LoadInt64(&v.Value)
}

func (v *CacheLine) NakedGet() int64 {
	return v.Value
}

func (v *CacheLine) Set(val int64) {
	atomic.StoreInt64(&v.Value, val)
}

func (v *CacheLine) NakedSet(val int64) {
	v.Value = val
}

func (v *CacheLine) Add(add int64) {
	atomic.AddInt64(&v.Value, add)
}
