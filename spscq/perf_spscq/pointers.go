package main

import (
	"os"
	"runtime"
	"runtime/pprof"
	"time"
	"unsafe"

	"github.com/fmstephe/queues/spscq"
)

func pqTest(msgCount, qSize int64) {
	q := spscq.NewPointerQ(qSize)
	done := make(chan bool)
	f, err := os.Create("pq.prof")
	if err != nil {
		panic(err.Error())
	}
	pprof.StartCPUProfile(f)
	go pqDequeue(msgCount, q, done)
	go pqEnqueue(msgCount, q, done)
	<-done
	<-done
	pprof.StopCPUProfile()
}

func pqEnqueue(msgCount int64, q *spscq.PointerQ, done chan bool) {
	runtime.LockOSThread()
	t := 1
	var v unsafe.Pointer
	for i := int64(0); i < msgCount; i++ {
		v = unsafe.Pointer(uintptr(uint(t)))
		w := q.Write(v)
		for w == false {
			w = q.Write(v)
		}
		t++
	}
	done <- true
}

func pqDequeue(msgCount int64, q *spscq.PointerQ, done chan bool) {
	runtime.LockOSThread()
	start := time.Now().UnixNano()
	sum := int64(0)
	checksum := int64(0)
	var v unsafe.Pointer
	for i := int64(0); i < msgCount; i++ {
		v = q.Read()
		for v == nil {
			v = q.Read()
		}
		sum += int64(uintptr(v))
		checksum += i + 1
	}
	nanos := time.Now().UnixNano() - start
	printTimings(msgCount, nanos, "pq")
	expect(sum, checksum)
	done <- true
}
