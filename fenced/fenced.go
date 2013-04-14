package fenced

import (
	"github.com/textnode/fencer"
	"sync/atomic"
)

// NB: The assumption has been made that the Go compiler won't reorder around calls to fencer.*Fence()

type Value struct {
//	before [8]int64
	value int64
//	after [8]int64
}

func (v *Value) Get() int64 {
//	fencer.MFence()
	return v.value
}

func (v *Value) Set(val int64) {
	v.value = val
	fencer.SFence()
}

func (v *Value) SetOrdered(val int64) {
	// How do we ask the compiler not to reorder method calls?
	v.value = val
	//v.Set(val)
}

// Is this method as strict as I require?
func (v *Value) GetAndIncrement() int64 {
	val := v.Get()
	v.Set(val+1)
	return val
}

// Is this method as strict as I require?
func (v *Value) IncrementAndGet() int64 {
	val := v.Get()
	val++
	v.Set(val)
	return val
}

func (v *Value) CompareAndSwap(oldVal, newVal int64) bool {
	return atomic.CompareAndSwapInt64(&v.value, oldVal, newVal)
}
