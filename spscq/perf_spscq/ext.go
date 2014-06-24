package main

import (
	"os"
	"runtime"
	"runtime/pprof"
	"time"
	"unsafe"

	"github.com/fmstephe/fatomic"
	"github.com/fmstephe/queues/spscq"
)

var ringBuffer []unsafe.Pointer
var mask int64

func extqTest(msgCount, batchSize, qSize int64) {
	q := spscq.NewExtQ(qSize)
	done := make(chan bool)
	f, err := os.Create("extq.prof")
	if err != nil {
		panic(err.Error())
	}
	ringBuffer = fatomic.CacheProtectedPointers(int(qSize))
	mask = qSize-1
	pprof.StartCPUProfile(f)
	go extqDequeue(msgCount, q, batchSize, done)
	go extqEnqueue(msgCount, q, batchSize, done)
	<-done
	<-done
	pprof.StopCPUProfile()
}

func extqEnqueue(msgCount int64, q *spscq.ExtQ, batchSize int64, done chan bool) {
	runtime.LockOSThread()
	var t int64
OUTER:
	for {
		low, high := q.WriteBuffer(batchSize)
		for low == high {
			low, high = q.WriteBuffer(batchSize)
		}
		for i := low; i < high; i++ {
			t++
			if t > msgCount {
				q.CommitWriteBuffer(i - low)
				break OUTER
			}
			ringBuffer[i & mask] = unsafe.Pointer(uintptr(t))
		}
		q.CommitWriteBuffer(high - low)
		if t == msgCount {
			break
		}
	}
	done <- true
}

func extqDequeue(msgCount int64, q *spscq.ExtQ, batchSize int64, done chan bool) {
	runtime.LockOSThread()
	start := time.Now().UnixNano()
	var sum int64
	var checksum int64
	var t int64
OUTER:
	for {
		low, high := q.ReadBuffer(batchSize)
		lim := 0
		for low == high {
			low, high = q.ReadBuffer(batchSize)
			lim++
			if lim > 100000 {
			}
		}
		for i := low; i < high; i++ {
			t++
			if t > msgCount {
				q.CommitReadBuffer(i - low)
				break OUTER
			}
			sum += int64(uintptr(ringBuffer[i & mask]))
			checksum += t
		}
		q.CommitReadBuffer(high - low)
		if t == msgCount {
			break
		}
	}
	nanos := time.Now().UnixNano() - start
	printTimings(msgCount, nanos, "extq")
	expect(sum, checksum)
	done <- true
}
