package atomicint

import (
	"sync/atomic"
)

type Value struct {
	before [7]int64
	value int64
	after [8]int64
}

func (v *Value) Get() int64 {
	return atomic.LoadInt64(&v.value)
}

func (v *Value) NakedGet() int64 {
	return v.value
}

func (v *Value) Set(val int64) {
	atomic.StoreInt64(&v.value, val)
}

func (v *Value) NakedSet(val int64) {
	v.value = val
}

func (v *Value) Add(add int64) {
	atomic.AddInt64(&v.value, add)
}
