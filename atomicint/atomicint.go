package atomicint

import "github.com/fmstephe/fatomic"
import "sync/atomic"

type CacheLine struct {
	before [7]int64
	Value int64
	after [8]int64
}

func (v *CacheLine) Get() int64 {
	return atomic.LoadInt64(&v.Value)
}

func (v *CacheLine) Set(val int64) {
	atomic.StoreInt64(&v.Value, val)
}

func (v *CacheLine) LazyStore(val int64) {
	fatomic.LazyStore(&v.Value, val)
}

func (v *CacheLine) Add(add int64) {
	atomic.AddInt64(&v.Value, add)
}
